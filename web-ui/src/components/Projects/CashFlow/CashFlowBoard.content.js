import { t } from 'intlayer';

const cashFlowContent = {
  key: 'cash-flow-board',
  content: {
    title: t({
      en: 'True Cost & Cash Flow Planning',
      uk: 'Собівартість та план грошових потоків',
    }),
    subtitle: t({
      en: 'Track multi-currency actuals (cash, time, goods), spread remaining target budgets across calendar months, and spot TagBudget bottlenecks early.',
      uk: 'Відстежуйте мультивалютні витрати (гроші, час, ресурси), розподіляйте залишок цільових бюджетів по місяцях та виявляйте перевантаження бюджету.',
    }),
    stackOngoingLabel: t({
      en: 'Stack ongoing burn on ceiling chart',
      uk: 'Накласти постійні витрати на ліміт',
    }),
    refreshLabel: t({
      en: 'Refresh',
      uk: 'Оновити',
    }),
    tagFilterPlaceholder: t({
      en: 'Filter by tag (e.g. home)',
      uk: 'Фільтр за тегом (напр. home)',
    }),
    totalTargetLabel: t({
      en: 'Target Budgets',
      uk: 'Цільові бюджети',
    }),
    actualCashLabel: t({
      en: 'Actual Cash Spent',
      uk: 'Фактичні гроші',
    }),
    remainingScheduledLabel: t({
      en: 'Scheduled Remaining',
      uk: 'Запланований залишок',
    }),
    ongoingBurnLabel: t({
      en: 'Ongoing Burn / mo',
      uk: 'Поточні витрати / міс',
    }),
    nonCashLabel: t({
      en: 'Time & Goods',
      uk: 'Час та ресурси',
    }),
    peakCommitmentLabel: t({
      en: 'Peak Cash Month',
      uk: 'Піковий місяць',
    }),
    chartTitle: t({
      en: 'Monthly Cash Demand vs. TagBudget Ceiling',
      uk: 'Місячний попит на кошти vs Ліміт бюджету',
    }),
    chartSubtitle: t({
      en: 'Actual cash combined with linearly spread target completion budgets compared against monthly MONEY TagBudget ceilings.',
      uk: 'Фактичні гроші разом із розподіленими залишками бюджетів відносно місячних лімітів TagBudget.',
    }),
    tableTitle: t({
      en: 'Asset Schedules & True-Cost Breakdown',
      uk: 'Графіки активів та структура витрат',
    }),
    tableSubtitle: t({
      en: 'Adjust start or target completion dates inline to spread capital commitments across months.',
      uk: 'Змінюйте дати початку або планового завершення прямо в таблиці для перерозподілу навантаження.',
    }),
    emptyTitle: t({
      en: 'No assets to plan yet',
      uk: 'Поки немає активів для планування',
    }),
    emptySubtitle: t({
      en: 'Create assets with optional budgets or target completion dates in the Assets tab to populate cash flow projections.',
      uk: 'Створіть активи (з бюджетом або безстрокові) у вкладці Активи, щоб побачити прогноз.',
    }),
  },
};

export default cashFlowContent;
