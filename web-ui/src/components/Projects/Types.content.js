import { t } from 'intlayer';

const typesContent = {
  key: 'types',
  content: {
    addType: t({
      en: 'Add Type',
      uk: 'Додати тип',
    }),
    categoryLabel: t({
      en: 'Category',
      uk: 'Категорія',
    }),
    selectCategoryPlaceholder: t({
      en: 'Select a category',
      uk: 'Оберіть категорію',
    }),
    nameLabel: t({
      en: 'Name',
      uk: 'Назва',
    }),
    namePlaceholder: t({
      en: 'e.g. Materials',
      uk: 'напр. Матеріали',
    }),
    descriptionLabel: t({
      en: 'Description',
      uk: 'Опис',
    }),
    descriptionPlaceholder: t({
      en: 'Type description...',
      uk: 'Опис типу...',
    }),
    noDescription: t({
      en: 'No description',
      uk: 'Немає опису',
    }),
    submitButton: t({
      en: 'Submit',
      uk: 'Зберегти',
    }),
    cancelButton: t({
      en: 'Cancel',
      uk: 'Скасувати',
    }),
    typeAddedSuccess: t({
      en: 'Type added successfully',
      uk: 'Тип успішно додано',
    }),
    typeRemovedSuccess: t({
      en: 'Type removed successfully',
      uk: 'Тип успішно видалено',
    }),
    failedToRemove: t({
      en: 'Failed to remove type',
      uk: 'Не вдалося видалити тип',
    }),
    failedToAddType: t({
      en: 'Failed to add type',
      uk: 'Не вдалося додати тип',
    }),
    categoryAndNameRequiredError: t({
      en: 'Category and Name are required',
      uk: 'Категорія та назва є обов’язковими',
    }),
  },
};

export default typesContent;
