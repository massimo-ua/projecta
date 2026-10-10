import React, { useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { useIntlayer } from 'react-intlayer';
import {
  Calculator,
  Wallet,
  Scale,
  BarChart3,
  Calendar,
} from 'lucide-react';
import { useProjectTotals } from '../../hooks/projects';
import { useProjectAnalytics } from '../../hooks/useProjectAnalytics';
import { Skeleton } from '@/components/ui/skeleton';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';
import {
  MonthlySpendingTrend,
  CategoryBreakdown,
  PaymentKindBreakdown,
  TopCostTypes,
  PaymentsVsAssets,
} from './Charts';
import './Total.css';

function TotalCard({ total }) {
  const isBalance = total.key.includes('balance');
  const Icon = isBalance ? Scale : Wallet;

  return (
    <div className="rounded-2xl border border-border/70 bg-card p-6 shadow-xs transition-all duration-300 hover:shadow-md hover:border-primary/40">
      <div className="flex items-center justify-between gap-3 mb-4">
        <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
          {total.title}
        </span>
        <div
          className={cn(
            'flex h-9 w-9 items-center justify-center rounded-xl ring-1',
            isBalance
              ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 ring-emerald-500/20'
              : 'bg-primary/10 text-primary ring-primary/20',
          )}
        >
          <Icon className="h-4 w-4" />
        </div>
      </div>
      <div className="flex items-baseline gap-2.5">
        <span className="text-3xl font-extrabold tracking-tight text-foreground">
          {total.amount}
        </span>
        {total.currency && (
          <Badge variant="outline" className="font-mono text-xs font-semibold border-border/80 bg-muted/30">
            {total.currency}
          </Badge>
        )}
      </div>
    </div>
  );
}

export default function Total() {
  const content = useIntlayer('total');
  const { projectId } = useParams();
  const [loadingTotals, totals, updateTotals] = useProjectTotals(projectId);
  const {
    loading: loadingAnalytics,
    timeframe,
    setTimeframe,
    currency,
    payments,
    assets,
  } = useProjectAnalytics(projectId);

  useEffect(() => {
    updateTotals();
  }, [projectId]);

  const timeframeOptions = [
    { key: 'all', label: String(content.timeframe.all) },
    { key: 'year', label: String(content.timeframe.year) },
    { key: '12m', label: String(content.timeframe.last12m) },
    { key: '6m', label: String(content.timeframe.last6m) },
  ];

  return (
    <div className="space-y-8">
      {/* KPI Summary Section */}
      <div className="space-y-4">
        <div className="flex items-center gap-2.5 pb-3 border-b border-border/60">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10 text-primary">
            <Calculator className="h-4 w-4" />
          </div>
          <h2 className="text-lg font-bold tracking-tight text-foreground">
            {String(content.summaryTitle)}
          </h2>
        </div>

        {loadingTotals ? (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
            {[1, 2, 3].map((i) => (
              <div key={i} className="rounded-2xl border border-border/60 p-6 space-y-4 bg-card">
                <div className="flex justify-between items-center">
                  <Skeleton className="h-4 w-28 rounded-md" />
                  <Skeleton className="h-8 w-8 rounded-xl" />
                </div>
                <Skeleton className="h-8 w-36 rounded-lg" />
              </div>
            ))}
          </div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
            {totals.map((total) => (
              <TotalCard key={total.key} total={total} />
            ))}
          </div>
        )}
      </div>

      {/* Analytics & Charts Section */}
      <div className="space-y-6">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-3 border-b border-border/60">
          <div className="flex items-center gap-2.5">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10 text-primary">
              <BarChart3 className="h-4 w-4" />
            </div>
            <h2 className="text-lg font-bold tracking-tight text-foreground">
              {String(content.analyticsTitle)}
            </h2>
          </div>

          {/* Timeframe Filter Controls */}
          <div className="flex items-center gap-1.5 p-1 bg-muted/50 rounded-xl border border-border/50 self-start sm:self-auto overflow-x-auto">
            <Calendar className="h-3.5 w-3.5 text-muted-foreground ml-1.5 mr-0.5 shrink-0" />
            {timeframeOptions.map((opt) => (
              <button
                key={opt.key}
                type="button"
                onClick={() => setTimeframe(opt.key)}
                className={cn(
                  'px-3 py-1 text-xs font-medium rounded-lg transition-all whitespace-nowrap',
                  timeframe === opt.key
                    ? 'bg-primary text-primary-foreground shadow-xs font-semibold'
                    : 'text-muted-foreground hover:text-foreground hover:bg-background/80',
                )}
              >
                {opt.label}
              </button>
            ))}
          </div>
        </div>

        {loadingAnalytics ? (
          <div className="space-y-6">
            <Skeleton className="h-80 w-full rounded-2xl" />
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <Skeleton className="h-80 w-full rounded-2xl" />
              <Skeleton className="h-80 w-full rounded-2xl" />
            </div>
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <Skeleton className="h-80 w-full rounded-2xl" />
              <Skeleton className="h-80 w-full rounded-2xl" />
            </div>
          </div>
        ) : (
          <div className="space-y-6">
            {/* Row 1: Full-width Monthly Spending Trend */}
            <MonthlySpendingTrend
              payments={payments}
              currency={currency}
              content={content}
            />

            {/* Row 2: Category Breakdown & Payment Kind Breakdown */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <CategoryBreakdown
                payments={payments}
                currency={currency}
                content={content}
              />
              <PaymentKindBreakdown
                payments={payments}
                currency={currency}
                content={content}
              />
            </div>

            {/* Row 3: Top Cost Types & Payments vs Assets */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <TopCostTypes
                payments={payments}
                currency={currency}
                content={content}
              />
              <PaymentsVsAssets
                payments={payments}
                assets={assets}
                currency={currency}
                content={content}
              />
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
