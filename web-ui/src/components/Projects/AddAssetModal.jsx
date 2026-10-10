import React, { useState } from 'react';
import PropTypes from 'prop-types';
import { useParams } from 'react-router-dom';
import { useIntlayer } from 'react-intlayer';
import { assetRepository } from '../../api';
import TagSelector from './TagSelector';
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
}) {
  const content = useIntlayer('assets');
  const { projectId } = useParams();

  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [status, setStatus] = useState('ACTIVE');
  const [startDate, setStartDate] = useState(new Date().toISOString().split('T')[0]);
  const [completedDate, setCompletedDate] = useState('');
  const [targetPrice, setTargetPrice] = useState('');
  const [currency, setCurrency] = useState('UAH');
  const [tags, setTags] = useState([]);
  const [withPayment, setWithPayment] = useState(false);
  const [initialInvestmentPrice, setInitialInvestmentPrice] = useState('');
  const [loading, setLoading] = useState(false);

  const handleAdd = async (e) => {
    e.preventDefault();
    if (!name.trim()) {
      toast.error(String(content?.validationRequiredFields || 'Name is required'));
      return;
    }

    setLoading(true);
    try {
      await assetRepository.addAsset(projectId, {
        name: name.trim(),
        description: description.trim(),
        status,
        startDate: new Date(startDate),
        completedDate: completedDate ? new Date(completedDate) : undefined,
        targetPrice: targetPrice ? Number(targetPrice) : undefined,
        targetCurrency: currency,
        price: withPayment && initialInvestmentPrice ? Number(initialInvestmentPrice) : (targetPrice ? Number(targetPrice) : 0),
        currency,
        withPayment,
        tags,
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
    setName('');
    setDescription('');
    setStatus('ACTIVE');
    setStartDate(new Date().toISOString().split('T')[0]);
    setCompletedDate('');
    setTargetPrice('');
    setCurrency('UAH');
    setTags([]);
    setWithPayment(false);
    setInitialInvestmentPrice('');
  };

  const handleCancel = () => {
    resetForm();
    onCancel();
  };

  return (
    <Dialog open={open} onOpenChange={(isOpen) => !isOpen && handleCancel()}>
      <DialogContent className="sm:max-w-[480px] rounded-2xl p-6 max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="text-xl font-bold tracking-tight">
            {String(content?.addAssetTitle || 'Add Asset')}
          </DialogTitle>
        </DialogHeader>
        <form onSubmit={handleAdd} className="space-y-4 py-2">
          {/* Asset Name */}
          <div className="space-y-2">
            <Label htmlFor="asset-name" className="text-xs font-semibold">
              {String(content?.nameLabel || 'Name')} *
            </Label>
            <Input
              id="asset-name"
              placeholder={String(content?.namePlaceholder || 'e.g. Mercedes Sprinter, Office Relocation')}
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="rounded-xl"
              required
            />
          </div>

          {/* Status & Currency */}
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="asset-status" className="text-xs font-semibold">
                {String(content?.statusLabel || 'Status')}
              </Label>
              <Select value={status} onValueChange={setStatus}>
                <SelectTrigger id="asset-status" className="rounded-xl">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent className="rounded-xl">
                  <SelectItem value="ACTIVE" className="rounded-lg">
                    {String(content?.activeStatus || 'Active')}
                  </SelectItem>
                  <SelectItem value="COMPLETED" className="rounded-lg">
                    {String(content?.completedStatus || 'Completed')}
                  </SelectItem>
                </SelectContent>
              </Select>
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

          {/* Lifecycle Dates */}
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="asset-start-date" className="text-xs font-semibold">
                {String(content?.startDateLabel || 'Start Date')} *
              </Label>
              <Input
                id="asset-start-date"
                type="date"
                value={startDate}
                onChange={(e) => setStartDate(e.target.value)}
                className="rounded-xl"
                required
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="asset-completed-date" className="text-xs font-semibold">
                {String(content?.completedDateLabel || 'Completed Date')}
              </Label>
              <Input
                id="asset-completed-date"
                type="date"
                value={completedDate}
                onChange={(e) => setCompletedDate(e.target.value)}
                className="rounded-xl"
              />
            </div>
          </div>

          {/* Target Valuation / Budget */}
          <div className="space-y-2">
            <Label htmlFor="asset-target-price" className="text-xs font-semibold">
              {String(content?.targetPriceLabel || 'Target Budget / Target Price (Optional)')}
            </Label>
            <Input
              id="asset-target-price"
              type="number"
              step="0.01"
              placeholder={String(content?.pricePlaceholder || '0.00')}
              value={targetPrice}
              onChange={(e) => setTargetPrice(e.target.value)}
              className="rounded-xl"
            />
          </div>

          {/* Description */}
          <div className="space-y-2">
            <Label htmlFor="asset-desc" className="text-xs font-semibold">
              {String(content?.descriptionLabel || 'Description')}
            </Label>
            <Textarea
              id="asset-desc"
              rows={2}
              placeholder={String(content?.descriptionPlaceholder || 'Asset description, purpose, scope...')}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="rounded-xl resize-none"
            />
          </div>

          {/* Tags */}
          <TagSelector
            id="asset-tags"
            tags={tags}
            onChange={setTags}
            projectId={projectId}
            label={String(content?.tagsLabel || 'Tags')}
            placeholder={String(content?.tagsPlaceholder || 'Type tag and press Enter')}
          />

          {/* Initial Investment Switch */}
          <div className="space-y-3 rounded-xl border border-border/70 p-3 bg-muted/20">
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label htmlFor="with-payment" className="text-xs font-semibold">
                  {String(content?.createPaymentLabel || 'Log Initial Investment')}
                </Label>
                <p className="text-xs text-muted-foreground">
                  {String(content?.createPaymentDescription || 'Automatically record an initial capital investment')}
                </p>
              </div>
              <Switch
                id="with-payment"
                checked={withPayment}
                onCheckedChange={setWithPayment}
              />
            </div>
            {withPayment && (
              <div className="pt-2">
                <Label htmlFor="initial-investment-amount" className="text-xs font-semibold">
                  {String(content?.priceLabel || 'Initial Investment Amount')}
                </Label>
                <Input
                  id="initial-investment-amount"
                  type="number"
                  step="0.01"
                  placeholder="0.00"
                  value={initialInvestmentPrice}
                  onChange={(e) => setInitialInvestmentPrice(e.target.value)}
                  className="rounded-xl mt-1.5"
                  required={withPayment}
                />
              </div>
            )}
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
              <span>{String(content?.submitButton || 'Save')}</span>
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
};
