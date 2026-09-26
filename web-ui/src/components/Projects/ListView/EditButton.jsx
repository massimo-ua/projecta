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
      className="h-8 px-2.5 rounded-lg text-xs font-medium gap-1.5 border-border/70 text-muted-foreground hover:text-foreground hover:border-border hover:bg-accent/60 transition-all"
      onClick={onClick}
    >
      <Pencil className="h-3.5 w-3.5" />
      <span>{String(content.edit)}</span>
    </Button>
  );
}

export default EditButton;
