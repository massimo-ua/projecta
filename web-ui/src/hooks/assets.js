import { useEffect, useState, useCallback } from 'react';
import { toast } from 'sonner';
import { assetRepository } from '../api';
import { PAGE_SIZE } from '../constants';

export default function useAssets(projectId) {
  const [loading, setLoading] = useState(false);
  const [assets, setAssets] = useState([]);
  const [total, setTotal] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [addModalOpened, setAddModalOpen] = useState(false);
  const [assetIdToEdit, setAssetIdToEdit] = useState('');
  const [linkModalAsset, setLinkModalAsset] = useState(null);

  const fetchAssets = useCallback((page = currentPage) => {
    if (!projectId) return;

    setLoading(true);
    assetRepository
      .getAssets(projectId, PAGE_SIZE, (page - 1) * PAGE_SIZE)
      .then(([data, count]) => {
        setAssets(data);
        setTotal(count);
      })
      .catch((err) => {
        console.error('Failed to load assets:', err);
      })
      .finally(() => setLoading(false));
  }, [projectId, currentPage]);

  useEffect(() => {
    fetchAssets(currentPage);
  }, [fetchAssets, currentPage]);

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
      fetchAssets(1);
    } else {
      setCurrentPage(1);
    }
  }, [currentPage, fetchAssets]);

  const openEditModal = useCallback((assetId) => {
    setAssetIdToEdit(assetId);
  }, []);

  const closeEditModal = useCallback(() => {
    setAssetIdToEdit('');
  }, []);

  const onEditSuccess = useCallback((successMessage) => {
    setAssetIdToEdit('');
    if (successMessage) toast.success(successMessage);
    fetchAssets(currentPage);
  }, [currentPage, fetchAssets]);

  const openLinkModal = useCallback((asset) => {
    setLinkModalAsset(asset);
  }, []);

  const closeLinkModal = useCallback(() => {
    setLinkModalAsset(null);
  }, []);

  const linkChild = useCallback(async (parentId, childId, sharePercentage = 100) => {
    try {
      await assetRepository.linkChild(projectId, parentId, childId, sharePercentage);
      toast.success('Child asset linked successfully');
      setLinkModalAsset(null);
      fetchAssets(currentPage);
    } catch (error) {
      toast.error(`Failed to link asset: ${error.message}`);
      throw error;
    }
  }, [projectId, currentPage, fetchAssets]);

  const unlinkChild = useCallback(async (parentId, childId) => {
    try {
      await assetRepository.unlinkChild(projectId, parentId, childId);
      toast.success('Child asset unlinked successfully');
      fetchAssets(currentPage);
    } catch (error) {
      toast.error(`Failed to unlink asset: ${error.message}`);
      throw error;
    }
  }, [projectId, currentPage, fetchAssets]);

  const removeAsset = useCallback(async (assetId, { successMessage, errorMessage } = {}) => {
    try {
      await assetRepository.removeAsset(projectId, assetId);
      if (successMessage) toast.success(successMessage);
      fetchAssets(currentPage);
    } catch (error) {
      if (errorMessage) {
        toast.error(`${errorMessage}: ${error.message}`);
      }
      console.error('Failed to remove asset:', error);
      throw error;
    }
  }, [projectId, currentPage, fetchAssets]);

  return {
    loading,
    assets,
    total,
    currentPage,
    pageSize: PAGE_SIZE,
    types: [],
    addModalOpened,
    assetIdToEdit,
    linkModalAsset,
    onPaginationChange,
    openAddModal,
    closeAddModal,
    onAddSuccess,
    openEditModal,
    closeEditModal,
    onEditSuccess,
    openLinkModal,
    closeLinkModal,
    linkChild,
    unlinkChild,
    removeAsset,
    refresh: () => fetchAssets(currentPage),
  };
}
