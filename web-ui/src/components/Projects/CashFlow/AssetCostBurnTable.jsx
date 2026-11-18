import React, { useState } from 'react';
import PropTypes from 'prop-types';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Check, X, Loader2 } from 'lucide-react';

function toDateInputValue(isoString) {
  if (!isoString) return '';
  return String(isoString).slice(0, 10);
}

function formatCents(cents, currency) {
  const major = (Number(cents) || 0) / 100;
  return `${major.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })} ${currency}`;
}

function ModeBadge({ mode, estimatedCompletion }) {
  const styleMap = {
    OVERDUE_TARGET: 'bg-red-500/10 text-red-600 dark:text-red-400 border-red-500/30',
    SCHEDULED_TARGET: 'bg-blue-500/10 text-blue-600 dark:text-blue-400 border-blue-500/30',
    ONGOING_RUN_RATE: 'bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/30',
    COMPLETED: 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/30',
    UNBUDGETED: 'bg-slate-500/10 text-slate-600 dark:text-slate-400 border-slate-500/30',
  };
  const labelMap = {
    OVERDUE_TARGET: 'Overdue Target',
    SCHEDULED_TARGET: 'Scheduled',
    ONGOING_RUN_RATE: estimatedCompletion ? `ETA ${estimatedCompletion}` : 'Ongoing Target',
    COMPLETED: 'Completed',
    UNBUDGETED: 'Unbudgeted Tracker',
  };

  return (
    <Badge variant="outline" className={`text-[10px] py-0 px-1.5 font-semibold ${styleMap[mode] || ''}`}>
      {labelMap[mode] || mode}
    </Badge>
  );
}

ModeBadge.propTypes = {
  mode: PropTypes.string.isRequired,
  estimatedCompletion: PropTypes.string,
};

function ScheduleEditorRow({
  item,
  currency,
  savingAssetId,
  onPreviewSchedule,
  onCommitSchedule,
}) {
  const [startVal, setStartVal] = useState(toDateInputValue(item.startDate));
  const [targetVal, setTargetVal] = useState(toDateInputValue(item.targetCompletionDate));

  const initialStart = toDateInputValue(item.startDate);
  const initialTarget = toDateInputValue(item.targetCompletionDate);
  const isDirty = startVal !== initialStart || targetVal !== initialTarget;
  const isSaving = savingAssetId === item.assetId;

  const handleSave = () => {
    onCommitSchedule(item.assetId, {
      startDate: startVal ? new Date(`${startVal}T00:00:00Z`) : undefined,
      targetCompletionDate: targetVal ? new Date(`${targetVal}T00:00:00Z`) : undefined,
      clearTargetDate: !targetVal,
    });
  };

  const handleClearTarget = () => {
    setTargetVal('');
    onPreviewSchedule(item.assetId, { clearTargetDate: true });
    onCommitSchedule(item.assetId, { clearTargetDate: true });
  };

  return (
    <tr className="border-b border-border/60 transition-colors hover:bg-muted/30">
      <td className="p-3 align-top">
        <div className="space-y-1">
          <div className="font-semibold text-sm text-foreground">{item.assetName}</div>
          <ModeBadge mode={item.forecastMode} estimatedCompletion={item.estimatedCompletion} />
        </div>
      </td>
      <td className="p-3 align-top">
        <div className="flex flex-wrap items-center gap-1.5">
          <Input
            type="date"
            aria-label={`Start date for ${item.assetName}`}
            value={startVal}
            onChange={(e) => {
              setStartVal(e.target.value);
              if (e.target.value) {
                onPreviewSchedule(item.assetId, {
                  startDate: e.target.value,
                  targetCompletionDate: targetVal || undefined,
                  clearTargetDate: !targetVal,
                });
              }
            }}
            className="h-8 w-[138px] text-xs rounded-lg"
          />
          <span className="text-muted-foreground text-xs">→</span>
          <Input
            type="date"
            aria-label={`Target completion date for ${item.assetName}`}
            value={targetVal}
            min={startVal || undefined}
            onChange={(e) => {
              setTargetVal(e.target.value);
              onPreviewSchedule(item.assetId, {
                startDate: startVal || undefined,
                targetCompletionDate: e.target.value || undefined,
                clearTargetDate: !e.target.value,
              });
            }}
            className="h-8 w-[138px] text-xs rounded-lg"
          />
          {targetVal && (
            <Button
              type="button"
              variant="ghost"
              size="icon"
              className="h-7 w-7 text-muted-foreground hover:text-destructive"
              title="Clear target date"
              disabled={isSaving}
              onClick={handleClearTarget}
            >
              <X className="h-3.5 w-3.5" />
            </Button>
          )}
          {isDirty && (
            <Button
              type="button"
              size="sm"
              className="h-7 px-2 text-xs gap-1"
              disabled={isSaving}
              onClick={handleSave}
            >
              {isSaving ? <Loader2 className="h-3 w-3 animate-spin" /> : <Check className="h-3 w-3" />}
              <span>Save</span>
            </Button>
          )}
        </div>
      </td>
      <td className="p-3 text-right font-mono text-xs align-top">
        {item.hasTargetBudget ? formatCents(item.targetHomeCents, currency) : (
          <span className="text-muted-foreground italic">No ceiling</span>
        )}
      </td>
      <td className="p-3 text-right font-mono text-xs align-top">
        <div>{formatCents(item.rolledUpTotalCashCents, currency)}</div>
        {item.hasTargetBudget && (
          <div className="text-[11px] text-muted-foreground">
            Left: {formatCents(item.residualTargetCents, currency)}
          </div>
        )}
      </td>
      <td className="p-3 text-right font-mono text-xs align-top">
        <div>{item.actualToDate.timeHours.toFixed(1)} hrs</div>
        <div className="text-[11px] text-muted-foreground">
          {formatCents(item.actualToDate.goodsCents, currency)}
        </div>
      </td>
      <td className="p-3 text-right font-mono text-xs align-top">
        <div>{formatCents(item.burnRateMonthly.cashCentsPerMonth, currency)}/mo</div>
        {item.burnRateMonthly.timeHoursPerMonth > 0 && (
          <div className="text-[11px] text-muted-foreground">
            {item.burnRateMonthly.timeHoursPerMonth.toFixed(1)} hrs/mo
          </div>
        )}
      </td>
    </tr>
  );
}

