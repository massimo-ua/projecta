import React, { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { projectsRepository } from '../../api';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
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
import { Loader2, Settings2 } from 'lucide-react';

const SUPPORTED_CURRENCIES = [
  { code: 'UAH', name: 'Ukrainian Hryvnia (UAH ₴)' },
  { code: 'USD', name: 'US Dollar (USD $)' },
  { code: 'EUR', name: 'Euro (EUR €)' },
  { code: 'PLN', name: 'Polish Zloty (PLN zł)' },
];

export default function Settings() {
  const { projectId } = useParams();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [mainCurrency, setMainCurrency] = useState('UAH');

  useEffect(() => {
    setLoading(true);
    projectsRepository
      .getProjects(100, 0)
      .then((projects) => {
        const current = projects.find((p) => p.id === projectId);
        if (current && current.mainCurrency) {
          setMainCurrency(current.mainCurrency);
        }
      })
      .catch((err) => {
        toast.error(`Failed to load project settings: ${err.message}`);
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
      toast.success('Project home currency updated successfully');
    } catch (err) {
      toast.error(`Failed to update project settings: ${err.message}`);
    } finally {
      setSaving(false);
    }
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
        <h2 className="text-lg font-semibold tracking-tight">Project Settings</h2>
      </div>

      <Card className="shadow-sm">
        <CardHeader>
          <CardTitle className="text-base">Home Currency</CardTitle>
          <CardDescription>
            Select the primary home currency for this project. All payments, assets, and summary totals will be converted to this currency using official National Bank of Ukraine (NBU) exchange rates.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSave} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="main-currency">Home Currency</Label>
              <Select value={mainCurrency} onValueChange={setMainCurrency} disabled={saving}>
                <SelectTrigger id="main-currency" className="w-full">
                  <SelectValue placeholder="Select currency" />
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
                Save Settings
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
