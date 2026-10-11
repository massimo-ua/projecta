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
} from 'date-fns';
import { investmentRepository, assetRepository, projectsRepository } from '../api';

export function useProjectAnalytics(projectId) {
  const [loading, setLoading] = useState(true);
  const [payments, setPayments] = useState([]);
  const [assets, setAssets] = useState([]);
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
    const filteredAssets = assets.filter((a) => isWithinTimeframe(a.rawAcquiredAt));

    return {
      payments: filteredPayments,
      assets: filteredAssets,
      hasData: filteredPayments.length > 0 || filteredAssets.length > 0,
    };
  }, [payments, assets, timeframe]);

  return {
    loading,
    timeframe,
    setTimeframe,
    currency,
    payments: filteredData.payments,
    assets: filteredData.assets,
    hasData: filteredData.hasData,
    reload: fetchData,
  };
}

export default useProjectAnalytics;
