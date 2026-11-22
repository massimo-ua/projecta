/* eslint-disable import/extensions */
import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import {
  toUtcCalendarDay,
  formatUtcMonthKey,
  distributeExactCentsByCalendarDays,
  evaluateCeilingAlert,
  buildChartSeries,
  calculateInstantSchedulePreview,
} from './CashFlowCalculator.js';

describe('CashFlowCalculator Domain Model', () => {
  it('normalizes dates to UTC calendar days safely', () => {
    const lateUtc = toUtcCalendarDay('2026-05-31T23:30:00Z');
    assert.equal(formatUtcMonthKey(lateUtc), '2026-05');
    assert.equal(toUtcCalendarDay('not-a-date'), null);
  });

  it('conserves exact integer cents across multi-month spans', () => {
    const dist = distributeExactCentsByCalendarDays(
      10000,
      '2026-05-10',
      '2026-07-10',
    );
    const keys = Object.keys(dist).sort();
    assert.deepEqual(keys, ['2026-05', '2026-06', '2026-07']);
    const sum = Object.values(dist).reduce((a, b) => a + b, 0);
    assert.equal(sum, 10000);
  });

  it('evaluates ceiling alert thresholds cleanly', () => {
    assert.deepEqual(evaluateCeilingAlert(5000, 0), { alertStatus: 'NO_CEILING', utilizationPct: 0 });
    assert.deepEqual(evaluateCeilingAlert(5000, 10000), { alertStatus: 'HEALTHY', utilizationPct: 50 });
    assert.deepEqual(evaluateCeilingAlert(8500, 10000), { alertStatus: 'AT_RISK', utilizationPct: 85 });
    assert.deepEqual(evaluateCeilingAlert(10001, 10000), { alertStatus: 'OVER_CEILING', utilizationPct: 100 });
  });

  it('builds chart series with optional ongoing burn stacking', () => {
    const forecast = {
      months: [
        {
          month: '2026-05',
          tagBudgetCeilingCents: 10000,
          actual: { cashCents: 3000, timeHours: 4.5, goodsCents: 500 },
          scheduledDemandCents: 4000,
          estimatedOngoingCashCents: 2000,
          assetDemands: [],
        },
      ],
    };

    const committedOnly = buildChartSeries(forecast, {
      includeOngoingInCommitted: false,
      asOfDate: '2026-05-10',
    });
    assert.equal(committedOnly[0].alertStatus, 'HEALTHY');
    assert.equal(committedOnly[0].committedCashCents, 7000);
    assert.equal(committedOnly[0].actualCash, 30);

    const stacked = buildChartSeries(forecast, {
      includeOngoingInCommitted: true,
      asOfDate: '2026-05-10',
    });
    assert.equal(stacked[0].alertStatus, 'AT_RISK');
    assert.equal(stacked[0].committedCashCents, 9000);
  });

  it('keeps all history plus current month and 5 months ahead across year boundary', () => {
    const monthKeys = [
      '2024-10',
      '2025-12',
      '2026-10',
      '2026-11',
      '2026-12',
      '2027-01',
      '2027-02',
      '2027-03',
      '2027-04',
      '2027-05',
      '2028-01',
    ];
    const forecast = {
      months: monthKeys.map((month) => ({
        month,
        tagBudgetCeilingCents: 10000,
        actual: { cashCents: 1000, timeHours: 0, goodsCents: 0 },
        scheduledDemandCents: 2000,
        estimatedOngoingCashCents: 500,
        assetDemands: [],
      })),
    };

    const series = buildChartSeries(forecast, { asOfDate: '2026-11-22' });
    assert.deepEqual(
      series.map((row) => row.month),
      [
        '2024-10',
        '2025-12',
        '2026-10',
        '2026-11',
        '2026-12',
        '2027-01',
        '2027-02',
        '2027-03',
        '2027-04',
      ],
    );
  });

  it('handles end-of-month anchor dates and ignores malformed month keys safely', () => {
    const forecast = {
      months: [
        { month: '2025-12' },
        { month: '2026-01' },
        { month: '2026-02' },
        { month: '2026-03' },
        { month: '2026-04' },
        { month: '2026-05' },
        { month: '2026-06' },
        { month: '2026-07' },
        { month: 'not-a-month' },
        { month: '2026-13' },
        { month: null },
      ],
    };

    const series = buildChartSeries(forecast, { asOfDate: '2026-01-31T23:59:59Z' });
    assert.deepEqual(
      series.map((row) => row.month),
      ['2025-12', '2026-01', '2026-02', '2026-03', '2026-04', '2026-05', '2026-06'],
    );
  });

  it('falls back to current UTC month when asOfDate is invalid and handles empty forecast', () => {
    assert.deepEqual(buildChartSeries(null, { asOfDate: '2026-11-22' }), []);

    const currentMonthKey = formatUtcMonthKey(new Date());
    const series = buildChartSeries(
      { months: [{ month: '1999-01' }, { month: currentMonthKey }, { month: '2099-12' }] },
      { asOfDate: 'invalid-date' },
    );
    assert.deepEqual(series.map((row) => row.month), ['1999-01', currentMonthKey]);
  });

  it('recomputes schedule preview optimistically when shifting target date', () => {
    const forecast = {
      months: [
        {
          month: '2026-05',
          tagBudgetCeilingCents: 10000,
          actual: { cashCents: 0, timeHours: 0, goodsCents: 0 },
          scheduledDemandCents: 6000,
          estimatedOngoingCashCents: 0,
          committedCashCents: 6000,
          alertStatus: 'HEALTHY',
          assetDemands: [{ assetId: 'a1', assetName: 'Kitchen', demandCents: 6000, isOverdue: false }],
        },
        {
          month: '2026-06',
          tagBudgetCeilingCents: 10000,
          actual: { cashCents: 0, timeHours: 0, goodsCents: 0 },
          scheduledDemandCents: 0,
          estimatedOngoingCashCents: 0,
          committedCashCents: 0,
          alertStatus: 'HEALTHY',
          assetDemands: [],
        },
      ],
      assets: [
        {
          assetId: 'a1',
          assetName: 'Kitchen',
          status: 'ACTIVE',
          startDate: '2026-05-01T00:00:00Z',
          targetCompletionDate: '2026-05-31T00:00:00Z',
          residualTargetCents: 6000,
        },
      ],
    };

    const preview = calculateInstantSchedulePreview(
      forecast,
      'a1',
      { targetCompletionDate: '2026-06-30' },
      { asOfDate: '2026-05-01' },
    );
    const total = preview.months.reduce((s, m) => s + m.scheduledDemandCents, 0);
    assert.equal(total, 6000);
    assert.ok(preview.months[1].scheduledDemandCents > 2500);
  });
});
