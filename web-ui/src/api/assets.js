import {
  fromISO, toDateView, toISO, toPrice, toPriceView,
} from './mappers';
import { Asset } from '../models/Asset';

const toDomain = ({
  asset_id,
  name,
  description,
  status,
  start_date,
  completed_date,
  target_completion_date,
  target_price,
  target_currency,
  target_home_amount,
  targetHomeAmount,
  progress_percentage,
  progressPercentage,
  direct_cost,
  direct_cost_home_amount,
  total_cost,
  price,
  currency,
  home_amount,
  home_currency,
  homeAmount,
  homeCurrency,
  acquired_at,
  children = [],
  parents = [],
  tags = [],
  type,
  project,
}) => {
  const finalPrice = total_cost !== undefined ? total_cost : (price !== undefined ? price : 0);
  const finalHomeAmount = home_amount !== undefined ? home_amount : (homeAmount !== undefined ? homeAmount : finalPrice);
  const finalHomeCurrency = home_currency || homeCurrency || project?.mainCurrency || currency;
  const finalCurrency = currency || target_currency || finalHomeCurrency;
  const finalTargetHome = target_home_amount !== undefined ? target_home_amount : targetHomeAmount;
  const finalProgress = progress_percentage !== undefined ? progress_percentage : progressPercentage;

  const childrenMapped = (children || []).map((c) => ({
    childId: c.child_id,
    childName: c.child_name,
    sharePercentage: c.share_percentage,
    totalCost: toPriceView(c.total_cost || 0),
    rawTotalCost: c.total_cost || 0,
  }));

  const parentsMapped = (parents || []).map((p) => ({
    parentId: p.parent_id,
    parentName: p.parent_name,
    sharePercentage: p.share_percentage,
  }));

  return new Asset({
    key: asset_id,
    id: asset_id,
    name,
    description,
    status: status || 'ACTIVE',
    startDate: toDateView(start_date || acquired_at),
    rawStartDate: start_date || acquired_at,
    completedDate: completed_date ? toDateView(completed_date) : null,
    rawCompletedDate: completed_date,
    targetCompletionDate: target_completion_date ? toDateView(target_completion_date) : null,
    rawTargetCompletionDate: target_completion_date,
    targetPrice: target_price ? toPriceView(target_price) : null,
    rawTargetPrice: target_price,
    targetCurrency: target_currency || finalCurrency,
    targetHomeAmount: finalTargetHome ? toPriceView(finalTargetHome) : null,
    rawTargetHomeAmount: finalTargetHome,
    progressPercentage: finalProgress,
    directCost: toPriceView(direct_cost || 0),
    rawDirectCost: direct_cost || 0,
    directCostHomeAmount: toPriceView(direct_cost_home_amount ?? direct_cost ?? 0),
    totalCost: toPriceView(finalPrice),
    rawTotalCost: finalPrice,
    price: toPriceView(finalPrice),
    rawPrice: finalPrice,
    currency: finalCurrency,
    homeAmount: toPriceView(finalHomeAmount),
    homeCurrency: finalHomeCurrency,
    acquiredAt: toDateView(acquired_at || start_date),
    rawAcquiredAt: acquired_at || start_date,
    children: childrenMapped,
    parents: parentsMapped,
    tags: tags || [],
    type: type?.name,
    category: type?.category?.name,
  });
};

const toEditAssetView = ({
  asset_id,
  name,
  description,
  status,
  start_date,
  completed_date,
  target_completion_date,
  target_price,
  target_currency,
  price,
  currency,
  acquired_at,
  tags = [],
}) => ({
  id: asset_id,
  name,
  description,
  status: status || 'ACTIVE',
  startDate: fromISO(start_date || acquired_at),
  completedDate: completed_date ? fromISO(completed_date) : '',
  targetCompletionDate: target_completion_date ? fromISO(target_completion_date) : '',
  targetPrice: target_price ? toPriceView(target_price) : '',
  targetCurrency: target_currency || currency || 'UAH',
  price: toPriceView(price || 0),
  currency: currency || target_currency || 'UAH',
  acquiredAt: fromISO(acquired_at || start_date),
  tags: tags || [],
});

