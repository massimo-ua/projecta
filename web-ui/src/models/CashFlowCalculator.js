const MS_PER_DAY = 24 * 60 * 60 * 1000;
const AT_RISK_THRESHOLD_RATIO = 0.85;

export function toUtcCalendarDay(dateInput) {
  if (!dateInput) return null;
  const d = dateInput instanceof Date ? dateInput : new Date(dateInput);
  if (Number.isNaN(d.getTime())) return null;
  return new Date(Date.UTC(d.getUTCFullYear(), d.getUTCMonth(), d.getUTCDate()));
}

export function utcMonthStart(dateInput) {
  const day = toUtcCalendarDay(dateInput);
  if (!day) return null;
  return new Date(Date.UTC(day.getUTCFullYear(), day.getUTCMonth(), 1));
}

export function parseMonthKey(monthStr) {
  if (typeof monthStr !== 'string') return null;
  const match = /^(\d{4})-(\d{2})$/.exec(monthStr.trim());
  if (!match) return null;
  const year = Number(match[1]);
  const monthIdx = Number(match[2]) - 1;
  if (monthIdx < 0 || monthIdx > 11) return null;
  return new Date(Date.UTC(year, monthIdx, 1));
}

export function formatUtcMonthKey(dateInput) {
  const d = utcMonthStart(dateInput);
  if (!d) return '';
  const year = d.getUTCFullYear();
  const month = String(d.getUTCMonth() + 1).padStart(2, '0');
  return `${year}-${month}`;
}

export function distributeExactCentsByCalendarDays(totalCents, startInput, endInput) {
  const cents = Math.trunc(Number(totalCents) || 0);
  const startDay = toUtcCalendarDay(startInput);
  const endDay = toUtcCalendarDay(endInput);
  if (cents <= 0 || !startDay || !endDay || endDay < startDay) {
    return {};
  }

  const totalDays = Math.floor((endDay.getTime() - startDay.getTime()) / MS_PER_DAY) + 1;
  const slices = [];

  let curMonth = utcMonthStart(startDay);
  const endMonth = utcMonthStart(endDay);

  while (curMonth <= endMonth) {
    const monthLastDay = new Date(
      Date.UTC(curMonth.getUTCFullYear(), curMonth.getUTCMonth() + 1, 0),
    );
    const overlapStart = startDay > curMonth ? startDay : curMonth;
    const overlapEnd = endDay < monthLastDay ? endDay : monthLastDay;

    if (overlapEnd >= overlapStart) {
      const days = Math.floor((overlapEnd.getTime() - overlapStart.getTime()) / MS_PER_DAY) + 1;
      const numer = cents * days;
      slices.push({
        monthKey: formatUtcMonthKey(curMonth),
        alloc: Math.floor(numer / totalDays),
        remainder: numer % totalDays,
      });
    }

    curMonth = new Date(Date.UTC(curMonth.getUTCFullYear(), curMonth.getUTCMonth() + 1, 1));
  }

  const sumAllocated = slices.reduce((acc, s) => acc + s.alloc, 0);
  const leftover = cents - sumAllocated;

  if (leftover > 0) {
    const order = slices.map((_, idx) => idx);
    order.sort((a, b) => {
      if (slices[b].remainder !== slices[a].remainder) {
        return slices[b].remainder - slices[a].remainder;
      }
      return a - b;
    });
    for (let i = 0; i < leftover && i < order.length; i += 1) {
      slices[order[i]].alloc += 1;
    }
  }

  return slices.reduce((acc, s) => {
    acc[s.monthKey] = s.alloc;
    return acc;
  }, {});
}

export function evaluateCeilingAlert(committedCents, ceilingCents) {
  const ceiling = Number(ceilingCents) || 0;
  const committed = Number(committedCents) || 0;
  if (ceiling <= 0) {
    return { alertStatus: 'NO_CEILING', utilizationPct: 0 };
  }
  const utilizationPct = Math.round(((committed / ceiling) * 100) * 10) / 10;
  if (committed > ceiling) {
    return { alertStatus: 'OVER_CEILING', utilizationPct };
  }
  if (committed >= ceiling * AT_RISK_THRESHOLD_RATIO) {
    return { alertStatus: 'AT_RISK', utilizationPct };
  }
  return { alertStatus: 'HEALTHY', utilizationPct };
}

