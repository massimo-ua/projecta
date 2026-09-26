import React, { useState } from 'react';
import PropTypes from 'prop-types';
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
import { Loader2 } from 'lucide-react';
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
      <DialogContent className="sm:max-w-[440px] rounded-2xl p-6">
        <DialogHeader>
          <DialogTitle className="text-xl font-bold tracking-tight">
            {String(content.createNewProjectTitle)}
          </DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4 py-2">
          <div className="space-y-2">
            <Label htmlFor="project-name" className="text-xs font-semibold">
              {String(content.projectNameLabel)}
            </Label>
            <Input
              id="project-name"
              placeholder={String(content.projectNamePlaceholder)}
              value={name}
              onChange={(e) => setName(e.target.value)}
              disabled={loading}
              className="rounded-xl h-10"
              required
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="project-description" className="text-xs font-semibold">
              {String(content.descriptionLabel)}
            </Label>
            <Textarea
              id="project-description"
              rows={3}
              placeholder={String(content.descriptionPlaceholder)}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              disabled={loading}
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
              <span>{String(content.createProjectButton)}</span>
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

AddProjectModal.propTypes = {
  open: PropTypes.bool.isRequired,
  onSuccess: PropTypes.func.isRequired,
  onCancel: PropTypes.func.isRequired,
};

export default AddProjectModal;
