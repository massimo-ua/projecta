import React from 'react';
import PropTypes from 'prop-types';
import {
  ResponsiveContainer,
  ComposedChart,
  Bar,
  Line,
  Cell,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
} from 'recharts';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';

function formatMoneyMajor(majorVal, currency) {
  const num = Number(majorVal) || 0;
  return `${num.toLocaleString(undefined, { minimumFractionDigits: 0, maximumFractionDigits: 2 })} ${currency}`;
}

function getDemandBarColor(alertStatus) {
  if (alertStatus === 'OVER_CEILING') return '#ef4444';
  if (alertStatus === 'AT_RISK') return '#f59e0b';
  return '#3b82f6';
}

function CashFlowTooltip({
  active,
  payload,
  currency,
  includeOngoingInCommitted,
}) {
  if (!active || !payload || !payload.length) return null;
  const row = payload[0]?.payload;
  if (!row) return null;

  return (
    <div className="rounded-xl border border-border bg-card p-3 shadow-lg text-xs space-y-2 max-w-xs">
      <div className="flex items-center justify-between gap-3 border-b border-border/60 pb-1.5">
        <span className="font-bold text-sm text-foreground">{row.month}</span>
        {row.alertStatus !== 'NO_CEILING' && (
          <Badge
            variant="outline"
            className="text-[10px] py-0 px-1.5"
          >
            {row.alertStatus} ({row.utilizationPct}%)
          </Badge>
        )}
      </div>

      <div className="space-y-1">
        <div className="flex justify-between gap-4">
          <span className="text-muted-foreground">Actual Cash:</span>
          <span className="font-mono font-semibold text-emerald-600 dark:text-emerald-400">
            {formatMoneyMajor(row.actualCash, currency)}
          </span>
        </div>
        <div className="flex justify-between gap-4">
          <span className="text-muted-foreground">Scheduled Demand:</span>
          <span className="font-mono font-semibold text-blue-600 dark:text-blue-400">
            {formatMoneyMajor(row.scheduledDemand, currency)}
          </span>
        </div>
        {row.estimatedOngoing > 0 && (
          <div className="flex justify-between gap-4">
            <span className="text-muted-foreground">
              {includeOngoingInCommitted ? 'Ongoing Burn (Stacked):' : 'Ongoing Burn (Info):'}
            </span>
            <span className="font-mono text-slate-500">
              {formatMoneyMajor(row.estimatedOngoing, currency)}
            </span>
          </div>
        )}
        {row.ceiling > 0 && (
          <div className="flex justify-between gap-4 pt-1 border-t border-border/50">
            <span className="text-muted-foreground">Monthly Ceiling:</span>
            <span className="font-mono font-semibold text-foreground">
              {formatMoneyMajor(row.ceiling, currency)}
            </span>
          </div>
        )}
      </div>

      {(row.timeHours > 0 || row.goods > 0) && (
        <div className="pt-1 border-t border-border/50 text-[11px] text-muted-foreground flex justify-between gap-3">
          <span>Non-Cash: {row.timeHours.toFixed(1)} hrs</span>
          <span>{formatMoneyMajor(row.goods, currency)} goods</span>
        </div>
      )}

      {row.assetDemands && row.assetDemands.length > 0 && (
        <div className="pt-1 border-t border-border/50 space-y-0.5">
          <p className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
            Scheduled Drivers
          </p>
          {row.assetDemands.slice(0, 5).map((d) => (
            <div key={d.assetId} className="flex justify-between gap-2 text-[11px]">
              <span className="truncate">
                {d.assetName}
                {d.isOverdue ? ' (overdue)' : ''}
              </span>
              <span className="font-mono shrink-0">
                {formatMoneyMajor(d.demandCents / 100, currency)}
              </span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

CashFlowTooltip.propTypes = {
  active: PropTypes.bool,
  payload: PropTypes.array,
  currency: PropTypes.string.isRequired,
  includeOngoingInCommitted: PropTypes.bool.isRequired,
};

export function MonthlyCashFlowChart({
  chartData,
  currency,
  includeOngoingInCommitted,
  title,
  subtitle,
}) {
  return (
    <Card className="border-border/60 shadow-sm">
      <CardHeader className="pb-2">
        <CardTitle className="text-base font-bold">{title}</CardTitle>
        <CardDescription className="text-xs">{subtitle}</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="h-[340px] w-full">
          <ResponsiveContainer width="100%" height="100%">
            <ComposedChart data={chartData} margin={{ top: 12, right: 16, left: 4, bottom: 8 }}>
              <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="rgba(148, 163, 184, 0.25)" />
              <XAxis
                dataKey="month"
                tick={{ fontSize: 12 }}
                axisLine={false}
                tickLine={false}
              />
              <YAxis
                tick={{ fontSize: 11 }}
                axisLine={false}
                tickLine={false}
                width={64}
              />
              <Tooltip
                content={(
                  <CashFlowTooltip
                    currency={currency}
                    includeOngoingInCommitted={includeOngoingInCommitted}
                  />
                )}
              />
              <Legend wrapperStyle={{ fontSize: '12px' }} />
              <Bar
                dataKey="actualCash"
                name="Actual Cash"
                stackId="cash"
                fill="#10b981"
                radius={[0, 0, 0, 0]}
              />
              <Bar
                dataKey="scheduledDemand"
                name="Scheduled Demand"
                stackId="cash"
                radius={[4, 4, 0, 0]}
              >
                {chartData.map((entry) => (
                  <Cell key={entry.month} fill={getDemandBarColor(entry.alertStatus)} />
                ))}
              </Bar>
              {includeOngoingInCommitted && (
                <Bar
                  dataKey="estimatedOngoing"
                  name="Ongoing Burn Preview"
                  stackId="cash"
                  fill="#94a3b8"
                  fillOpacity={0.65}
                  radius={[4, 4, 0, 0]}
                />
              )}
              <Line
                type="stepAfter"
                dataKey="ceiling"
                name="TagBudget Ceiling"
                stroke="#f97316"
                strokeWidth={2}
                strokeDasharray="5 5"
                dot={false}
              />
            </ComposedChart>
          </ResponsiveContainer>
        </div>
      </CardContent>
    </Card>
  );
}

MonthlyCashFlowChart.propTypes = {
  chartData: PropTypes.array.isRequired,
  currency: PropTypes.string.isRequired,
  includeOngoingInCommitted: PropTypes.bool.isRequired,
  title: PropTypes.string.isRequired,
  subtitle: PropTypes.string.isRequired,
};
