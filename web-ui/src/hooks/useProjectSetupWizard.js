import { useState, useMemo, useCallback } from 'react';
import { useLocale } from 'react-intlayer';
import {
  TagBudgetRule,
  ProjectBudgetSchedule,
  BUDGET_FREQUENCIES,
  BUDGET_RESOURCE_TYPES,
  cleanTagName,
  getActiveCalendarMonths,
} from '../models';
import { projectsRepository } from '../api';

const INITIAL_DRAFT = Object.freeze({
  tag: '',
  frequency: BUDGET_FREQUENCIES.MONTHLY,
  targetMonth: 1,
  resourceType: BUDGET_RESOURCE_TYPES.MONEY,
  amount: '',
  timeHours: '',
  timeHourlyRate: '',
  goodsQuantity: '',
  goodsUnit: '',
  goodsItemName: '',
  description: '',
});

export function useProjectSetupWizard() {
  const { locale } = useLocale();
  const [step, setStep] = useState(1);
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [mainCurrency, setMainCurrency] = useState('UAH');
  const [startDate, setStartDate] = useState(() => {
    const today = new Date();
    return today.toISOString().split('T')[0];
  });
  const [endDate, setEndDate] = useState('');
  const [tagBudgets, setTagBudgets] = useState([]);
  const [draftRule, setDraftRule] = useState(INITIAL_DRAFT);
  const [editingTag, setEditingTag] = useState(null);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState(null);

  const activeMonths = useMemo(
    () => getActiveCalendarMonths(startDate, endDate || null),
    [startDate, endDate],
  );

  const schedulePreview = useMemo(() => new ProjectBudgetSchedule({
    tagBudgets,
    payments: [],
    startDate,
    endDate: endDate || null,
    locale,
  }), [tagBudgets, startDate, endDate, locale]);

  const canGoNext = useMemo(() => {
    if (step === 1) {
      if (name.trim().length < 3) return false;
      if (endDate && startDate && endDate < startDate) return false;
      return true;
    }
    return true;
  }, [step, name, startDate, endDate]);

  const isDraftValid = useMemo(() => {
    const cleaned = cleanTagName(draftRule.tag);
    if (!cleaned) return false;
    if (draftRule.resourceType === BUDGET_RESOURCE_TYPES.TIME) {
      const hours = parseFloat(draftRule.timeHours) || 0;
      const rate = parseFloat(draftRule.timeHourlyRate) || 0;
      return hours > 0 && rate > 0;
    }
    const amt = parseFloat(draftRule.amount) || 0;
    return amt > 0;
  }, [draftRule]);

  const prevStep = useCallback(() => {
    if (step > 1) {
      setStep((prev) => prev - 1);
    }
  }, [step]);

  const goToStep = useCallback((target) => {
    if (target >= 1 && target <= 3) {
      setStep(target);
    }
  }, []);

  const updateDraft = useCallback((fields) => {
    setDraftRule((prev) => ({ ...prev, ...fields }));
  }, []);

  const resetDraft = useCallback(() => {
    setDraftRule(INITIAL_DRAFT);
    setEditingTag(null);
  }, []);

  const selectPresetTag = useCallback((preset) => {
    setDraftRule((prev) => ({
      ...prev,
      tag: preset,
    }));
  }, []);

  const startEditTag = useCallback((tagToEdit) => {
    const existing = tagBudgets.find((t) => t.tag === tagToEdit);
    if (!existing) return;
    setEditingTag(tagToEdit);
    setDraftRule({
      tag: existing.tag,
      frequency: existing.frequency,
      targetMonth: existing.targetMonth,
      resourceType: existing.resourceType,
      amount: existing.amount ? String(existing.amount) : '',
      timeHours: existing.timeHours !== null ? String(existing.timeHours) : '',
      timeHourlyRate: existing.timeHourlyRate !== null ? String(existing.timeHourlyRate) : '',
      goodsQuantity: existing.goodsQuantity !== null ? String(existing.goodsQuantity) : '',
      goodsUnit: existing.goodsUnit || '',
      goodsItemName: existing.goodsItemName || '',
      description: existing.description || '',
    });
  }, [tagBudgets]);

  const addOrUpdateTagRule = useCallback(() => {
    const cleaned = cleanTagName(draftRule.tag);
    if (!cleaned) return false;

    let ruleAmount = parseFloat(draftRule.amount) || 0;
    const hours = parseFloat(draftRule.timeHours) || null;
    const rate = parseFloat(draftRule.timeHourlyRate) || null;

    if (draftRule.resourceType === BUDGET_RESOURCE_TYPES.TIME && hours && rate) {
      ruleAmount = Math.round(hours * rate * 100) / 100;
    }

    if (ruleAmount <= 0) return false;

    const newRule = new TagBudgetRule({
      tag: cleaned,
      frequency: draftRule.frequency,
      targetMonth: draftRule.targetMonth,
      resourceType: draftRule.resourceType,
      amount: ruleAmount,
      timeHours: hours,
      timeHourlyRate: rate,
      goodsQuantity: parseFloat(draftRule.goodsQuantity) || null,
      goodsUnit: draftRule.goodsUnit,
      goodsItemName: draftRule.goodsItemName,
      description: draftRule.description,
    });

    setTagBudgets((prev) => {
      const matchIndex = prev.findIndex(
        (t) => t.tag === cleaned || (editingTag && t.tag === editingTag),
      );
      if (matchIndex >= 0) {
        const copy = [...prev];
        copy[matchIndex] = newRule;
        return copy;
      }
      return [...prev, newRule];
    });

    resetDraft();
    return true;
  }, [draftRule, editingTag, resetDraft]);

  const nextStep = useCallback(() => {
    if (step === 2 && isDraftValid) {
      addOrUpdateTagRule();
    }
    if (step < 3 && canGoNext) {
      setStep((prev) => prev + 1);
    }
  }, [step, canGoNext, isDraftValid, addOrUpdateTagRule]);

  const resetWizard = useCallback(() => {
    setStep(1);
    setName('');
    setDescription('');
    setMainCurrency('UAH');
    setStartDate(new Date().toISOString().split('T')[0]);
    setEndDate('');
    setTagBudgets([]);
    setDraftRule(INITIAL_DRAFT);
    setEditingTag(null);
    setSubmitting(false);
    setError(null);
  }, []);

  const removeTagRule = useCallback((tagToRemove) => {
    const cleaned = cleanTagName(tagToRemove);
    setTagBudgets((prev) => prev.filter((t) => t.tag !== cleaned));
    if (editingTag === cleaned) {
      resetDraft();
    }
  }, [editingTag, resetDraft]);

  const submitWizard = useCallback(async (onSuccess) => {
    if (name.trim().length < 3) return;
    setSubmitting(true);
    setError(null);
    try {
      const project = await projectsRepository.createProject({
        name: name.trim(),
        description: description.trim(),
        mainCurrency,
        startDate,
        endDate: endDate || undefined,
        tagBudgets,
      });
      if (typeof onSuccess === 'function') {
        onSuccess(project);
      }
    } catch (err) {
      setError(err?.message || 'Failed to create project');
      throw err;
    } finally {
      setSubmitting(false);
    }
  }, [name, description, mainCurrency, startDate, endDate, tagBudgets]);

  return {
    step,
    name,
    setName,
    description,
    setDescription,
    mainCurrency,
    setMainCurrency,
    startDate,
    setStartDate,
    endDate,
    setEndDate,
    tagBudgets,
    activeMonths,
    draftRule,
    editingTag,
    submitting,
    error,
    schedulePreview,
    canGoNext,
    isDraftValid,
    nextStep,
    prevStep,
    goToStep,
    updateDraft,
    resetDraft,
    selectPresetTag,
    startEditTag,
    addOrUpdateTagRule,
    removeTagRule,
    submitWizard,
    resetWizard,
  };
}

export default useProjectSetupWizard;
