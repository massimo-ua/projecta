import React, { useState, useEffect } from 'react';
import PropTypes from 'prop-types';
import { useParams } from 'react-router-dom';
import { useIntlayer } from 'react-intlayer';
import { investmentRepository, assetRepository } from '../../api';
import TagSelector from './TagSelector';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
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
  DialogFooter,
} from '@/components/ui/dialog';
import { toast } from 'sonner';
import { Loader2, X, Plus } from 'lucide-react';
import { cn } from '@/lib/utils';

const SUPPORTED_CURRENCIES = ['UAH', 'USD', 'EUR', 'PLN'];

export default function AddPaymentModal({
  open,
  onSuccess,
  onCancel,
  assets: propAssets,
  defaultAssetId,
}) {
  const content = useIntlayer('payments');
  const { projectId } = useParams();

  const resourceTypes = [
    { key: 'MONEY', label: String(content?.moneyResource || 'Капітал (Гроші)') },
    { key: 'TIME', label: String(content?.timeResource || 'Праця (Час)') },
    { key: 'GOODS', label: String(content?.goodsResource || 'Товари та матеріали') },
  ];

  const [loadedAssets, setLoadedAssets] = useState([]);
  const [assetId, setAssetId] = useState(defaultAssetId || '');
  const [isSplit, setIsSplit] = useState(false);
  const [allocations, setAllocations] = useState([]);
  const [resourceType, setResourceType] = useState('MONEY');
  const [amount, setAmount] = useState('');
  const [currency, setCurrency] = useState('UAH');
  const [date, setDate] = useState(new Date().toISOString().split('T')[0]);
  const [description, setDescription] = useState('');
  const [tags, setTags] = useState([]);

  // Resource specifics
  const [timeHours, setTimeHours] = useState('');
  const [timeHourlyRate, setTimeHourlyRate] = useState('');
  const [goodsItemName, setGoodsItemName] = useState('');
  const [goodsQuantity, setGoodsQuantity] = useState('');
  const [goodsUnit, setGoodsUnit] = useState('pcs');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (propAssets && propAssets.length > 0) {
      setLoadedAssets(propAssets);
      if (!assetId && !defaultAssetId) {
        setAssetId(propAssets[0].id);
      }
    } else if (projectId && open) {
      assetRepository
        .getAssets(projectId, 100, 0)
        .then(([items]) => {
          setLoadedAssets(items);
          if (!assetId && items.length > 0) {
            setAssetId(defaultAssetId || items[0].id);
          }
        })
        .catch((err) => console.error('Failed to load assets for investment modal', err));
    }
  }, [projectId, open, propAssets, defaultAssetId]);

  useEffect(() => {
    if (defaultAssetId) {
      setAssetId(defaultAssetId);
    }
  }, [defaultAssetId]);

  // Auto-calculate amount for Time
  useEffect(() => {
    if (resourceType === 'TIME' && timeHours && timeHourlyRate) {
      const h = Number(timeHours);
      const r = Number(timeHourlyRate);
      if (!isNaN(h) && !isNaN(r)) {
        setAmount((h * r).toFixed(2));
      }
    }
  }, [resourceType, timeHours, timeHourlyRate]);

  const toggleSharedInvestment = () => {
    if (!isSplit) {
      setIsSplit(true);
      if (allocations.length < 2) {
        const firstId = assetId || loadedAssets[0]?.id || '';
        const second = loadedAssets.find((a) => a.id !== firstId);
        const secondId = second ? second.id : '';
        setAllocations([
          { assetId: firstId, sharePercentage: '50' },
          { assetId: secondId, sharePercentage: '50' },
        ]);
      }
    } else {
      setIsSplit(false);
      if (allocations[0]?.assetId) {
        setAssetId(allocations[0].assetId);
      }
    }
  };

  const handleAddAllocation = () => {
    const selectedIds = new Set(allocations.map((a) => a.assetId));
    const available = loadedAssets.find((a) => !selectedIds.has(a.id));
    const nextAssetId = available ? available.id : (loadedAssets[0]?.id || '');
    setAllocations([...allocations, { assetId: nextAssetId, sharePercentage: '' }]);
  };

  const handleRemoveAllocation = (index) => {
    setAllocations(allocations.filter((_, idx) => idx !== index));
  };

  const handleAllocationChange = (index, field, value) => {
    setAllocations(allocations.map((a, idx) => (idx === index ? { ...a, [field]: value } : a)));
  };

  const handleEqualSplit = () => {
    if (allocations.length === 0) return;
    const count = allocations.length;
    const share = +(100 / count).toFixed(2);
    let remaining = 100;
    const updated = allocations.map((a, idx) => {
      if (idx === count - 1) {
        return { ...a, sharePercentage: String(+(remaining).toFixed(2)) };
      }
      remaining -= share;
      return { ...a, sharePercentage: String(share) };
    });
    setAllocations(updated);
  };

  const totalShare = isSplit ? allocations.reduce((sum, a) => sum + (Number(a.sharePercentage) || 0), 0) : 100;
  const isShareValid = !isSplit || Math.abs(totalShare - 100) < 0.01;

  const handleAdd = async (e) => {
    e.preventDefault();
    if (!amount || !date) {
      toast.error(String(content?.validationRequiredFields || 'Asset, Amount, and Date are required'));
      return;
    }

    if (isSplit) {
      if (allocations.length < 2) {
        toast.error(String(content?.sharedInvestmentMinAssets || 'Shared investment must include at least 2 assets'));
        return;
      }
      for (const a of allocations) {
        if (!a.assetId) {
          toast.error(String(content?.selectAssetPlaceholder || 'Select asset'));
          return;
        }
        if (!a.sharePercentage || Number(a.sharePercentage) <= 0) {
          toast.error(String(content?.sharesMustEqual100 || 'Total allocation shares must equal 100%'));
          return;
        }
      }
      const uniqueAssets = new Set(allocations.map((a) => a.assetId));
      if (uniqueAssets.size !== allocations.length) {
        toast.error(String(content?.duplicateAssetInAllocations || 'Duplicate asset selected in allocations'));
        return;
      }
      if (!isShareValid) {
        toast.error(String(content?.sharesMustEqual100 || 'Total allocation shares must equal 100%'));
        return;
      }
    } else if (!assetId) {
      toast.error(String(content?.validationRequiredFields || 'Asset, Amount, and Date are required'));
      return;
    }

    setLoading(true);
    try {
      await investmentRepository.addInvestment(projectId, {
        assetId: isSplit ? allocations[0].assetId : assetId,
        assetAllocations: isSplit
          ? allocations.map((a) => ({
              assetId: a.assetId,
              sharePercentage: Number(a.sharePercentage),
            }))
          : undefined,
        resourceType,
        amount: Number(amount),
        currency,
        date: new Date(date),
        description: description.trim(),
        tags,
        timeHours: resourceType === 'TIME' && timeHours ? Number(timeHours) : undefined,
        timeHourlyRate: resourceType === 'TIME' && timeHourlyRate ? Number(timeHourlyRate) : undefined,
        goodsItemName: resourceType === 'GOODS' ? goodsItemName.trim() : undefined,
        goodsQuantity: resourceType === 'GOODS' && goodsQuantity ? Number(goodsQuantity) : undefined,
        goodsUnit: resourceType === 'GOODS' ? goodsUnit.trim() : undefined,
      });

      toast.success(String(content?.paymentAddedSuccess || 'Investment logged successfully'));
      resetForm();
      onSuccess();
    } catch (e) {
      toast.error(`${String(content?.failedToAdd || 'Failed to log investment')}: ${e.message}`);
      console.error('Failed to log investment', e.message);
    } finally {
      setLoading(false);
    }
  };

  const resetForm = () => {
    setIsSplit(false);
    setAllocations([]);
    setAssetId(defaultAssetId || (loadedAssets[0]?.id || ''));
    setResourceType('MONEY');
    setAmount('');
    setCurrency('UAH');
    setDate(new Date().toISOString().split('T')[0]);
    setDescription('');
    setTags([]);
    setTimeHours('');
    setTimeHourlyRate('');
    setGoodsItemName('');
    setGoodsQuantity('');
    setGoodsUnit('pcs');
  };

  const handleCancel = () => {
    resetForm();
    onCancel();
  };

  return (
    <Dialog open={open} onOpenChange={(isOpen) => !isOpen && handleCancel()}>
      <DialogContent className="sm:max-w-[500px] rounded-2xl p-6 max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="text-xl font-bold tracking-tight">
            {String(content?.addPaymentTitle || 'Log Investment')}
          </DialogTitle>
        </DialogHeader>

        <form onSubmit={handleAdd} className="space-y-4 py-2">
          {/* Target Asset Selector */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label htmlFor="investment-asset" className="text-xs font-semibold">
                {isSplit
                  ? String(content?.allocationsBreakdown || 'Розподіл за активами')
                  : `${String(content?.assetTargetLabel || 'Цільовий актив')} *`}
              </Label>
              <Button
                type="button"
                variant="ghost"
                size="sm"
                className="h-6 text-xs text-primary px-1.5 hover:bg-primary/10"
                onClick={toggleSharedInvestment}
              >
                {isSplit
                  ? String(content?.singleAssetMode || 'Один актив (100%)')
                  : String(content?.splitAcrossAssets || 'Спільна інвестиція (кілька активів)')}
              </Button>
            </div>

            {!isSplit ? (
              <Select value={assetId} onValueChange={setAssetId}>
                <SelectTrigger id="investment-asset" className="rounded-xl">
                  <SelectValue placeholder={String(content?.selectAssetPlaceholder || 'Select asset')} />
                </SelectTrigger>
                <SelectContent className="rounded-xl max-h-56">
                  {loadedAssets.map((asset) => (
                    <SelectItem key={asset.id} value={asset.id} className="rounded-lg">
                      {asset.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            ) : (
              <div className="space-y-2 p-3 rounded-xl bg-muted/30 border border-border/60">
                {allocations.map((alloc, idx) => {
                  const shareNum = Number(alloc.sharePercentage) || 0;
                  const estimatedAllocAmount = amount ? ((Number(amount) * shareNum) / 100).toFixed(2) : null;
                  return (
                    <div key={idx} className="flex items-center gap-2">
                      <div className="flex-1">
                        <Select
                          value={alloc.assetId}
                          onValueChange={(val) => handleAllocationChange(idx, 'assetId', val)}
                        >
                          <SelectTrigger className="rounded-lg h-9 text-xs">
                            <SelectValue placeholder={String(content?.selectAssetPlaceholder || 'Select asset')} />
                          </SelectTrigger>
                          <SelectContent className="rounded-xl max-h-56">
                            {loadedAssets.map((asset) => (
                              <SelectItem key={asset.id} value={asset.id} className="rounded-lg text-xs">
                                {asset.name}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      </div>
                      <div className="w-24 relative flex items-center">
                        <Input
                          type="number"
                          step="0.1"
                          min="0"
                          max="100"
                          placeholder="%"
                          value={alloc.sharePercentage}
                          onChange={(e) => handleAllocationChange(idx, 'sharePercentage', e.target.value)}
                          className="rounded-lg h-9 text-xs pr-6"
                        />
                        <span className="absolute right-2 text-xs text-muted-foreground pointer-events-none">%</span>
                      </div>
                      {estimatedAllocAmount && (
                        <span className="text-[11px] font-mono text-muted-foreground w-16 text-right truncate">
                          {estimatedAllocAmount}
                        </span>
                      )}
                      {allocations.length > 1 && (
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon"
                          className="h-8 w-8 text-muted-foreground hover:text-destructive shrink-0"
                          onClick={() => handleRemoveAllocation(idx)}
                        >
                          <X className="h-4 w-4" />
                        </Button>
                      )}
                    </div>
                  );
                })}

                <div className="flex items-center justify-between pt-2 border-t border-border/40">
                  <div className="flex items-center gap-2">
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      className="h-7 text-xs rounded-lg gap-1"
                      onClick={handleAddAllocation}
                    >
                      <Plus className="h-3 w-3" />
                      <span>{String(content?.addAssetAllocation || 'Додати актив')}</span>
                    </Button>
                    {allocations.length > 1 && (
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        className="h-7 text-xs text-muted-foreground"
                        onClick={handleEqualSplit}
                      >
                        {String(content?.equalSplit || 'Порівну')}
                      </Button>
                    )}
                  </div>
                  <div className="flex items-center gap-1.5 text-xs font-mono">
                    <span className="text-muted-foreground">{String(content?.totalShareLabel || 'Загалом')}:</span>
                    <span className={cn('font-semibold', isShareValid ? 'text-emerald-600 dark:text-emerald-400' : 'text-amber-600 dark:text-amber-400')}>
                      {totalShare.toFixed(1)}% / 100%
                    </span>
                  </div>
                </div>
              </div>
            )}
          </div>

          {/* Resource Type & Date */}
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="resource-type" className="text-xs font-semibold">
                {String(content?.resourceTypeLabel || 'Resource Type')}
              </Label>
              <Select value={resourceType} onValueChange={setResourceType}>
                <SelectTrigger id="resource-type" className="rounded-xl">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent className="rounded-xl">
                  {resourceTypes.map((rt) => (
                    <SelectItem key={rt.key} value={rt.key} className="rounded-lg">
                      {rt.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label htmlFor="investment-date" className="text-xs font-semibold">
                {String(content?.dateLabel || 'Date')} *
              </Label>
              <Input
                id="investment-date"
                type="date"
                value={date}
                onChange={(e) => setDate(e.target.value)}
                className="rounded-xl"
                required
              />
            </div>
          </div>

          {/* Conditional Resource Fields: Labor (TIME) */}
          {resourceType === 'TIME' && (
            <div className="grid grid-cols-2 gap-4 p-3 rounded-xl bg-muted/30 border border-border/60">
              <div className="space-y-1.5">
                <Label htmlFor="time-hours" className="text-xs font-semibold">
                  {String(content?.timeHoursLabel || 'Hours Worked')}
                </Label>
                <Input
                  id="time-hours"
                  type="number"
                  step="0.25"
                  placeholder="e.g. 8.0"
                  value={timeHours}
                  onChange={(e) => setTimeHours(e.target.value)}
                  className="rounded-xl h-9 text-xs"
                />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="time-rate" className="text-xs font-semibold">
                  {String(content?.timeHourlyRateLabel || 'Hourly Rate')}
                </Label>
                <Input
                  id="time-rate"
                  type="number"
                  step="0.01"
                  placeholder="e.g. 50.00"
                  value={timeHourlyRate}
                  onChange={(e) => setTimeHourlyRate(e.target.value)}
                  className="rounded-xl h-9 text-xs"
                />
              </div>
            </div>
          )}

          {/* Conditional Resource Fields: Goods & Materials (GOODS) */}
          {resourceType === 'GOODS' && (
            <div className="space-y-3 p-3 rounded-xl bg-muted/30 border border-border/60">
              <div className="space-y-1.5">
                <Label htmlFor="goods-item-name" className="text-xs font-semibold">
                  {String(content?.goodsItemNameLabel || 'Item / Material Name')}
                </Label>
                <Input
                  id="goods-item-name"
                  placeholder="e.g. Engine Oil, Tires, Cement"
                  value={goodsItemName}
                  onChange={(e) => setGoodsItemName(e.target.value)}
                  className="rounded-xl h-9 text-xs"
                />
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-1.5">
                  <Label htmlFor="goods-quantity" className="text-xs font-semibold">
                    {String(content?.goodsQuantityLabel || 'Quantity')}
                  </Label>
                  <Input
                    id="goods-quantity"
                    type="number"
                    step="0.01"
                    placeholder="e.g. 4"
                    value={goodsQuantity}
                    onChange={(e) => setGoodsQuantity(e.target.value)}
                    className="rounded-xl h-9 text-xs"
                  />
                </div>
                <div className="space-y-1.5">
                  <Label htmlFor="goods-unit" className="text-xs font-semibold">
                    {String(content?.goodsUnitLabel || 'Unit')}
                  </Label>
                  <Input
                    id="goods-unit"
                    placeholder="e.g. pcs, kg, liters"
                    value={goodsUnit}
                    onChange={(e) => setGoodsUnit(e.target.value)}
                    className="rounded-xl h-9 text-xs"
                  />
                </div>
              </div>
            </div>
          )}

          {/* Financial Valuation Amount & Currency */}
          <div className="grid grid-cols-3 gap-4">
            <div className="col-span-2 space-y-2">
              <Label htmlFor="investment-amount" className="text-xs font-semibold">
                {String(content?.amountLabel || 'Valuation Amount')} *
              </Label>
              <Input
                id="investment-amount"
                type="number"
                step="0.01"
                placeholder={String(content?.amountPlaceholder || '0.00')}
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                className="rounded-xl"
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="investment-currency" className="text-xs font-semibold">
                {String(content?.currencyLabel || 'Currency')}
              </Label>
              <Select value={currency} onValueChange={setCurrency}>
                <SelectTrigger id="investment-currency" className="rounded-xl">
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
          </div>

          {/* Tags */}
          <TagSelector
            id="investment-tags"
            tags={tags}
            onChange={setTags}
            projectId={projectId}
          />

          {/* Description */}
          <div className="space-y-2">
            <Label htmlFor="investment-desc" className="text-xs font-semibold">
              {String(content?.descriptionLabel || 'Description')}
            </Label>
            <Textarea
              id="investment-desc"
              rows={2}
              placeholder={String(content?.descriptionPlaceholder || 'Investment description, details...')}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="rounded-xl resize-none"
            />
          </div>

          <DialogFooter className="pt-3 gap-2 sm:gap-0">
            <Button
              type="button"
              variant="outline"
              onClick={handleCancel}
              disabled={loading}
              className="rounded-xl"
            >
              {String(content?.cancelButton || 'Cancel')}
            </Button>
            <Button
              type="submit"
              disabled={loading || (isSplit ? (!isShareValid || allocations.length < 2) : !assetId) || !amount}
              className="rounded-xl font-semibold shadow-sm shadow-primary/20 gap-2"
            >
              {loading && <Loader2 className="h-4 w-4 animate-spin" />}
              <span>{String(content?.submitButton || 'Save')}</span>
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

AddPaymentModal.propTypes = {
  open: PropTypes.bool.isRequired,
  onSuccess: PropTypes.func.isRequired,
  onCancel: PropTypes.func.isRequired,
  assets: PropTypes.array,
  defaultAssetId: PropTypes.string,
};
