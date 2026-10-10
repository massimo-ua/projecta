import React, { useEffect, useState } from 'react';
import PropTypes from 'prop-types';
import { useParams } from 'react-router-dom';
import { useIntlayer } from 'react-intlayer';
import { investmentRepository, assetRepository } from '../../api';
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

const SUPPORTED_CURRENCIES = ['UAH', 'USD', 'EUR', 'PLN'];
const RESOURCE_TYPES = [
  { key: 'MONEY', label: 'Capital (Money)' },
  { key: 'TIME', label: 'Labor (Time)' },
  { key: 'GOODS', label: 'Goods & Materials' },
];

export default function EditPaymentModal({
  onSuccess,
  onCancel,
  paymentId = null,
  assets: propAssets,
}) {
  const content = useIntlayer('payments');
  const { projectId } = useParams();

  const [loadedAssets, setLoadedAssets] = useState([]);
  const [assetId, setAssetId] = useState('');
  const [resourceType, setResourceType] = useState('MONEY');
  const [amount, setAmount] = useState('');
  const [currency, setCurrency] = useState('UAH');
  const [date, setDate] = useState('');
  const [description, setDescription] = useState('');
  const [tags, setTags] = useState([]);
  const [tagInput, setTagInput] = useState('');

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
    } else if (projectId) {
      assetRepository
        .getAssets(projectId, 100, 0)
        .then(([items]) => setLoadedAssets(items))
        .catch((err) => console.error('Failed to load assets for edit modal', err));
    }
  }, [projectId, propAssets]);

  useEffect(() => {
    if (!paymentId) return;

    investmentRepository
      .getInvestment(projectId, paymentId)
      .then((inv) => {
        setAssetId(inv.assetId || '');
        setResourceType(inv.resourceType || 'MONEY');
        setAmount(inv.amount || '');
        setCurrency(inv.currency || 'UAH');
        setDescription(inv.description || '');
        setTags(Array.isArray(inv.tags) ? inv.tags : []);
        setTimeHours(inv.timeHours || '');
        setTimeHourlyRate(inv.timeHourlyRate || '');
        setGoodsItemName(inv.goodsItemName || '');
        setGoodsQuantity(inv.goodsQuantity || '');
        setGoodsUnit(inv.goodsUnit || 'pcs');

        if (inv.date) {
          const dt = new Date(inv.date);
          if (!isNaN(dt.getTime())) {
            setDate(dt.toISOString().split('T')[0]);
          }
        }
      })
      .catch((err) => {
        toast.error(`${String(content?.failedToLoadDetails || 'Failed to load details')}: ${err.message}`);
      });
  }, [paymentId, projectId, content]);

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

  const handleAddTag = () => {
    const trimmed = tagInput.trim().replace(/^#/, '');
    if (trimmed && !tags.includes(trimmed)) {
      setTags([...tags, trimmed]);
      setTagInput('');
    }
  };

  const handleRemoveTag = (tagToRemove) => {
    setTags(tags.filter((t) => t !== tagToRemove));
  };

  const handleTagKeyDown = (e) => {
    if (e.key === 'Enter' || e.key === ',') {
      e.preventDefault();
      handleAddTag();
    }
  };

  const handleUpdate = async (e) => {
    e.preventDefault();
    if (!assetId || !amount || !date) {
      toast.error(String(content?.validationRequiredFields || 'Asset, Amount, and Date are required'));
      return;
    }

    setLoading(true);
    try {
      await investmentRepository.updateInvestment(projectId, {
        id: paymentId,
        assetId,
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

      toast.success(String(content?.paymentUpdatedSuccess || 'Investment updated successfully'));
      onSuccess();
    } catch (e) {
      toast.error(`${String(content?.failedToUpdate || 'Failed to update investment')}: ${e.message}`);
      console.error('Failed to update investment', e.message);
    } finally {
      setLoading(false);
    }
  };

  const handleCancel = () => onCancel();

  return (
    <Dialog open={Boolean(paymentId)} onOpenChange={(isOpen) => !isOpen && handleCancel()}>
      <DialogContent className="sm:max-w-[500px] rounded-2xl p-6 max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="text-xl font-bold tracking-tight">
            {String(content?.editPaymentTitle || 'Edit Investment')}
          </DialogTitle>
        </DialogHeader>

        <form onSubmit={handleUpdate} className="space-y-4 py-2">
          {/* Target Asset Selector */}
          <div className="space-y-2">
            <Label htmlFor="edit-investment-asset" className="text-xs font-semibold">
              {String(content?.assetTargetLabel || 'Target Asset')} *
            </Label>
            <Select value={assetId} onValueChange={setAssetId}>
              <SelectTrigger id="edit-investment-asset" className="rounded-xl">
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
          </div>

          {/* Resource Type & Date */}
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="edit-resource-type" className="text-xs font-semibold">
                {String(content?.resourceTypeLabel || 'Resource Type')}
              </Label>
              <Select value={resourceType} onValueChange={setResourceType}>
                <SelectTrigger id="edit-resource-type" className="rounded-xl">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent className="rounded-xl">
                  {RESOURCE_TYPES.map((rt) => (
                    <SelectItem key={rt.key} value={rt.key} className="rounded-lg">
                      {rt.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label htmlFor="edit-investment-date" className="text-xs font-semibold">
                {String(content?.dateLabel || 'Date')} *
              </Label>
              <Input
                id="edit-investment-date"
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
                <Label htmlFor="edit-time-hours" className="text-xs font-semibold">
                  {String(content?.timeHoursLabel || 'Hours Worked')}
                </Label>
                <Input
                  id="edit-time-hours"
                  type="number"
                  step="0.25"
                  placeholder="e.g. 8.0"
                  value={timeHours}
                  onChange={(e) => setTimeHours(e.target.value)}
                  className="rounded-xl h-9 text-xs"
                />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="edit-time-rate" className="text-xs font-semibold">
                  {String(content?.timeHourlyRateLabel || 'Hourly Rate')}
                </Label>
                <Input
                  id="edit-time-rate"
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
                <Label htmlFor="edit-goods-item-name" className="text-xs font-semibold">
                  {String(content?.goodsItemNameLabel || 'Item / Material Name')}
                </Label>
                <Input
                  id="edit-goods-item-name"
                  placeholder="e.g. Engine Oil, Tires, Cement"
                  value={goodsItemName}
                  onChange={(e) => setGoodsItemName(e.target.value)}
                  className="rounded-xl h-9 text-xs"
                />
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-1.5">
                  <Label htmlFor="edit-goods-quantity" className="text-xs font-semibold">
                    {String(content?.goodsQuantityLabel || 'Quantity')}
                  </Label>
                  <Input
                    id="edit-goods-quantity"
                    type="number"
                    step="0.01"
                    placeholder="e.g. 4"
                    value={goodsQuantity}
                    onChange={(e) => setGoodsQuantity(e.target.value)}
                    className="rounded-xl h-9 text-xs"
                  />
                </div>
                <div className="space-y-1.5">
                  <Label htmlFor="edit-goods-unit" className="text-xs font-semibold">
                    {String(content?.goodsUnitLabel || 'Unit')}
                  </Label>
                  <Input
                    id="edit-goods-unit"
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
              <Label htmlFor="edit-investment-amount" className="text-xs font-semibold">
                {String(content?.amountLabel || 'Valuation Amount')} *
              </Label>
              <Input
                id="edit-investment-amount"
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
              <Label htmlFor="edit-investment-currency" className="text-xs font-semibold">
                {String(content?.currencyLabel || 'Currency')}
              </Label>
              <Select value={currency} onValueChange={setCurrency}>
                <SelectTrigger id="edit-investment-currency" className="rounded-xl">
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

          {/* Tags Input */}
          <div className="space-y-2">
            <Label htmlFor="edit-investment-tags" className="text-xs font-semibold">
              {String(content?.tagsLabel || 'Tags')}
            </Label>
            <div className="flex gap-2">
              <Input
                id="edit-investment-tags"
                placeholder={String(content?.tagsPlaceholder || 'e.g. labor, parts, down-payment')}
                value={tagInput}
                onChange={(e) => setTagInput(e.target.value)}
                onKeyDown={handleTagKeyDown}
                className="rounded-xl"
              />
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={handleAddTag}
                className="rounded-xl shrink-0 px-3"
              >
                <Plus className="h-4 w-4" />
              </Button>
            </div>
            {tags.length > 0 && (
              <div className="flex gap-1.5 flex-wrap pt-1">
                {tags.map((tag) => (
                  <Badge key={tag} variant="secondary" className="gap-1 rounded-md text-xs py-0.5">
                    <span>#{tag}</span>
                    <button
                      type="button"
                      onClick={() => handleRemoveTag(tag)}
                      className="text-muted-foreground hover:text-foreground"
                    >
                      <X className="h-3 w-3" />
                    </button>
                  </Badge>
                ))}
              </div>
            )}
          </div>

          {/* Description */}
          <div className="space-y-2">
            <Label htmlFor="edit-investment-desc" className="text-xs font-semibold">
              {String(content?.descriptionLabel || 'Description')}
            </Label>
            <Textarea
              id="edit-investment-desc"
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
              disabled={loading || !assetId || !amount}
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

EditPaymentModal.propTypes = {
  paymentId: PropTypes.string,
  onSuccess: PropTypes.func.isRequired,
  onCancel: PropTypes.func.isRequired,
  assets: PropTypes.array,
};
