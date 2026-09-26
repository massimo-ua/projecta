export const ROLES = Object.freeze({
  ADMINISTRATOR: 'Administrator',
  USER: 'User',
});

export const AVAILABLE_ROLES = Object.freeze([
  ROLES.ADMINISTRATOR,
  ROLES.USER,
]);

export class User {
  constructor({
    id,
    customerId,
    firstName = '',
    lastName = '',
    displayName = '',
    roles = [],
  } = {}) {
    this.id = id || customerId || '';
    this.customerId = this.id;
    this.firstName = firstName || '';
    this.lastName = lastName || '';
    this.displayName = displayName || '';
    this.roles = Array.isArray(roles) ? [...roles] : [];
  }

  get fullName() {
    const combined = `${this.firstName} ${this.lastName}`.trim();
    return combined || this.displayName || this.id;
  }

  get initials() {
    const name = this.fullName;
    if (!name) return '?';
    const parts = name.split(/\s+/).filter(Boolean);
    if (parts.length >= 2) {
      return (parts[0][0] + parts[1][0]).toUpperCase();
    }
    return name.slice(0, 2).toUpperCase();
  }

  hasRole(role) {
    if (!role) return false;
    const target = String(role).toLowerCase();
    return this.roles.some((r) => String(r).toLowerCase() === target);
  }

  get isAdmin() {
    return this.hasRole(ROLES.ADMINISTRATOR) || this.hasRole('admin');
  }

  get isAdministrator() {
    return this.isAdmin;
  }

  get isUser() {
    return this.hasRole(ROLES.USER);
  }

  get canManageRoles() {
    return this.isAdmin;
  }

  static validateRoles(roles) {
    if (!Array.isArray(roles) || roles.length === 0) {
      return { valid: false, error: 'At least one role must be assigned' };
    }
    const invalid = roles.filter((r) => !AVAILABLE_ROLES.includes(r));
    if (invalid.length > 0) {
      return { valid: false, error: `Invalid role: ${invalid.join(', ')}` };
    }
    return { valid: true };
  }
}

export default User;
