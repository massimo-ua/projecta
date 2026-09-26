import { t } from 'intlayer';

const invitationsManagementContent = {
  key: 'invitations-management',
  content: {
    title: t({
      en: 'Invitations',
      uk: 'Запрошення',
    }),
    subtitle: t({
      en: 'Create, view and manage user invitations',
      uk: 'Створюйте, переглядайте та керуйте запрошеннями для користувачів',
    }),
    rolesTab: t({
      en: 'Roles Management',
      uk: 'Керування ролями',
    }),
    invitationsTab: t({
      en: 'Invitations',
      uk: 'Запрошення',
    }),
    createButton: t({
      en: 'Create Invitation',
      uk: 'Створити запрошення',
    }),
    noInvitations: t({
      en: 'No invitations created yet',
      uk: 'Запрошень ще не створено',
    }),
    emailLabel: t({
      en: 'Google Email',
      uk: 'Google Email',
    }),
    statusLabel: t({
      en: 'Status',
      uk: 'Статус',
    }),
    createdLabel: t({
      en: 'Created',
      uk: 'Створено',
    }),
    expiresLabel: t({
      en: 'Expires',
      uk: 'Діє до',
    }),
    statusPending: t({
      en: 'Pending',
      uk: 'Очікує',
    }),
    statusCompleted: t({
      en: 'Completed',
      uk: 'Використано',
    }),
    statusExpired: t({
      en: 'Expired',
      uk: 'Прострочено',
    }),
    deleteTooltip: t({
      en: 'Delete Invitation',
      uk: 'Видалити запрошення',
    }),
    deleteConfirmTitle: t({
      en: 'Delete this invitation?',
      uk: 'Видалити це запрошення?',
    }),
    deleteConfirmDesc: t({
      en: 'Once deleted, this invitation is no longer valid and any precreated user record will be removed.',
      uk: 'Після видалення запрошення стане недійсним, а попередньо створений запис користувача буде видалено.',
    }),
    deleteSuccess: t({
      en: 'Invitation deleted successfully',
      uk: 'Запрошення успішно видалено',
    }),
    deleteError: t({
      en: 'Failed to delete invitation',
      uk: 'Не вдалося видалити запрошення',
    }),
    modalTitle: t({
      en: 'New Invitation',
      uk: 'Нове запрошення',
    }),
    modalDesc: t({
      en: 'Set Google email for the user. Invitation link with unique code valid for 7 days will be generated.',
      uk: 'Вкажіть Google email користувача. Буде згенеровано посилання з унікальним кодом, дійсним 7 днів.',
    }),
    modalNotice: t({
      en: 'This invitation code appears only once! Copy the link and share it with the user.',
      uk: 'Цей код запрошення відображається лише один раз! Скопіюйте посилання та надішліть його користувачу.',
    }),
    emailPlaceholder: t({
      en: 'user@gmail.com',
      uk: 'user@gmail.com',
    }),
    submitButton: t({
      en: 'Generate Invitation',
      uk: 'Згенерувати запрошення',
    }),
    invitationLinkLabel: t({
      en: 'Invitation Link',
      uk: 'Посилання для запрошення',
    }),
    invitationCodeLabel: t({
      en: 'Unique Code',
      uk: 'Унікальний код',
    }),
    copyLinkButton: t({
      en: 'Copy Link',
      uk: 'Скопіювати посилання',
    }),
    copyCodeButton: t({
      en: 'Copy Code',
      uk: 'Скопіювати код',
    }),
    linkCopied: t({
      en: 'Invitation link copied to clipboard',
      uk: 'Посилання на запрошення скопійовано',
    }),
    codeCopied: t({
      en: 'Code copied to clipboard',
      uk: 'Код скопійовано',
    }),
    doneButton: t({
      en: 'Done',
      uk: 'Готово',
    }),
    cancelButton: t({
      en: 'Cancel',
      uk: 'Скасувати',
    }),
  },
};

export default invitationsManagementContent;
