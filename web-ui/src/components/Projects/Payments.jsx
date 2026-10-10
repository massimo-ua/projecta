import React, { useState, useEffect } from 'react';
import PropTypes from 'prop-types';
import { useParams, useNavigate, useSearchParams } from 'react-router-dom';
import { useIntlayer, useLocale } from 'react-intlayer';
import { getLocalizedUrl } from 'intlayer';
import {
  DollarSign,
  Calendar,
  FileUp,
  Package,
  X,
  Clock,
  Briefcase,
  User,
  Tag,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Checkbox } from '@/components/ui/checkbox';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { toast } from 'sonner';
import useInvestments from '../../hooks/investments';
import AddPaymentModal from './AddPaymentModal';
import EditPaymentModal from './EditPaymentModal';
import ImportStatementModal from './ImportStatementModal';
import CreateAssetFromPaymentsModal from './CreateAssetFromPaymentsModal';
import { ListView } from './ListView';
import { EditButton } from './ListView/EditButton';
import { RemoveButton } from './ListView/RemoveButton';
import { CopyableText } from './ListView/CopyableText';
import { DetailItem } from './ListView/DetailItem';
import './Payments.css';

function InvestmentMainContent({ investment }) {
  const content = useIntlayer('payments');
  const isTime = investment.resourceType === 'TIME';
  const isGoods = investment.resourceType === 'GOODS';

  return (
    <div className="flex flex-col gap-1.5">
      <div className="flex items-center gap-2 flex-wrap">
        <span className="flex items-center gap-1 text-xs text-muted-foreground font-mono">
          <Calendar className="h-3 w-3" />
          {investment.date || investment.paymentDate}
        </span>

        {/* Target Asset Badge */}
        {investment.isMultiAsset ? (
          <Badge variant="outline" className="text-xs font-semibold gap-1 border-primary/30 bg-primary/5 text-primary rounded-md">
            <Package className="h-3 w-3" />
            <span>{String(content.multiAssetBadge || 'Кілька активів')} ({investment.assetAllocations.length})</span>
          </Badge>
        ) : (
          <Badge variant="outline" className="text-xs font-semibold gap-1 border-primary/30 bg-primary/5 text-primary rounded-md">
            <Package className="h-3 w-3" />
            <span>{investment.assetName || investment.type || String(content.defaultCostType || 'General Operations')}</span>
          </Badge>
        )}

        {/* Resource Type Badge */}
        {isTime && (
          <Badge variant="secondary" className="text-[11px] font-medium gap-1 bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/20 rounded-md">
            <Clock className="h-3 w-3" />
            <span>{investment.resourceSummary || String(content.laborBadge || 'Праця')}</span>
          </Badge>
        )}

        {isGoods && (
          <Badge variant="secondary" className="text-[11px] font-medium gap-1 bg-violet-500/10 text-violet-600 dark:text-violet-400 border-violet-500/20 rounded-md">
            <Briefcase className="h-3 w-3" />
            <span>{investment.resourceSummary || String(content.goodsBadge || 'Матеріали')}</span>
          </Badge>
        )}

        {/* Contributor */}
        {investment.contributorName && (
          <span className="flex items-center gap-1 text-[11px] text-muted-foreground">
            <User className="h-3 w-3" />
            <span>{investment.contributorName}</span>
          </span>
        )}
      </div>

      {/* Description */}
      {investment.description && (
        <span className="font-semibold text-sm text-foreground">
          {investment.description}
        </span>
      )}

      {/* Tags */}
      {Array.isArray(investment.tags) && investment.tags.length > 0 && (
        <div className="flex gap-1.5 flex-wrap items-center mt-0.5">
          {investment.tags.map((tag) => (
            <Badge
              key={tag}
              variant="outline"
              className="text-[10px] font-mono px-1.5 py-0 rounded-md border-border/60 text-muted-foreground"
            >
              #{tag}
            </Badge>
          ))}
        </div>
      )}
    </div>
  );
}

InvestmentMainContent.propTypes = {
  investment: PropTypes.object.isRequired,
};

function InvestmentAmount({ investment }) {
  return (
    <div className="flex flex-col items-end gap-0.5">
      <span className="px-2.5 py-1 rounded-xl text-sm font-bold tracking-tight inline-block bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20">
        {investment.formattedAmount}
      </span>
      {investment.hasDifferentHomeCurrency && investment.formattedHomeAmount && (
        <span className="text-[11px] text-muted-foreground font-mono">
          {investment.formattedHomeAmount}
        </span>
      )}
    </div>
  );
}

