import React, { useState } from 'react';
import { useIntlayer } from 'react-intlayer';
import { projectsRepository } from '../../api';
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

export function AddProjectModal({ open, onSuccess, onCancel }) {
  const content = useIntlayer('projects');
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!name.trim()) {
      toast.error(String(content.nameRequiredError));
      return;
    }

    setLoading(true);
    try {
      await projectsRepository.createProject({
        name: name.trim(),
        description: description.trim(),
      });
      toast.success(String(content.projectCreatedSuccess));
      resetForm();
      onSuccess();
    } catch (error) {
      toast.error(`${String(content.failedToCreateProject)}: ${error.message || 'Unknown error'}`);
    } finally {
      setLoading(false);
    }
  };

  const resetForm = () => {
    setName('');
    setDescription('');
  };

  const handleCancel = () => {
    resetForm();
    onCancel();
  };

  return (
    <Dialog open={open} onOpenChange={(isOpen) => !isOpen && handleCancel()}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>{String(content.createNewProjectTitle)}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4 py-2">
          <div className="space-y-2">
            <Label htmlFor="project-name">{String(content.projectNameLabel)}</Label>
            <Input
              id="project-name"
              placeholder={String(content.projectNamePlaceholder)}
              value={name}
              onChange={(e) => setName(e.target.value)}
              disabled={loading}
              required
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="project-description">{String(content.descriptionLabel)}</Label>
            <Textarea
              id="project-description"
              rows={3}
              placeholder={String(content.descriptionPlaceholder)}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              disabled={loading}
            />
          </div>

          <DialogFooter className="pt-2">
            <Button type="button" variant="outline" onClick={handleCancel} disabled={loading}>
              {String(content.cancelButton)}
            </Button>
            <Button type="submit" disabled={loading}>
              {String(content.createProjectButton)}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

export default AddProjectModal;
