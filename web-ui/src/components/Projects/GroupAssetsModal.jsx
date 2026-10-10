import React, { useState } from 'react';
import PropTypes from 'prop-types';
import { useIntlayer } from 'react-intlayer';
import { assetRepository } from '../../api';
import TagSelector from './TagSelector';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog';
import { toast } from 'sonner';
import { Loader2, Boxes, Package } from 'lucide-react';

export default function GroupAssetsModal({
  open,
  onCancel,
  onSuccess,
  selectedAssets = [],
  projectId,
}) {
  const content = useIntlayer('assets');
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [tags, setTags] = useState([]);
  const [loading, setLoading] = useState(false);

  // Group assets costs summary
  const totalsByCurrency = selectedAssets.reduce((acc, a) => {
    const curr = a.currency || 'UAH';
    const amount = Number(a.rawTotalCost || a.rawPrice || 0);
    acc[curr] = (acc[curr] || 0) + amount;
    return acc;
  }, {});

  const handleGroup = async (e) => {
    e.preventDefault();
    if (!name.trim()) {
      toast.error(String(content?.validationRequiredFields || 'Name is required'));
      return;
    }

    setLoading(true);
    try {
      await assetRepository.groupAssets(projectId, {
        name: name.trim(),
        description: description.trim(),
        childAssetIds: selectedAssets.map((a) => a.id),
        tags,
      });
      setName('');
      setDescription('');
      setTags([]);
      onSuccess();
    } catch (err) {
      toast.error(err?.message || String(content?.failedToGroup || 'Failed to group assets'));
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    setName('');
    setDescription('');
    setTags([]);
    onCancel();
  };

  return (
    <Dialog open={open} onOpenChange={(isOpen) => !isOpen && handleClose()}>
      <DialogContent className="sm:max-w-[560px] max-h-[90vh] overflow-y-auto rounded-2xl p-6">
        <DialogHeader>
          <div className="flex items-center gap-2">
            <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-primary/10 text-primary">
              <Boxes className="h-5 w-5" />
            </div>
            <div>
              <DialogTitle className="text-lg font-bold tracking-tight">
                {String(content?.groupModalTitle || 'Group Assets into New Asset')}
              </DialogTitle>
              <p className="text-xs text-muted-foreground mt-0.5">
                {String(content?.groupModalDesc || 'The new asset will dynamically aggregate all costs from selected component assets.')}
              </p>
            </div>
          </div>
        </DialogHeader>

        <form onSubmit={handleGroup} className="space-y-4 py-2">
          {/* Selected Component Assets Preview */}
          <div className="rounded-xl border border-border/70 bg-muted/40 p-3 space-y-2">
            <div className="flex items-center justify-between text-xs font-semibold text-muted-foreground">
              <span>{String(content?.componentsSummary || 'Selected Component Assets')} ({selectedAssets.length})</span>
              <div className="flex gap-2">
                {Object.entries(totalsByCurrency).map(([curr, amt]) => (
                  <Badge key={curr} variant="default" className="text-[11px] font-mono font-bold">
                    ~ {amt} {curr}
                  </Badge>
                ))}
              </div>
            </div>
            <div className="max-h-36 overflow-y-auto space-y-1 pr-1">
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

          {/* Group Asset Name */}
          <div className="space-y-1.5">
            <Label htmlFor="group-name" className="text-xs font-semibold">
              {String(content?.nameLabel || 'Group Asset Name')} *
            </Label>
            <Input
              id="group-name"
              placeholder={String(content?.namePlaceholder || 'e.g. Rent, Office Equipment...')}
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="rounded-xl"
              required
              autoFocus
            />
          </div>

          {/* Group Asset Description */}
          <div className="space-y-1.5">
            <Label htmlFor="group-description" className="text-xs font-semibold">
              {String(content?.descriptionLabel || 'Description')}
            </Label>
            <Textarea
              id="group-description"
              placeholder={String(content?.descriptionPlaceholder || 'Optional notes...')}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="rounded-xl min-h-[70px] resize-none"
            />
          </div>

          {/* Tags */}
          <div className="space-y-1.5">
            <Label className="text-xs font-semibold">
              {String(content?.tagsLabel || 'Tags')}
            </Label>
            <TagSelector
              projectId={projectId}
              selectedTags={tags}
              onChange={setTags}
            />
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
              disabled={loading}
              className="gap-2 rounded-xl font-semibold shadow-sm shadow-primary/20"
            >
              {loading && <Loader2 className="h-4 w-4 animate-spin" />}
              <span>{String(content?.groupAssets || 'Group into New Asset')}</span>
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

GroupAssetsModal.propTypes = {
  open: PropTypes.bool.isRequired,
  onCancel: PropTypes.func.isRequired,
  onSuccess: PropTypes.func.isRequired,
  selectedAssets: PropTypes.arrayOf(PropTypes.object).isRequired,
  projectId: PropTypes.string.isRequired,
};
