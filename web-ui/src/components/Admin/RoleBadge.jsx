import React from 'react';
import PropTypes from 'prop-types';
import { Badge } from '@/components/ui/badge';
import { ShieldCheck, User as UserIcon } from 'lucide-react';
import { ROLES } from '../../models/User.js';

export function RoleBadge({ role }) {
  const isRoleAdmin = role?.toLowerCase() === ROLES.ADMINISTRATOR.toLowerCase();

  if (isRoleAdmin) {
    return (
      <Badge
        variant="default"
        className="gap-1 px-2.5 py-0.5 rounded-lg font-medium text-xs bg-amber-600/15 text-amber-800 dark:bg-amber-400/15 dark:text-amber-300 border border-amber-500/20"
      >
        <ShieldCheck className="h-3.5 w-3.5" />
        <span>{ROLES.ADMINISTRATOR}</span>
      </Badge>
    );
  }

  return (
    <Badge
      variant="secondary"
      className="gap-1 px-2.5 py-0.5 rounded-lg font-medium text-xs bg-primary/10 text-primary border border-primary/20"
    >
      <UserIcon className="h-3.5 w-3.5" />
      <span>{ROLES.USER}</span>
    </Badge>
  );
}

RoleBadge.propTypes = {
  role: PropTypes.string.isRequired,
};

export default RoleBadge;
