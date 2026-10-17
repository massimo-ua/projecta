import React, { useMemo } from 'react';
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Cell,
  ResponsiveContainer,
} from 'recharts';
import { Scale } from 'lucide-react';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card';
import { cn } from '@/lib/utils';
import {
  getPaymentHomeAmount,
  getAssetHomeAmount,
  getAssetTargetHomeAmount,
} from '../../../hooks/useProjectAnalytics';
import { ChartTooltip } from './ChartTooltip';

export function PaymentsVsAssets({
  payments = [],
  assets = [],
  currency = '',
  content = {},
}) {
  const targetLabel = String(content.charts?.paymentsVsAssets?.targetLabel || 'Target Budget');
  const actualLabel = String(content.charts?.paymentsVsAssets?.actualLabel || 'Actual Invested');
  const remainingLabel = String(content.charts?.paymentsVsAssets?.remainingLabel || 'Remaining Budget');
  const overBudgetLabel = String(content.charts?.paymentsVsAssets?.overBudgetLabel || 'Over Budget');
  const activeStatusLabel = String(content.cards?.active || 'Active');
  const completedStatusLabel = String(content.cards?.completed || 'Completed');

  const {
    hasBudgetData,
    hasFallbackData,
    budgetChartData,
    fallbackChartData,
    summaryMetrics,
  } = useMemo(() => {
    const assetById = new Map(assets.map((a) => [a.id, a]));
    const budgetedAssets = assets.filter((a) => getAssetTargetHomeAmount(a) > 0);

    if (budgetedAssets.length > 0) {
      let totalTarget = 0;
      let totalActual = 0;

      const items = budgetedAssets.map((a) => {
        const target = Math.round(getAssetTargetHomeAmount(a) * 100) / 100;
        const actual = Math.round(getAssetHomeAmount(a) * 100) / 100;

        const parentHasTarget = Array.isArray(a.parents) && a.parents.some((pLink) => {
          const parentObj = assetById.get(pLink.parentId);
          return parentObj && getAssetTargetHomeAmount(parentObj) > 0;
        });

        if (!parentHasTarget) {
          totalTarget += target;
          totalActual += actual;
        }

        return {
          name: a.name,
          target,
          actual,
          isOverBudget: actual > target,
        };
      });

      items.sort((a, b) => Math.max(b.target, b.actual) - Math.max(a.target, a.actual));

      const roundedTarget = Math.round(totalTarget * 100) / 100;
      const roundedActual = Math.round(totalActual * 100) / 100;
      const variance = Math.round((roundedTarget - roundedActual) * 100) / 100;

      return {
        hasBudgetData: true,
        hasFallbackData: false,
        budgetChartData: items.slice(0, 6),
        fallbackChartData: [],
        summaryMetrics: [
          { label: targetLabel, value: roundedTarget, tone: 'default' },
          { label: actualLabel, value: roundedActual, tone: 'default' },
          {
            label: variance >= 0 ? remainingLabel : overBudgetLabel,
            value: Math.abs(variance),
            tone: variance >= 0 ? 'emerald' : 'rose',
          },
        ],
      };
    }

    // Fallback when no target budgets are configured: Active vs. Completed Root Assets
    const rootAssets = assets.filter((a) => !a.parents || a.parents.length === 0);
    const activeSum = Math.round(
      rootAssets
        .filter((a) => a.isActive)
        .reduce((sum, a) => sum + getAssetHomeAmount(a), 0) * 100,
    ) / 100;
    const completedSum = Math.round(
      rootAssets
        .filter((a) => a.isCompleted)
        .reduce((sum, a) => sum + getAssetHomeAmount(a), 0) * 100,
    ) / 100;
    const totalPayments = Math.round(
      payments.reduce((sum, p) => sum + getPaymentHomeAmount(p), 0) * 100,
    ) / 100;

    return {
      hasBudgetData: false,
      hasFallbackData: activeSum > 0 || completedSum > 0 || totalPayments > 0,
      budgetChartData: [],
      fallbackChartData: [
        {
          name: activeStatusLabel,
          amount: activeSum,
          fill: '#3b82f6',
        },
        {
          name: completedStatusLabel,
          amount: completedSum,
          fill: '#10b981',
        },
      ],
      summaryMetrics: [
        { label: String(content.labels?.payments || 'Total Invested'), value: totalPayments, tone: 'default' },
        { label: activeStatusLabel, value: activeSum, tone: 'default' },
        { label: completedStatusLabel, value: completedSum, tone: 'emerald' },
      ],
    };
  }, [
    payments,
    assets,
    content,
    targetLabel,
    actualLabel,
    remainingLabel,
    overBudgetLabel,
    activeStatusLabel,
    completedStatusLabel,
  ]);

  const hasData = hasBudgetData || hasFallbackData;

  return (
    <Card className="rounded-2xl border-border/70 shadow-xs flex flex-col">
      <CardHeader className="pb-2">
        <div className="flex items-center gap-2">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-sky-500/10 text-sky-600 dark:text-sky-400">
            <Scale className="h-4 w-4" />
          </div>
          <CardTitle className="text-base font-bold tracking-tight">
            {String(content.charts?.paymentsVsAssets?.title || 'Budget vs. Actual')}
          </CardTitle>
        </div>
        <CardDescription className="text-xs">
          {hasBudgetData
            ? String(content.charts?.paymentsVsAssets?.description || 'Planned target budgets compared to actual investments')
            : String(content.charts?.paymentsVsAssets?.fallbackDescription || 'Active vs. completed asset value allocation')}
        </CardDescription>
      </CardHeader>
      <CardContent className="flex-1 flex flex-col justify-between pt-2">
        {!hasData ? (
          <div className="flex h-64 items-center justify-center text-xs text-muted-foreground">
            {String(content.labels?.noData || 'No data')}
          </div>
        ) : (
          <div className="space-y-4">
            <div className="h-56 w-full min-w-0">
              <ResponsiveContainer width="100%" height="100%">
                {hasBudgetData ? (
                  <BarChart data={budgetChartData} margin={{ top: 16, right: 16, left: 0, bottom: 5 }}>
                    <CartesianGrid strokeDasharray="3 3" vertical={false} className="stroke-border/40" />
                    <XAxis
                      dataKey="name"
                      tickLine={false}
                      axisLine={false}
                      tick={{ fill: 'currentColor', fontSize: 11 }}
                      className="text-muted-foreground font-medium"
                    />
                    <YAxis
                      tickLine={false}
                      axisLine={false}
                      tick={{ fill: 'currentColor', fontSize: 11 }}
                      className="text-muted-foreground font-mono"
                      tickFormatter={(val) => new Intl.NumberFormat(undefined, { notation: 'compact' }).format(val)}
                    />
                    <Tooltip content={<ChartTooltip currency={currency} />} />
                    <Bar
                      dataKey="target"
                      name={targetLabel}
                      fill="#64748b"
                      radius={[6, 6, 0, 0]}
                    />
                    <Bar
                      dataKey="actual"
                      name={actualLabel}
                      radius={[6, 6, 0, 0]}
                    >
                      {budgetChartData.map((entry) => (
                        <Cell
                          key={`actual-${entry.name}`}
                          fill={entry.isOverBudget ? '#f43f5e' : '#10b981'}
                        />
                      ))}
                    </Bar>
                  </BarChart>
                ) : (
                  <BarChart data={fallbackChartData} margin={{ top: 16, right: 16, left: 0, bottom: 5 }}>
                    <CartesianGrid strokeDasharray="3 3" vertical={false} className="stroke-border/40" />
                    <XAxis
                      dataKey="name"
                      tickLine={false}
                      axisLine={false}
                      tick={{ fill: 'currentColor', fontSize: 11 }}
                      className="text-muted-foreground font-medium"
                    />
                    <YAxis
                      tickLine={false}
                      axisLine={false}
                      tick={{ fill: 'currentColor', fontSize: 11 }}
                      className="text-muted-foreground font-mono"
                      tickFormatter={(val) => new Intl.NumberFormat(undefined, { notation: 'compact' }).format(val)}
                    />
                    <Tooltip content={<ChartTooltip currency={currency} />} />
                    <Bar
                      dataKey="amount"
                      name={String(content.labels?.amount || 'Amount')}
                      radius={[8, 8, 0, 0]}
                    >
                      {fallbackChartData.map((entry) => (
                        <Cell key={`bar-${entry.name}`} fill={entry.fill} />
                      ))}
                    </Bar>
                  </BarChart>
                )}
              </ResponsiveContainer>
            </div>

            {/* Metrics cards */}
            <div className="grid grid-cols-3 gap-2 text-xs">
              {summaryMetrics.map((metric) => (
                <div key={metric.label} className="p-2.5 rounded-xl border border-border/60 bg-card/60">
                  <span className="text-[11px] font-medium text-muted-foreground block truncate" title={metric.label}>
                    {metric.label}
                  </span>
                  <span
                    className={cn(
                      'font-mono font-bold text-xs sm:text-sm',
                      metric.tone === 'rose' && 'text-rose-600 dark:text-rose-400',
                      metric.tone === 'emerald' && 'text-emerald-600 dark:text-emerald-400',
                      metric.tone === 'default' && 'text-foreground',
                    )}
                  >
                    {new Intl.NumberFormat().format(metric.value)} {currency}
                  </span>
                </div>
              ))}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

export default PaymentsVsAssets;

