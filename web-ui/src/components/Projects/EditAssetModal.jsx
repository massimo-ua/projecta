import React, { useEffect, useState } from 'react';
import PropTypes from 'prop-types';
import { useParams } from 'react-router-dom';
import { useIntlayer } from 'react-intlayer';
import { assetRepository } from '../../api';
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

export default function EditAssetModal({
  onSuccess,
  onCancel,
  assetId = null,
  types = [],
}) {
  const content = useIntlayer('assets');
  const { projectId } = useParams();

  const [typeId, setTypeId] = useState('');
  const [price, setPrice] = useState('');
  const [currency, setCurrency] = useState('UAH');
  const [acquiredAt, setAcquiredAt] = useState('');
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!assetId) return;

    assetRepository
      .getAsset(projectId, assetId)
      .then((asset) => {
        setTypeId(asset.typeId || '');
        setPrice(asset.price || '');
        setCurrency(asset.currency || 'UAH');
        setName(asset.name || '');
        setDescription(asset.description || '');

        if (asset.acquiredAt) {
          const formattedDate = new Date(asset.acquiredAt).toISOString().split('T')[0];
          setAcquiredAt(formattedDate);
        }
      })
      .catch((err) => {
        toast.error(`${String(content?.failedToLoadDetails || 'Failed to load asset details')}: ${err.message}`);
      });
  }, [assetId, projectId, content]);

  const handleUpdate = async (e) => {
    e.preventDefault();
    if (!typeId || !price || !name || !acquiredAt) {
      toast.error(String(content?.validationRequiredFields || 'Type, Price, Name and Acquired Date are required'));
      return;
    }

    setLoading(true);
    try {
      await assetRepository.updateAsset(projectId, {
        id: assetId,
        typeId,
        price: Number(price),
        currency,
        acquiredAt: new Date(acquiredAt),
        name,
        description,
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
      <DialogContent className="sm:max-w-[480px]">
        <DialogHeader>
          <DialogTitle>{String(content?.editAssetTitle || 'Edit Asset')}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleUpdate} className="space-y-4 py-2">
          <div className="space-y-2">
            <Label htmlFor="edit-asset-type">{String(content?.typeLabel || 'Type')}</Label>
            <Select value={typeId} onValueChange={setTypeId}>
              <SelectTrigger id="edit-asset-type">
                <SelectValue placeholder={String(content?.selectTypePlaceholder || 'Select type')} />
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

          <div className="grid grid-cols-3 gap-4">
            <div className="col-span-2 space-y-2">
              <Label htmlFor="edit-asset-price">{String(content?.priceLabel || 'Price')}</Label>
              <Input
                id="edit-asset-price"
                type="number"
                step="0.01"
                placeholder={String(content?.pricePlaceholder || '0.00')}
                value={price}
                onChange={(e) => setPrice(e.target.value)}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="edit-asset-currency">{String(content?.currencyLabel || 'Currency')}</Label>
              <Select value={currency} onValueChange={setCurrency}>
                <SelectTrigger id="edit-asset-currency">
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
            <Label htmlFor="edit-asset-acquired">{String(content?.acquiredAtLabel || 'Acquired At')}</Label>
            <Input
              id="edit-asset-acquired"
              type="date"
              value={acquiredAt}
              onChange={(e) => setAcquiredAt(e.target.value)}
              required
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="edit-asset-name">{String(content?.nameLabel || 'Name')}</Label>
            <Input
              id="edit-asset-name"
              placeholder={String(content?.namePlaceholder || 'Asset name')}
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="edit-asset-desc">{String(content?.descriptionLabel || 'Description')}</Label>
            <Textarea
              id="edit-asset-desc"
              rows={3}
              placeholder={String(content?.descriptionPlaceholder || 'Asset description...')}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
          </div>

          <DialogFooter className="pt-2">
            <Button type="button" variant="outline" onClick={handleCancel} disabled={loading}>
              {String(content?.cancelButton || 'Cancel')}
            </Button>
            <Button type="submit" disabled={loading}>
              {String(content?.submitButton || 'Submit')}
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
  types: PropTypes.arrayOf(
    PropTypes.shape({
      id: PropTypes.string.isRequired,
      name: PropTypes.string.isRequired,
      category: PropTypes.string,
    }),
  ),
};
