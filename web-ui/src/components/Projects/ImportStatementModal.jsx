import React, { useState, useRef } from 'react';
import PropTypes from 'prop-types';
import { format, parseISO } from 'date-fns';
import { useIntlayer } from 'react-intlayer';
import {
  Loader2,
  UploadCloud,
  AlertTriangle,
  FileText,
  ArrowLeft,
  CheckSquare,
  Square,
} from 'lucide-react';
import { paymentRepository } from '../../api';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog';
import { toast } from 'sonner';

export default function ImportStatementModal({
  open,
  projectId,
  types,
  onCancel,
  onSuccess,
}) {
  const content = useIntlayer('payments');
  const fileInputRef = useRef(null);

  const [parsing, setParsing] = useState(false);
  const [importing, setImporting] = useState(false);
  const [statement, setStatement] = useState(null);
  const [items, setItems] = useState([]);
  const [defaultTypeId, setDefaultTypeId] = useState('');

  const resetState = () => {
    setParsing(false);
    setImporting(false);
    setStatement(null);
    setItems([]);
    setDefaultTypeId('');
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  const handleClose = () => {
    resetState();
    onCancel();
  };

  const handleFileChange = async (e) => {
    const file = e.target.files?.[0];
    if (!file) return;

    if (!file.name.toLowerCase().endsWith('.pdf')) {
      toast.error('Only PDF files are supported');
      return;
    }

    setParsing(true);
    try {
      const result = await paymentRepository.parseStatement(projectId, file);
      setStatement(result);

      const initialItems = (result.transactions || []).map((tx, index) => ({
        id: `tx-${index}`,
        selected: !tx.is_duplicate,
        docNum: tx.doc_num,
        date: tx.date,
        amount: tx.amount / 100,
        currency: tx.currency || result.currency || 'PLN',
        description: tx.description,
        typeId: defaultTypeId || (types[0]?.id || ''),
        isDuplicate: Boolean(tx.is_duplicate),
        kind: 'UPON_COMPLETION',
      }));

      setItems(initialItems);
      toast.success(String(content?.statementParsed || 'Statement parsed successfully'));
    } catch (err) {
      toast.error(`${String(content?.failedToParseStatement || 'Failed to parse statement')}: ${err.message}`);
    } finally {
      setParsing(false);
    }
  };

  const handleToggleSelectAll = () => {
    const anyUnselected = items.some((item) => !item.selected);
    setItems(items.map((item) => ({ ...item, selected: anyUnselected })));
  };

  const handleToggleRow = (id) => {
    setItems(
      items.map((item) => (item.id === id ? { ...item, selected: !item.selected } : item)),
    );
  };

  const handleRowChange = (id, field, value) => {
    setItems(
      items.map((item) => (item.id === id ? { ...item, [field]: value } : item)),
    );
  };

  const handleApplyDefaultType = () => {
    if (!defaultTypeId) return;
    setItems(
      items.map((item) => (item.selected ? { ...item, typeId: defaultTypeId } : item)),
    );
  };

  const handleImport = async () => {
    const selectedItems = items.filter((item) => item.selected);

    if (selectedItems.length === 0) {
      toast.error('Please select at least one payment to import');
      return;
    }

    const missingType = selectedItems.some((item) => !item.typeId);
    if (missingType) {
      toast.error(String(content?.selectTypeForSelectedWarning || 'Please select a Cost Type for all checked payments'));
      return;
    }

    setImporting(true);
    try {
      const payload = selectedItems.map((item) => ({
        typeId: item.typeId,
        amount: item.amount,
        currency: item.currency,
        paymentDate: new Date(item.date),
        description: item.description,
        paymentKind: item.kind,
      }));

      await paymentRepository.addPaymentsBatch(projectId, payload);
      toast.success(String(content?.paymentsImportedSuccess || 'Payments imported successfully'));
      resetState();
      onSuccess();
    } catch (err) {
      toast.error(`${String(content?.failedToImportPayments || 'Failed to import payments')}: ${err.message}`);
    } finally {
      setImporting(false);
    }
  };

  const selectedCount = items.filter((item) => item.selected).length;
  const selectedTotal = items
    .filter((item) => item.selected)
    .reduce((sum, item) => sum + item.amount, 0);

  return (
    <Dialog open={open} onOpenChange={(isOpen) => !isOpen && handleClose()}>
      <DialogContent className="max-w-4xl max-h-[90vh] flex flex-col p-6 overflow-hidden">
        <DialogHeader className="pb-3 border-b">
          <DialogTitle className="text-xl font-bold flex items-center gap-2">
            <FileText className="h-5 w-5 text-primary" />
            <span>{String(content?.importStatementTitle || 'Import Statement (Kredobank)')}</span>
          </DialogTitle>
          <DialogDescription className="text-xs text-muted-foreground">
            {String(content?.importStatementDesc || 'Upload a Kredobank PDF account statement to extract and import payments.')}
          </DialogDescription>
        </DialogHeader>

        {!statement ? (
          <div className="py-10 flex flex-col items-center justify-center">
            {parsing ? (
              <div className="flex flex-col items-center gap-3">
                <Loader2 className="h-10 w-10 animate-spin text-primary" />
                <p className="text-sm font-medium text-muted-foreground">
                  {String(content?.uploading || 'Parsing PDF statement...')}
                </p>
              </div>
            ) : (
              <button
                type="button"
                onClick={() => fileInputRef.current?.click()}
                className="w-full max-w-lg border-2 border-dashed border-border/80 hover:border-primary/60 rounded-2xl p-8 flex flex-col items-center justify-center gap-3 cursor-pointer transition-colors bg-muted/20 hover:bg-muted/30 text-left"
              >
                <div className="h-14 w-14 rounded-full bg-primary/10 flex items-center justify-center text-primary">
                  <UploadCloud className="h-7 w-7" />
                </div>
                <div className="text-center">
                  <p className="text-sm font-semibold text-foreground">
                    {String(content?.dropFilePrompt || 'Choose a PDF statement or drag & drop here')}
                  </p>
                  <p className="text-xs text-muted-foreground mt-1">
                    Kredobank PDF statement (.pdf)
                  </p>
                </div>
                <input
                  ref={fileInputRef}
                  type="file"
                  accept=".pdf,application/pdf"
                  onChange={handleFileChange}
                  className="hidden"
                />
              </button>
            )}
          </div>
        ) : (
          <div className="flex-1 overflow-hidden flex flex-col gap-4 py-2">
            {/* Metadata Bar */}
            <div className="flex items-center justify-between bg-muted/40 p-3 rounded-xl text-xs gap-3 flex-wrap">
              <div className="flex items-center gap-4 flex-wrap">
                {statement.account && (
                  <div>
                    <span className="text-muted-foreground mr-1">
                      {`${String(content?.accountLabel || 'Account')}:`}
                    </span>
                    <span className="font-mono font-medium">{statement.account}</span>
                  </div>
                )}
                {statement.period && (
                  <div>
                    <span className="text-muted-foreground mr-1">
                      {`${String(content?.periodLabel || 'Period')}:`}
                    </span>
                    <span className="font-medium">{statement.period}</span>
                  </div>
                )}
                <div>
                  <span className="text-muted-foreground mr-1">
                    {`${String(content?.currencyLabel || 'Currency')}:`}
                  </span>
                  <Badge variant="outline" className="font-mono text-[11px]">
                    {statement.currency}
                  </Badge>
                </div>
              </div>

              <Button
                variant="ghost"
                size="sm"
                onClick={resetState}
                className="text-xs h-7 text-muted-foreground hover:text-foreground"
              >
                <ArrowLeft className="h-3.5 w-3.5 mr-1" />
                <span>{String(content?.selectAnotherFile || 'Change file')}</span>
              </Button>
            </div>

            {/* Bulk Actions */}
            <div className="flex items-center justify-between gap-4 flex-wrap pb-1">
              <div className="flex items-center gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleToggleSelectAll}
                  className="h-8 text-xs gap-1.5"
                >
                  {items.every((item) => item.selected) ? (
                    <CheckSquare className="h-3.5 w-3.5" />
                  ) : (
                    <Square className="h-3.5 w-3.5" />
                  )}
                  <span>{String(content?.selectAll || 'Select all')}</span>
                </Button>
                <span className="text-xs text-muted-foreground">
                  {`${selectedCount} / ${items.length} ${String(content?.selectedCount || 'selected')}`}
                </span>
              </div>

              <div className="flex items-center gap-2">
                <div className="w-48">
                  <Select value={defaultTypeId} onValueChange={setDefaultTypeId}>
                    <SelectTrigger className="h-8 text-xs rounded-lg">
                      <SelectValue placeholder={String(content?.defaultCostType || 'Default Cost Type')} />
                    </SelectTrigger>
                    <SelectContent className="rounded-xl">
                      {types.map((type) => (
                        <SelectItem key={type.id} value={type.id} className="text-xs">
                          {`${type.name} [${type.category}]`}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <Button
                  variant="secondary"
                  size="sm"
                  onClick={handleApplyDefaultType}
                  disabled={!defaultTypeId}
                  className="h-8 text-xs"
                >
                  <span>{String(content?.applyToAll || 'Apply to all')}</span>
                </Button>
              </div>
            </div>

            {/* Transactions Preview Table */}
            <div className="flex-1 overflow-y-auto border border-border/70 rounded-xl">
              <table className="w-full text-left text-xs border-collapse">
                <thead className="bg-muted/60 sticky top-0 z-10 border-b border-border/70">
                  <tr>
                    <th className="p-2.5 w-8" aria-label="Selection" />
                    <th className="p-2.5 w-24">{String(content?.dateLabel || 'Date')}</th>
                    <th className="p-2.5 w-28 text-right">{String(content?.amountLabel || 'Amount')}</th>
                    <th className="p-2.5 min-w-[200px]">{String(content?.descriptionLabel || 'Description')}</th>
                    <th className="p-2.5 w-48">{String(content?.typeLabel || 'Type')}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border/50">
                  {items.map((item) => (
                    <tr
                      key={item.id}
                      className={`hover:bg-muted/30 transition-colors ${
                        item.selected ? 'bg-background' : 'opacity-50 bg-muted/10'
                      }`}
                    >
                      <td className="p-2.5 align-middle">
                        <input
                          type="checkbox"
                          aria-label={`Select transaction ${item.docNum}`}
                          checked={item.selected}
                          onChange={() => handleToggleRow(item.id)}
                          className="rounded border-gray-300 text-primary focus:ring-primary h-4 w-4 cursor-pointer"
                        />
                      </td>
                      <td className="p-2.5 font-mono text-[11px] align-middle whitespace-nowrap">
                        {item.date ? format(parseISO(item.date), 'dd/MM/yyyy') : ''}
                      </td>
                      <td className="p-2.5 text-right font-mono font-bold align-middle whitespace-nowrap">
                        {`${item.amount.toFixed(2)} ${item.currency}`}
                      </td>
                      <td className="p-2.5 align-middle">
                        <div className="flex flex-col gap-1">
                          <Input
                            aria-label={`Description for transaction ${item.docNum}`}
                            value={item.description}
                            onChange={(e) => handleRowChange(item.id, 'description', e.target.value)}
                            className="h-7 text-xs font-normal rounded-md"
                          />
                          {item.isDuplicate && (
                            <span className="flex items-center gap-1 text-[10px] text-amber-600 dark:text-amber-400 font-medium">
                              <AlertTriangle className="h-3 w-3" />
                              <span>{String(content?.potentialDuplicate || 'Potential duplicate')}</span>
                            </span>
                          )}
                        </div>
                      </td>
                      <td className="p-2.5 align-middle">
                        <Select
                          value={item.typeId}
                          onValueChange={(val) => handleRowChange(item.id, 'typeId', val)}
                        >
                          <SelectTrigger className="h-7 text-xs rounded-md">
                            <SelectValue placeholder="Type..." />
                          </SelectTrigger>
                          <SelectContent className="rounded-xl">
                            {types.map((type) => (
                              <SelectItem key={type.id} value={type.id} className="text-xs">
                                {`${type.name} [${type.category}]`}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            {/* Total Footer */}
            <div className="flex items-center justify-between text-xs px-2 pt-1 font-medium">
              <span className="text-muted-foreground">
                {`${selectedCount} ${String(content?.selectedCount || 'selected')}`}
              </span>
              <span className="font-semibold text-sm">
                {`Total: ${selectedTotal.toFixed(2)} ${statement?.currency || ''}`}
              </span>
            </div>
          </div>
        )}

        <DialogFooter className="pt-3 border-t flex justify-end gap-2">
          <Button variant="outline" onClick={handleClose} disabled={importing}>
            <span>{String(content?.cancelButton || 'Cancel')}</span>
          </Button>
          {statement && (
            <Button
              onClick={handleImport}
              disabled={importing || selectedCount === 0}
              className="gap-2"
            >
              {importing && <Loader2 className="h-4 w-4 animate-spin" />}
              <span>{`${String(content?.importPaymentsCount || 'Import Payments')} (${selectedCount})`}</span>
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

ImportStatementModal.propTypes = {
  open: PropTypes.bool.isRequired,
  projectId: PropTypes.string.isRequired,
  types: PropTypes.arrayOf(
    PropTypes.shape({
      id: PropTypes.string,
      name: PropTypes.string,
      category: PropTypes.string,
    }),
  ),
  onCancel: PropTypes.func.isRequired,
  onSuccess: PropTypes.func.isRequired,
};

ImportStatementModal.defaultProps = {
  types: [],
};
