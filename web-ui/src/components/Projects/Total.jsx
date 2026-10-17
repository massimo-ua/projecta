import React from 'react';
import PropTypes from 'prop-types';
import { useParams } from 'react-router-dom';
import { useIntlayer } from 'react-intlayer';
import {
  Calculator,
  Wallet,
  Target,
  CheckCircle2,
  Flame,
  BarChart3,
  Calendar,
} from 'lucide-react';
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

function KpiCard({
  title,
  value,
  currency,
  icon: Icon,
  iconTone = 'primary',
  badges = [],
}) {
  const toneClasses = {
    primary: 'bg-primary/10 text-primary ring-primary/20',
    emerald: 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 ring-emerald-500/20',
    amber: 'bg-amber-500/10 text-amber-600 dark:text-amber-400 ring-amber-500/20',
    indigo: 'bg-indigo-500/10 text-indigo-600 dark:text-indigo-400 ring-indigo-500/20',
    rose: 'bg-rose-500/10 text-rose-600 dark:text-rose-400 ring-rose-500/20',
  };

  return (
    <div className="rounded-2xl border border-border/70 bg-card p-5 shadow-xs transition-all duration-300 hover:shadow-md hover:border-primary/40 flex flex-col justify-between gap-3">
      <div>
        <div className="flex items-center justify-between gap-3 mb-3">
          <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider truncate">
            {title}
          </span>
          <div
            className={cn(
              'flex h-9 w-9 shrink-0 items-center justify-center rounded-xl ring-1',
              toneClasses[iconTone] || toneClasses.primary,
            )}
          >
            <Icon className="h-4 w-4" />
          </div>
        </div>
        <div className="flex items-baseline gap-2 flex-wrap">
          <span className="text-2xl sm:text-3xl font-extrabold tracking-tight text-foreground font-mono">
            {value}
          </span>
          {currency && (
            <Badge variant="outline" className="font-mono text-xs font-semibold border-border/80 bg-muted/30">
              {currency}
            </Badge>
          )}
        </div>
      </div>

      {badges.length > 0 && (
        <div className="flex items-center flex-wrap gap-1.5 pt-1 border-t border-border/40">
          {badges.map((b) => (
            <Badge
              key={b.label}
              variant={b.variant || 'secondary'}
              className={cn(
                'text-[11px] font-medium px-2 py-0.5',
                b.tone === 'rose' && 'bg-rose-500/10 text-rose-600 dark:text-rose-400 border-rose-500/20',
                b.tone === 'emerald' && 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20',
              )}
            >
              {b.label}
            </Badge>
          ))}
        </div>
      )}
    </div>
  );
}

KpiCard.propTypes = {
  title: PropTypes.string.isRequired,
  value: PropTypes.string.isRequired,
  currency: PropTypes.string,
  icon: PropTypes.elementType.isRequired,
  iconTone: PropTypes.string,
  badges: PropTypes.arrayOf(
    PropTypes.shape({
      label: PropTypes.string.isRequired,
      variant: PropTypes.string,
      tone: PropTypes.string,
    }),
  ),
};

