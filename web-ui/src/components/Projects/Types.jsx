import React, { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { Skeleton, Tag } from 'antd';
import { BuildOutlined } from '@ant-design/icons';
import { ListView } from './ListView';
import AddTypeModal from './AddTypeModal';
import { typesRepository } from '../../api';
import { DEFAULT_OFFSET, PAGE_SIZE } from '../../constants';
import useTypes from '../../hooks/types';
import { RemoveButton } from './ListView/RemoveButton';

export default function Types() {
  const { projectId } = useParams();
  const [ loading, types, total, setFilter ] = useTypes();
  const [ addModalOpened, setAddModalOpen ] = useState(false);
  const [ currentPage, setCurrentPage ] = useState(1);

  const onPaginationChange = (nextPage) => {
    setCurrentPage(nextPage);
  };

  useEffect(() => {
    setFilter({
      projectId,
      limit: PAGE_SIZE,
      offset: (currentPage - 1) * PAGE_SIZE,
    });
  }, [ currentPage ]);

  useEffect(() => {
    setFilter({
      projectId,
      limit: PAGE_SIZE,
      offset: DEFAULT_OFFSET,
    });
  }, [ projectId, setFilter ]);

  const onAddTypeClick = () => {
    if (!addModalOpened) {
      setAddModalOpen(true);
    }
  };
  const onCancel = () => setAddModalOpen(false);
  const onSucces = () => {
    setAddModalOpen(false);
    setFilter({
      projectId,
      limit: PAGE_SIZE,
      offset: DEFAULT_OFFSET,
    });
  };

  const onRemoveButtonClick = (typeId) => {
    typesRepository.removeType(projectId, typeId)
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

  return loading ? <Skeleton active/> : (
    <div>
      <ListView
        loading={ loading }
        items={ types }
        total={ total }
        pageSize={ PAGE_SIZE }
        currentPage={ currentPage }
        onPaginationChange={ onPaginationChange }
        onAddButtonClick={ onAddTypeClick }
        addButtonIcon={ <BuildOutlined/> }
        addButtonText="Add Type"
        addButtonDisabled={ addModalOpened }
        renderItemMainContent={ (type) => (
          <div>
            <Tag>{ type.category }</Tag>
            <span>{ type.name }</span>
          </div>
        ) }
        renderItemDetails={ (type) => (
          <div>
            <span>{ type.description }</span>
          </div>
        ) }
        renderItemActions={ (type) => (
          <RemoveButton onRemove={ () => onRemoveButtonClick(type.id) }/>
        ) }
      />

      <AddTypeModal open={ addModalOpened } onSuccess={ onSucces } onCancel={ onCancel }/>
    </div>
  );
}
