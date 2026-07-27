import React from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { useIntlayer, useLocale } from 'react-intlayer';
import { Locales, getLocalizedUrl } from 'intlayer';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { toast } from 'sonner';
import { User, Globe } from 'lucide-react';
import HomeLayout from '../../Layout';

export function UserProfileSettings() {
  const content = useIntlayer('user-profile-settings');
  const { locale, setLocale, availableLocales } = useLocale();
  const { pathname, search } = useLocation();
  const navigate = useNavigate();

  const handleLanguageChange = (newLocale) => {
    setLocale(newLocale);
    const newUrl = getLocalizedUrl(`${pathname}${search}`, newLocale);
    navigate(newUrl);
    toast.success(String(content.saveSuccess));
  };

  return (
    <HomeLayout>
      <div className="p-4 space-y-6 max-w-xl mx-auto">
        <div className="flex items-center gap-3 pb-2 border-b">
          <User className="h-6 w-6 text-primary" />
          <div>
            <h1 className="text-xl font-bold tracking-tight">{String(content.title)}</h1>
            <p className="text-sm text-muted-foreground">{String(content.subtitle)}</p>
          </div>
        </div>

        <Card className="shadow-sm">
          <CardHeader>
            <div className="flex items-center gap-2">
              <Globe className="h-5 w-5 text-primary" />
              <CardTitle className="text-base">{String(content.languageSectionTitle)}</CardTitle>
            </div>
            <CardDescription>
              {String(content.languageSectionDesc)}
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="user-language-select">{String(content.selectLanguageLabel)}</Label>
              <Select value={locale} onValueChange={handleLanguageChange}>
                <SelectTrigger id="user-language-select" className="w-full">
                  <SelectValue placeholder="Select language" />
                </SelectTrigger>
                <SelectContent>
                  {availableLocales.map((loc) => (
                    <SelectItem key={loc} value={loc}>
                      {loc === Locales.UKRAINIAN ? String(content.languages.uk) : String(content.languages.en)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </CardContent>
        </Card>
      </div>
    </HomeLayout>
  );
}

export default UserProfileSettings;
