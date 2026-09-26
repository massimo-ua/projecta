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
  },
};

export default userProfileSettingsContent;
