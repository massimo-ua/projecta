import { useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { useLocale } from 'react-intlayer';
import { Locales, getHTMLTextDir, getLocalizedUrl } from 'intlayer';

/**
 * Updates the HTML <html> element's `lang` and `dir` attributes based on the current locale
 * and syncs locale state with URL path while preserving user's language choice across routes.
 */
export const useI18nHTMLAttributes = () => {
  const { locale, setLocale } = useLocale();
  const location = useLocation();
  const navigate = useNavigate();

  useEffect(() => {
    const isUkPath = location.pathname === '/uk' || location.pathname.startsWith('/uk/');

    if (isUkPath && locale !== Locales.UKRAINIAN) {
      setLocale(Locales.UKRAINIAN);
    } else if (!isUkPath && locale === Locales.UKRAINIAN) {
      const localizedUrl = getLocalizedUrl(`${location.pathname}${location.search}`, locale);
      if (localizedUrl !== location.pathname) {
        navigate(localizedUrl, { replace: true });
      }
    }
  }, [location.pathname, locale]);

  useEffect(() => {
    if (locale) {
      document.documentElement.lang = locale;
      document.documentElement.dir = getHTMLTextDir(locale);
    }
  }, [locale]);
};

export default useI18nHTMLAttributes;
