import { useParams } from 'react-router-dom';
import React, { useEffect, useState } from 'react';
import { Skeleton, Table } from 'antd';
import useCategories from '../../hooks/categories';
import { DEFAULT_OFFSET, PAGE_SIZE } from '../../constants';

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
  const [loading, categories, total, setFilter] = useCategories();
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
  }, []);
  return loading ? <Skeleton active /> : (
    <Table
      dataSource={categories}
      columns={columns}
      pagination={{
        total,
        current: currentPage,
        pageSize: PAGE_SIZE,
        position: [PAGE_SIZE < total ? 'bottomRight' : 'none'],
        onChange: onPaginationChange,
      }}
    />
  );
}
