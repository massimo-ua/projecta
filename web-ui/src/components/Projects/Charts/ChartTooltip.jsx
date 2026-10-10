import React from 'react';

export function ChartTooltip({ active, payload, label, currency }) {
  if (!active || !payload || !payload.length) return null;

  return (
    <div className="rounded-xl border border-border/80 bg-popover/95 backdrop-blur-md px-3.5 py-2.5 shadow-xl text-xs space-y-1.5 z-50">
      {label && <p className="font-semibold text-foreground border-b border-border/40 pb-1">{label}</p>}
      {payload.map((item, index) => {
        const val = item.value ?? 0;
        const color = item.color || item.payload?.fill || item.fill || '#3b82f6';
        const name = item.name || item.dataKey;
        const percentage = item.payload?.percentage;

        return (
          <div key={index} className="space-y-0.5">
            <div className="flex items-center justify-between gap-3">
              <div className="flex items-center gap-1.5">
                <span className="h-2.5 w-2.5 rounded-full shrink-0" style={{ backgroundColor: color }} />
                <span className="text-muted-foreground">{name}:</span>
              </div>
              <span className="font-bold text-foreground font-mono">
                {new Intl.NumberFormat().format(val)} {currency}
              </span>
            </div>
            {percentage !== undefined && (
              <p className="text-[11px] text-muted-foreground pl-4">
                {percentage}% of total
              </p>
            )}
          </div>
        );
      })}
    </div>
  );
}

export default ChartTooltip;
