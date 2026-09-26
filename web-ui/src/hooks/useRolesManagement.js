import { useEffect, useState, useCallback } from 'react';
import { toast } from 'sonner';
import { usersRepository } from '../api';
import { User } from '../models/User.js';
import { PAGE_SIZE } from '../constants';

export function useRolesManagement() {
  const [loading, setLoading] = useState(false);
  const [users, setUsers] = useState([]);
  const [total, setTotal] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [editingUser, setEditingUser] = useState(null);
  const [isUpdating, setIsUpdating] = useState(false);

  const fetchUsers = useCallback((page = currentPage) => {
    setLoading(true);
    usersRepository
      .getUsers(PAGE_SIZE, (page - 1) * PAGE_SIZE)
      .then(([data, count]) => {
        setUsers(data);
        setTotal(count);
      })
      .catch((err) => {
        toast.error(`Failed to load users: ${err.message}`);
      })
      .finally(() => setLoading(false));
  }, [currentPage]);

  useEffect(() => {
    fetchUsers(currentPage);
  }, [fetchUsers, currentPage]);

  const onPaginationChange = useCallback((nextPage) => {
    setCurrentPage(nextPage);
  }, []);

  const openEditRoles = useCallback((user) => {
    setEditingUser(user);
  }, []);

  const closeEditRoles = useCallback(() => {
    setEditingUser(null);
  }, []);

  const saveRoles = useCallback(async (userId, newRoles, { successMessage, errorMessage } = {}) => {
    const validation = User.validateRoles(newRoles);
    if (!validation.valid) {
      toast.error(validation.error);
      return false;
    }

    setIsUpdating(true);
    try {
      const updatedUser = await usersRepository.assignRoles(userId, newRoles);
      setUsers((prevUsers) => prevUsers.map((u) => (u.id === userId ? updatedUser : u)));
      if (editingUser && editingUser.id === userId) {
        setEditingUser(null);
      }
      if (successMessage) {
        toast.success(successMessage);
      }
      return true;
    } catch (err) {
      if (errorMessage) {
        toast.error(`${errorMessage}: ${err.message}`);
      } else {
        toast.error(`Failed to assign roles: ${err.message}`);
      }
      return false;
    } finally {
      setIsUpdating(false);
    }
  }, [editingUser]);

  return {
    loading,
    users,
    total,
    currentPage,
    pageSize: PAGE_SIZE,
    editingUser,
    isUpdating,
    onPaginationChange,
    openEditRoles,
    closeEditRoles,
    saveRoles,
    refresh: () => fetchUsers(currentPage),
  };
}

export default useRolesManagement;
