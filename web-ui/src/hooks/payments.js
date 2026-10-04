import { useEffect, useState, useCallback, useMemo } from 'react';
import { toast } from 'sonner';
import { paymentRepository } from '../api';
import useTypes from './types';
import { PAGE_SIZE } from '../constants';

export default function usePayments(projectId) {
  const [loading, setLoading] = useState(false);
  const [payments, setPayments] = useState([]);
  const [total, setTotal] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [addModalOpened, setAddModalOpen] = useState(false);
  const [paymentIdToEdit, setPaymentIdToEdit] = useState('');
  const [, types, , setTypesFilter] = useTypes();

  const [filters, setFiltersState] = useState({
    typeId: '',
    fromDate: '',
    toDate: '',
  });

  const [selectedPaymentsMap, setSelectedPaymentsMap] = useState({});

  const setFilter = useCallback((key, value) => {
    setFiltersState((prev) => {
      if (prev[key] === value) return prev;
      return { ...prev, [key]: value };
    });
    setCurrentPage(1);
  }, []);

  const resetFilters = useCallback(() => {
    setFiltersState({
      typeId: '',
      fromDate: '',
      toDate: '',
    });
    setCurrentPage(1);
  }, []);

  const fetchPayments = useCallback((page = currentPage) => {
    if (!projectId) return;

    setLoading(true);
    paymentRepository
      .getPayments(projectId, PAGE_SIZE, (page - 1) * PAGE_SIZE, filters)
      .then(([data, count]) => {
        setPayments(data);
        setTotal(count);
      })
      .catch((err) => {
        console.error('Failed to load payments:', err);
      })
      .finally(() => setLoading(false));
  }, [projectId, currentPage, filters]);

  useEffect(() => {
    fetchPayments(currentPage);
  }, [fetchPayments, currentPage]);

  useEffect(() => {
    if (projectId) {
      setTypesFilter({
        projectId,
        limit: 100,
        offset: 0,
      });
    }
  }, [projectId, setTypesFilter]);

  const onPaginationChange = useCallback((nextPage) => {
    setCurrentPage(nextPage);
  }, []);

  const toggleSelectPayment = useCallback((payment) => {
    setSelectedPaymentsMap((prev) => {
      const next = { ...prev };
      if (next[payment.id]) {
        delete next[payment.id];
      } else {
        next[payment.id] = payment;
      }
      return next;
    });
  }, []);

  const toggleSelectAllPage = useCallback(() => {
    setSelectedPaymentsMap((prev) => {
      const allOnPageSelected = payments.length > 0 && payments.every((p) => prev[p.id]);
      const next = { ...prev };
      if (allOnPageSelected) {
        payments.forEach((p) => {
          delete next[p.id];
        });
      } else {
        payments.forEach((p) => {
          next[p.id] = p;
        });
      }
      return next;
    });
  }, [payments]);

  const clearSelection = useCallback(() => {
    setSelectedPaymentsMap({});
  }, []);

  const selectedPaymentsList = useMemo(
    () => Object.values(selectedPaymentsMap),
    [selectedPaymentsMap],
  );

  const selectedPaymentIds = useMemo(
    () => Object.keys(selectedPaymentsMap),
    [selectedPaymentsMap],
  );

  const isAllPageSelected = useMemo(
    () => payments.length > 0 && payments.every((p) => Boolean(selectedPaymentsMap[p.id])),
    [payments, selectedPaymentsMap],
  );

  const isSomePageSelected = useMemo(
    () => payments.length > 0 && payments.some((p) => Boolean(selectedPaymentsMap[p.id])),
    [payments, selectedPaymentsMap],
  );

  const openAddModal = useCallback(() => {
    setAddModalOpen(true);
  }, []);

  const closeAddModal = useCallback(() => {
    setAddModalOpen(false);
  }, []);

  const onAddSuccess = useCallback((successMessage) => {
    setAddModalOpen(false);
    if (successMessage) toast.success(successMessage);
    if (currentPage === 1) {
      fetchPayments(1);
    } else {
      setCurrentPage(1);
    }
  }, [currentPage, fetchPayments]);

  const openEditModal = useCallback((paymentId) => {
    setPaymentIdToEdit(paymentId);
  }, []);

  const closeEditModal = useCallback(() => {
    setPaymentIdToEdit('');
  }, []);

  const onEditSuccess = useCallback((successMessage) => {
    setPaymentIdToEdit('');
    if (successMessage) toast.success(successMessage);
    fetchPayments(currentPage);
  }, [currentPage, fetchPayments]);

  const removePayment = useCallback(async (paymentId, { successMessage, errorMessage } = {}) => {
    try {
      await paymentRepository.removePayment(projectId, paymentId);
      setSelectedPaymentsMap((prev) => {
        if (!prev[paymentId]) return prev;
        const next = { ...prev };
        delete next[paymentId];
        return next;
      });
      if (successMessage) toast.success(successMessage);
      fetchPayments(currentPage);
    } catch (error) {
      if (errorMessage) {
        toast.error(`${errorMessage}: ${error.message}`);
      }
      console.error('Failed to remove payment:', error);
      throw error;
    }
  }, [projectId, currentPage, fetchPayments]);

  return {
    loading,
    payments,
    total,
    currentPage,
    pageSize: PAGE_SIZE,
    types,
    filters,
    setFilter,
    resetFilters,
    selectedPaymentsMap,
    selectedPaymentsList,
    selectedPaymentIds,
    selectedCount: selectedPaymentIds.length,
    toggleSelectPayment,
    toggleSelectAllPage,
    isAllPageSelected,
    isSomePageSelected,
    clearSelection,
    addModalOpened,
    paymentIdToEdit,
    onPaginationChange,
    openAddModal,
    closeAddModal,
    onAddSuccess,
    openEditModal,
    closeEditModal,
    onEditSuccess,
    removePayment,
    refresh: () => fetchPayments(currentPage),
  };
}
