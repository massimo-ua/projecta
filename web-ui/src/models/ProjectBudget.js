/* eslint-disable max-classes-per-file */
/**
 * Pure domain model for Project Tag Budgets and Annual 12-Month Schedules.
 * Following docs/frontend-design-principals.md (Layer 3: Domain Model).
 * ZERO imports from React or any UI framework.
 */

export const BUDGET_FREQUENCIES = Object.freeze({
  ONCE: 'ONCE',
  DAILY: 'DAILY',
  WEEKLY: 'WEEKLY',
  MONTHLY: 'MONTHLY',
  ANNUALLY: 'ANNUALLY',
});

export const BUDGET_RESOURCE_TYPES = Object.freeze({
  MONEY: 'MONEY',
  TIME: 'TIME',
  GOODS: 'GOODS',
});

export const PRESET_PROJECT_TAGS = Object.freeze([
  'labor',
  'materials',
  'equipment',
  'rent',
  'logistics',
  'services',
  'permits',
  'maintenance',
]);

/**
 * Normalizes a tag string by stripping leading '#', whitespace, and lowercasing.
 */
export function cleanTagName(tag) {
  if (!tag || typeof tag !== 'string') return '';
  return tag.trim().replace(/^#+/, '').trim().toLowerCase();
}

/**
 * Safely parses date-only YYYY-MM-DD or standard date strings without timezone shifts.
 */
export function parseDateSafe(input) {
  if (!input) return new Date();
  if (input instanceof Date) {
    return Number.isNaN(input.getTime()) ? new Date() : input;
  }
  if (typeof input === 'string') {
    const trimmed = input.trim();
    const match = trimmed.match(/^(\d{4})-(\d{1,2})(?:-(\d{1,2}))?/);
    if (match) {
      const year = parseInt(match[1], 10);
      const month = parseInt(match[2], 10) - 1;
      const day = match[3] ? parseInt(match[3], 10) : 1;
      return new Date(year, month, day);
    }
    const d = new Date(trimmed);
    return Number.isNaN(d.getTime()) ? new Date() : d;
  }
  return new Date();
}

/**
 * Calculates the active calendar months (1..12) for a project
 * spanning from startDate up to optional endDate.
 */
export function getActiveCalendarMonths(startDate, endDate) {
  const start = parseDateSafe(startDate);
  const end = endDate ? parseDateSafe(endDate) : null;
  const startYear = start.getFullYear();
  const startMonth = start.getMonth();
  const activeMonths = new Set();
  for (let i = 0; i < 12; i += 1) {
    const d = new Date(startYear, startMonth + i, 1);
    if (end && d > end) {
      break;
    }
    activeMonths.add(d.getMonth() + 1);
  }
  return activeMonths;
}

/**
 * Polymorphic strategy calculation for each frequency over a 12-month schedule.
 */
export const FREQUENCY_STRATEGIES = Object.freeze({
  ONCE: {
    getMonthlyAmount(amount, { targetMonth = 1, monthIndex } = {}) {
      const calendarMonth = typeof monthIndex === 'number' ? monthIndex + 1 : 0;
      return calendarMonth === targetMonth ? amount : 0;
    },
  },
  ANNUALLY: {
    getMonthlyAmount(amount, { targetMonth = 1, monthIndex } = {}) {
      const calendarMonth = typeof monthIndex === 'number' ? monthIndex + 1 : 0;
      return calendarMonth === targetMonth ? amount : 0;
    },
  },
  MONTHLY: {
    getMonthlyAmount(amount) {
      return amount;
    },
  },
  WEEKLY: {
    getMonthlyAmount(amount) {
      return Math.round(amount * (52 / 12) * 100) / 100;
    },
  },
  DAILY: {
    getMonthlyAmount(amount, { year, monthIndex }) {
      const daysInMonth = new Date(year, monthIndex + 1, 0).getDate();
      return Math.round(amount * daysInMonth * 100) / 100;
    },
  },
});

/**
 * Domain entity representing a planned budget rule for a single investment tag.
 */
export class TagBudgetRule {
  constructor({
    tag = '',
    frequency = BUDGET_FREQUENCIES.MONTHLY,
    targetMonth = 1,
    resourceType = BUDGET_RESOURCE_TYPES.MONEY,
    amount = 0,
    timeHours = null,
    timeHourlyRate = null,
    goodsQuantity = null,
    goodsUnit = '',
    goodsItemName = '',
    description = '',
    annualTotal = null,
  } = {}) {
    this.tag = cleanTagName(tag);
    this.frequency = String(frequency || BUDGET_FREQUENCIES.MONTHLY).toUpperCase();
    const parsedMonth = parseInt(targetMonth, 10);
    this.targetMonth = Number.isInteger(parsedMonth) && parsedMonth >= 1 && parsedMonth <= 12
      ? parsedMonth
      : 1;
    this.resourceType = String(resourceType || BUDGET_RESOURCE_TYPES.MONEY).toUpperCase();

    this.amount = typeof amount === 'number' ? amount : parseFloat(amount) || 0;
    this.timeHours = timeHours !== null && timeHours !== undefined && timeHours !== ''
      ? parseFloat(timeHours)
      : null;
    this.timeHourlyRate = timeHourlyRate !== null && timeHourlyRate !== undefined && timeHourlyRate !== ''
      ? parseFloat(timeHourlyRate)
      : null;
    this.goodsQuantity = goodsQuantity !== null && goodsQuantity !== undefined && goodsQuantity !== ''
      ? parseFloat(goodsQuantity)
      : null;
    this.goodsUnit = goodsUnit ? String(goodsUnit).trim() : '';
    this.goodsItemName = goodsItemName ? String(goodsItemName).trim() : '';
    this.description = description ? String(description).trim() : '';
    this.cachedAnnualTotal = annualTotal !== null && annualTotal !== undefined
      ? parseFloat(annualTotal)
      : null;
  }

  get displayTag() {
    return this.tag ? `#${this.tag}` : '';
  }

  get effectiveAmount() {
    if (
      this.resourceType === BUDGET_RESOURCE_TYPES.TIME
      && this.timeHours > 0
      && this.timeHourlyRate > 0
    ) {
      return Math.round(this.timeHours * this.timeHourlyRate * 100) / 100;
    }
    return this.amount > 0 ? this.amount : 0;
  }

  get isValid() {
    return Boolean(this.tag && this.effectiveAmount > 0);
  }

  getMonthlyProjection(scheduleIndex, year, monthIndex) {
    const strategy = FREQUENCY_STRATEGIES[this.frequency] || FREQUENCY_STRATEGIES.MONTHLY;
    return strategy.getMonthlyAmount(this.effectiveAmount, {
      scheduleIndex,
      targetMonth: this.targetMonth,
      year,
      monthIndex,
    });
  }

  getAnnualTotal(startDate = null, endDate = null) {
    if (this.cachedAnnualTotal !== null && this.cachedAnnualTotal >= 0 && !startDate && !endDate) {
      return this.cachedAnnualTotal;
    }
    const validDate = parseDateSafe(startDate);
    const parsedEnd = endDate ? parseDateSafe(endDate) : null;
    const startYear = validDate.getFullYear();
    const startMonth = validDate.getMonth();
    let total = 0;
    for (let i = 0; i < 12; i += 1) {
      const d = new Date(startYear, startMonth + i, 1);
      if (parsedEnd && d > parsedEnd) {
        break;
      }
      total += this.getMonthlyProjection(i + 1, d.getFullYear(), d.getMonth());
    }
    return Math.round(total * 100) / 100;
  }

  toDTO() {
    return {
      tag: this.tag,
      frequency: this.frequency,
      target_month: this.targetMonth,
      resource_type: this.resourceType,
      amount: Math.round(this.effectiveAmount * 100),
      time_hours: this.timeHours,
      time_hourly_rate: this.timeHourlyRate !== null ? Math.round(this.timeHourlyRate * 100) : null,
      goods_quantity: this.goodsQuantity,
      goods_unit: this.goodsUnit || '',
      goods_item_name: this.goodsItemName || '',
      description: this.description || '',
    };
  }

  static fromDTO(dto) {
    if (!dto) return null;
    return new TagBudgetRule({
      tag: dto.tag,
      frequency: dto.frequency,
      targetMonth: dto.target_month || 1,
      resourceType: dto.resource_type || BUDGET_RESOURCE_TYPES.MONEY,
      amount: typeof dto.amount === 'number' ? dto.amount / 100 : 0,
      timeHours: dto.time_hours ?? null,
      timeHourlyRate: typeof dto.time_hourly_rate === 'number' ? dto.time_hourly_rate / 100 : null,
      goodsQuantity: dto.goods_quantity ?? null,
      goodsUnit: dto.goods_unit || '',
      goodsItemName: dto.goods_item_name || '',
      description: dto.description || '',
      annualTotal: typeof dto.annual_total === 'number' ? dto.annual_total / 100 : null,
    });
  }
}

/**
 * Domain service calculating 12-month projected vs. actual schedules and tag comparisons.
 */
export class ProjectBudgetSchedule {
  constructor({
    tagBudgets = [],
    payments = [],
    startDate = null,
    endDate = null,
    locale = 'en',
  } = {}) {
    this.tagRules = Array.isArray(tagBudgets)
      ? tagBudgets
        .map((tb) => (tb instanceof TagBudgetRule ? tb : new TagBudgetRule(tb)))
        .filter((tb) => tb && tb.isValid)
      : [];

    this.startDate = parseDateSafe(startDate);
    this.endDate = endDate ? parseDateSafe(endDate) : null;
    this.locale = locale || 'en';

    const parsedPayments = ProjectBudgetSchedule.#parsePayments(payments);
    this.months = this.#calculateMonths(parsedPayments);
    this.tagComparisons = this.#calculateTagComparisons(parsedPayments);

    this.totalAnnualBudget = Math.round(
      this.tagRules.reduce(
        (sum, tb) => sum + tb.getAnnualTotal(this.startDate, this.endDate),
        0,
      ) * 100,
    ) / 100;

    this.totalActualInvested = Math.round(
      this.months.reduce((sum, m) => sum + m.actual, 0) * 100,
    ) / 100;

    this.annualVariance = Math.round(
      (this.totalAnnualBudget - this.totalActualInvested) * 100,
    ) / 100;
    this.budgetUtilization = this.totalAnnualBudget > 0
      ? Math.round((this.totalActualInvested / this.totalAnnualBudget) * 100)
      : 0;

    this.overBudgetTagsCount = this.tagComparisons.filter((tc) => tc.isOverBudget).length;
    this.isOverAnnualBudget = this.totalActualInvested > this.totalAnnualBudget;
  }

  static #parsePayments(payments) {
    return (Array.isArray(payments) ? payments : []).map((p) => {
      const rawDateStr = p.paymentDate || p.date || p.rawDate;
      const parsed = rawDateStr ? parseDateSafe(rawDateStr) : null;
      const valid = parsed && !Number.isNaN(parsed.getTime());
      const amount = typeof p.rawHomeAmount === 'number'
        ? p.rawHomeAmount / 100
        : parseFloat(String(p.homeAmount || p.amount || 0).replace(/,/g, '')) || 0;

      let tagList = [];
      if (Array.isArray(p.tags) && p.tags.length > 0) {
        tagList = p.tags.map(cleanTagName).filter(Boolean);
      } else if (p.category && typeof p.category === 'string') {
        tagList = p.category.split(',').map(cleanTagName).filter(Boolean);
      }

      return {
        date: valid ? parsed : null,
        amount,
        tags: tagList,
      };
    });
  }

  #calculateMonths(parsedPayments) {
    const startYear = this.startDate.getFullYear();
    const startMonth = this.startDate.getMonth();
    const startWindow = new Date(startYear, startMonth, 1);
    const endWindow = new Date(startYear, startMonth + 12, 1);
    const endCap = this.endDate
      ? new Date(
        this.endDate.getFullYear(),
        this.endDate.getMonth(),
        this.endDate.getDate(),
        23,
        59,
        59,
        999,
      )
      : null;
    const intlLocale = (this.locale && this.locale.startsWith('uk')) ? 'uk-UA' : 'en-US';

    const monthList = [];
    let cumulativeProjected = 0;
    let cumulativeActual = 0;

    for (let i = 0; i < 12; i += 1) {
      const d = new Date(startYear, startMonth + i, 1);
      const year = d.getFullYear();
      const monthIndex = d.getMonth();
      const scheduleIndex = i + 1;
      const key = `${year}-${String(monthIndex + 1).padStart(2, '0')}`;
      const label = d.toLocaleString(intlLocale, { month: 'short', year: 'numeric' });

      // Calculate projected by tag
      const projectedByTag = {};
      let projectedSum = 0;
      const isPastEndDate = this.endDate && d > this.endDate;

      this.tagRules.forEach((rule) => {
        const amt = isPastEndDate ? 0 : rule.getMonthlyProjection(scheduleIndex, year, monthIndex);
        projectedByTag[rule.tag] = amt;
        projectedSum += amt;
      });

      // Calculate actual by tag for payments in this calendar month
      const actualByTag = {};
      let actualSum = 0;

      parsedPayments.forEach((p) => {
        if (!p.date || p.date < startWindow || p.date >= endWindow) return;
        if (endCap && p.date > endCap) return;
        if (p.date.getFullYear() === year && p.date.getMonth() === monthIndex) {
          actualSum += p.amount;
          if (p.tags.length === 0) {
            actualByTag.unplanned = (actualByTag.unplanned || 0) + p.amount;
          } else {
            const share = p.amount / p.tags.length;
            p.tags.forEach((t) => {
              actualByTag[t] = (actualByTag[t] || 0) + share;
            });
          }
        }
      });

      projectedSum = Math.round(projectedSum * 100) / 100;
      actualSum = Math.round(actualSum * 100) / 100;
      cumulativeProjected = Math.round((cumulativeProjected + projectedSum) * 100) / 100;
      cumulativeActual = Math.round((cumulativeActual + actualSum) * 100) / 100;

      monthList.push({
        scheduleIndex,
        key,
        label,
        year,
        monthIndex,
        projected: projectedSum,
        actual: actualSum,
        cumulativeProjected,
        cumulativeActual,
        variance: Math.round((projectedSum - actualSum) * 100) / 100,
        isOverBudget: actualSum > projectedSum,
        byTag: projectedByTag,
        actualByTag,
      });
    }

    return monthList;
  }

  #calculateTagComparisons(parsedPayments) {
    const startYear = this.startDate.getFullYear();
    const startMonth = this.startDate.getMonth();
    const startWindow = new Date(startYear, startMonth, 1);
    const endWindow = new Date(startYear, startMonth + 12, 1);
    const endCap = this.endDate
      ? new Date(
        this.endDate.getFullYear(),
        this.endDate.getMonth(),
        this.endDate.getDate(),
        23,
        59,
        59,
        999,
      )
      : null;

    const tagTotals = new Map();

    parsedPayments.forEach((p) => {
      if (!p.date || p.date < startWindow || p.date >= endWindow) {
        return; // Filter strictly to the 12-month schedule window
      }
      if (endCap && p.date > endCap) {
        return; // Exclude payments after the project end date
      }

      const { amount } = p;
      const tagList = p.tags;

      if (tagList.length === 0) {
        tagTotals.set('unplanned', (tagTotals.get('unplanned') || 0) + amount);
      } else {
        const share = amount / tagList.length;
        tagList.forEach((t) => {
          tagTotals.set(t, (tagTotals.get(t) || 0) + share);
        });
      }
    });

    const comparisons = [];
    const processedTags = new Set();

    this.tagRules.forEach((rule) => {
      processedTags.add(rule.tag);
      const projected = rule.getAnnualTotal(this.startDate, this.endDate);
      const actual = Math.round((tagTotals.get(rule.tag) || 0) * 100) / 100;
      const variance = Math.round((projected - actual) * 100) / 100;
      const utilizationPercentage = projected > 0
        ? Math.round((actual / projected) * 100)
        : 0;

      comparisons.push({
        tag: rule.tag,
        displayTag: `#${rule.tag}`,
        rule,
        frequency: rule.frequency,
        resourceType: rule.resourceType,
        projectedAnnual: projected,
        actualAnnual: actual,
        variance,
        utilizationPercentage,
        isOverBudget: actual > projected,
        isUnplanned: false,
      });
    });

    // Handle unplanned / untagged or tags without a budget rule
    tagTotals.forEach((actualRaw, tag) => {
      if (processedTags.has(tag)) return;
      const actual = Math.round(actualRaw * 100) / 100;
      if (actual <= 0) return;

      comparisons.push({
        tag,
        displayTag: tag === 'unplanned' ? 'unplanned' : `#${tag}`,
        rule: null,
        frequency: BUDGET_FREQUENCIES.ONCE,
        resourceType: BUDGET_RESOURCE_TYPES.MONEY,
        projectedAnnual: 0,
        actualAnnual: actual,
        variance: -actual,
        utilizationPercentage: 100,
        isOverBudget: true,
        isUnplanned: true,
      });
    });

    return comparisons;
  }
}
