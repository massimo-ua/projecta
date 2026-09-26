import React from 'react';
import PropTypes from 'prop-types';
import { useParams } from 'react-router-dom';
import { useIntlayer } from 'react-intlayer';
import { Package, Calendar } from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import useAssets from '../../hooks/assets';
import AddAssetModal from './AddAssetModal';
import EditAssetModal from './EditAssetModal';
import { ListView } from './ListView';
import { EditButton } from './ListView/EditButton';
import { RemoveButton } from './ListView/RemoveButton';
import { CopyableText } from './ListView/CopyableText';
import { DetailItem } from './ListView/DetailItem';
import './Assets.css';

function AssetMainContent({ asset }) {
  return (
    <div className="flex flex-col gap-1.5">
      <div className="flex items-baseline gap-2 flex-wrap">
        <span className="flex items-center gap-1 text-xs text-muted-foreground font-mono">
          <Calendar className="h-3 w-3" />
          {asset.acquiredAt}
        </span>
        <span className="font-semibold text-base text-foreground">{asset.name}</span>
      </div>
      <div className="flex gap-1.5 flex-wrap">
        <Badge variant="outline" className="rounded-md text-[11px] font-medium border-border/70 text-muted-foreground">{asset.category}</Badge>
        <Badge variant="secondary" className="rounded-md text-[11px] font-medium">{asset.type}</Badge>
      </div>
      {asset.description && (
        <span className="text-xs text-muted-foreground/80 mt-0.5 line-clamp-1">{asset.description}</span>
      )}
    </div>
  );
}

AssetMainContent.propTypes = {
  asset: PropTypes.shape({
    acquiredAt: PropTypes.string,
    name: PropTypes.string,
    category: PropTypes.string,
    type: PropTypes.string,
    description: PropTypes.string,
  }).isRequired,
};

function AssetAmount({ asset }) {
  return (
    <div className="flex flex-col items-end gap-0.5">
      <span className="px-2.5 py-1 rounded-xl text-sm font-bold tracking-tight inline-block bg-primary/10 text-primary border border-primary/20">
        {asset.formattedPrice}
      </span>
      {asset.hasDifferentHomeCurrency && asset.formattedHomeAmount && (
        <span className="text-[11px] text-muted-foreground font-mono">
          {asset.formattedHomeAmount}
        </span>
      )}
    </div>
  );
}

AssetAmount.propTypes = {
  asset: PropTypes.shape({
    formattedPrice: PropTypes.string,
    hasDifferentHomeCurrency: PropTypes.bool,
    formattedHomeAmount: PropTypes.string,
  }).isRequired,
};

function AssetDetails({ asset, typeLabel, categoryLabel }) {
  return (
    <div className="space-y-2">
      <DetailItem label="ID">
        <CopyableText text={asset.id} truncate />
      </DetailItem>
      <DetailItem label={typeLabel}>
        <span className="text-sm text-foreground">{asset.type}</span>
      </DetailItem>
      <DetailItem label={categoryLabel}>
        <span className="text-sm text-foreground">{asset.category}</span>
      </DetailItem>
    </div>
  );
}

AssetDetails.propTypes = {
  asset: PropTypes.shape({
    id: PropTypes.string,
    type: PropTypes.string,
    category: PropTypes.string,
  }).isRequired,
  typeLabel: PropTypes.string.isRequired,
  categoryLabel: PropTypes.string.isRequired,
};

export function Assets() {
  const content = useIntlayer('assets');
  const { projectId } = useParams();
  const {
    loading,
    assets,
    total,
    currentPage,
    pageSize,
    types,
    addModalOpened,
    assetIdToEdit,
    onPaginationChange,
    openAddModal,
    closeAddModal,
    onAddSuccess,
    openEditModal,
    closeEditModal,
    onEditSuccess,
    removeAsset,
  } = useAssets(projectId);

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
        renderItemMainContent={(asset) => <AssetMainContent asset={asset} />}
        renderItemAmount={(asset) => <AssetAmount asset={asset} />}
        renderItemDetails={(asset) => (
          <AssetDetails
            asset={asset}
            typeLabel={String(content.typeLabel)}
            categoryLabel={String(content.categoryLabel)}
          />
        )}
        renderItemActions={(asset) => (
          <>
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
        types={types}
      />

      <EditAssetModal
        projectId={projectId}
        assetId={assetIdToEdit}
        open={!!assetIdToEdit}
        onCancel={closeEditModal}
        onSuccess={() => onEditSuccess(String(content.assetUpdatedSuccess))}
        types={types}
      />
    </>
  );
}

export default Assets;
