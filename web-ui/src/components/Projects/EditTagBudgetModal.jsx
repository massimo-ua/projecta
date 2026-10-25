import React, { useState, useEffect, useMemo } from 'react';
import PropTypes from 'prop-types';
import { useIntlayer, useLocale } from 'react-intlayer';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Loader2, Trash2 } from 'lucide-react';
import {
  TagBudgetRule,
  BUDGET_FREQUENCIES,
  BUDGET_RESOURCE_TYPES,
  cleanTagName,
  getActiveCalendarMonths,
} from '../../models';

export function EditTagBudgetModal({
  open,
  tag = '',
  existingRule = null,
  currency = 'UAH',
  startDate = null,
  endDate = null,
  onSave,
  onClose,
}) {
  const content = useIntlayer('tags');
  const { locale } = useLocale();
  const intlLocale = (locale && locale.startsWith('uk')) ? 'uk-UA' : 'en-US';

  const activeMonths = useMemo(
    () => getActiveCalendarMonths(startDate, endDate || null),
    [startDate, endDate],
  );
  const isShortPeriod = Boolean(endDate && activeMonths.size < 12);

  const [customTag, setCustomTag] = useState(tag || '');
  const [frequency, setFrequency] = useState(BUDGET_FREQUENCIES.MONTHLY);
  const [targetMonth, setTargetMonth] = useState(1);
  const [resourceType, setResourceType] = useState(BUDGET_RESOURCE_TYPES.MONEY);
  const [amount, setAmount] = useState('');
  const [timeHours, setTimeHours] = useState('');
  const [timeHourlyRate, setTimeHourlyRate] = useState('');
  const [goodsQuantity, setGoodsQuantity] = useState('');
  const [goodsUnit, setGoodsUnit] = useState('');
  const [goodsItemName, setGoodsItemName] = useState('');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    setCustomTag(tag || (existingRule ? existingRule.tag : ''));
    if (existingRule) {
      setFrequency(existingRule.frequency || BUDGET_FREQUENCIES.MONTHLY);
      setTargetMonth(existingRule.targetMonth || 1);
      setResourceType(existingRule.resourceType || BUDGET_RESOURCE_TYPES.MONEY);
      setAmount(existingRule.amount ? String(existingRule.amount) : '');
      setTimeHours(existingRule.timeHours !== null ? String(existingRule.timeHours) : '');
      setTimeHourlyRate(existingRule.timeHourlyRate !== null ? String(existingRule.timeHourlyRate) : '');
      setGoodsQuantity(existingRule.goodsQuantity !== null ? String(existingRule.goodsQuantity) : '');
      setGoodsUnit(existingRule.goodsUnit || '');
      setGoodsItemName(existingRule.goodsItemName || '');
    } else {
      setFrequency(BUDGET_FREQUENCIES.MONTHLY);
      setTargetMonth(1);
      setResourceType(BUDGET_RESOURCE_TYPES.MONEY);
      setAmount('');
      setTimeHours('');
      setTimeHourlyRate('');
      setGoodsQuantity('');
      setGoodsUnit('');
      setGoodsItemName('');
    }
  }, [existingRule, open, tag]);

  const effectiveTag = tag || customTag;

  const previewRule = useMemo(() => {
    const cleaned = cleanTagName(effectiveTag);
    let ruleAmount = parseFloat(amount) || 0;
    const hours = parseFloat(timeHours) || null;
    const rate = parseFloat(timeHourlyRate) || null;

    if (resourceType === BUDGET_RESOURCE_TYPES.TIME && hours && rate) {
      ruleAmount = Math.round(hours * rate * 100) / 100;
    }

    return new TagBudgetRule({
      tag: cleaned,
      frequency,
      targetMonth,
      resourceType,
      amount: ruleAmount,
      timeHours: hours,
      timeHourlyRate: rate,
      goodsQuantity: parseFloat(goodsQuantity) || null,
      goodsUnit,
      goodsItemName,
    });
  }, [
    effectiveTag,
    frequency,
    targetMonth,
    resourceType,
    amount,
    timeHours,
    timeHourlyRate,
    goodsQuantity,
    goodsUnit,
    goodsItemName,
  ]);

  const handleSave = async (e) => {
    e.preventDefault();
    if (!previewRule.isValid) return;

    setLoading(true);
    try {
      await onSave(previewRule);
      onClose();
    } finally {
      setLoading(false);
    }
  };

  const handleRemove = async () => {
    setLoading(true);
    try {
      await onSave(null);
      onClose();
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={(isOpen) => !isOpen && onClose()}>
      <DialogContent className="sm:max-w-[500px] rounded-2xl p-6">
        <DialogHeader>
          <DialogTitle className="text-lg font-bold tracking-tight">
            <span>{tag ? `${String(content.budgetModalTitle || 'Configure Tag Budget')}: #${tag}` : String(content.budgetModalTitle || 'Configure Tag Budget')}</span>
          </DialogTitle>
          <DialogDescription className="text-xs text-muted-foreground">
            {String(content.budgetModalSubtitle || 'Set expected frequency and resource allocation')}
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSave} className="space-y-4 py-2">
          {!tag && (
            <div className="space-y-1.5">
              <Label htmlFor="tag-edit-custom-tag" className="text-xs font-semibold">
                {String(content.tagNameLabel || 'Category / Tag Name')}
              </Label>
              <Input
                id="tag-edit-custom-tag"
                placeholder={String(content.tagPlaceholder || 'materials, equipment, etc.')}
                value={customTag}
                onChange={(e) => setCustomTag(e.target.value)}
                className="h-9 rounded-xl text-xs"
                required
              />
            </div>
          )}

          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1.5">
              <Label htmlFor="tag-edit-frequency" className="text-xs font-semibold">
                {String(content.frequencyLabel || 'Frequency')}
              </Label>
              <Select value={frequency} onValueChange={setFrequency}>
                <SelectTrigger id="tag-edit-frequency" className="h-9 rounded-xl text-xs">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent className="rounded-xl">
                  {Object.keys(BUDGET_FREQUENCIES).map((f) => (
                    <SelectItem key={f} value={f} className="text-xs">
                      {String(content.frequencies?.[f] || f)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="tag-edit-resource-type" className="text-xs font-semibold">
                {String(content.resourceTypeLabel || 'Resource Type')}
              </Label>
              <Select value={resourceType} onValueChange={setResourceType}>
                <SelectTrigger id="tag-edit-resource-type" className="h-9 rounded-xl text-xs">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent className="rounded-xl">
                  {Object.keys(BUDGET_RESOURCE_TYPES).map((r) => (
                    <SelectItem key={r} value={r} className="text-xs">
                      {String(content.resources?.[r] || r)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          {(frequency === BUDGET_FREQUENCIES.ONCE || frequency === BUDGET_FREQUENCIES.ANNUALLY) && (
            <div className="space-y-1.5">
              <Label htmlFor="tag-edit-target-month" className="text-xs font-semibold">
                {String(content.targetMonthLabel || 'Scheduled Month (1–12)')}
              </Label>
              <Select
                value={String(targetMonth)}
                onValueChange={(val) => setTargetMonth(parseInt(val, 10) || 1)}
              >
                <SelectTrigger id="tag-edit-target-month" className="h-9 rounded-xl text-xs">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent className="rounded-xl">
                  {Array.from({ length: 12 }, (_, i) => i + 1).map((m) => {
                    const isOutside = activeMonths && activeMonths.size > 0 && !activeMonths.has(m);
                    return (
                      <SelectItem key={m} value={String(m)} className="text-xs">
                        <span>
                          {`${m} - ${String(content.months?.[m] || `${content.monthLabel || 'Month'} ${m}`)}`}
                          {isOutside ? ` (${String(content.outsidePeriodBadge || 'outside period')})` : ''}
                        </span>
                      </SelectItem>
                    );
                  })}
                </SelectContent>
              </Select>
              {activeMonths && activeMonths.size > 0 && !activeMonths.has(targetMonth) && (
                <p className="text-[11px] text-amber-600 dark:text-amber-400 font-medium pt-0.5">
                  {String(content.monthOutsidePeriodWarning || 'This month is outside the active project period (planned total will be 0)')}
                </p>
              )}
            </div>
          )}

          {resourceType === BUDGET_RESOURCE_TYPES.MONEY && (
            <div className="space-y-1.5">
              <Label htmlFor="tag-edit-amount" className="text-xs font-semibold">
                <span>{`${String(content.amountLabel || 'Expected Amount')} (${currency})`}</span>
              </Label>
              <Input
                id="tag-edit-amount"
                type="number"
                min="0"
                step="any"
                placeholder="0.00"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                className="h-9 rounded-xl text-xs"
                required
              />
            </div>
          )}

          {resourceType === BUDGET_RESOURCE_TYPES.TIME && (
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label htmlFor="tag-edit-hours" className="text-xs font-semibold">
                  {String(content.hoursLabel || 'Expected Hours')}
                </Label>
                <Input
                  id="tag-edit-hours"
                  type="number"
                  min="0"
                  step="any"
                  placeholder="20"
                  value={timeHours}
                  onChange={(e) => setTimeHours(e.target.value)}
                  className="h-9 rounded-xl text-xs"
                  required
                />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="tag-edit-rate" className="text-xs font-semibold">
                  <span>{`${String(content.hourlyRateLabel || 'Hourly Rate')} (${currency}${String(content.perHourSuffix || '/h')})`}</span>
                </Label>
                <Input
                  id="tag-edit-rate"
                  type="number"
                  min="0"
                  step="any"
                  placeholder="25.00"
                  value={timeHourlyRate}
                  onChange={(e) => setTimeHourlyRate(e.target.value)}
                  className="h-9 rounded-xl text-xs"
                  required
                />
              </div>
            </div>
          )}

          {resourceType === BUDGET_RESOURCE_TYPES.GOODS && (
            <div className="space-y-3">
              <div className="grid grid-cols-4 gap-3">
                <div className="space-y-1.5 col-span-2">
                  <Label htmlFor="tag-edit-goods-name" className="text-xs font-semibold">
                    {String(content.goodsNameLabel || 'Item Name')}
                  </Label>
                  <Input
                    id="tag-edit-goods-name"
                    placeholder={String(content.goodsItemNamePlaceholder || 'Material / goods')}
                    value={goodsItemName}
                    onChange={(e) => setGoodsItemName(e.target.value)}
                    className="h-9 rounded-xl text-xs"
                  />
                </div>
                <div className="space-y-1.5">
                  <Label htmlFor="tag-edit-goods-qty" className="text-xs font-semibold">
                    {String(content.goodsQtyLabel || 'Quantity')}
                  </Label>
                  <Input
                    id="tag-edit-goods-qty"
                    type="number"
                    min="0"
                    step="any"
                    placeholder="1"
                    value={goodsQuantity}
                    onChange={(e) => setGoodsQuantity(e.target.value)}
                    className="h-9 rounded-xl text-xs"
                  />
                </div>
                <div className="space-y-1.5">
                  <Label htmlFor="tag-edit-goods-unit" className="text-xs font-semibold">
                    {String(content.goodsUnitLabel || 'Unit')}
                  </Label>
                  <Input
                    id="tag-edit-goods-unit"
                    placeholder={String(content.goodsUnitPlaceholder || 'pcs, kg, etc.')}
                    value={goodsUnit}
                    onChange={(e) => setGoodsUnit(e.target.value)}
                    className="h-9 rounded-xl text-xs"
                  />
                </div>
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="tag-edit-goods-amount" className="text-xs font-semibold">
                  <span>{`${String(content.amountLabel || 'Valuation Amount')} (${currency})`}</span>
                </Label>
                <Input
                  id="tag-edit-goods-amount"
                  type="number"
                  min="0"
                  step="any"
                  placeholder="0.00"
                  value={amount}
                  onChange={(e) => setAmount(e.target.value)}
                  className="h-9 rounded-xl text-xs"
                  required
                />
              </div>
            </div>
          )}

          {/* Projection Preview Banner */}
          {previewRule.isValid && (
            <div className="rounded-xl border border-primary/20 bg-primary/5 p-3 flex items-center justify-between text-xs">
              <span className="text-muted-foreground">
                {`${isShortPeriod
                  ? String(content.plannedPeriodTitle || 'Planned for Period')
                  : String(content.plannedAnnualTitle || 'Planned Annual')}:`}
              </span>
              <span className="font-extrabold text-primary">
                <span>
                  {`${new Intl.NumberFormat(intlLocale).format(previewRule.getAnnualTotal(startDate, endDate))} ${currency}${isShortPeriod ? '' : String(content.perYearSuffix || '/yr')}`}
                </span>
              </span>
            </div>
          )}

          <DialogFooter className="pt-2 flex items-center justify-between sm:justify-between w-full">
            {existingRule ? (
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={handleRemove}
                disabled={loading}
                className="rounded-xl h-9 text-xs text-rose-500 hover:text-rose-600 hover:bg-rose-500/10 gap-1.5"
              >
                <Trash2 className="h-3.5 w-3.5" />
                <span>{String(content.removeBudgetButton || 'Remove Budget')}</span>
              </Button>
            ) : (
              <div />
            )}

            <div className="flex items-center gap-2">
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={onClose}
                disabled={loading}
                className="rounded-xl h-9 text-xs"
              >
                {String(content.cancelButton || 'Cancel')}
              </Button>
              <Button
                type="submit"
                size="sm"
                disabled={loading || !previewRule.isValid}
                className="rounded-xl h-9 text-xs font-semibold gap-1.5"
              >
                {loading && <Loader2 className="h-3.5 w-3.5 animate-spin" />}
                <span>{String(content.saveBudgetButton || 'Save Budget Rule')}</span>
              </Button>
            </div>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

EditTagBudgetModal.propTypes = {
  open: PropTypes.bool.isRequired,
  tag: PropTypes.string,
  existingRule: PropTypes.shape({
    tag: PropTypes.string,
    frequency: PropTypes.string,
    targetMonth: PropTypes.number,
    resourceType: PropTypes.string,
    amount: PropTypes.number,
    timeHours: PropTypes.number,
    timeHourlyRate: PropTypes.number,
    goodsQuantity: PropTypes.number,
    goodsUnit: PropTypes.string,
    goodsItemName: PropTypes.string,
  }),
  currency: PropTypes.string,
  startDate: PropTypes.string,
  endDate: PropTypes.string,
  onSave: PropTypes.func.isRequired,
  onClose: PropTypes.func.isRequired,
};

EditTagBudgetModal.defaultProps = {
  tag: '',
  existingRule: null,
  currency: 'UAH',
  startDate: null,
  endDate: null,
};

export default EditTagBudgetModal;
