import { t } from 'intlayer';

const assetsContent = {
  key: 'assets',
  content: {
    addAsset: t({
      en: 'Add Asset',
      uk: 'Додати актив',
    }),
    addAssetTitle: t({
      en: 'Add Asset',
      uk: 'Додати актив',
    }),
    editAssetTitle: t({
      en: 'Edit Asset',
      uk: 'Редагувати актив',
    }),
    typeLabel: t({
      en: 'Type',
      uk: 'Тип',
    }),
    selectTypePlaceholder: t({
      en: 'Select type',
      uk: 'Оберіть тип',
    }),
    categoryLabel: t({
      en: 'Category',
      uk: 'Категорія',
    }),
    createPaymentLabel: t({
      en: 'Create Payment',
      uk: 'Створити платіж',
    }),
    createPaymentDescription: t({
      en: 'Automatically record an associated payment entry',
      uk: 'Автоматично створити відповідний платіжний запис',
    }),
    priceLabel: t({
      en: 'Price',
      uk: 'Ціна',
    }),
    pricePlaceholder: t({
      en: '0.00',
      uk: '0.00',
    }),
    currencyLabel: t({
      en: 'Currency',
      uk: 'Валюта',
    }),
    acquiredAtLabel: t({
      en: 'Acquired At',
      uk: 'Дата придбання',
    }),
    nameLabel: t({
      en: 'Name',
      uk: 'Назва',
    }),
    namePlaceholder: t({
      en: 'Asset name',
      uk: 'Назва активу',
    }),
    descriptionLabel: t({
      en: 'Description',
      uk: 'Опис',
    }),
    descriptionPlaceholder: t({
      en: 'Asset description...',
      uk: 'Опис активу...',
    }),
    cancelButton: t({
      en: 'Cancel',
      uk: 'Скасувати',
    }),
    submitButton: t({
      en: 'Submit',
      uk: 'Зберегти',
    }),
    validationRequiredFields: t({
      en: 'Type, Price, Name and Acquired Date are required',
      uk: 'Тип, Ціна, Назва та Дата придбання є обовʼязковими',
    }),
    assetAddedSuccess: t({
      en: 'Asset added successfully',
      uk: 'Актив успішно додано',
    }),
    assetRemovedSuccess: t({
      en: 'Asset removed successfully',
      uk: 'Актив успішно видалено',
    }),
    assetUpdatedSuccess: t({
      en: 'Asset updated successfully',
      uk: 'Актив успішно оновлено',
    }),
    failedToAdd: t({
      en: 'Failed to add asset',
      uk: 'Не вдалося додати актив',
    }),
    failedToUpdate: t({
      en: 'Failed to update asset',
      uk: 'Не вдалося оновити актив',
    }),
    failedToRemove: t({
      en: 'Failed to remove asset',
      uk: 'Не вдалося видалити актив',
    }),
    failedToLoadDetails: t({
      en: 'Failed to load asset details',
      uk: 'Не вдалося завантажити деталі активу',
    }),
  },
};

export default assetsContent;
