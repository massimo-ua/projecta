const toNonNegativeInt = (value) => {
  const num = Number(value);
  if (!Number.isFinite(num) || num <= 0) return 0;
  return Math.floor(num);
};

export class Pagination {
  constructor({ total = 0, pageSize = 10, currentPage = 1 } = {}) {
    this.total = toNonNegativeInt(total);
    this.pageSize = toNonNegativeInt(pageSize);
    this.totalPages = this.pageSize > 0 ? Math.ceil(this.total / this.pageSize) : 0;
    this.firstPage = 1;

    const requestedPage = toNonNegativeInt(currentPage) || 1;
    this.currentPage = this.totalPages > 0
      ? Math.min(requestedPage, this.totalPages)
      : 1;

    Object.freeze(this);
  }

  get hasMultiplePages() {
    return this.totalPages > 1;
  }

  get lastPage() {
    return Math.max(this.firstPage, this.totalPages);
  }

  get isFirstPage() {
    return this.currentPage <= this.firstPage;
  }

  get isLastPage() {
    return this.currentPage >= this.lastPage;
  }

  get canGoPrevious() {
    return this.hasMultiplePages && !this.isFirstPage;
  }

  get canGoNext() {
    return this.hasMultiplePages && !this.isLastPage;
  }

  get showToBeginning() {
    return this.canGoPrevious;
  }

  get showToEnd() {
    return this.canGoNext;
  }

  get previousPage() {
    return Math.max(this.firstPage, this.currentPage - 1);
  }

  get nextPage() {
    return Math.min(this.lastPage, this.currentPage + 1);
  }

  get offset() {
    return (this.currentPage - 1) * this.pageSize;
  }
}

export default Pagination;
