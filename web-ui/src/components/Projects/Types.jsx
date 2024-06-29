import React, { useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { useTypes } from '../../hooks/types.js';
import { Skeleton, Table } from 'antd';

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

export function Types() {
  const { projectId} = useParams();
  const [loading, types, setFilter] = useTypes();

  const onChange = console.log.bind('TypesTable.onChange');

  useEffect(() => {
    setFilter({
      projectId,
      limit: 10,
      offset: 0,
    });
  }, []);
  return loading ? <Skeleton active /> : <Table
    dataSource={types}
    columns={columns}
    onChange={onChange}
    showSorterTooltip={{
      target: 'sorter-icon',
    }}
  />;
}