InvestmentAmount.propTypes = {
  investment: PropTypes.object.isRequired,
};

function InvestmentDetails({ investment }) {
  const content = useIntlayer('payments');

  const resourceTypeName = investment.isTime
    ? String(content.timeResource || 'Праця (Час)')
    : investment.isGoods
      ? String(content.goodsResource || 'Товари та матеріали')
      : String(content.moneyResource || 'Капітал (Гроші)');

  return (
    <div className="space-y-2">
      <DetailItem label={String(content.detailId || 'ID')}>
        <CopyableText text={investment.id} truncate />
      </DetailItem>
      {investment.isMultiAsset ? (
        <DetailItem label={String(content.allocationsBreakdown || 'Розподіл за активами')}>
          <div className="space-y-1.5 w-full">
            {investment.assetAllocations.map((alloc) => (
              <div key={alloc.assetId} className="flex justify-between items-center text-sm py-1 border-b border-border/30 last:border-b-0">
                <span className="font-medium text-foreground">{alloc.assetName}</span>
                <span className="font-mono text-xs text-muted-foreground">
                  {alloc.sharePercentage}% {alloc.amount ? `(${alloc.amount} ${investment.currency})` : ''}
                </span>
              </div>
            ))}
          </div>
        </DetailItem>
      ) : (
        <DetailItem label={String(content.assetTargetLabel || 'Цільовий актив')}>
          <span className="text-sm font-semibold text-foreground">
            {investment.assetName || investment.type || String(content.defaultCostType || 'General Operations')}
          </span>
        </DetailItem>
      )}
      <DetailItem label={String(content.resourceTypeLabel || 'Тип ресурсу')}>
        <span className="text-sm text-foreground">{resourceTypeName}</span>
      </DetailItem>
      {investment.isTime && investment.timeHours && (
        <DetailItem label={String(content.timeBreakdown || 'Деталізація часу')}>
          <span className="text-sm text-foreground">
            {investment.timeHours} {String(content.timeHoursLabel || 'год')} {investment.timeHourlyRate ? `@ ${investment.timeHourlyRate} ${investment.currency}` : ''}
          </span>
        </DetailItem>
      )}
      {investment.isGoods && (
        <DetailItem label={String(content.goodsBreakdown || 'Деталізація товарів/матеріалів')}>
          <span className="text-sm text-foreground">
            {investment.goodsItemName} {investment.goodsQuantity ? `(${investment.goodsQuantity} ${investment.goodsUnit || ''})` : ''}
          </span>
        </DetailItem>
      )}
      {investment.tags && investment.tags.length > 0 && (
        <DetailItem label={String(content.tagsLabel || 'Теги')}>
          <span className="text-sm text-foreground font-mono">
            {investment.tags.map((t) => `#${t}`).join(', ')}
          </span>
        </DetailItem>
      )}
    </div>
  );
}

InvestmentDetails.propTypes = {
  investment: PropTypes.object.isRequired,
};

