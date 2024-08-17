import { useEffect, useState } from 'react';
import { assetRepository } from '../api';

export default function useAssets() {
  const [loading, setLoading] = useState(false);
  const [assets, setPayments] = useState([]);
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
    assetRepository.getAssets(projectId, limit, offset)
      .then(([assets, total]) => {
        setPayments(assets);
        setTotal(total);
      })
      .finally(() => setLoading(false));
  }, [filter]);

  return [loading, assets, total, setFilter];
}
