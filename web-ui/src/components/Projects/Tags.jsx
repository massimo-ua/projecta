import React, { useState } from 'react';
import PropTypes from 'prop-types';
import { useParams, useNavigate } from 'react-router-dom';
import { useIntlayer, useLocale } from 'react-intlayer';
import { getLocalizedUrl } from 'intlayer';
import {
  Tag as TagIcon,
  Search,
  SlidersHorizontal,
  ArrowRight,
  DollarSign,
  Clock,
  Briefcase,
  Layers,
  TrendingUp,
  Inbox,
  RefreshCw,
  Plus,
  Edit2,
  Calendar,
} from 'lucide-react';
import { toast } from 'sonner';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Skeleton } from '@/components/ui/skeleton';
import { cn } from '@/lib/utils';
import { useProjectTags } from '../../hooks/useProjectTags';
import { EditTagBudgetModal } from './EditTagBudgetModal';
import { cleanTagName } from '../../models';

function TagKpiCard({ title, value, subtitle, icon: Icon, colorClass }) {
  return (
    <div className="rounded-2xl border border-border/70 bg-card p-5 shadow-xs transition-all duration-300 hover:shadow-md hover:border-primary/40">
      <div className="flex items-center justify-between gap-3 mb-3">
        <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
          {title}
        </span>
        <div className={`flex h-8 w-8 items-center justify-center rounded-xl ring-1 ${colorClass}`}>
          <Icon className="h-4 w-4" />
        </div>
      </div>
      <div className="flex flex-col gap-0.5">
        <span className="text-2xl font-extrabold tracking-tight text-foreground">
          {value}
        </span>
        {subtitle && (
          <span className="text-xs text-muted-foreground">
            {subtitle}
          </span>
        )}
      </div>
    </div>
  );
}

TagKpiCard.propTypes = {
  title: PropTypes.string.isRequired,
  value: PropTypes.oneOfType([PropTypes.string, PropTypes.number]).isRequired,
  subtitle: PropTypes.string,
  icon: PropTypes.elementType.isRequired,
  colorClass: PropTypes.string.isRequired,
};

