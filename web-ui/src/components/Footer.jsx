import React from 'react';
import { useIntlayer } from 'react-intlayer';

export default function Footer() {
  const content = useIntlayer('footer');
  const startYear = 2024;
  const currentYear = new Date().getFullYear();
  const devPeriod = startYear === currentYear ? currentYear : `${startYear}–${currentYear}`;
  const appVersion = import.meta.env.VITE_APP_VERSION;
  const formattedVersion = appVersion
    ? (appVersion.startsWith('v') ? appVersion : (/^\d+\.\d+/.test(appVersion) ? `v${appVersion}` : appVersion))
    : '';

  return (
    <div className="w-full flex flex-col sm:flex-row items-center justify-between gap-2 text-xs text-muted-foreground font-medium">
      <div className="flex items-center gap-2">
        <span className="h-1.5 w-1.5 rounded-full bg-emerald-500 animate-pulse" />
        <span>Projecta</span>
      </div>
      <div>
        {`© ${devPeriod} ${String(content.createdBy)} Massimo UA${formattedVersion ? ` • ${formattedVersion}` : ''}`}
      </div>
    </div>
  );
}