ScheduleEditorRow.propTypes = {
  item: PropTypes.object.isRequired,
  currency: PropTypes.string.isRequired,
  savingAssetId: PropTypes.string,
  onPreviewSchedule: PropTypes.func.isRequired,
  onCommitSchedule: PropTypes.func.isRequired,
};

export function AssetCostBurnTable({
  assets,
  currency,
  savingAssetId,
  onPreviewSchedule,
  onCommitSchedule,
  title,
  subtitle,
}) {
  return (
    <Card className="border-border/60 shadow-sm">
      <CardHeader className="pb-2">
        <CardTitle className="text-base font-bold">{title}</CardTitle>
        <CardDescription className="text-xs">{subtitle}</CardDescription>
      </CardHeader>
      <CardContent className="p-0">
        <div className="overflow-x-auto">
          <table className="w-full caption-bottom text-sm">
            <thead className="[&_tr]:border-b border-border/60 bg-muted/30">
              <tr className="border-b border-border/60">
                <th className="h-10 px-3 text-left align-middle text-xs font-medium text-muted-foreground">Asset &amp; Mode</th>
                <th className="h-10 px-3 text-left align-middle text-xs font-medium text-muted-foreground">Schedule Window</th>
                <th className="h-10 px-3 text-right align-middle text-xs font-medium text-muted-foreground">Target Budget</th>
                <th className="h-10 px-3 text-right align-middle text-xs font-medium text-muted-foreground">Cash Spent / Residual</th>
                <th className="h-10 px-3 text-right align-middle text-xs font-medium text-muted-foreground">Time &amp; Goods</th>
                <th className="h-10 px-3 text-right align-middle text-xs font-medium text-muted-foreground">6m Avg Burn</th>
              </tr>
            </thead>
            <tbody>
              {assets.map((item) => (
                <ScheduleEditorRow
                  key={item.assetId}
                  item={item}
                  currency={currency}
                  savingAssetId={savingAssetId}
                  onPreviewSchedule={onPreviewSchedule}
                  onCommitSchedule={onCommitSchedule}
                />
              ))}
            </tbody>
          </table>
        </div>
      </CardContent>
    </Card>
  );
}

AssetCostBurnTable.propTypes = {
  assets: PropTypes.array.isRequired,
  currency: PropTypes.string.isRequired,
  savingAssetId: PropTypes.string,
  onPreviewSchedule: PropTypes.func.isRequired,
  onCommitSchedule: PropTypes.func.isRequired,
  title: PropTypes.string.isRequired,
  subtitle: PropTypes.string.isRequired,
};
