import { useParams } from 'react-router-dom';
import React, { useEffect, useState } from 'react';
import { Button, Skeleton, Table, Tag } from 'antd';
import usePayments from '../../hooks/payments';
import { DollarOutlined } from '@ant-design/icons';
import AddPaymentModal from './AddPaymentModal';
import RemovePaymentButton from './RemovePaymentButton';
import { paymentRepository } from '../../api';
import { DEFAULT_OFFSET, PAGE_SIZE } from '../../constants';

const columns = [
  {
    title: 'ID',
    dataIndex: 'id',
    key: 'ID',
  },
  {
    title: 'Date',
    dataIndex: 'paymentDate',
    key: 'paymentDate',
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
    render: (_, payment) => (<Tag>{payment.category}</Tag>),
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

export function Payments() {
  const { projectId } = useParams();
  const [loading, payments, total, setFilter] = usePayments();
  const [addModalOpened, setAddModalOpen] = useState(false);
  const [currentPage, setCurrentPage] = useState(1);

  const onPaginationChange = (nextPage) => {
    setCurrentPage(nextPage);
  };
  const onAddButtonClick = () => {
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

  const onRemoveButtonClick = (paymentId) => {
    paymentRepository.removePayment(projectId, paymentId)
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
    <Button disabled={addModalOpened} style={{ margin: '10px' }} icon={<DollarOutlined />} type="primary" onClick={onAddButtonClick}>Add Payment</Button>
    <Table
      dataSource={payments}
      columns={[...columns, {
        title: 'Action',
        key: 'action',
        render: (_, payment) => (
          <RemovePaymentButton paymentId={payment.id} onClick={onRemoveButtonClick}/>
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
    <AddPaymentModal open={addModalOpened} onCancel={onCancel} onSuccess={onSucces} />
    </div>
  );
}
