import {
  useState,
  useEffect,
  useCallback,
  useMemo,
} from 'react';
import {
  parseISO,
  isAfter,
  subMonths,
  startOfYear,
  startOfMonth,
  eachMonthOfInterval,
} from 'date-fns';
import { useLocale } from 'react-intlayer';
import { investmentRepository, assetRepository, projectsRepository } from '../api';
import { ProjectBudgetSchedule } from '../models';

export function getPaymentHomeAmount(p) {
  if (!p) return 0;
  if (typeof p.rawHomeAmount === 'number') {
    return p.rawHomeAmount / 100;
  }
  return parseFloat(String(p.homeAmount || p.amount || '0').replace(/,/g, '')) || 0;
}

export function getAssetHomeAmount(a) {
  if (!a) return 0;
  if (typeof a.rawTotalCost === 'number' && (!a.homeCurrency || a.currency === a.homeCurrency)) {
    return a.rawTotalCost / 100;
  }
  const parsedHome = parseFloat(String(a.homeAmount || '').replace(/,/g, ''));
  if (!Number.isNaN(parsedHome) && parsedHome > 0) {
    return parsedHome;
  }
  if (typeof a.rawTotalCost === 'number') {
    return a.rawTotalCost / 100;
  }
  return parseFloat(String(a.totalCost || a.price || '0').replace(/,/g, '')) || 0;
}

export function getAssetTargetHomeAmount(a) {
  if (!a) return 0;
  if (typeof a.rawTargetHomeAmount === 'number' && a.rawTargetHomeAmount > 0) {
    return a.rawTargetHomeAmount / 100;
  }
  const parsedTargetHome = parseFloat(String(a.targetHomeAmount || '').replace(/,/g, ''));
  if (!Number.isNaN(parsedTargetHome) && parsedTargetHome > 0) {
    return parsedTargetHome;
  }
  if (typeof a.rawTargetPrice === 'number' && a.rawTargetPrice > 0) {
    return a.rawTargetPrice / 100;
  }
  return parseFloat(String(a.targetPrice || '0').replace(/,/g, '')) || 0;
}

