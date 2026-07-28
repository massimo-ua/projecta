import { t } from 'intlayer';

const errorBoundaryContent = {
  key: 'error-boundary',
  content: {
    title: t({
      en: 'Something went wrong',
      uk: 'Щось пішло не так',
    }),
    description: t({
      en: 'An unexpected error has occurred. Please try again or return to the home page.',
      uk: 'Сталася неочікувана помилка. Будь ласка, спробуйте ще раз або поверніться на головну сторінку.',
    }),
    tryAgain: t({
      en: 'Try again',
      uk: 'Спробувати знову',
    }),
    goHome: t({
      en: 'Go to Home',
      uk: 'На головну',
    }),
    reloadPage: t({
      en: 'Reload page',
      uk: 'Оновити сторінку',
    }),
    detailsTitle: t({
      en: 'Error Details',
      uk: 'Деталі помилки',
    }),
  },
};

export default errorBoundaryContent;
