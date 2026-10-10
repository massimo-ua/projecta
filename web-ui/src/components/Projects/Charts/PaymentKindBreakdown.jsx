import React, { useMemo } from 'react';
import {
  PieChart,
  Pie,
  Cell,
  Tooltip,
  ResponsiveContainer,
} from 'recharts';
import { Layers } from 'lucide-react';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { ChartTooltip } from './ChartTooltip';

export function PaymentKindBreakdown({ payments = [], currency = '', content = {} }) {
  const paymentKindData = useMemo(() => {
    const kindsMap = {
      MONEY: {
        key: 'MONEY',
        name: String(content?.kinds?.money || 'Капітал (Гроші)'),
        amount: 0,
        count: 0,
        fill: '#10b981',
      },
      TIME: {
        key: 'TIME',
        name: String(content?.kinds?.time || 'Праця (Час)'),
        amount: 0,
        count: 0,
        fill: '#f59e0b',
      },
      GOODS: {
        key: 'GOODS',
        name: String(content?.kinds?.goods || 'Товари та матеріали'),
        amount: 0,
        count: 0,
        fill: '#8b5cf6',
      },
      DOWN_PAYMENT: {
        key: 'DOWN_PAYMENT',
        name: String(content.kinds?.downPayment || 'Down Payment'),
        amount: 0,
        count: 0,
        fill: '#f43f5e',
      },
      UPON_COMPLETION: {
        key: 'UPON_COMPLETION',
        name: String(content.kinds?.uponCompletion || 'Direct Payment'),
        amount: 0,
        count: 0,
        fill: '#06b6d4',
      },
      CREDIT_PAYMENT: {
        key: 'CREDIT_PAYMENT',
        name: String(content.kinds?.credit || 'Credit'),
        amount: 0,
        count: 0,
        fill: '#0ea5e9',
      },
    };
    let total = 0;

    payments.forEach((p) => {
      const kindKey = p.kind || p.resourceType || 'MONEY';
      const amount = parseFloat(p.homeAmount) || (p.rawHomeAmount ? p.rawHomeAmount / 100 : 0) || 0;
      total += amount;

      if (kindsMap[kindKey]) {
        kindsMap[kindKey].amount += amount;
        kindsMap[kindKey].count += 1;
      } else {
        if (!kindsMap.OTHER) {
          kindsMap.OTHER = {
            key: 'OTHER',
            name: String(content.kinds?.other || 'Other'),
            amount: 0,
            count: 0,
            fill: '#64748b',
          };
        }
        kindsMap.OTHER.amount += amount;
        kindsMap.OTHER.count += 1;
      }
    });

    return Object.values(kindsMap)
      .filter((k) => k.amount > 0 || k.count > 0)
      .map((k) => ({
        ...k,
        amount: Math.round(k.amount * 100) / 100,
        percentage: total > 0 ? Math.round((k.amount / total) * 100) : 0,
      }));
  }, [payments, content]);

  return (
    <Card className="rounded-2xl border-border/70 shadow-xs flex flex-col">
      <CardHeader className="pb-2">
        <div className="flex items-center gap-2">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-rose-500/10 text-rose-600 dark:text-rose-400">
            <Layers className="h-4 w-4" />
          </div>
          <CardTitle className="text-base font-bold tracking-tight">
            {String(content.charts?.paymentKinds?.title || 'Resource Breakdown')}
          </CardTitle>
        </div>
        <CardDescription className="text-xs">
          {String(content.charts?.paymentKinds?.description || 'Investments across capital, labor hours, and goods')}
        </CardDescription>
      </CardHeader>
      <CardContent className="flex-1 flex flex-col justify-between pt-2">
        {paymentKindData.length === 0 ? (
          <div className="flex h-64 items-center justify-center text-xs text-muted-foreground">
            {String(content.labels?.noData || 'No data')}
          </div>
        ) : (
          <div className="space-y-4">
            <div className="h-56 w-full min-w-0">
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={paymentKindData}
                    dataKey="amount"
                    nameKey="name"
                    cx="50%"
                    cy="50%"
                    innerRadius={55}
                    outerRadius={85}
                    paddingAngle={3}
                  >
                    {paymentKindData.map((entry) => (
                      <Cell key={`cell-${entry.key}`} fill={entry.fill} />
                    ))}
                  </Pie>
                  <Tooltip content={<ChartTooltip currency={currency} />} />
                </PieChart>
              </ResponsiveContainer>
            </div>

            {/* List breakdown */}
            <div className="space-y-2">
              {paymentKindData.map((item) => (
                <div
                  key={item.key}
                  className="flex items-center justify-between text-xs py-1.5 px-2.5 rounded-xl border border-border/50 bg-card/60"
                >
                  <div className="flex items-center gap-2">
                    <span
                      className="h-2.5 w-2.5 rounded-full shrink-0"
                      style={{ backgroundColor: item.fill }}
                    />
                    <span className="font-medium text-foreground">{item.name}</span>
                    <Badge variant="outline" className="text-[10px] px-1.5 py-0 font-mono">
                      {item.count}
                    </Badge>
                  </div>
                  <div className="flex items-center gap-2">
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
