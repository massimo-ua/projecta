import React from 'react';
import { useIntlayer } from 'react-intlayer';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion';
import { ChevronLeft, ChevronRight, Plus, Inbox } from 'lucide-react';

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
      <div className="space-y-4">
        <div className="flex justify-between items-center pb-2">
          <Skeleton className="h-10 w-36 rounded-xl" />
          <Skeleton className="h-5 w-20 rounded-md" />
        </div>
        <div className="space-y-3">
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className="rounded-2xl border border-border/60 p-5 space-y-3 bg-card">
              <div className="flex justify-between items-center">
                <Skeleton className="h-5 w-1/3 rounded-md" />
                <Skeleton className="h-5 w-24 rounded-md" />
              </div>
              <Skeleton className="h-4 w-1/2 rounded-md" />
            </div>
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-5">
      {/* Top Action Bar */}
      <div className="flex items-center justify-between gap-4 pb-2">
        <Button
          disabled={addButtonDisabled}
          onClick={onAddButtonClick}
          className="gap-2 font-semibold shadow-sm shadow-primary/20 hover:shadow-md transition-all rounded-xl h-10 px-4"
        >
          {addButtonIcon || <Plus className="h-4 w-4" />}
          <span>{addButtonText}</span>
        </Button>

        {total > 0 && (
          <span className="text-xs text-muted-foreground font-medium bg-muted/50 px-2.5 py-1 rounded-full border border-border/40">
            {String(content.totalItemsLabel)}: <strong className="text-foreground">{total}</strong>
          </span>
        )}
      </div>

      {/* Item Cards List */}
      {!items || items.length === 0 ? (
        <div className="flex flex-col items-center justify-center p-12 text-center rounded-2xl border border-dashed border-border/80 bg-muted/20">
          <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-muted text-muted-foreground mb-3">
            <Inbox className="h-6 w-6" />
          </div>
          <p className="text-sm font-semibold text-foreground">{String(content.noRecords)}</p>
        </div>
      ) : (
        <div className="space-y-3">
          {items.map((item, idx) => (
            <div
              key={item.id || idx}
              className="group rounded-2xl border border-border/70 bg-card p-4 sm:p-5 shadow-xs transition-all duration-200 hover:border-border hover:shadow-sm"
            >
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
                <div className="flex-1 min-w-0">
                  {renderItemMainContent(item)}
                </div>

                {/* Amount and Quick Actions directly visible on desktop */}
                <div className="flex items-center justify-between sm:justify-end gap-3 self-stretch sm:self-auto pt-2 sm:pt-0 border-t border-border/40 sm:border-0">
                  {renderItemAmount && (
                    <div className="whitespace-nowrap">
                      {renderItemAmount(item)}
                    </div>
                  )}

                  {/* Actions visible on desktop and mobile */}
                  {renderItemActions && (
                    <div className="flex items-center gap-1.5 shrink-0">
                      {renderItemActions(item)}
                    </div>
                  )}
                </div>
              </div>

              {/* Details Accordion for secondary technical details */}
              {renderItemDetails && (
                <Accordion type="single" collapsible className="w-full mt-2">
                  <AccordionItem value="details" className="border-t border-border/40 border-b-0">
                    <AccordionTrigger className="py-1.5 text-[11px] font-medium text-muted-foreground hover:no-underline hover:text-foreground">
                      {String(content.details || content.detailsAndActions || 'Details')}
                    </AccordionTrigger>
                    <AccordionContent className="pt-2">
                      <div className="text-xs bg-muted/30 p-3 rounded-xl border border-border/50 space-y-1.5">
                        {renderItemDetails(item)}
                      </div>
                    </AccordionContent>
                  </AccordionItem>
                </Accordion>
              )}
            </div>
          ))}
        </div>
      )}

      {/* Modern Pagination Controls */}
      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-3 pt-6 border-t border-border/50">
          <Button
            variant="outline"
            size="sm"
            disabled={currentPage <= 1}
            onClick={() => onPaginationChange(currentPage - 1)}
            className="gap-1.5 text-xs rounded-xl h-9 px-3 border-border/70 text-muted-foreground hover:text-foreground"
          >
            <ChevronLeft className="h-3.5 w-3.5" />
            <span>{String(content.previous)}</span>
          </Button>

          <span className="text-xs text-muted-foreground font-medium px-2">
            {String(content.pageOf)} <strong className="text-foreground">{currentPage}</strong> {String(content.of)} {totalPages}
          </span>

          <Button
            variant="outline"
            size="sm"
            disabled={currentPage >= totalPages}
            onClick={() => onPaginationChange(currentPage + 1)}
            className="gap-1.5 text-xs rounded-xl h-9 px-3 border-border/70 text-muted-foreground hover:text-foreground"
          >
            <span>{String(content.next)}</span>
            <ChevronRight className="h-3.5 w-3.5" />
          </Button>
        </div>
      )}
    </div>
  );
}

export default ListView;
