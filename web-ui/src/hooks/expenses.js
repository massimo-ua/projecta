import { useEffect, useState } from 'react';
import { paymentRepository } from '../api';

export default function useExpenses() {
  const [loading, setLoading] = useState(false);
  const [payments, setPayments] = useState([]);
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
    paymentRepository.getPayments(projectId, limit, offset)
      .then(([payments, total]) => {
        setPayments(payments);
        setTotal(total);
      })
      .finally(() => setLoading(false));
  }, [filter]);

  return [loading, payments, total, setFilter];
}
