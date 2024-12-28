import React, { useEffect, useState, useCallback } from 'react';
import { useParams } from 'react-router-dom';
import { Skeleton, Typography, notification } from 'antd';
import { PieChartOutlined } from '@ant-design/icons';
import { ListView } from './ListView';
import AddCategoryModal from './AddCategoryModal';
import { categoriesRepository } from '../../api';
import { DEFAULT_OFFSET, PAGE_SIZE } from '../../constants';
import useCategories from '../../hooks/categories';
import { RemoveButton } from './ListView/RemoveButton';
import { DetailItem } from './ListView/DetailItem';
import { CopyableText } from './ListView/CopyableText';
import './Payments.css';

const { Text } = Typography;

const renderCategoryMainContent = (category) => (
  <div>
    <span>{category.name}</span>
  </div>
);

const renderCategoryDetails = (category) => (
  <div className="details-grid">
    <DetailItem label="ID">
      <CopyableText text={category.id} truncate />
    </DetailItem>
    <DetailItem label="Description">
      <Text>{category.description}</Text>
    </DetailItem>
  </div>
);

export function Categories() {
  const { projectId } = useParams();
  const [loading, categories, total, setFilter] = useCategories();
  const [addModalOpened, setAddModalOpen] = useState(false);
  const [currentPage, setCurrentPage] = useState(1);

  const resetFilter = useCallback(() => {
    setFilter({
      projectId,
      limit: PAGE_SIZE,
      offset: DEFAULT_OFFSET,
    });
  }, [projectId, setFilter]);

  const updatePageFilter = useCallback((page) => {
    setFilter({
      projectId,
      limit: PAGE_SIZE,
      offset: (page - 1) * PAGE_SIZE,
    });
  }, [projectId, setFilter]);

  useEffect(() => {
    updatePageFilter(currentPage);
  }, [currentPage, updatePageFilter]);

  useEffect(() => {
    resetFilter();
  }, [resetFilter]);

  const onAddCategoryClick = () => {
    !addModalOpened && setAddModalOpen(true);
  };

  const handleModalClose = () => {
    setAddModalOpen(false);
  };

  const handleModalSuccess = () => {
    handleModalClose();
    resetFilter();
  };

  const handleRemoveCategory = async (categoryId) => {
    try {
      await categoriesRepository.removeCategory(projectId, categoryId);
      resetFilter();
    } catch (error) {
      notification.error({
        message: 'Failed to remove category',
        description: `Category ${categoryId} could not be removed. ${error.message}`,
      });
      console.error('Failed to remove category:', error);
    }
  };

  const renderCategoryActions = (category) => (
    <RemoveButton onRemove={() => handleRemoveCategory(category.id)} />
  );

  if (loading) {
    return <Skeleton active />;
  }

  return (
    <div>
      <ListView
        loading={loading}
        items={categories}
        total={total}
        pageSize={PAGE_SIZE}
        currentPage={currentPage}
        onPaginationChange={setCurrentPage}
        onAddButtonClick={onAddCategoryClick}
        addButtonIcon={<PieChartOutlined />}
        addButtonText="Add Category"
        addButtonDisabled={addModalOpened}
        renderItemMainContent={renderCategoryMainContent}
        renderItemDetails={renderCategoryDetails}
        renderItemActions={renderCategoryActions}
      />

      <AddCategoryModal
        open={addModalOpened}
        onSuccess={handleModalSuccess}
        onCancel={handleModalClose}
      />
    </div>
  );
}
