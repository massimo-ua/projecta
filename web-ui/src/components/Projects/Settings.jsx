import React, { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { useIntlayer, useLocale } from 'react-intlayer';
import { getLocalizedUrl } from 'intlayer';
import { projectsRepository } from '../../api';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
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
      <div className="space-y-4 max-w-xl">
        <Skeleton className="h-44 w-full rounded-2xl" />
        <Skeleton className="h-36 w-full rounded-2xl" />
      </div>
    );
  }

  return (
    <div className="space-y-6 max-w-xl">
      <div className="flex items-center gap-2.5 pb-3 border-b border-border/60">
        <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10 text-primary">
          <Settings2 className="h-4 w-4" />
        </div>
        <h2 className="text-lg font-bold tracking-tight text-foreground">{String(content.title)}</h2>
      </div>

      <div className="rounded-2xl border border-border/70 bg-card p-6 shadow-xs space-y-5">
        <div>
          <h3 className="text-base font-semibold text-foreground">{String(content.homeCurrencyTitle)}</h3>
          <p className="text-xs text-muted-foreground mt-0.5">
            {String(content.homeCurrencyDesc)}
          </p>
        </div>

        <form onSubmit={handleSave} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="main-currency" className="text-xs font-semibold">{String(content.homeCurrencyTitle)}</Label>
            <Select value={mainCurrency} onValueChange={setMainCurrency} disabled={saving}>
              <SelectTrigger id="main-currency" className="w-full rounded-xl">
                <SelectValue placeholder={String(content.selectCurrencyPlaceholder)} />
              </SelectTrigger>
              <SelectContent className="rounded-xl">
                {SUPPORTED_CURRENCIES.map((c) => (
                  <SelectItem key={c.code} value={c.code} className="rounded-lg">
                    {c.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="flex justify-end pt-2">
            <Button type="submit" disabled={saving} className="gap-2 font-semibold rounded-xl px-5 h-10 shadow-sm shadow-primary/20">
              {saving && <Loader2 className="h-4 w-4 animate-spin" />}
              <span>{String(content.saveSettingsButton)}</span>
            </Button>
          </div>
        </form>
      </div>

      <div className="rounded-2xl border border-border/70 bg-card p-6 shadow-xs space-y-4">
        <div>
          <div className="flex items-center gap-2">
            <Share2 className="h-4 w-4 text-primary" />
            <h3 className="text-base font-semibold text-foreground">{String(content.projectSharingTitle)}</h3>
          </div>
          <p className="text-xs text-muted-foreground mt-0.5">
            {String(content.projectSharingDesc)}
          </p>
        </div>

        <div className="space-y-2">
          <Label htmlFor="share-url" className="text-xs font-semibold">{String(content.shareableLinkLabel)}</Label>
          <div className="flex gap-2">
            <Input
              id="share-url"
              readOnly
              value={shareUrl}
              onClick={(e) => e.target.select()}
              className="font-mono text-xs bg-muted/50 rounded-xl text-muted-foreground cursor-default focus-visible:ring-0 select-all"
            />
            <Button
              type="button"
              variant="outline"
              onClick={handleCopyShareUrl}
              className="gap-2 shrink-0 rounded-xl border-border/70 hover:border-primary/40 text-xs font-semibold"
            >
              {copied ? <Check className="h-3.5 w-3.5 text-emerald-500" /> : <Copy className="h-3.5 w-3.5" />}
              <span>{copied ? String(content.copiedButton) : String(content.copyButton)}</span>
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
