import React, { useState } from 'react';
import PropTypes from 'prop-types';
import { useParams, useNavigate } from 'react-router-dom';
import { useIntlayer, useLocale } from 'react-intlayer';
import { getLocalizedUrl } from 'intlayer';
import {
  Package,
  Calendar,
  Link2,
  Unlink,
  CheckCircle2,
  Clock,
  ArrowRight,
  Boxes,
  DollarSign,
  FolderPlus,
} from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { Label } from '@/components/ui/label';
import useAssets from '../../hooks/assets';
import AddAssetModal from './AddAssetModal';
import EditAssetModal from './EditAssetModal';
import LinkChildAssetModal from './LinkChildAssetModal';
import AddPaymentModal from './AddPaymentModal';
import GroupAssetsModal from './GroupAssetsModal';
import AssignToParentModal from './AssignToParentModal';
import { ListView } from './ListView';
import { EditButton } from './ListView/EditButton';
import { RemoveButton } from './ListView/RemoveButton';
import { CopyableText } from './ListView/CopyableText';
import { DetailItem } from './ListView/DetailItem';
import { cn } from '@/lib/utils';
import './Assets.css';

function AssetMainContent({ asset }) {
  const isActive = asset.status === 'ACTIVE';

  return (
    <div className="flex flex-col gap-1.5">
      <div className="flex items-center gap-2 flex-wrap">
        <Badge
          variant="outline"
          className={cn(
            'text-[11px] font-semibold gap-1 rounded-md px-2 py-0.5',
            isActive
              ? 'border-emerald-500/30 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400'
              : 'border-blue-500/30 bg-blue-500/10 text-blue-600 dark:text-blue-400',
          )}
        >
          {isActive ? <Clock className="h-3 w-3" /> : <CheckCircle2 className="h-3 w-3" />}
          <span>{asset.status || 'ACTIVE'}</span>
        </Badge>

        <span className="font-bold text-base text-foreground tracking-tight">
          {asset.name}
        </span>
      </div>

      {/* Dates */}
      <div className="flex items-center gap-1.5 text-xs text-muted-foreground font-mono">
        <Calendar className="h-3 w-3 text-muted-foreground/70" />
        <span>{asset.startDate || asset.acquiredAt}</span>
        {asset.completedDate && (
          <>
            <ArrowRight className="h-3 w-3 text-muted-foreground/50" />
            <span>{asset.completedDate}</span>
          </>
        )}
      </div>

      {/* Description */}
      {asset.description && (
        <span className="text-xs text-muted-foreground mt-0.5 line-clamp-2">
          {asset.description}
        </span>
      )}

      {/* Tags */}
      {asset.tags && asset.tags.length > 0 && (
        <div className="flex gap-1 flex-wrap mt-0.5">
          {asset.tags.map((t) => (
            <Badge key={t} variant="outline" className="text-[10px] py-0 px-1.5 font-normal rounded-md text-muted-foreground">
              #{t}
            </Badge>
          ))}
        </div>
      )}

      {/* Hierarchy summaries */}
      <div className="flex gap-2 flex-wrap items-center mt-1">
        {asset.hasChildren && (
          <Badge variant="secondary" className="text-[11px] font-normal gap-1 rounded-md">
            <Boxes className="h-3 w-3 text-primary" />
            <span>{asset.children.length} component{asset.children.length > 1 ? 's' : ''}</span>
          </Badge>
        )}

        {asset.hasParents && (
          <span className="text-[11px] text-muted-foreground">
            Part of: {asset.parents.map((p) => `${p.parentName} (${p.sharePercentage}%)`).join(', ')}
          </span>
        )}
      </div>
    </div>
  );
}

AssetMainContent.propTypes = {
  asset: PropTypes.object.isRequired,
};

