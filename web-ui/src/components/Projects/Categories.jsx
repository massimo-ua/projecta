import { useParams } from 'react-router-dom';
import React, { useEffect } from 'react';
import { Skeleton, Table } from 'antd';
import useCategories from '../../hooks/categories.js';

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

export function Categories() {
  const { projectId } = useParams();
  const [loading, categories, setFilter] = useCategories();

  const onChange = console.log.bind('CategoriesTable.onChange');

  useEffect(() => {
    setFilter({
      projectId,
      limit: 10,
      offset: 0,
    });
  }, []);
  return loading ? <Skeleton active /> : (
    <Table
      dataSource={categories}
      columns={columns}
      onChange={onChange}
      showSorterTooltip={{
        target: 'sorter-icon',
      }}
    />
  );
}
