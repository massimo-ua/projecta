import React, { useState } from 'react';
import { useParams } from 'react-router-dom';
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

export default function AddAssetModal(props) {
  const { projectId } = useParams();
  const { open, onSuccess, onCancel, types = [] } = props;

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
      toast.error('Type, Price, Name and Acquired Date are required');
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
      toast.success('Asset added successfully');
      resetForm();
      onSuccess();
    } catch (e) {
      toast.error(`Failed to add asset: ${e.message}`);
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
      <DialogContent className="sm:max-w-[480px]">
        <DialogHeader>
          <DialogTitle>Add Asset</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleAdd} className="space-y-4 py-2">
          <div className="space-y-2">
            <Label htmlFor="asset-type">Type</Label>
            <Select value={typeId} onValueChange={setTypeId}>
              <SelectTrigger id="asset-type">
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

          <div className="flex items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <Label htmlFor="with-payment" className="text-sm font-medium">Create Payment</Label>
              <p className="text-xs text-muted-foreground">Automatically record an associated payment entry</p>
            </div>
            <Switch
              id="with-payment"
              checked={withPayment}
              onCheckedChange={setWithPayment}
            />
          </div>

          <div className="grid grid-cols-3 gap-4">
            <div className="col-span-2 space-y-2">
              <Label htmlFor="asset-price">Price</Label>
              <Input
                id="asset-price"
                type="number"
                step="0.01"
                placeholder="0.00"
                value={price}
                onChange={(e) => setPrice(e.target.value)}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="asset-currency">Currency</Label>
              <Select value={currency} onValueChange={setCurrency}>
                <SelectTrigger id="asset-currency">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="UAH">UAH</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="asset-acquired">Acquired At</Label>
            <Input
              id="asset-acquired"
              type="date"
              value={acquiredAt}
              onChange={(e) => setAcquiredAt(e.target.value)}
              required
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="asset-name">Name</Label>
            <Input
              id="asset-name"
              placeholder="Asset name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="asset-desc">Description</Label>
            <Textarea
              id="asset-desc"
              rows={3}
              placeholder="Asset description..."
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
