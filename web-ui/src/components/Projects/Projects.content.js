import { t } from 'intlayer';

const projectsContent = {
  key: 'projects',
  content: {
    title: t({
      en: 'Projects',
      uk: 'Проєкти',
    }),
    subtitle: t({
      en: 'Manage and track all your active projects',
      uk: 'Керуйте та відстежуйте усі ваші активні проєкти',
    }),
    createProjectButton: t({
      en: 'Create Project',
      uk: 'Створити проєкт',
    }),
    noProjectsFound: t({
      en: 'No projects found',
      uk: 'Проєктів не знайдено',
    }),
    noProjectsDesc: t({
      en: 'Get started by creating your first project',
      uk: 'Розпочніть, створивши свій перший проєкт',
    }),
    createNewProjectTitle: t({
      en: 'Create New Project',
      uk: 'Створити новий проєкт',
    }),
    projectNameLabel: t({
      en: 'Project Name',
      uk: 'Назва проєкту',
    }),
    projectNamePlaceholder: t({
      en: 'e.g. Apartment Renovation',
      uk: 'напр. Ремонт квартири',
    }),
    descriptionLabel: t({
      en: 'Description',
      uk: 'Опис',
    }),
    descriptionPlaceholder: t({
      en: 'Project goals, scope, and notes...',
      uk: 'Цілі проєкту, обсяг та примітки...',
    }),
    cancelButton: t({
      en: 'Cancel',
      uk: 'Скасувати',
    }),
    nameRequiredError: t({
      en: 'Project name is required (minimum 3 characters)',
      uk: 'Назва проєкту є обов’язковою (мінімум 3 символи)',
    }),
    projectCreatedSuccess: t({
      en: 'Project created successfully',
      uk: 'Проєкт успішно створено',
    }),
    failedToCreateProject: t({
      en: 'Failed to create project',
      uk: 'Не вдалося створити проєкт',
    }),
    viewDetailsLink: t({
      en: 'View details',
      uk: 'Детальніше',
    }),
    noDescription: t({
      en: 'No description provided.',
      uk: 'Опис відсутній.',
    }),
    shareTokenNotAvailable: t({
      en: 'Share token not available',
      uk: 'Токен поширення недоступний',
    }),
    shareLinkCopied: t({
      en: 'Share link copied to clipboard!',
      uk: 'Посилання на проєкт скопійовано в буфер обміну!',
    }),
    failedToCopyShareLink: t({
      en: 'Failed to copy share link',
      uk: 'Не вдалося скопіювати посилання',
    }),
    copyShareLinkTooltip: t({
      en: 'Copy project share link',
      uk: 'Скопіювати посилання на проєкт',
    }),
    sharedTag: t({
      en: 'Shared',
      uk: 'Спільний',
    }),
    sharedBy: t({
      en: 'by',
      uk: 'від',
    }),
    ongoingBadge: t({
      en: 'Ongoing',
      uk: 'Безстроковий',
    }),
    wizard: {
      title: t({
        en: 'Project Setup Wizard',
        uk: 'Майстер налаштування проєкту',
      }),
      subtitle: t({
        en: 'Configure your project currency, investment categories, and projected annual budget',
        uk: 'Налаштуйте валюту проєкту, категорії інвестицій та плановий річний бюджет',
      }),
      steps: {
        basicsTitle: t({ en: '1. Project & Currency', uk: '1. Проєкт та валюта' }),
        categoriesTitle: t({ en: '2. Investment Tags & Budget', uk: '2. Теги інвестицій та бюджет' }),
        scheduleTitle: t({ en: '3. Annual Schedule Preview', uk: '3. Річний графік інвестицій' }),
      },
      homeCurrencyLabel: t({ en: 'Home Currency', uk: 'Основна валюта проєкту' }),
      currencyPlaceholder: t({ en: 'Currency', uk: 'Валюта' }),
      monthLabel: t({ en: 'Month', uk: 'Місяць' }),
      perHourSuffix: t({ en: '/h', uk: '/год' }),
      perYearSuffix: t({ en: '/yr', uk: '/рік' }),
      homeCurrencyHint: t({
        en: 'All planned tag budgets and schedule projections are calculated in this currency.',
        uk: 'Усі планові бюджети за тегами та річний графік розраховуються у цій валюті.',
      }),
      startDateLabel: t({ en: 'Project Start Date', uk: 'Дата початку проєкту' }),
      endDateLabel: t({ en: 'Target End Date', uk: 'Цільова дата завершення' }),
      optionalBadge: t({ en: 'Optional', uk: 'Необов’язково' }),
      clearEndDateButton: t({ en: 'Clear end date (Ongoing project)', uk: 'Очистити дату (Безстроковий проєкт)' }),
      clearButton: t({ en: 'Clear', uk: 'Очистити' }),
      endDateHint: t({
        en: 'Leave empty if this project is ongoing without a deadline (e.g. "Stay healthy"). If set, projections are capped at this goal date.',
        uk: 'Залиште порожнім, якщо проєкт безстроковий без фіксованого терміну (напр. "Здоровий спосіб життя"). Якщо вказано, розрахунки завершуються на цій даті.',
      }),
      endDateSetHint: t({
        en: 'Goal target date active. Investments and budget projections will conclude at this date.',
        uk: 'Цільова дата встановлена. Інвестиції та прогнози бюджету завершаться цією датою.',
      }),
      invalidEndDateError: t({
        en: 'End date cannot be earlier than start date.',
        uk: 'Дата завершення не може бути раніше дати початку.',
      }),
      presetTagsTitle: t({
        en: 'Quick-Add Suggested Categories (Tags)',
        uk: 'Швидке додавання типових категорій (тегів)',
      }),
      tagRuleTitle: t({
        en: 'Define Category Investment Expectation',
        uk: 'Визначте очікувані інвестиції для категорії',
      }),
      tagLabel: t({ en: 'Category (Tag)', uk: 'Категорія (Тег)' }),
      tagPlaceholder: t({ en: 'e.g. labor, materials, rent', uk: 'напр. праця, матеріали, оренда' }),
      frequencyLabel: t({ en: 'Investment Frequency', uk: 'Періодичність інвестицій' }),
      targetMonthLabel: t({ en: 'Scheduled Month (1–12)', uk: 'Плановий місяць (1–12)' }),
      monthOutsidePeriodWarning: t({
        en: 'This month is outside the active project period (planned total will be 0)',
        uk: 'Цей місяць за межами активного періоду проєкту (плановий підсумок буде 0)',
      }),
      outsidePeriodBadge: t({
        en: 'outside period',
        uk: 'поза періодом',
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
      resourceTypeLabel: t({ en: 'Resource Type', uk: 'Тип ресурсу' }),
      amountLabel: t({ en: 'Expected Amount', uk: 'Очікувана сума' }),
      hoursLabel: t({ en: 'Expected Hours', uk: 'Очікувані години' }),
      hourlyRateLabel: t({ en: 'Hourly Rate', uk: 'Погодинна ставка' }),
      goodsNameLabel: t({ en: 'Item / Material Name', uk: 'Назва товару / матеріалу' }),
      goodsItemPlaceholder: t({ en: 'Concrete, cables...', uk: 'Бетон, кабелі...' }),
      goodsQtyLabel: t({ en: 'Quantity', uk: 'Кількість' }),
      goodsUnitLabel: t({ en: 'Unit', uk: 'Одиниця виміру' }),
      goodsUnitPlaceholder: t({ en: 'pcs, kg...', uk: 'шт, кг...' }),
      addCategoryButton: t({ en: 'Add Category to Budget', uk: 'Додати категорію до бюджету' }),
      updateCategoryButton: t({ en: 'Update Category', uk: 'Оновити категорію' }),
      cancelEditCategoryButton: t({ en: 'Cancel Edit', uk: 'Скасувати редагування' }),
      configuredCategoriesTitle: t({
        en: 'Configured Budget Categories',
        uk: 'Налаштовані категорії бюджету',
      }),
      noCategoriesYetTitle: t({
        en: 'No budget categories added yet (Optional)',
        uk: 'Категорії бюджету ще не додані (Необов’язково)',
      }),
      noCategoriesYetDesc: t({
        en: 'Add expected investments by tag above, or skip this step and configure them later in the Tags tab.',
        uk: 'Додайте очікувані інвестиції за тегами вище або пропустіть цей крок і налаштуйте їх пізніше у вкладці Теги.',
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
      preview: {
        totalAnnualBudgetTitle: t({ en: 'Projected Annual Budget', uk: 'Прогнозований річний бюджет' }),
        avgMonthlyPlannedTitle: t({ en: 'Avg. Monthly Projection', uk: 'Середній план на місяць' }),
        categoriesCountTitle: t({ en: 'Planned Categories', uk: 'Заплановано категорій' }),
        monthlyScheduleChartTitle: t({
          en: 'Projected 12-Month Investment Schedule',
          uk: 'Прогнозований 12-місячний графік інвестицій',
        }),
        categoryBreakdownTitle: t({
          en: 'Annual Budget Breakdown by Category',
          uk: 'Розподіл річного бюджету за категоріями',
        }),
        annualTotalLabel: t({ en: 'Annual Total', uk: 'Річна сума' }),
        monthLabel: t({ en: 'Month', uk: 'Місяць' }),
        projectedLabel: t({ en: 'Projected', uk: 'План' }),
      },
      backButton: t({ en: 'Back', uk: 'Назад' }),
      nextButton: t({ en: 'Next Step', uk: 'Далі' }),
      skipStepButton: t({ en: 'Skip & Preview', uk: 'Пропустити' }),
      completeWizardButton: t({
        en: 'Create Project & Save Budget',
        uk: 'Створити проєкт та зберегти бюджет',
      }),
    },
  },
};

export default projectsContent;
