import { useParams } from 'react-router-dom';
import React, { useEffect, useState } from 'react';
import { Button, Row, Skeleton, Table, Tag } from 'antd';
import usePayments from '../../hooks/payments';
import { DollarOutlined, EditOutlined } from '@ant-design/icons';
import AddPaymentModal from './AddPaymentModal';
import RemovePaymentButton from './RemovePaymentButton';
import { paymentRepository } from '../../api';
import { DEFAULT_OFFSET, PAGE_SIZE } from '../../constants';
import EditPaymentModal from './EditPaymentModal';

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
  const [paymentIdToEdit, setPaymentIdToEdit] = useState('');
  const [currentPage, setCurrentPage] = useState(1);

  const onPaginationChange = (nextPage) => {
    setCurrentPage(nextPage);
  };
  const onAddButtonClick = () => {
    if (!addModalOpened) {
      setAddModalOpen(true);
    }
  };

  const onEditButtonClick = (paymentId) => {
    if (!paymentIdToEdit) {
      setPaymentIdToEdit(paymentId);
    }
  }

  useEffect(() => {
    setFilter({
      projectId,
      limit: PAGE_SIZE,
      offset: (currentPage - 1) * PAGE_SIZE,
    });
  }, [currentPage]);

  const onAddCancel = () => setAddModalOpen(false);
  const onAddSuccess = () => {
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

  const onEditSuccess = () => {
    setPaymentIdToEdit('');
    setFilter({
      projectId,
      limit: PAGE_SIZE,
      offset: DEFAULT_OFFSET,
    });
  }
  const onEditCancel = () => setPaymentIdToEdit('');

  return loading ? <Skeleton active /> : (
    <div>
    <Button disabled={addModalOpened} style={{ margin: '10px' }} icon={<DollarOutlined />} type="primary" onClick={onAddButtonClick}>Add Payment</Button>
    <Table
      dataSource={payments}
      columns={[...columns, {
        title: 'Action',
        key: 'action',
        render: (_, payment) => (
          <Row style={{ gap: '2px'}}>
            <EditOutlined onClick={() => onEditButtonClick(payment.id)}/>
            <RemovePaymentButton paymentId={payment.id} onClick={onRemoveButtonClick}/>
          </Row>
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
      <AddPaymentModal open={addModalOpened} onCancel={onAddCancel} onSuccess={onAddSuccess} />
      <EditPaymentModal paymentId={paymentIdToEdit} onCancel={onEditCancel} onSuccess={onEditSuccess} />
    </div>
  );
}
