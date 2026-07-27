import React, { useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { useIntlayer } from 'react-intlayer';
import { useProjectTotals } from '../../hooks/projects';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { Calculator } from 'lucide-react';
import './Total.css';

function TotalCard({ total }) {
  return (
    <Card className="shadow-sm hover:shadow-md transition-shadow">
      <CardHeader className="p-5 pb-2">
        <CardTitle className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
          {total.title}
        </CardTitle>
      </CardHeader>
      <CardContent className="p-5 pt-0">
        <div className="flex items-baseline gap-2">
          <span className="text-3xl font-bold tracking-tight text-foreground">
            {total.amount}
          </span>
          {total.currency && (
            <span className="text-sm font-semibold text-muted-foreground">
              {total.currency}
            </span>
          )}
        </div>
      </CardContent>
    </Card>
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
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6 p-4">
        <Skeleton className="h-28 w-full rounded-xl" />
        <Skeleton className="h-28 w-full rounded-xl" />
        <Skeleton className="h-28 w-full rounded-xl" />
      </div>
    );
  }

  return (
    <div className="p-4 space-y-4">
      <div className="flex items-center gap-2 pb-2 border-b">
        <Calculator className="h-5 w-5 text-primary" />
        <h2 className="text-lg font-semibold tracking-tight">{String(content.summaryTitle)}</h2>
      </div>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
        {totals.map((total) => (
          <TotalCard key={total.key} total={total} />
        ))}
      </div>
    </div>
  );
}
