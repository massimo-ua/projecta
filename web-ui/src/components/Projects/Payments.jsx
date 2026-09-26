import React from 'react';
import PropTypes from 'prop-types';
import { useParams } from 'react-router-dom';
import { useIntlayer } from 'react-intlayer';
import { DollarSign, Calendar } from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';
import usePayments from '../../hooks/payments';
import AddPaymentModal from './AddPaymentModal';
import EditPaymentModal from './EditPaymentModal';
import { ListView } from './ListView';
import { EditButton } from './ListView/EditButton';
import { RemoveButton } from './ListView/RemoveButton';
import { CopyableText } from './ListView/CopyableText';
import { DetailItem } from './ListView/DetailItem';
import './Payments.css';

function PaymentMainContent({ payment }) {
  return (
    <div className="flex flex-col gap-1.5">
      <div className="flex items-baseline gap-2 flex-wrap">
        <span className="flex items-center gap-1 text-xs text-muted-foreground font-mono">
          <Calendar className="h-3 w-3" />
          {payment.paymentDate}
        </span>
        <span className="font-semibold text-base text-foreground">{payment.description}</span>
      </div>
      <div className="flex gap-1.5 flex-wrap">
        <Badge variant="outline" className="rounded-md text-[11px] font-medium border-border/70 text-muted-foreground">{payment.category}</Badge>
        <Badge variant="secondary" className="rounded-md text-[11px] font-medium">{payment.type}</Badge>
      </div>
    </div>
  );
}

PaymentMainContent.propTypes = {
  payment: PropTypes.shape({
    paymentDate: PropTypes.string,
    description: PropTypes.string,
    category: PropTypes.string,
    type: PropTypes.string,
  }).isRequired,
};

function PaymentAmount({ payment }) {
  return (
    <div className="flex flex-col items-end gap-0.5">
      <span
        className={cn(
          'px-2.5 py-1 rounded-xl text-sm font-bold tracking-tight inline-block',
          payment.isDownPayment
            ? 'bg-rose-500/10 text-rose-600 dark:text-rose-400 border border-rose-500/20'
            : 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20',
        )}
      >
        {payment.formattedAmount}
      </span>
      {payment.hasDifferentHomeCurrency && payment.formattedHomeAmount && (
        <span className="text-[11px] text-muted-foreground font-mono">
          {payment.formattedHomeAmount}
        </span>
      )}
    </div>
  );
}

PaymentAmount.propTypes = {
  payment: PropTypes.shape({
    isDownPayment: PropTypes.bool,
    formattedAmount: PropTypes.string,
    hasDifferentHomeCurrency: PropTypes.bool,
    formattedHomeAmount: PropTypes.string,
  }).isRequired,
};

function PaymentDetails({ payment, typeLabel, categoryLabel }) {
  return (
    <div className="space-y-2">
      <DetailItem label="ID">
        <CopyableText text={payment.id} truncate />
      </DetailItem>
      <DetailItem label={typeLabel}>
        <span className="text-sm text-foreground">{payment.type}</span>
      </DetailItem>
      <DetailItem label={categoryLabel}>
        <span className="text-sm text-foreground">{payment.category}</span>
      </DetailItem>
    </div>
  );
}

PaymentDetails.propTypes = {
  payment: PropTypes.shape({
    id: PropTypes.string,
    type: PropTypes.string,
    category: PropTypes.string,
  }).isRequired,
  typeLabel: PropTypes.string.isRequired,
  categoryLabel: PropTypes.string.isRequired,
};

export function Payments() {
  const content = useIntlayer('payments');
  const { projectId } = useParams();
  const {
    loading,
    payments,
    total,
    currentPage,
    pageSize,
    types,
    addModalOpened,
    paymentIdToEdit,
    onPaginationChange,
    openAddModal,
    closeAddModal,
    onAddSuccess,
    openEditModal,
    closeEditModal,
    onEditSuccess,
    removePayment,
  } = usePayments(projectId);

  const handleRemove = (paymentId) => {
    removePayment(paymentId, {
      successMessage: String(content.paymentRemovedSuccess),
      errorMessage: String(content.failedToRemove),
    });
  };

  return (
    <>
      <ListView
        loading={loading}
        items={payments}
        total={total}
        pageSize={pageSize}
        currentPage={currentPage}
        onPaginationChange={onPaginationChange}
        onAddButtonClick={openAddModal}
        addButtonIcon={<DollarSign className="h-4 w-4" />}
        addButtonText={String(content.addPayment)}
        addButtonDisabled={addModalOpened}
        renderItemMainContent={(payment) => <PaymentMainContent payment={payment} />}
        renderItemAmount={(payment) => <PaymentAmount payment={payment} />}
        renderItemDetails={(payment) => (
          <PaymentDetails
            payment={payment}
            typeLabel={String(content.typeLabel)}
            categoryLabel={String(content.categoryLabel)}
          />
        )}
        renderItemActions={(payment) => (
          <>
            <EditButton onClick={() => openEditModal(payment.id)} />
            <RemoveButton onRemove={() => handleRemove(payment.id)} />
          </>
        )}
      />

      <AddPaymentModal
        types={types}
        projectId={projectId}
        open={addModalOpened}
        onCancel={closeAddModal}
        onSuccess={() => onAddSuccess(String(content.paymentAddedSuccess))}
      />

      <EditPaymentModal
        types={types}
        projectId={projectId}
        paymentId={paymentIdToEdit}
        open={!!paymentIdToEdit}
        onCancel={closeEditModal}
        onSuccess={() => onEditSuccess(String(content.paymentUpdatedSuccess))}
      />
    </>
  );
}

export default Payments;