const buildAssetPayloadDTO = ({
  name,
  description,
  status = 'ACTIVE',
  startDate,
  acquiredAt,
  completedDate,
  targetCompletionDate,
  clearTargetDate = false,
  targetPrice,
  targetCurrency,
  price,
  currency,
  withPayment,
  tags,
}, includeWithPayment = false) => {
  const chosenTarget = targetPrice ? toPrice(targetPrice) : 0;
  const chosenPrice = price ? toPrice(price) : chosenTarget;
  const chosenCurrency = targetCurrency || currency || 'UAH';
  const effectiveDate = startDate || acquiredAt || new Date();

  const dto = {
    name,
    description: description || '',
    status: status || 'ACTIVE',
    start_date: toISO(effectiveDate),
    acquired_at: toISO(effectiveDate),
    tags: Array.isArray(tags) ? tags : [],
  };

  if (chosenTarget > 0) {
    dto.target_price = chosenTarget;
    dto.target_currency = chosenCurrency;
  }
  if (chosenPrice > 0) {
    dto.price = chosenPrice;
    dto.currency = chosenCurrency;
  }
  if (includeWithPayment) {
    dto.with_payment = Boolean(withPayment);
  }
  if (completedDate) {
    dto.completed_date = toISO(completedDate);
  }
  if (targetCompletionDate) {
    dto.target_completion_date = toISO(targetCompletionDate);
  } else if (clearTargetDate) {
    dto.clear_target_date = true;
  }

  return dto;
};

const toAddAssetDTO = (payload) => buildAssetPayloadDTO(payload, true);

const toUpdateAssetDTO = (payload) => buildAssetPayloadDTO(payload, false);

export class AssetRepository {
  #request;

  constructor(request) {
    this.#request = request;
  }

  async getAssets(projectId, limit = 50, offset = 0) {
    const query = new URLSearchParams({ limit: String(limit), offset: String(offset) }).toString();
    const resourceUrl = `/projects/${projectId}/assets`;
    const url = query ? `${resourceUrl}?${query}` : resourceUrl;
    const response = await this.#request.get(url);

    const { assets, total } = response;
    return [assets.map(toDomain), total];
  }

  async getAsset(projectId, assetId) {
    const response = await this.#request.get(`/projects/${projectId}/assets/${assetId}`);
    return toEditAssetView(response);
  }

