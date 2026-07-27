import React, { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { useIntlayer } from 'react-intlayer';
import { typesRepository } from '../../api';
import useCategories from '../../hooks/categories.js';
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

export default function AddTypeModal(props) {
  const content = useIntlayer('types');
  const { projectId } = useParams();
  const { open, onSuccess, onCancel } = props;
  const [categoryId, setCategoryId] = useState('');
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [loading, setLoading] = useState(false);
  const [, categories, , setCategoriesFilter] = useCategories();

  useEffect(() => {
    if (open) {
      setCategoriesFilter({
        projectId,
        limit: 100,
        offset: 0,
      });
    }
  }, [open, projectId, setCategoriesFilter]);

  const handleAdd = async (e) => {
    e.preventDefault();
    if (!categoryId || !name) {
      toast.error(String(content.categoryAndNameRequiredError));
      return;
    }

    setLoading(true);
    try {
      await typesRepository.addType(projectId, {
        categoryId,
        name,
        description,
      });
      toast.success(String(content.typeAddedSuccess));
      setCategoryId('');
      setName('');
      setDescription('');
      onSuccess();
    } catch (e) {
      toast.error(`${String(content.failedToAddType)}: ${e.message}`);
      console.error('Failed to add type', e.message);
    } finally {
      setLoading(false);
    }
  };

  const handleCancel = () => {
    setCategoryId('');
    setName('');
    setDescription('');
    onCancel();
  };

  return (
    <Dialog open={open} onOpenChange={(isOpen) => !isOpen && handleCancel()}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>{String(content.addType)}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleAdd} className="space-y-4 py-2">
          <div className="space-y-2">
            <Label htmlFor="type-category">{String(content.categoryLabel)}</Label>
            <Select value={categoryId} onValueChange={setCategoryId}>
              <SelectTrigger id="type-category">
                <SelectValue placeholder={String(content.selectCategoryPlaceholder)} />
              </SelectTrigger>
              <SelectContent>
                {categories.map((category) => (
                  <SelectItem key={category.id} value={category.id}>
                    {category.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label htmlFor="type-name">{String(content.nameLabel)}</Label>
            <Input
              id="type-name"
              placeholder={String(content.namePlaceholder)}
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="type-desc">{String(content.descriptionLabel)}</Label>
            <Textarea
              id="type-desc"
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
