import { t } from 'intlayer';

const projectSettingsContent = {
  key: 'project-settings',
  content: {
    title: t({
      en: 'Project Settings',
      uk: 'Налаштування проєкту',
    }),
    homeCurrencyTitle: t({
      en: 'Home Currency',
      uk: 'Основна валюта',
    }),
    homeCurrencyDesc: t({
      en: 'Select the primary home currency for this project. All payments, assets, and summary totals will be converted to this currency using official National Bank of Ukraine (NBU) exchange rates.',
      uk: 'Оберіть основну валюту для цього проєкту. Усі платежі, активи та підсумкові суми перераховуватимуться у цю валюту за офіційними курсами НБУ.',
    }),
    selectCurrencyPlaceholder: t({
      en: 'Select currency',
      uk: 'Оберіть валюту',
    }),
    saveSettingsButton: t({
      en: 'Save Settings',
      uk: 'Зберегти налаштування',
    }),
    settingsUpdatedSuccess: t({
      en: 'Project home currency updated successfully',
      uk: 'Основну валюту проєкту успішно оновлено',
    }),
    failedToLoad: t({
      en: 'Failed to load project settings',
      uk: 'Не вдалося завантажити налаштування проєкту',
    }),
    failedToUpdate: t({
      en: 'Failed to update project settings',
      uk: 'Не вдалося оновити налаштування проєкту',
    }),
    projectSharingTitle: t({
      en: 'Project Sharing',
      uk: 'Поширення проєкту',
    }),
    projectSharingDesc: t({
      en: 'Share this link with team members to grant them access to this project.',
      uk: 'Поділіться цим посиланням із членами команди, щоб надати їм доступ до цього проєкту.',
    }),
    shareableLinkLabel: t({
      en: 'Shareable Link',
      uk: 'Посилання для доступу',
    }),
    copyButton: t({
      en: 'Copy',
      uk: 'Скопіювати',
    }),
    copiedButton: t({
      en: 'Copied',
      uk: 'Скопійовано',
    }),
    shareLinkCopied: t({
      en: 'Share link copied to clipboard!',
      uk: 'Посилання на проєкт скопійовано в буфер обміну!',
    }),
    failedToCopyLink: t({
      en: 'Failed to copy link',
      uk: 'Не вдалося скопіювати посилання',
    }),
  },
};

export default projectSettingsContent;
