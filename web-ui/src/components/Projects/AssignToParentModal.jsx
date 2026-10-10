import React, { useState } from 'react';
import PropTypes from 'prop-types';
import { useIntlayer } from 'react-intlayer';
import { assetRepository } from '../../api';
import { Button } from '@/components/ui/button';
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
import { Loader2, FolderPlus, Package } from 'lucide-react';

export default function AssignToParentModal({
  open,
  onCancel,
  onSuccess,
  selectedAssets = [],
  allAssets = [],
  projectId,
}) {
  const content = useIntlayer('assets');
  const [selectedParentId, setSelectedParentId] = useState('');
  const [loading, setLoading] = useState(false);

  const selectedIds = new Set(selectedAssets.map((a) => a.id));
  const candidateParents = allAssets.filter((a) => !selectedIds.has(a.id));

  const handleAssign = async (e) => {
    e.preventDefault();
    if (!selectedParentId) {
      toast.error(String(content?.selectParentAsset || 'Please select a parent asset'));
      return;
    }

    setLoading(true);
    try {
      await assetRepository.linkChildren(
        projectId,
        selectedParentId,
        selectedAssets.map((a) => a.id),
      );
      setSelectedParentId('');
      toast.success(String(content?.assetsGroupedSuccess || 'Assets added to parent successfully'));
      onSuccess();
    } catch (err) {
      toast.error(err?.message || String(content?.failedToGroup || 'Failed to link assets to parent'));
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    setSelectedParentId('');
    onCancel();
  };

  return (
    <Dialog open={open} onOpenChange={(isOpen) => !isOpen && handleClose()}>
      <DialogContent className="sm:max-w-[480px] rounded-2xl p-6">
        <DialogHeader>
          <div className="flex items-center gap-2">
            <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-primary/10 text-primary">
              <FolderPlus className="h-5 w-5" />
            </div>
            <div>
              <DialogTitle className="text-lg font-bold tracking-tight">
                {String(content?.assignModalTitle || 'Add to Existing Parent Asset')}
              </DialogTitle>
              <p className="text-xs text-muted-foreground mt-0.5">
                {String(content?.assignModalDesc || 'The selected assets will be linked as components under the chosen parent asset.')}
              </p>
            </div>
          </div>
        </DialogHeader>

        <form onSubmit={handleAssign} className="space-y-4 py-2">
          {/* Selected Assets to Link */}
          <div className="rounded-xl border border-border/70 bg-muted/40 p-3 space-y-1.5">
            <div className="text-xs font-semibold text-muted-foreground">
              {String(content?.componentsSummary || 'Assets to Link')} ({selectedAssets.length})
            </div>
            <div className="max-h-32 overflow-y-auto space-y-1 pr-1">
              {selectedAssets.map((a) => (
                <div
                  key={a.id}
                  className="flex items-center justify-between p-1.5 rounded-lg bg-background/80 border border-border/50 text-xs"
                >
                  <div className="flex items-center gap-1.5 truncate">
                    <Package className="h-3.5 w-3.5 text-muted-foreground shrink-0" />
                    <span className="font-medium truncate">{a.name}</span>
                  </div>
                  <span className="font-mono text-muted-foreground shrink-0">
                    {a.formattedTotalCost || a.formattedPrice}
                  </span>
                </div>
              ))}
            </div>
          </div>

          {/* Parent Asset Selector */}
          <div className="space-y-1.5">
            <Label className="text-xs font-semibold">
              {String(content?.selectParentAsset || 'Select Parent Asset')} *
            </Label>
            {candidateParents.length === 0 ? (
              <p className="text-xs text-muted-foreground italic">
                No eligible parent assets available.
              </p>
            ) : (
              <Select value={selectedParentId} onValueChange={setSelectedParentId}>
                <SelectTrigger className="rounded-xl">
                  <SelectValue placeholder={String(content?.selectParentAsset || 'Choose parent asset...')} />
                </SelectTrigger>
                <SelectContent className="rounded-xl max-h-56">
                  {candidateParents.map((parent) => (
                    <SelectItem key={parent.id} value={parent.id} className="rounded-lg text-xs">
                      {parent.name} ({parent.formattedTotalCost || parent.formattedPrice})
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
          </div>

          <DialogFooter className="pt-2">
            <Button
              type="button"
              variant="outline"
              onClick={handleClose}
              disabled={loading}
              className="rounded-xl"
            >
              {String(content?.cancel || 'Cancel')}
            </Button>
            <Button
              type="submit"
              disabled={loading || !selectedParentId || candidateParents.length === 0}
              className="gap-2 rounded-xl font-semibold shadow-sm shadow-primary/20"
            >
              {loading && <Loader2 className="h-4 w-4 animate-spin" />}
              <span>{String(content?.addToParentAsset || 'Add to Parent Asset')}</span>
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

AssignToParentModal.propTypes = {
  open: PropTypes.bool.isRequired,
  onCancel: PropTypes.func.isRequired,
  onSuccess: PropTypes.func.isRequired,
  selectedAssets: PropTypes.arrayOf(PropTypes.object).isRequired,
  allAssets: PropTypes.arrayOf(PropTypes.object).isRequired,
  projectId: PropTypes.string.isRequired,
};
