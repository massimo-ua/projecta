import React, { useState } from 'react';
import { useParams } from 'react-router-dom';
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

const SUPPORTED_CURRENCIES = ['UAH', 'USD', 'EUR', 'PLN'];

export default function AddPaymentModal(props) {
  const { projectId } = useParams();
  const { open, onSuccess, onCancel, types = [] } = props;

  const [typeId, setTypeId] = useState('');
  const [paymentKind, setPaymentKind] = useState('UPON_COMPLETION');
  const [amount, setAmount] = useState('');
  const [currency, setCurrency] = useState('UAH');
  const [paymentDate, setPaymentDate] = useState(new Date().toISOString().split('T')[0]);
  const [description, setDescription] = useState('');
  const [loading, setLoading] = useState(false);

  const handleAdd = async (e) => {
    e.preventDefault();
    if (!typeId || !amount || !paymentDate) {
      toast.error('Type, Amount, and Payment Date are required');
      return;
    }

    setLoading(true);
    try {
      await paymentRepository.addPayment(projectId, {
        typeId,
        amount: Number(amount),
        currency,
        paymentDate: new Date(paymentDate),
        description,
        paymentKind,
      });
      toast.success('Payment added successfully');
      resetForm();
      onSuccess();
    } catch (e) {
      toast.error(`Failed to add payment: ${e.message}`);
      console.error('Failed to add payment', e.message);
    } finally {
      setLoading(false);
    }
  };

  const resetForm = () => {
    setTypeId('');
    setPaymentKind('UPON_COMPLETION');
    setAmount('');
    setCurrency('UAH');
    setPaymentDate(new Date().toISOString().split('T')[0]);
    setDescription('');
  };

  const handleCancel = () => {
    resetForm();
    onCancel();
  };

  return (
    <Dialog open={open} onOpenChange={(isOpen) => !isOpen && handleCancel()}>
      <DialogContent className="sm:max-w-[480px]">
        <DialogHeader>
          <DialogTitle>Add Payment</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleAdd} className="space-y-4 py-2">
          <div className="space-y-2">
            <Label htmlFor="payment-type">Type</Label>
            <Select value={typeId} onValueChange={setTypeId}>
              <SelectTrigger id="payment-type">
                <SelectValue placeholder="Select type" />
              </SelectTrigger>
              <SelectContent>
                {types.map((type) => (
                  <SelectItem key={type.id} value={type.id}>
                    {type.name} [{type.category}]
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="payment-kind">Kind</Label>
              <Select value={paymentKind} onValueChange={setPaymentKind}>
                <SelectTrigger id="payment-kind">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {Object.entries(PaymentKind).map(([id, label]) => (
                    <SelectItem key={id} value={id}>
                      {label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label htmlFor="payment-date">Date</Label>
              <Input
                id="payment-date"
                type="date"
                value={paymentDate}
                onChange={(e) => setPaymentDate(e.target.value)}
                required
              />
            </div>
          </div>

          <div className="grid grid-cols-3 gap-4">
            <div className="col-span-2 space-y-2">
              <Label htmlFor="payment-amount">Amount</Label>
              <Input
                id="payment-amount"
                type="number"
                step="0.01"
                placeholder="0.00"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="payment-currency">Currency</Label>
              <Select value={currency} onValueChange={setCurrency}>
                <SelectTrigger id="payment-currency">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {SUPPORTED_CURRENCIES.map((c) => (
                    <SelectItem key={c} value={c}>
                      {c}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="payment-desc">Description</Label>
            <Textarea
              id="payment-desc"
              rows={3}
              placeholder="Payment description..."
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
          </div>

          <DialogFooter className="pt-2">
            <Button type="button" variant="outline" onClick={handleCancel} disabled={loading}>
              Cancel
            </Button>
            <Button type="submit" disabled={loading}>
              Submit
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
