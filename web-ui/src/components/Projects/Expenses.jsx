import { useParams } from 'react-router-dom';
import React, { useEffect, useState } from 'react';
import { Button, Skeleton, Table, Tag } from 'antd';
import useExpenses from '../../hooks/expenses';
import { DollarOutlined } from '@ant-design/icons';
import AddExpenseModal from './AddExpenseModal';
import RemoveExpenseButton from './RemoveExpenseButton';
import { expensesRepository } from '../../api';
import { DEFAULT_OFFSET, PAGE_SIZE } from '../../constants';

const columns = [
  {
    title: 'ID',
    dataIndex: 'id',
    key: 'ID',
  },
  {
    title: 'Date',
    dataIndex: 'expenseDate',
    key: 'expenseDate',
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
    render: (_, expense) => (<Tag>{expense.category}</Tag>),
  },
  {
    title: 'Type',
    dataIndex: 'type',
    key: 'type',
  },
  {
    title: 'Amount',
    dataIndex: 'amount',
    key: 'amount',
  },
  {
    title: 'Currency',
    dataIndex: 'currency',
    key: 'currency',
  },
];

export function Expenses() {
  const { projectId } = useParams();
  const [loading, expenses, total, setFilter] = useExpenses();
  const [addModalOpened, setAddModalOpen] = useState(false);
  const [currentPage, setCurrentPage] = useState(1);

  const onPaginationChange = (nextPage) => {
    setCurrentPage(nextPage);
  };
  const onAddExpenseClick = () => {
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

  const onRemoveButtonClick = (expenseId) => {
    expensesRepository.removeExpense(projectId, expenseId)
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
    <Button disabled={addModalOpened} style={{ margin: '10px' }} icon={<DollarOutlined />} type="primary" onClick={onAddExpenseClick}>Add Expense</Button>
    <Table
      dataSource={expenses}
      columns={[...columns, {
        title: 'Action',
        key: 'action',
        render: (_, expense) => (
          <RemoveExpenseButton expenseId={expense.id} onClick={onRemoveButtonClick}/>
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
    <AddExpenseModal open={addModalOpened} onCancel={onCancel} onSuccess={onSucces} />
    </div>
  );
}