function AssetAmount({ asset }) {
  return (
    <div className="flex flex-col items-end gap-1">
      {/* Total Rolled-up Cost */}
      <span className="px-3 py-1 rounded-xl text-sm font-extrabold tracking-tight inline-block bg-primary/10 text-primary border border-primary/20">
        {asset.formattedTotalCost || asset.formattedPrice}
      </span>

      {/* Direct Cost if different */}
      {asset.hasChildren && asset.directCost && asset.directCost !== asset.totalCost && (
        <span className="text-[11px] text-muted-foreground font-mono">
          Direct: {asset.formattedDirectCost}
        </span>
      )}

      {/* Home Currency conversion */}
      {asset.hasDifferentHomeCurrency && asset.formattedHomeAmount && (
        <span className="text-[11px] text-muted-foreground font-mono">
          {asset.formattedHomeAmount}
        </span>
      )}

      {/* Target Budget Progress Bar */}
      {asset.hasTarget && (
        <div className="w-28 sm:w-36 space-y-1 mt-1 text-right">
          <div className="flex justify-between text-[10px] text-muted-foreground font-mono">
            <span>Budget</span>
            <span>{asset.targetProgress}%</span>
          </div>
          <div className="h-1.5 w-full bg-muted rounded-full overflow-hidden">
            <div
              className={cn(
                'h-full rounded-full transition-all duration-300',
                asset.targetProgress > 100 ? 'bg-rose-500' : 'bg-primary',
              )}
              style={{ width: `${asset.targetProgress}%` }}
            />
          </div>
        </div>
      )}
    </div>
  );
}

AssetAmount.propTypes = {
  asset: PropTypes.object.isRequired,
};

