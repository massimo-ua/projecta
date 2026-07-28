import { t } from 'intlayer';

const paymentsContent = {
  key: 'payments',
  content: {
    addPayment: t({
      en: 'Add Payment',
      uk: 'Додати платіж',
    }),
    addPaymentTitle: t({
      en: 'Add Payment',
      uk: 'Додати платіж',
    }),
    editPaymentTitle: t({
      en: 'Edit Payment',
      uk: 'Редагувати платіж',
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
    kindLabel: t({
      en: 'Kind',
      uk: 'Вид',
    }),
    kindDownPayment: t({
      en: 'Down Payment',
      uk: 'Аванс',
    }),
    kindCreditPayment: t({
      en: 'Credit Payment',
      uk: 'Оплата в кредит',
    }),
    kindUponCompletion: t({
      en: 'Upon Completion',
      uk: 'По завершенню',
    }),
    dateLabel: t({
      en: 'Date',
      uk: 'Дата',
    }),
    amountLabel: t({
      en: 'Amount',
      uk: 'Сума',
    }),
    amountPlaceholder: t({
      en: '0.00',
      uk: '0.00',
    }),
    currencyLabel: t({
      en: 'Currency',
      uk: 'Валюта',
    }),
    descriptionLabel: t({
      en: 'Description',
      uk: 'Опис',
    }),
    descriptionPlaceholder: t({
      en: 'Payment description...',
      uk: 'Опис платежу...',
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
      en: 'Type, Amount, and Payment Date are required',
      uk: 'Тип, Сума та Дата платежу є обовʼязковими',
    }),
    paymentAddedSuccess: t({
      en: 'Payment added successfully',
      uk: 'Платіж успішно додано',
    }),
    paymentRemovedSuccess: t({
      en: 'Payment removed successfully',
      uk: 'Платіж успішно видалено',
    }),
    paymentUpdatedSuccess: t({
      en: 'Payment updated successfully',
      uk: 'Платіж успішно оновлено',
    }),
    failedToAdd: t({
      en: 'Failed to add payment',
      uk: 'Не вдалося додати платіж',
    }),
    failedToUpdate: t({
      en: 'Failed to update payment',
      uk: 'Не вдалося оновити платіж',
    }),
    failedToRemove: t({
      en: 'Failed to remove payment',
      uk: 'Не вдалося видалити платіж',
    }),
    failedToLoadDetails: t({
      en: 'Failed to load payment details',
      uk: 'Не вдалося завантажити деталі платежу',
    }),
  },
};

export default paymentsContent;
