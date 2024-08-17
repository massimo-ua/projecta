import { useParams } from 'react-router-dom';
import React, { useEffect, useState } from 'react';
import { Button, Skeleton, Table, Tag } from 'antd';
import useAssets from '../../hooks/assets';
import { CarryOutOutlined } from '@ant-design/icons';
import AddAssetModal from './AddAssetModal';
import RemoveAssetButton from './RemoveAssetButton';
import { assetRepository } from '../../api';
import { DEFAULT_OFFSET, PAGE_SIZE } from '../../constants';

const columns = [
  {
    title: 'ID',
    dataIndex: 'id',
    key: 'ID',
  },
  {
    title: 'Date',
    dataIndex: 'acquiredAt',
    key: 'acquiredAt',
  },
  {
    title: 'Name',
    dataIndex: 'name',
    key: 'name',
  },
  {
    title: 'Description',
    dataIndex: 'description',
    key: 'description',
  },
  {
    title: 'Category',
    dataIndex: 'category',
    key: 'category',
    render: (_, asset) => (<Tag>{asset.category}</Tag>),
  },
  {
    title: 'Type',
    dataIndex: 'type',
    key: 'type',
  },
  {
    title: 'Price',
    dataIndex: 'price',
    key: 'price',
  },
  {
    title: 'Currency',
    dataIndex: 'currency',
    key: 'currency',
  },
];

export function Assets() {
  const { projectId } = useParams();
  const [loading, assets, total, setFilter] = useAssets();
  const [addModalOpened, setAddModalOpen] = useState(false);
  const [currentPage, setCurrentPage] = useState(1);

  const onPaginationChange = (nextPage) => {
    setCurrentPage(nextPage);
  };
  const onAddButtonClick = () => {
    if (!addModalOpened) {
      setAddModalOpen(true);
    }
  };

  useEffect(() => {
    setFilter({
      projectId,
      limit: PAGE_SIZE,
      offset: (currentPage - 1) * PAGE_SIZE,
    });
  }, [currentPage]);

  const onCancel = () => setAddModalOpen(false);
  const onSucces = () => {
    setAddModalOpen(false);
    setFilter({
      projectId,
      limit: PAGE_SIZE,
      offset: DEFAULT_OFFSET,
    });
  };

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
      })
  };

  return loading ? <Skeleton active /> : (
    <div>
    <Button disabled={addModalOpened} style={{ margin: '10px' }} icon={<CarryOutOutlined />} type="primary" onClick={onAddButtonClick}>Add Asset</Button>
    <Table
      dataSource={assets}
      columns={[...columns, {
        title: 'Action',
        key: 'action',
        render: (_, asset) => (
          <RemoveAssetButton assetId={asset.id} onClick={onRemoveButtonClick}/>
        ),
      }]}
      showSorterTooltip={{
        target: 'sorter-icon',
      }}
      pagination={{
        total,
        current: currentPage,
        pageSize: PAGE_SIZE,
        position: [PAGE_SIZE < total ? 'bottomRight' : 'none'],
        onChange: onPaginationChange,
      }}
    />
    <AddAssetModal open={addModalOpened} onCancel={onCancel} onSuccess={onSucces} />
    </div>
  );
}
