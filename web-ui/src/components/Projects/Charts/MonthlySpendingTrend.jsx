import React, { useMemo } from 'react';
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts';
import { format, parseISO } from 'date-fns';
import { TrendingUp } from 'lucide-react';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card';
import { ChartTooltip } from './ChartTooltip';

export function MonthlySpendingTrend({ payments = [], currency = '', content = {} }) {
  const chartData = useMemo(() => {
    const map = new Map();
    const sorted = [...payments].sort((a, b) => {
      const dateA = a.rawDate || '';
      const dateB = b.rawDate || '';
      return dateA.localeCompare(dateB);
    });

    sorted.forEach((p) => {
      if (!p.rawDate) return;
      try {
        const date = parseISO(p.rawDate);
        if (Number.isNaN(date.getTime())) return;
        const monthKey = format(date, 'yyyy-MM');
        const monthLabel = format(date, 'MMM yyyy');
        const amount = parseFloat(p.homeAmount) || (p.rawHomeAmount ? p.rawHomeAmount / 100 : 0) || 0;

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

    return Array.from(map.values()).map((item) => ({
      ...item,
      amount: Math.round(item.amount * 100) / 100,
    }));
  }, [payments]);

  const totalPeriodSpending = useMemo(() => {
    return chartData.reduce((sum, item) => sum + item.amount, 0);
  }, [chartData]);

  return (
    <Card className="rounded-2xl border-border/70 shadow-xs">
      <CardHeader className="flex flex-row items-start justify-between pb-2 gap-4">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10 text-primary">
              <TrendingUp className="h-4 w-4" />
            </div>
            <CardTitle className="text-base font-bold tracking-tight">
              {String(content.charts?.monthlyTrend?.title || 'Monthly Spending Trend')}
            </CardTitle>
          </div>
          <CardDescription className="text-xs">
            {String(content.charts?.monthlyTrend?.description || 'Spending timeline in project currency')}
          </CardDescription>
        </div>
        {chartData.length > 0 && (
          <div className="text-right">
            <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider block">
              {String(content.labels?.amount || 'Amount')}
            </span>
            <span className="text-lg font-extrabold tracking-tight text-foreground font-mono">
              {new Intl.NumberFormat().format(totalPeriodSpending)} {currency}
            </span>
          </div>
        )}
      </CardHeader>
      <CardContent className="pt-4">
        {chartData.length === 0 ? (
          <div className="flex h-64 items-center justify-center text-xs text-muted-foreground">
            {String(content.labels?.noData || 'No data')}
          </div>
        ) : (
          <div className="h-72 w-full min-w-0">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={chartData} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
                <defs>
                  <linearGradient id="spendingGradient" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.4} />
                    <stop offset="95%" stopColor="#3b82f6" stopOpacity={0.0} />
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
                <Area
                  type="monotone"
                  dataKey="amount"
                  name={String(content.charts?.monthlyTrend?.title || 'Spending')}
                  stroke="#3b82f6"
                  strokeWidth={2.5}
                  fillOpacity={1}
                  fill="url(#spendingGradient)"
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
