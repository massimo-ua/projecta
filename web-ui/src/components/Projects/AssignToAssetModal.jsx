import React, { useState } from 'react';
import PropTypes from 'prop-types';
import { useIntlayer } from 'react-intlayer';
import { toast } from 'sonner';
import { Loader2, Package, FolderPlus } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
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
  DialogDescription,
} from '@/components/ui/dialog';
import { assetRepository } from '../../api';

export default function AssignToAssetModal({
  open,
  onSuccess,
  onCancel,
  projectId,
  investments = [],
  availableAssets = [],
}) {
  const content = useIntlayer('payments');
  const [targetAssetId, setTargetAssetId] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!targetAssetId) {
      toast.error(String(content.selectAssetPlaceholder || 'Please select an asset'));
      return;
    }
    if (!investments || investments.length === 0) {
      return;
    }

    setLoading(true);
    try {
      const investmentIds = investments.map((inv) => inv.id);
      await assetRepository.assignInvestments(projectId, targetAssetId, investmentIds);
      const chosenAsset = availableAssets.find((a) => a.id === targetAssetId);
      toast.success(
        chosenAsset
          ? `${String(content.assignSuccess || 'Investments assigned to asset successfully')}: ${chosenAsset.name}`
          : String(content.assignSuccess || 'Investments assigned to asset successfully'),
      );
      setTargetAssetId('');
      onSuccess();
    } catch (err) {
      toast.error(`${String(content.failedToAssign || 'Failed to assign investments to asset')}: ${err.message}`);
      console.error('Failed to assign investments to asset:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleOpenChange = (isOpen) => {
    if (!isOpen) {
      setTargetAssetId('');
      onCancel();
    }
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-[480px] rounded-2xl p-6">
        <DialogHeader>
          <div className="flex items-center gap-2">
            <div className="p-2 rounded-xl bg-primary/10 text-primary">
              <FolderPlus className="h-5 w-5" />
            </div>
            <div>
              <DialogTitle className="text-lg font-bold tracking-tight">
                {String(content.assignToAssetTitle || 'Add Investments to Existing Asset')}
              </DialogTitle>
              <DialogDescription className="text-xs text-muted-foreground mt-0.5">
                {String(content.assignToAssetDesc || 'Assign the selected investments to an existing project asset.')}
              </DialogDescription>
            </div>
          </div>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4 pt-2">
          {/* Target Asset Selection */}
          <div className="space-y-1.5">
            <Label className="text-xs font-semibold">
              {String(content.assetTargetLabel || 'Target Asset')} *
            </Label>
            <Select value={targetAssetId} onValueChange={setTargetAssetId}>
              <SelectTrigger className="rounded-xl h-10 text-sm">
                <SelectValue placeholder={String(content.selectAssetPlaceholder || 'Select target asset...')} />
              </SelectTrigger>
              <SelectContent className="rounded-xl max-h-56">
                {availableAssets.map((asset) => (
                  <SelectItem key={asset.id} value={asset.id} className="rounded-lg text-sm">
                    {asset.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* Selected Investments Summary */}
          <div className="space-y-2 rounded-xl bg-muted/40 border border-border/50 p-3.5">
            <div className="flex items-center justify-between">
              <span className="text-xs font-semibold text-foreground">
                {String(content.selectedPayments || 'Selected investments')}
              </span>
              <Badge variant="outline" className="font-mono text-xs">
                {investments.length}
              </Badge>
            </div>

            <div className="max-h-36 overflow-y-auto space-y-1.5 pr-1">
              {investments.map((inv) => (
                <div
                  key={inv.id}
                  className="flex items-center justify-between text-xs py-1 border-b border-border/20 last:border-b-0"
                >
                  <span className="truncate max-w-[240px] text-muted-foreground">
                    {inv.description || inv.date || inv.id}
                  </span>
                  <span className="font-mono font-medium text-foreground ml-2 shrink-0">
                    {inv.formattedAmount || `${(inv.rawAmount / 100).toFixed(2)} ${inv.currency}`}
                  </span>
                </div>
              ))}
            </div>
          </div>

          <DialogFooter className="gap-2 sm:gap-0 pt-2">
            <Button
              type="button"
              variant="outline"
              onClick={() => handleOpenChange(false)}
              disabled={loading}
              className="rounded-xl h-9 text-xs"
            >
              {String(content.cancelButton || 'Cancel')}
            </Button>
            <Button
              type="submit"
              disabled={loading || !targetAssetId || investments.length === 0}
              className="gap-2 rounded-xl h-9 text-xs font-semibold shadow-sm shadow-primary/20"
            >
              {loading ? (
                <>
                  <Loader2 className="h-3.5 w-3.5 animate-spin" />
                  <span>Saving...</span>
                </>
              ) : (
                <>
                  <Package className="h-3.5 w-3.5" />
                  <span>{String(content.assignButton || 'Assign Investments')}</span>
                </>
              )}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

AssignToAssetModal.propTypes = {
  open: PropTypes.bool.isRequired,
  onSuccess: PropTypes.func.isRequired,
  onCancel: PropTypes.func.isRequired,
  projectId: PropTypes.string.isRequired,
  investments: PropTypes.arrayOf(PropTypes.object).isRequired,
  availableAssets: PropTypes.arrayOf(PropTypes.object).isRequired,
};

