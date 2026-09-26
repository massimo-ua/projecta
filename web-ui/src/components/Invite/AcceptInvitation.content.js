import { t } from 'intlayer';

const acceptInvitationContent = {
  key: 'accept-invitation',
  content: {
    invitedTitle: t({
      en: "You're Invited!",
      uk: 'Вас запрошено!',
    }),
    invitedSubtitle: t({
      en: 'Join Projecta using your invitation link.',
      uk: 'Приєднуйтесь до Projecta за вашим запрошенням.',
    }),
    invitationFor: t({
      en: 'This invitation is for',
      uk: 'Це запрошення призначене для',
    }),
    signInInstructions: t({
      en: 'Sign in with your Google account to complete registration.',
      uk: 'Увійдіть за допомогою свого облікового запису Google, щоб завершити реєстрацію.',
    }),
    loading: t({
      en: 'Verifying invitation...',
      uk: 'Перевірка запрошення...',
    }),
    invalidTitle: t({
      en: 'Invalid or Expired Invitation',
      uk: 'Недійсне або прострочене запрошення',
    }),
    invalidSubtitle: t({
      en: 'This invitation link is invalid, expired, or has already been used.',
      uk: 'Це посилання на запрошення недійсне, прострочене або вже було використане.',
    }),
    loginInstead: t({
      en: 'Go to Login',
      uk: 'Перейти до входу',
    }),
    loginSuccess: t({
      en: 'Welcome to Projecta!',
      uk: 'Ласкаво просимо до Projecta!',
    }),
  },
};

export default acceptInvitationContent;
