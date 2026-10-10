import React, { useEffect, useState } from 'react';
import PropTypes from 'prop-types';
import { useParams } from 'react-router-dom';
import { useIntlayer } from 'react-intlayer';
import { assetRepository } from '../../api';
import TagSelector from './TagSelector';
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

export default function EditAssetModal({
  onSuccess,
  onCancel,
  assetId = null,
}) {
  const content = useIntlayer('assets');
  const { projectId } = useParams();

  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [status, setStatus] = useState('ACTIVE');
  const [startDate, setStartDate] = useState('');
  const [completedDate, setCompletedDate] = useState('');
  const [targetPrice, setTargetPrice] = useState('');
  const [currency, setCurrency] = useState('UAH');
  const [tags, setTags] = useState([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!assetId) return;

    assetRepository
      .getAsset(projectId, assetId)
      .then((asset) => {
        setName(asset.name || '');
        setDescription(asset.description || '');
        setStatus(asset.status || 'ACTIVE');
        setTargetPrice(asset.targetPrice || '');
        setCurrency(asset.targetCurrency || asset.currency || 'UAH');
        setTags(Array.isArray(asset.tags) ? asset.tags : []);

        if (asset.startDate || asset.acquiredAt) {
          const dt = new Date(asset.startDate || asset.acquiredAt);
          if (!isNaN(dt.getTime())) {
            setStartDate(dt.toISOString().split('T')[0]);
          }
        }
        if (asset.completedDate) {
          const dt = new Date(asset.completedDate);
          if (!isNaN(dt.getTime())) {
            setCompletedDate(dt.toISOString().split('T')[0]);
          }
        } else {
          setCompletedDate('');
        }
      })
      .catch((err) => {
        toast.error(`${String(content?.failedToLoadDetails || 'Failed to load asset details')}: ${err.message}`);
      });
  }, [assetId, projectId, content]);

  const handleUpdate = async (e) => {
    e.preventDefault();
    if (!name.trim()) {
      toast.error(String(content?.validationRequiredFields || 'Name is required'));
      return;
    }

    setLoading(true);
    try {
      await assetRepository.updateAsset(projectId, {
        id: assetId,
        name: name.trim(),
        description: description.trim(),
        status,
        startDate: startDate ? new Date(startDate) : new Date(),
        completedDate: completedDate ? new Date(completedDate) : undefined,
        targetPrice: targetPrice ? Number(targetPrice) : undefined,
        targetCurrency: currency,
        currency,
        tags,
      });
      toast.success(String(content?.assetUpdatedSuccess || 'Asset updated successfully'));
      onSuccess();
    } catch (e) {
      toast.error(`${String(content?.failedToUpdate || 'Failed to update asset')}: ${e.message}`);
      console.error('Failed to update asset', e.message);
    } finally {
      setLoading(false);
    }
  };

  const handleCancel = () => onCancel();

  return (
    <Dialog open={Boolean(assetId)} onOpenChange={(isOpen) => !isOpen && handleCancel()}>
      <DialogContent className="sm:max-w-[480px] rounded-2xl p-6 max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="text-xl font-bold tracking-tight">
            {String(content?.editAssetTitle || 'Edit Asset')}
          </DialogTitle>
        </DialogHeader>
        <form onSubmit={handleUpdate} className="space-y-4 py-2">
          {/* Asset Name */}
          <div className="space-y-2">
            <Label htmlFor="edit-asset-name" className="text-xs font-semibold">
              {String(content?.nameLabel || 'Name')} *
            </Label>
            <Input
              id="edit-asset-name"
              placeholder={String(content?.namePlaceholder || 'Asset name')}
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="rounded-xl"
              required
            />
          </div>

          {/* Status & Currency */}
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="edit-asset-status" className="text-xs font-semibold">
                {String(content?.statusLabel || 'Status')}
              </Label>
              <Select value={status} onValueChange={setStatus}>
                <SelectTrigger id="edit-asset-status" className="rounded-xl">
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
              <Label htmlFor="edit-asset-currency" className="text-xs font-semibold">
                {String(content?.currencyLabel || 'Currency')}
              </Label>
              <Select value={currency} onValueChange={setCurrency}>
                <SelectTrigger id="edit-asset-currency" className="rounded-xl">
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
              <Label htmlFor="edit-asset-start-date" className="text-xs font-semibold">
                {String(content?.startDateLabel || 'Start Date')} *
              </Label>
              <Input
                id="edit-asset-start-date"
                type="date"
                value={startDate}
                onChange={(e) => setStartDate(e.target.value)}
                className="rounded-xl"
                required
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="edit-asset-completed-date" className="text-xs font-semibold">
                {String(content?.completedDateLabel || 'Completed Date')}
              </Label>
              <Input
                id="edit-asset-completed-date"
                type="date"
                value={completedDate}
                onChange={(e) => setCompletedDate(e.target.value)}
                className="rounded-xl"
              />
            </div>
          </div>

          {/* Target Valuation / Budget */}
          <div className="space-y-2">
            <Label htmlFor="edit-asset-target-price" className="text-xs font-semibold">
              {String(content?.targetPriceLabel || 'Target Budget / Target Price (Optional)')}
            </Label>
            <Input
              id="edit-asset-target-price"
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
            <Label htmlFor="edit-asset-desc" className="text-xs font-semibold">
              {String(content?.descriptionLabel || 'Description')}
            </Label>
            <Textarea
              id="edit-asset-desc"
              rows={3}
              placeholder={String(content?.descriptionPlaceholder || 'Asset description...')}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="rounded-xl resize-none"
            />
          </div>

          {/* Tags */}
          <TagSelector
            id="edit-asset-tags"
            tags={tags}
            onChange={setTags}
            projectId={projectId}
            label={String(content?.tagsLabel || 'Tags')}
            placeholder={String(content?.tagsPlaceholder || 'Type tag and press Enter')}
          />

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

EditAssetModal.propTypes = {
  assetId: PropTypes.string,
  onSuccess: PropTypes.func.isRequired,
  onCancel: PropTypes.func.isRequired,
};
