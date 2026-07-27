import { t } from 'intlayer';

const assetsContent = {
  key: 'assets',
  content: {
    addAsset: t({
      en: 'Add Asset',
      uk: 'Додати актив',
    }),
    typeLabel: t({
      en: 'Type',
      uk: 'Тип',
    }),
    categoryLabel: t({
      en: 'Category',
      uk: 'Категорія',
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
    failedToRemove: t({
      en: 'Failed to remove asset',
      uk: 'Не вдалося видалити актив',
    }),
  },
};

export default assetsContent;
