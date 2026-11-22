import React from 'react';
import PropTypes from 'prop-types';
import { useIntlayer } from 'react-intlayer';
import {
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Pagination } from '../../../models/Pagination';

const NAV_BUTTON_CLASS = 'gap-1.5 text-xs rounded-xl h-9 px-3 border-border/70 text-muted-foreground hover:text-foreground';

export function PaginationControls({ pagination, onPageChange }) {
  const content = useIntlayer('list-view');

  if (!pagination.hasMultiplePages) {
    return null;
  }

  return (
    <nav
      aria-label="Pagination"
      className="flex flex-wrap items-center justify-center gap-2 sm:gap-3 pt-6 border-t border-border/50"
    >
      {pagination.showToBeginning && (
        <Button
          variant="outline"
          size="sm"
          onClick={() => onPageChange?.(pagination.firstPage)}
          className={NAV_BUTTON_CLASS}
        >
          <ChevronsLeft className="h-3.5 w-3.5" />
          <span>{String(content.toBeginning)}</span>
        </Button>
      )}

      <Button
        variant="outline"
        size="sm"
        disabled={!pagination.canGoPrevious}
        onClick={() => onPageChange?.(pagination.previousPage)}
        className={NAV_BUTTON_CLASS}
      >
        <ChevronLeft className="h-3.5 w-3.5" />
        <span>{String(content.previous)}</span>
      </Button>

      <span className="text-xs text-muted-foreground font-medium px-2">
        {String(content.pageOf)}
        {' '}
        <strong className="text-foreground">{pagination.currentPage}</strong>
        {' '}
        {String(content.of)}
        {' '}
        {pagination.totalPages}
      </span>

      <Button
        variant="outline"
        size="sm"
        disabled={!pagination.canGoNext}
        onClick={() => onPageChange?.(pagination.nextPage)}
        className={NAV_BUTTON_CLASS}
      >
        <span>{String(content.next)}</span>
        <ChevronRight className="h-3.5 w-3.5" />
      </Button>

      {pagination.showToEnd && (
        <Button
          variant="outline"
          size="sm"
          onClick={() => onPageChange?.(pagination.lastPage)}
          className={NAV_BUTTON_CLASS}
        >
          <span>{String(content.toEnd)}</span>
          <ChevronsRight className="h-3.5 w-3.5" />
        </Button>
      )}
    </nav>
  );
}

PaginationControls.propTypes = {
  pagination: PropTypes.instanceOf(Pagination).isRequired,
  onPageChange: PropTypes.func,
};

PaginationControls.defaultProps = {
  onPageChange: undefined,
};

export default PaginationControls;
