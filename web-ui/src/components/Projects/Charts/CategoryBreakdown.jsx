import React, { useMemo } from 'react';
import {
  PieChart,
  Pie,
  Cell,
  Tooltip,
  ResponsiveContainer,
} from 'recharts';
import { Tag } from 'lucide-react';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card';
import { ChartTooltip } from './ChartTooltip';
import { CHART_COLORS } from './colors';

export function CategoryBreakdown({ payments = [], currency = '', content = {} }) {
  const uncategorizedLabel = String(content.labels?.uncategorized || 'General');

  const categoryData = useMemo(() => {
    const map = new Map();
    let total = 0;

    payments.forEach((p) => {
      const tagList = Array.isArray(p.tags) && p.tags.length > 0
        ? p.tags
        : (p.category ? p.category.split(',').map((s) => s.trim()).filter(Boolean) : [uncategorizedLabel]);

      const effectiveTags = tagList.length > 0 ? tagList : [uncategorizedLabel];
      const amount = parseFloat(p.homeAmount) || (p.rawHomeAmount ? p.rawHomeAmount / 100 : 0) || 0;
      total += amount;

      const sharePerTag = amount / effectiveTags.length;
      effectiveTags.forEach((tag) => {
        const cleanTag = tag.startsWith('#') ? tag : `#${tag}`;
        map.set(cleanTag, (map.get(cleanTag) || 0) + sharePerTag);
      });
    });

    const sorted = Array.from(map.entries())
      .map(([name, amount]) => ({
        name,
        amount: Math.round(amount * 100) / 100,
        percentage: total > 0 ? Math.round((amount / total) * 100) : 0,
      }))
      .sort((a, b) => b.amount - a.amount);

    return sorted;
  }, [payments, uncategorizedLabel]);

  return (
    <Card className="rounded-2xl border-border/70 shadow-xs flex flex-col">
      <CardHeader className="pb-2">
        <div className="flex items-center gap-2">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
            <Tag className="h-4 w-4" />
          </div>
          <CardTitle className="text-base font-bold tracking-tight">
            {String(content.charts?.byCategory?.title || 'Investments by Tag')}
          </CardTitle>
        </div>
        <CardDescription className="text-xs">
          {String(content.charts?.byCategory?.description || 'Breakdown across resource tags')}
        </CardDescription>
      </CardHeader>
      <CardContent className="flex-1 flex flex-col justify-between pt-2">
        {categoryData.length === 0 ? (
          <div className="flex h-64 items-center justify-center text-xs text-muted-foreground">
            {String(content.labels?.noData || 'No data')}
          </div>
        ) : (
          <div className="space-y-4">
            <div className="h-56 w-full min-w-0">
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={categoryData}
                    dataKey="amount"
                    nameKey="name"
                    cx="50%"
                    cy="50%"
                    innerRadius={55}
                    outerRadius={85}
                    paddingAngle={3}
                  >
                    {categoryData.map((entry, index) => (
                      <Cell
                        key={`cell-${entry.name}`}
                        fill={CHART_COLORS[index % CHART_COLORS.length]}
                      />
                    ))}
                  </Pie>
                  <Tooltip content={<ChartTooltip currency={currency} />} />
                </PieChart>
              </ResponsiveContainer>
            </div>

            {/* Category/Tag breakdown legend list */}
            <div className="space-y-1.5 max-h-40 overflow-y-auto pr-1">
              {categoryData.map((item, index) => (
                <div
                  key={item.name}
                  className="flex items-center justify-between text-xs py-1 px-2 rounded-lg hover:bg-muted/50 transition-colors"
                >
                  <div className="flex items-center gap-2 min-w-0">
                    <span
                      className="h-2.5 w-2.5 rounded-full shrink-0"
                      style={{ backgroundColor: CHART_COLORS[index % CHART_COLORS.length] }}
                    />
                    <span className="truncate font-mono font-medium text-foreground" title={item.name}>
                      {item.name}
                    </span>
                  </div>
                  <div className="flex items-center gap-2 shrink-0 ml-2">
                    <span className="text-[11px] font-mono text-muted-foreground">
                      {item.percentage}%
                    </span>
                    <span className="font-mono font-semibold text-foreground">
                      {new Intl.NumberFormat().format(item.amount)} {currency}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

export default CategoryBreakdown;
