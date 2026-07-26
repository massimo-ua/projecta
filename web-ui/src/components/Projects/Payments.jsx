import { useParams } from 'react-router-dom';
import React, { useEffect, useState } from 'react';
import { Badge } from '@/components/ui/badge';
import { DollarSign } from 'lucide-react';
import usePayments from '../../hooks/payments';
import useTypes from '../../hooks/types';
import AddPaymentModal from './AddPaymentModal';
import { paymentRepository } from '../../api';
import { DEFAULT_OFFSET, PAGE_SIZE } from '../../constants';
import EditPaymentModal from './EditPaymentModal';
import { ListView } from './ListView';
import { EditButton } from './ListView/EditButton';
import { RemoveButton } from './ListView/RemoveButton';
import { CopyableText } from './ListView/CopyableText';
import { DetailItem } from './ListView/DetailItem';
import { toast } from 'sonner';
import './Payments.css';

export function Payments() {
  const { projectId } = useParams();
  const [loading, payments, total, setFilter] = usePayments();
  const [addModalOpened, setAddModalOpen] = useState(false);
  const [paymentIdToEdit, setPaymentIdToEdit] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const [, types, , setTypesFilter] = useTypes();

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
    setTypesFilter({
      projectId,
      limit: 100,
      offset: 0,
    });
  }, [currentPage, projectId, setFilter]);

  const onAddCancel = () => setAddModalOpen(false);
  const onAddSuccess = () => {
    setAddModalOpen(false);
    toast.success('Payment added successfully');
    setFilter({
      projectId,
      limit: PAGE_SIZE,
      offset: DEFAULT_OFFSET,
    });
  };

  const onRemoveButtonClick = (paymentId) => {
    paymentRepository.removePayment(projectId, paymentId)
      .then(() => {
        toast.success('Payment removed successfully');
        setFilter({
          projectId,
          limit: PAGE_SIZE,
          offset: DEFAULT_OFFSET,
        });
      })
      .catch((error) => {
        toast.error(`Failed to remove payment: ${error.message}`);
        console.error(error);
      });
  };

  const onEditSuccess = () => {
    setPaymentIdToEdit('');
    toast.success('Payment updated successfully');
    setFilter({
      projectId,
      limit: PAGE_SIZE,
      offset: DEFAULT_OFFSET,
    });
  };

  const onEditCancel = () => setPaymentIdToEdit('');

  const renderPaymentMainContent = (payment) => (
    <div className="flex flex-col gap-1">
      <div className="flex items-baseline gap-2 flex-wrap">
        <span className="text-xs text-muted-foreground whitespace-nowrap">
          {payment.paymentDate}
        </span>
        <span className="font-semibold text-base text-foreground">{payment.description}</span>
      </div>
      <div className="flex gap-1.5 flex-wrap">
        <Badge variant="outline">{payment.category}</Badge>
        <Badge variant="secondary">{payment.type}</Badge>
      </div>
    </div>
  );

  const renderPaymentAmount = (payment) => {
    const isDiff = payment.currency !== payment.homeCurrency;
    return (
      <div className="flex flex-col items-end">
        <span
          className={
            payment.kind === 'DOWN_PAYMENT'
              ? 'text-red-600 dark:text-red-400 font-bold'
              : 'text-green-600 dark:text-green-400 font-bold'
          }
        >
          {payment.amount} {payment.currency}
        </span>
        {isDiff && payment.homeAmount && (
          <span className="text-xs text-muted-foreground font-medium">
            ≈ {payment.homeAmount} {payment.homeCurrency}
          </span>
        )}
      </div>
    );
  };

  const renderPaymentDetails = (payment) => (
    <div className="space-y-2">
      <DetailItem label="ID">
        <CopyableText text={payment.id} truncate />
      </DetailItem>
      <DetailItem label="Type">
        <span className="text-sm text-foreground">{payment.type}</span>
      </DetailItem>
      <DetailItem label="Category">
        <span className="text-sm text-foreground">{payment.category}</span>
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
        addButtonIcon={<DollarSign className="h-4 w-4" />}
        addButtonText="Add Payment"
        addButtonDisabled={addModalOpened}
        renderItemMainContent={renderPaymentMainContent}
        renderItemAmount={renderPaymentAmount}
        renderItemDetails={renderPaymentDetails}
        renderItemActions={renderPaymentActions}
      />

      <AddPaymentModal
        types={types}
        projectId={projectId}
        open={addModalOpened}
        onCancel={onAddCancel}
        onSuccess={onAddSuccess}
      />

      <EditPaymentModal
        types={types}
        projectId={projectId}
        paymentId={paymentIdToEdit}
        open={!!paymentIdToEdit}
        onCancel={onEditCancel}
        onSuccess={onEditSuccess}
      />
    </>
  );
}
