import React, { useState, useEffect, useRef } from 'react';
import PropTypes from 'prop-types';
import { useParams, useNavigate } from 'react-router-dom';
import { useIntlayer, useLocale } from 'react-intlayer';
import { getLocalizedUrl } from 'intlayer';
import {
  DollarSign,
  Calendar,
  FileUp,
  Package,
  X,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Checkbox } from '@/components/ui/checkbox';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { toast } from 'sonner';
import { cn } from '@/lib/utils';
import usePayments from '../../hooks/payments';
import AddPaymentModal from './AddPaymentModal';
import EditPaymentModal from './EditPaymentModal';
import ImportStatementModal from './ImportStatementModal';
import CreateAssetFromPaymentsModal from './CreateAssetFromPaymentsModal';
import { ListView } from './ListView';
import { EditButton } from './ListView/EditButton';
import { RemoveButton } from './ListView/RemoveButton';
import { CopyableText } from './ListView/CopyableText';
import { DetailItem } from './ListView/DetailItem';
import './Payments.css';

function PaymentMainContent({ payment }) {
  return (
    <div className="flex flex-col gap-1.5">
      <div className="flex items-baseline gap-2 flex-wrap">
        <span className="flex items-center gap-1 text-xs text-muted-foreground font-mono">
          <Calendar className="h-3 w-3" />
          {payment.paymentDate}
        </span>
        <span className="font-semibold text-base text-foreground">{payment.description}</span>
      </div>
      <div className="flex gap-1.5 flex-wrap">
        <Badge variant="outline" className="rounded-md text-[11px] font-medium border-border/70 text-muted-foreground">{payment.category}</Badge>
        <Badge variant="secondary" className="rounded-md text-[11px] font-medium">{payment.type}</Badge>
      </div>
    </div>
  );
}

PaymentMainContent.propTypes = {
  payment: PropTypes.shape({
    paymentDate: PropTypes.string,
    description: PropTypes.string,
    category: PropTypes.string,
    type: PropTypes.string,
  }).isRequired,
};

function PaymentAmount({ payment }) {
  return (
    <div className="flex flex-col items-end gap-0.5">
      <span
        className={cn(
          'px-2.5 py-1 rounded-xl text-sm font-bold tracking-tight inline-block',
          payment.isDownPayment
            ? 'bg-rose-500/10 text-rose-600 dark:text-rose-400 border border-rose-500/20'
            : 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20',
        )}
      >
        {payment.formattedAmount}
      </span>
      {payment.hasDifferentHomeCurrency && payment.formattedHomeAmount && (
        <span className="text-[11px] text-muted-foreground font-mono">
          {payment.formattedHomeAmount}
        </span>
      )}
    </div>
  );
}

PaymentAmount.propTypes = {
  payment: PropTypes.shape({
    isDownPayment: PropTypes.bool,
    formattedAmount: PropTypes.string,
    hasDifferentHomeCurrency: PropTypes.bool,
    formattedHomeAmount: PropTypes.string,
  }).isRequired,
};

function PaymentDetails({ payment, typeLabel, categoryLabel }) {
  return (
    <div className="space-y-2">
      <DetailItem label="ID">
        <CopyableText text={payment.id} truncate />
      </DetailItem>
      <DetailItem label={typeLabel}>
        <span className="text-sm text-foreground">{payment.type}</span>
      </DetailItem>
      <DetailItem label={categoryLabel}>
        <span className="text-sm text-foreground">{payment.category}</span>
      </DetailItem>
    </div>
  );
}

PaymentDetails.propTypes = {
  payment: PropTypes.shape({
    id: PropTypes.string,
    type: PropTypes.string,
    category: PropTypes.string,
  }).isRequired,
  typeLabel: PropTypes.string.isRequired,
  categoryLabel: PropTypes.string.isRequired,
};

function DebouncedDateInput({
  value,
  onChange,
  className,
}) {
  const [localValue, setLocalValue] = useState(value || '');
  const timerRef = useRef(null);

  useEffect(() => {
    setLocalValue(value || '');
  }, [value]);

  useEffect(() => () => {
    if (timerRef.current) {
      clearTimeout(timerRef.current);
    }
  }, []);

  const isValidFilterDate = (val) => {
    if (!val) return true;
    return /^\d{4}-\d{2}-\d{2}$/.test(val) && parseInt(val.slice(0, 4), 10) >= 1000;
  };

  const handleChange = (e) => {
    const nextVal = e.target.value;
    setLocalValue(nextVal);

    if (timerRef.current) {
      clearTimeout(timerRef.current);
      timerRef.current = null;
    }

    if (!nextVal) {
      if (value !== '') {
        onChange('');
      }
      return;
    }

    if (isValidFilterDate(nextVal)) {
      timerRef.current = setTimeout(() => {
        timerRef.current = null;
        onChange(nextVal);
      }, 500);
    }
  };

  const handleBlur = () => {
    if (timerRef.current) {
      clearTimeout(timerRef.current);
      timerRef.current = null;
    }

    if (!localValue) {
      if (value !== '') {
        onChange('');
      }
    } else if (isValidFilterDate(localValue)) {
      if (localValue !== value) {
        onChange(localValue);
      }
    } else {
      setLocalValue(value || '');
    }
  };

  const handleKeyDown = (e) => {
    if (e.key === 'Enter') {
      if (timerRef.current) {
        clearTimeout(timerRef.current);
        timerRef.current = null;
      }
      if (!localValue) {
        if (value !== '') {
          onChange('');
        }
      } else if (isValidFilterDate(localValue)) {
        if (localValue !== value) {
          onChange(localValue);
        }
      } else {
        setLocalValue(value || '');
      }
    }
  };

  return (
    <Input
      type="date"
      value={localValue}
      onChange={handleChange}
      onBlur={handleBlur}
      onKeyDown={handleKeyDown}
      className={className}
    />
  );
}

