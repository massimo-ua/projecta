import React, { useState } from 'react';
import PropTypes from 'prop-types';
import { useParams } from 'react-router-dom';
import { useIntlayer } from 'react-intlayer';
import { assetRepository } from '../../api';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
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

export default function AddAssetModal({
  open,
  onSuccess,
  onCancel,
  types = [],
}) {
  const content = useIntlayer('assets');
  const { projectId } = useParams();

  const [typeId, setTypeId] = useState('');
  const [withPayment, setWithPayment] = useState(false);
  const [price, setPrice] = useState('');
  const [currency, setCurrency] = useState('UAH');
  const [acquiredAt, setAcquiredAt] = useState(new Date().toISOString().split('T')[0]);
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [loading, setLoading] = useState(false);

  const handleAdd = async (e) => {
    e.preventDefault();
    if (!typeId || !price || !name || !acquiredAt) {
      toast.error(String(content?.validationRequiredFields || 'Type, Price, Name and Acquired Date are required'));
      return;
    }

    setLoading(true);
    try {
      await assetRepository.addAsset(projectId, {
        typeId,
        price: Number(price),
        currency,
        acquiredAt: new Date(acquiredAt),
        name,
        description,
        withPayment,
      });
      toast.success(String(content?.assetAddedSuccess || 'Asset added successfully'));
      resetForm();
      onSuccess();
    } catch (e) {
      toast.error(`${String(content?.failedToAdd || 'Failed to add asset')}: ${e.message}`);
      console.error('Failed to add asset', e.message);
    } finally {
      setLoading(false);
    }
  };

  const resetForm = () => {
    setTypeId('');
    setWithPayment(false);
    setPrice('');
    setCurrency('UAH');
    setAcquiredAt(new Date().toISOString().split('T')[0]);
    setName('');
    setDescription('');
  };

  const handleCancel = () => {
    resetForm();
    onCancel();
  };

  return (
    <Dialog open={open} onOpenChange={(isOpen) => !isOpen && handleCancel()}>
      <DialogContent className="sm:max-w-[480px] rounded-2xl p-6">
        <DialogHeader>
          <DialogTitle className="text-xl font-bold tracking-tight">
            {String(content?.addAssetTitle || 'Add Asset')}
          </DialogTitle>
        </DialogHeader>
        <form onSubmit={handleAdd} className="space-y-4 py-2">
          <div className="space-y-2">
            <Label htmlFor="asset-type" className="text-xs font-semibold">
              {String(content?.typeLabel || 'Type')}
            </Label>
            <Select value={typeId} onValueChange={setTypeId}>
              <SelectTrigger id="asset-type" className="rounded-xl">
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

          <div className="flex items-center justify-between rounded-xl border border-border/70 p-3 bg-muted/20">
            <div className="space-y-0.5">
              <Label htmlFor="with-payment" className="text-xs font-semibold">
                {String(content?.createPaymentLabel || 'Create Payment')}
              </Label>
              <p className="text-xs text-muted-foreground">
                {String(content?.createPaymentDescription || 'Automatically record an associated payment entry')}
              </p>
            </div>
            <Switch
              id="with-payment"
              checked={withPayment}
              onCheckedChange={setWithPayment}
            />
          </div>

          <div className="grid grid-cols-3 gap-4">
            <div className="col-span-2 space-y-2">
              <Label htmlFor="asset-price" className="text-xs font-semibold">
                {String(content?.priceLabel || 'Price')}
              </Label>
              <Input
                id="asset-price"
                type="number"
                step="0.01"
                placeholder={String(content?.pricePlaceholder || '0.00')}
                value={price}
                onChange={(e) => setPrice(e.target.value)}
                className="rounded-xl"
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="asset-currency" className="text-xs font-semibold">
                {String(content?.currencyLabel || 'Currency')}
              </Label>
              <Select value={currency} onValueChange={setCurrency}>
                <SelectTrigger id="asset-currency" className="rounded-xl">
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
            <Label htmlFor="asset-acquired" className="text-xs font-semibold">
              {String(content?.acquiredAtLabel || 'Acquired At')}
            </Label>
            <Input
              id="asset-acquired"
              type="date"
              value={acquiredAt}
              onChange={(e) => setAcquiredAt(e.target.value)}
              className="rounded-xl"
              required
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="asset-name" className="text-xs font-semibold">
              {String(content?.nameLabel || 'Name')}
            </Label>
            <Input
              id="asset-name"
              placeholder={String(content?.namePlaceholder || 'Asset name')}
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="rounded-xl"
              required
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="asset-desc" className="text-xs font-semibold">
              {String(content?.descriptionLabel || 'Description')}
            </Label>
            <Textarea
              id="asset-desc"
              rows={3}
              placeholder={String(content?.descriptionPlaceholder || 'Asset description...')}
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

AddAssetModal.propTypes = {
  open: PropTypes.bool.isRequired,
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
