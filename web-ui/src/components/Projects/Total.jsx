import { Card, Skeleton, Statistic } from 'antd';
import { useParams } from 'react-router-dom';
import React, { useEffect } from 'react';
import { useProjectTotals } from '../../hooks/projects';

export default function Total() {
  const { projectId } = useParams();
  const [loading, totals, updateTotals] = useProjectTotals(projectId);

  useEffect(() => {
    updateTotals();
  }, []);

  return (
    loading ? <Skeleton active />
      : totals.map((total) => (
        <Card key={total.key} style={{ width: 250, margin: '10px' }}>
          <Statistic title={total.title} value={`${total.amount} ${total.currency}`} />
        </Card>
      ))
  );
}