function AssetDetails({ asset, onUnlinkChild, onViewInvestments }) {
  return (
    <div className="space-y-3">
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
        <DetailItem label="Asset ID">
          <CopyableText text={asset.id} truncate />
        </DetailItem>
        <DetailItem label="Status">
          <span className="text-sm font-medium">{asset.status}</span>
        </DetailItem>
        <DetailItem label="Start Date">
          <span className="text-sm">{asset.startDate || asset.acquiredAt}</span>
        </DetailItem>
        {asset.completedDate && (
          <DetailItem label="Completed Date">
            <span className="text-sm">{asset.completedDate}</span>
          </DetailItem>
        )}
        <DetailItem label="Direct Cost">
          <div className="flex items-center gap-2">
            <span className="text-sm font-mono font-medium">{asset.formattedDirectCost}</span>
            {onViewInvestments && (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => onViewInvestments(asset.id)}
                className="h-6 px-1.5 text-xs text-muted-foreground hover:text-primary gap-1"
                title="View investments for this asset"
              >
                <span>Investments</span>
                <ArrowRight className="h-3 w-3" />
              </Button>
            )}
          </div>
        </DetailItem>
        <DetailItem label="Total Cost (Rollup)">
          <span className="text-sm font-mono font-bold text-primary">{asset.formattedTotalCost}</span>
        </DetailItem>
        {asset.hasTarget && (
          <DetailItem label="Target Price / Budget">
            <span className="text-sm font-mono">{asset.formattedTargetPrice}</span>
          </DetailItem>
        )}
        {asset.tags && asset.tags.length > 0 && (
          <DetailItem label="Tags">
            <div className="flex gap-1 flex-wrap">
              {asset.tags.map((t) => (
                <Badge key={t} variant="secondary" className="text-xs py-0.5">
                  #{t}
                </Badge>
              ))}
            </div>
          </DetailItem>
        )}
      </div>

      {/* Linked Components / Sub-Assets */}
      {asset.hasChildren && (
        <div className="pt-2 border-t border-border/50 space-y-2">
          <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider block">
            Components & Sub-Assets ({asset.children.length})
          </span>
          <div className="space-y-1.5">
            {asset.children.map((child) => (
              <div
                key={child.childId}
                className="flex items-center justify-between gap-3 p-2 rounded-xl bg-background/80 border border-border/60 text-xs"
              >
                <div className="flex items-center gap-2">
                  <Package className="h-3.5 w-3.5 text-muted-foreground" />
                  <span className="font-semibold">{child.childName}</span>
                  <Badge variant="outline" className="text-[10px] px-1.5 py-0 h-4 font-mono">
                    {child.sharePercentage}%
                  </Badge>
                </div>
                <div className="flex items-center gap-2">
                  <span className="font-mono text-muted-foreground">
                    {child.totalCost} {asset.currency}
                  </span>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => onUnlinkChild(asset.id, child.childId)}
                    className="h-6 w-6 p-0 text-muted-foreground hover:text-destructive rounded-lg"
                    title="Unlink component"
                  >
                    <Unlink className="h-3 w-3" />
                  </Button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Parent Assets */}
      {asset.hasParents && (
        <div className="pt-2 border-t border-border/50 space-y-1.5">
          <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider block">
            Component Of ({asset.parents.length})
          </span>
          <div className="flex gap-2 flex-wrap">
            {asset.parents.map((p) => (
              <Badge key={p.parentId} variant="secondary" className="text-xs py-1 px-2.5">
                {p.parentName} ({p.sharePercentage}%)
              </Badge>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

AssetDetails.propTypes = {
  asset: PropTypes.object.isRequired,
  onUnlinkChild: PropTypes.func.isRequired,
  onViewInvestments: PropTypes.func,
};

AssetDetails.defaultProps = {
  onViewInvestments: undefined,
};


export function Assets() {
  const content = useIntlayer('assets');
  const { projectId } = useParams();
  const navigate = useNavigate();
  const { locale } = useLocale();
  const [investAssetId, setInvestAssetId] = useState('');
  const [selectedAssetsMap, setSelectedAssetsMap] = useState({});
  const [groupModalOpened, setGroupModalOpened] = useState(false);
  const [assignModalOpened, setAssignModalOpened] = useState(false);

  const {
    loading,
    assets,
    total,
    currentPage,
    pageSize,
    addModalOpened,
    assetIdToEdit,
    linkModalAsset,
    onPaginationChange,
    openAddModal,
    closeAddModal,
    onAddSuccess,
    openEditModal,
    closeEditModal,
    onEditSuccess,
    openLinkModal,
    closeLinkModal,
    linkChild,
    unlinkChild,
    removeAsset,
    refresh,
  } = useAssets(projectId);

  const selectedAssets = Object.values(selectedAssetsMap);
  const selectedCount = selectedAssets.length;

  const toggleSelectAsset = (asset) => {
    setSelectedAssetsMap((prev) => {
      const next = { ...prev };
      if (next[asset.id]) {
        delete next[asset.id];
      } else {
        next[asset.id] = asset;
      }
      return next;
    });
  };

  const clearSelection = () => {
    setSelectedAssetsMap({});
  };

  const isAllPageSelected = assets.length > 0 && assets.every((a) => selectedAssetsMap[a.id]);

  const toggleSelectAllPage = () => {
    if (isAllPageSelected) {
      clearSelection();
    } else {
      const next = { ...selectedAssetsMap };
      assets.forEach((a) => {
        next[a.id] = a;
      });
      setSelectedAssetsMap(next);
    }
  };

  const handleViewInvestments = (assetId) => {
    navigate(getLocalizedUrl(`/projects/${projectId}/payments?assetId=${assetId}`, locale));
  };

  const handleRemove = (assetId) => {
    removeAsset(assetId, {
      successMessage: String(content.assetRemovedSuccess),
      errorMessage: String(content.failedToRemove),
    });
  };

  return (
    <>
      <ListView
        loading={loading}
        items={assets}
        total={total}
        pageSize={pageSize}
        currentPage={currentPage}
        onPaginationChange={onPaginationChange}
        onAddButtonClick={openAddModal}
        addButtonIcon={<Package className="h-4 w-4" />}
        addButtonText={String(content.addAsset)}
        addButtonDisabled={addModalOpened}
        filterBar={assets.length > 0 ? (
          <div className="flex items-center justify-end px-1">
            <div className="flex items-center gap-2">
              <Checkbox
                id="select-all-assets-page"
                checked={isAllPageSelected}
                onCheckedChange={toggleSelectAllPage}
              />
              <Label htmlFor="select-all-assets-page" className="text-xs text-muted-foreground cursor-pointer select-none">
                {String(content.selectAll || 'Select all on page')}
              </Label>
            </div>
          </div>
        ) : null}
        batchActionBar={selectedCount > 0 ? (
          <div className="flex items-center justify-between gap-3 p-3.5 rounded-2xl bg-primary/5 border border-primary/20 shadow-xs flex-wrap">
            <div className="flex items-center gap-3">
              <div className="flex items-center gap-2">
                <Badge variant="default" className="rounded-lg font-bold">
                  {selectedCount}
                </Badge>
                <span className="text-xs font-semibold text-foreground">
                  {String(content.selectedAssets || 'assets selected')}
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

            <div className="flex items-center gap-2">
              <Button
                size="sm"
                variant="outline"
                onClick={() => setAssignModalOpened(true)}
                className="gap-2 rounded-xl font-semibold shadow-xs"
              >
                <FolderPlus className="h-4 w-4" />
                <span>{String(content.addToParentAsset || 'Add to Existing Asset')}</span>
              </Button>

              <Button
                size="sm"
                onClick={() => setGroupModalOpened(true)}
                className="gap-2 rounded-xl font-semibold shadow-sm shadow-primary/20"
              >
                <Boxes className="h-4 w-4" />
                <span>{String(content.groupAssets || 'Group into New Asset')}</span>
              </Button>
            </div>
          </div>
        ) : null}
        renderItemPrefix={(asset) => (
          <Checkbox
            checked={Boolean(selectedAssetsMap[asset.id])}
            onCheckedChange={() => toggleSelectAsset(asset)}
            aria-label={`Select asset ${asset.name}`}
          />
        )}
        renderItemMainContent={(asset) => <AssetMainContent asset={asset} />}
        renderItemAmount={(asset) => <AssetAmount asset={asset} />}
        renderItemDetails={(asset) => (
          <AssetDetails
            asset={asset}
            onUnlinkChild={unlinkChild}
            onViewInvestments={handleViewInvestments}
          />
        )}
        renderItemActions={(asset) => (
          <>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setInvestAssetId(asset.id)}
              className="h-8 px-2.5 gap-1.5 rounded-xl text-xs font-medium text-emerald-600 dark:text-emerald-400 hover:text-emerald-700 hover:bg-emerald-500/10"
              title={String(content.investInAsset || 'Log Investment')}
            >
              <DollarSign className="h-3.5 w-3.5" />
              <span className="hidden sm:inline">{String(content.investInAsset || 'Invest')}</span>
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => openLinkModal(asset)}
              className="h-8 px-2.5 gap-1.5 rounded-xl text-xs font-medium text-muted-foreground hover:text-foreground hover:bg-accent/80"
              title="Link Component Asset"
            >
              <Link2 className="h-3.5 w-3.5" />
              <span className="hidden sm:inline">Link</span>
            </Button>
            <EditButton onClick={() => openEditModal(asset.id)} />
            <RemoveButton onRemove={() => handleRemove(asset.id)} />
          </>
        )}
      />

      <AddAssetModal
        projectId={projectId}
        open={addModalOpened}
        onCancel={closeAddModal}
        onSuccess={() => onAddSuccess(String(content.assetAddedSuccess))}
      />

      <EditAssetModal
        projectId={projectId}
        assetId={assetIdToEdit}
        open={!!assetIdToEdit}
        onCancel={closeEditModal}
        onSuccess={() => onEditSuccess(String(content.assetUpdatedSuccess))}
      />

      {linkModalAsset && (
        <LinkChildAssetModal
          open={Boolean(linkModalAsset)}
          onClose={closeLinkModal}
          onLink={linkChild}
          parentAsset={linkModalAsset}
          allAssets={assets}
        />
      )}

      {investAssetId && (
        <AddPaymentModal
          open={Boolean(investAssetId)}
          onCancel={() => setInvestAssetId('')}
          onSuccess={() => {
            setInvestAssetId('');
            refresh();
          }}
          assets={assets}
          defaultAssetId={investAssetId}
        />
      )}

      {groupModalOpened && (
        <GroupAssetsModal
          open={groupModalOpened}
          projectId={projectId}
          selectedAssets={selectedAssets}
          onCancel={() => setGroupModalOpened(false)}
          onSuccess={() => {
            setGroupModalOpened(false);
            clearSelection();
            refresh();
          }}
        />
      )}

      {assignModalOpened && (
        <AssignToParentModal
          open={assignModalOpened}
          projectId={projectId}
          selectedAssets={selectedAssets}
          allAssets={assets}
          onCancel={() => setAssignModalOpened(false)}
          onSuccess={() => {
            setAssignModalOpened(false);
            clearSelection();
            refresh();
          }}
        />
      )}
    </>
  );
}

export default Assets;
