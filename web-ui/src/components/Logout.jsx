import React from 'react';
import { authProvider } from '../api';
import { useNavigate } from 'react-router-dom';
import { useLocale } from 'react-intlayer';
import { getLocalizedUrl } from 'intlayer';
import { Button } from '@/components/ui/button';
import { LogOut } from 'lucide-react';

export default function Logout() {
  const navigate = useNavigate();
  const { locale } = useLocale();

  const onClick = () => {
    authProvider.logout();
    const loginUrl = getLocalizedUrl('/login', locale);
    navigate(loginUrl);
  };

  return authProvider.isAuthenticated() ? (
    <Button
      variant="ghost"
      size="icon"
      onClick={onClick}
      title="Logout"
      className="rounded-full hover:bg-destructive/10 hover:text-destructive transition-colors"
    >
      <LogOut className="h-4 w-4" />
    </Button>
  ) : null;
}
