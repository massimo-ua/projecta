import React from 'react';
import { useIntlayer } from 'react-intlayer';
import { Button } from '@/components/ui/button';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog';
import { Trash2 } from 'lucide-react';

export function RemoveButton({ onRemove }) {
  const content = useIntlayer('list-view');

  return (
    <AlertDialog>
      <AlertDialogTrigger asChild>
        <Button
          variant="outline"
          size="sm"
          className="h-8 text-xs font-medium text-destructive border-destructive/30 hover:bg-destructive/10 hover:text-destructive gap-1.5"
        >
          <Trash2 className="h-3.5 w-3.5" />
          {String(content.remove)}
        </Button>
      </AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{String(content.confirmRemoval)}</AlertDialogTitle>
          <AlertDialogDescription>
            {String(content.removeWarning)}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>{String(content.cancel)}</AlertDialogCancel>
          <AlertDialogAction
            onClick={onRemove}
            className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
          >
            {String(content.remove)}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
