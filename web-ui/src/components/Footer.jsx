import React from 'react';
import { useIntlayer } from 'react-intlayer';

export default function Footer() {
  const content = useIntlayer('footer');
  const startYear = 2024;
  const currentYear = new Date().getFullYear();
  const devPeriod = startYear === currentYear ? currentYear : `${startYear}-${currentYear}`;

  return (
    <div className="w-full text-right font-medium text-xs text-muted-foreground">
      {`Projecta Web UI ©${devPeriod} ${String(content.createdBy)} Massimo UA`}
    </div>
  );
}
