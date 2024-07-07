import { useEffect, useState } from 'react';
import { typesRepository } from '../api';

export default function useTypes() {
  const [loading, setLoading] = useState(false);
  const [types, setTypes] = useState([]);
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
    typesRepository.getTypes(projectId, limit, offset)
      .then((data) => {
        setTypes(data);
      })
      .finally(() => setLoading(false));
  }, [filter]);

  return [loading, types, setFilter];
}