function TagCard({ tag, currency, onSelectTag, onConfigureBudget, content }) {
  const { locale } = useLocale();
  const intlLocale = (locale && locale.startsWith('uk')) ? 'uk-UA' : 'en-US';
  const hasMoney = tag.resourceTypes.includes('MONEY');
  const hasTime = tag.resourceTypes.includes('TIME');
  const hasGoods = tag.resourceTypes.includes('GOODS');

  return (
    <div className="group flex flex-col justify-between rounded-2xl border border-border/70 bg-card p-5 shadow-xs transition-all duration-200 hover:shadow-md hover:border-primary/50">
      <div className="space-y-4">
        {/* Header: Tag Pill, Count & Budget Badge */}
        <div className="flex items-start justify-between gap-2">
          <div className="flex items-center gap-1.5 flex-wrap">
            <span className="font-mono text-base font-bold text-primary group-hover:underline">
              <span>{`#${tag.name}`}</span>
            </span>
            {tag.tagBudget && (
              <Badge variant="outline" className="text-[10px] px-1.5 py-0 rounded-md bg-primary/5 border-primary/20 text-primary">
                {String(content.frequencies?.[tag.tagBudget.frequency] || tag.tagBudget.frequency)}
              </Badge>
            )}
          </div>
          <Badge
            variant="secondary"
            className="text-xs font-semibold px-2 py-0.5 rounded-lg border-border/60 bg-muted/60"
          >
            <span>{`${tag.count} ${tag.count === 1 ? String(content.singleInvestment || 'investment') : String(content.investmentsCount || 'investments')}`}</span>
          </Badge>
        </div>

        {/* Valuation & Budget comparison */}
        <div className="space-y-2">
          <div className="flex items-baseline gap-1.5">
            <span className="text-xl font-extrabold text-foreground tracking-tight">
              {tag.formattedAmount}
            </span>
            <span className="text-xs font-mono font-medium text-muted-foreground">
              {currency}
            </span>
          </div>

          {/* Budget progress bar if tag has budget rule */}
          {tag.tagBudget && (
            <div className="space-y-1 pt-1 border-t border-border/40">
              <div className="flex items-center justify-between text-[11px]">
                <span className="text-muted-foreground">
                  {String(content.plannedAnnualTitle || 'Planned Annual')}:
                </span>
                <span className="font-semibold text-foreground">
                  <span>{`${new Intl.NumberFormat(intlLocale).format(tag.projectedAnnualAmount)} ${currency}`}</span>
                </span>
              </div>
              <div className="w-full bg-muted/60 h-2 rounded-full overflow-hidden">
                <div
                  className={cn(
                    'h-full transition-all rounded-full',
                    tag.isOverBudget ? 'bg-orange-500' : 'bg-emerald-500',
                  )}
                  style={{ width: `${Math.min(tag.budgetUtilization || 0, 100)}%` }}
                />
              </div>
              <div className="flex items-center justify-between text-[10px]">
                <span className={cn(
                  'font-bold',
                  tag.isOverBudget ? 'text-orange-500' : 'text-muted-foreground',
                )}>
                  <span>{`${tag.budgetUtilization || 0}%`}</span>
                  {tag.isOverBudget && (
                    <span className="ml-1">
                      <span>{`(${String(content.overbudgetBadge || 'Overbudget')})`}</span>
                    </span>
                  )}
                </span>
              </div>
            </div>
          )}

          {/* Resource types involved */}
          <div className="flex items-center gap-1.5 flex-wrap pt-1">
            {hasMoney && (
              <Badge variant="outline" className="text-[10px] font-medium gap-1 px-1.5 py-0.5 rounded-md text-emerald-600 dark:text-emerald-400 border-emerald-500/20 bg-emerald-500/5">
                <DollarSign className="h-2.5 w-2.5" />
                <span>{String(content.capitalBadge || 'Capital')}</span>
              </Badge>
            )}
            {hasTime && (
              <Badge variant="outline" className="text-[10px] font-medium gap-1 px-1.5 py-0.5 rounded-md text-amber-600 dark:text-amber-400 border-amber-500/20 bg-amber-500/5">
                <Clock className="h-2.5 w-2.5" />
                <span>{String(content.laborBadge || 'Labor')}</span>
              </Badge>
            )}
            {hasGoods && (
              <Badge variant="outline" className="text-[10px] font-medium gap-1 px-1.5 py-0.5 rounded-md text-violet-600 dark:text-violet-400 border-violet-500/20 bg-violet-500/5">
                <Briefcase className="h-2.5 w-2.5" />
                <span>{String(content.goodsBadge || 'Goods')}</span>
              </Badge>
            )}
          </div>
        </div>
      </div>

      {/* Footer Actions */}
      <div className="pt-3 mt-4 border-t border-border/50 flex items-center justify-between gap-2">
        <Button
          variant="outline"
          size="sm"
          onClick={() => onConfigureBudget(tag.name, tag.tagBudget)}
          className="text-xs font-medium h-8 rounded-xl px-2.5 gap-1 text-muted-foreground hover:text-foreground"
        >
          <Edit2 className="h-3 w-3" />
          <span>
            {tag.tagBudget
              ? String(content.editBudgetButton || 'Edit Budget')
              : String(content.configureBudgetButton || 'Configure Budget')}
          </span>
        </Button>

        <Button
          variant="ghost"
          size="sm"
          onClick={() => onSelectTag(tag.name)}
          className="text-xs font-semibold h-8 rounded-xl text-primary hover:bg-primary/10 transition-colors gap-1 px-2.5"
        >
          <span>{String(content.viewInvestments || 'View')}</span>
          <ArrowRight className="h-3.5 w-3.5 transition-transform group-hover:translate-x-1" />
        </Button>
      </div>
    </div>
  );
}

TagCard.propTypes = {
  tag: PropTypes.shape({
    name: PropTypes.string.isRequired,
    count: PropTypes.number.isRequired,
    formattedAmount: PropTypes.string.isRequired,
    resourceTypes: PropTypes.arrayOf(PropTypes.string).isRequired,
    tagBudget: PropTypes.shape({}),
    projectedAnnualAmount: PropTypes.number,
    budgetUtilization: PropTypes.number,
    isOverBudget: PropTypes.bool,
  }).isRequired,
  currency: PropTypes.string.isRequired,
  onSelectTag: PropTypes.func.isRequired,
  onConfigureBudget: PropTypes.func.isRequired,
  content: PropTypes.shape({}).isRequired,
};

