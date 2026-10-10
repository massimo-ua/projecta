import { t } from 'intlayer';

const paymentsContent = {
  key: 'payments',
  content: {
    addPayment: t({
      en: 'Log Investment',
      uk: 'Зафіксувати інвестицію',
    }),
    addPaymentTitle: t({
      en: 'Log Investment',
      uk: 'Зафіксувати інвестицію',
    }),
    editPaymentTitle: t({
      en: 'Edit Investment',
      uk: 'Редагувати інвестицію',
    }),
    investmentsTitle: t({
      en: 'Investments',
      uk: 'Інвестиції',
    }),
    resourceTypeLabel: t({
      en: 'Resource Type',
      uk: 'Тип ресурсу',
    }),
    moneyResource: t({
      en: 'Capital (Money)',
      uk: 'Капітал (Гроші)',
    }),
    timeResource: t({
      en: 'Labor (Time)',
      uk: 'Праця (Час)',
    }),
    goodsResource: t({
      en: 'Goods & Materials',
      uk: 'Товари та матеріали',
    }),
    assetTargetLabel: t({
      en: 'Target Asset',
      uk: 'Цільовий актив',
    }),
    selectAssetPlaceholder: t({
      en: 'Select asset',
      uk: 'Оберіть актив',
    }),
    timeHoursLabel: t({
      en: 'Hours Worked',
      uk: 'Відпрацьовані години',
    }),
    timeHourlyRateLabel: t({
      en: 'Hourly Rate',
      uk: 'Погодинна ставка',
    }),
    goodsQuantityLabel: t({
      en: 'Quantity',
      uk: 'Кількість',
    }),
    goodsUnitLabel: t({
      en: 'Unit (e.g. pcs, kg, m)',
      uk: 'Одиниця (напр. шт, кг, м)',
    }),
    goodsItemNameLabel: t({
      en: 'Item / Material Name',
      uk: 'Назва товару / матеріалу',
    }),
    tagsLabel: t({
      en: 'Tags',
      uk: 'Теги',
    }),
    tagsPlaceholder: t({
      en: 'Type tag and press Enter',
      uk: 'Введіть тег та натисніть Enter',
    }),
    filterAllAssets: t({
      en: 'All Assets',
      uk: 'Усі активи',
    }),
    filterAllResourceTypes: t({
      en: 'All Resources',
      uk: 'Усі ресурси',
    }),
    filterTagPlaceholder: t({
      en: 'Filter by tag...',
      uk: 'Фільтр за тегом...',
    }),
    typeLabel: t({
      en: 'Asset',
      uk: 'Актив',
    }),
    selectTypePlaceholder: t({
      en: 'Select asset',
      uk: 'Оберіть актив',
    }),
    categoryLabel: t({
      en: 'Tags / Resource',
      uk: 'Теги / Ресурс',
    }),
    kindLabel: t({
      en: 'Resource',
      uk: 'Ресурс',
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
      en: 'Valuation Amount',
      uk: 'Грошова оцінка',
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
      en: 'Investment description, details...',
      uk: 'Опис інвестиції, деталі...',
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
      en: 'Asset, Amount, and Date are required',
      uk: 'Актив, Сума та Дата є обовʼязковими',
    }),
    paymentAddedSuccess: t({
      en: 'Investment logged successfully',
      uk: 'Інвестицію успішно зафіксовано',
    }),
    paymentRemovedSuccess: t({
      en: 'Investment removed successfully',
      uk: 'Інвестицію видалено',
    }),
    paymentUpdatedSuccess: t({
      en: 'Investment updated successfully',
      uk: 'Інвестицію оновлено',
    }),
    failedToAdd: t({
      en: 'Failed to log investment',
      uk: 'Не вдалося зафіксувати інвестицію',
    }),
    failedToUpdate: t({
      en: 'Failed to update investment',
      uk: 'Не вдалося оновити інвестицію',
    }),
    failedToRemove: t({
      en: 'Failed to remove investment',
      uk: 'Не вдалося видалити інвестицію',
    }),
    failedToLoadDetails: t({
      en: 'Failed to load details',
      uk: 'Не вдалося завантажити деталі',
    }),
    importStatement: t({
      en: 'Import Statement',
      uk: 'Імпортувати виписку',
    }),
    importStatementTitle: t({
      en: 'Import Statement (Kredobank)',
      uk: 'Імпорт виписки (Кредобанк)',
    }),
    importStatementDesc: t({
      en: 'Upload a Kredobank PDF account statement to extract and import payments as investments.',
      uk: 'Завантажте PDF-виписку Кредобанку для вилучення та імпорту витрат як інвестицій.',
    }),
    dropFilePrompt: t({
      en: 'Choose a PDF statement or drag & drop here',
      uk: 'Оберіть PDF-виписку або перетягніть її сюди',
    }),
    uploading: t({
      en: 'Parsing PDF statement...',
      uk: 'Обробка PDF-виписки...',
    }),
    defaultCostType: t({
      en: 'Default Target Asset',
      uk: 'Цільовий актив за замовчуванням',
    }),
    applyToAll: t({
      en: 'Apply to all',
      uk: 'Застосувати до всіх',
    }),
    potentialDuplicate: t({
      en: 'Potential duplicate',
      uk: 'Можливий дублікат',
    }),
    importPaymentsCount: t({
      en: 'Import Investments',
      uk: 'Імпортувати інвестиції',
    }),
    selectedCount: t({
      en: 'selected',
      uk: 'обрано',
    }),
    paymentsImportedSuccess: t({
      en: 'Investments imported successfully',
      uk: 'Інвестиції успішно імпортовано',
    }),
    failedToParseStatement: t({
      en: 'Failed to parse statement',
      uk: 'Не вдалося обробити виписку',
    }),
    failedToImportPayments: t({
      en: 'Failed to import investments',
      uk: 'Не вдалося імпортувати інвестиції',
    }),
    accountLabel: t({
      en: 'Account',
      uk: 'Рахунок',
    }),
    periodLabel: t({
      en: 'Period',
      uk: 'Період',
    }),
    selectAnotherFile: t({
      en: 'Change file',
      uk: 'Змінити файл',
    }),
    selectAll: t({
      en: 'Select all',
      uk: 'Вибрати всі',
    }),
    selectTypeForSelectedWarning: t({
      en: 'Please select a Target Asset for all checked items',
      uk: 'Будь ласка, оберіть цільовий актив для всіх вибраних записів',
    }),
    filterAllTypes: t({
      en: 'All Assets',
      uk: 'Усі активи',
    }),
    filterDateFrom: t({
      en: 'From date',
      uk: 'Дата з',
    }),
    filterDateTo: t({
      en: 'To date',
      uk: 'Дата по',
    }),
    resetFilters: t({
      en: 'Reset filters',
      uk: 'Скинути фільтри',
    }),
    selectedPayments: t({
      en: 'investments selected',
      uk: 'інвестицій вибрано',
    }),
    createAssetFromSelected: t({
      en: 'Create Asset from Selected',
      uk: 'Створити актив з вибраних',
    }),
    clearSelection: t({
      en: 'Clear selection',
      uk: 'Очистити вибір',
    }),
    createAssetTitle: t({
      en: 'Create Asset from Investments',
      uk: 'Створити актив з інвестицій',
    }),
    createAssetDesc: t({
      en: 'Summarize the selected investments into a new asset record.',
      uk: 'Обʼєднайте вибрані інвестиції у новий актив.',
    }),
    assetNameLabel: t({
      en: 'Asset Name',
      uk: 'Назва активу',
    }),
    assetNamePlaceholder: t({
      en: 'Enter asset name...',
      uk: 'Введіть назву активу...',
    }),
    targetCurrencyLabel: t({
      en: 'Resulting Currency',
      uk: 'Підсумкова валюта',
    }),
    estimatedTotalLabel: t({
      en: 'Total Price',
      uk: 'Підсумкова ціна',
    }),
    currencyConversionNotice: t({
      en: 'Investments in differing currencies will be converted into',
      uk: 'Інвестиції в інших валютах буде сконвертовано у',
    }),
    assetCreatedSuccess: t({
      en: 'Asset created successfully from selected investments',
      uk: 'Актив успішно створено з вибраних інвестицій',
    }),
    viewInAssets: t({
      en: 'View in Assets',
      uk: 'Переглянути в активах',
    }),
    failedToCreateAsset: t({
      en: 'Failed to create asset',
      uk: 'Не вдалося створити актив',
    }),
  },
};

export default paymentsContent;
