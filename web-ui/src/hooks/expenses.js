import { useEffect, useState } from 'react';
import { expensesRepository } from '../api';

export default function useExpenses() {
  const [loading, setLoading] = useState(false);
  const [expenses, setExpenses] = useState([]);
  const [total, setTotal] = useState(0);
  const [filter, setFilter] = useState();

  useEffect(() => {
    if (!filter) {
      return;
    }

    const {
      projectId,
      limit,
      offset,
    } = filter;

    setLoading(true);
    expensesRepository.getExpenses(projectId, limit, offset)
      .then(([expenses, total]) => {
        setExpenses(expenses);
        setTotal(total);
      })
      .finally(() => setLoading(false));
  }, [filter]);

  return [loading, expenses, total, setFilter];
}
