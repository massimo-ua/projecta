import React, { useMemo, useState } from 'react';
import {
  PieChart,
  Pie,
  Cell,
  Tooltip,
  ResponsiveContainer,
} from 'recharts';
import { Layers, Users } from 'lucide-react';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';
import { getPaymentHomeAmount } from '../../../hooks/useProjectAnalytics';
import { ChartTooltip } from './ChartTooltip';
import { CHART_COLORS } from './colors';

export function PaymentKindBreakdown({
  payments = [],
  project = null,
  currency = '',
  content = {},
}) {
  const [breakdownMode, setBreakdownMode] = useState('resource');
  const hoursUnit = String(content.labels?.hoursUnit || 'hrs');

  const paymentKindData = useMemo(() => {
    const kindsMap = {
      MONEY: {
        key: 'MONEY',
        name: String(content?.kinds?.money || 'Capital (Money)'),
        amount: 0,
        count: 0,
        hours: 0,
        fill: '#10b981',
      },
      TIME: {
        key: 'TIME',
        name: String(content?.kinds?.time || 'Labor (Time)'),
        amount: 0,
        count: 0,
        hours: 0,
        fill: '#f59e0b',
      },
      GOODS: {
        key: 'GOODS',
        name: String(content?.kinds?.goods || 'Goods & Materials'),
        amount: 0,
        count: 0,
        hours: 0,
        fill: '#8b5cf6',
      },
      DOWN_PAYMENT: {
        key: 'DOWN_PAYMENT',
        name: String(content.kinds?.downPayment || 'Down Payment'),
        amount: 0,
        count: 0,
        hours: 0,
        fill: '#f43f5e',
      },
      UPON_COMPLETION: {
        key: 'UPON_COMPLETION',
        name: String(content.kinds?.uponCompletion || 'Direct Investment'),
        amount: 0,
        count: 0,
        hours: 0,
        fill: '#06b6d4',
      },
      CREDIT_PAYMENT: {
        key: 'CREDIT_PAYMENT',
        name: String(content.kinds?.credit || 'Credit'),
        amount: 0,
        count: 0,
        hours: 0,
        fill: '#0ea5e9',
      },
    };
    let total = 0;

    payments.forEach((p) => {
      const kindKey = p.kind || p.resourceType || 'MONEY';
      const amount = getPaymentHomeAmount(p);
      const hrs = kindKey === 'TIME' ? (Number(p.timeHours) || 0) : 0;
      total += amount;

      if (kindsMap[kindKey]) {
        kindsMap[kindKey].amount += amount;
        kindsMap[kindKey].count += 1;
        kindsMap[kindKey].hours += hrs;
      } else {
        if (!kindsMap.OTHER) {
          kindsMap.OTHER = {
            key: 'OTHER',
            name: String(content.kinds?.other || 'Other'),
            amount: 0,
            count: 0,
            hours: 0,
            fill: '#64748b',
          };
        }
        kindsMap.OTHER.amount += amount;
        kindsMap.OTHER.count += 1;
        kindsMap.OTHER.hours += hrs;
      }
    });

    return Object.values(kindsMap)
      .filter((k) => k.amount > 0 || k.count > 0)
      .map((k) => ({
        ...k,
        amount: Math.round(k.amount * 100) / 100,
        hours: Math.round(k.hours * 10) / 10,
        percentage: total > 0 ? Math.round((k.amount / total) * 100) : 0,
      }));
  }, [payments, content]);

  const contributorData = useMemo(() => {
    const map = new Map();
    let total = 0;
    const defaultName =
      project?.owner?.name ||
      String(content.labels?.unassignedContributor || 'Project Owner');

    payments.forEach((p) => {
      const name = p.contributorName || defaultName;
      const amount = getPaymentHomeAmount(p);
      const hrs = p.resourceType === 'TIME' ? (Number(p.timeHours) || 0) : 0;
      total += amount;

      if (!map.has(name)) {
        map.set(name, {
          key: name,
          name,
          amount: 0,
          count: 0,
          hours: 0,
        });
      }
      const entry = map.get(name);
      entry.amount += amount;
      entry.count += 1;
      entry.hours += hrs;
    });

    return Array.from(map.values())
      .sort((a, b) => b.amount - a.amount)
      .map((item, idx) => ({
        ...item,
        amount: Math.round(item.amount * 100) / 100,
        hours: Math.round(item.hours * 10) / 10,
        percentage: total > 0 ? Math.round((item.amount / total) * 100) : 0,
        fill: CHART_COLORS[idx % CHART_COLORS.length],
      }));
  }, [payments, project, content]);

  const activeData = breakdownMode === 'contributor' ? contributorData : paymentKindData;
  const HeaderIcon = breakdownMode === 'contributor' ? Users : Layers;

  return (
    <Card className="rounded-2xl border-border/70 shadow-xs flex flex-col">
      <CardHeader className="flex flex-col sm:flex-row sm:items-start justify-between pb-2 gap-3">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-rose-500/10 text-rose-600 dark:text-rose-400">
              <HeaderIcon className="h-4 w-4" />
            </div>
            <CardTitle className="text-base font-bold tracking-tight">
              {String(content.charts?.paymentKinds?.title || 'Resource & Team Breakdown')}
            </CardTitle>
          </div>
          <CardDescription className="text-xs">
            {String(content.charts?.paymentKinds?.description || 'Capital, labor hours, goods, and contributor shares')}
          </CardDescription>
        </div>

        <div className="flex items-center gap-1 p-1 bg-muted/50 rounded-xl border border-border/50 self-start">
          <button
            type="button"
            onClick={() => setBreakdownMode('resource')}
            className={cn(
              'px-2.5 py-1 text-xs font-medium rounded-lg transition-all whitespace-nowrap',
              breakdownMode === 'resource'
                ? 'bg-primary text-primary-foreground shadow-xs font-semibold'
                : 'text-muted-foreground hover:text-foreground',
            )}
          >
            {String(content.charts?.paymentKinds?.byResource || 'By Resource')}
          </button>
          <button
            type="button"
            onClick={() => setBreakdownMode('contributor')}
            className={cn(
              'px-2.5 py-1 text-xs font-medium rounded-lg transition-all whitespace-nowrap',
              breakdownMode === 'contributor'
                ? 'bg-primary text-primary-foreground shadow-xs font-semibold'
                : 'text-muted-foreground hover:text-foreground',
            )}
          >
            {String(content.charts?.paymentKinds?.byContributor || 'By Contributor')}
          </button>
        </div>
      </CardHeader>
      <CardContent className="flex-1 flex flex-col justify-between pt-2">
        {activeData.length === 0 ? (
          <div className="flex h-64 items-center justify-center text-xs text-muted-foreground">
            {String(content.labels?.noData || 'No data')}
          </div>
        ) : (
          <div className="space-y-4">
            <div className="h-56 w-full min-w-0">
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={activeData}
                    dataKey="amount"
                    nameKey="name"
                    cx="50%"
                    cy="50%"
                    innerRadius={55}
                    outerRadius={85}
                    paddingAngle={3}
                  >
                    {activeData.map((entry) => (
                      <Cell key={`cell-${entry.key}`} fill={entry.fill} />
                    ))}
                  </Pie>
                  <Tooltip content={<ChartTooltip currency={currency} />} />
                </PieChart>
              </ResponsiveContainer>
            </div>

            {/* List breakdown */}
            <div className="space-y-2 max-h-40 overflow-y-auto pr-1">
              {activeData.map((item) => (
                <div
                  key={item.key}
                  className="flex items-center justify-between text-xs py-1.5 px-2.5 rounded-xl border border-border/50 bg-card/60"
                >
                  <div className="flex items-center gap-2 min-w-0">
                    <span
                      className="h-2.5 w-2.5 rounded-full shrink-0"
                      style={{ backgroundColor: item.fill }}
                    />
                    <span className="font-medium text-foreground truncate" title={item.name}>
                      {item.name}
                    </span>
                    <Badge variant="outline" className="text-[10px] px-1.5 py-0 font-mono shrink-0">
                      {item.count}
                    </Badge>
                    {item.hours > 0 && (
                      <Badge variant="secondary" className="text-[10px] px-1.5 py-0 font-mono shrink-0">
                        {item.hours} {hoursUnit}
                      </Badge>
                    )}
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

export default PaymentKindBreakdown;

