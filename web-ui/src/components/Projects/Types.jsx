import React, { useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { Skeleton, Table } from 'antd';

import useTypes from '../../hooks/types';

const columns = [
  {
    title: 'ID',
    dataIndex: 'id',
    key: 'ID',
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

  useEffect(() => {
    setFilter({
      projectId,
      limit: 10,
      offset: 0,
    });
  }, [projectId, setFilter]);

  return loading ? <Skeleton active /> : (
    <Table
      dataSource={types}
      columns={columns}
      showSorterTooltip={{
        target: 'sorter-icon',
      }}
    />
  );
}
