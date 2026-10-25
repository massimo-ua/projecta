import { t } from 'intlayer';

const tagsContent = {
  key: 'tags',
  content: {
    title: t({
      en: 'Project Tags',
      uk: 'Теги проєкту',
    }),
    subtitle: t({
      en: 'Track, filter, and analyze investments categorized by tags',
      uk: 'Відстежуйте, фільтруйте та аналізуйте інвестиції за тегами',
    }),
    searchPlaceholder: t({
      en: 'Search tags by name...',
      uk: 'Пошук тегів за назвою...',
    }),
    totalTagsCard: t({
      en: 'Total Tags',
      uk: 'Всього тегів',
    }),
    totalTagsSubtitle: t({
      en: 'Distinct taxonomy labels',
      uk: 'Унікальні категорії проєкту',
    }),
    taggedInvestmentsCard: t({
      en: 'Tagged Allocations',
      uk: 'Теговані інвестиції',
    }),
    taggedInvestmentsSubtitle: t({
      en: 'Resource allocations',
      uk: 'Розподілені вкладення',
    }),
    topTagCard: t({
      en: 'Most Used Tag',
      uk: 'Найчастіший тег',
    }),
    totalValuationCard: t({
      en: 'Total Tagged Value',
      uk: 'Загальна сума за тегами',
    }),
    totalValuationSubtitle: t({
      en: 'Across all tags',
      uk: 'За всіма тегами',
    }),
    noTagsLabel: t({
      en: 'No tags',
      uk: 'Немає тегів',
    }),
    capitalBadge: t({
      en: 'Capital',
      uk: 'Капітал',
    }),
    laborBadge: t({
      en: 'Labor',
      uk: 'Праця',
    }),
    goodsBadge: t({
      en: 'Goods',
      uk: 'Матеріали',
    }),
    refreshButton: t({
      en: 'Refresh',
      uk: 'Оновити',
    }),
    configureBudgetButton: t({
      en: 'Configure Tag Budget',
      uk: 'Налаштувати бюджет тегу',
    }),
    editBudgetButton: t({
      en: 'Edit Budget',
      uk: 'Змінити бюджет',
    }),
    removeBudgetButton: t({
      en: 'Remove Budget',
      uk: 'Видалити бюджет',
    }),
    plannedAnnualTitle: t({
      en: 'Planned Annual',
      uk: 'Річний план',
    }),
    plannedPeriodTitle: t({
      en: 'Planned for Period',
      uk: 'План на період',
    }),
    monthOutsidePeriodWarning: t({
      en: 'This month is outside the active project period (planned total will be 0)',
      uk: 'Цей місяць за межами активного періоду проєкту (плановий підсумок буде 0)',
    }),
    outsidePeriodBadge: t({
      en: 'outside period',
      uk: 'поза періодом',
    }),
    actualInvestedTitle: t({
      en: 'Actual Invested',
      uk: 'Фактично',
    }),
    overbudgetBadge: t({
      en: 'Overbudget',
      uk: 'Перевитрата',
    }),
    budgetModalTitle: t({
      en: 'Configure Category (Tag) Budget',
      uk: 'Налаштування бюджету категорії (тегу)',
    }),
    budgetModalSubtitle: t({
      en: 'Set expected frequency and resource allocation for this category',
      uk: 'Встановіть очікувану періодичність та обсяг вкладень для цієї категорії',
    }),
    targetMonthLabel: t({
      en: 'Scheduled Month (1–12)',
      uk: 'Плановий місяць (1–12)',
    }),
    months: {
      1: t({ en: 'January', uk: 'Січень' }),
      2: t({ en: 'February', uk: 'Лютий' }),
      3: t({ en: 'March', uk: 'Березень' }),
      4: t({ en: 'April', uk: 'Квітень' }),
      5: t({ en: 'May', uk: 'Травень' }),
      6: t({ en: 'June', uk: 'Червень' }),
      7: t({ en: 'July', uk: 'Липень' }),
      8: t({ en: 'August', uk: 'Серпень' }),
      9: t({ en: 'September', uk: 'Вересень' }),
      10: t({ en: 'October', uk: 'Жовтень' }),
      11: t({ en: 'November', uk: 'Листопад' }),
      12: t({ en: 'December', uk: 'Грудень' }),
    },
    frequencyLabel: t({
      en: 'Investment Frequency',
      uk: 'Періодичність інвестицій',
    }),
    resourceTypeLabel: t({
      en: 'Resource Type',
      uk: 'Тип ресурсу',
    }),
    amountLabel: t({
      en: 'Expected Amount',
      uk: 'Очікувана сума',
    }),
    hoursLabel: t({
      en: 'Expected Hours',
      uk: 'Очікувані години',
    }),
    hourlyRateLabel: t({
      en: 'Hourly Rate',
      uk: 'Погодинна ставка',
    }),
    goodsNameLabel: t({
      en: 'Item / Material Name',
      uk: 'Назва товару / матеріалу',
    }),
    goodsQtyLabel: t({
      en: 'Quantity',
      uk: 'Кількість',
    }),
    goodsUnitLabel: t({
      en: 'Unit',
      uk: 'Одиниця виміру',
    }),
    goodsUnitPlaceholder: t({
      en: 'pcs, kg, etc.',
      uk: 'шт, кг тощо',
    }),
    goodsItemNamePlaceholder: t({
      en: 'Material / goods',
      uk: 'Матеріали / товари',
    }),
    tagNameLabel: t({
      en: 'Category / Tag Name',
      uk: 'Назва категорії / тегу',
    }),
    tagPlaceholder: t({
      en: 'materials, equipment, etc.',
      uk: 'матеріали, обладнання тощо',
    }),
    monthLabel: t({
      en: 'Month',
      uk: 'Місяць',
    }),
    perHourSuffix: t({
      en: '/h',
      uk: '/год',
    }),
    perYearSuffix: t({
      en: '/yr',
      uk: '/рік',
    }),
    saveBudgetButton: t({
      en: 'Save Budget Rule',
      uk: 'Зберегти правило бюджету',
    }),
    cancelButton: t({
      en: 'Cancel',
      uk: 'Скасувати',
    }),
    budgetSavedSuccess: t({
      en: 'Category budget updated',
      uk: 'Бюджет категорії оновлено',
    }),
    budgetSaveError: t({
      en: 'Failed to update category budget',
      uk: 'Не вдалося оновити бюджет категорії',
    }),
    frequencies: {
      ONCE: t({ en: 'Once', uk: 'Одноразово' }),
      DAILY: t({ en: 'Daily', uk: 'Щодня' }),
      WEEKLY: t({ en: 'Weekly', uk: 'Щотижня' }),
      MONTHLY: t({ en: 'Monthly', uk: 'Щомісяця' }),
      ANNUALLY: t({ en: 'Annually', uk: 'Щорічно' }),
    },
    resources: {
      MONEY: t({ en: 'Capital (Money)', uk: 'Капітал (Гроші)' }),
      TIME: t({ en: 'Labor (Time)', uk: 'Праця (Час)' }),
      GOODS: t({ en: 'Goods & Materials', uk: 'Товари та матеріали' }),
    },
    sortLabel: t({
      en: 'Sort by:',
      uk: 'Сортувати:',
    }),
    sortByCount: t({
      en: 'Most Used',
      uk: 'За кількістю',
    }),
    sortByAmount: t({
      en: 'Highest Amount',
      uk: 'За сумою',
    }),
    sortByName: t({
      en: 'Name (A-Z)',
      uk: 'За назвою (А-Я)',
    }),
    viewInvestments: t({
      en: 'View Investments',
      uk: 'Переглянути інвестиції',
    }),
    investmentsCount: t({
      en: 'investments',
      uk: 'інвестицій',
    }),
    singleInvestment: t({
      en: 'investment',
      uk: 'інвестиція',
    }),
    resourcesUsed: t({
      en: 'Resources:',
      uk: 'Ресурси:',
    }),
    noTagsTitle: t({
      en: 'No tags in this project yet',
      uk: 'У цьому проєкті ще немає тегів',
    }),
    noTagsDescription: t({
      en: 'Tags are added when creating or importing investments (e.g. #labor, #maintenance, #down-payment).',
      uk: 'Теги додаються під час фіксації або імпорту інвестицій (наприклад #праця, #обслуговування, #аванс).',
    }),
    noSearchMatchTitle: t({
      en: 'No tags match your search',
      uk: 'Жоден тег не відповідає критеріям пошуку',
    }),
    noSearchMatchDescription: t({
      en: 'Try typing a different keyword or clear the search input.',
      uk: 'Спробуйте інше ключове слово або очистіть рядок пошуку.',
    }),
    goToInvestments: t({
      en: 'Go to Investments',
      uk: 'Перейти до інвестицій',
    }),
  },
};

export default tagsContent;
