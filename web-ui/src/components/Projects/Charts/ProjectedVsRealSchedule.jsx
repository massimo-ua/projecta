import React, { useMemo, useState } from 'react';
import PropTypes from 'prop-types';
import { useNavigate, useParams } from 'react-router-dom';
import { useLocale } from 'react-intlayer';
import { getLocalizedUrl } from 'intlayer';
import {
  ResponsiveContainer,
  BarChart,
  Bar,
  Cell,
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
} from 'recharts';
import {
  CalendarRange,
  ArrowUpRight,
  AlertTriangle,
  CheckCircle2,
  Tag,
  SlidersHorizontal,
} from 'lucide-react';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { cn } from '@/lib/utils';

export function ProjectedVsRealSchedule({ budgetSchedule, currency = 'UAH', content = {} }) {
  const { projectId } = useParams();
  const { locale } = useLocale();
  const navigate = useNavigate();

  const [viewMode, setViewMode] = useState('monthly'); // 'monthly' | 'cumulative' | 'byTag'
  const [selectedTag, setSelectedTag] = useState('ALL');
  const [showTable, setShowTable] = useState(false);

  const t = content.charts?.projectedSchedule || {};

  const hasTagBudgets = Boolean(budgetSchedule && budgetSchedule.tagRules && budgetSchedule.tagRules.length > 0);

  const availableTags = useMemo(() => {
    if (!budgetSchedule) return [];
    return budgetSchedule.tagRules.map((r) => r.tag);
  }, [budgetSchedule]);

  // Compute filtered series data if a specific tag is filtered
  const chartData = useMemo(() => {
    if (!budgetSchedule || !budgetSchedule.months) return [];

    let cumProj = 0;
    let cumAct = 0;

    return budgetSchedule.months.map((m) => {
      let proj = m.projected;
      let act = m.actual;

      if (selectedTag !== 'ALL') {
        proj = m.byTag[selectedTag] || 0;
        act = m.actualByTag[selectedTag] || 0;
      }

      cumProj = Math.round((cumProj + proj) * 100) / 100;
      cumAct = Math.round((cumAct + act) * 100) / 100;

      return {
        key: m.key,
        label: m.label,
        scheduleIndex: m.scheduleIndex,
        projected: proj,
        actual: act,
        cumulativeProjected: cumProj,
        cumulativeActual: cumAct,
        variance: Math.round((proj - act) * 100) / 100,
        isOverBudget: act > proj,
      };
    });
  }, [budgetSchedule, selectedTag]);

  // Tag breakdown data
  const tagBreakdownData = useMemo(() => {
    if (!budgetSchedule || !budgetSchedule.tagComparisons) return [];
    return budgetSchedule.tagComparisons.map((c) => ({
      name: c.isUnplanned ? String(t.unplannedTagLabel || '#unplanned') : c.displayTag,
      tag: c.tag,
      planned: c.projectedAnnual,
      actual: c.actualAnnual,
      variance: c.variance,
      utilization: c.utilizationPercentage,
      isOverBudget: c.isOverBudget,
      isUnplanned: c.isUnplanned,
    }));
  }, [budgetSchedule, t]);

  const summary = useMemo(() => {
    if (!budgetSchedule) {
      return {
        planned: 0,
        actual: 0,
        variance: 0,
        utilization: 0,
        isOverBudget: false,
      };
    }

    if (selectedTag === 'ALL') {
      return {
        planned: budgetSchedule.totalAnnualBudget,
        actual: budgetSchedule.totalActualInvested,
        variance: budgetSchedule.annualVariance,
        utilization: budgetSchedule.budgetUtilization,
        isOverBudget: budgetSchedule.isOverAnnualBudget,
      };
    }

    const filteredPlanned = chartData.reduce((sum, d) => sum + d.projected, 0);
    const filteredActual = chartData.reduce((sum, d) => sum + d.actual, 0);
    const filteredVariance = Math.round((filteredPlanned - filteredActual) * 100) / 100;
    const filteredUtil = filteredPlanned > 0
      ? Math.round((filteredActual / filteredPlanned) * 100)
      : 0;

    return {
      planned: Math.round(filteredPlanned * 100) / 100,
      actual: Math.round(filteredActual * 100) / 100,
      variance: filteredVariance,
      utilization: filteredUtil,
      isOverBudget: filteredActual > filteredPlanned,
    };
  }, [budgetSchedule, selectedTag, chartData]);

  if (!hasTagBudgets) {
    return (
      <Card className="rounded-2xl border-border/70 shadow-xs">
        <CardHeader>
          <div className="flex items-center gap-2">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10 text-primary">
              <CalendarRange className="h-4 w-4" />
            </div>
            <CardTitle className="text-base font-bold tracking-tight">
              {String(t.title || 'Projected vs. Real Annual Investment Schedule')}
            </CardTitle>
          </div>
          <CardDescription className="text-xs">
            {String(t.description || '12-month comparison of planned tag budgets against real investments')}
          </CardDescription>
        </CardHeader>
        <CardContent className="p-8 text-center space-y-4">
          <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-muted mx-auto text-muted-foreground">
            <Tag className="h-7 w-7" />
          </div>
          <div className="max-w-md mx-auto space-y-1">
            <h4 className="text-sm font-bold text-foreground">
              {String(t.noBudgetEmptyTitle || 'No Category Budgets Configured Yet')}
            </h4>
            <p className="text-xs text-muted-foreground leading-relaxed">
              {String(t.noBudgetEmptyDesc || 'Define expected investment frequencies and amounts per tag to compare your projected annual schedule with real investments.')}
            </p>
          </div>
          {projectId && (
            <Button
              onClick={() => navigate(getLocalizedUrl(`/projects/${projectId}/tags`, locale))}
              className="gap-2 rounded-xl text-xs font-semibold h-9 px-4"
            >
              <span>{String(t.configureInTagsButton || 'Configure Category Budgets')}</span>
              <ArrowUpRight className="h-3.5 w-3.5" />
            </Button>
          )}
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="rounded-2xl border-border/70 shadow-xs">
      <CardHeader className="pb-3 border-b border-border/40">
        <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
          <div>
            <div className="flex items-center gap-2">
              <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10 text-primary">
                <CalendarRange className="h-4 w-4" />
              </div>
              <CardTitle className="text-base font-bold tracking-tight">
                {String(t.title || 'Projected vs. Real Annual Investment Schedule')}
              </CardTitle>
            </div>
            <CardDescription className="text-xs mt-0.5">
              {String(t.description || '12-month comparison of planned tag budgets against real investments')}
            </CardDescription>
          </div>

          {/* Controls: Mode Switch & Tag Filter */}
          <div className="flex flex-wrap items-center gap-2">
            <div className="flex items-center p-1 bg-muted/50 rounded-xl border border-border/50">
              <button
                type="button"
                onClick={() => setViewMode('monthly')}
                className={cn(
                  'px-2.5 py-1 text-xs font-medium rounded-lg transition-all',
                  viewMode === 'monthly'
                    ? 'bg-primary text-primary-foreground shadow-xs font-semibold'
                    : 'text-muted-foreground hover:text-foreground',
                )}
              >
                {String(t.monthlyMode || 'Monthly Schedule')}
              </button>
              <button
                type="button"
                onClick={() => setViewMode('cumulative')}
                className={cn(
                  'px-2.5 py-1 text-xs font-medium rounded-lg transition-all',
                  viewMode === 'cumulative'
                    ? 'bg-primary text-primary-foreground shadow-xs font-semibold'
                    : 'text-muted-foreground hover:text-foreground',
                )}
              >
                {String(t.cumulativeMode || 'Cumulative Trajectory')}
              </button>
              <button
                type="button"
                onClick={() => setViewMode('byTag')}
                className={cn(
                  'px-2.5 py-1 text-xs font-medium rounded-lg transition-all',
                  viewMode === 'byTag'
                    ? 'bg-primary text-primary-foreground shadow-xs font-semibold'
                    : 'text-muted-foreground hover:text-foreground',
                )}
              >
                {String(t.byTagMode || 'By Category (Tag)')}
              </button>
            </div>

            {viewMode !== 'byTag' && (
              <Select value={selectedTag} onValueChange={setSelectedTag}>
                <SelectTrigger className="h-8 text-xs rounded-xl min-w-[130px]">
                  <SelectValue placeholder={String(t.tagFilterPlaceholder || 'Tag')} />
                </SelectTrigger>
                <SelectContent className="rounded-xl">
                  <SelectItem value="ALL" className="text-xs">
                    {String(t.allCategoriesFilter || 'All Categories')}
                  </SelectItem>
                  {availableTags.map((tag) => (
                    <SelectItem key={tag} value={tag} className="text-xs">
                      <span>{`#${tag}`}</span>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}

            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => setShowTable((prev) => !prev)}
              className="h-8 text-xs rounded-xl gap-1 text-muted-foreground"
            >
              <SlidersHorizontal className="h-3 w-3" />
              <span>{showTable ? String(t.hideTableButton || 'Hide Table') : String(t.showTableButton || 'Table')}</span>
            </Button>
          </div>
        </div>

        {/* Summary KPI Strip */}
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 pt-4">
          <div className="rounded-xl border border-border/60 p-2.5 bg-muted/20">
            <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">
              {String(t.plannedAnnualTitle || 'Planned Annual Budget')}
            </span>
            <p className="text-sm sm:text-base font-extrabold text-foreground mt-0.5">
              <span>{`${new Intl.NumberFormat().format(summary.planned)} ${currency}`}</span>
            </p>
          </div>

          <div className="rounded-xl border border-border/60 p-2.5 bg-muted/20">
            <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">
              {String(t.realInvestedTitle || 'Real Invested')}
            </span>
            <p className="text-sm sm:text-base font-extrabold text-foreground mt-0.5">
              <span>{`${new Intl.NumberFormat().format(summary.actual)} ${currency}`}</span>
            </p>
          </div>

          <div className={cn(
            'rounded-xl border p-2.5',
            summary.isOverBudget
              ? 'border-rose-500/30 bg-rose-500/10 text-rose-600 dark:text-rose-400'
              : 'border-emerald-500/30 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400',
          )}>
            <span className="text-[10px] font-bold uppercase tracking-wider">
              {summary.isOverBudget
                ? String(t.overBudgetTitle || 'Over Budget')
                : String(t.remainingBudgetTitle || 'Remaining Budget')}
            </span>
            <p className="text-sm sm:text-base font-extrabold mt-0.5">
              <span>{`${new Intl.NumberFormat().format(Math.abs(summary.variance))} ${currency}`}</span>
            </p>
          </div>

          <div className="rounded-xl border border-border/60 p-2.5 bg-muted/20">
            <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider">
              {String(t.adherenceTitle || 'Budget Utilization')}
            </span>
            <p className="text-sm sm:text-base font-extrabold text-foreground mt-0.5 flex items-center gap-1.5">
              <span>{`${summary.utilization}%`}</span>
              {summary.isOverBudget ? (
                <AlertTriangle className="h-4 w-4 text-rose-500" />
              ) : (
                <CheckCircle2 className="h-4 w-4 text-emerald-500" />
              )}
            </p>
          </div>
        </div>
      </CardHeader>

      <CardContent className="pt-4 space-y-4">
        {/* CHART VISUALIZATION */}
        <div className="h-72 w-full">
          <ResponsiveContainer width="100%" height="100%">
            {viewMode === 'monthly' && (
              <BarChart data={chartData} margin={{ top: 10, right: 10, left: -10, bottom: 0 }}>
                <CartesianGrid strokeDasharray="3 3" vertical={false} opacity={0.2} />
                <XAxis
                  dataKey="label"
                  tickLine={false}
                  axisLine={false}
                  tick={{ fontSize: 10, fill: 'currentColor', opacity: 0.7 }}
                />
                <YAxis
                  tickLine={false}
                  axisLine={false}
                  tick={{ fontSize: 10, fill: 'currentColor', opacity: 0.7 }}
                  tickFormatter={(v) => new Intl.NumberFormat().format(v)}
                />
                <Tooltip
                  formatter={(value, name) => [
                    `${new Intl.NumberFormat().format(value)} ${currency}`,
                    name === 'projected' ? String(t.projectedSeriesLabel || 'Projected') : String(t.actualSeriesLabel || 'Real Invested'),
                  ]}
                  contentStyle={{
                    borderRadius: '0.75rem',
                    fontSize: '11px',
                    background: 'hsl(var(--card))',
                    border: '1px solid hsl(var(--border))',
                  }}
                />
                <Legend
                  formatter={(val) => (val === 'projected' ? String(t.projectedSeriesLabel || 'Projected') : String(t.actualSeriesLabel || 'Real Invested'))}
                  wrapperStyle={{ fontSize: '11px', paddingTop: '8px' }}
                />
                <Bar
                  dataKey="projected"
                  fill="hsl(var(--primary))"
                  radius={[4, 4, 0, 0]}
                  maxBarSize={28}
                />
                <Bar
                  dataKey="actual"
                  radius={[4, 4, 0, 0]}
                  maxBarSize={28}
                >
                  {chartData.map((entry) => (
                    <Cell
                      key={`actual-bar-${entry.key}`}
                      fill={entry.isOverBudget ? '#f43f5e' : '#10b981'}
                    />
                  ))}
                </Bar>
              </BarChart>
            )}

            {viewMode === 'cumulative' && (
              <AreaChart data={chartData} margin={{ top: 10, right: 10, left: -10, bottom: 0 }}>
                <defs>
                  <linearGradient id="projGrad" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="hsl(var(--primary))" stopOpacity={0.4} />
                    <stop offset="95%" stopColor="hsl(var(--primary))" stopOpacity={0} />
                  </linearGradient>
                  <linearGradient id="actGrad" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#10b981" stopOpacity={0.4} />
                    <stop offset="95%" stopColor="#10b981" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" vertical={false} opacity={0.2} />
                <XAxis
                  dataKey="label"
                  tickLine={false}
                  axisLine={false}
                  tick={{ fontSize: 10, fill: 'currentColor', opacity: 0.7 }}
                />
                <YAxis
                  tickLine={false}
                  axisLine={false}
                  tick={{ fontSize: 10, fill: 'currentColor', opacity: 0.7 }}
                  tickFormatter={(v) => new Intl.NumberFormat().format(v)}
                />
                <Tooltip
                  formatter={(value, name) => [
                    `${new Intl.NumberFormat().format(value)} ${currency}`,
                    name === 'cumulativeProjected' ? String(t.projectedSeriesLabel || 'Projected') : String(t.actualSeriesLabel || 'Real Invested'),
                  ]}
                  contentStyle={{
                    borderRadius: '0.75rem',
                    fontSize: '11px',
                    background: 'hsl(var(--card))',
                    border: '1px solid hsl(var(--border))',
                  }}
                />
                <Legend
                  formatter={(val) => (val === 'cumulativeProjected' ? String(t.projectedSeriesLabel || 'Projected') : String(t.actualSeriesLabel || 'Real Invested'))}
                  wrapperStyle={{ fontSize: '11px', paddingTop: '8px' }}
                />
                <Area
                  type="monotone"
                  dataKey="cumulativeProjected"
                  stroke="hsl(var(--primary))"
                  strokeWidth={2}
                  fillOpacity={1}
                  fill="url(#projGrad)"
                />
                <Area
                  type="monotone"
                  dataKey="cumulativeActual"
                  stroke="#10b981"
                  strokeWidth={2}
                  fillOpacity={1}
                  fill="url(#actGrad)"
                />
              </AreaChart>
            )}

            {viewMode === 'byTag' && (
              <BarChart
                layout="vertical"
                data={tagBreakdownData}
                margin={{ top: 10, right: 20, left: 20, bottom: 0 }}
              >
                <CartesianGrid strokeDasharray="3 3" horizontal={false} opacity={0.2} />
                <XAxis
                  type="number"
                  tickLine={false}
                  axisLine={false}
                  tick={{ fontSize: 10, fill: 'currentColor', opacity: 0.7 }}
                  tickFormatter={(v) => new Intl.NumberFormat().format(v)}
                />
                <YAxis
                  type="category"
                  dataKey="name"
                  tickLine={false}
                  axisLine={false}
                  tick={{ fontSize: 10, fill: 'currentColor', opacity: 0.8 }}
                  width={90}
                />
                <Tooltip
                  formatter={(value, name) => [
                    `${new Intl.NumberFormat().format(value)} ${currency}`,
                    name === 'planned' ? String(t.projectedSeriesLabel || 'Projected') : String(t.actualSeriesLabel || 'Real Invested'),
                  ]}
                  contentStyle={{
                    borderRadius: '0.75rem',
                    fontSize: '11px',
                    background: 'hsl(var(--card))',
                    border: '1px solid hsl(var(--border))',
                  }}
                />
                <Legend
                  formatter={(val) => (val === 'planned' ? String(t.projectedSeriesLabel || 'Projected') : String(t.actualSeriesLabel || 'Real Invested'))}
                  wrapperStyle={{ fontSize: '11px', paddingTop: '8px' }}
                />
                <Bar
                  dataKey="planned"
                  fill="hsl(var(--primary))"
                  radius={[0, 4, 4, 0]}
                  maxBarSize={20}
                />
                <Bar
                  dataKey="actual"
                  radius={[0, 4, 4, 0]}
                  maxBarSize={20}
                >
                  {tagBreakdownData.map((entry) => (
                    <Cell
                      key={`actual-tag-${entry.tag}`}
                      fill={entry.isOverBudget ? '#f43f5e' : '#10b981'}
                    />
                  ))}
                </Bar>
              </BarChart>
            )}
          </ResponsiveContainer>
        </div>

        {/* Compact Monthly Breakdown Table (Toggleable) */}
        {showTable && (
          <div className="rounded-xl border border-border/60 overflow-hidden">
            <div className="overflow-x-auto">
              <table className="w-full text-xs text-left">
                <thead className="bg-muted/40 text-muted-foreground uppercase text-[10px] tracking-wider border-b border-border/60">
                  <tr>
                    <th className="py-2.5 px-3">{String(t.monthColumnTitle || 'Month')}</th>
                    <th className="py-2.5 px-3 text-right">{String(t.projectedSeriesLabel || 'Projected')}</th>
                    <th className="py-2.5 px-3 text-right">{String(t.actualSeriesLabel || 'Real Invested')}</th>
                    <th className="py-2.5 px-3 text-right">{String(t.varianceColumnTitle || 'Variance')}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border/40">
                  {chartData.map((d) => (
                    <tr key={d.key} className="hover:bg-muted/20 transition-colors">
                      <td className="py-2 px-3 font-semibold text-foreground">{d.label}</td>
                      <td className="py-2 px-3 text-right text-muted-foreground">
                        <span>{`${new Intl.NumberFormat().format(d.projected)} ${currency}`}</span>
                      </td>
                      <td className="py-2 px-3 text-right font-medium text-foreground">
                        <span>{`${new Intl.NumberFormat().format(d.actual)} ${currency}`}</span>
                      </td>
                      <td className={cn(
                        'py-2 px-3 text-right font-bold',
                        d.variance < 0 ? 'text-rose-500' : 'text-emerald-500',
                      )}>
                        <span>{`${d.variance > 0 ? '+' : ''}${new Intl.NumberFormat().format(d.variance)} ${currency}`}</span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

ProjectedVsRealSchedule.propTypes = {
  budgetSchedule: PropTypes.shape({
    totalAnnualBudget: PropTypes.number,
    totalActualInvested: PropTypes.number,
    annualVariance: PropTypes.number,
    budgetUtilization: PropTypes.number,
    isOverAnnualBudget: PropTypes.bool,
    months: PropTypes.arrayOf(PropTypes.shape({})),
    tagRules: PropTypes.arrayOf(PropTypes.shape({})),
    tagComparisons: PropTypes.arrayOf(PropTypes.shape({})),
  }),
  currency: PropTypes.string,
  content: PropTypes.shape({}),
};

export default ProjectedVsRealSchedule;
