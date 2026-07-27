import { t } from 'intlayer';

const loginContent = {
  key: 'login',
  content: {
    title: t({
      en: 'Welcome back',
      uk: 'З поверненням',
    }),
    subtitle: t({
      en: 'Sign in to your Projecta account to continue',
      uk: 'Увійдіть у свій акаунт Projecta для продовження',
    }),
    orContinueWith: t({
      en: 'Or continue with',
      uk: 'Або продовжіть за допомогою',
    }),
    usernameLabel: t({
      en: 'Username',
      uk: 'Ім’я користувача',
    }),
    usernamePlaceholder: t({
      en: 'name@example.com',
      uk: 'name@example.com',
    }),
    passwordLabel: t({
      en: 'Password',
      uk: 'Пароль',
    }),
    signInButton: t({
      en: 'Sign In',
      uk: 'Увійти',
    }),
    inputRequiredError: t({
      en: 'Please input your username and password',
      uk: 'Будь ласка, введіть ім’я користувача та пароль',
    }),
    loginSuccess: t({
      en: 'Login successful',
      uk: 'Успішний вхід',
    }),
    loginFailed: t({
      en: 'Login failed',
      uk: 'Помилка входу',
    }),
  },
};

export default loginContent;
