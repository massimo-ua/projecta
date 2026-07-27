import React, { useState } from 'react';
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
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>{String(content.addCategory)}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleAdd} className="space-y-4 py-2">
          <div className="space-y-2">
            <Label htmlFor="category-name">{String(content.nameLabel)}</Label>
            <Input
              id="category-name"
              placeholder={String(content.namePlaceholder)}
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="category-desc">{String(content.descriptionLabel)}</Label>
            <Textarea
              id="category-desc"
              rows={3}
              placeholder={String(content.descriptionPlaceholder)}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
          </div>
          <DialogFooter className="pt-2">
            <Button type="button" variant="outline" onClick={handleCancel} disabled={loading}>
              {String(content.cancelButton)}
            </Button>
            <Button type="submit" disabled={loading}>
              {String(content.submitButton)}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
