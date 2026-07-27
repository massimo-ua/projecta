import React from 'react';
import { TrendingUp } from 'lucide-react';
import { Link } from 'react-router-dom';
import { useLocale } from 'react-intlayer';
import { getLocalizedUrl } from 'intlayer';

export function Logo() {
  const { locale } = useLocale();
  const homePath = getLocalizedUrl('/', locale);

  return (
    <Link to={homePath} className="flex items-center gap-2.5 font-bold text-xl tracking-tight text-primary hover:opacity-90 transition-opacity">
      <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary text-primary-foreground shadow-sm">
        <TrendingUp className="h-5 w-5" />
      </div>
      <span>Projecta</span>
    </Link>
  );
}
