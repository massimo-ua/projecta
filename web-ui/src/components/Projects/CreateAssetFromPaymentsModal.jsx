import React, { useState, useEffect, useMemo } from 'react';
import PropTypes from 'prop-types';
import { useIntlayer } from 'react-intlayer';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
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
  DialogFooter,
  DialogDescription,
} from '@/components/ui/dialog';
import { toast } from 'sonner';
import { Loader2, AlertCircle, Coins, Layers } from 'lucide-react';
import { assetRepository } from '../../api';

const SUPPORTED_CURRENCIES = ['UAH', 'USD', 'EUR', 'PLN'];

export default function CreateAssetFromPaymentsModal({
  open,
  onSuccess,
  onCancel,
  projectId,
  types = [],
  payments = [],
}) {
  const content = useIntlayer('payments');

  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [typeId, setTypeId] = useState('');
  const [currency, setCurrency] = useState('UAH');
  const [acquiredAt, setAcquiredAt] = useState('');
  const [loading, setLoading] = useState(false);

  // Check currency homogeneity
  const allSameCurrency = useMemo(() => {
    if (!payments || payments.length === 0) return true;
    const firstCurr = payments[0].currency;
    return payments.every((p) => p.currency === firstCurr);
  }, [payments]);

  // Check type homogeneity
  const allSameType = useMemo(() => {
    if (!payments || payments.length === 0) return true;
    const firstTypeId = payments[0].typeId;
    return Boolean(firstTypeId && payments.every((p) => p.typeId === firstTypeId));
  }, [payments]);

  // Reset / compute defaults when opened or payments change
  useEffect(() => {
    if (!open || !payments || payments.length === 0) return;

    // Currency default: identical currency if all match, else homeCurrency (or UAH)
    const initialCurrency = allSameCurrency
      ? payments[0].currency
      : (payments[0]?.homeCurrency || 'UAH');
    setCurrency(initialCurrency);

    // Type default: if all share the same type, preselect it
    if (allSameType && payments[0].typeId) {
      setTypeId(payments[0].typeId);
    } else {
      setTypeId('');
    }

    // Acquired date default: latest payment date
    let latestDate = null;
    payments.forEach((p) => {
      let d = null;
      if (p.rawDate) {
        d = new Date(p.rawDate);
      } else if (p.paymentDate) {
        const parts = p.paymentDate.split('/');
        if (parts.length === 3) {
          d = new Date(`${parts[2]}-${parts[1]}-${parts[0]}`);
        }
      }
      if (d && !Number.isNaN(d.getTime())) {
        if (!latestDate || d > latestDate) {
          latestDate = d;
        }
      }
    });

    const dateStr = latestDate
      ? latestDate.toISOString().split('T')[0]
      : new Date().toISOString().split('T')[0];
    setAcquiredAt(dateStr);

    setName('');
    setDescription('');
  }, [open, payments, allSameCurrency, allSameType]);

  // Estimated total calculation for preview
  const estimatedTotal = useMemo(() => {
    if (!payments || payments.length === 0) return null;

    if (allSameCurrency && currency === payments[0].currency) {
      const sumMinor = payments.reduce((acc, p) => acc + (p.rawAmount || 0), 0);
      return (sumMinor / 100).toFixed(2);
    }

    const homeCurr = payments[0]?.homeCurrency || 'UAH';
    if (!allSameCurrency && currency === homeCurr) {
      const sumMinor = payments.reduce((acc, p) => acc + (p.rawHomeAmount || p.rawAmount || 0), 0);
      return (sumMinor / 100).toFixed(2);
    }

    return null;
  }, [payments, allSameCurrency, currency]);

  const handleSubmit = async (e) => {
    e.preventDefault();

    if (!name.trim()) {
      toast.error(`${String(content?.assetNameLabel || 'Name')} is required`);
      return;
    }

    if (!typeId) {
      toast.error(String(content?.selectTypePlaceholder || 'Please select a type'));
      return;
    }

    setLoading(true);
    try {
      const paymentIds = payments.map((p) => p.id);
      const createdAsset = await assetRepository.createAssetFromPayments(projectId, {
        paymentIds,
        name: name.trim(),
        description: description.trim(),
        typeId,
        acquiredAt: acquiredAt ? new Date(acquiredAt) : undefined,
        targetCurrency: currency,
      });

      onSuccess(createdAsset);
    } catch (err) {
      toast.error(`${String(content?.failedToCreateAsset || 'Failed to create asset')}: ${err.message}`);
      console.error('Failed to create asset from payments:', err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={(isOpen) => !isOpen && onCancel()}>
      <DialogContent className="sm:max-w-[500px] rounded-2xl p-6">
        <DialogHeader>
          <DialogTitle className="text-xl font-bold tracking-tight">
            {String(content?.createAssetTitle || 'Create Asset from Payments')}
          </DialogTitle>
          <DialogDescription className="text-xs text-muted-foreground">
            {String(content?.createAssetDesc || 'Summarize the selected payments into a new asset record.')}
          </DialogDescription>
        </DialogHeader>

        {/* Selected Payments Summary Banner */}
        <div className="rounded-xl border border-border/70 p-3 bg-muted/20 space-y-2">
          <div className="flex items-center justify-between">
            <span className="flex items-center gap-1.5 text-xs font-semibold text-foreground">
              <Layers className="h-4 w-4 text-primary" />
              <span>{payments.length} {String(content?.selectedPayments || 'payments selected')}</span>
            </span>
            {estimatedTotal && (
              <span className="flex items-center gap-1 text-sm font-bold text-foreground">
                <Coins className="h-4 w-4 text-emerald-500" />
                <span>{estimatedTotal} {currency}</span>
              </span>
            )}
          </div>

          {!allSameCurrency && (
            <div className="flex items-center gap-1.5 text-[11px] text-amber-600 dark:text-amber-400">
              <AlertCircle className="h-3.5 w-3.5 shrink-0" />
              <span>
                {String(content?.currencyConversionNotice || 'Payments in differing currencies will be converted into')} {currency}.
              </span>
            </div>
          )}
        </div>

        <form onSubmit={handleSubmit} className="space-y-4 py-2">
          {/* Asset Name */}
          <div className="space-y-2">
            <Label htmlFor="asset-name" className="text-xs font-semibold">
              {String(content?.assetNameLabel || 'Asset Name')} *
            </Label>
            <Input
              id="asset-name"
              placeholder={String(content?.assetNamePlaceholder || 'Enter asset name...')}
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="rounded-xl"
              required
            />
          </div>

          {/* Cost Type */}
          <div className="space-y-2">
            <Label htmlFor="asset-type" className="text-xs font-semibold">
              {String(content?.typeLabel || 'Type')} *
            </Label>
            <Select value={typeId} onValueChange={setTypeId}>
              <SelectTrigger id="asset-type" className="rounded-xl">
                <SelectValue placeholder={String(content?.selectTypePlaceholder || 'Select type')} />
              </SelectTrigger>
              <SelectContent className="rounded-xl">
                {types.map((type) => (
                  <SelectItem key={type.id} value={type.id} className="rounded-lg">
                    {type.name} [{type.category}]
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* Currency and Acquired Date */}
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="asset-currency" className="text-xs font-semibold">
                {String(content?.targetCurrencyLabel || 'Resulting Currency')}
              </Label>
              <Select value={currency} onValueChange={setCurrency}>
                <SelectTrigger id="asset-currency" className="rounded-xl">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent className="rounded-xl">
                  {SUPPORTED_CURRENCIES.map((c) => (
                    <SelectItem key={c} value={c} className="rounded-lg">
                      {c}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label htmlFor="asset-acquired" className="text-xs font-semibold">
                {String(content?.dateLabel || 'Acquired Date')}
              </Label>
              <Input
                id="asset-acquired"
                type="date"
                value={acquiredAt}
                onChange={(e) => setAcquiredAt(e.target.value)}
                className="rounded-xl"
                required
              />
            </div>
          </div>

          {/* Description */}
          <div className="space-y-2">
            <Label htmlFor="asset-desc" className="text-xs font-semibold">
              {String(content?.descriptionLabel || 'Description')}
            </Label>
            <Textarea
              id="asset-desc"
              rows={2}
              placeholder={String(content?.descriptionPlaceholder || 'Asset description...')}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="rounded-xl resize-none"
            />
          </div>

          <DialogFooter className="pt-3 gap-2 sm:gap-0">
            <Button
              type="button"
              variant="outline"
              onClick={onCancel}
              disabled={loading}
              className="rounded-xl"
            >
              {String(content?.cancelButton || 'Cancel')}
            </Button>
            <Button
              type="submit"
              disabled={loading}
              className="rounded-xl font-semibold shadow-sm shadow-primary/20 gap-2"
            >
              {loading && <Loader2 className="h-4 w-4 animate-spin" />}
              <span>{String(content?.submitButton || 'Submit')}</span>
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

CreateAssetFromPaymentsModal.propTypes = {
  open: PropTypes.bool.isRequired,
  onSuccess: PropTypes.func.isRequired,
  onCancel: PropTypes.func.isRequired,
  projectId: PropTypes.string.isRequired,
  types: PropTypes.arrayOf(
    PropTypes.shape({
      id: PropTypes.string.isRequired,
      name: PropTypes.string.isRequired,
      category: PropTypes.string,
    }),
  ),
  payments: PropTypes.arrayOf(
    PropTypes.shape({
      id: PropTypes.string.isRequired,
      currency: PropTypes.string,
      homeCurrency: PropTypes.string,
      rawAmount: PropTypes.number,
      rawHomeAmount: PropTypes.number,
      typeId: PropTypes.string,
    }),
  ),
};

CreateAssetFromPaymentsModal.defaultProps = {
  types: [],
  payments: [],
};

