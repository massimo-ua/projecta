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
      monthlyTrend: {
        title: t({
          en: 'Monthly Spending Trend',
          uk: 'Динаміка витрат за місяцями',
        }),
        description: t({
          en: 'Spending timeline in project currency',
          uk: 'Хронологія витрат у валюті проєкту',
        }),
      },
      byCategory: {
        title: t({
          en: 'Expenses by Category',
          uk: 'Витрати за категоріями',
        }),
        description: t({
          en: 'Payment breakdown across categories',
          uk: 'Розподіл платежів за категоріями',
        }),
      },
      paymentKinds: {
        title: t({
          en: 'Payment Kinds',
          uk: 'Види платежів',
        }),
        description: t({
          en: 'Down payments vs upon completion vs credit',
          uk: 'Аванси та оплати по завершенню робіт',
        }),
      },
      topTypes: {
        title: t({
          en: 'Top Cost Types',
          uk: 'Основні типи витрат',
        }),
        description: t({
          en: 'Top spending subcategories',
          uk: 'Найбільші статті витрат',
        }),
      },
      paymentsVsAssets: {
        title: t({
          en: 'Payments vs. Assets',
          uk: 'Платежі проти активів',
        }),
        description: t({
          en: 'Comparison of cash spent and asset values',
          uk: 'Порівняння витрачених коштів та вартості активів',
        }),
      },
    },
    kinds: {
      downPayment: t({
        en: 'Down Payment',
        uk: 'Авансовий платіж',
      }),
      uponCompletion: t({
        en: 'Upon Completion',
        uk: 'По завершенню',
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
        en: 'Total Payments',
        uk: 'Всього платежів',
      }),
      assets: t({
        en: 'Total Assets',
        uk: 'Всього активів',
      }),
      other: t({
        en: 'Other',
        uk: 'Інше',
      }),
      uncategorized: t({
        en: 'Uncategorized',
        uk: 'Без категорії',
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
        en: 'Record payments or select another timeframe to see analytics.',
        uk: 'Додайте платежі або оберіть інший період для відображення графіків.',
      }),
    },
  },
};

export default totalContent;
