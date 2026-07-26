import React, { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { Skeleton } from '@/components/ui/skeleton';
import { Badge } from '@/components/ui/badge';
import { Boxes } from 'lucide-react';
import { ListView } from './ListView';
import AddTypeModal from './AddTypeModal';
import { typesRepository } from '../../api';
import { DEFAULT_OFFSET, PAGE_SIZE } from '../../constants';
import useTypes from '../../hooks/types';
import { RemoveButton } from './ListView/RemoveButton';
import { toast } from 'sonner';

export default function Types() {
  const { projectId } = useParams();
  const [loading, types, total, setFilter] = useTypes();
  const [addModalOpened, setAddModalOpen] = useState(false);
  const [currentPage, setCurrentPage] = useState(1);

  const onPaginationChange = (nextPage) => {
    setCurrentPage(nextPage);
  };

  useEffect(() => {
    setFilter({
      projectId,
      limit: PAGE_SIZE,
      offset: (currentPage - 1) * PAGE_SIZE,
    });
  }, [currentPage]);

  useEffect(() => {
    setFilter({
      projectId,
      limit: PAGE_SIZE,
      offset: DEFAULT_OFFSET,
    });
  }, [projectId, setFilter]);

  const onAddTypeClick = () => {
    if (!addModalOpened) {
      setAddModalOpen(true);
    }
  };
  const onCancel = () => setAddModalOpen(false);
  const onSucces = () => {
    setAddModalOpen(false);
    toast.success('Type added successfully');
    setFilter({
      projectId,
      limit: PAGE_SIZE,
      offset: DEFAULT_OFFSET,
    });
  };

  const onRemoveButtonClick = (typeId) => {
    typesRepository.removeType(projectId, typeId)
      .then(() => {
        toast.success('Type removed successfully');
        setFilter({
          projectId,
          limit: PAGE_SIZE,
          offset: DEFAULT_OFFSET,
        });
      })
      .catch((error) => {
        toast.error(`Failed to remove type: ${error.message}`);
        console.error(error);
      });
  };

  return loading ? <Skeleton className="h-48 w-full rounded-xl" /> : (
    <div>
      <ListView
        loading={loading}
        items={types}
        total={total}
        pageSize={PAGE_SIZE}
        currentPage={currentPage}
        onPaginationChange={onPaginationChange}
        onAddButtonClick={onAddTypeClick}
        addButtonIcon={<Boxes className="h-4 w-4" />}
        addButtonText="Add Type"
        addButtonDisabled={addModalOpened}
        renderItemMainContent={(type) => (
          <div className="flex items-center gap-2">
            <Badge variant="secondary">{type.category}</Badge>
            <span className="font-semibold text-base text-foreground">{type.name}</span>
          </div>
        )}
        renderItemDetails={(type) => (
          <div>
            <span className="text-sm text-muted-foreground">{type.description || 'No description'}</span>
          </div>
        )}
        renderItemActions={(type) => (
          <RemoveButton onRemove={() => onRemoveButtonClick(type.id)} />
        )}
      />

      <AddTypeModal open={addModalOpened} onSuccess={onSucces} onCancel={onCancel} />
    </div>
  );
}
