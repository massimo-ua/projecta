import { useEffect, useState, useCallback } from 'react';
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

  const fetchPayments = useCallback((page = currentPage) => {
    if (!projectId) return;

    setLoading(true);
    paymentRepository
      .getPayments(projectId, PAGE_SIZE, (page - 1) * PAGE_SIZE)
      .then(([data, count]) => {
        setPayments(data);
        setTotal(count);
      })
      .catch((err) => {
        console.error('Failed to load payments:', err);
      })
      .finally(() => setLoading(false));
  }, [projectId, currentPage]);

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
