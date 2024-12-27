import { useParams } from 'react-router-dom';
import React, { useEffect, useState } from 'react';
import { Space, Tag, Typography } from 'antd';
import { CarryOutOutlined } from '@ant-design/icons';
import styled from 'styled-components';
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
import './Assets.css';

const { Text } = Typography;

const MainContent = styled.div`
  display: flex;
  flex-direction: column;
  gap: 4px;
`;

const HeaderRow = styled.div`
  display: flex;
  align-items: baseline;
  gap: 8px;
  flex-wrap: nowrap;
`;

const TagsRow = styled.div`
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
`;

export function Assets() {
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
    setFilter({
      projectId,
      limit: PAGE_SIZE,
      offset: DEFAULT_OFFSET,
    });
  };

  const onEditSuccess = () => {
    setAssetIdToEdit('');
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
        setFilter({
          projectId,
          limit: PAGE_SIZE,
          offset: DEFAULT_OFFSET,
        });
      })
      .catch((error) => {
        console.error(error);
      });
  };

  const renderAssetMainContent = (asset) => (
    <MainContent>
      <HeaderRow>
        <Text type="secondary" style={{ fontSize: '12px', whiteSpace: 'nowrap' }}>
          {asset.acquiredAt}
        </Text>
        <Text strong>{asset.name}</Text>
      </HeaderRow>
      <TagsRow>
        <Tag>{asset.category}</Tag>
        <Tag>{asset.type}</Tag>
      </TagsRow>
      <Text type="secondary">{asset.description}</Text>
    </MainContent>
  );

  const renderAssetAmount = (asset) => (
    `${asset.price} ${asset.currency}`
  );

  const renderAssetDetails = (asset) => (
    <div className="details-grid">
      <DetailItem label="ID">
        <CopyableText text={asset.id} truncate />
      </DetailItem>
      <DetailItem label="Type">
        <Text>{asset.type}</Text>
      </DetailItem>
      <DetailItem label="Category">
        <Text>{asset.category}</Text>
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
        addButtonIcon={<CarryOutOutlined />}
        addButtonText="Add Asset"
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