  async updateAsset(projectId, asset) {
    return await this.#request.put(
      `/projects/${projectId}/assets/${asset.id}`,
      toUpdateAssetDTO(asset),
    );
  }

  async addAsset(projectId, asset) {
    const response = await this.#request.post(`/projects/${projectId}/assets`, toAddAssetDTO(asset));
    return toDomain(response);
  }

  async linkChild(projectId, parentId, childId, sharePercentage = 100) {
    return await this.#request.post(
      `/projects/${projectId}/assets/${parentId}/children`,
      {
        child_asset_id: childId,
        share_percentage: Number(sharePercentage),
      },
    );
  }

  async unlinkChild(projectId, parentId, childId) {
    return await this.#request.delete(
      `/projects/${projectId}/assets/${parentId}/children/${childId}`,
    );
  }

  async createAssetFromPayments(projectId, {
    paymentIds, name, description, acquiredAt, startDate, targetCurrency,
  }) {
    const payload = {
      payment_ids: paymentIds,
      name,
      description,
    };
    if (acquiredAt || startDate) payload.acquired_at = toISO(acquiredAt || startDate);
    if (targetCurrency) payload.target_currency = targetCurrency;

    const response = await this.#request.post(`/projects/${projectId}/assets/from-payments`, payload);
    return toDomain(response);
  }

  async removeAsset(projectId, assetId) {
    return await this.#request.delete(`/projects/${projectId}/assets/${assetId}`);
  }

  async assignInvestments(projectId, assetId, investmentIds) {
    return await this.#request.post(
      `/projects/${projectId}/assets/${assetId}/investments`,
      { investment_ids: investmentIds },
    );
  }

  async groupAssets(projectId, { name, description, childAssetIds, tags }) {
    const response = await this.#request.post(`/projects/${projectId}/assets/group`, {
      name,
      description: description || '',
      child_asset_ids: childAssetIds,
      tags: Array.isArray(tags) ? tags : [],
    });
    return toDomain(response);
  }

  async linkChildren(projectId, parentId, childAssetIds) {
    return await this.#request.post(
      `/projects/${projectId}/assets/${parentId}/children/batch`,
      { child_asset_ids: childAssetIds },
    );
  }

  async rescheduleAsset(projectId, assetId, { startDate, targetCompletionDate, clearTargetDate = false }) {
    const payload = { clear_target_date: Boolean(clearTargetDate) };
    if (startDate) payload.start_date = toISO(startDate);
    if (clearTargetDate) {
      payload.target_completion_date = '';
    } else if (targetCompletionDate) {
      payload.target_completion_date = toISO(targetCompletionDate);
    }

    const res = await this.#request.patch(
      `/projects/${projectId}/assets/${assetId}/schedule`,
      payload,
    );
    return {
      asset: res?.asset ? toDomain(res.asset) : null,
      warnings: Array.isArray(res?.warnings) ? res.warnings : [],
    };
  }

  async getCashFlowForecast(projectId, { fromMonth, toMonth, tag } = {}) {
    const params = new URLSearchParams();
    if (fromMonth) params.set('from_month', fromMonth);
    if (toMonth) params.set('to_month', toMonth);
    if (tag) params.set('tag', tag);
    const qs = params.toString();
    const res = await this.#request.get(
      `/projects/${projectId}/cashflow-forecast${qs ? `?${qs}` : ''}`,
    );

    return {
      projectId: res?.project_id || projectId,
      currency: res?.currency || 'UAH',
      fromMonth: res?.from_month || '',
      toMonth: res?.to_month || '',
      trailingWindowMonths: res?.trailing_window_months || 6,
      summary: {
        totalTargetCents: res?.summary?.total_target_cents || 0,
        totalActualCashCents: res?.summary?.total_actual_cash_cents || 0,
        totalRemainingScheduledCents: res?.summary?.total_remaining_scheduled_cents || 0,
        totalTimeHours: res?.summary?.total_time_hours || 0,
        totalGoodsCents: res?.summary?.total_goods_cents || 0,
        estimatedOngoingMonthlyCashCents: res?.summary?.estimated_ongoing_monthly_cash_cents || 0,
        monthlyBudgetCeilingAvgCents: res?.summary?.monthly_budget_ceiling_avg_cents || 0,
        peakMonth: res?.summary?.peak_month || '',
        peakCommittedCashCents: res?.summary?.peak_committed_cash_cents || 0,
        alertMonthsCount: res?.summary?.alert_months_count || 0,
      },
      months: (res?.months || []).map((m) => ({
        month: m.month,
        tagBudgetCeilingCents: m.tag_budget_ceiling_cents || 0,
        actual: {
          cashCents: m.actual?.cash_cents || 0,
          timeHours: m.actual?.time_hours || 0,
          goodsCents: m.actual?.goods_cents || 0,
        },
        scheduledDemandCents: m.scheduled_demand_cents || 0,
        estimatedOngoingCashCents: m.estimated_ongoing_cash_cents || 0,
        committedCashCents: m.committed_cash_cents || 0,
        cashUtilizationPct: m.cash_utilization_pct || 0,
        alertStatus: m.alert_status || 'NO_CEILING',
        assetDemands: (m.asset_demands || []).map((d) => ({
          assetId: d.asset_id,
          assetName: d.asset_name,
          demandCents: d.demand_cents || 0,
          isOverdue: Boolean(d.is_overdue),
        })),
      })),
      assets: (res?.assets || []).map((a) => ({
        assetId: a.asset_id,
        assetName: a.asset_name,
        status: a.status || 'ACTIVE',
        startDate: a.start_date,
        targetCompletionDate: a.target_completion_date || null,
        completedDate: a.completed_date || null,
        hasTargetBudget: Boolean(a.has_target_budget),
        targetHomeCents: a.target_home_cents || 0,
        actualToDate: {
          cashCents: a.actual_to_date?.cash_cents || 0,
          timeHours: a.actual_to_date?.time_hours || 0,
          goodsCents: a.actual_to_date?.goods_cents || 0,
        },
        rolledUpTotalCashCents: a.rolled_up_total_cash_cents || 0,
        residualTargetCents: a.residual_target_cents || 0,
        burnRateMonthly: {
          cashCentsPerMonth: a.burn_rate_monthly?.cash_cents_per_month || 0,
          timeHoursPerMonth: a.burn_rate_monthly?.time_hours_per_month || 0,
          goodsCentsPerMonth: a.burn_rate_monthly?.goods_cents_per_month || 0,
        },
        forecastMode: a.forecast_mode || 'UNBUDGETED',
        estimatedCompletion: a.estimated_completion || null,
      })),
    };
  }
}

export default AssetRepository;
