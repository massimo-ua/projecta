import React, { useMemo, useState } from 'react';
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
import { cn } from '@/lib/utils';
import { getPaymentHomeAmount, getAssetHomeAmount } from '../../../hooks/useProjectAnalytics';
import { ChartTooltip } from './ChartTooltip';
import { CHART_COLORS } from './colors';

export function TopCostTypes({
  payments = [],
  assets = [],
  currency = '',
  content = {},
}) {
  const [groupingMode, setGroupingMode] = useState('direct');
  const uncategorizedLabel = String(content.labels?.uncategorized || 'General');
  const otherLabel = String(content.labels?.other || 'Other');

  const topCostTypesData = useMemo(() => {
    const assetById = new Map(assets.map((a) => [a.id, a]));
    const map = new Map();

    // Helper to resolve root ancestor shares for a given assetId
    const resolveRootShares = (assetId, visited = new Set()) => {
      if (!assetId || visited.has(assetId)) {
        return [{ name: assetById.get(assetId)?.name || uncategorizedLabel, weight: 1 }];
      }
      visited.add(assetId);
      const assetObj = assetById.get(assetId);
      if (!assetObj) {
        return [{ name: uncategorizedLabel, weight: 1 }];
      }
      if (!Array.isArray(assetObj.parents) || assetObj.parents.length === 0) {
        return [{ name: assetObj.name || uncategorizedLabel, weight: 1 }];
      }

      const rootShares = [];
      assetObj.parents.forEach((pLink) => {
        const shareWeight = (Number(pLink.sharePercentage) || 100) / 100;
        const parentRoots = resolveRootShares(pLink.parentId, new Set(visited));
        parentRoots.forEach((pr) => {
          rootShares.push({
            name: pr.name || pLink.parentName || uncategorizedLabel,
            weight: shareWeight * pr.weight,
          });
        });
      });
      return rootShares.length > 0
        ? rootShares
        : [{ name: assetObj.name || uncategorizedLabel, weight: 1 }];
    };

    const addAmount = (name, amount) => {
      if (amount <= 0) return;
      const cleanName = name || uncategorizedLabel;
      map.set(cleanName, (map.get(cleanName) || 0) + amount);
    };

    if (payments.length > 0) {
      payments.forEach((p) => {
        const totalPaymentAmt = getPaymentHomeAmount(p);
        if (totalPaymentAmt <= 0) return;

        const allocations = Array.isArray(p.assetAllocations) && p.assetAllocations.length > 0
          ? p.assetAllocations.map((a) => ({
              assetId: a.assetId,
              assetName: a.assetName || assetById.get(a.assetId)?.name,
              weight: (Number(a.sharePercentage) || 100) / 100,
            }))
          : [
              {
                assetId: p.assetId,
                assetName: p.assetName || assetById.get(p.assetId)?.name,
                weight: 1,
              },
            ];

        allocations.forEach((alloc) => {
          const allocatedAmt = totalPaymentAmt * alloc.weight;
          if (groupingMode === 'groups' && alloc.assetId) {
            const roots = resolveRootShares(alloc.assetId);
            roots.forEach((r) => addAmount(r.name, allocatedAmt * r.weight));
          } else {
            addAmount(alloc.assetName || uncategorizedLabel, allocatedAmt);
          }
        });
      });
    } else if (assets.length > 0) {
      // Fallback to asset valuations when no payments are present in list
      const sourceAssets = groupingMode === 'groups'
        ? assets.filter((a) => !a.parents || a.parents.length === 0)
        : assets.filter((a) => !a.children || a.children.length === 0);

      sourceAssets.forEach((a) => {
        addAmount(a.name, getAssetHomeAmount(a));
      });
    }

    const totalSum = Array.from(map.values()).reduce((s, v) => s + v, 0);

    const sorted = Array.from(map.entries())
      .map(([name, amount]) => ({
        name,
        amount: Math.round(amount * 100) / 100,
        percentage: totalSum > 0 ? Math.round((amount / totalSum) * 100) : 0,
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
        percentage: totalSum > 0 ? Math.round((otherAmount / totalSum) * 100) : 0,
      },
    ];
  }, [payments, assets, groupingMode, uncategorizedLabel, otherLabel]);

  return (
    <Card className="rounded-2xl border-border/70 shadow-xs flex flex-col">
      <CardHeader className="flex flex-col sm:flex-row sm:items-start justify-between pb-2 gap-3">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-indigo-500/10 text-indigo-600 dark:text-indigo-400">
              <Boxes className="h-4 w-4" />
            </div>
            <CardTitle className="text-base font-bold tracking-tight">
              {String(content.charts?.topTypes?.title || 'Top Assets by Investment')}
            </CardTitle>
          </div>
          <CardDescription className="text-xs">
            {String(content.charts?.topTypes?.description || 'Assets with highest resource allocation')}
          </CardDescription>
        </div>

        <div className="flex items-center gap-1 p-1 bg-muted/50 rounded-xl border border-border/50 self-start">
          <button
            type="button"
            onClick={() => setGroupingMode('direct')}
            className={cn(
              'px-2.5 py-1 text-xs font-medium rounded-lg transition-all whitespace-nowrap',
              groupingMode === 'direct'
                ? 'bg-primary text-primary-foreground shadow-xs font-semibold'
                : 'text-muted-foreground hover:text-foreground',
            )}
          >
            {String(content.charts?.topTypes?.directAssets || 'Direct Assets')}
          </button>
          <button
            type="button"
            onClick={() => setGroupingMode('groups')}
            className={cn(
              'px-2.5 py-1 text-xs font-medium rounded-lg transition-all whitespace-nowrap',
              groupingMode === 'groups'
                ? 'bg-primary text-primary-foreground shadow-xs font-semibold'
                : 'text-muted-foreground hover:text-foreground',
            )}
          >
            {String(content.charts?.topTypes?.rootGroups || 'Root Groups')}
          </button>
        </div>
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
                  width={120}
                  tick={{ fill: 'currentColor', fontSize: 11 }}
                  className="text-muted-foreground"
                />
                <Tooltip content={<ChartTooltip currency={currency} />} />
                <Bar
                  dataKey="amount"
                  name={String(content.labels?.amount || 'Amount')}
                  radius={[0, 6, 6, 0]}
                >
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

