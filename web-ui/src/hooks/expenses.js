import { useEffect, useState } from 'react';
import { expensesRepository } from '../api';

export default function useExpenses() {
  const [loading, setLoading] = useState(false);
  const [expenses, setExpenses] = useState([]);
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
      .then((data) => {
        setExpenses(data);
      })
      .finally(() => setLoading(false));
  }, [filter]);

  return [loading, expenses, setFilter];
}
