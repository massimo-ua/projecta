import React from 'react';
import { useNavigate } from 'react-router-dom';
import { useIntlayer, useLocale } from 'react-intlayer';
import { getLocalizedUrl } from 'intlayer';
import { ShieldCheck, ChevronLeft, ChevronRight, Inbox, Shield, Mail } from 'lucide-react';
import HomeLayout from '../../Layout';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { RoleBadge } from './RoleBadge';
import { EditRolesModal } from './EditRolesModal';
import { useRolesManagement } from '../../hooks/useRolesManagement';
import { CopyableText } from '../Projects/ListView/CopyableText';

export function RolesManagement() {
  const content = useIntlayer('roles-management');
  const { locale } = useLocale();
  const navigate = useNavigate();
  const {
    loading,
    users,
    total,
    currentPage,
    pageSize,
    editingUser,
    isUpdating,
    onPaginationChange,
    openEditRoles,
    closeEditRoles,
    saveRoles,
  } = useRolesManagement();

  const totalPages = Math.ceil(total / pageSize);

  const handleSaveRoles = (userId, roles) => saveRoles(userId, roles, {
    successMessage: String(content.saveSuccess),
    errorMessage: String(content.saveError),
  });

  return (
    <HomeLayout>
      <div className="space-y-6 max-w-5xl mx-auto">
        {/* Navigation Tabs */}
        <div className="flex items-center gap-2 border-b border-border/60 pb-3">
          <Button
            variant="secondary"
            size="sm"
            className="gap-2 rounded-xl font-semibold shadow-2xs"
          >
            <Shield className="h-4 w-4 text-amber-500" />
            <span>{String(content.title)}</span>
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => navigate(getLocalizedUrl('/admin/invitations', locale))}
            className="gap-2 rounded-xl text-muted-foreground hover:text-foreground"
          >
            <Mail className="h-4 w-4" />
            <span>Invitations</span>
          </Button>
        </div>

        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-4 border-b border-border/60">
          <div className="flex items-center gap-3">
            <div className="flex h-11 w-11 items-center justify-center rounded-2xl bg-amber-500/10 text-amber-600 dark:text-amber-400 ring-1 ring-amber-500/20">
              <Shield className="h-6 w-6" />
            </div>
            <div>
              <div className="flex items-center gap-2.5">
                <h1 className="text-xl sm:text-2xl font-bold tracking-tight text-foreground">
                  {String(content.title)}
                </h1>
                {!loading && total > 0 && (
                  <Badge variant="secondary" className="font-mono text-xs px-2 py-0.5 rounded-full font-semibold">
                    {total}
                  </Badge>
                )}
              </div>
              <p className="text-xs sm:text-sm text-muted-foreground mt-0.5">
                {String(content.subtitle)}
              </p>
            </div>
          </div>
        </div>

        {/* User List */}
        {loading ? (
          <div className="space-y-3">
            {[1, 2, 3, 4].map((i) => (
              <div key={i} className="rounded-2xl border border-border/60 p-5 space-y-3 bg-card">
                <div className="flex justify-between items-center">
                  <div className="flex items-center gap-3">
                    <Skeleton className="h-11 w-11 rounded-2xl" />
                    <div className="space-y-1.5">
                      <Skeleton className="h-5 w-36 rounded-md" />
                      <Skeleton className="h-3 w-48 rounded-md" />
                    </div>
                  </div>
                  <Skeleton className="h-9 w-28 rounded-xl" />
                </div>
              </div>
            ))}
          </div>
        ) : users.length === 0 ? (
          <div className="flex flex-col items-center justify-center p-12 text-center rounded-2xl border border-dashed border-border/80 bg-muted/20">
            <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-muted text-muted-foreground mb-3">
              <Inbox className="h-6 w-6" />
            </div>
            <p className="text-sm font-semibold text-foreground">
              {String(content.noUsers)}
            </p>
          </div>
        ) : (
          <div className="space-y-3">
            {users.map((user) => (
              <div
                key={user.id}
                className="group rounded-2xl border border-border/70 bg-card p-4 sm:p-5 shadow-xs transition-all duration-200 hover:border-border hover:shadow-sm"
              >
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                  {/* User info */}
                  <div className="flex items-center gap-3.5 min-w-0">
                    <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl bg-primary/10 text-primary font-bold text-sm ring-1 ring-primary/20">
                      {user.initials}
                    </div>
                    <div className="min-w-0">
                      <div className="flex items-center gap-2">
                        <span className="font-semibold text-sm sm:text-base text-foreground truncate">
                          {user.fullName}
                        </span>
                        {user.displayName && user.displayName !== user.fullName && (
                          <span className="text-xs text-muted-foreground/80 hidden sm:inline">
                            (@{user.displayName})
                          </span>
                        )}
                      </div>
                      <div className="flex items-center gap-1.5 text-xs text-muted-foreground mt-0.5">
                        <span className="font-mono text-[11px]">
                          <CopyableText text={user.id} truncate />
                        </span>
                      </div>
                    </div>
                  </div>

                  {/* Roles and Actions */}
                  <div className="flex items-center justify-between sm:justify-end gap-3 pt-2 sm:pt-0 border-t border-border/40 sm:border-0">
                    <div className="flex flex-wrap items-center gap-1.5">
                      {user.roles.map((r) => (
                        <RoleBadge key={r} role={r} />
                      ))}
                    </div>

                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => openEditRoles(user)}
                      className="gap-1.5 text-xs rounded-xl h-9 px-3 border-border/70 text-foreground hover:bg-accent/80 transition-colors shrink-0"
                    >
                      <ShieldCheck className="h-3.5 w-3.5 text-primary" />
                      <span>{String(content.editRoles)}</span>
                    </Button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="flex items-center justify-center gap-3 pt-4 border-t border-border/50">
            <Button
              variant="outline"
              size="sm"
              disabled={currentPage <= 1}
              onClick={() => onPaginationChange(currentPage - 1)}
              className="gap-1.5 text-xs rounded-xl h-9 px-3 border-border/70 text-muted-foreground hover:text-foreground"
            >
              <ChevronLeft className="h-3.5 w-3.5" />
              <span>{String(content.previous)}</span>
            </Button>

            <span className="text-xs text-muted-foreground font-medium px-2">
              {String(content.pageOf)} <strong className="text-foreground">{currentPage}</strong> {String(content.of)} {totalPages}
            </span>

            <Button
              variant="outline"
              size="sm"
              disabled={currentPage >= totalPages}
              onClick={() => onPaginationChange(currentPage + 1)}
              className="gap-1.5 text-xs rounded-xl h-9 px-3 border-border/70 text-muted-foreground hover:text-foreground"
            >
              <span>{String(content.next)}</span>
              <ChevronRight className="h-3.5 w-3.5" />
            </Button>
          </div>
        )}

        {/* Edit Modal */}
        <EditRolesModal
          open={Boolean(editingUser)}
          user={editingUser}
          onClose={closeEditRoles}
          onSave={handleSaveRoles}
          isSaving={isUpdating}
        />
      </div>
    </HomeLayout>
  );
}

export default RolesManagement;
