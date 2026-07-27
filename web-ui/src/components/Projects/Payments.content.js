import { t } from 'intlayer';

const paymentsContent = {
  key: 'payments',
  content: {
    addPayment: t({
      en: 'Add Payment',
      uk: 'Додати платіж',
    }),
    typeLabel: t({
      en: 'Type',
      uk: 'Тип',
    }),
    categoryLabel: t({
      en: 'Category',
      uk: 'Категорія',
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
    failedToRemove: t({
      en: 'Failed to remove payment',
      uk: 'Не вдалося видалити платіж',
    }),
  },
};

export default paymentsContent;
