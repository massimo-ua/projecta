import React, { useMemo, useState } from 'react';
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ReferenceLine,
  ResponsiveContainer,
} from 'recharts';
import {
  format,
  parseISO,
  startOfMonth,
  eachMonthOfInterval,
} from 'date-fns';
import { TrendingUp } from 'lucide-react';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card';
import { cn } from '@/lib/utils';
import { getPaymentHomeAmount } from '../../../hooks/useProjectAnalytics';
import { ChartTooltip } from './ChartTooltip';

export function MonthlySpendingTrend({
  payments = [],
  currency = '',
  totalTargetBudget = 0,
  content = {},
}) {
  const [viewMode, setViewMode] = useState('monthly');

  const { chartData, totalPeriodSpending, avgMonthlyBurn } = useMemo(() => {
    const map = new Map();
    let minDate = null;
    let maxDate = null;

    payments.forEach((p) => {
      if (!p.rawDate) return;
      try {
        const date = parseISO(p.rawDate);
        if (Number.isNaN(date.getTime())) return;
        if (!minDate || date < minDate) minDate = date;
        if (!maxDate || date > maxDate) maxDate = date;

        const monthKey = format(date, 'yyyy-MM');
        const monthLabel = format(date, 'MMM yyyy');
        const amount = getPaymentHomeAmount(p);

        if (!map.has(monthKey)) {
          map.set(monthKey, {
            key: monthKey,
            label: monthLabel,
            amount: 0,
            count: 0,
          });
        }
        const entry = map.get(monthKey);
        entry.amount += amount;
        entry.count += 1;
      } catch {
        // skip invalid date
      }
    });

    if (!minDate || !maxDate) {
      return { chartData: [], totalPeriodSpending: 0, avgMonthlyBurn: 0 };
    }

    const months = eachMonthOfInterval({
      start: startOfMonth(minDate),
      end: startOfMonth(maxDate),
    });

    let cumulative = 0;
    const series = months.map((monthDate) => {
      const monthKey = format(monthDate, 'yyyy-MM');
      const monthLabel = format(monthDate, 'MMM yyyy');
      const existing = map.get(monthKey);
      const monthlyAmount = existing ? Math.round(existing.amount * 100) / 100 : 0;
      cumulative = Math.round((cumulative + monthlyAmount) * 100) / 100;

      return {
        key: monthKey,
        label: monthLabel,
        amount: monthlyAmount,
        cumulative,
        count: existing ? existing.count : 0,
      };
    });

    const avg = series.length > 0 ? Math.round((cumulative / series.length) * 100) / 100 : 0;

    return {
      chartData: series,
      totalPeriodSpending: cumulative,
      avgMonthlyBurn: avg,
    };
  }, [payments]);

  const isCumulative = viewMode === 'cumulative';
  const activeDataKey = isCumulative ? 'cumulative' : 'amount';
  const strokeColor = isCumulative ? '#10b981' : '#3b82f6';
  const gradientId = isCumulative ? 'cumulativeGradient' : 'spendingGradient';

  const referenceValue = isCumulative
    ? (totalTargetBudget > 0 ? totalTargetBudget : null)
    : (avgMonthlyBurn > 0 && chartData.length > 1 ? avgMonthlyBurn : null);

  const referenceLabel = isCumulative
    ? String(content.charts?.monthlyTrend?.targetBudgetLine || 'Target Budget')
    : String(content.charts?.monthlyTrend?.avgBurnLine || 'Avg. Monthly');

  return (
    <Card className="rounded-2xl border-border/70 shadow-xs">
      <CardHeader className="flex flex-col sm:flex-row sm:items-start justify-between pb-2 gap-4">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10 text-primary">
              <TrendingUp className="h-4 w-4" />
            </div>
            <CardTitle className="text-base font-bold tracking-tight">
              {String(content.charts?.monthlyTrend?.title || 'Monthly Investment Trend')}
            </CardTitle>
          </div>
          <CardDescription className="text-xs">
            {String(content.charts?.monthlyTrend?.description || 'Investment timeline in project currency')}
          </CardDescription>
        </div>

        <div className="flex items-center justify-between sm:justify-end gap-4">
          {chartData.length > 0 && (
            <div className="flex items-center gap-1 p-1 bg-muted/50 rounded-xl border border-border/50">
              <button
                type="button"
                onClick={() => setViewMode('monthly')}
                className={cn(
                  'px-2.5 py-1 text-xs font-medium rounded-lg transition-all',
                  !isCumulative
                    ? 'bg-primary text-primary-foreground shadow-xs font-semibold'
                    : 'text-muted-foreground hover:text-foreground',
                )}
              >
                {String(content.charts?.monthlyTrend?.monthlyMode || 'Monthly')}
              </button>
              <button
                type="button"
                onClick={() => setViewMode('cumulative')}
                className={cn(
                  'px-2.5 py-1 text-xs font-medium rounded-lg transition-all',
                  isCumulative
                    ? 'bg-primary text-primary-foreground shadow-xs font-semibold'
                    : 'text-muted-foreground hover:text-foreground',
                )}
              >
                {String(content.charts?.monthlyTrend?.cumulativeMode || 'Cumulative')}
              </button>
            </div>
          )}

          {chartData.length > 0 && (
            <div className="text-right">
              <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider block">
                {isCumulative
                  ? String(content.labels?.payments || 'Total Invested')
                  : String(content.charts?.monthlyTrend?.avgBurnLine || 'Avg. Monthly')}
              </span>
              <span className="text-lg font-extrabold tracking-tight text-foreground font-mono">
                {new Intl.NumberFormat().format(isCumulative ? totalPeriodSpending : avgMonthlyBurn)} {currency}
              </span>
            </div>
          )}
        </div>
      </CardHeader>
      <CardContent className="pt-4">
        {chartData.length === 0 ? (
          <div className="flex h-64 items-center justify-center text-xs text-muted-foreground">
            {String(content.labels?.noData || 'No data')}
          </div>
        ) : (
          <div className="h-72 w-full min-w-0">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={chartData} margin={{ top: 14, right: 16, left: 0, bottom: 0 }}>
                <defs>
                  <linearGradient id={gradientId} x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor={strokeColor} stopOpacity={0.4} />
                    <stop offset="95%" stopColor={strokeColor} stopOpacity={0.0} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" vertical={false} className="stroke-border/40" />
                <XAxis
                  dataKey="label"
                  tickLine={false}
                  axisLine={false}
                  tick={{ fill: 'currentColor', fontSize: 11 }}
                  className="text-muted-foreground font-mono"
                />
                <YAxis
                  tickLine={false}
                  axisLine={false}
                  tick={{ fill: 'currentColor', fontSize: 11 }}
                  className="text-muted-foreground font-mono"
                  tickFormatter={(val) => new Intl.NumberFormat(undefined, { notation: 'compact' }).format(val)}
                />
                <Tooltip content={<ChartTooltip currency={currency} />} />
                {referenceValue !== null && (
                  <ReferenceLine
                    y={referenceValue}
                    stroke="#f59e0b"
                    strokeDasharray="4 4"
                    strokeWidth={1.5}
                    label={{
                      value: `${referenceLabel}: ${new Intl.NumberFormat(undefined, { notation: 'compact' }).format(referenceValue)} ${currency}`,
                      position: 'insideTopRight',
                      fill: '#f59e0b',
                      fontSize: 11,
                    }}
                  />
                )}
                <Area
                  type="monotone"
                  dataKey={activeDataKey}
                  name={
                    isCumulative
                      ? String(content.charts?.monthlyTrend?.cumulativeMode || 'Cumulative')
                      : String(content.charts?.monthlyTrend?.monthlyMode || 'Monthly')
                  }
                  stroke={strokeColor}
                  strokeWidth={2.5}
                  fillOpacity={1}
                  fill={`url(#${gradientId})`}
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

export default MonthlySpendingTrend;

