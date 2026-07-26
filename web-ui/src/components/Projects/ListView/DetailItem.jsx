import React from 'react';

export function DetailItem({ label, children }) {
  return (
    <div className="flex flex-col sm:flex-row sm:items-center gap-1 sm:gap-3 py-1">
      <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider min-w-[80px]">
        {label}:
      </span>
      <div className="text-sm font-medium text-foreground">{children}</div>
    </div>
  );
}
