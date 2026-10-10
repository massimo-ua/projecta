import React, { useState } from 'react';
import PropTypes from 'prop-types';
import { useIntlayer } from 'react-intlayer';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
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
import { Loader2, Link2 } from 'lucide-react';

export default function LinkChildAssetModal({
  open,
  onClose,
  onLink,
  parentAsset,
  allAssets = [],
}) {
  const content = useIntlayer('assets');
  const [selectedChildId, setSelectedChildId] = useState('');
  const [sharePercentage, setSharePercentage] = useState('100');
  const [loading, setLoading] = useState(false);

  const existingChildIds = new Set(
    (parentAsset?.children || []).map((c) => c.childId || c.child_id),
  );

  // Eligible child assets: not self, and not already a child
  const eligibleCandidates = allAssets.filter(
    (a) => a.id !== parentAsset?.id && !existingChildIds.has(a.id),
  );

  const handleLink = async (e) => {
    e.preventDefault();
    if (!selectedChildId) {
      toast.error('Please select an asset to link as component');
      return;
    }
    const shareNum = Number(sharePercentage);
    if (isNaN(shareNum) || shareNum <= 0 || shareNum > 100) {
      toast.error('Share percentage must be between 0.1% and 100%');
      return;
    }

    setLoading(true);
    try {
      await onLink(parentAsset.id, selectedChildId, shareNum);
      setSelectedChildId('');
      setSharePercentage('100');
    } catch {
      // toast is already handled in hook
    } finally {
      setLoading(false);
    }
  };

  const handleCancel = () => {
    setSelectedChildId('');
    setSharePercentage('100');
    onClose();
  };

  return (
    <Dialog open={open} onOpenChange={(isOpen) => !isOpen && handleCancel()}>
      <DialogContent className="sm:max-w-[460px] rounded-2xl p-6">
        <DialogHeader>
          <div className="flex items-center gap-2">
            <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-primary/10 text-primary">
              <Link2 className="h-5 w-5" />
            </div>
            <div>
              <DialogTitle className="text-lg font-bold tracking-tight">
                {String(content?.linkModalTitle || 'Link Component Asset')}
              </DialogTitle>
              <p className="text-xs text-muted-foreground mt-0.5">
                Parent: <span className="font-semibold text-foreground">{parentAsset?.name}</span>
              </p>
            </div>
          </div>
        </DialogHeader>

        <form onSubmit={handleLink} className="space-y-4 py-2">
          {/* Child Asset Selector */}
          <div className="space-y-2">
            <Label htmlFor="candidate-child" className="text-xs font-semibold">
              {String(content?.selectChildAsset || 'Select Component / Sub-Asset')} *
            </Label>
            {eligibleCandidates.length === 0 ? (
              <p className="text-xs text-muted-foreground p-3 rounded-xl border border-dashed border-border/70 bg-muted/20">
                No eligible candidate assets available to link.
              </p>
            ) : (
              <Select value={selectedChildId} onValueChange={setSelectedChildId}>
                <SelectTrigger id="candidate-child" className="rounded-xl">
                  <SelectValue placeholder="Choose asset..." />
                </SelectTrigger>
                <SelectContent className="rounded-xl max-h-56">
                  {eligibleCandidates.map((asset) => (
                    <SelectItem key={asset.id} value={asset.id} className="rounded-lg">
                      <div className="flex items-center justify-between gap-3 w-full">
                        <span className="font-medium">{asset.name}</span>
                        <span className="text-xs text-muted-foreground font-mono">
                          {asset.formattedTotalCost || asset.formattedPrice}
                        </span>
                      </div>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
          </div>

          {/* Allocation Share Percentage */}
          <div className="space-y-2">
            <div className="flex justify-between items-center">
              <Label htmlFor="share-pct" className="text-xs font-semibold">
                {String(content?.sharePercentageLabel || 'Allocation Share (%)')} *
              </Label>
              <span className="text-[11px] text-muted-foreground">Default: 100%</span>
            </div>
            <Input
              id="share-pct"
              type="number"
              min="0.1"
              max="100"
              step="0.1"
              value={sharePercentage}
              onChange={(e) => setSharePercentage(e.target.value)}
              className="rounded-xl"
              required
            />
            <p className="text-[11px] text-muted-foreground">
              What percentage of the child asset's total cost is attributed to &ldquo;{parentAsset?.name}&rdquo;?
            </p>
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
              disabled={loading || !selectedChildId || eligibleCandidates.length === 0}
              className="rounded-xl font-semibold shadow-sm shadow-primary/20 gap-2"
            >
              {loading && <Loader2 className="h-4 w-4 animate-spin" />}
              <span>{String(content?.submitButton || 'Link Asset')}</span>
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

LinkChildAssetModal.propTypes = {
  open: PropTypes.bool.isRequired,
  onClose: PropTypes.func.isRequired,
  onLink: PropTypes.func.isRequired,
  parentAsset: PropTypes.shape({
    id: PropTypes.string,
    name: PropTypes.string,
    children: PropTypes.array,
  }),
  allAssets: PropTypes.array,
};