DebouncedDateInput.propTypes = {
  value: PropTypes.string,
  onChange: PropTypes.func.isRequired,
  className: PropTypes.string,
};

DebouncedDateInput.defaultProps = {
  value: '',
  className: '',
};

export function Payments() {
  const content = useIntlayer('payments');
  const { projectId } = useParams();
  const navigate = useNavigate();
  const { locale } = useLocale();

  const {
    loading,
    payments,
    total,
    currentPage,
    pageSize,
    types,
    filters,
    setFilter,
    resetFilters,
    selectedPaymentsMap,
    selectedPaymentsList,
    selectedCount,
    toggleSelectPayment,
    toggleSelectAllPage,
    isAllPageSelected,
    clearSelection,
    addModalOpened,
    paymentIdToEdit,
    onPaginationChange,
    openAddModal,
    closeAddModal,
    onAddSuccess,
    openEditModal,
    closeEditModal,
    onEditSuccess,
    removePayment,
  } = usePayments(projectId);

  const [importModalOpened, setImportModalOpened] = useState(false);
  const openImportModal = () => setImportModalOpened(true);
  const closeImportModal = () => setImportModalOpened(false);

  const [createAssetModalOpened, setCreateAssetModalOpened] = useState(false);
  const openCreateAssetModal = () => setCreateAssetModalOpened(true);
  const closeCreateAssetModal = () => setCreateAssetModalOpened(false);

  const handleRemove = (paymentId) => {
    removePayment(paymentId, {
      successMessage: String(content.paymentRemovedSuccess),
      errorMessage: String(content.failedToRemove),
    });
  };

  const handleAssetCreated = () => {
    closeCreateAssetModal();
    clearSelection();
    toast.success(String(content.assetCreatedSuccess || 'Asset created successfully from selected payments'), {
      action: {
        label: String(content.viewInAssets || 'View in Assets'),
        onClick: () => navigate(getLocalizedUrl(`/projects/${projectId}/assets`, locale)),
      },
    });
  };

  const hasActiveFilters = Boolean(filters.typeId || filters.fromDate || filters.toDate);

  return (
    <>
      <ListView
        loading={loading}
        items={payments}
        total={total}
        pageSize={pageSize}
        currentPage={currentPage}
        onPaginationChange={onPaginationChange}
        onAddButtonClick={openAddModal}
        addButtonIcon={<DollarSign className="h-4 w-4" />}
        addButtonText={String(content.addPayment)}
        addButtonDisabled={addModalOpened}
        extraActions={(
          <Button
            variant="outline"
            onClick={openImportModal}
            className="gap-2 font-semibold shadow-sm hover:shadow-md transition-all rounded-xl h-10 px-4 border-dashed hover:border-primary/50"
          >
            <FileUp className="h-4 w-4 text-primary" />
            <span>{String(content.importStatement || 'Import Statement')}</span>
          </Button>
        )}
        filterBar={(
          <div className="rounded-2xl border border-border/60 bg-card p-4 space-y-3">
            <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3 flex-wrap">
              <div className="flex items-center gap-3 flex-wrap flex-1">
                {/* Type Filter */}
                <div className="w-full sm:w-48">
                  <Select
                    value={filters.typeId || 'ALL'}
                    onValueChange={(val) => setFilter('typeId', val === 'ALL' ? '' : val)}
                  >
                    <SelectTrigger className="rounded-xl h-9 text-xs">
                      <SelectValue placeholder={String(content.filterAllTypes || 'All Types')} />
                    </SelectTrigger>
                    <SelectContent className="rounded-xl">
                      <SelectItem value="ALL" className="rounded-lg text-xs font-medium">
                        {String(content.filterAllTypes || 'All Types')}
                      </SelectItem>
                      {types.map((type) => (
                        <SelectItem key={type.id} value={type.id} className="rounded-lg text-xs">
                          {type.name} [{type.category}]
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>

                {/* Date Range: From */}
                <div className="flex items-center gap-1.5 w-full sm:w-auto">
                  <Label className="text-[11px] text-muted-foreground whitespace-nowrap">
                    {String(content.filterDateFrom || 'From')}:
                  </Label>
                  <DebouncedDateInput
                    value={filters.fromDate}
                    onChange={(val) => setFilter('fromDate', val)}
                    className="rounded-xl h-9 text-xs w-full sm:w-36"
                  />
                </div>

                {/* Date Range: To */}
                <div className="flex items-center gap-1.5 w-full sm:w-auto">
                  <Label className="text-[11px] text-muted-foreground whitespace-nowrap">
                    {String(content.filterDateTo || 'To')}:
                  </Label>
                  <DebouncedDateInput
                    value={filters.toDate}
                    onChange={(val) => setFilter('toDate', val)}
                    className="rounded-xl h-9 text-xs w-full sm:w-36"
                  />
                </div>

                {/* Reset Filters */}
                {hasActiveFilters && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={resetFilters}
                    className="gap-1 text-xs text-muted-foreground hover:text-foreground h-9 px-2.5 rounded-xl"
                  >
                    <X className="h-3.5 w-3.5" />
                    <span>{String(content.resetFilters || 'Reset')}</span>
                  </Button>
                )}
              </div>

              {/* Select All on Page Toggle */}
              {payments.length > 0 && (
                <div className="flex items-center gap-2 pt-2 sm:pt-0 border-t sm:border-0 border-border/40">
                  <Checkbox
                    id="select-all-page"
                    checked={isAllPageSelected}
                    onCheckedChange={toggleSelectAllPage}
                  />
                  <Label htmlFor="select-all-page" className="text-xs text-muted-foreground cursor-pointer select-none">
                    {String(content.selectAll || 'Select all on page')}
                  </Label>
                </div>
              )}
            </div>
          </div>
        )}
        batchActionBar={selectedCount > 0 ? (
          <div className="flex items-center justify-between gap-3 p-3.5 rounded-2xl bg-primary/5 border border-primary/20 shadow-xs flex-wrap">
            <div className="flex items-center gap-3">
              <div className="flex items-center gap-2">
                <Badge variant="default" className="rounded-lg font-bold">
                  {selectedCount}
                </Badge>
                <span className="text-xs font-semibold text-foreground">
                  {String(content.selectedPayments || 'payments selected')}
                </span>
              </div>
              <Button
                variant="ghost"
                size="sm"
                onClick={clearSelection}
                className="text-xs text-muted-foreground hover:text-foreground h-7 px-2"
              >
                {String(content.clearSelection || 'Clear')}
              </Button>
            </div>

            <Button
              size="sm"
              onClick={openCreateAssetModal}
              className="gap-2 rounded-xl font-semibold shadow-sm shadow-primary/20"
            >
              <Package className="h-4 w-4" />
              <span>{String(content.createAssetFromSelected || 'Create Asset from Selected')}</span>
            </Button>
          </div>
        ) : null}
        renderItemPrefix={(payment) => (
          <Checkbox
            checked={Boolean(selectedPaymentsMap[payment.id])}
            onCheckedChange={() => toggleSelectPayment(payment)}
            aria-label={`Select payment ${payment.description}`}
          />
        )}
        renderItemMainContent={(payment) => <PaymentMainContent payment={payment} />}
        renderItemAmount={(payment) => <PaymentAmount payment={payment} />}
        renderItemDetails={(payment) => (
          <PaymentDetails
            payment={payment}
            typeLabel={String(content.typeLabel)}
            categoryLabel={String(content.categoryLabel)}
          />
        )}
        renderItemActions={(payment) => (
          <>
            <EditButton onClick={() => openEditModal(payment.id)} />
            <RemoveButton onRemove={() => handleRemove(payment.id)} />
          </>
        )}
      />

      <AddPaymentModal
        types={types}
        projectId={projectId}
        open={addModalOpened}
        onCancel={closeAddModal}
        onSuccess={() => onAddSuccess(String(content.paymentAddedSuccess))}
      />

      <ImportStatementModal
        types={types}
        projectId={projectId}
        open={importModalOpened}
        onCancel={closeImportModal}
        onSuccess={() => {
          closeImportModal();
          onAddSuccess(String(content.paymentsImportedSuccess || 'Payments imported successfully'));
        }}
      />

      <EditPaymentModal
        types={types}
        projectId={projectId}
        paymentId={paymentIdToEdit}
        open={!!paymentIdToEdit}
        onCancel={closeEditModal}
        onSuccess={() => onEditSuccess(String(content.paymentUpdatedSuccess))}
      />

      <CreateAssetFromPaymentsModal
        open={createAssetModalOpened}
        onCancel={closeCreateAssetModal}
        onSuccess={handleAssetCreated}
        projectId={projectId}
        types={types}
        payments={selectedPaymentsList}
      />
    </>
  );
}

export default Payments;
