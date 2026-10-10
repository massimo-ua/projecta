import {
  useState,
  useEffect,
  useCallback,
  useMemo,
} from 'react';
import { investmentRepository, projectsRepository, assetRepository } from '../api';

export function useProjectTags(projectId) {
  const [loading, setLoading] = useState(true);
  const [tagStats, setTagStats] = useState([]);
  const [currency, setCurrency] = useState('UAH');
  const [searchQuery, setSearchQuery] = useState('');
  const [sortBy, setSortBy] = useState('count'); // 'count' | 'amount' | 'name'

  const fetchTagsAndStats = useCallback(async () => {
    if (!projectId) return;
    setLoading(true);

    try {
      const [distinctTags, investmentsRes, assetsRes, projectRes] = await Promise.all([
        investmentRepository.getTags(projectId).catch(() => []),
        investmentRepository.getInvestments(projectId, 1000, 0).catch(() => [[], 0]),
        assetRepository.getAssets(projectId, 1000, 0).catch(() => [[], 0]),
        projectsRepository.getProject(projectId).catch(() => null),
      ]);

      const [investments] = investmentsRes;
      const loadedInvestments = Array.isArray(investments) ? investments : [];
      const [assets] = assetsRes;
      const loadedAssets = Array.isArray(assets) ? assets : [];

      if (projectRes?.mainCurrency) {
        setCurrency(projectRes.mainCurrency);
      }

      // Collect all tags from distinctTags endpoint, investment models, and asset models
      const allUniqueTags = new Set(distinctTags || []);
      loadedInvestments.forEach((inv) => {
        if (Array.isArray(inv.tags)) {
          inv.tags.forEach((t) => {
            if (t) allUniqueTags.add(t);
          });
        }
      });
      loadedAssets.forEach((ast) => {
        if (Array.isArray(ast.tags)) {
          ast.tags.forEach((t) => {
            if (t) allUniqueTags.add(t);
          });
        }
      });

      // Build stats for each tag
      const stats = Array.from(allUniqueTags).map((tag) => {
        const matchingInvestments = loadedInvestments.filter(
          (inv) => Array.isArray(inv.tags) && inv.tags.includes(tag),
        );
        const matchingAssets = loadedAssets.filter(
          (ast) => Array.isArray(ast.tags) && ast.tags.includes(tag),
        );

        let totalAmount = 0;
        const resourceTypesSet = new Set();
        let latestDate = null;

        matchingInvestments.forEach((inv) => {
          const invAmount = parseFloat(inv.homeAmount)
            || (inv.rawHomeAmount ? inv.rawHomeAmount / 100 : 0)
            || parseFloat(inv.amount)
            || (inv.rawAmount ? inv.rawAmount / 100 : 0)
            || 0;
          totalAmount += invAmount;

          if (inv.resourceType) {
            resourceTypesSet.add(inv.resourceType);
          }

          const invDate = inv.date || inv.rawDate;
          if (invDate && (!latestDate || invDate > latestDate)) {
            latestDate = invDate;
          }
        });

        return {
          name: tag,
          count: matchingInvestments.length + matchingAssets.length,
          investmentCount: matchingInvestments.length,
          assetCount: matchingAssets.length,
          totalAmount,
          formattedAmount: new Intl.NumberFormat().format(Math.round(totalAmount)),
          resourceTypes: Array.from(resourceTypesSet),
          recentDate: latestDate,
        };
      });

      setTagStats(stats);
    } catch (err) {
      console.error('Failed to load project tags', err);
    } finally {
      setLoading(false);
    }
  }, [projectId]);

  useEffect(() => {
    fetchTagsAndStats();
  }, [fetchTagsAndStats]);

  const filteredTags = useMemo(() => {
    let result = [...tagStats];

    if (searchQuery.trim()) {
      const q = searchQuery.trim().toLowerCase().replace(/^#/, '');
      result = result.filter((t) => t.name.toLowerCase().includes(q));
    }

    result.sort((a, b) => {
      if (sortBy === 'amount') {
        return b.totalAmount - a.totalAmount;
      }
      if (sortBy === 'name') {
        return a.name.localeCompare(b.name);
      }
      // default: count desc, then amount desc
      if (b.count !== a.count) {
        return b.count - a.count;
      }
      return b.totalAmount - a.totalAmount;
    });

    return result;
  }, [tagStats, searchQuery, sortBy]);

  const totalInvestmentsTagged = useMemo(
    () => tagStats.reduce((sum, t) => sum + t.count, 0),
    [tagStats],
  );

  const totalAmountTagged = useMemo(
    () => tagStats.reduce((sum, t) => sum + t.totalAmount, 0),
    [tagStats],
  );

  const topTag = useMemo(() => {
    if (tagStats.length === 0) return null;
    return [...tagStats].sort((a, b) => b.count - a.count)[0];
  }, [tagStats]);

  return {
    loading,
    currency,
    searchQuery,
    setSearchQuery,
    sortBy,
    setSortBy,
    tags: filteredTags,
    totalTagsCount: tagStats.length,
    totalInvestmentsTagged,
    totalAmountTagged,
    formattedTotalAmountTagged: new Intl.NumberFormat().format(Math.round(totalAmountTagged)),
    topTag,
    reload: fetchTagsAndStats,
  };
}

export default useProjectTags;
