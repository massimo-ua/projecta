import { t } from 'intlayer';

const totalContent = {
  key: 'total',
  content: {
    summaryTitle: t({
      en: 'Project Summary & Totals',
      uk: 'Підсумок та загальні показники проєкту',
    }),
    analyticsTitle: t({
      en: 'Project Analytics & Charts',
      uk: 'Аналітика та діаграми проєкту',
    }),
    timeframe: {
      label: t({
        en: 'Timeframe',
        uk: 'Період',
      }),
      all: t({
        en: 'All Time',
        uk: 'За весь час',
      }),
      year: t({
        en: 'This Year',
        uk: 'Цей рік',
      }),
      last12m: t({
        en: 'Last 12 Months',
        uk: 'Останні 12 місяців',
      }),
      last6m: t({
        en: 'Last 6 Months',
        uk: 'Останні 6 місяців',
      }),
    },
    charts: {
      projectedSchedule: {
        title: t({
          en: 'Projected vs. Real Annual Investment Schedule',
          uk: 'Річний графік: Прогнозовані та фактичні інвестиції',
        }),
        description: t({
          en: '12-month comparison of planned tag budgets against real investments',
          uk: 'Помісячне порівняння планових бюджетів за тегами з фактичними інвестиціями',
        }),
        monthlyMode: t({ en: 'Monthly Schedule', uk: 'За місяцями' }),
        cumulativeMode: t({ en: 'Cumulative Trajectory', uk: 'Накопичувально' }),
        byTagMode: t({ en: 'By Category (Tag)', uk: 'За категоріями (тегами)' }),
        allCategoriesFilter: t({ en: 'All Categories', uk: 'Усі категорії' }),
        tagFilterPlaceholder: t({ en: 'Tag', uk: 'Тег' }),
        showTableButton: t({ en: 'Table', uk: 'Таблиця' }),
        hideTableButton: t({ en: 'Hide Table', uk: 'Приховати' }),
        unplannedTagLabel: t({ en: 'Unplanned / Untagged', uk: 'Позапланові / Без бюджету' }),
        plannedAnnualTitle: t({ en: 'Planned Annual Budget', uk: 'Плановий річний бюджет' }),
        realInvestedTitle: t({ en: 'Real Invested', uk: 'Фактично інвестовано' }),
        remainingBudgetTitle: t({ en: 'Remaining Budget', uk: 'Залишок бюджету' }),
        overBudgetTitle: t({ en: 'Over Budget', uk: 'Перевитрата бюджету' }),
        adherenceTitle: t({ en: 'Budget Utilization', uk: 'Використання бюджету' }),
        projectedSeriesLabel: t({ en: 'Projected', uk: 'План' }),
        actualSeriesLabel: t({ en: 'Real Invested', uk: 'Факт' }),
        monthColumnTitle: t({ en: 'Month', uk: 'Місяць' }),
        varianceColumnTitle: t({ en: 'Variance', uk: 'Різниця' }),
        noBudgetEmptyTitle: t({
          en: 'No Category Budgets Configured Yet',
          uk: 'Бюджети за категоріями ще не налаштовані',
        }),
        noBudgetEmptyDesc: t({
          en: 'Define expected investment frequencies and amounts per tag to compare your projected annual schedule with real investments.',
          uk: 'Визначте очікувану періодичність та обсяг інвестицій для тегів, щоб порівнювати річний план із фактичними витратами.',
        }),
        configureInTagsButton: t({
          en: 'Configure Category Budgets',
          uk: 'Налаштувати бюджети категорій',
        }),
      },
      monthlyTrend: {
        title: t({
          en: 'Monthly Investment Trend',
          uk: 'Динаміка інвестицій за місяцями',
        }),
        description: t({
          en: 'Investment timeline in project currency',
          uk: 'Хронологія інвестицій у валюті проєкту',
        }),
        monthlyMode: t({
          en: 'Monthly',
          uk: 'Помісячно',
        }),
        cumulativeMode: t({
          en: 'Cumulative',
          uk: 'Накопичувально',
        }),
        avgBurnLine: t({
          en: 'Avg. Monthly',
          uk: 'Сер. за місяць',
        }),
        targetBudgetLine: t({
          en: 'Target Budget',
          uk: 'Цільовий бюджет',
        }),
      },
      byCategory: {
        title: t({
          en: 'Investments by Tag',
          uk: 'Інвестиції за тегами',
        }),
        description: t({
          en: 'Breakdown across resource tags',
          uk: 'Розподіл за тегами ресурсів',
        }),
      },
      paymentKinds: {
        title: t({
          en: 'Resource & Team Breakdown',
          uk: 'Розподіл за ресурсами та учасниками',
        }),
        description: t({
          en: 'Capital, labor hours, goods, and contributor shares',
          uk: 'Капітал, години праці, матеріали та внесок учасників',
        }),
        byResource: t({
          en: 'By Resource',
          uk: 'За ресурсами',
        }),
        byContributor: t({
          en: 'By Contributor',
          uk: 'За учасниками',
        }),
      },
      topTypes: {
        title: t({
          en: 'Top Assets by Investment',
          uk: 'Основні активи за інвестиціями',
        }),
        description: t({
          en: 'Assets with highest resource allocation',
          uk: 'Активи з найбільшим обсягом інвестицій',
        }),
        rootGroups: t({
          en: 'Root Groups',
          uk: 'Групи активів',
        }),
        directAssets: t({
          en: 'Direct Assets',
          uk: 'Окремі активи',
        }),
      },
      paymentsVsAssets: {
        title: t({
          en: 'Budget vs. Actual',
          uk: 'Бюджет проти факту',
        }),
        description: t({
          en: 'Planned target budgets compared to actual investments',
          uk: 'Плановий цільовий бюджет порівняно з фактичними інвестиціями',
        }),
        fallbackDescription: t({
          en: 'Active vs. completed asset value allocation',
          uk: 'Розподіл вартості між активними та завершеними активами',
        }),
        targetLabel: t({
          en: 'Target Budget',
          uk: 'Цільовий бюджет',
        }),
        actualLabel: t({
          en: 'Actual Invested',
          uk: 'Фактично інвестовано',
        }),
        remainingLabel: t({
          en: 'Remaining Budget',
          uk: 'Залишок бюджету',
        }),
        overBudgetLabel: t({
          en: 'Over Budget',
          uk: 'Перевищення бюджету',
        }),
      },
    },
    cards: {
      totalPayments: t({
        en: 'Total Invested',
        uk: 'Всього інвестовано',
      }),
      projectBalance: t({
        en: 'Project Balance',
        uk: 'Баланс проекту',
      }),
      targetBudget: t({
        en: 'Target Budget',
        uk: 'Цільовий бюджет',
      }),
      assetsProgress: t({
        en: 'Assets Progress',
        uk: 'Прогрес активів',
      }),
      avgMonthlyBurn: t({
        en: 'Avg. Monthly Burn',
        uk: 'Середні витрати / міс.',
      }),
      investmentsCount: t({
        en: 'investments',
        uk: 'інвестицій',
      }),
      laborHours: t({
        en: 'hrs labor',
        uk: 'год праці',
      }),
      utilized: t({
        en: 'utilized',
        uk: 'використано',
      }),
      noBudgetSet: t({
        en: 'No target budgets set',
        uk: 'Цільовий бюджет не задано',
      }),
      overBudgetAssets: t({
        en: 'over budget',
        uk: 'понад бюджет',
      }),
      completed: t({
        en: 'completed',
        uk: 'завершено',
      }),
      active: t({
        en: 'active',
        uk: 'активних',
      }),
      activeMonths: t({
        en: 'active months',
        uk: 'активних міс.',
      }),
    },
    kinds: {
      money: t({
        en: 'Capital (Money)',
        uk: 'Капітал (Гроші)',
      }),
      time: t({
        en: 'Labor (Time)',
        uk: 'Праця (Час)',
      }),
      goods: t({
        en: 'Goods & Materials',
        uk: 'Товари та матеріали',
      }),
      downPayment: t({
        en: 'Down Payment',
        uk: 'Авансовий платіж',
      }),
      uponCompletion: t({
        en: 'Direct Investment',
        uk: 'Пряма інвестиція',
      }),
      credit: t({
        en: 'Credit',
        uk: 'Кредитний платіж',
      }),
      other: t({
        en: 'Other',
        uk: 'Інше',
      }),
    },
    labels: {
      payments: t({
        en: 'Total Invested',
        uk: 'Всього інвестовано',
      }),
      balance: t({
        en: 'Project Balance',
        uk: 'Баланс проекту',
      }),
      assets: t({
        en: 'Total Assets',
        uk: 'Всього активів',
      }),
      hoursUnit: t({
        en: 'hrs',
        uk: 'год',
      }),
      unassignedContributor: t({
        en: 'Project Owner',
        uk: 'Власник проєкту',
      }),
      other: t({
        en: 'Other',
        uk: 'Інше',
      }),
      uncategorized: t({
        en: 'General',
        uk: 'Загальні',
      }),
      amount: t({
        en: 'Amount',
        uk: 'Сума',
      }),
      share: t({
        en: 'Share',
        uk: 'Частка',
      }),
      noData: t({
        en: 'No chart data available for this timeframe',
        uk: 'Немає даних для діаграм за обраний період',
      }),
      noDataDesc: t({
        en: 'Record investments or select another timeframe to see analytics.',
        uk: 'Додайте інвестиції або оберіть інший період для відображення графіків.',
      }),
    },
  },
};

export default totalContent;

