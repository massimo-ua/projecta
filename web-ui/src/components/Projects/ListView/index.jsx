import React from 'react';
import { useIntlayer } from 'react-intlayer';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion';
import { ChevronLeft, ChevronRight, Plus } from 'lucide-react';

export function ListView({
  loading,
  items,
  total,
  currentPage,
  pageSize,
  onPaginationChange,
  onAddButtonClick,
  addButtonIcon,
  addButtonText,
  addButtonDisabled,
  renderItemMainContent,
  renderItemAmount,
  renderItemDetails,
  renderItemActions,
}) {
  const content = useIntlayer('list-view');
  const totalPages = Math.ceil(total / pageSize);

  if (loading) {
    return (
      <div className="space-y-4 p-4">
        <Skeleton className="h-10 w-36 rounded-md" />
        <Skeleton className="h-24 w-full rounded-xl" />
        <Skeleton className="h-24 w-full rounded-xl" />
        <Skeleton className="h-24 w-full rounded-xl" />
      </div>
    );
  }

  return (
    <div className="space-y-4 p-2 sm:p-4">
      {/* Top Bar / Add Button */}
      <div className="flex items-center justify-between pb-2">
        <Button
          disabled={addButtonDisabled}
          onClick={onAddButtonClick}
          className="gap-2 font-semibold shadow-sm"
        >
          {addButtonIcon || <Plus className="h-4 w-4" />}
          <span>{addButtonText}</span>
        </Button>

        {total > 0 && (
          <span className="text-xs text-muted-foreground font-medium">
            {String(content.totalItemsLabel)}: {total}
          </span>
        )}
      </div>

      {/* Item Cards List */}
      {!items || items.length === 0 ? (
        <div className="flex flex-col items-center justify-center p-12 text-center rounded-xl border border-dashed bg-muted/20">
          <p className="text-sm text-muted-foreground">{String(content.noRecords)}</p>
        </div>
      ) : (
        <div className="space-y-3">
          {items.map((item, idx) => (
            <Card key={item.id || idx} className="shadow-sm hover:border-muted-foreground/30 transition-colors">
              <CardContent className="p-4 sm:p-5">
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 pb-2">
                  <div className="flex-1 min-w-0">
                    {renderItemMainContent(item)}
                  </div>
                  {renderItemAmount && (
                    <div className="text-base font-semibold tracking-tight whitespace-nowrap self-start sm:self-auto">
                      {renderItemAmount(item)}
                    </div>
                  )}
                </div>

                <Accordion type="single" collapsible className="w-full">
                  <AccordionItem value="details" className="border-t mt-2 pt-1 border-b-0">
                    <AccordionTrigger className="py-2 text-xs font-medium text-muted-foreground hover:no-underline hover:text-foreground">
                      {String(content.detailsAndActions)}
                    </AccordionTrigger>
                    <AccordionContent className="pt-2 space-y-4">
                      {renderItemDetails && (
                        <div className="space-y-2 text-sm bg-muted/40 p-3 rounded-lg border">
                          {renderItemDetails(item)}
                        </div>
                      )}
                      {renderItemActions && (
                        <div className="flex items-center gap-2 pt-1">
                          {renderItemActions(item)}
                        </div>
                      )}
                    </AccordionContent>
                  </AccordionItem>
                </Accordion>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* Pagination Controls */}
      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-3 pt-4 border-t">
          <Button
            variant="outline"
            size="sm"
            disabled={currentPage <= 1}
            onClick={() => onPaginationChange(currentPage - 1)}
            className="gap-1 text-xs"
          >
            <ChevronLeft className="h-3.5 w-3.5" />
            {String(content.previous)}
          </Button>
          <span className="text-xs text-muted-foreground font-medium">
            {String(content.pageOf)} {currentPage} {String(content.of)} {totalPages}
          </span>
          <Button
            variant="outline"
            size="sm"
            disabled={currentPage >= totalPages}
            onClick={() => onPaginationChange(currentPage + 1)}
            className="gap-1 text-xs"
          >
            {String(content.next)}
            <ChevronRight className="h-3.5 w-3.5" />
          </Button>
        </div>
      )}
    </div>
  );
}
