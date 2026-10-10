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
    taggedInvestmentsCard: t({
      en: 'Tagged Allocations',
      uk: 'Теговані інвестиції',
    }),
    topTagCard: t({
      en: 'Most Used Tag',
      uk: 'Найчастіший тег',
    }),
    totalValuationCard: t({
      en: 'Total Tagged Value',
      uk: 'Загальна сума за тегами',
    }),
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
