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
import { Boxes } from 'lucide-react';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card';
import { ChartTooltip } from './ChartTooltip';
import { CHART_COLORS } from './colors';

export function TopCostTypes({ payments = [], currency = '', content = {} }) {
  const uncategorizedLabel = String(content.labels?.uncategorized || 'Uncategorized');
  const otherLabel = String(content.labels?.other || 'Other');

  const topCostTypesData = useMemo(() => {
    const map = new Map();

    payments.forEach((p) => {
      const typeName = p.type || uncategorizedLabel;
      const amount = parseFloat(p.homeAmount) || (p.rawHomeAmount ? p.rawHomeAmount / 100 : 0) || 0;
      map.set(typeName, (map.get(typeName) || 0) + amount);
    });

    const sorted = Array.from(map.entries())
      .map(([name, amount]) => ({
        name,
        amount: Math.round(amount * 100) / 100,
      }))
      .sort((a, b) => b.amount - a.amount);

    if (sorted.length <= 8) {
      return sorted;
    }

    const top8 = sorted.slice(0, 8);
    const otherAmount = sorted.slice(8).reduce((sum, item) => sum + item.amount, 0);

    return [
      ...top8,
      {
        name: otherLabel,
        amount: Math.round(otherAmount * 100) / 100,
      },
    ];
  }, [payments, uncategorizedLabel, otherLabel]);

  return (
    <Card className="rounded-2xl border-border/70 shadow-xs flex flex-col">
      <CardHeader className="pb-2">
        <div className="flex items-center gap-2">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-indigo-500/10 text-indigo-600 dark:text-indigo-400">
            <Boxes className="h-4 w-4" />
          </div>
          <CardTitle className="text-base font-bold tracking-tight">
            {String(content.charts?.topTypes?.title || 'Top Cost Types')}
          </CardTitle>
        </div>
        <CardDescription className="text-xs">
          {String(content.charts?.topTypes?.description || 'Top spending subcategories')}
        </CardDescription>
      </CardHeader>
      <CardContent className="flex-1 flex flex-col justify-between pt-2">
        {topCostTypesData.length === 0 ? (
          <div className="flex h-64 items-center justify-center text-xs text-muted-foreground">
            {String(content.labels?.noData || 'No data')}
          </div>
        ) : (
          <div className="h-72 w-full min-w-0">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart
                layout="vertical"
                data={topCostTypesData}
                margin={{ top: 10, right: 30, left: 10, bottom: 5 }}
              >
                <CartesianGrid strokeDasharray="3 3" horizontal={false} className="stroke-border/40" />
                <XAxis
                  type="number"
                  tickLine={false}
                  axisLine={false}
                  tick={{ fill: 'currentColor', fontSize: 11 }}
                  className="text-muted-foreground font-mono"
                  tickFormatter={(val) => new Intl.NumberFormat(undefined, { notation: 'compact' }).format(val)}
                />
                <YAxis
                  type="category"
                  dataKey="name"
                  tickLine={false}
                  axisLine={false}
                  width={110}
                  tick={{ fill: 'currentColor', fontSize: 11 }}
                  className="text-muted-foreground"
                />
                <Tooltip content={<ChartTooltip currency={currency} />} />
                <Bar dataKey="amount" radius={[0, 6, 6, 0]}>
                  {topCostTypesData.map((entry, index) => (
                    <Cell
                      key={`bar-${entry.name}`}
                      fill={entry.name === otherLabel ? '#94a3b8' : CHART_COLORS[index % CHART_COLORS.length]}
                    />
                  ))}
                </Bar>
              </BarChart>
            </ResponsiveContainer>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

export default TopCostTypes;