export function Payments() {
  const content = useIntlayer('payments');
  const { projectId } = useParams();
  const navigate = useNavigate();
  const { locale } = useLocale();

  const {
    loading,
    investments,
    total,
    currentPage,
    pageSize,
    availableTags,
    availableAssets,
    filters,
    setFilter,
    resetFilters,
    selectedInvestmentsMap,
    selectedInvestmentsList,
    selectedCount,
    toggleSelectInvestment,
    toggleSelectAllPage,
    isAllPageSelected,
    clearSelection,
    addModalOpened,
    investmentIdToEdit,
    onPaginationChange,
    openAddModal,
    closeAddModal,
    onAddSuccess,
    openEditModal,
    closeEditModal,
    onEditSuccess,
    removeInvestment,
  } = useInvestments(projectId);

  const [searchParams] = useSearchParams();
  const urlTag = searchParams.get('tag');

  useEffect(() => {
    if (urlTag) {
      setFilter('tag', urlTag);
    }
  }, [urlTag, setFilter]);

  const [importModalOpened, setImportModalOpened] = useState(false);
  const openImportModal = () => setImportModalOpened(true);
  const closeImportModal = () => setImportModalOpened(false);

  const [createAssetModalOpened, setCreateAssetModalOpened] = useState(false);
  const openCreateAssetModal = () => setCreateAssetModalOpened(true);
  const closeCreateAssetModal = () => setCreateAssetModalOpened(false);

  const handleRemove = (investmentId) => {
    removeInvestment(investmentId, {
      successMessage: String(content.paymentRemovedSuccess || 'Investment removed'),
      errorMessage: String(content.failedToRemove || 'Failed to remove investment'),
    });
  };

  const handleAssetCreated = () => {
    closeCreateAssetModal();
    clearSelection();
    toast.success(String(content.assetCreatedSuccess || 'Asset created successfully from selected investments'), {
      action: {
        label: String(content.viewInAssets || 'View in Assets'),
        onClick: () => navigate(getLocalizedUrl(`/projects/${projectId}/assets`, locale)),
      },
    });
  };

  const hasActiveFilters = Boolean(filters.assetId || filters.resourceType || filters.tag);

  return (
    <>
      <ListView
        loading={loading}
        items={investments}
        total={total}
        pageSize={pageSize}
        currentPage={currentPage}
        onPaginationChange={onPaginationChange}
        onAddButtonClick={openAddModal}
        addButtonIcon={<DollarSign className="h-4 w-4" />}
        addButtonText={String(content.addPayment || 'Log Investment')}
        addButtonDisabled={addModalOpened}
        extraActions={(
          <Button
            variant="outline"
            onClick={openImportModal}
            className="gap-2 font-semibold shadow-sm hover:shadow-md transition-all rounded-xl h-10 px-4 border-dashed hover:border-primary/50"
          >
            <FileUp className="h-4 w-4 text-primary" />
            <span>{String(content.importStatement || 'Import Statement')}</span>
          </Button>
        )}
        filterBar={(
          <div className="rounded-2xl border border-border/60 bg-card p-4 space-y-3">
            <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3 flex-wrap">
              <div className="flex items-center gap-3 flex-wrap flex-1">
                {/* Asset Filter */}
                <div className="w-full sm:w-48">
                  <Select
                    value={filters.assetId || 'ALL'}
                    onValueChange={(val) => setFilter('assetId', val === 'ALL' ? '' : val)}
                  >
                    <SelectTrigger className="rounded-xl h-9 text-xs">
                      <SelectValue placeholder={String(content.filterAllAssets || 'All Assets')} />
                    </SelectTrigger>
                    <SelectContent className="rounded-xl max-h-56">
                      <SelectItem value="ALL" className="rounded-lg text-xs font-medium">
                        {String(content.filterAllAssets || 'All Assets')}
                      </SelectItem>
                      {availableAssets.map((asset) => (
                        <SelectItem key={asset.id} value={asset.id} className="rounded-lg text-xs">
                          {asset.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>

                {/* Resource Type Filter */}
                <div className="w-full sm:w-40">
                  <Select
                    value={filters.resourceType || 'ALL'}
                    onValueChange={(val) => setFilter('resourceType', val === 'ALL' ? '' : val)}
                  >
                    <SelectTrigger className="rounded-xl h-9 text-xs">
                      <SelectValue placeholder={String(content.filterAllResourceTypes || 'All Resources')} />
                    </SelectTrigger>
                    <SelectContent className="rounded-xl">
                      <SelectItem value="ALL" className="rounded-lg text-xs font-medium">
                        {String(content.filterAllResourceTypes || 'All Resources')}
                      </SelectItem>
                      <SelectItem value="MONEY" className="rounded-lg text-xs">
                        {String(content.moneyResource || 'Capital (Money)')}
                      </SelectItem>
                      <SelectItem value="TIME" className="rounded-lg text-xs">
                        {String(content.timeResource || 'Labor (Time)')}
                      </SelectItem>
                      <SelectItem value="GOODS" className="rounded-lg text-xs">
                        {String(content.goodsResource || 'Goods & Materials')}
                      </SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                {/* Tag Filter */}
                <div className="w-full sm:w-44">
                  <div className="relative">
                    <Tag className="absolute left-2.5 top-2.5 h-3.5 w-3.5 text-muted-foreground" />
                    <Input
                      placeholder={String(content.filterTagPlaceholder || 'Filter by tag...')}
                      value={filters.tag || ''}
                      onChange={(e) => setFilter('tag', e.target.value.trim().replace(/^#/, ''))}
                      className="rounded-xl h-9 text-xs pl-8"
                    />
                  </div>
                </div>

                {/* Available Quick Tag Pills */}
                {availableTags.length > 0 && !filters.tag && (
                  <div className="hidden lg:flex items-center gap-1 text-[11px] text-muted-foreground overflow-x-auto">
                    <span>Popular:</span>
                    {availableTags.slice(0, 4).map((t) => (
                      <button
                        key={t}
                        type="button"
                        onClick={() => setFilter('tag', t)}
                        className="px-2 py-0.5 rounded-md bg-muted/60 hover:bg-muted text-foreground transition-colors font-mono"
                      >
                        #{t}
                      </button>
                    ))}
                  </div>
                )}

                {/* Reset Filters */}
                {hasActiveFilters && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={resetFilters}
                    className="gap-1 text-xs text-muted-foreground hover:text-foreground h-9 px-2.5 rounded-xl"
                  >
                    <X className="h-3.5 w-3.5" />
                    <span>{String(content.resetFilters || 'Reset')}</span>
                  </Button>
                )}
              </div>

              {/* Select All on Page Toggle */}
              {investments.length > 0 && (
                <div className="flex items-center gap-2 pt-2 sm:pt-0 border-t sm:border-0 border-border/40">
                  <Checkbox
                    id="select-all-page"
                    checked={isAllPageSelected}
                    onCheckedChange={toggleSelectAllPage}
                  />
                  <Label htmlFor="select-all-page" className="text-xs text-muted-foreground cursor-pointer select-none">
                    {String(content.selectAll || 'Select all on page')}
                  </Label>
                </div>
              )}
            </div>
          </div>
        )}
        batchActionBar={selectedCount > 0 ? (
          <div className="flex items-center justify-between gap-3 p-3.5 rounded-2xl bg-primary/5 border border-primary/20 shadow-xs flex-wrap">
            <div className="flex items-center gap-3">
              <div className="flex items-center gap-2">
                <Badge variant="default" className="rounded-lg font-bold">
                  {selectedCount}
                </Badge>
                <span className="text-xs font-semibold text-foreground">
                  {String(content.selectedPayments || 'investments selected')}
                </span>
              </div>
              <Button
                variant="ghost"
                size="sm"
                onClick={clearSelection}
                className="text-xs text-muted-foreground hover:text-foreground h-7 px-2"
              >
                {String(content.clearSelection || 'Clear')}
              </Button>
            </div>

            <Button
              size="sm"
              onClick={openCreateAssetModal}
              className="gap-2 rounded-xl font-semibold shadow-sm shadow-primary/20"
            >
              <Package className="h-4 w-4" />
              <span>{String(content.createAssetFromSelected || 'Create Asset from Selected')}</span>
            </Button>
          </div>
        ) : null}
        renderItemPrefix={(inv) => (
          <Checkbox
            checked={Boolean(selectedInvestmentsMap[inv.id])}
            onCheckedChange={() => toggleSelectInvestment(inv)}
            aria-label={`Select investment ${inv.description}`}
          />
        )}
        renderItemMainContent={(inv) => <InvestmentMainContent investment={inv} />}
        renderItemAmount={(inv) => <InvestmentAmount investment={inv} />}
        renderItemDetails={(inv) => <InvestmentDetails investment={inv} />}
        renderItemActions={(inv) => (
          <>
            <EditButton onClick={() => openEditModal(inv.id)} />
            <RemoveButton onRemove={() => handleRemove(inv.id)} />
          </>
        )}
      />

      <AddPaymentModal
        assets={availableAssets}
        projectId={projectId}
        open={addModalOpened}
        onCancel={closeAddModal}
        onSuccess={() => onAddSuccess(String(content.paymentAddedSuccess))}
      />

      <ImportStatementModal
        assets={availableAssets}
        projectId={projectId}
        open={importModalOpened}
        onCancel={closeImportModal}
        onSuccess={() => {
          closeImportModal();
          onAddSuccess(String(content.paymentsImportedSuccess || 'Investments imported successfully'));
        }}
      />

      <EditPaymentModal
        assets={availableAssets}
        projectId={projectId}
        paymentId={investmentIdToEdit}
        open={Boolean(investmentIdToEdit)}
        onCancel={closeEditModal}
        onSuccess={() => onEditSuccess(String(content.paymentUpdatedSuccess))}
      />

      <CreateAssetFromPaymentsModal
        open={createAssetModalOpened}
        onCancel={closeCreateAssetModal}
        onSuccess={handleAssetCreated}
        projectId={projectId}
        payments={selectedInvestmentsList}
      />
    </>
  );
}

export const Investments = Payments;
export default Payments;
