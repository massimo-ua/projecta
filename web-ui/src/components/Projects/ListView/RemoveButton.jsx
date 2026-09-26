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
          className="h-8 px-2.5 rounded-lg text-xs font-medium text-muted-foreground border-border/70 hover:text-destructive hover:border-destructive/30 hover:bg-destructive/10 transition-all gap-1.5"
        >
          <Trash2 className="h-3.5 w-3.5" />
          <span>{String(content.remove)}</span>
        </Button>
      </AlertDialogTrigger>
      <AlertDialogContent className="rounded-2xl sm:max-w-md">
        <AlertDialogHeader>
          <AlertDialogTitle className="text-lg font-bold">{String(content.confirmRemoval)}</AlertDialogTitle>
          <AlertDialogDescription className="text-sm text-muted-foreground">
            {String(content.removeWarning)}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter className="gap-2 sm:gap-0 pt-2">
          <AlertDialogCancel className="rounded-xl">{String(content.cancel)}</AlertDialogCancel>
          <AlertDialogAction
            onClick={onRemove}
            className="rounded-xl bg-destructive text-destructive-foreground hover:bg-destructive/90 transition-colors"
          >
            {String(content.remove)}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}

export default RemoveButton;