export default function Total() {
  const content = useIntlayer('total');
  const { projectId } = useParams();
  const {
    loading,
    timeframe,
    setTimeframe,
    currency,
    project,
    payments,
    assets,
    kpis,
  } = useProjectAnalytics(projectId);

  const timeframeOptions = [
    { key: 'all', label: String(content.timeframe.all) },
    { key: 'year', label: String(content.timeframe.year) },
    { key: '12m', label: String(content.timeframe.last12m) },
    { key: '6m', label: String(content.timeframe.last6m) },
  ];

  const fmt = (num) => new Intl.NumberFormat().format(num || 0);

  const investedBadges = [
    {
      label: `${kpis?.investmentCount || 0} ${String(content.cards?.investmentsCount || 'investments')}`,
      variant: 'secondary',
    },
  ];
  if (kpis?.totalLaborHours > 0) {
    investedBadges.push({
      label: `${kpis.totalLaborHours} ${String(content.cards?.laborHours || 'hrs labor')}`,
      variant: 'outline',
    });
  }

  const budgetBadges = [];
  if (kpis?.totalTargetBudget > 0 && kpis?.budgetUtilization !== null) {
    budgetBadges.push({
      label: `${kpis.budgetUtilization}% ${String(content.cards?.utilized || 'utilized')}`,
      variant: 'outline',
      tone: kpis.budgetUtilization > 100 ? 'rose' : 'emerald',
    });
    if (kpis.overBudgetCount > 0) {
      budgetBadges.push({
        label: `${kpis.overBudgetCount} ${String(content.cards?.overBudgetAssets || 'over budget')}`,
        variant: 'outline',
        tone: 'rose',
      });
    }
  } else {
    budgetBadges.push({
      label: String(content.cards?.noBudgetSet || 'No target budgets set'),
      variant: 'secondary',
    });
  }

  const assetsBadges = [
    {
      label: `${kpis?.completedAssetsCount || 0} ${String(content.cards?.completed || 'completed')}`,
      variant: 'outline',
      tone: 'emerald',
    },
    {
      label: `${kpis?.activeAssetsCount || 0} ${String(content.cards?.active || 'active')}`,
      variant: 'secondary',
    },
  ];

  const burnBadges = [
    {
      label: `${kpis?.activeMonthsCount || 0} ${String(content.cards?.activeMonths || 'active months')}`,
      variant: 'secondary',
    },
  ];

  return (
    <div className="space-y-8">
      {/* KPI Summary Section with Unified Timeframe Selector */}
      <div className="space-y-4">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-3 border-b border-border/60">
          <div className="flex items-center gap-2.5">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10 text-primary">
              <Calculator className="h-4 w-4" />
            </div>
            <h2 className="text-lg font-bold tracking-tight text-foreground">
              {String(content.summaryTitle)}
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

        {loading ? (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5">
            {[1, 2, 3, 4].map((i) => (
              <div key={i} className="rounded-2xl border border-border/60 p-5 space-y-4 bg-card">
                <div className="flex justify-between items-center">
                  <Skeleton className="h-4 w-28 rounded-md" />
                  <Skeleton className="h-9 w-9 rounded-xl" />
                </div>
                <Skeleton className="h-8 w-36 rounded-lg" />
                <Skeleton className="h-5 w-24 rounded-md" />
              </div>
            ))}
          </div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5">
            <KpiCard
              title={String(content.cards?.totalPayments || 'Total Invested')}
              value={fmt(kpis?.totalInvested)}
              currency={currency}
              icon={Wallet}
              iconTone="primary"
              badges={investedBadges}
            />
            <KpiCard
              title={String(content.cards?.targetBudget || 'Target Budget')}
              value={kpis?.totalTargetBudget > 0 ? fmt(kpis.totalTargetBudget) : '—'}
              currency={kpis?.totalTargetBudget > 0 ? currency : ''}
              icon={Target}
              iconTone={kpis?.budgetUtilization > 100 ? 'rose' : 'emerald'}
              badges={budgetBadges}
            />
            <KpiCard
              title={String(content.cards?.assetsProgress || 'Assets Progress')}
              value={`${kpis?.completedAssetsCount || 0} / ${kpis?.totalAssetsCount || 0}`}
              icon={CheckCircle2}
              iconTone="indigo"
              badges={assetsBadges}
            />
            <KpiCard
              title={String(content.cards?.avgMonthlyBurn || 'Avg. Monthly Burn')}
              value={fmt(kpis?.avgMonthlyBurn)}
              currency={currency}
              icon={Flame}
              iconTone="amber"
              badges={burnBadges}
            />
          </div>
        )}
      </div>

      {/* Analytics & Charts Section */}
      <div className="space-y-6">
        <div className="flex items-center gap-2.5 pb-3 border-b border-border/60">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10 text-primary">
            <BarChart3 className="h-4 w-4" />
          </div>
          <h2 className="text-lg font-bold tracking-tight text-foreground">
            {String(content.analyticsTitle)}
          </h2>
        </div>

        {loading ? (
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
            {/* Row 1: Full-width Monthly / Cumulative Investment Trend */}
            <MonthlySpendingTrend
              payments={payments}
              currency={currency}
              totalTargetBudget={kpis?.totalTargetBudget || 0}
              content={content}
            />

            {/* Row 2: Category Breakdown & Resource / Contributor Breakdown */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <CategoryBreakdown
                payments={payments}
                currency={currency}
                content={content}
              />
              <PaymentKindBreakdown
                payments={payments}
                project={project}
                currency={currency}
                content={content}
              />
            </div>

            {/* Row 3: Top Assets by Investment & Budget vs. Actual */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <TopCostTypes
                payments={payments}
                assets={assets}
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

