import React, { useState, useEffect } from 'react';
import PropTypes from 'prop-types';
import { useIntlayer } from 'react-intlayer';
import { Button } from '@/components/ui/button';
import { Switch } from '@/components/ui/switch';
import { Label } from '@/components/ui/label';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog';
import { Loader2, ShieldCheck, User as UserIcon } from 'lucide-react';
import { ROLES, User } from '../../models/User.js';

export function EditRolesModal({
  open,
  user,
  onClose,
  onSave,
  isSaving,
}) {
  const content = useIntlayer('roles-management');
  const [selectedRoles, setSelectedRoles] = useState([]);

  useEffect(() => {
    if (user) {
      setSelectedRoles([...user.roles]);
    }
  }, [user]);

  const handleToggleRole = (role) => {
    setSelectedRoles((prev) => {
      const exists = prev.some((r) => r.toLowerCase() === role.toLowerCase());
      if (exists) {
        return prev.filter((r) => r.toLowerCase() !== role.toLowerCase());
      }
      return [...prev, role];
    });
  };

  const validation = User.validateRoles(selectedRoles);
  const isValid = validation.valid;

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!isValid || !user) return;
    await onSave(user.id, selectedRoles);
  };

  return (
    <Dialog open={open} onOpenChange={(isOpen) => !isOpen && onClose()}>
      <DialogContent className="sm:max-w-[480px] rounded-2xl p-6">
        <DialogHeader className="space-y-1">
          <DialogTitle className="text-xl font-bold tracking-tight">
            {String(content.editModalTitle)}
          </DialogTitle>
          <DialogDescription className="text-xs text-muted-foreground">
            {user ? `${user.fullName} (${user.id})` : String(content.editModalDesc)}
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4 py-2">
          <div className="space-y-3">
            {/* User Role Card */}
            <div className="flex items-center justify-between p-3.5 rounded-xl border border-border/70 bg-card hover:bg-accent/40 transition-colors">
              <div className="flex items-center gap-3">
                <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-primary/10 text-primary">
                  <UserIcon className="h-4 w-4" />
                </div>
                <div>
                  <Label htmlFor="role-user-switch" className="text-sm font-semibold cursor-pointer">
                    {String(content.roleUserTitle)}
                  </Label>
                  <p className="text-xs text-muted-foreground mt-0.5">
                    {String(content.roleUserDesc)}
                  </p>
                </div>
              </div>
              <Switch
                id="role-user-switch"
                checked={selectedRoles.some((r) => r.toLowerCase() === ROLES.USER.toLowerCase())}
                onCheckedChange={() => handleToggleRole(ROLES.USER)}
                disabled={isSaving}
              />
            </div>

            {/* Administrator Role Card */}
            <div className="flex items-center justify-between p-3.5 rounded-xl border border-border/70 bg-card hover:bg-accent/40 transition-colors">
              <div className="flex items-center gap-3">
                <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-amber-500/10 text-amber-600 dark:text-amber-400">
                  <ShieldCheck className="h-4 w-4" />
                </div>
                <div>
                  <Label htmlFor="role-admin-switch" className="text-sm font-semibold cursor-pointer">
                    {String(content.roleAdminTitle)}
                  </Label>
                  <p className="text-xs text-muted-foreground mt-0.5">
                    {String(content.roleAdminDesc)}
                  </p>
                </div>
              </div>
              <Switch
                id="role-admin-switch"
                checked={selectedRoles.some((r) => r.toLowerCase() === ROLES.ADMINISTRATOR.toLowerCase())}
                onCheckedChange={() => handleToggleRole(ROLES.ADMINISTRATOR)}
                disabled={isSaving}
              />
            </div>
          </div>

          {!isValid && (
            <p className="text-xs text-destructive font-medium pt-1">
              {String(content.atLeastOneRoleError)}
            </p>
          )}

          <DialogFooter className="pt-3 gap-2 sm:gap-0">
            <Button
              type="button"
              variant="outline"
              onClick={onClose}
              disabled={isSaving}
              className="rounded-xl"
            >
              {String(content.cancel)}
            </Button>
            <Button
              type="submit"
              disabled={!isValid || isSaving}
              className="rounded-xl font-semibold shadow-sm shadow-primary/20 gap-2"
            >
              {isSaving && <Loader2 className="h-4 w-4 animate-spin" />}
              <span>{String(content.save)}</span>
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

EditRolesModal.propTypes = {
  open: PropTypes.bool.isRequired,
  user: PropTypes.shape({
    id: PropTypes.string,
    fullName: PropTypes.string,
    roles: PropTypes.arrayOf(PropTypes.string),
  }),
  onClose: PropTypes.func.isRequired,
  onSave: PropTypes.func.isRequired,
  isSaving: PropTypes.bool,
};

EditRolesModal.defaultProps = {
  user: null,
  isSaving: false,
};

export default EditRolesModal;
