import { Locales } from 'intlayer';

/** @type {import('intlayer').IntlayerConfig} */
const config = {
  internationalization: {
    locales: [
      Locales.ENGLISH,
      Locales.UKRAINIAN,
    ],
    defaultLocale: Locales.ENGLISH,
  },
  routing: {
    mode: 'prefix-no-default',
    enableProxy: false,
  },
};

export default config;
