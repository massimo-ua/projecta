import { t } from 'intlayer';

const userProfileSettingsContent = {
  key: 'user-profile-settings',
  content: {
    title: t({
      en: 'User Profile Settings',
      uk: 'Налаштування профілю',
    }),
    subtitle: t({
      en: 'Manage your profile and interface preferences',
      uk: 'Керуйте вашим профілем та налаштуваннями інтерфейсу',
    }),
    languageSectionTitle: t({
      en: 'Interface Language',
      uk: 'Мова інтерфейсу',
    }),
    languageSectionDesc: t({
      en: 'Select your preferred language for the application interface.',
      uk: 'Оберіть бажану мову інтерфейсу застосунку.',
    }),
    selectLanguageLabel: t({
      en: 'Language',
      uk: 'Мова',
    }),
    languages: {
      en: t({
        en: 'English',
        uk: 'Англійська',
      }),
      uk: t({
        en: 'Ukrainian',
        uk: 'Українська',
      }),
    },
    themeSectionTitle: t({
      en: 'Appearance & Theme',
      uk: 'Зовнішній вигляд та тема',
    }),
    themeSectionDesc: t({
      en: 'Customize how Projecta looks on your device. Choose between light and dark themes.',
      uk: 'Налаштуйте вигляд Projecta на вашому пристрої. Оберіть світлу або темну тему.',
    }),
    selectThemeLabel: t({
      en: 'Theme',
      uk: 'Тема',
    }),
    themes: {
      light: t({
        en: 'Light Mode',
        uk: 'Світла тема',
      }),
      dark: t({
        en: 'Dark Mode',
        uk: 'Темна тема',
      }),
    },
    saveSuccess: t({
      en: 'Preferences updated successfully',
      uk: 'Налаштування успішно оновлено',
    }),
    displayNameSectionTitle: t({
      en: 'Display Name',
      uk: "Відображуване ім'я",
    }),
    displayNameSectionDesc: t({
      en: 'Update the name that is displayed across your workspaces and projects.',
      uk: "Оновіть ім'я, яке відображається у ваших робочих просторах та проєктах.",
    }),
    displayNameLabel: t({
      en: 'Display Name',
      uk: "Відображуване ім'я",
    }),
    displayNamePlaceholder: t({
      en: 'Enter your display name',
      uk: "Введіть ваше відображуване ім'я",
    }),
    saveDisplayNameBtn: t({
      en: 'Save',
      uk: 'Зберегти',
    }),
    savingDisplayNameBtn: t({
      en: 'Saving...',
      uk: 'Збереження...',
    }),
    displayNameSuccess: t({
      en: 'Display name updated successfully',
      uk: "Відображуване ім'я успішно оновлено",
    }),
    displayNameError: t({
      en: 'Failed to update display name',
      uk: "Не вдалося оновити відображуване ім'я",
    }),
    displayNameValidation: t({
      en: 'Display name must be less than 255 characters',
      uk: "Відображуване ім'я має бути менше 255 символів",
    }),
  },
};

export default userProfileSettingsContent;
