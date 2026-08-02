import { t } from 'intlayer';

const acceptShareContent = {
  key: 'accept-share',
  content: {
    joiningProject: t({
      en: 'Joining Project...',
      uk: 'Приєднання до проєкту...',
    }),
    projectShared: t({
      en: 'Project Shared!',
      uk: 'Доступ до проєкту отримано!',
    }),
    unableToJoinProject: t({
      en: 'Unable to Join Project',
      uk: 'Не вдалося приєднатися до проєкту',
    }),
    processingInvitation: t({
      en: 'Processing your invitation to collaborate on this project.',
      uk: 'Обробка вашого запрошення для спільної роботи над цим проєктом.',
    }),
    accessGrantedRedirecting: t({
      en: 'You now have access to this project. Redirecting to project page...',
      uk: 'Тепер ви маєте доступ до цього проєкту. Перенаправляємо на сторінку проєкту...',
    }),
    alreadyOwner: t({
      en: 'You already have access to this project as owner. Redirecting to project page...',
      uk: 'Ви вже маєте доступ до цього проєкту як власник. Перенаправляємо на сторінку проєкту...',
    }),
    alreadyOwnerToast: t({
      en: 'You already have access to this project as owner.',
      uk: 'Ви вже маєте доступ до цього проєкту як власник.',
    }),
    invalidOrExpiredLink: t({
      en: 'The share link may be invalid or expired.',
      uk: 'Посилання для доступу може бути недійсним або застарілим.',
    }),
    noTokenProvided: t({
      en: 'No share token provided.',
      uk: 'Токен доступу не надано.',
    }),
    failedToAcceptShare: t({
      en: 'Failed to accept project share.',
      uk: 'Не вдалося прийняти доступ до проєкту.',
    }),
    projectWord: t({
      en: 'Project',
      uk: 'Проєкт',
    }),
    addedToList: t({
      en: 'added to your list!',
      uk: 'додано до вашого списку!',
    }),
    backToProjectsButton: t({
      en: 'Back to Projects',
      uk: 'Повернутися до проєктів',
    }),
  },
};

export default acceptShareContent;
