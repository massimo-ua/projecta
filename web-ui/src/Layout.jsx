import React from 'react';
import { useNavigate } from 'react-router-dom';
import { useIntlayer, useLocale } from 'react-intlayer';
import { getLocalizedUrl } from 'intlayer';
import { User } from 'lucide-react';
import PropTypes from 'prop-types';
import { Logo } from './components/Logo';
import Logout from './components/Logout';
import { AppFooter, ErrorBoundary } from './components/index.js';
import { Toaster } from '@/components/ui/sonner';
import { Button } from '@/components/ui/button';
import { authProvider } from './api';
import { useI18nHTMLAttributes } from './hooks/useI18nHTMLAttributes';

export default function HomeLayout({ children }) {
  useI18nHTMLAttributes();
  const content = useIntlayer('layout');
  const { locale } = useLocale();
  const navigate = useNavigate();
  const isAuthenticated = authProvider.isAuthenticated();

  const handleProfileClick = () => {
    const profileUrl = getLocalizedUrl('/profile', locale);
    navigate(profileUrl);
  };

  return (
    <div className="min-h-screen flex flex-col bg-background text-foreground">
      <header className="sticky top-0 z-40 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <div className="container flex h-14 items-center justify-between px-4 sm:px-8">
          <Logo />
          {isAuthenticated && (
            <div className="flex items-center gap-2">
              <Button
                variant="ghost"
                size="icon"
                onClick={handleProfileClick}
                title={String(content.profileSettingsTooltip)}
                className="rounded-full hover:bg-accent transition-colors"
              >
                <User className="h-4 w-4" />
              </Button>
              <Logout />
            </div>
          )}
        </div>
      </header>
      <main className="flex-1 container px-4 sm:px-8 py-6">
        <ErrorBoundary>
          {children}
        </ErrorBoundary>
      </main>
      <footer className="border-t py-4 px-4 sm:px-8 bg-muted/30">
        <AppFooter />
      </footer>
      <Toaster />
    </div>
  );
}

HomeLayout.propTypes = {
  children: PropTypes.node.isRequired,
};
