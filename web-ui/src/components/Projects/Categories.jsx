import React, { useEffect, useState, useCallback } from 'react';
import { useParams } from 'react-router-dom';
import { Skeleton } from '@/components/ui/skeleton';
import { PieChart } from 'lucide-react';
import { ListView } from './ListView';
import AddCategoryModal from './AddCategoryModal';
import { categoriesRepository } from '../../api';
import { DEFAULT_OFFSET, PAGE_SIZE } from '../../constants';
import useCategories from '../../hooks/categories';
import { RemoveButton } from './ListView/RemoveButton';
import { DetailItem } from './ListView/DetailItem';
import { CopyableText } from './ListView/CopyableText';
import { toast } from 'sonner';
import './Payments.css';

const renderCategoryMainContent = (category) => (
  <div>
    <span className="font-semibold text-base text-foreground">{category.name}</span>
  </div>
);

const renderCategoryDetails = (category) => (
  <div className="space-y-2">
    <DetailItem label="ID">
      <CopyableText text={category.id} truncate />
    </DetailItem>
    <DetailItem label="Description">
      <span className="text-sm text-muted-foreground">{category.description || 'No description'}</span>
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
      toast.success('Category removed successfully');
      resetFilter();
    } catch (error) {
      toast.error(`Failed to remove category: ${error.message}`);
      console.error('Failed to remove category:', error);
    }
  };

  const renderCategoryActions = (category) => (
    <RemoveButton onRemove={() => handleRemoveCategory(category.id)} />
  );

  if (loading) {
    return <Skeleton className="h-48 w-full rounded-xl" />;
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
        addButtonIcon={<PieChart className="h-4 w-4" />}
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
