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
          en: 'Monthly Investment Trend',
          uk: 'Динаміка інвестицій за місяцями',
        }),
        description: t({
          en: 'Investment timeline in project currency',
          uk: 'Хронологія інвестицій у валюті проєкту',
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
          en: 'Resource Breakdown',
          uk: 'Розподіл за видами ресурсів',
        }),
        description: t({
          en: 'Capital, labor hours, and goods investments',
          uk: 'Капітал, праця (години) та товари й матеріали',
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
      },
      paymentsVsAssets: {
        title: t({
          en: 'Investments vs. Assets',
          uk: 'Інвестиції та активи',
        }),
        description: t({
          en: 'Direct resource allocations versus asset valuations',
          uk: 'Прямі інвестиції проти оцінки активів',
        }),
      },
    },
    kinds: {
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
        en: 'Total Investments',
        uk: 'Всього інвестицій',
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
