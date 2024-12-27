import { useParams } from 'react-router-dom';
import React, { useEffect, useState } from 'react';
import { Space, Tag, Typography } from 'antd';
import { DollarOutlined } from '@ant-design/icons';
import styled from 'styled-components';
import usePayments from '../../hooks/payments';
import AddPaymentModal from './AddPaymentModal';
import { paymentRepository } from '../../api';
import { DEFAULT_OFFSET, PAGE_SIZE } from '../../constants';
import EditPaymentModal from './EditPaymentModal';
import { ListView } from './ListView';
import { EditButton } from './ListView/EditButton';
import { RemoveButton } from './ListView/RemoveButton';
import { CopyableText } from './ListView/CopyableText';
import { DetailItem } from './ListView/DetailItem';
import './Payments.css';

const { Text } = Typography;

const MainContent = styled.div`
  display: flex;
  flex-direction: column;
  gap: 4px;
`;

const HeaderRow = styled.div`
  display: flex;
  align-items: baseline;
  gap: 8px;
  flex-wrap: nowrap;
`;

const TagsRow = styled.div`
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
`;

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
  };

  useEffect(() => {
    setFilter({
      projectId,
      limit: PAGE_SIZE,
      offset: (currentPage - 1) * PAGE_SIZE,
    });
  }, [currentPage, projectId, setFilter]);

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
      });
  };

  const onEditSuccess = () => {
    setPaymentIdToEdit('');
    setFilter({
      projectId,
      limit: PAGE_SIZE,
      offset: DEFAULT_OFFSET,
    });
  };

  const onEditCancel = () => setPaymentIdToEdit('');

  const renderPaymentMainContent = (payment) => (
    <MainContent>
      <HeaderRow>
        <Text type="secondary" style={{ fontSize: '12px', whiteSpace: 'nowrap' }}>
          {payment.paymentDate}
        </Text>
        <Text strong>{payment.description}</Text>
      </HeaderRow>
      <TagsRow>
        <Tag>{payment.category}</Tag>
        <Tag>{payment.type}</Tag>
      </TagsRow>
    </MainContent>
  );

  const renderPaymentAmount = (payment) => (
    <span style={{
      color: payment.kind === 'DOWN_PAYMENT' ? 'red' : 'green',
    }}>
      {payment.amount} {payment.currency}
    </span>
  );

  const renderPaymentDetails = (payment) => (
    <div className="details-grid">
      <DetailItem label="ID">
        <CopyableText text={payment.id} truncate />
      </DetailItem>
      <DetailItem label="Type">
        <Text>{payment.type}</Text>
      </DetailItem>
      <DetailItem label="Category">
        <Text>{payment.category}</Text>
      </DetailItem>
    </div>
  );

  const renderPaymentActions = (payment) => (
    <>
      <EditButton onClick={() => onEditButtonClick(payment.id)} />
      <RemoveButton onRemove={() => onRemoveButtonClick(payment.id)} />
    </>
  );

  return (
    <>
      <ListView
        loading={loading}
        items={payments}
        total={total}
        pageSize={PAGE_SIZE}
        currentPage={currentPage}
        onPaginationChange={onPaginationChange}
        onAddButtonClick={onAddButtonClick}
        addButtonIcon={<DollarOutlined />}
        addButtonText="Add Payment"
        addButtonDisabled={addModalOpened}
        renderItemMainContent={renderPaymentMainContent}
        renderItemAmount={renderPaymentAmount}
        renderItemDetails={renderPaymentDetails}
        renderItemActions={renderPaymentActions}
      />

      <AddPaymentModal
        projectId={projectId}
        open={addModalOpened}
        onCancel={onAddCancel}
        onSuccess={onAddSuccess}
      />

      <EditPaymentModal
        projectId={projectId}
        paymentId={paymentIdToEdit}
        open={!!paymentIdToEdit}
        onCancel={onEditCancel}
        onSuccess={onEditSuccess}
      />
    </>
  );
}
