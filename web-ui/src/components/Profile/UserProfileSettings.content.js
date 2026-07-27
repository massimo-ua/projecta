import { t } from 'intlayer';

const userProfileSettingsContent = {
  key: 'user-profile-settings',
  content: {
    title: t({
      en: 'User Profile Settings',
      uk: 'Налаштування профілю користувача',
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
      en: 'Select your preferred language for the application interface. This choice will be saved for your profile.',
      uk: 'Оберіть бажану мову інтерфейсу застосунку. Цей вибір буде збережено для вашого профілю.',
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
    saveSuccess: t({
      en: 'Interface language updated successfully',
      uk: 'Мову інтерфейсу успішно оновлено',
    }),
  },
};

export default userProfileSettingsContent;
