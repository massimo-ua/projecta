import React, { useEffect, useState } from 'react';
import PropTypes from 'prop-types';
import { useParams } from 'react-router-dom';
import { useIntlayer } from 'react-intlayer';
import { paymentRepository } from '../../api';
import { PaymentKind } from '../../constants';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog';
import { toast } from 'sonner';
import { Loader2 } from 'lucide-react';

const SUPPORTED_CURRENCIES = ['UAH', 'USD', 'EUR', 'PLN'];

export default function EditPaymentModal({
  onSuccess,
  onCancel,
  paymentId = null,
  types = [],
}) {
  const content = useIntlayer('payments');
  const { projectId } = useParams();

  const [typeId, setTypeId] = useState('');
  const [paymentKind, setPaymentKind] = useState('UPON_COMPLETION');
  const [amount, setAmount] = useState('');
  const [currency, setCurrency] = useState('UAH');
  const [paymentDate, setPaymentDate] = useState('');
  const [description, setDescription] = useState('');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!paymentId) return;

    paymentRepository
      .getPayment(projectId, paymentId)
      .then((payment) => {
        setTypeId(payment.typeId || '');
        setAmount(payment.amount || '');
        setCurrency(payment.currency || 'UAH');
        setPaymentKind(payment.kind || 'UPON_COMPLETION');
        setDescription(payment.description || '');

        if (payment.paymentDate) {
          const formattedDate = new Date(payment.paymentDate).toISOString().split('T')[0];
          setPaymentDate(formattedDate);
        }
      })
      .catch((err) => {
        toast.error(`${String(content?.failedToLoadDetails || 'Failed to load payment details')}: ${err.message}`);
      });
  }, [paymentId, projectId, content]);

  const getPaymentKindLabel = (kindKey) => {
    switch (kindKey) {
      case 'DOWN_PAYMENT':
        return String(content?.kindDownPayment || PaymentKind.DOWN_PAYMENT);
      case 'CREDIT_PAYMENT':
        return String(content?.kindCreditPayment || PaymentKind.CREDIT_PAYMENT);
      case 'UPON_COMPLETION':
        return String(content?.kindUponCompletion || PaymentKind.UPON_COMPLETION);
      default:
        return PaymentKind[kindKey] || kindKey;
    }
  };

  const handleUpdate = async (e) => {
    e.preventDefault();
    if (!typeId || !amount || !paymentDate) {
      toast.error(String(content?.validationRequiredFields || 'Type, Amount, and Payment Date are required'));
      return;
    }

    setLoading(true);
    try {
      await paymentRepository.updatePayment(projectId, {
        id: paymentId,
        typeId,
        amount: Number(amount),
        currency,
        paymentDate: new Date(paymentDate),
        description,
        paymentKind,
      });
      toast.success(String(content?.paymentUpdatedSuccess || 'Payment updated successfully'));
      onSuccess();
    } catch (e) {
      toast.error(`${String(content?.failedToUpdate || 'Failed to update payment')}: ${e.message}`);
      console.error('Failed to update payment', e.message);
    } finally {
      setLoading(false);
    }
  };

  const handleCancel = () => onCancel();

  return (
    <Dialog open={Boolean(paymentId)} onOpenChange={(isOpen) => !isOpen && handleCancel()}>
      <DialogContent className="sm:max-w-[480px] rounded-2xl p-6">
        <DialogHeader>
          <DialogTitle className="text-xl font-bold tracking-tight">
            {String(content?.editPaymentTitle || 'Edit Payment')}
          </DialogTitle>
        </DialogHeader>
        <form onSubmit={handleUpdate} className="space-y-4 py-2">
          <div className="space-y-2">
            <Label htmlFor="edit-payment-type" className="text-xs font-semibold">
              {String(content?.typeLabel || 'Type')}
            </Label>
            <Select value={typeId} onValueChange={setTypeId}>
              <SelectTrigger id="edit-payment-type" className="rounded-xl">
                <SelectValue placeholder={String(content?.selectTypePlaceholder || 'Select type')} />
              </SelectTrigger>
              <SelectContent className="rounded-xl">
                {types.map((type) => (
                  <SelectItem key={type.id} value={type.id} className="rounded-lg">
                    {type.name} [{type.category}]
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="edit-payment-kind" className="text-xs font-semibold">
                {String(content?.kindLabel || 'Kind')}
              </Label>
              <Select value={paymentKind} onValueChange={setPaymentKind}>
                <SelectTrigger id="edit-payment-kind" className="rounded-xl">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent className="rounded-xl">
                  {Object.keys(PaymentKind).map((id) => (
                    <SelectItem key={id} value={id} className="rounded-lg">
                      {getPaymentKindLabel(id)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label htmlFor="edit-payment-date" className="text-xs font-semibold">
                {String(content?.dateLabel || 'Date')}
              </Label>
              <Input
                id="edit-payment-date"
                type="date"
                value={paymentDate}
                onChange={(e) => setPaymentDate(e.target.value)}
                className="rounded-xl"
                required
              />
            </div>
          </div>

          <div className="grid grid-cols-3 gap-4">
            <div className="col-span-2 space-y-2">
              <Label htmlFor="edit-payment-amount" className="text-xs font-semibold">
                {String(content?.amountLabel || 'Amount')}
              </Label>
              <Input
                id="edit-payment-amount"
                type="number"
                step="0.01"
                placeholder={String(content?.amountPlaceholder || '0.00')}
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                className="rounded-xl"
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="edit-payment-currency" className="text-xs font-semibold">
                {String(content?.currencyLabel || 'Currency')}
              </Label>
              <Select value={currency} onValueChange={setCurrency}>
                <SelectTrigger id="edit-payment-currency" className="rounded-xl">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent className="rounded-xl">
                  {SUPPORTED_CURRENCIES.map((c) => (
                    <SelectItem key={c} value={c} className="rounded-lg">
                      {c}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="edit-payment-desc" className="text-xs font-semibold">
              {String(content?.descriptionLabel || 'Description')}
            </Label>
            <Textarea
              id="edit-payment-desc"
              rows={3}
              placeholder={String(content?.descriptionPlaceholder || 'Payment description...')}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="rounded-xl resize-none"
            />
          </div>

          <DialogFooter className="pt-3 gap-2 sm:gap-0">
            <Button
              type="button"
              variant="outline"
              onClick={handleCancel}
              disabled={loading}
              className="rounded-xl"
            >
              {String(content?.cancelButton || 'Cancel')}
            </Button>
            <Button
              type="submit"
              disabled={loading}
              className="rounded-xl font-semibold shadow-sm shadow-primary/20 gap-2"
            >
              {loading && <Loader2 className="h-4 w-4 animate-spin" />}
              <span>{String(content?.submitButton || 'Submit')}</span>
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

EditPaymentModal.propTypes = {
  paymentId: PropTypes.string,
  onSuccess: PropTypes.func.isRequired,
  onCancel: PropTypes.func.isRequired,
  types: PropTypes.arrayOf(
    PropTypes.shape({
      id: PropTypes.string.isRequired,
      name: PropTypes.string.isRequired,
      category: PropTypes.string,
    }),
  ),
};
