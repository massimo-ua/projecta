/* eslint-disable import/extensions */
import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { Pagination } from './Pagination.js';

describe('Pagination Domain Model', () => {
  describe('single page and empty collections', () => {
    it('hides pagination and boundary links when collection is empty', () => {
      const pagination = new Pagination({ total: 0, pageSize: 5, currentPage: 1 });

      assert.equal(pagination.totalPages, 0);
      assert.equal(pagination.currentPage, 1);
      assert.equal(pagination.hasMultiplePages, false);
      assert.equal(pagination.isFirstPage, true);
      assert.equal(pagination.isLastPage, true);
      assert.equal(pagination.showToBeginning, false);
      assert.equal(pagination.showToEnd, false);
      assert.equal(pagination.canGoPrevious, false);
      assert.equal(pagination.canGoNext, false);
      assert.equal(pagination.firstPage, 1);
      assert.equal(pagination.lastPage, 1);
      assert.equal(pagination.previousPage, 1);
      assert.equal(pagination.nextPage, 1);
      assert.equal(pagination.offset, 0);
    });

    it('hides pagination and boundary links when all items fit on one page', () => {
      const pagination = new Pagination({ total: 5, pageSize: 5, currentPage: 1 });

      assert.equal(pagination.totalPages, 1);
      assert.equal(pagination.hasMultiplePages, false);
      assert.equal(pagination.showToBeginning, false);
      assert.equal(pagination.showToEnd, false);
      assert.equal(pagination.canGoPrevious, false);
      assert.equal(pagination.canGoNext, false);
    });
  });

  describe('multi-page navigation visibility and targets', () => {
    it('hides "To the beginning" and shows "To the end" on the first page', () => {
      const pagination = new Pagination({ total: 23, pageSize: 5, currentPage: 1 });

      assert.equal(pagination.totalPages, 5);
      assert.equal(pagination.currentPage, 1);
      assert.equal(pagination.hasMultiplePages, true);
      assert.equal(pagination.isFirstPage, true);
      assert.equal(pagination.isLastPage, false);
      assert.equal(pagination.showToBeginning, false);
      assert.equal(pagination.showToEnd, true);
      assert.equal(pagination.canGoPrevious, false);
      assert.equal(pagination.canGoNext, true);
      assert.equal(pagination.firstPage, 1);
      assert.equal(pagination.previousPage, 1);
      assert.equal(pagination.nextPage, 2);
      assert.equal(pagination.lastPage, 5);
      assert.equal(pagination.offset, 0);
    });

    it('shows both "To the beginning" and "To the end" on middle pages', () => {
      const pagination = new Pagination({ total: 23, pageSize: 5, currentPage: 3 });

      assert.equal(pagination.totalPages, 5);
      assert.equal(pagination.currentPage, 3);
      assert.equal(pagination.isFirstPage, false);
      assert.equal(pagination.isLastPage, false);
      assert.equal(pagination.showToBeginning, true);
      assert.equal(pagination.showToEnd, true);
      assert.equal(pagination.canGoPrevious, true);
      assert.equal(pagination.canGoNext, true);
      assert.equal(pagination.firstPage, 1);
      assert.equal(pagination.previousPage, 2);
      assert.equal(pagination.nextPage, 4);
      assert.equal(pagination.lastPage, 5);
      assert.equal(pagination.offset, 10);
    });

    it('shows "To the beginning" and hides "To the end" on the last page', () => {
      const pagination = new Pagination({ total: 23, pageSize: 5, currentPage: 5 });

      assert.equal(pagination.totalPages, 5);
      assert.equal(pagination.currentPage, 5);
      assert.equal(pagination.isFirstPage, false);
      assert.equal(pagination.isLastPage, true);
      assert.equal(pagination.showToBeginning, true);
      assert.equal(pagination.showToEnd, false);
      assert.equal(pagination.canGoPrevious, true);
      assert.equal(pagination.canGoNext, false);
      assert.equal(pagination.firstPage, 1);
      assert.equal(pagination.previousPage, 4);
      assert.equal(pagination.nextPage, 5);
      assert.equal(pagination.lastPage, 5);
      assert.equal(pagination.offset, 20);
    });

    it('handles two-page collections accurately on both boundaries', () => {
      const first = new Pagination({ total: 6, pageSize: 5, currentPage: 1 });
      assert.equal(first.totalPages, 2);
      assert.equal(first.showToBeginning, false);
      assert.equal(first.showToEnd, true);
      assert.equal(first.lastPage, 2);

      const second = new Pagination({ total: 6, pageSize: 5, currentPage: 2 });
      assert.equal(second.totalPages, 2);
      assert.equal(second.showToBeginning, true);
      assert.equal(second.showToEnd, false);
      assert.equal(second.firstPage, 1);
    });
  });

  describe('boundary discipline for invalid and out-of-range inputs', () => {
    it('clamps currentPage below 1 to the first page', () => {
      const pagination = new Pagination({ total: 20, pageSize: 5, currentPage: -2 });

      assert.equal(pagination.currentPage, 1);
      assert.equal(pagination.showToBeginning, false);
      assert.equal(pagination.showToEnd, true);
    });

    it('clamps currentPage above totalPages to the last page', () => {
      const pagination = new Pagination({ total: 20, pageSize: 5, currentPage: 99 });

      assert.equal(pagination.currentPage, 4);
      assert.equal(pagination.showToBeginning, true);
      assert.equal(pagination.showToEnd, false);
    });

    it('floors fractional page numbers', () => {
      const pagination = new Pagination({ total: 20, pageSize: 5, currentPage: 2.8 });

      assert.equal(pagination.currentPage, 2);
      assert.equal(pagination.offset, 5);
    });

    it('normalizes negative total and non-positive pageSize safely', () => {
      const negativeTotal = new Pagination({ total: -10, pageSize: 5, currentPage: 2 });
      assert.equal(negativeTotal.totalPages, 0);
      assert.equal(negativeTotal.currentPage, 1);
      assert.equal(negativeTotal.hasMultiplePages, false);

      const zeroPageSize = new Pagination({ total: 20, pageSize: 0, currentPage: 1 });
      assert.equal(zeroPageSize.totalPages, 0);
      assert.equal(zeroPageSize.hasMultiplePages, false);
      assert.equal(zeroPageSize.showToBeginning, false);
      assert.equal(zeroPageSize.showToEnd, false);

      const defaultState = new Pagination();
      assert.equal(defaultState.totalPages, 0);
      assert.equal(defaultState.currentPage, 1);
      assert.equal(defaultState.hasMultiplePages, false);

      const nonFinite = new Pagination({
        total: Number.NaN,
        pageSize: Number.POSITIVE_INFINITY,
        currentPage: Number.NaN,
      });
      assert.equal(nonFinite.totalPages, 0);
      assert.equal(nonFinite.currentPage, 1);
      assert.equal(nonFinite.hasMultiplePages, false);
      assert.equal(Object.isFrozen(nonFinite), true);
    });
  });
});
