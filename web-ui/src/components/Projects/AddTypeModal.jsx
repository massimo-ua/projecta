import React, { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
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
      toast.error('Category and Name are required');
      return;
    }

    setLoading(true);
    try {
      await typesRepository.addType(projectId, {
        categoryId,
        name,
        description,
      });
      toast.success('Type added successfully');
      setCategoryId('');
      setName('');
      setDescription('');
      onSuccess();
    } catch (e) {
      toast.error(`Failed to add type: ${e.message}`);
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
          <DialogTitle>Add Type</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleAdd} className="space-y-4 py-2">
          <div className="space-y-2">
            <Label htmlFor="type-category">Category</Label>
            <Select value={categoryId} onValueChange={setCategoryId}>
              <SelectTrigger id="type-category">
                <SelectValue placeholder="Select a category" />
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
            <Label htmlFor="type-name">Name</Label>
            <Input
              id="type-name"
              placeholder="e.g. Materials"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="type-desc">Description</Label>
            <Textarea
              id="type-desc"
              rows={3}
              placeholder="Type description..."
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