function TagEmptyState({ hasSearch, onGoToInvestments, content }) {
  if (hasSearch) {
    return (
      <div className="flex flex-col items-center justify-center p-12 text-center rounded-2xl border border-dashed border-border/80 bg-card/50">
        <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-muted/60 text-muted-foreground mb-4">
          <Search className="h-6 w-6" />
        </div>
        <h3 className="text-base font-semibold text-foreground mb-1">
          {String(content.noSearchMatchTitle || 'No tags match your search')}
        </h3>
        <p className="text-xs text-muted-foreground max-w-sm mb-4">
          {String(content.noSearchMatchDescription || 'Try searching for another keyword.')}
        </p>
      </div>
    );
  }

  return (
    <div className="flex flex-col items-center justify-center p-12 text-center rounded-2xl border border-dashed border-border/80 bg-card/50">
      <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-primary/10 text-primary mb-4">
        <Inbox className="h-6 w-6" />
      </div>
      <h3 className="text-base font-semibold text-foreground mb-1">
        {String(content.noTagsTitle || 'No tags in this project yet')}
      </h3>
      <p className="text-xs text-muted-foreground max-w-md mb-5">
        {String(content.noTagsDescription || 'Tags are added when creating or importing investments.')}
      </p>
      <Button
        variant="default"
        size="sm"
        onClick={onGoToInvestments}
        className="rounded-xl font-semibold gap-2"
      >
        <DollarSign className="h-4 w-4" />
        <span>{String(content.goToInvestments || 'Go to Investments')}</span>
      </Button>
    </div>
  );
}

TagEmptyState.propTypes = {
  hasSearch: PropTypes.bool.isRequired,
  onGoToInvestments: PropTypes.func.isRequired,
  content: PropTypes.shape({}).isRequired,
};

