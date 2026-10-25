import {
  useState,
  useEffect,
  useCallback,
  useMemo,
} from 'react';
import { investmentRepository, projectsRepository, assetRepository } from '../api';
import { cleanTagName, TagBudgetRule } from '../models';

export function useProjectTags(projectId) {
  const [loading, setLoading] = useState(true);
  const [tagStats, setTagStats] = useState([]);
  const [project, setProject] = useState(null);
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

      setProject(projectRes);
      if (projectRes?.mainCurrency) {
        setCurrency(projectRes.mainCurrency);
      }

      // Collect all tags from distinctTags endpoint, investment models,
      // asset models, and project tag budgets
      const allUniqueTags = new Set((distinctTags || []).map(cleanTagName).filter(Boolean));
      if (Array.isArray(projectRes?.tagBudgets)) {
        projectRes.tagBudgets.forEach((tb) => {
          const c = cleanTagName(tb.tag);
          if (c) allUniqueTags.add(c);
        });
      }
      loadedInvestments.forEach((inv) => {
        if (Array.isArray(inv.tags)) {
          inv.tags.forEach((t) => {
            const c = cleanTagName(t);
            if (c) allUniqueTags.add(c);
          });
        }
      });
      loadedAssets.forEach((ast) => {
        if (Array.isArray(ast.tags)) {
          ast.tags.forEach((t) => {
            const c = cleanTagName(t);
            if (c) allUniqueTags.add(c);
          });
        }
      });

      // Build stats for each tag
      const stats = Array.from(allUniqueTags).map((tag) => {
        const matchingInvestments = loadedInvestments.filter(
          (inv) => Array.isArray(inv.tags) && inv.tags.map(cleanTagName).includes(tag),
        );
        const matchingAssets = loadedAssets.filter(
          (ast) => Array.isArray(ast.tags) && ast.tags.map(cleanTagName).includes(tag),
        );

        let totalAmount = 0;
        const resourceTypesSet = new Set();
        let latestDate = null;

        matchingInvestments.forEach((inv) => {
          let invAmount;
          if (typeof inv.rawHomeAmount === 'number') {
            invAmount = inv.rawHomeAmount / 100;
          } else if (inv.homeAmount) {
            invAmount = parseFloat(String(inv.homeAmount).replace(/,/g, '')) || 0;
          } else if (typeof inv.rawAmount === 'number') {
            invAmount = inv.rawAmount / 100;
          } else {
            invAmount = parseFloat(String(inv.amount || '0').replace(/,/g, '')) || 0;
          }

          const tagCount = Array.isArray(inv.tags) && inv.tags.length > 0 ? inv.tags.length : 1;
          const proportionalAmount = invAmount / tagCount;
          totalAmount += proportionalAmount;

          if (inv.resourceType) {
            resourceTypesSet.add(inv.resourceType);
          }

          const invDate = inv.date || inv.rawDate;
          if (invDate && (!latestDate || invDate > latestDate)) {
            latestDate = invDate;
          }
        });

        const rawBudget = (projectRes?.tagBudgets || []).find(
          (tb) => cleanTagName(tb.tag) === tag,
        ) || null;
        let tagBudget = null;
        if (rawBudget instanceof TagBudgetRule) {
          tagBudget = rawBudget;
        } else if (rawBudget) {
          tagBudget = TagBudgetRule.fromDTO(rawBudget) || new TagBudgetRule(rawBudget);
        }

        const projectedAnnualAmount = tagBudget
          ? tagBudget.getAnnualTotal(projectRes?.startDate, projectRes?.endDate)
          : 0;

        const budgetUtilization = projectedAnnualAmount > 0
          ? Math.round((totalAmount / projectedAnnualAmount) * 100)
          : null;

        const isOverBudget = projectedAnnualAmount > 0 && totalAmount > projectedAnnualAmount;

        return {
          name: tag,
          count: matchingInvestments.length + matchingAssets.length,
          investmentCount: matchingInvestments.length,
          assetCount: matchingAssets.length,
          totalAmount: Math.round(totalAmount * 100) / 100,
          formattedAmount: new Intl.NumberFormat().format(Math.round(totalAmount)),
          resourceTypes: Array.from(resourceTypesSet),
          recentDate: latestDate,
          tagBudget,
          projectedAnnualAmount,
          budgetUtilization,
          isOverBudget,
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

  const saveTagBudgets = useCallback(async (updatedRules) => {
    if (!projectId) return null;
    const updatedProject = await projectsRepository.updateProjectBudget(projectId, updatedRules);
    setProject(updatedProject);
    await fetchTagsAndStats();
    return updatedProject;
  }, [projectId, fetchTagsAndStats]);

  const filteredTags = useMemo(() => {
    let result = [...tagStats];

    if (searchQuery.trim()) {
      const q = cleanTagName(searchQuery);
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
    () => Math.round(tagStats.reduce((sum, t) => sum + t.totalAmount, 0) * 100) / 100,
    [tagStats],
  );

  const topTag = useMemo(() => {
    if (tagStats.length === 0) return null;
    return [...tagStats].sort((a, b) => b.count - a.count)[0];
  }, [tagStats]);

  return {
    loading,
    currency,
    project,
    tagBudgets: project?.tagBudgets || [],
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
    saveTagBudgets,
    reload: fetchTagsAndStats,
  };
}

export default useProjectTags;
