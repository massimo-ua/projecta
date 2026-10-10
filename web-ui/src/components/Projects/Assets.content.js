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
    statusLabel: t({
      en: 'Status',
      uk: 'Статус',
    }),
    activeStatus: t({
      en: 'Active',
      uk: 'В процесі',
    }),
    completedStatus: t({
      en: 'Completed',
      uk: 'Завершено',
    }),
    startDateLabel: t({
      en: 'Start Date',
      uk: 'Дата початку',
    }),
    completedDateLabel: t({
      en: 'Completed Date',
      uk: 'Дата завершення',
    }),
    targetPriceLabel: t({
      en: 'Target Price / Budget',
      uk: 'Цільова вартість / Бюджет',
    }),
    directCostLabel: t({
      en: 'Direct Cost',
      uk: 'Прямі витрати',
    }),
    totalCostLabel: t({
      en: 'Total Cost',
      uk: 'Загальна вартість',
    }),
    progressLabel: t({
      en: 'Progress',
      uk: 'Прогрес',
    }),
    linkSubAsset: t({
      en: 'Link Sub-Asset',
      uk: 'Привʼязати під-актив',
    }),
    unlinkSubAsset: t({
      en: 'Unlink',
      uk: 'Відвʼязати',
    }),
    subAssetsLabel: t({
      en: 'Sub-Assets',
      uk: 'Складові активи',
    }),
    parentAssetsLabel: t({
      en: 'Part of',
      uk: 'Входить до',
    }),
    shareLabel: t({
      en: 'Allocation Share',
      uk: 'Частка',
    }),
    noSubAssets: t({
      en: 'No sub-assets',
      uk: 'Немає під-активів',
    }),
    linkModalTitle: t({
      en: 'Link Component Asset',
      uk: 'Привʼязати складовий актив',
    }),
    selectChildAsset: t({
      en: 'Select Child Asset',
      uk: 'Оберіть під-актив',
    }),
    sharePercentageLabel: t({
      en: 'Share (%)',
      uk: 'Частка (%)',
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
      en: 'Log Initial Investment',
      uk: 'Зафіксувати початкову інвестицію',
    }),
    createPaymentDescription: t({
      en: 'Automatically create an investment entry for this asset',
      uk: 'Автоматично створити запис про інвестицію для цього активу',
    }),
    priceLabel: t({
      en: 'Initial Investment Amount',
      uk: 'Сума початкової інвестиції',
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
      en: 'Start Date',
      uk: 'Дата початку',
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
    tagsLabel: t({
      en: 'Tags',
      uk: 'Теги',
    }),
    tagsPlaceholder: t({
      en: 'Type tag and press Enter',
      uk: 'Введіть тег та натисніть Enter',
    }),
    suggestedTags: t({
      en: 'Existing tags:',
      uk: 'Існуючі теги:',
    }),
    cancelButton: t({
      en: 'Cancel',
      uk: 'Скасувати',
    }),
    submitButton: t({
      en: 'Save',
      uk: 'Зберегти',
    }),
    validationRequiredFields: t({
      en: 'Name is required',
      uk: 'Назва є обовʼязковою',
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
    investInAsset: t({
      en: 'Log Investment',
      uk: 'Інвестувати',
    }),
    viewInvestments: t({
      en: 'View Investments',
      uk: 'Переглянути інвестиції',
    }),
    groupAssets: t({
      en: 'Group into New Asset',
      uk: 'Згрупувати в новий актив',
    }),
    addToParentAsset: t({
      en: 'Add to Existing Asset',
      uk: 'Додати до існуючого активу',
    }),
    groupModalTitle: t({
      en: 'Group Assets into New Asset',
      uk: 'Згрупувати активи в новий актив',
    }),
    groupModalDesc: t({
      en: 'The new asset will dynamically aggregate all costs from the selected component assets.',
      uk: 'Новий актив динамічно підсумовуватиме всі витрати з вибраних активів-компонентів.',
    }),
    assignModalTitle: t({
      en: 'Add to Existing Parent Asset',
      uk: 'Додати до існуючого батьківського активу',
    }),
    assignModalDesc: t({
      en: 'The selected assets will be linked as components under the chosen parent asset.',
      uk: 'Вибрані активи будуть підпорядковані як компоненти обраного активу.',
    }),
    assetsGroupedSuccess: t({
      en: 'Assets grouped successfully',
      uk: 'Активи успішно згруповано',
    }),
    failedToGroup: t({
      en: 'Failed to group assets',
      uk: 'Не вдалося згрупувати активи',
    }),
    selectedAssets: t({
      en: 'assets selected',
      uk: 'активів вибрано',
    }),
    selectAll: t({
      en: 'Select all on page',
      uk: 'Вибрати всі на сторінці',
    }),
    clearSelection: t({
      en: 'Clear',
      uk: 'Очистити',
    }),
    componentsSummary: t({
      en: 'Selected Component Assets',
      uk: 'Вибрані активи-компоненти',
    }),
    aggregatedCostEstimate: t({
      en: 'Aggregated Total Cost',
      uk: 'Загальна вартість групи',
    }),
    selectParentAsset: t({
      en: 'Select Parent Asset',
      uk: 'Оберіть батьківський актив',
    }),
  },
};

export default assetsContent;