export function Tags() {
  const content = useIntlayer('tags');
  const { projectId } = useParams();
  const navigate = useNavigate();
  const { locale } = useLocale();

  const {
    loading,
    currency,
    project,
    searchQuery,
    setSearchQuery,
    sortBy,
    setSortBy,
    tags,
    tagBudgets,
    totalTagsCount,
    totalInvestmentsTagged,
    formattedTotalAmountTagged,
    topTag,
    saveTagBudgets,
    reload,
  } = useProjectTags(projectId);

  const [modalState, setModalState] = useState({
    open: false,
    tag: '',
    existingRule: null,
  });

  const handleSelectTag = (tagName) => {
    const targetUrl = `${getLocalizedUrl(`/projects/${projectId}/investments`, locale)}?tag=${encodeURIComponent(tagName)}`;
    navigate(targetUrl);
  };

  const handleGoToInvestments = () => {
    navigate(getLocalizedUrl(`/projects/${projectId}/investments`, locale));
  };

  const handleOpenBudgetModal = (tagName = '', existingRule = null) => {
    setModalState({
      open: true,
      tag: tagName,
      existingRule,
    });
  };

  const handleCloseBudgetModal = () => {
    setModalState({
      open: false,
      tag: '',
      existingRule: null,
    });
  };

  const handleSaveBudgetRule = async (newRule) => {
    const cleaned = cleanTagName(modalState.tag || (newRule ? newRule.tag : ''));
    let updatedRules = [...tagBudgets];

    if (!newRule) {
      updatedRules = updatedRules.filter((r) => cleanTagName(r.tag) !== cleaned);
    } else {
      const idx = updatedRules.findIndex((r) => cleanTagName(r.tag) === cleanTagName(newRule.tag));
      if (idx >= 0) {
        updatedRules[idx] = newRule;
      } else {
        updatedRules.push(newRule);
      }
    }

    try {
      await saveTagBudgets(updatedRules);
      toast.success(String(content.budgetSavedSuccess || 'Category budget updated'));
    } catch (err) {
      toast.error(`${String(content.budgetSaveError || 'Failed to update category budget')}: ${err.message || ''}`);
    }
  };

  return (
    <div className="space-y-6">
      {/* View Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-4 border-b border-border/60">
        <div>
          <div className="flex items-center gap-2.5">
            <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-primary/10 text-primary ring-1 ring-primary/20">
              <TagIcon className="h-4 w-4" />
            </div>
            <h1 className="text-2xl font-bold tracking-tight text-foreground">
              {String(content.title || 'Project Tags')}
            </h1>
          </div>
          <p className="text-xs text-muted-foreground mt-1">
            {String(content.subtitle || 'Track and analyze investments categorized by tags')}
          </p>
        </div>

        <div className="flex items-center gap-2 self-start sm:self-auto">
          <Button
            variant="default"
            size="sm"
            onClick={() => handleOpenBudgetModal('')}
            className="rounded-xl gap-2 text-xs font-semibold h-9 shadow-xs"
          >
            <Plus className="h-3.5 w-3.5" />
            <span>{String(content.configureBudgetButton || 'Configure Tag Budget')}</span>
          </Button>

          <Button
            variant="outline"
            size="sm"
            onClick={reload}
            className="rounded-xl gap-2 text-xs font-semibold h-9"
          >
            <RefreshCw className="h-3.5 w-3.5" />
            <span>{String(content.refreshButton || 'Refresh')}</span>
          </Button>
        </div>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <TagKpiCard
          title={String(content.totalTagsCard || 'Total Tags')}
          value={loading ? '...' : totalTagsCount}
          subtitle={String(content.totalTagsSubtitle || 'Distinct taxonomy labels')}
          icon={TagIcon}
          colorClass="bg-primary/10 text-primary ring-primary/20"
        />
        <TagKpiCard
          title={String(content.taggedInvestmentsCard || 'Tagged Allocations')}
          value={loading ? '...' : totalInvestmentsTagged}
          subtitle={String(content.taggedInvestmentsSubtitle || 'Resource allocations')}
          icon={Layers}
          colorClass="bg-blue-500/10 text-blue-600 dark:text-blue-400 ring-blue-500/20"
        />
        <TagKpiCard
          title={String(content.totalValuationCard || 'Total Tagged Value')}
          value={loading ? '...' : `${formattedTotalAmountTagged} ${currency}`}
          subtitle={String(content.totalValuationSubtitle || 'Across all tags')}
          icon={DollarSign}
          colorClass="bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 ring-emerald-500/20"
        />
        <TagKpiCard
          title={String(content.topTagCard || 'Most Used Tag')}
          value={loading ? '...' : topTag ? `#${topTag.name}` : '—'}
          subtitle={topTag
            ? `${topTag.count} ${String(content.investmentsCount || 'investments')}`
            : String(content.noTagsLabel || 'No tags')}
          icon={TrendingUp}
          colorClass="bg-amber-500/10 text-amber-600 dark:text-amber-400 ring-amber-500/20"
        />
      </div>

      {/* Filter and Sort Toolbar */}
      <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3 p-4 rounded-2xl border border-border/60 bg-card">
        {/* Search */}
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder={String(content.searchPlaceholder || 'Search tags by name...')}
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-9 rounded-xl h-9 text-xs"
          />
        </div>

        {/* Sort dropdown */}
        <div className="flex items-center gap-2 self-end sm:self-auto">
          <SlidersHorizontal className="h-3.5 w-3.5 text-muted-foreground" />
          <span className="text-xs text-muted-foreground hidden sm:inline">
            {String(content.sortLabel || 'Sort by:')}
          </span>
          <Select value={sortBy} onValueChange={setSortBy}>
            <SelectTrigger className="rounded-xl h-9 text-xs w-36">
              <SelectValue />
            </SelectTrigger>
            <SelectContent className="rounded-xl">
              <SelectItem value="count" className="rounded-lg text-xs">
                {String(content.sortByCount || 'Most Used')}
              </SelectItem>
              <SelectItem value="amount" className="rounded-lg text-xs">
                {String(content.sortByAmount || 'Highest Amount')}
              </SelectItem>
              <SelectItem value="name" className="rounded-lg text-xs">
                {String(content.sortByName || 'Name (A-Z)')}
              </SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      {/* Tags Grid / List */}
      {loading ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
          {[1, 2, 3, 4, 5, 6].map((i) => (
            <div key={i} className="rounded-2xl border border-border/60 p-5 space-y-4 bg-card">
              <div className="flex justify-between items-center">
                <Skeleton className="h-5 w-24 rounded-lg" />
                <Skeleton className="h-5 w-16 rounded-md" />
              </div>
              <Skeleton className="h-7 w-28 rounded-lg" />
              <Skeleton className="h-8 w-full rounded-xl" />
            </div>
          ))}
        </div>
      ) : tags.length > 0 ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
          {tags.map((tag) => (
            <TagCard
              key={tag.name}
              tag={tag}
              currency={currency}
              onSelectTag={handleSelectTag}
              onConfigureBudget={(name, rule) => handleOpenBudgetModal(name, rule)}
              content={content}
            />
          ))}
        </div>
      ) : (
        <TagEmptyState
          hasSearch={Boolean(searchQuery.trim())}
          onGoToInvestments={handleGoToInvestments}
          content={content}
        />
      )}

      {/* Edit / Add Tag Budget Modal */}
      <EditTagBudgetModal
        open={modalState.open}
        tag={modalState.tag}
        existingRule={modalState.existingRule}
        currency={currency}
        startDate={project?.startDate}
        endDate={project?.endDate}
        onSave={handleSaveBudgetRule}
        onClose={handleCloseBudgetModal}
      />
    </div>
  );
}

export default Tags;
