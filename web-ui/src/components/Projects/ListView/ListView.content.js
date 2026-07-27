import { t } from 'intlayer';

const listViewContent = {
  key: 'list-view',
  content: {
    detailsAndActions: t({
      en: 'Details & Actions',
      uk: 'Деталі та дії',
    }),
    totalItemsLabel: t({
      en: 'Total items',
      uk: 'Всього елементів',
    }),
    noRecords: t({
      en: 'No records available',
      uk: 'Записи відсутні',
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
    remove: t({
      en: 'Remove',
      uk: 'Видалити',
    }),
    confirmRemoval: t({
      en: 'Confirm removal',
      uk: 'Підтвердження видалення',
    }),
    removeWarning: t({
      en: 'Are you sure you want to remove this item? This action cannot be undone.',
      uk: 'Ви впевнені, що хочете видалити цей елемент? Цю дію неможливо скасувати.',
    }),
    cancel: t({
      en: 'Cancel',
      uk: 'Скасувати',
    }),
    edit: t({
      en: 'Edit',
      uk: 'Редагувати',
    }),
  },
};

export default listViewContent;
