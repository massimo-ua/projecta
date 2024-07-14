import { useParams } from 'react-router-dom';
import React, { useEffect, useState } from 'react';
import { Button, Skeleton, Table, Tag } from 'antd';
import useExpenses from '../../hooks/expenses';
import { DollarOutlined } from '@ant-design/icons';
import AddExpenseModal from './AddExpenseModal';
import RemoveExpenseButton from './RemoveExpenseButton';
import { expensesRepository } from '../../api';

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

  const onCancel = () => setAddModalOpen(false);
  const onSucces = () => {
    setAddModalOpen(false);
    setFilter({
      projectId,
      limit: 10,
      offset: 0,
    });
  };

  const onRemoveButtonClick = (expenseId) => {
    expensesRepository.removeExpense(projectId, expenseId)
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
      onChange={onChange}
      showSorterTooltip={{
        target: 'sorter-icon',
      }}
    />
    <AddExpenseModal open={addModalOpened} onCancel={onCancel} onSuccess={onSucces} />
    </div>
  );
}
