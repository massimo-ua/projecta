import { useParams } from 'react-router-dom';
import React, { useEffect, useState } from 'react';
import { useIntlayer } from 'react-intlayer';
import { Badge } from '@/components/ui/badge';
import { Package } from 'lucide-react';
import useAssets from '../../hooks/assets';
import useTypes from '../../hooks/types';
import AddAssetModal from './AddAssetModal';
import EditAssetModal from './EditAssetModal';
import { DEFAULT_OFFSET, PAGE_SIZE } from '../../constants';
import { ListView } from './ListView';
import { EditButton } from './ListView/EditButton';
import { RemoveButton } from './ListView/RemoveButton';
import { CopyableText } from './ListView/CopyableText';
import { DetailItem } from './ListView/DetailItem';
import { assetRepository } from '../../api';
import { toast } from 'sonner';
import './Assets.css';

export function Assets() {
  const content = useIntlayer('assets');
  const { projectId } = useParams();
  const [loading, assets, total, setFilter] = useAssets();
  const [, types, , setTypesFilter] = useTypes();
  const [addModalOpened, setAddModalOpen] = useState(false);
  const [assetIdToEdit, setAssetIdToEdit] = useState('');
  const [currentPage, setCurrentPage] = useState(1);

  const onPaginationChange = (nextPage) => {
    setCurrentPage(nextPage);
  };

  const onAddButtonClick = () => {
    if (!addModalOpened) {
      setAddModalOpen(true);
    }
  };

  const onEditButtonClick = (assetId) => {
    if (!assetIdToEdit) {
      setAssetIdToEdit(assetId);
    }
  };

  useEffect(() => {
    setFilter({
      projectId,
      limit: PAGE_SIZE,
      offset: (currentPage - 1) * PAGE_SIZE,
    });
    setTypesFilter({
      projectId,
      limit: 100,
      offset: 0,
    });
  }, [currentPage, projectId, setFilter, setTypesFilter]);

  const onAddCancel = () => setAddModalOpen(false);
  const onAddSuccess = () => {
    setAddModalOpen(false);
    toast.success(String(content.assetAddedSuccess));
    setFilter({
      projectId,
      limit: PAGE_SIZE,
      offset: DEFAULT_OFFSET,
    });
  };

  const onEditSuccess = () => {
    setAssetIdToEdit('');
    toast.success(String(content.assetUpdatedSuccess));
    setFilter({
      projectId,
      limit: PAGE_SIZE,
      offset: DEFAULT_OFFSET,
    });
  };

  const onEditCancel = () => setAssetIdToEdit('');

  const onRemoveButtonClick = (assetId) => {
    assetRepository.removeAsset(projectId, assetId)
      .then(() => {
        toast.success(String(content.assetRemovedSuccess));
        setFilter({
          projectId,
          limit: PAGE_SIZE,
          offset: DEFAULT_OFFSET,
        });
      })
      .catch((error) => {
        toast.error(`${String(content.failedToRemove)}: ${error.message}`);
        console.error(error);
      });
  };

  const renderAssetMainContent = (asset) => (
    <div className="flex flex-col gap-1">
      <div className="flex items-baseline gap-2 flex-wrap">
        <span className="text-xs text-muted-foreground whitespace-nowrap">
          {asset.acquiredAt}
        </span>
        <span className="font-semibold text-base text-foreground">{asset.name}</span>
      </div>
      <div className="flex gap-1.5 flex-wrap">
        <Badge variant="outline">{asset.category}</Badge>
        <Badge variant="secondary">{asset.type}</Badge>
      </div>
      {asset.description && (
        <span className="text-xs text-muted-foreground mt-0.5">{asset.description}</span>
      )}
    </div>
  );

  const renderAssetAmount = (asset) => {
    const isDiff = asset.currency !== asset.homeCurrency;
    return (
      <div className="flex flex-col items-end">
        <span className="font-bold text-foreground">
          {asset.price} {asset.currency}
        </span>
        {isDiff && asset.homeAmount && (
          <span className="text-xs text-muted-foreground font-medium">
            ≈ {asset.homeAmount} {asset.homeCurrency}
          </span>
        )}
      </div>
    );
  };

  const renderAssetDetails = (asset) => (
    <div className="space-y-2">
      <DetailItem label="ID">
        <CopyableText text={asset.id} truncate />
      </DetailItem>
      <DetailItem label={String(content.typeLabel)}>
        <span className="text-sm text-foreground">{asset.type}</span>
      </DetailItem>
      <DetailItem label={String(content.categoryLabel)}>
        <span className="text-sm text-foreground">{asset.category}</span>
      </DetailItem>
    </div>
  );

  const renderAssetActions = (asset) => (
    <>
      <EditButton onClick={() => onEditButtonClick(asset.id)} />
      <RemoveButton onRemove={() => onRemoveButtonClick(asset.id)} />
    </>
  );

  return (
    <>
      <ListView
        loading={loading}
        items={assets}
        total={total}
        pageSize={PAGE_SIZE}
        currentPage={currentPage}
        onPaginationChange={onPaginationChange}
        onAddButtonClick={onAddButtonClick}
        addButtonIcon={<Package className="h-4 w-4" />}
        addButtonText={String(content.addAsset)}
        addButtonDisabled={addModalOpened}
        renderItemMainContent={renderAssetMainContent}
        renderItemAmount={renderAssetAmount}
        renderItemDetails={renderAssetDetails}
        renderItemActions={renderAssetActions}
      />

      <AddAssetModal
        projectId={projectId}
        open={addModalOpened}
        onCancel={onAddCancel}
        onSuccess={onAddSuccess}
        types={types}
      />

      <EditAssetModal
        projectId={projectId}
        assetId={assetIdToEdit}
        open={!!assetIdToEdit}
        onCancel={onEditCancel}
        onSuccess={onEditSuccess}
        types={types}
      />
    </>
  );
}
