import { t } from 'intlayer';

const rolesManagementContent = {
  key: 'roles-management',
  content: {
    title: t({
      en: 'Roles Management',
      uk: 'Керування ролями',
    }),
    subtitle: t({
      en: 'Manage user access and administrative roles',
      uk: 'Керування доступом користувачів та адміністративними ролями',
    }),
    totalUsers: t({
      en: 'Total Users',
      uk: 'Всього користувачів',
    }),
    noUsers: t({
      en: 'No users found',
      uk: 'Користувачів не знайдено',
    }),
    editRoles: t({
      en: 'Edit Roles',
      uk: 'Змінити ролі',
    }),
    assignedRoles: t({
      en: 'Assigned Roles',
      uk: 'Призначені ролі',
    }),
    userId: t({
      en: 'User ID',
      uk: 'ID користувача',
    }),
    saveSuccess: t({
      en: 'User roles updated successfully',
      uk: 'Ролі користувача успішно оновлено',
    }),
    saveError: t({
      en: 'Failed to update user roles',
      uk: 'Не вдалося оновити ролі користувача',
    }),
    editModalTitle: t({
      en: 'Manage User Roles',
      uk: 'Керування ролями користувача',
    }),
    editModalDesc: t({
      en: 'Configure permissions and system roles for this user.',
      uk: 'Налаштуйте дозволи та системні ролі для цього користувача.',
    }),
    roleUserTitle: t({
      en: 'User',
      uk: 'Користувач',
    }),
    roleUserDesc: t({
      en: 'Standard access to projects, payments, and assets',
      uk: 'Стандартний доступ до проєктів, платежів та активів',
    }),
    roleAdminTitle: t({
      en: 'Administrator',
      uk: 'Адміністратор',
    }),
    roleAdminDesc: t({
      en: 'Full access to system configuration and user roles management',
      uk: 'Повний доступ до налаштувань системи та керування ролями',
    }),
    atLeastOneRoleError: t({
      en: 'User must have at least one role assigned',
      uk: 'Користувач повинен мати принаймні одну призначену роль',
    }),
    cancel: t({
      en: 'Cancel',
      uk: 'Скасувати',
    }),
    save: t({
      en: 'Save Changes',
      uk: 'Зберегти зміни',
    }),
    previous: t({
      en: 'Previous',
      uk: 'Попередня',
    }),
    next: t({
      en: 'Next',
      uk: 'Наступна',
    }),
    pageOf: t({
      en: 'Page',
      uk: 'Сторінка',
    }),
    of: t({
      en: 'of',
      uk: 'з',
    }),
  },
};

export default rolesManagementContent;
