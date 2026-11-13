import React from 'react';
import PropTypes from 'prop-types';
import { useParams } from 'react-router-dom';
import { useIntlayer } from 'react-intlayer';
import { useProjectCashFlow } from '../../../hooks/useProjectCashFlow';
import { MonthlyCashFlowChart } from './MonthlyCashFlowChart';
import { AssetCostBurnTable } from './AssetCostBurnTable';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { Skeleton } from '@/components/ui/skeleton';
import { RefreshCw, AlertTriangle, TrendingUp } from 'lucide-react';

function formatMoneyCents(cents, currency) {
  const major = (Number(cents) || 0) / 100;
  return `${major.toLocaleString(undefined, { minimumFractionDigits: 0, maximumFractionDigits: 2 })} ${currency}`;
}

function KpiCard({ label, value, subtext, highlight }) {
  return (
    <Card className={`border-border/60 shadow-sm ${highlight ? 'border-amber-500/40 bg-amber-500/5' : ''}`}>
      <CardContent className="p-4 space-y-1">
        <p className="text-xs font-medium text-muted-foreground">{label}</p>
        <p className="text-lg font-bold font-mono text-foreground">{value}</p>
        {subtext && <p className="text-[11px] text-muted-foreground">{subtext}</p>}
      </CardContent>
    </Card>
  );
}

KpiCard.propTypes = {
  label: PropTypes.string.isRequired,
  value: PropTypes.string.isRequired,
  subtext: PropTypes.string,
  highlight: PropTypes.bool,
};

export function CashFlowBoard() {
  const { projectId } = useParams();
  const content = useIntlayer('cash-flow-board');
  const {
    forecast,
    chartData,
    loading,
    error,
    savingAssetId,
    selectedTag,
    setSelectedTag,
    includeOngoingInCommitted,
    setIncludeOngoingInCommitted,
    previewSchedule,
    commitSchedule,
    reload,
  } = useProjectCashFlow(projectId);

  const currency = forecast?.currency || 'UAH';
  const summary = forecast?.summary || {};
  const assets = forecast?.assets || [];

  return (
    <div className="space-y-6">
      {/* Header & Controls */}
      <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <TrendingUp className="h-5 w-5 text-primary" />
            <h2 className="text-xl font-bold tracking-tight text-foreground">
              {String(content.title || 'True Cost & Cash Flow Planning')}
            </h2>
          </div>
          <p className="text-xs text-muted-foreground max-w-2xl">
            {String(content.subtitle)}
          </p>
        </div>

        <div className="flex flex-wrap items-center gap-3">
          <Input
            type="text"
            value={selectedTag}
            onChange={(e) => setSelectedTag(e.target.value)}
            placeholder={String(content.tagFilterPlaceholder || 'Filter by tag...')}
            className="h-9 w-44 rounded-xl text-xs"
          />

          <div className="flex items-center gap-2 rounded-xl border border-border/60 px-3 py-1.5 bg-card">
            <Switch
              id="stack-ongoing-burn"
              checked={includeOngoingInCommitted}
              onCheckedChange={setIncludeOngoingInCommitted}
            />
            <Label htmlFor="stack-ongoing-burn" className="text-xs font-medium cursor-pointer">
              {String(content.stackOngoingLabel || 'Stack ongoing burn')}
            </Label>
          </div>

          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => reload()}
            className="h-9 rounded-xl gap-1.5 text-xs"
          >
            <RefreshCw className="h-3.5 w-3.5" />
            <span>{String(content.refreshLabel || 'Refresh')}</span>
          </Button>
        </div>
      </div>

      {error && (
        <div className="flex items-center gap-2 rounded-xl border border-destructive/40 bg-destructive/10 p-3 text-xs text-destructive">
          <AlertTriangle className="h-4 w-4 shrink-0" />
          <span>{error}</span>
        </div>
      )}

      {loading && !forecast ? (
        <div className="space-y-4">
          <div className="grid grid-cols-2 lg:grid-cols-5 gap-3">
            <Skeleton className="h-24 rounded-xl" />
            <Skeleton className="h-24 rounded-xl" />
            <Skeleton className="h-24 rounded-xl" />
            <Skeleton className="h-24 rounded-xl" />
            <Skeleton className="h-24 rounded-xl" />
          </div>
          <Skeleton className="h-[340px] rounded-xl" />
        </div>
      ) : (
        <>
          {/* Summary Strip */}
          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-3">
            <KpiCard
              label={String(content.totalTargetLabel || 'Target Budgets')}
              value={formatMoneyCents(summary.totalTargetCents, currency)}
            />
            <KpiCard
              label={String(content.actualCashLabel || 'Actual Cash')}
              value={formatMoneyCents(summary.totalActualCashCents, currency)}
            />
            <KpiCard
              label={String(content.remainingScheduledLabel || 'Scheduled Left')}
              value={formatMoneyCents(summary.totalRemainingScheduledCents, currency)}
            />
            <KpiCard
              label={String(content.ongoingBurnLabel || 'Ongoing Burn / mo')}
              value={formatMoneyCents(summary.estimatedOngoingMonthlyCashCents, currency)}
              subtext="6m trailing undated"
            />
            <KpiCard
              label={String(content.nonCashLabel || 'Time & Goods')}
              value={`${(Number(summary.totalTimeHours) || 0).toFixed(1)} hrs`}
              subtext={`${formatMoneyCents(summary.totalGoodsCents, currency)} goods`}
            />
            <KpiCard
              label={String(content.peakCommitmentLabel || 'Peak Cash Month')}
              value={summary.peakMonth || '—'}
              subtext={
                summary.alertMonthsCount > 0
                  ? `${summary.alertMonthsCount} ceiling alert month(s)`
                  : formatMoneyCents(summary.peakCommittedCashCents, currency)
              }
              highlight={summary.alertMonthsCount > 0}
            />
          </div>

          {/* Chart */}
          {chartData.length > 0 && (
            <MonthlyCashFlowChart
              chartData={chartData}
              currency={currency}
              includeOngoingInCommitted={includeOngoingInCommitted}
              title={String(content.chartTitle || 'Monthly Cash Demand')}
              subtitle={String(content.chartSubtitle || '')}
            />
          )}

          {/* Asset Schedule & True-Cost Breakdown */}
          {assets.length > 0 ? (
            <AssetCostBurnTable
              assets={assets}
              currency={currency}
              savingAssetId={savingAssetId}
              onPreviewSchedule={previewSchedule}
              onCommitSchedule={commitSchedule}
              title={String(content.tableTitle || 'Asset Schedules')}
              subtitle={String(content.tableSubtitle || '')}
            />
          ) : (
            <Card className="border-dashed border-border/70">
              <CardContent className="p-8 text-center space-y-1">
                <p className="text-sm font-semibold text-foreground">{String(content.emptyTitle)}</p>
                <p className="text-xs text-muted-foreground">{String(content.emptySubtitle)}</p>
              </CardContent>
            </Card>
          )}
        </>
      )}
    </div>
  );
}
