/* eslint-disable import/extensions */
import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import {
  cleanTagName,
  parseDateSafe,
  getActiveCalendarMonths,
  FREQUENCY_STRATEGIES,
  TagBudgetRule,
  ProjectBudgetSchedule,
  BUDGET_FREQUENCIES,
  BUDGET_RESOURCE_TYPES,
} from './ProjectBudget.js';

describe('ProjectBudget Domain Model', () => {
  describe('cleanTagName', () => {
    it('normalizes tags by trimming, stripping hash and lowercasing', () => {
      assert.equal(cleanTagName(' #Labor '), 'labor');
      assert.equal(cleanTagName('###MATERIALS'), 'materials');
      assert.equal(cleanTagName(''), '');
      assert.equal(cleanTagName(null), '');
    });
  });

  describe('FREQUENCY_STRATEGIES', () => {
    it('ONCE only allocates to target calendar month', () => {
      const once = FREQUENCY_STRATEGIES.ONCE;
      assert.equal(once.getMonthlyAmount(500, { monthIndex: 0, targetMonth: 1 }), 500);
      assert.equal(once.getMonthlyAmount(500, { monthIndex: 1, targetMonth: 1 }), 0);
      assert.equal(once.getMonthlyAmount(500, { monthIndex: 11, targetMonth: 12 }), 500);
      assert.equal(once.getMonthlyAmount(500, { targetMonth: 1 }), 0); // monthIndex required
    });

    it('ANNUALLY allocates to target month in calendar year', () => {
      const annually = FREQUENCY_STRATEGIES.ANNUALLY;
      assert.equal(annually.getMonthlyAmount(20000, { monthIndex: 0, targetMonth: 1 }), 20000);
      assert.equal(annually.getMonthlyAmount(20000, { monthIndex: 1, targetMonth: 1 }), 0);
      assert.equal(annually.getMonthlyAmount(20000, { monthIndex: 11, targetMonth: 12 }), 20000);
      assert.equal(annually.getMonthlyAmount(20000, { monthIndex: 9, targetMonth: 12 }), 0);
    });

    it('MONTHLY allocates to every month', () => {
      const monthly = FREQUENCY_STRATEGIES.MONTHLY;
      assert.equal(monthly.getMonthlyAmount(100), 100);
    });

    it('WEEKLY calculates 52/12 average per month', () => {
      const weekly = FREQUENCY_STRATEGIES.WEEKLY;
      assert.equal(weekly.getMonthlyAmount(100), 433.33);
    });

    it('DAILY calculates based on days in month', () => {
      const daily = FREQUENCY_STRATEGIES.DAILY;
      // Jan has 31 days
      assert.equal(daily.getMonthlyAmount(10, { year: 2026, monthIndex: 0 }), 310);
      // Feb 2026 has 28 days
      assert.equal(daily.getMonthlyAmount(10, { year: 2026, monthIndex: 1 }), 280);
      // Feb 2024 (leap year) has 29 days
      assert.equal(daily.getMonthlyAmount(10, { year: 2024, monthIndex: 1 }), 290);
    });
  });

  describe('TagBudgetRule', () => {
    it('calculates effectiveAmount for MONEY', () => {
      const rule = new TagBudgetRule({
        tag: '#Equipment',
        frequency: BUDGET_FREQUENCIES.MONTHLY,
        amount: 2500,
        resourceType: BUDGET_RESOURCE_TYPES.MONEY,
      });
      assert.equal(rule.tag, 'equipment');
      assert.equal(rule.effectiveAmount, 2500);
      assert.equal(rule.isValid, true);
      assert.equal(rule.getAnnualTotal(new Date('2026-01-01')), 30000);
    });

    it('calculates effectiveAmount for TIME from hours and rate', () => {
      const rule = new TagBudgetRule({
        tag: 'labor',
        frequency: BUDGET_FREQUENCIES.WEEKLY,
        resourceType: BUDGET_RESOURCE_TYPES.TIME,
        timeHours: 20,
        timeHourlyRate: 25,
      });
      assert.equal(rule.effectiveAmount, 500); // 20 * 25
      assert.equal(rule.isValid, true);
      // Weekly 500 = 2166.67 per month * 12 = 26000.04
      assert.equal(rule.getAnnualTotal(new Date('2026-01-01')), 26000.04);
    });

    it('calculates effectiveAmount and getAnnualTotal for ANNUALLY frequency', () => {
      const rule = new TagBudgetRule({
        tag: 'materials',
        frequency: BUDGET_FREQUENCIES.ANNUALLY,
        targetMonth: 12,
        resourceType: BUDGET_RESOURCE_TYPES.MONEY,
        amount: 20000,
      });
      assert.equal(rule.effectiveAmount, 20000);
      assert.equal(rule.getAnnualTotal(), 20000);
      assert.equal(rule.getAnnualTotal('2026-10-25', null), 20000);
      // 3-month project (Oct 2026 to Dec 2026) includes December (month 12) => 20000
      assert.equal(rule.getAnnualTotal('2026-10-25', '2026-12-31'), 20000);
      // Project ending before December (Oct 2026 to Nov 2026) => 0
      assert.equal(rule.getAnnualTotal('2026-10-25', '2026-11-30'), 0);
    });

    it('serializes to DTO in minor cents and deserializes from DTO', () => {
      const rule = new TagBudgetRule({
        tag: 'materials',
        frequency: BUDGET_FREQUENCIES.ONCE,
        targetMonth: 3,
        resourceType: BUDGET_RESOURCE_TYPES.GOODS,
        amount: 1500.50,
        goodsQuantity: 10,
        goodsUnit: 'pcs',
        goodsItemName: 'Concrete',
        description: 'Foundation',
      });

      const dto = rule.toDTO();
      assert.equal(dto.tag, 'materials');
      assert.equal(dto.frequency, 'ONCE');
      assert.equal(dto.target_month, 3);
      assert.equal(dto.resource_type, 'GOODS');
      assert.equal(dto.amount, 150050); // Minor units
      assert.equal(dto.goods_quantity, 10);
      assert.equal(dto.goods_unit, 'pcs');
      assert.equal(dto.goods_item_name, 'Concrete');
      assert.equal(dto.description, 'Foundation');

      const reconstructed = TagBudgetRule.fromDTO(dto);
      assert.equal(reconstructed.tag, 'materials');
      assert.equal(reconstructed.amount, 1500.50);
      assert.equal(reconstructed.targetMonth, 3);
      assert.equal(reconstructed.resourceType, 'GOODS');
      assert.equal(reconstructed.goodsItemName, 'Concrete');
    });
  });

  describe('ProjectBudgetSchedule', () => {
    const startDate = new Date('2026-01-01');

    it('generates 12 months with correct projected totals', () => {
      const tagBudgets = [
        new TagBudgetRule({
          tag: 'rent',
          frequency: BUDGET_FREQUENCIES.MONTHLY,
          amount: 1000,
        }),
        new TagBudgetRule({
          tag: 'permits',
          frequency: BUDGET_FREQUENCIES.ONCE,
          targetMonth: 1,
          amount: 5000,
        }),
      ];

      const schedule = new ProjectBudgetSchedule({
        tagBudgets,
        payments: [],
        startDate,
      });

      assert.equal(schedule.months.length, 12);
      assert.equal(schedule.totalAnnualBudget, 17000); // 12000 (rent) + 5000 (permits)
      assert.equal(schedule.months[0].projected, 6000); // Month 1: 1000 + 5000
      assert.equal(schedule.months[1].projected, 1000); // Month 2: 1000
      assert.equal(schedule.months[0].cumulativeProjected, 6000);
      assert.equal(schedule.months[1].cumulativeProjected, 7000);
      assert.equal(schedule.months[11].cumulativeProjected, 17000);
    });

    it('splits payments across multiple tags and computes actuals', () => {
      const tagBudgets = [
        new TagBudgetRule({ tag: 'labor', frequency: BUDGET_FREQUENCIES.MONTHLY, amount: 2000 }),
        new TagBudgetRule({ tag: 'materials', frequency: BUDGET_FREQUENCIES.MONTHLY, amount: 3000 }),
      ];

      const payments = [
        // Payment in Month 1 (Jan 2026) split across 2 tags: 1000 each
        {
          rawDate: '2026-01-15T10:00:00Z',
          rawHomeAmount: 200000, // 2,000.00
          tags: ['labor', 'materials'],
        },
        // Unplanned payment in Month 2 (Feb 2026)
        {
          rawDate: '2026-02-10T10:00:00Z',
          homeAmount: 500,
          tags: ['unplanned-expense'],
        },
      ];

      const schedule = new ProjectBudgetSchedule({
        tagBudgets,
        payments,
        startDate,
      });

      assert.equal(schedule.totalActualInvested, 2500);

      // Month 1 actual
      assert.equal(schedule.months[0].actual, 2000);
      assert.equal(schedule.months[0].actualByTag.labor, 1000);
      assert.equal(schedule.months[0].actualByTag.materials, 1000);

      // Month 2 actual
      assert.equal(schedule.months[1].actual, 500);
      assert.equal(schedule.months[1].actualByTag['unplanned-expense'], 500);

      // Tag comparisons
      const laborComp = schedule.tagComparisons.find((c) => c.tag === 'labor');
      assert.ok(laborComp);
      assert.equal(laborComp.projectedAnnual, 24000);
      assert.equal(laborComp.actualAnnual, 1000);
      assert.equal(laborComp.isOverBudget, false);

      const unplannedComp = schedule.tagComparisons.find((c) => c.tag === 'unplanned-expense');
      assert.ok(unplannedComp);
      assert.equal(unplannedComp.isUnplanned, true);
      assert.equal(unplannedComp.actualAnnual, 500);
      assert.equal(unplannedComp.isOverBudget, true);
    });

    it('filters out payments outside the 12-month schedule window consistently', () => {
      const tagBudgets = [
        new TagBudgetRule({ tag: 'rent', frequency: BUDGET_FREQUENCIES.MONTHLY, amount: 1000 }),
      ];

      const payments = [
        { rawDate: '2025-12-31', amount: 999, tags: ['rent'] }, // Prior to schedule
        { rawDate: '2026-06-15', amount: 500, tags: ['rent'] }, // Inside schedule (Month 6)
        { rawDate: '2027-01-01', amount: 888, tags: ['rent'] }, // After schedule (Month 13)
      ];

      const schedule = new ProjectBudgetSchedule({
        tagBudgets,
        payments,
        startDate: '2026-01-01',
      });

      assert.equal(schedule.totalActualInvested, 500);
      const rentComp = schedule.tagComparisons.find((c) => c.tag === 'rent');
      assert.equal(rentComp.actualAnnual, 500);
    });

    it('supports Ukrainian locale for month labels', () => {
      const schedule = new ProjectBudgetSchedule({
        tagBudgets: [],
        payments: [],
        startDate: '2026-01-01',
        locale: 'uk',
      });

      assert.equal(schedule.months[0].label.includes('січ'), true);
      assert.equal(schedule.months[0].label.includes('2026'), true);
    });

    it('caps projected budget when project has a target endDate', () => {
      const tagBudgets = [
        new TagBudgetRule({ tag: 'car-saving', frequency: BUDGET_FREQUENCIES.MONTHLY, amount: 1000 }),
      ];

      // Ongoing project: 12 * 1000 = 12000
      const ongoingSchedule = new ProjectBudgetSchedule({
        tagBudgets,
        startDate: '2026-01-01',
      });
      assert.equal(ongoingSchedule.totalAnnualBudget, 12000);
      assert.equal(ongoingSchedule.months[0].projected, 1000);
      assert.equal(ongoingSchedule.months[11].projected, 1000);

      // Targeted project ending in month 6 (June 30, 2026): 6 * 1000 = 6000
      const targetedSchedule = new ProjectBudgetSchedule({
        tagBudgets,
        startDate: '2026-01-01',
        endDate: '2026-06-30',
      });
      assert.equal(targetedSchedule.totalAnnualBudget, 6000);
      assert.equal(targetedSchedule.months[0].projected, 1000); // Month 1 (Jan)
      assert.equal(targetedSchedule.months[5].projected, 1000); // Month 6 (Jun)
      assert.equal(targetedSchedule.months[6].projected, 0); // Month 7 (Jul) -> completed
      assert.equal(targetedSchedule.months[11].projected, 0); // Month 12 (Dec) -> completed
    });

    it('handles ANNUALLY tag budget placed in calendar target month for short project', () => {
      const tagBudgets = [
        new TagBudgetRule({
          tag: 'materials',
          frequency: BUDGET_FREQUENCIES.ANNUALLY,
          targetMonth: 12, // December
          amount: 20000,
        }),
      ];

      const schedule = new ProjectBudgetSchedule({
        tagBudgets,
        startDate: '2026-10-25',
        endDate: '2026-12-31',
      });

      assert.equal(schedule.totalAnnualBudget, 20000);
      assert.equal(schedule.months[0].projected, 0); // Oct 2026
      assert.equal(schedule.months[1].projected, 0); // Nov 2026
      assert.equal(schedule.months[2].projected, 20000); // Dec 2026
      assert.equal(schedule.months[3].projected, 0); // Jan 2027 (after end date)
    });

    it('excludes payments past endDate from tag comparisons and monthly actuals', () => {
      const tagBudgets = [
        new TagBudgetRule({
          tag: 'labor',
          frequency: BUDGET_FREQUENCIES.MONTHLY,
          amount: 1000,
        }),
      ];

      const payments = [
        { date: '2026-10-15', amount: 800, tags: ['labor'] },
        { date: '2026-11-20', amount: 900, tags: ['labor'] },
        // Payment after endDate (project ends Dec 31, 2026)
        { date: '2027-01-15', amount: 1500, tags: ['labor'] },
      ];

      const schedule = new ProjectBudgetSchedule({
        tagBudgets,
        payments,
        startDate: '2026-10-01',
        endDate: '2026-12-31',
      });

      // Oct, Nov, Dec -> 3 months * 1000 = 3000 projected
      assert.equal(schedule.totalAnnualBudget, 3000);
      // Total actual spend should only include Oct (800) + Nov (900) = 1700; Jan (1500) excluded!
      assert.equal(schedule.totalActualInvested, 1700);

      const laborComparison = schedule.tagComparisons.find((c) => c.tag === 'labor');
      assert.ok(laborComparison);
      assert.equal(laborComparison.projectedAnnual, 3000);
      assert.equal(laborComparison.actualAnnual, 1700); // 1500 after endDate ignored
      assert.equal(laborComparison.isOverBudget, false);
    });
  });

  describe('getActiveCalendarMonths', () => {
    it('returns all 12 calendar months when project has no endDate', () => {
      const active = getActiveCalendarMonths('2026-10-25', null);
      assert.equal(active.size, 12);
      for (let m = 1; m <= 12; m += 1) {
        assert.ok(active.has(m));
      }
    });

    it('returns only active months when project duration is less than 12 months', () => {
      const active = getActiveCalendarMonths('2026-10-25', '2026-12-31');
      assert.equal(active.size, 3);
      assert.ok(active.has(10)); // October
      assert.ok(active.has(11)); // November
      assert.ok(active.has(12)); // December
      assert.ok(!active.has(1)); // January is outside
      assert.ok(!active.has(3)); // March is outside
    });

    it('returns all 12 calendar months when project spans more than 12 months', () => {
      const active = getActiveCalendarMonths('2026-01-01', '2027-06-30');
      assert.equal(active.size, 12);
    });
  });

  describe('parseDateSafe', () => {
    it('parses YYYY-MM-DD string without timezone day shift', () => {
      const d = parseDateSafe('2026-05-15');
      assert.equal(d.getFullYear(), 2026);
      assert.equal(d.getMonth(), 4); // 0-indexed May
      assert.equal(d.getDate(), 15);
    });

    it('parses ISO 8601 strings to local calendar date without day shift', () => {
      const d = parseDateSafe('2026-10-25T00:00:00.000Z');
      assert.equal(d.getFullYear(), 2026);
      assert.equal(d.getMonth(), 9); // October
      assert.equal(d.getDate(), 25);
    });

    it('parses Date object or returns valid Date on fallback', () => {
      const now = new Date();
      assert.equal(parseDateSafe(now).getTime(), now.getTime());
      assert.ok(parseDateSafe(null) instanceof Date);
    });
  });
});
