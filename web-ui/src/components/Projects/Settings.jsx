import React, { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { useIntlayer, useLocale } from 'react-intlayer';
import { getLocalizedUrl } from 'intlayer';
import { projectsRepository } from '../../api';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Skeleton } from '@/components/ui/skeleton';
import { toast } from 'sonner';
import { Loader2, Settings2, Share2, Check, Copy } from 'lucide-react';

const SUPPORTED_CURRENCIES = [
  { code: 'UAH', name: 'Ukrainian Hryvnia (UAH ₴)' },
  { code: 'USD', name: 'US Dollar (USD $)' },
  { code: 'EUR', name: 'Euro (EUR €)' },
  { code: 'PLN', name: 'Polish Zloty (PLN zł)' },
];

export default function Settings() {
  const content = useIntlayer('project-settings');
  const { locale } = useLocale();
  const { projectId } = useParams();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [mainCurrency, setMainCurrency] = useState('UAH');
  const [project, setProject] = useState(null);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    setLoading(true);
    projectsRepository
      .getProjects(100, 0)
      .then((projects) => {
        const current = projects.find((p) => p.id === projectId);
        if (current) {
          setProject(current);
          if (current.mainCurrency) {
            setMainCurrency(current.mainCurrency);
          }
        }
      })
      .catch((err) => {
        toast.error(`${String(content.failedToLoad)}: ${err.message}`);
      })
      .finally(() => {
        setLoading(false);
      });
  }, [projectId]);

  const handleSave = async (e) => {
    e.preventDefault();
    setSaving(true);
    try {
      await projectsRepository.updateProjectSettings(projectId, { mainCurrency });
      toast.success(String(content.settingsUpdatedSuccess));
    } catch (err) {
      toast.error(`${String(content.failedToUpdate)}: ${err.message}`);
    } finally {
      setSaving(false);
    }
  };

  const shareUrl = project?.shareToken
    ? `${window.location.origin}${getLocalizedUrl(`/projects/share/${project.shareToken}`, locale)}`
    : '';

  const handleCopyShareUrl = () => {
    if (!shareUrl) return;
    navigator.clipboard.writeText(shareUrl).then(() => {
      setCopied(true);
      toast.success(String(content.shareLinkCopied));
      setTimeout(() => setCopied(false), 2000);
    }).catch(() => {
      toast.error(String(content.failedToCopyLink));
    });
  };

  if (loading) {
    return (
      <div className="p-4 space-y-4">
        <Skeleton className="h-44 w-full max-w-lg rounded-xl" />
      </div>
    );
  }

  return (
    <div className="p-4 space-y-6 max-w-xl">
      <div className="flex items-center gap-2 pb-2 border-b">
        <Settings2 className="h-5 w-5 text-primary" />
        <h2 className="text-lg font-semibold tracking-tight">{String(content.title)}</h2>
      </div>

      <Card className="shadow-sm">
        <CardHeader>
          <CardTitle className="text-base">{String(content.homeCurrencyTitle)}</CardTitle>
          <CardDescription>
            {String(content.homeCurrencyDesc)}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSave} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="main-currency">{String(content.homeCurrencyTitle)}</Label>
              <Select value={mainCurrency} onValueChange={setMainCurrency} disabled={saving}>
                <SelectTrigger id="main-currency" className="w-full">
                  <SelectValue placeholder={String(content.selectCurrencyPlaceholder)} />
                </SelectTrigger>
                <SelectContent>
                  {SUPPORTED_CURRENCIES.map((c) => (
                    <SelectItem key={c.code} value={c.code}>
                      {c.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="flex justify-end pt-2">
              <Button type="submit" disabled={saving} className="gap-2 font-semibold">
                {saving && <Loader2 className="h-4 w-4 animate-spin" />}
                {String(content.saveSettingsButton)}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>

      <Card className="shadow-sm">
        <CardHeader>
          <div className="flex items-center gap-2">
            <Share2 className="h-4 w-4 text-primary" />
            <CardTitle className="text-base">{String(content.projectSharingTitle)}</CardTitle>
          </div>
          <CardDescription>
            {String(content.projectSharingDesc)}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="share-url">{String(content.shareableLinkLabel)}</Label>
            <div className="flex gap-2">
              <Input
                id="share-url"
                readOnly
                value={shareUrl}
                onClick={(e) => e.target.select()}
                className="font-mono text-xs bg-muted text-muted-foreground cursor-default focus-visible:ring-0 select-all"
              />
              <Button
                type="button"
                variant="outline"
                onClick={handleCopyShareUrl}
                className="gap-2 shrink-0"
              >
                {copied ? <Check className="h-4 w-4 text-green-500" /> : <Copy className="h-4 w-4" />}
                {copied ? String(content.copiedButton) : String(content.copyButton)}
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
