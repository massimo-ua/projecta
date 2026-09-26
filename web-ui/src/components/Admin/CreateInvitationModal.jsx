import React, { useState } from 'react';
import PropTypes from 'prop-types';
import { useIntlayer, useLocale } from 'react-intlayer';
import { getLocalizedUrl } from 'intlayer';
import { toast } from 'sonner';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Loader2, Copy, Check, AlertTriangle, Mail } from 'lucide-react';
import { invitationsRepository } from '../../api';

export function CreateInvitationModal({ open, onClose, onSuccess }) {
  const content = useIntlayer('invitations-management');
  const { locale } = useLocale();
  const [email, setEmail] = useState('');
  const [loading, setLoading] = useState(false);
  const [createdData, setCreatedData] = useState(null);
  const [linkCopied, setLinkCopied] = useState(false);
  const [codeCopied, setCodeCopied] = useState(false);

  const resetState = () => {
    setEmail('');
    setCreatedData(null);
    setLinkCopied(false);
    setCodeCopied(false);
    setLoading(false);
  };

  const handleClose = () => {
    resetState();
    onClose();
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!email) return;

    setLoading(true);
    try {
      const data = await invitationsRepository.create(email.trim());
      const inviteUrl = `${window.location.origin}${getLocalizedUrl(`/invite/${data.code}`, locale)}`;
      setCreatedData({
        ...data,
        link: inviteUrl,
      });
      if (onSuccess) {
        onSuccess();
      }
    } catch (err) {
      toast.error(err.message || String(content.createError));
    } finally {
      setLoading(false);
    }
  };

  const handleCopyLink = async () => {
    if (!createdData?.link) return;
    try {
      await navigator.clipboard.writeText(createdData.link);
      setLinkCopied(true);
      toast.success(String(content.linkCopied));
      setTimeout(() => setLinkCopied(false), 2000);
    } catch {
      toast.error('Failed to copy to clipboard');
    }
  };

  const handleCopyCode = async () => {
    if (!createdData?.code) return;
    try {
      await navigator.clipboard.writeText(createdData.code);
      setCodeCopied(true);
      toast.success(String(content.codeCopied));
      setTimeout(() => setCodeCopied(false), 2000);
    } catch {
      toast.error('Failed to copy to clipboard');
    }
  };

  return (
    <Dialog open={open} onOpenChange={(isOpen) => !isOpen && handleClose()}>
      <DialogContent className="sm:max-w-[500px] rounded-2xl p-6">
        <DialogHeader className="space-y-1">
          <DialogTitle className="text-xl font-bold tracking-tight">
            {String(content.modalTitle)}
          </DialogTitle>
          <DialogDescription className="text-xs text-muted-foreground">
            {String(content.modalDesc)}
          </DialogDescription>
        </DialogHeader>

        {!createdData ? (
          <form onSubmit={handleSubmit} className="space-y-4 py-2">
            <div className="space-y-2">
              <Label htmlFor="invite-email" className="text-xs font-semibold text-foreground">
                {String(content.emailLabel)}
              </Label>
              <div className="relative">
                <Mail className="absolute left-3.5 top-3.5 h-4 w-4 text-muted-foreground" />
                <Input
                  id="invite-email"
                  type="email"
                  placeholder={String(content.emailPlaceholder)}
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  className="pl-10 h-11 rounded-xl"
                  disabled={loading}
                  required
                />
              </div>
            </div>

            <DialogFooter className="pt-3 gap-2 sm:gap-0">
              <Button
                type="button"
                variant="outline"
                onClick={handleClose}
                disabled={loading}
                className="rounded-xl"
              >
                {String(content.cancelButton)}
              </Button>
              <Button
                type="submit"
                disabled={!email || loading}
                className="rounded-xl font-semibold shadow-sm shadow-primary/20 gap-2"
              >
                {loading && <Loader2 className="h-4 w-4 animate-spin" />}
                <span>{String(content.submitButton)}</span>
              </Button>
            </DialogFooter>
          </form>
        ) : (
          <div className="space-y-4 py-2">
            {/* Warning Alert */}
            <div className="flex items-start gap-3 p-3.5 rounded-xl border border-amber-500/30 bg-amber-500/10 text-amber-700 dark:text-amber-400 text-xs leading-relaxed">
              <AlertTriangle className="h-5 w-5 shrink-0 mt-0.5" />
              <div>
                <p className="font-semibold">{String(content.modalNotice)}</p>
              </div>
            </div>

            {/* Generated Link */}
            <div className="space-y-1.5">
              <Label className="text-xs font-semibold text-foreground">
                {String(content.invitationLinkLabel)}
              </Label>
              <div className="flex gap-2">
                <Input
                  readOnly
                  value={createdData.link}
                  className="h-10 rounded-xl font-mono text-xs bg-muted/50"
                />
                <Button
                  type="button"
                  variant="outline"
                  size="icon"
                  onClick={handleCopyLink}
                  title={String(content.copyLinkButton)}
                  className="h-10 w-10 shrink-0 rounded-xl"
                >
                  {linkCopied ? (
                    <Check className="h-4 w-4 text-emerald-500" />
                  ) : (
                    <Copy className="h-4 w-4" />
                  )}
                </Button>
              </div>
            </div>

            {/* Unique Code */}
            <div className="space-y-1.5">
              <Label className="text-xs font-semibold text-foreground">
                {String(content.invitationCodeLabel)}
              </Label>
              <div className="flex gap-2">
                <Input
                  readOnly
                  value={createdData.code}
                  className="h-10 rounded-xl font-mono text-xs bg-muted/50 truncate"
                />
                <Button
                  type="button"
                  variant="outline"
                  size="icon"
                  onClick={handleCopyCode}
                  title={String(content.copyCodeButton)}
                  className="h-10 w-10 shrink-0 rounded-xl"
                >
                  {codeCopied ? (
                    <Check className="h-4 w-4 text-emerald-500" />
                  ) : (
                    <Copy className="h-4 w-4" />
                  )}
                </Button>
              </div>
            </div>

            <DialogFooter className="pt-2">
              <Button
                type="button"
                onClick={handleClose}
                className="w-full rounded-xl font-semibold"
              >
                {String(content.doneButton)}
              </Button>
            </DialogFooter>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}

CreateInvitationModal.propTypes = {
  open: PropTypes.bool.isRequired,
  onClose: PropTypes.func.isRequired,
  onSuccess: PropTypes.func,
};

CreateInvitationModal.defaultProps = {
  onSuccess: null,
};

export default CreateInvitationModal;
