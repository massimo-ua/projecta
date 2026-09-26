import React, { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { useIntlayer, useLocale } from 'react-intlayer';
import { getLocalizedUrl } from 'intlayer';
import {
  Mail,
  Plus,
  Trash2,
  Inbox,
  Shield,
  Clock,
  CheckCircle2,
  AlertCircle,
  Calendar,
} from 'lucide-react';
import { format } from 'date-fns';
import { toast } from 'sonner';
import HomeLayout from '../../Layout';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog';
import { CreateInvitationModal } from './CreateInvitationModal';
import { invitationsRepository } from '../../api';

export function InvitationsManagement() {
  const content = useIntlayer('invitations-management');
  const { locale } = useLocale();
  const navigate = useNavigate();

  const [invitations, setInvitations] = useState([]);
  const [loading, setLoading] = useState(true);
  const [modalOpen, setModalOpen] = useState(false);

  const fetchInvitations = useCallback(async () => {
    setLoading(true);
    try {
      const list = await invitationsRepository.list();
      setInvitations(list);
    } catch (err) {
      toast.error(err.message || 'Failed to fetch invitations');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchInvitations();
  }, [fetchInvitations]);

  const handleDelete = async (id) => {
    try {
      await invitationsRepository.delete(id);
      toast.success(String(content.deleteSuccess));
      fetchInvitations();
    } catch (err) {
      toast.error(err.message || String(content.deleteError));
    }
  };

  const getStatusInfo = (inv) => {
    const isExpired = new Date(inv.expires_at) < new Date();
    if (inv.status === 'COMPLETED') {
      return {
        label: String(content.statusCompleted),
        icon: CheckCircle2,
        className: 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20',
      };
    }
    if (isExpired) {
      return {
        label: String(content.statusExpired),
        icon: AlertCircle,
        className: 'bg-rose-500/10 text-rose-600 dark:text-rose-400 border-rose-500/20',
      };
    }
    return {
      label: String(content.statusPending),
      icon: Clock,
      className: 'bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/20',
    };
  };

  return (
    <HomeLayout>
      <div className="space-y-6 max-w-5xl mx-auto">
        {/* Navigation Tabs */}
        <div className="flex items-center gap-2 border-b border-border/60 pb-3">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => navigate(getLocalizedUrl('/admin/roles', locale))}
            className="gap-2 rounded-xl text-muted-foreground hover:text-foreground"
          >
            <Shield className="h-4 w-4" />
            <span>{String(content.rolesTab)}</span>
          </Button>
          <Button
            variant="secondary"
            size="sm"
            className="gap-2 rounded-xl font-semibold shadow-2xs"
          >
            <Mail className="h-4 w-4 text-primary" />
            <span>{String(content.invitationsTab)}</span>
          </Button>
        </div>

        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-2">
          <div className="flex items-center gap-3">
            <div className="flex h-11 w-11 items-center justify-center rounded-2xl bg-primary/10 text-primary ring-1 ring-primary/20">
              <Mail className="h-6 w-6" />
            </div>
            <div>
              <div className="flex items-center gap-2.5">
                <h1 className="text-xl sm:text-2xl font-bold tracking-tight text-foreground">
                  {String(content.title)}
                </h1>
                {!loading && invitations.length > 0 && (
                  <Badge variant="secondary" className="font-mono text-xs px-2 py-0.5 rounded-full font-semibold">
                    {invitations.length}
                  </Badge>
                )}
              </div>
              <p className="text-xs sm:text-sm text-muted-foreground mt-0.5">
                {String(content.subtitle)}
              </p>
            </div>
          </div>

          <Button
            onClick={() => setModalOpen(true)}
            className="gap-2 rounded-xl h-10 px-4 font-semibold shadow-sm shadow-primary/20"
          >
            <Plus className="h-4 w-4" />
            <span>{String(content.createButton)}</span>
          </Button>
        </div>

        {/* Invitations List */}
        {loading ? (
          <div className="space-y-3">
            {[1, 2, 3].map((i) => (
              <div key={i} className="rounded-2xl border border-border/60 p-5 space-y-3 bg-card">
                <div className="flex justify-between items-center">
                  <div className="flex items-center gap-3">
                    <Skeleton className="h-11 w-11 rounded-2xl" />
                    <div className="space-y-1.5">
                      <Skeleton className="h-5 w-40 rounded-md" />
                      <Skeleton className="h-3 w-48 rounded-md" />
                    </div>
                  </div>
                  <Skeleton className="h-8 w-20 rounded-xl" />
                </div>
              </div>
            ))}
          </div>
        ) : invitations.length === 0 ? (
          <div className="flex flex-col items-center justify-center p-12 text-center rounded-2xl border border-dashed border-border/80 bg-muted/20">
            <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-muted text-muted-foreground mb-3">
              <Inbox className="h-6 w-6" />
            </div>
            <p className="text-sm font-semibold text-foreground">
              {String(content.noInvitations)}
            </p>
          </div>
        ) : (
          <div className="space-y-3">
            {invitations.map((inv) => {
              const statusInfo = getStatusInfo(inv);
              const StatusIcon = statusInfo.icon;

              let formattedCreated = '';
              try {
                formattedCreated = format(new Date(inv.created_at), 'PPP');
              } catch {
                formattedCreated = inv.created_at;
              }

              let formattedExpires = '';
              try {
                formattedExpires = format(new Date(inv.expires_at), 'PPP');
              } catch {
                formattedExpires = inv.expires_at;
              }

              return (
                <div
                  key={inv.id}
                  className="group rounded-2xl border border-border/70 bg-card p-4 sm:p-5 shadow-xs transition-all duration-200 hover:border-border hover:shadow-sm"
                >
                  <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                    {/* Info */}
                    <div className="flex items-center gap-3.5 min-w-0">
                      <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl bg-primary/10 text-primary font-bold text-sm ring-1 ring-primary/20">
                        <Mail className="h-5 w-5" />
                      </div>
                      <div className="min-w-0">
                        <div className="flex items-center gap-2">
                          <span className="font-semibold text-sm sm:text-base text-foreground truncate">
                            {inv.email}
                          </span>
                        </div>
                        <div className="flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-muted-foreground mt-0.5">
                          <span className="flex items-center gap-1">
                            <Calendar className="h-3 w-3" />
                            {String(content.createdLabel)}: {formattedCreated}
                          </span>
                          <span className="flex items-center gap-1">
                            <Clock className="h-3 w-3" />
                            {String(content.expiresLabel)}: {formattedExpires}
                          </span>
                        </div>
                      </div>
                    </div>

                    {/* Status & Actions */}
                    <div className="flex items-center justify-between sm:justify-end gap-3 pt-2 sm:pt-0 border-t border-border/40 sm:border-0">
                      <Badge
                        variant="outline"
                        className={`gap-1.5 py-1 px-2.5 rounded-lg text-xs font-semibold ${statusInfo.className}`}
                      >
                        <StatusIcon className="h-3.5 w-3.5" />
                        <span>{statusInfo.label}</span>
                      </Badge>

                      <AlertDialog>
                        <AlertDialogTrigger asChild>
                          <Button
                            variant="ghost"
                            size="icon"
                            title={String(content.deleteTooltip)}
                            className="h-9 w-9 rounded-xl text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors shrink-0"
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </AlertDialogTrigger>
                        <AlertDialogContent className="rounded-2xl sm:max-w-md">
                          <AlertDialogHeader>
                            <AlertDialogTitle className="text-lg font-bold">
                              {String(content.deleteConfirmTitle)}
                            </AlertDialogTitle>
                            <AlertDialogDescription className="text-sm text-muted-foreground">
                              {String(content.deleteConfirmDesc)}
                            </AlertDialogDescription>
                          </AlertDialogHeader>
                          <AlertDialogFooter className="gap-2 sm:gap-0 pt-2">
                            <AlertDialogCancel className="rounded-xl">
                              {String(content.cancelButton)}
                            </AlertDialogCancel>
                            <AlertDialogAction
                              onClick={() => handleDelete(inv.id)}
                              className="rounded-xl bg-destructive text-destructive-foreground hover:bg-destructive/90 transition-colors"
                            >
                              {String(content.deleteTooltip)}
                            </AlertDialogAction>
                          </AlertDialogFooter>
                        </AlertDialogContent>
                      </AlertDialog>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        )}

        <CreateInvitationModal
          open={modalOpen}
          onClose={() => setModalOpen(false)}
          onSuccess={fetchInvitations}
        />
      </div>
    </HomeLayout>
  );
}

export default InvitationsManagement;
