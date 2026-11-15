import {
  useState,
  useEffect,
  useCallback,
  useMemo,
} from 'react';
import { toast } from 'sonner';
import { assetRepository } from '../api';
import {
  buildChartSeries,
  calculateInstantSchedulePreview,
} from '../models/CashFlowCalculator';

export function useProjectCashFlow(projectId) {
  const [forecast, setForecast] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [savingAssetId, setSavingAssetId] = useState(null);
  const [selectedTag, setSelectedTag] = useState('');
  const [includeOngoingInCommitted, setIncludeOngoingInCommitted] = useState(false);

  const loadForecast = useCallback(async (tagOverride) => {
    if (!projectId) return;
    setLoading(true);
    setError(null);
    const activeTag = tagOverride !== undefined ? tagOverride : selectedTag;
    try {
      const data = await assetRepository.getCashFlowForecast(projectId, {
        tag: activeTag || undefined,
      });
      setForecast(data);
    } catch (err) {
      setError(err?.message || 'Failed to load cash flow forecast');
    } finally {
      setLoading(false);
    }
  }, [projectId, selectedTag]);

  useEffect(() => {
    loadForecast();
  }, [loadForecast]);

  const previewSchedule = useCallback((assetId, nextSchedule) => {
    setForecast((prev) => calculateInstantSchedulePreview(prev, assetId, nextSchedule));
  }, []);

  const commitSchedule = useCallback(async (assetId, nextSchedule) => {
    if (!projectId || !assetId) return;
    setSavingAssetId(assetId);
    setForecast((prev) => calculateInstantSchedulePreview(prev, assetId, nextSchedule));
    try {
      const res = await assetRepository.rescheduleAsset(projectId, assetId, nextSchedule);
      if (Array.isArray(res?.warnings) && res.warnings.length > 0) {
        toast.warning(`Schedule saved (${res.warnings.length} child timeline notice)`);
      } else {
        toast.success('Schedule updated');
      }
      await loadForecast();
    } catch (err) {
      toast.error(err?.message || 'Failed to update schedule');
      await loadForecast();
    } finally {
      setSavingAssetId(null);
    }
  }, [projectId, loadForecast]);

  const chartData = useMemo(
    () => buildChartSeries(forecast, { includeOngoingInCommitted }),
    [forecast, includeOngoingInCommitted],
  );

  const availableTags = useMemo(() => {
    const assetCount = forecast?.assets?.length || 0;
    return { assetCount };
  }, [forecast]);

  return {
    forecast,
    chartData,
    loading,
    error,
    savingAssetId,
    selectedTag,
    setSelectedTag,
    includeOngoingInCommitted,
    setIncludeOngoingInCommitted,
    previewSchedule,
    commitSchedule,
    reload: loadForecast,
    availableTags,
  };
}
