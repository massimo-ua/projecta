import React, { useState } from 'react';
import PropTypes from 'prop-types';
import { useParams } from 'react-router-dom';
import { useIntlayer } from 'react-intlayer';
import { categoriesRepository } from '../../api';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog';
import { Loader2 } from 'lucide-react';
import { toast } from 'sonner';

export default function AddCategoryModal(props) {
  const content = useIntlayer('categories');
  const { projectId } = useParams();
  const { open, onSuccess, onCancel } = props;
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [loading, setLoading] = useState(false);

  const handleAdd = async (e) => {
    e.preventDefault();
    if (!name) {
      toast.error(String(content.nameRequiredError));
      return;
    }
    setLoading(true);
    try {
      await categoriesRepository.addCategory(projectId, { name, description });
      toast.success(String(content.categoryCreatedSuccess));
      setName('');
      setDescription('');
      onSuccess();
    } catch (e) {
      toast.error(`${String(content.failedToAddCategory)}: ${e.message}`);
      console.error('Failed to add category', e.message);
    } finally {
      setLoading(false);
    }
  };

  const handleCancel = () => {
    setName('');
    setDescription('');
    onCancel();
  };

  return (
    <Dialog open={open} onOpenChange={(isOpen) => !isOpen && handleCancel()}>
      <DialogContent className="sm:max-w-[440px] rounded-2xl p-6">
        <DialogHeader>
          <DialogTitle className="text-xl font-bold tracking-tight">
            {String(content.addCategory)}
          </DialogTitle>
        </DialogHeader>
        <form onSubmit={handleAdd} className="space-y-4 py-2">
          <div className="space-y-2">
            <Label htmlFor="category-name" className="text-xs font-semibold">
              {String(content.nameLabel)}
            </Label>
            <Input
              id="category-name"
              placeholder={String(content.namePlaceholder)}
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="rounded-xl"
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="category-desc" className="text-xs font-semibold">
              {String(content.descriptionLabel)}
            </Label>
            <Textarea
              id="category-desc"
              rows={3}
              placeholder={String(content.descriptionPlaceholder)}
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
              {String(content.cancelButton)}
            </Button>
            <Button
              type="submit"
              disabled={loading}
              className="rounded-xl font-semibold shadow-sm shadow-primary/20 gap-2"
            >
              {loading && <Loader2 className="h-4 w-4 animate-spin" />}
              <span>{String(content.submitButton)}</span>
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

AddCategoryModal.propTypes = {
  open: PropTypes.bool.isRequired,
  onSuccess: PropTypes.func.isRequired,
  onCancel: PropTypes.func.isRequired,
};
