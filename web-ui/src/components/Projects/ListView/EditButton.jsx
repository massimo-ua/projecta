import React from 'react';
import { useIntlayer } from 'react-intlayer';
import { Button } from '@/components/ui/button';
import { Pencil } from 'lucide-react';

export function EditButton({ onClick }) {
  const content = useIntlayer('list-view');

  return (
    <Button
      variant="outline"
      size="sm"
      className="h-8 text-xs font-medium gap-1.5"
      onClick={onClick}
    >
      <Pencil className="h-3.5 w-3.5" />
      {String(content.edit)}
    </Button>
  );
}
