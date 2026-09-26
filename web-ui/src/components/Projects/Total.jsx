import React, { useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { useIntlayer } from 'react-intlayer';
import { useProjectTotals } from '../../hooks/projects';
import { Skeleton } from '@/components/ui/skeleton';
import { Badge } from '@/components/ui/badge';
import { Calculator, Wallet, TrendingDown } from 'lucide-react';
import './Total.css';

function TotalCard({ total }) {
  return (
    <div className="rounded-2xl border border-border/70 bg-card p-6 shadow-xs transition-all duration-300 hover:shadow-md hover:border-primary/40">
      <div className="flex items-center justify-between gap-3 mb-4">
        <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
          {total.title}
        </span>
        <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-primary/10 text-primary ring-1 ring-primary/20">
          <Wallet className="h-4 w-4" />
        </div>
      </div>
      <div className="flex items-baseline gap-2.5">
        <span className="text-3xl font-extrabold tracking-tight text-foreground">
          {total.amount}
        </span>
        {total.currency && (
          <Badge variant="outline" className="font-mono text-xs font-semibold border-border/80 bg-muted/30">
            {total.currency}
          </Badge>
        )}
      </div>
    </div>
  );
}

export default function Total() {
  const content = useIntlayer('total');
  const { projectId } = useParams();
  const [loading, totals, updateTotals] = useProjectTotals(projectId);

  useEffect(() => {
    updateTotals();
  }, []);

  if (loading) {
    return (
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
        {[1, 2, 3].map((i) => (
          <div key={i} className="rounded-2xl border border-border/60 p-6 space-y-4 bg-card">
            <div className="flex justify-between items-center">
              <Skeleton className="h-4 w-28 rounded-md" />
              <Skeleton className="h-8 w-8 rounded-xl" />
            </div>
            <Skeleton className="h-8 w-36 rounded-lg" />
          </div>
        ))}
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-2.5 pb-3 border-b border-border/60">
        <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10 text-primary">
          <Calculator className="h-4 w-4" />
        </div>
        <h2 className="text-lg font-bold tracking-tight text-foreground">
          {String(content.summaryTitle)}
        </h2>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
        {totals.map((total) => (
          <TotalCard key={total.key} total={total} />
        ))}
      </div>
    </div>
  );
}
