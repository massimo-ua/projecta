import { useEffect, useState } from 'react';
import { categoriesRepository } from '../api';

export default function useCategories() {
  const [loading, setLoading] = useState(false);
  const [categories, setCategories] = useState([]);
  const [filter, setFilter] = useState();
  const [total, setTotal] = useState(0);

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
    categoriesRepository.getCategories(projectId, limit, offset)
      .then(([data, total]) => {
        setCategories(data);
        setTotal(total);
      })
      .finally(() => setLoading(false));
  }, [filter]);

  return [loading, categories, total, setFilter];
}