export function buildChartSeries(forecast, { includeOngoingInCommitted = false } = {}) {
  const months = forecast?.months || [];
  return months.map((m) => {
    const actualCashCents = Number(m.actual?.cashCents || 0);
    const scheduledCents = Number(m.scheduledDemandCents || 0);
    const ongoingCents = Number(m.estimatedOngoingCashCents || 0);
    const ceilingCents = Number(m.tagBudgetCeilingCents || 0);
    const baseCommitted = actualCashCents + scheduledCents;
    const effectiveTotal = includeOngoingInCommitted ? baseCommitted + ongoingCents : baseCommitted;
    const alertEval = evaluateCeilingAlert(effectiveTotal, ceilingCents);

    return {
      month: m.month,
      actualCashCents,
      scheduledDemandCents: scheduledCents,
      estimatedOngoingCashCents: ongoingCents,
      committedCashCents: effectiveTotal,
      ceilingCents,
      actualCash: actualCashCents / 100,
      scheduledDemand: scheduledCents / 100,
      estimatedOngoing: ongoingCents / 100,
      ceiling: ceilingCents / 100,
      timeHours: Number(m.actual?.timeHours || 0),
      goods: Number(m.actual?.goodsCents || 0) / 100,
      alertStatus: alertEval.alertStatus,
      utilizationPct: alertEval.utilizationPct,
      assetDemands: m.assetDemands || [],
    };
  });
}

export function calculateInstantSchedulePreview(
  forecast,
  assetId,
  { startDate, targetCompletionDate, clearTargetDate = false } = {},
  { asOfDate = new Date() } = {},
) {
  if (!forecast || !Array.isArray(forecast.months)) return forecast;

  const targetAsset = (forecast.assets || []).find((a) => a.assetId === assetId);
  if (!targetAsset) return forecast;

  const asOfDay = toUtcCalendarDay(asOfDate) || toUtcCalendarDay(new Date());
  const currentMonth = utcMonthStart(asOfDay);
  const currentMonthKey = formatUtcMonthKey(currentMonth);

  const nextStart = startDate
    ? toUtcCalendarDay(startDate)
    : toUtcCalendarDay(targetAsset.startDate);
  let nextTarget = null;
  if (!clearTargetDate) {
    nextTarget = targetCompletionDate
      ? toUtcCalendarDay(targetCompletionDate)
      : toUtcCalendarDay(targetAsset.targetCompletionDate);
  }

  let newAllocations = {};
  let isOverdue = false;

  if (targetAsset.status === 'ACTIVE' && nextTarget && targetAsset.residualTargetCents > 0) {
    if (nextTarget < currentMonth) {
      isOverdue = true;
      newAllocations = { [currentMonthKey]: targetAsset.residualTargetCents };
    } else {
      const effectiveStart = nextStart && nextStart > asOfDay ? nextStart : asOfDay;
      const boundedStart = effectiveStart > nextTarget ? nextTarget : effectiveStart;
      newAllocations = distributeExactCentsByCalendarDays(
        targetAsset.residualTargetCents,
        boundedStart,
        nextTarget,
      );
    }
  }

  const nextMonths = forecast.months.map((m) => {
    const existingWithoutAsset = (m.assetDemands || []).filter((d) => d.assetId !== assetId);
    const removedCents = (m.assetDemands || [])
      .filter((d) => d.assetId === assetId)
      .reduce((sum, d) => sum + (Number(d.demandCents) || 0), 0);
    const addedCents = newAllocations[m.month] || 0;

    const currentScheduled = Number(m.scheduledDemandCents) || 0;
    const scheduledDemandCents = Math.max(0, currentScheduled - removedCents + addedCents);
    const committedCashCents = (Number(m.actual?.cashCents) || 0) + scheduledDemandCents;
    const { alertStatus, utilizationPct } = evaluateCeilingAlert(
      committedCashCents,
      m.tagBudgetCeilingCents,
    );

    const nextDemands = addedCents > 0
      ? [
        ...existingWithoutAsset,
        {
          assetId,
          assetName: targetAsset.assetName,
          demandCents: addedCents,
          isOverdue,
        },
      ]
      : existingWithoutAsset;

    return {
      ...m,
      scheduledDemandCents,
      committedCashCents,
      cashUtilizationPct: utilizationPct,
      alertStatus,
      assetDemands: nextDemands,
    };
  });

  return {
    ...forecast,
    months: nextMonths,
  };
}
