import React, { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { Button, Skeleton, Table, Tag } from 'antd';

import useTypes from '../../hooks/types';
import { BuildOutlined } from '@ant-design/icons';
import AddTypeModal from './AddTypeModal.jsx';
import RemoveTypeButton from './RemoveTypeButton.jsx';
import { typesRepository } from '../../api/index.js';

const columns = [
  {
    title: 'ID',
    dataIndex: 'id',
    key: 'ID',
  },
  {
    title: 'Category',
    dataIndex: 'category',
    key: 'category',
    render: (_, type) => (<Tag>{type.category}</Tag>),
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
];

export default function Types() {
  const { projectId } = useParams();
  const [loading, types, setFilter] = useTypes();
  const [addModalOpened, setAddModalOpen] = useState(false);

  useEffect(() => {
    setFilter({
      projectId,
      limit: 10,
      offset: 0,
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
    setFilter({
      projectId,
      limit: 10,
      offset: 0,
    });
  };

  const onRemoveButtonClick = (typeId) => {
    typesRepository.removeType(projectId, typeId)
      .then(() => {
        setFilter({
          projectId,
          limit: 10,
          offset: 0,
        });
      })
      .catch((error) => {
        console.error(error);
      })
  };

  return loading ? <Skeleton active /> : (
    <div>
      <Button disabled={addModalOpened} style={{ margin: '10px' }} icon={<BuildOutlined />} type="primary" onClick={onAddTypeClick}>Add Type</Button>
      <Table
        dataSource={types}
        columns={[...columns,   {
          title: 'Actions',
          key: 'actions',
          render: (_, type) => (
            <RemoveTypeButton typeId={type.id} onClick={onRemoveButtonClick}/>
          ),
        }]}
        showSorterTooltip={{
          target: 'sorter-icon',
        }}
      />
      <AddTypeModal open={addModalOpened} onSuccess={onSucces} onCancel={onCancel} />
    </div>
  );
}
