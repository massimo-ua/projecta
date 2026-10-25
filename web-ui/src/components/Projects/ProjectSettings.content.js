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
    timelineTitle: t({
      en: 'Project Timeline',
      uk: 'Терміни проєкту',
    }),
    timelineDesc: t({
      en: 'Configure the project start date and optional target end date. Ongoing projects have no end date.',
      uk: 'Налаштуйте дату початку проєкту та необов’язкову цільову дату завершення. Безстрокові проєкти не мають дати завершення.',
    }),
    startDateLabel: t({
      en: 'Start Date',
      uk: 'Дата початку',
    }),
    endDateLabel: t({
      en: 'Target End Date (Optional)',
      uk: 'Цільова дата завершення (Необов’язково)',
    }),
    clearEndDateButton: t({
      en: 'Clear end date (Ongoing)',
      uk: 'Очистити дату (Безстроковий)',
    }),
    endDateHint: t({
      en: 'Leave empty for an ongoing project without a deadline.',
      uk: 'Залиште порожнім для безстрокового проєкту без фіксованого терміну.',
    }),
    invalidEndDateError: t({
      en: 'End date cannot be earlier than start date.',
      uk: 'Дата завершення не може бути раніше дати початку.',
    }),
    settingsUpdatedSuccess: t({
      en: 'Project settings updated successfully',
      uk: 'Налаштування проєкту успішно оновлено',
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
