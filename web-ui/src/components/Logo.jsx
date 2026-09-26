import React from 'react';
import { TrendingUp } from 'lucide-react';
import { Link } from 'react-router-dom';
import { useLocale } from 'react-intlayer';
import { getLocalizedUrl } from 'intlayer';

export function Logo() {
  const { locale } = useLocale();
  const homePath = getLocalizedUrl('/', locale);

  return (
    <Link
      to={homePath}
      className="group flex items-center gap-2.5 font-bold text-xl tracking-tight transition-transform hover:scale-[1.02]"
    >
      <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-gradient-to-tr from-primary via-indigo-600 to-primary text-primary-foreground shadow-md shadow-primary/20 ring-1 ring-white/20 transition-all duration-300 group-hover:shadow-lg group-hover:shadow-primary/30">
        <TrendingUp className="h-5 w-5 transition-transform duration-300 group-hover:rotate-6" />
      </div>
      <span className="font-bold text-lg tracking-tight text-foreground transition-colors group-hover:text-primary">
        Projecta
      </span>
    </Link>
  );
}

export default Logo;
