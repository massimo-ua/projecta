import { useParams } from 'react-router-dom';
import React, { useEffect, useState } from 'react';
import { Button, Skeleton, Table } from 'antd';
import useExpenses from '../../hooks/expenses';
import { DollarOutlined } from '@ant-design/icons';
import AddExpenseModal from './AddExpenseModal';

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
  const [loading, expenses, setFilter] = useExpenses();
  const [addModalOpened, setAddModalOpen] = useState(false);

  const onChange = (e) => console.log('ExpensesTable.onChange', e);
  const onAddExpenseClick = () => {
    if (!addModalOpened) {
      setAddModalOpen(true);
    }
  };

  useEffect(() => {
    setFilter({
      projectId,
      limit: 10,
      offset: 0,
    });
  }, []);

  return loading ? <Skeleton active /> : (
    <div>
    <Button disabled={addModalOpened} style={{ margin: '10px' }} icon={<DollarOutlined />} type="primary" onClick={onAddExpenseClick}>Add Expense</Button>
    <Table
      dataSource={expenses}
      columns={columns}
      onChange={onChange}
      showSorterTooltip={{
        target: 'sorter-icon',
      }}
    />
    <AddExpenseModal open={addModalOpened} setOpen={setAddModalOpen} />
    </div>
  );
}
