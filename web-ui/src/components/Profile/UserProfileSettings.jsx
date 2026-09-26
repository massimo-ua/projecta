import React from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { useIntlayer, useLocale } from 'react-intlayer';
import { Locales, getLocalizedUrl } from 'intlayer';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { toast } from 'sonner';
import { User, Globe, Sun, Moon, Palette } from 'lucide-react';
import HomeLayout from '../../Layout';
import { useTheme } from '../../hooks/useTheme';

export function UserProfileSettings() {
  const content = useIntlayer('user-profile-settings');
  const { locale, setLocale, availableLocales } = useLocale();
  const { theme, setTheme } = useTheme();
  const { pathname, search } = useLocation();
  const navigate = useNavigate();

  const handleLanguageChange = (newLocale) => {
    setLocale(newLocale);
    const newUrl = getLocalizedUrl(`${pathname}${search}`, newLocale);
    navigate(newUrl);
    toast.success(String(content.saveSuccess));
  };

  const handleThemeChange = (newTheme) => {
    setTheme(newTheme);
    toast.success(String(content.saveSuccess));
  };

  return (
    <HomeLayout>
      <div className="space-y-6 max-w-xl mx-auto">
        <div className="flex items-center gap-3 pb-4 border-b border-border/60">
          <div className="flex h-11 w-11 items-center justify-center rounded-2xl bg-primary/10 text-primary ring-1 ring-primary/20">
            <User className="h-6 w-6" />
          </div>
          <div>
            <h1 className="text-xl sm:text-2xl font-bold tracking-tight text-foreground">
              {String(content.title)}
            </h1>
            <p className="text-xs sm:text-sm text-muted-foreground mt-0.5">
              {String(content.subtitle)}
            </p>
          </div>
        </div>

        {/* Language Card */}
        <div className="rounded-2xl border border-border/70 bg-card p-6 shadow-xs space-y-4">
          <div className="flex items-center gap-2.5">
            <Globe className="h-4 w-4 text-primary" />
            <h3 className="text-base font-semibold text-foreground">{String(content.languageSectionTitle)}</h3>
          </div>
          <p className="text-xs text-muted-foreground leading-relaxed">
            {String(content.languageSectionDesc)}
          </p>

          <div className="space-y-2 pt-1">
            <Label htmlFor="user-language-select" className="text-xs font-semibold text-foreground">
              {String(content.selectLanguageLabel)}
            </Label>
            <Select value={locale} onValueChange={handleLanguageChange}>
              <SelectTrigger id="user-language-select" className="w-full rounded-xl">
                <SelectValue placeholder="Select language" />
              </SelectTrigger>
              <SelectContent className="rounded-xl">
                {availableLocales.map((loc) => (
                  <SelectItem key={loc} value={loc} className="rounded-lg">
                    {loc === Locales.UKRAINIAN ? String(content.languages.uk) : String(content.languages.en)}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>

        {/* Theme Card */}
        <div className="rounded-2xl border border-border/70 bg-card p-6 shadow-xs space-y-4">
          <div className="flex items-center gap-2.5">
            <Palette className="h-4 w-4 text-primary" />
            <h3 className="text-base font-semibold text-foreground">{String(content.themeSectionTitle)}</h3>
          </div>
          <p className="text-xs text-muted-foreground leading-relaxed">
            {String(content.themeSectionDesc)}
          </p>

          <div className="space-y-2 pt-1">
            <Label htmlFor="user-theme-select" className="text-xs font-semibold text-foreground">
              {String(content.selectThemeLabel)}
            </Label>
            <Select value={theme} onValueChange={handleThemeChange}>
              <SelectTrigger id="user-theme-select" className="w-full rounded-xl">
                <SelectValue />
              </SelectTrigger>
              <SelectContent className="rounded-xl">
                <SelectItem value="light" className="rounded-lg">
                  <span className="flex items-center gap-2">
                    <Sun className="h-3.5 w-3.5 text-amber-500" />
                    <span>{String(content.themes?.light || 'Light Mode')}</span>
                  </span>
                </SelectItem>
                <SelectItem value="dark" className="rounded-lg">
                  <span className="flex items-center gap-2">
                    <Moon className="h-3.5 w-3.5 text-indigo-400" />
                    <span>{String(content.themes?.dark || 'Dark Mode')}</span>
                  </span>
                </SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>
      </div>
    </HomeLayout>
  );
}

export default UserProfileSettings;
