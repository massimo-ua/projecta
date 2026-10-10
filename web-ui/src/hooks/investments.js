import {
  useEffect,
  useState,
  useCallback,
  useMemo,
} from 'react';
import { toast } from 'sonner';
import { investmentRepository, assetRepository } from '../api';
import { PAGE_SIZE } from '../constants';

export default function useInvestments(projectId) {
  const [loading, setLoading] = useState(false);
  const [investments, setInvestments] = useState([]);
  const [total, setTotal] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [addModalOpened, setAddModalOpen] = useState(false);
  const [investmentIdToEdit, setInvestmentIdToEdit] = useState('');
  const [availableTags, setAvailableTags] = useState([]);
  const [availableAssets, setAvailableAssets] = useState([]);

  const [filters, setFiltersState] = useState({
    assetId: '',
    resourceType: '',
    tag: '',
  });

  const [selectedInvestmentsMap, setSelectedInvestmentsMap] = useState({});

  const setFilter = useCallback((key, value) => {
    setFiltersState((prev) => {
      if (prev[key] === value) return prev;
      return { ...prev, [key]: value };
    });
    setCurrentPage(1);
  }, []);

  const resetFilters = useCallback(() => {
    setFiltersState({
      assetId: '',
      resourceType: '',
      tag: '',
    });
    setCurrentPage(1);
  }, []);

  const fetchInvestments = useCallback((page = currentPage) => {
    if (!projectId) return;

    setLoading(true);
    investmentRepository
      .getInvestments(projectId, PAGE_SIZE, (page - 1) * PAGE_SIZE, filters)
      .then(([data, count]) => {
        setInvestments(data);
        setTotal(count);
      })
      .catch((err) => {
        console.error('Failed to load investments:', err);
      })
      .finally(() => setLoading(false));
  }, [projectId, currentPage, filters]);

  const fetchTags = useCallback(() => {
    if (!projectId) return;
    investmentRepository
      .getTags(projectId)
      .then((tags) => setAvailableTags(tags))
      .catch((err) => console.error('Failed to load tags:', err));
  }, [projectId]);

  const fetchAssets = useCallback(() => {
    if (!projectId) return;
    assetRepository
      .getAssets(projectId, 100, 0)
      .then(([items]) => setAvailableAssets(items))
      .catch((err) => console.error('Failed to load assets in useInvestments:', err));
  }, [projectId]);

  useEffect(() => {
    fetchInvestments(currentPage);
  }, [fetchInvestments, currentPage]);

  useEffect(() => {
    fetchTags();
    fetchAssets();
  }, [fetchTags, fetchAssets]);

  const onPaginationChange = useCallback((nextPage) => {
    setCurrentPage(nextPage);
  }, []);

  const toggleSelectInvestment = useCallback((inv) => {
    setSelectedInvestmentsMap((prev) => {
      const next = { ...prev };
      if (next[inv.id]) {
        delete next[inv.id];
      } else {
        next[inv.id] = inv;
      }
      return next;
    });
  }, []);

  const toggleSelectAllPage = useCallback(() => {
    setSelectedInvestmentsMap((prev) => {
      const allOnPageSelected = investments.length > 0 && investments.every((p) => prev[p.id]);
      const next = { ...prev };
      if (allOnPageSelected) {
        investments.forEach((p) => {
          delete next[p.id];
        });
      } else {
        investments.forEach((p) => {
          next[p.id] = p;
        });
      }
      return next;
    });
  }, [investments]);

  const clearSelection = useCallback(() => {
    setSelectedInvestmentsMap({});
  }, []);

  const selectedInvestmentsList = useMemo(
    () => Object.values(selectedInvestmentsMap),
    [selectedInvestmentsMap],
  );

  const selectedInvestmentIds = useMemo(
    () => Object.keys(selectedInvestmentsMap),
    [selectedInvestmentsMap],
  );

  const isAllPageSelected = useMemo(
    () => investments.length > 0 && investments.every((p) => Boolean(selectedInvestmentsMap[p.id])),
    [investments, selectedInvestmentsMap],
  );

  const isSomePageSelected = useMemo(
    () => investments.length > 0 && investments.some((p) => Boolean(selectedInvestmentsMap[p.id])),
    [investments, selectedInvestmentsMap],
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
    fetchTags();
    fetchAssets();
    if (currentPage === 1) {
      fetchInvestments(1);
    } else {
      setCurrentPage(1);
    }
  }, [currentPage, fetchInvestments, fetchTags, fetchAssets]);

  const openEditModal = useCallback((invId) => {
    setInvestmentIdToEdit(invId);
  }, []);

  const closeEditModal = useCallback(() => {
    setInvestmentIdToEdit('');
  }, []);

  const onEditSuccess = useCallback((successMessage) => {
    setInvestmentIdToEdit('');
    if (successMessage) toast.success(successMessage);
    fetchTags();
    fetchAssets();
    fetchInvestments(currentPage);
  }, [currentPage, fetchInvestments, fetchTags, fetchAssets]);

  const removeInvestment = useCallback(async (invId, { successMessage, errorMessage } = {}) => {
    try {
      await investmentRepository.removeInvestment(projectId, invId);
      setSelectedInvestmentsMap((prev) => {
        if (!prev[invId]) return prev;
        const next = { ...prev };
        delete next[invId];
        return next;
      });
      if (successMessage) toast.success(successMessage);
      fetchTags();
      fetchInvestments(currentPage);
    } catch (error) {
      if (errorMessage) {
        toast.error(`${errorMessage}: ${error.message}`);
      }
      console.error('Failed to remove investment:', error);
      throw error;
    }
  }, [projectId, currentPage, fetchInvestments, fetchTags]);

  return {
    loading,
    investments,
    total,
    currentPage,
    pageSize: PAGE_SIZE,
    availableTags,
    availableAssets,
    filters,
    setFilter,
    resetFilters,
    selectedInvestmentsMap,
    selectedInvestmentsList,
    selectedInvestmentIds,
    selectedCount: selectedInvestmentIds.length,
    toggleSelectInvestment,
    toggleSelectAllPage,
    isAllPageSelected,
    isSomePageSelected,
    clearSelection,
    addModalOpened,
    investmentIdToEdit,
    onPaginationChange,
    openAddModal,
    closeAddModal,
    onAddSuccess,
    openEditModal,
    closeEditModal,
    onEditSuccess,
    removeInvestment,
    refresh: () => {
      fetchInvestments(currentPage);
      fetchTags();
      fetchAssets();
    },
  };
}
