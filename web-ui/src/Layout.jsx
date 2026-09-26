import React from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useIntlayer, useLocale } from 'react-intlayer';
import { Locales, getLocalizedUrl } from 'intlayer';
import { User, Sun, Moon, Globe, ShieldCheck } from 'lucide-react';
import PropTypes from 'prop-types';
import { Logo } from './components/Logo';
import Logout from './components/Logout';
import { AppFooter, ErrorBoundary } from './components/index.js';
import { Toaster } from '@/components/ui/sonner';
import { Button } from '@/components/ui/button';
import { authProvider } from './api';
import { useI18nHTMLAttributes } from './hooks/useI18nHTMLAttributes';
import { useTheme } from './hooks/useTheme';
import { useCurrentUser } from './hooks/useCurrentUser';

export default function HomeLayout({ children }) {
  useI18nHTMLAttributes();
  const content = useIntlayer('layout');
  const { locale, setLocale } = useLocale();
  const location = useLocation();
  const navigate = useNavigate();
  const { theme, toggleTheme } = useTheme();
  const isAuthenticated = authProvider.isAuthenticated();
  const { isAdmin } = useCurrentUser();

  const handleProfileClick = () => {
    const profileUrl = getLocalizedUrl('/profile', locale);
    navigate(profileUrl);
  };

  const handleAdminRolesClick = () => {
    const adminUrl = getLocalizedUrl('/admin/roles', locale);
    navigate(adminUrl);
  };

  const handleToggleLanguage = () => {
    const nextLocale = locale === Locales.UKRAINIAN ? Locales.ENGLISH : Locales.UKRAINIAN;
    setLocale(nextLocale);
    const newUrl = getLocalizedUrl(`${location.pathname}${location.search}`, nextLocale);
    navigate(newUrl);
  };

  return (
    <div className="min-h-screen flex flex-col bg-background text-foreground transition-colors duration-200">
      <header className="sticky top-0 z-40 w-full border-b border-border/60 bg-background/80 backdrop-blur-md backdrop-saturate-150">
        <div className="container max-w-7xl flex h-16 items-center justify-between px-4 sm:px-6 lg:px-8">
          <Logo />

          <div className="flex items-center gap-1.5 sm:gap-2">
            {/* Language Switcher */}
            <Button
              variant="ghost"
              size="sm"
              onClick={handleToggleLanguage}
              title={String(content.languageTooltip || 'Switch language')}
              className="h-9 px-2.5 rounded-lg text-xs font-semibold gap-1.5 text-muted-foreground hover:text-foreground hover:bg-accent/80 transition-colors"
            >
              <Globe className="h-3.5 w-3.5" />
              <span>{locale === Locales.UKRAINIAN ? 'UA' : 'EN'}</span>
            </Button>

            {/* Dark / Light Mode Toggle */}
            <Button
              variant="ghost"
              size="icon"
              onClick={toggleTheme}
              title={String(content.themeToggleTooltip || 'Toggle theme')}
              className="h-9 w-9 rounded-lg text-muted-foreground hover:text-foreground hover:bg-accent/80 transition-colors"
            >
              {theme === 'dark' ? (
                <Sun className="h-4 w-4 text-amber-400 transition-transform hover:rotate-45" />
              ) : (
                <Moon className="h-4 w-4 text-slate-700 transition-transform hover:-rotate-12" />
              )}
            </Button>

            {isAuthenticated && (
              <>
                <div className="h-4 w-px bg-border/80 mx-1 hidden sm:block" />
                {isAdmin && (
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={handleAdminRolesClick}
                    title={String(content.rolesManagementTooltip || 'Roles Management')}
                    className="h-9 w-9 rounded-lg text-muted-foreground hover:text-foreground hover:bg-accent/80 transition-colors"
                  >
                    <ShieldCheck className="h-4 w-4 text-amber-500/90 dark:text-amber-400" />
                  </Button>
                )}
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={handleProfileClick}
                  title={String(content.profileSettingsTooltip)}
                  className="h-9 w-9 rounded-lg text-muted-foreground hover:text-foreground hover:bg-accent/80 transition-colors"
                >
                  <User className="h-4 w-4" />
                </Button>
                <Logout />
              </>
            )}
          </div>
        </div>
      </header>

      <main className="flex-1 container max-w-7xl px-4 sm:px-6 lg:px-8 py-8">
        <ErrorBoundary>
          {children}
        </ErrorBoundary>
      </main>

      <footer className="border-t border-border/40 py-5 px-4 sm:px-6 lg:px-8 bg-muted/20">
        <div className="container max-w-7xl">
          <AppFooter />
        </div>
      </footer>
      <Toaster richColors position="top-right" />
    </div>
  );
}

HomeLayout.propTypes = {
  children: PropTypes.node.isRequired,
};
