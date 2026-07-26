import React from 'react';

export default function Footer() {
  const startYear = 2024;
  const currentYear = new Date().getFullYear();
  const devPeriod = startYear === currentYear ? currentYear : `${startYear}-${currentYear}`;

  return (
    <div className="w-full text-right font-medium text-xs text-muted-foreground">
      {`Projecta Web UI ©${devPeriod} Created by Massimo UA`}
    </div>
  );
}