export function useProjectAnalytics(projectId) {
  const { locale } = useLocale();
  const [loading, setLoading] = useState(true);
  const [payments, setPayments] = useState([]);
  const [assets, setAssets] = useState([]);
  const [project, setProject] = useState(null);
  const [currency, setCurrency] = useState('UAH');
  const [timeframe, setTimeframe] = useState('all');

  const fetchData = useCallback(async () => {
    if (!projectId) return;
    setLoading(true);
    try {
      const [investmentsRes, assetsRes, projectRes] = await Promise.all([
        investmentRepository.getInvestments(projectId, 1000, 0).catch(() => [[], 0]),
        assetRepository.getAssets(projectId, 1000, 0).catch(() => [[], 0]),
        projectsRepository.getProject(projectId).catch(() => null),
      ]);

      const [loadedInvestments] = investmentsRes;
      const [loadedAssets] = assetsRes;

      setPayments(Array.isArray(loadedInvestments) ? loadedInvestments : []);
      setAssets(Array.isArray(loadedAssets) ? loadedAssets : []);
      setProject(projectRes);
      if (projectRes?.mainCurrency) {
        setCurrency(projectRes.mainCurrency);
      } else if (loadedInvestments?.[0]?.homeCurrency) {
        setCurrency(loadedInvestments[0].homeCurrency);
      }
    } catch {
      // Ignore load error, empty fallback
    } finally {
      setLoading(false);
    }
  }, [projectId]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const filteredData = useMemo(() => {
    const now = new Date();

    const isWithinTimeframe = (dateStr) => {
      if (!dateStr || timeframe === 'all') return true;
      try {
        const itemDate = parseISO(dateStr);
        if (Number.isNaN(itemDate.getTime())) return true;
        if (timeframe === 'year') {
          return isAfter(itemDate, startOfYear(now));
        }
        if (timeframe === '12m') {
          return isAfter(itemDate, subMonths(now, 12));
        }
        if (timeframe === '6m') {
          return isAfter(itemDate, subMonths(now, 6));
        }
      } catch {
        return true;
      }
      return true;
    };

    const filteredPayments = payments.filter((p) => isWithinTimeframe(p.rawDate));

    const activeAssetIdsInPeriod = new Set();
    filteredPayments.forEach((p) => {
      if (p.assetId) activeAssetIdsInPeriod.add(p.assetId);
      if (Array.isArray(p.assetAllocations)) {
        p.assetAllocations.forEach((alloc) => {
          if (alloc.assetId) activeAssetIdsInPeriod.add(alloc.assetId);
        });
      }
    });

    const filteredAssets = timeframe === 'all'
      ? assets
      : assets.filter((a) => (
        isWithinTimeframe(a.rawAcquiredAt)
        || isWithinTimeframe(a.rawCompletedDate)
        || activeAssetIdsInPeriod.has(a.id)
      ));

    const totalInvested = Math.round(
      filteredPayments.reduce((sum, p) => sum + getPaymentHomeAmount(p), 0) * 100,
    ) / 100;

    const totalLaborHours = Math.round(
      filteredPayments.reduce((sum, p) => {
        if (p.resourceType === 'TIME' && p.timeHours) {
          return sum + (Number(p.timeHours) || 0);
        }
        return sum;
      }, 0) * 10,
    ) / 10;

    // Calculate budget schedule from tag budgets
    const budgetSchedule = new ProjectBudgetSchedule({
      tagBudgets: project?.tagBudgets || [],
      payments: filteredPayments,
      startDate: project?.startDate,
      endDate: project?.endDate,
      locale,
    });

    const hasTagBudgets = Array.isArray(project?.tagBudgets) && project.tagBudgets.length > 0;

    // Calculate target budget without double-counting parent/child hierarchies
    const assetById = new Map(assets.map((a) => [a.id, a]));
    let totalTargetBudget = 0;
    let budgetedActualCost = 0;
    let overBudgetCount = 0;
    let budgetedAssetsCount = 0;

    filteredAssets.forEach((a) => {
      const targetAmt = getAssetTargetHomeAmount(a);
      const actualAmt = getAssetHomeAmount(a);
      if (targetAmt > 0) {
        // Check if any parent also has a target budget to avoid double-counting in total
        const parentHasTarget = Array.isArray(a.parents) && a.parents.some((pLink) => {
          const parentAsset = assetById.get(pLink.parentId);
          return parentAsset && getAssetTargetHomeAmount(parentAsset) > 0;
        });

        budgetedAssetsCount += 1;
        if (actualAmt > targetAmt) {
          overBudgetCount += 1;
        }
        if (!parentHasTarget) {
          totalTargetBudget += targetAmt;
          budgetedActualCost += actualAmt;
        }
      }
    });

    let budgetUtilization;
    let budgetVariance;
    let isTagBasedBudget = false;

    if (hasTagBudgets) {
      isTagBasedBudget = true;
      totalTargetBudget = budgetSchedule.totalAnnualBudget;
      budgetedActualCost = totalInvested;
      budgetUtilization = budgetSchedule.budgetUtilization;
      budgetVariance = budgetSchedule.annualVariance;
      overBudgetCount = budgetSchedule.overBudgetTagsCount;
    } else {
      totalTargetBudget = Math.round(totalTargetBudget * 100) / 100;
      budgetedActualCost = Math.round(budgetedActualCost * 100) / 100;
      budgetUtilization = totalTargetBudget > 0
        ? Math.round((budgetedActualCost / totalTargetBudget) * 100)
        : null;
      budgetVariance = Math.round((totalTargetBudget - budgetedActualCost) * 100) / 100;
    }

    const completedAssetsCount = filteredAssets.filter((a) => a.isCompleted).length;
    const activeAssetsCount = filteredAssets.filter((a) => a.isActive).length;
    const totalAssetsCount = filteredAssets.length;

    // Calculate active months span and average monthly burn rate
    let minDate = null;
    let maxDate = null;
    filteredPayments.forEach((p) => {
      if (!p.rawDate) return;
      try {
        const d = parseISO(p.rawDate);
        if (Number.isNaN(d.getTime())) return;
        if (!minDate || d < minDate) minDate = d;
        if (!maxDate || d > maxDate) maxDate = d;
      } catch {
        // ignore invalid date
      }
    });

    let activeMonthsCount = 0;
    if (minDate && maxDate) {
      const months = eachMonthOfInterval({
        start: startOfMonth(minDate),
        end: startOfMonth(maxDate),
      });
      activeMonthsCount = Math.max(1, months.length);
    }

    const avgMonthlyBurn = activeMonthsCount > 0
      ? Math.round((totalInvested / activeMonthsCount) * 100) / 100
      : 0;

    return {
      payments: filteredPayments,
      assets: filteredAssets,
      allAssets: assets,
      hasData: filteredPayments.length > 0 || filteredAssets.length > 0,
      budgetSchedule,
      kpis: {
        totalInvested,
        investmentCount: filteredPayments.length,
        totalLaborHours,
        totalTargetBudget,
        budgetedActualCost,
        budgetUtilization,
        budgetVariance,
        overBudgetCount,
        budgetedAssetsCount,
        totalAssetsCount,
        completedAssetsCount,
        activeAssetsCount,
        avgMonthlyBurn,
        activeMonthsCount,
        isTagBasedBudget,
      },
    };
  }, [payments, assets, timeframe, project, locale]);

  return {
    loading,
    timeframe,
    setTimeframe,
    currency,
    project,
    payments: filteredData.payments,
    assets: filteredData.assets,
    allAssets: filteredData.allAssets,
    kpis: filteredData.kpis,
    budgetSchedule: filteredData.budgetSchedule,
    hasData: filteredData.hasData,
    reload: fetchData,
  };
}

export default useProjectAnalytics;
