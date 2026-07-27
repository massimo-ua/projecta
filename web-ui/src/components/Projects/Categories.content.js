import { t } from 'intlayer';

const categoriesContent = {
  key: 'categories',
  content: {
    addCategory: t({
      en: 'Add Category',
      uk: 'Додати категорію',
    }),
    descriptionLabel: t({
      en: 'Description',
      uk: 'Опис',
    }),
    noDescription: t({
      en: 'No description',
      uk: 'Немає опису',
    }),
    categoryRemovedSuccess: t({
      en: 'Category removed successfully',
      uk: 'Категорію успішно видалено',
    }),
    failedToRemove: t({
      en: 'Failed to remove category',
      uk: 'Не вдалося видалити категорію',
    }),
    nameLabel: t({
      en: 'Name',
      uk: 'Назва',
    }),
    namePlaceholder: t({
      en: 'e.g. Construction',
      uk: 'напр. Будівництво',
    }),
    descriptionPlaceholder: t({
      en: 'Category details...',
      uk: 'Деталі категорії...',
    }),
    submitButton: t({
      en: 'Submit',
      uk: 'Зберегти',
    }),
    cancelButton: t({
      en: 'Cancel',
      uk: 'Скасувати',
    }),
    nameRequiredError: t({
      en: 'Category name is required',
      uk: 'Назва категорії є обов’язковою',
    }),
    categoryCreatedSuccess: t({
      en: 'Category created successfully',
      uk: 'Категорію успішно створено',
    }),
    failedToAddCategory: t({
      en: 'Failed to add category',
      uk: 'Не вдалося додати категорію',
    }),
  },
};

export default categoriesContent;
