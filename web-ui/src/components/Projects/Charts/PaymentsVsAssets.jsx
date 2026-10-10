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
import { ChartTooltip } from './ChartTooltip';

export function PaymentsVsAssets({ payments = [], assets = [], currency = '', content = {} }) {
  const { chartData, totalPayments, totalAssets, netBalance } = useMemo(() => {
    const pSum = payments.reduce((sum, p) => {
      const amount = parseFloat(p.homeAmount) || (p.rawHomeAmount ? p.rawHomeAmount / 100 : 0) || 0;
      return sum + amount;
    }, 0);

    const aSum = assets.reduce((sum, a) => {
      const price = parseFloat(a.homeAmount) || parseFloat(a.price) || 0;
      return sum + price;
    }, 0);

    const roundedP = Math.round(pSum * 100) / 100;
    const roundedA = Math.round(aSum * 100) / 100;
    const roundedNet = Math.round((pSum - aSum) * 100) / 100;

    return {
      totalPayments: roundedP,
      totalAssets: roundedA,
      netBalance: roundedNet,
      chartData: [
        {
          name: String(content.labels?.payments || 'Payments'),
          amount: roundedP,
          fill: '#3b82f6',
        },
        {
          name: String(content.labels?.assets || 'Assets'),
          amount: roundedA,
          fill: '#10b981',
        },
      ],
    };
  }, [payments, assets, content]);

  const hasData = totalPayments > 0 || totalAssets > 0;

  return (
    <Card className="rounded-2xl border-border/70 shadow-xs flex flex-col">
      <CardHeader className="pb-2">
        <div className="flex items-center gap-2">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-sky-500/10 text-sky-600 dark:text-sky-400">
            <Scale className="h-4 w-4" />
          </div>
          <CardTitle className="text-base font-bold tracking-tight">
            {String(content.charts?.paymentsVsAssets?.title || 'Payments vs. Assets')}
          </CardTitle>
        </div>
        <CardDescription className="text-xs">
          {String(content.charts?.paymentsVsAssets?.description || 'Comparison of cash spent and asset values')}
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
                <BarChart data={chartData} margin={{ top: 20, right: 20, left: 0, bottom: 5 }}>
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
                  <Bar dataKey="amount" radius={[8, 8, 0, 0]}>
                    {chartData.map((entry) => (
                      <Cell key={`bar-${entry.name}`} fill={entry.fill} />
                    ))}
                  </Bar>
                </BarChart>
              </ResponsiveContainer>
            </div>

            {/* Metrics cards */}
            <div className="grid grid-cols-3 gap-2 text-xs">
              <div className="p-2.5 rounded-xl border border-border/60 bg-card/60">
                <span className="text-[11px] font-medium text-muted-foreground block truncate">
                  {String(content.labels?.payments || 'Payments')}
                </span>
                <span className="font-mono font-bold text-foreground text-xs sm:text-sm">
                  {new Intl.NumberFormat().format(totalPayments)} {currency}
                </span>
              </div>
              <div className="p-2.5 rounded-xl border border-border/60 bg-card/60">
                <span className="text-[11px] font-medium text-muted-foreground block truncate">
                  {String(content.labels?.assets || 'Assets')}
                </span>
                <span className="font-mono font-bold text-foreground text-xs sm:text-sm">
                  {new Intl.NumberFormat().format(totalAssets)} {currency}
                </span>
              </div>
              <div className="p-2.5 rounded-xl border border-border/60 bg-card/60">
                <span className="text-[11px] font-medium text-muted-foreground block truncate">
                  {String(content.labels?.balance || 'Баланс проекту')}
                </span>
                <span className="font-mono font-bold text-foreground text-xs sm:text-sm">
                  {new Intl.NumberFormat().format(netBalance)} {currency}
                </span>
              </div>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

export default PaymentsVsAssets;
