import React from 'react';
import PropTypes from 'prop-types';
import { useIntlayer } from 'react-intlayer';
import {
  ResponsiveContainer,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  CartesianGrid,
} from 'recharts';
import {
  Plus,
  Trash2,
  Edit2,
  Check,
  ChevronRight,
  ChevronLeft,
  Calendar,
  Sparkles,
  Loader2,
  Info,
  X,
} from 'lucide-react';
import { toast } from 'sonner';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog';
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
import { useProjectSetupWizard } from '../../hooks/useProjectSetupWizard';
import {
  BUDGET_FREQUENCIES,
  BUDGET_RESOURCE_TYPES,
  PRESET_PROJECT_TAGS,
} from '../../models';

const CURRENCIES = [
  { code: 'UAH', label: 'UAH ₴' },
  { code: 'USD', label: 'USD $' },
  { code: 'EUR', label: 'EUR €' },
  { code: 'PLN', label: 'PLN zł' },
];

export function ProjectSetupWizardModal({ open, onSuccess, onCancel }) {
  const content = useIntlayer('projects');
  const w = content.wizard || {};

  const {
    step,
    name,
    setName,
    description,
    setDescription,
    mainCurrency,
    setMainCurrency,
    startDate,
    setStartDate,
    endDate,
    setEndDate,
    tagBudgets,
    activeMonths,
    draftRule,
    editingTag,
    submitting,
    schedulePreview,
    canGoNext,
    isDraftValid,
    nextStep,
    prevStep,
    updateDraft,
    resetDraft,
    selectPresetTag,
    startEditTag,
    addOrUpdateTagRule,
    removeTagRule,
    submitWizard,
    resetWizard,
  } = useProjectSetupWizard();

  const handleClose = () => {
    resetWizard();
    onCancel();
  };

  const handleFinish = async () => {
    try {
      await submitWizard((createdProject) => {
        toast.success(String(content.projectCreatedSuccess));
        resetWizard();
        onSuccess(createdProject);
      });
    } catch (err) {
      toast.error(`${String(content.failedToCreateProject)}: ${err.message || ''}`);
    }
  };

  const getStepClass = (num) => {
    if (step === num) {
      return 'bg-primary/10 text-primary border-primary/30 shadow-xs';
    }
    if (step > num) {
      return 'bg-muted/40 text-muted-foreground border-border/40';
    }
    return 'bg-transparent text-muted-foreground/60 border-border/20';
  };

  const getStepBadgeClass = (num) => {
    if (step >= num) {
      return 'bg-primary text-primary-foreground';
    }
    return 'bg-muted text-muted-foreground';
  };

  return (
    <Dialog open={open} onOpenChange={(isOpen) => !isOpen && handleClose()}>
      <DialogContent className="sm:max-w-[760px] max-h-[92vh] overflow-y-auto rounded-2xl p-6">
        <DialogHeader className="border-b border-border/50 pb-4">
          <div className="flex items-center gap-2">
            <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-primary/10 text-primary">
              <Sparkles className="h-5 w-5" />
            </div>
            <div>
              <DialogTitle className="text-xl font-bold tracking-tight">
                {String(w.title || 'Project Setup Wizard')}
              </DialogTitle>
              <DialogDescription className="text-xs text-muted-foreground mt-0.5">
                {String(w.subtitle || '')}
              </DialogDescription>
            </div>
          </div>

          {/* Stepper Indicator */}
          <div className="grid grid-cols-3 gap-2 pt-3">
            {[
              { num: 1, title: String(w.steps?.basicsTitle || '1. Project & Currency') },
              { num: 2, title: String(w.steps?.categoriesTitle || '2. Tags & Budget') },
              { num: 3, title: String(w.steps?.scheduleTitle || '3. Annual Schedule') },
            ].map((s) => (
              <div
                key={s.num}
                className={`flex items-center gap-2 rounded-xl px-3 py-2 text-xs font-semibold border transition-all ${getStepClass(s.num)}`}
              >
                <div
                  className={`flex h-5 w-5 items-center justify-center rounded-full text-[10px] font-bold ${getStepBadgeClass(s.num)}`}
                >
                  {step > s.num ? <Check className="h-3 w-3" /> : s.num}
                </div>
                <span className="truncate">{s.title}</span>
              </div>
            ))}
          </div>
        </DialogHeader>

        {/* STEP 1: BASICS & CURRENCY */}
        {step === 1 && (
          <div className="space-y-4 py-2">
            <div className="space-y-2">
              <Label htmlFor="wizard-project-name" className="text-xs font-semibold">
                {String(content.projectNameLabel)}
                <span className="text-rose-500 ml-0.5">*</span>
              </Label>
              <Input
                id="wizard-project-name"
                placeholder={String(content.projectNamePlaceholder)}
                value={name}
                onChange={(e) => setName(e.target.value)}
                className="rounded-xl h-10"
                required
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="wizard-project-desc" className="text-xs font-semibold">
                {String(content.descriptionLabel)}
              </Label>
              <Textarea
                id="wizard-project-desc"
                rows={3}
                placeholder={String(content.descriptionPlaceholder)}
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                className="rounded-xl resize-none"
              />
            </div>

            <div className="space-y-4 pt-1">
              <div className="space-y-2">
                <Label htmlFor="wizard-currency" className="text-xs font-semibold">
                  {String(w.homeCurrencyLabel || 'Home Currency')}
                  <span className="text-rose-500 ml-0.5">*</span>
                </Label>
                <Select value={mainCurrency} onValueChange={setMainCurrency}>
                  <SelectTrigger id="wizard-currency" className="rounded-xl h-10">
                    <SelectValue placeholder={String(w.currencyPlaceholder || 'Currency')} />
                  </SelectTrigger>
                  <SelectContent className="rounded-xl">
                    {CURRENCIES.map((c) => (
                      <SelectItem key={c.code} value={c.code}>
                        {c.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <p className="text-[11px] text-muted-foreground leading-snug">
                  {String(w.homeCurrencyHint || '')}
                </p>
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label
                    htmlFor="wizard-start-date"
                    className="text-xs font-semibold flex items-center gap-1.5"
                  >
                    <Calendar className="h-3.5 w-3.5 text-muted-foreground" />
                    <span>{String(w.startDateLabel || 'Project Start Date')}</span>
                  </Label>
                  <Input
                    id="wizard-start-date"
                    type="date"
                    value={startDate}
                    onChange={(e) => setStartDate(e.target.value)}
                    className="rounded-xl h-10"
                  />
                </div>

                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <Label
                      htmlFor="wizard-end-date"
                      className="text-xs font-semibold flex items-center gap-1.5"
                    >
                      <Calendar className="h-3.5 w-3.5 text-muted-foreground" />
                      <span>{String(w.endDateLabel || 'Target End Date')}</span>
                    </Label>
                    <span className="text-[10px] text-muted-foreground font-medium">
                      {String(w.optionalBadge || 'Optional')}
                    </span>
                  </div>
                  <div className="relative flex items-center">
                    <Input
                      id="wizard-end-date"
                      type="date"
                      value={endDate}
                      min={startDate || undefined}
                      onChange={(e) => setEndDate(e.target.value)}
                      className="rounded-xl h-10 pr-9"
                    />
                    {endDate && (
                      <button
                        type="button"
                        onClick={() => setEndDate('')}
                        title={String(w.clearEndDateButton || 'Clear end date')}
                        className="absolute right-2.5 p-1 rounded-md text-muted-foreground hover:text-foreground hover:bg-muted/80 transition-colors"
                      >
                        <X className="h-3.5 w-3.5" />
                      </button>
                    )}
                  </div>
                </div>
              </div>

              {endDate && startDate && endDate < startDate ? (
                <p className="text-[11px] text-rose-500 font-medium">
                  {String(w.invalidEndDateError || 'End date cannot be earlier than start date.')}
                </p>
              ) : (
                <p className="text-[11px] text-muted-foreground leading-snug">
                  {endDate
                    ? String(w.endDateSetHint || 'Goal target date active. Investments and budget projections will conclude at this date.')
                    : String(w.endDateHint || 'Leave empty if this project is ongoing without a deadline (e.g. "Stay healthy"). If set, projections are capped at this goal date.')}
                </p>
              )}
            </div>
          </div>
        )}

        {/* STEP 2: INVESTMENT TAGS & BUDGET (SKIPPABLE 0+ TAGS) */}
        {step === 2 && (
          <div className="space-y-5 py-2">
            {/* Quick Add Presets */}
            <div className="space-y-2">
              <span className="text-xs font-semibold text-muted-foreground">
                {String(w.presetTagsTitle || 'Quick-Add Suggested Categories (Tags)')}
              </span>
              <div className="flex flex-wrap gap-1.5">
                {PRESET_PROJECT_TAGS.map((tag) => (
                  <Button
                    key={tag}
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() => selectPresetTag(tag)}
                    className="h-7 text-xs rounded-lg px-2.5 bg-muted/30 hover:bg-primary/10 hover:text-primary transition-all"
                  >
                    <span>{`#${tag}`}</span>
                  </Button>
                ))}
              </div>
            </div>

            {/* Tag Rule Form Card */}
            <div className="rounded-2xl border border-border/80 bg-muted/10 p-4 space-y-3.5">
              <div className="flex items-center justify-between">
                <h4 className="text-xs font-bold uppercase tracking-wider text-foreground">
                  {String(w.tagRuleTitle || 'Define Category Investment Expectation')}
                </h4>
                {editingTag && (
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    onClick={resetDraft}
                    className="h-6 text-xs text-muted-foreground"
                  >
                    {String(w.cancelEditCategoryButton || 'Cancel Edit')}
                  </Button>
                )}
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                <div className="space-y-1.5">
                  <Label htmlFor="tag-rule-name" className="text-xs font-medium">
                    {String(w.tagLabel || 'Category (Tag)')}
                  </Label>
                  <Input
                    id="tag-rule-name"
                    placeholder={String(w.tagPlaceholder || 'labor')}
                    value={draftRule.tag}
                    onChange={(e) => updateDraft({ tag: e.target.value })}
                    className="h-9 rounded-xl text-xs"
                  />
                </div>

                <div className="space-y-1.5">
                  <Label htmlFor="tag-rule-frequency" className="text-xs font-medium">
                    {String(w.frequencyLabel || 'Frequency')}
                  </Label>
                  <Select
                    value={draftRule.frequency}
                    onValueChange={(val) => updateDraft({ frequency: val })}
                  >
                    <SelectTrigger id="tag-rule-frequency" className="h-9 rounded-xl text-xs">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent className="rounded-xl">
                      {Object.keys(BUDGET_FREQUENCIES).map((f) => (
                        <SelectItem key={f} value={f} className="text-xs">
                          {String(w.frequencies?.[f] || f)}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>

                {(draftRule.frequency === BUDGET_FREQUENCIES.ONCE
                  || draftRule.frequency === BUDGET_FREQUENCIES.ANNUALLY) && (
                  <div className="space-y-1.5">
                    <Label htmlFor="tag-rule-target-month" className="text-xs font-medium">
                      {String(w.targetMonthLabel || 'Scheduled Month (1–12)')}
                    </Label>
                    <Select
                      value={String(draftRule.targetMonth || 1)}
                      onValueChange={(val) => updateDraft({ targetMonth: parseInt(val, 10) || 1 })}
                    >
                      <SelectTrigger id="tag-rule-target-month" className="h-9 rounded-xl text-xs">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent className="rounded-xl">
                        {Array.from({ length: 12 }, (_, i) => i + 1).map((m) => {
                          const isOutside = activeMonths
                            && activeMonths.size > 0
                            && !activeMonths.has(m);
                          const outsideText = isOutside
                            ? ` (${String(w.outsidePeriodBadge || 'outside period')})`
                            : '';
                          return (
                            <SelectItem key={m} value={String(m)} className="text-xs">
                              <span>
                                {`${m} - ${String(w.months?.[m] || `${w.monthLabel || 'Month'} ${m}`)}`}
                                {outsideText}
                              </span>
                            </SelectItem>
                          );
                        })}
                      </SelectContent>
                    </Select>
                    {activeMonths && activeMonths.size > 0
                      && !activeMonths.has(draftRule.targetMonth || 1) && (
                      <p className="text-[11px] text-amber-600 dark:text-amber-400 font-medium pt-0.5">
                        {String(w.monthOutsidePeriodWarning
                          || 'This month is outside the active project period (planned total will be 0)')}
                      </p>
                    )}
                  </div>
                )}

                <div className="space-y-1.5">
                  <Label htmlFor="tag-rule-resource-type" className="text-xs font-medium">
                    {String(w.resourceTypeLabel || 'Resource Type')}
                  </Label>
                  <Select
                    value={draftRule.resourceType}
                    onValueChange={(val) => updateDraft({ resourceType: val })}
                  >
                    <SelectTrigger id="tag-rule-resource-type" className="h-9 rounded-xl text-xs">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent className="rounded-xl">
                      {Object.keys(BUDGET_RESOURCE_TYPES).map((r) => (
                        <SelectItem key={r} value={r} className="text-xs">
                          {String(w.resources?.[r] || r)}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              </div>

              {/* Dynamic Resource Inputs */}
              {draftRule.resourceType === BUDGET_RESOURCE_TYPES.MONEY && (
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-1">
                  <div className="space-y-1.5">
                    <Label htmlFor="tag-rule-amount" className="text-xs font-medium">
                      <span>{`${String(w.amountLabel || 'Expected Amount')} (${mainCurrency})`}</span>
                    </Label>
                    <Input
                      id="tag-rule-amount"
                      type="number"
                      min="0"
                      step="any"
                      placeholder="0.00"
                      value={draftRule.amount}
                      onChange={(e) => updateDraft({ amount: e.target.value })}
                      className="h-9 rounded-xl text-xs"
                    />
                  </div>
                </div>
              )}

              {draftRule.resourceType === BUDGET_RESOURCE_TYPES.TIME && (
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-1">
                  <div className="space-y-1.5">
                    <Label htmlFor="tag-rule-hours" className="text-xs font-medium">
                      {String(w.hoursLabel || 'Expected Hours')}
                    </Label>
                    <Input
                      id="tag-rule-hours"
                      type="number"
                      min="0"
                      step="any"
                      placeholder="40"
                      value={draftRule.timeHours}
                      onChange={(e) => updateDraft({ timeHours: e.target.value })}
                      className="h-9 rounded-xl text-xs"
                    />
                  </div>
                  <div className="space-y-1.5">
                    <Label htmlFor="tag-rule-hourly-rate" className="text-xs font-medium">
                      <span>{`${String(w.hourlyRateLabel || 'Hourly Rate')} (${mainCurrency}${String(w.perHourSuffix || '/h')})`}</span>
                    </Label>
                    <Input
                      id="tag-rule-hourly-rate"
                      type="number"
                      min="0"
                      step="any"
                      placeholder="25.00"
                      value={draftRule.timeHourlyRate}
                      onChange={(e) => updateDraft({ timeHourlyRate: e.target.value })}
                      className="h-9 rounded-xl text-xs"
                    />
                  </div>
                </div>
              )}

              {draftRule.resourceType === BUDGET_RESOURCE_TYPES.GOODS && (
                <div className="grid grid-cols-1 sm:grid-cols-4 gap-3 pt-1">
                  <div className="space-y-1.5 sm:col-span-2">
                    <Label htmlFor="tag-rule-goods-name" className="text-xs font-medium">
                      {String(w.goodsNameLabel || 'Item / Material Name')}
                    </Label>
                    <Input
                      id="tag-rule-goods-name"
                      placeholder={String(w.goodsItemPlaceholder || 'Concrete, cables...')}
                      value={draftRule.goodsItemName}
                      onChange={(e) => updateDraft({ goodsItemName: e.target.value })}
                      className="h-9 rounded-xl text-xs"
                    />
                  </div>
                  <div className="space-y-1.5">
                    <Label htmlFor="tag-rule-goods-qty" className="text-xs font-medium">
                      {String(w.goodsQtyLabel || 'Quantity')}
                    </Label>
                    <Input
                      id="tag-rule-goods-qty"
                      type="number"
                      min="0"
                      step="any"
                      placeholder="10"
                      value={draftRule.goodsQuantity}
                      onChange={(e) => updateDraft({ goodsQuantity: e.target.value })}
                      className="h-9 rounded-xl text-xs"
                    />
                  </div>
                  <div className="space-y-1.5">
                    <Label htmlFor="tag-rule-goods-unit" className="text-xs font-medium">
                      {String(w.goodsUnitLabel || 'Unit')}
                    </Label>
                    <Input
                      id="tag-rule-goods-unit"
                      placeholder={String(w.goodsUnitPlaceholder || 'pcs, kg...')}
                      value={draftRule.goodsUnit}
                      onChange={(e) => updateDraft({ goodsUnit: e.target.value })}
                      className="h-9 rounded-xl text-xs"
                    />
                  </div>
                  <div className="space-y-1.5 sm:col-span-2">
                    <Label htmlFor="tag-rule-goods-amount" className="text-xs font-medium">
                      <span>{`${String(w.amountLabel || 'Valuation Amount')} (${mainCurrency})`}</span>
                    </Label>
                    <Input
                      id="tag-rule-goods-amount"
                      type="number"
                      min="0"
                      step="any"
                      placeholder="0.00"
                      value={draftRule.amount}
                      onChange={(e) => updateDraft({ amount: e.target.value })}
                      className="h-9 rounded-xl text-xs"
                    />
                  </div>
                </div>
              )}

              <div className="flex justify-end pt-2">
                <Button
                  type="button"
                  size="sm"
                  onClick={addOrUpdateTagRule}
                  disabled={!isDraftValid}
                  className="rounded-xl gap-1.5 h-8 px-4 text-xs font-semibold"
                >
                  <Plus className="h-3.5 w-3.5" />
                  <span>
                    {editingTag
                      ? String(w.updateCategoryButton || 'Update Category')
                      : String(w.addCategoryButton || 'Add Category to Budget')}
                  </span>
                </Button>
              </div>
            </div>

            {/* List of Configured Tag Budgets */}
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <h4 className="text-xs font-bold uppercase tracking-wider text-muted-foreground">
                  <span>{`${String(w.configuredCategoriesTitle || 'Configured Categories')} (${tagBudgets.length})`}</span>
                </h4>
              </div>

              {tagBudgets.length === 0 ? (
                <div className="rounded-2xl border border-dashed border-border/80 p-6 text-center bg-muted/10">
                  <Info className="h-5 w-5 text-muted-foreground mx-auto mb-2" />
                  <p className="text-xs font-semibold text-foreground">
                    {String(w.noCategoriesYetTitle || 'No budget categories added yet (Optional)')}
                  </p>
                  <p className="text-[11px] text-muted-foreground max-w-sm mx-auto mt-1">
                    {String(w.noCategoriesYetDesc || '')}
                  </p>
                </div>
              ) : (
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-2.5">
                  {tagBudgets.map((tb) => {
                    const annualAmt = tb.getAnnualTotal(startDate, endDate);
                    const formattedEff = new Intl.NumberFormat().format(tb.effectiveAmount);
                    const formattedAnn = new Intl.NumberFormat().format(annualAmt);
                    return (
                      <div
                        key={tb.tag}
                        className="flex items-center justify-between rounded-xl border border-border/70 p-3 bg-card shadow-2xs"
                      >
                        <div className="space-y-1">
                          <div className="flex items-center gap-2">
                            <span className="font-bold text-xs text-primary">{tb.displayTag}</span>
                            <Badge variant="outline" className="text-[10px] px-1.5 py-0 rounded-md">
                              {String(w.frequencies?.[tb.frequency] || tb.frequency)}
                            </Badge>
                          </div>
                          <div className="text-[11px] text-muted-foreground">
                            <span>{`${formattedEff} ${mainCurrency} → `}</span>
                            <span className="font-semibold text-foreground">
                              {`${formattedAnn} ${mainCurrency}${String(w.perYearSuffix || '/yr')}`}
                            </span>
                          </div>
                        </div>
                        <div className="flex items-center gap-1">
                          <Button
                            type="button"
                            variant="ghost"
                            size="icon"
                            onClick={() => startEditTag(tb.tag)}
                            className="h-7 w-7 text-muted-foreground hover:text-foreground rounded-lg"
                          >
                            <Edit2 className="h-3.5 w-3.5" />
                          </Button>
                          <Button
                            type="button"
                            variant="ghost"
                            size="icon"
                            onClick={() => removeTagRule(tb.tag)}
                            className="h-7 w-7 text-muted-foreground hover:text-rose-500 rounded-lg"
                          >
                            <Trash2 className="h-3.5 w-3.5" />
                          </Button>
                        </div>
                      </div>
                    );
                  })}
                </div>
              )}
            </div>
          </div>
        )}

        {/* STEP 3: ANNUAL SCHEDULE PREVIEW */}
        {step === 3 && (
          <div className="space-y-5 py-2">
            {/* KPI Summary Banner */}
            <div className="grid grid-cols-3 gap-3">
              <div className="rounded-2xl border border-border/70 p-3.5 bg-primary/5">
                <span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
                  {String(w.preview?.totalAnnualBudgetTitle || 'Projected Annual Budget')}
                </span>
                <p className="text-xl font-black text-primary mt-1">
                  <span>{`${new Intl.NumberFormat().format(schedulePreview.totalAnnualBudget)} ${mainCurrency}`}</span>
                </p>
              </div>
              <div className="rounded-2xl border border-border/70 p-3.5 bg-muted/20">
                <span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
                  {String(w.preview?.avgMonthlyPlannedTitle || 'Avg. Monthly Projection')}
                </span>
                <p className="text-xl font-black text-foreground mt-1">
                  <span>{`${new Intl.NumberFormat().format(Math.round(schedulePreview.totalAnnualBudget / 12))} ${mainCurrency}`}</span>
                </p>
              </div>
              <div className="rounded-2xl border border-border/70 p-3.5 bg-muted/20">
                <span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
                  {String(w.preview?.categoriesCountTitle || 'Planned Categories')}
                </span>
                <p className="text-xl font-black text-foreground mt-1">
                  {tagBudgets.length}
                </p>
              </div>
            </div>

            {/* Recharts 12-Month Bar Chart */}
            <div className="rounded-2xl border border-border/70 p-4 bg-card shadow-2xs space-y-2">
              <h4 className="text-xs font-bold uppercase tracking-wider text-foreground">
                {String(w.preview?.monthlyScheduleChartTitle || 'Projected 12-Month Schedule')}
              </h4>
              <div className="h-56 w-full pt-2">
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart
                    data={schedulePreview.months}
                    margin={{
                      top: 10,
                      right: 10,
                      left: -20,
                      bottom: 0,
                    }}
                  >
                    <CartesianGrid strokeDasharray="3 3" vertical={false} opacity={0.2} />
                    <XAxis
                      dataKey="label"
                      tickLine={false}
                      axisLine={false}
                      tick={{ fontSize: 10, fill: 'currentColor', opacity: 0.7 }}
                    />
                    <YAxis
                      tickLine={false}
                      axisLine={false}
                      tick={{ fontSize: 10, fill: 'currentColor', opacity: 0.7 }}
                      tickFormatter={(val) => new Intl.NumberFormat().format(val)}
                    />
                    <Tooltip
                      formatter={(val) => [
                        `${new Intl.NumberFormat().format(val)} ${mainCurrency}`,
                        String(w.preview?.projectedLabel || 'Projected'),
                      ]}
                      contentStyle={{
                        borderRadius: '0.75rem',
                        fontSize: '11px',
                        background: 'hsl(var(--card))',
                        border: '1px solid hsl(var(--border))',
                      }}
                    />
                    <Bar
                      dataKey="projected"
                      fill="hsl(var(--primary))"
                      radius={[6, 6, 0, 0]}
                    />
                  </BarChart>
                </ResponsiveContainer>
              </div>
            </div>

            {/* Per-Tag Annual Breakdown */}
            {tagBudgets.length > 0 && (
              <div className="rounded-2xl border border-border/70 p-4 bg-card shadow-2xs space-y-2.5">
                <h4 className="text-xs font-bold uppercase tracking-wider text-muted-foreground">
                  {String(w.preview?.categoryBreakdownTitle || 'Annual Budget Breakdown by Category')}
                </h4>
                <div className="space-y-2">
                  {tagBudgets.map((tb) => {
                    const annual = tb.getAnnualTotal(startDate, endDate);
                    const pct = schedulePreview.totalAnnualBudget > 0
                      ? Math.round((annual / schedulePreview.totalAnnualBudget) * 100)
                      : 0;
                    return (
                      <div key={tb.tag} className="flex items-center justify-between text-xs">
                        <div className="flex items-center gap-2">
                          <span className="font-semibold text-primary">{tb.displayTag}</span>
                          <span className="text-muted-foreground text-[11px]">{`(${pct}%)`}</span>
                        </div>
                        <span className="font-bold text-foreground">
                          {`${new Intl.NumberFormat().format(annual)} ${mainCurrency}`}
                        </span>
                      </div>
                    );
                  })}
                </div>
              </div>
            )}
          </div>
        )}

        {/* Modal Footer Controls */}
        <div className="flex items-center justify-between border-t border-border/50 pt-4 mt-2">
          {step === 1 ? (
            <Button
              type="button"
              variant="outline"
              onClick={handleClose}
              className="rounded-xl h-9 px-4 text-xs font-semibold"
            >
              {String(content.cancelButton)}
            </Button>
          ) : (
            <Button
              type="button"
              variant="outline"
              onClick={prevStep}
              className="rounded-xl h-9 px-3.5 text-xs font-semibold gap-1.5"
            >
              <ChevronLeft className="h-4 w-4" />
              <span>{String(w.backButton || 'Back')}</span>
            </Button>
          )}

          {step < 3 ? (
            <div className="flex items-center gap-2">
              {step === 2 && (
                <Button
                  type="button"
                  variant="ghost"
                  onClick={nextStep}
                  className="rounded-xl h-9 px-3.5 text-xs font-semibold text-muted-foreground"
                >
                  {String(w.skipStepButton || 'Skip & Preview')}
                </Button>
              )}
              <Button
                type="button"
                onClick={nextStep}
                disabled={!canGoNext}
                className="rounded-xl h-9 px-4 text-xs font-semibold gap-1.5"
              >
                <span>{String(w.nextButton || 'Next Step')}</span>
                <ChevronRight className="h-4 w-4" />
              </Button>
            </div>
          ) : (
            <Button
              type="button"
              onClick={handleFinish}
              disabled={submitting}
              className="rounded-xl h-9 px-5 text-xs font-semibold gap-2 shadow-sm shadow-primary/20"
            >
              {submitting ? (
                <>
                  <Loader2 className="h-4 w-4 animate-spin" />
                  <span>{String(w.completeWizardButton || 'Saving...')}</span>
                </>
              ) : (
                <>
                  <Check className="h-4 w-4" />
                  <span>{String(w.completeWizardButton || 'Create Project & Save Budget')}</span>
                </>
              )}
            </Button>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}

ProjectSetupWizardModal.propTypes = {
  open: PropTypes.bool.isRequired,
  onSuccess: PropTypes.func.isRequired,
  onCancel: PropTypes.func.isRequired,
};

export default ProjectSetupWizardModal;
