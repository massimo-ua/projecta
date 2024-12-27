import { Card, Skeleton, Statistic } from 'antd';
import { useParams } from 'react-router-dom';
import React, { useEffect } from 'react';
import { useProjectTotals } from '../../hooks/projects';
import './Total.css';

function TotalCard({ total }) {
  return (
    <div className="TotalCard_container">
      <Card bordered style={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
        <Statistic
          title={total.title}
          value={total.amount}
          suffix={total.currency}
          style={{ height: '100%', display: 'flex', flexDirection: 'column', justifyContent: 'center' }}
        />
      </Card>
    </div>
  );
}

export default function Total() {
  const { projectId } = useParams();
  const [loading, totals, updateTotals] = useProjectTotals(projectId);

  useEffect(() => {
    updateTotals();
  }, []);

  if (loading) {
    return (
      <div className="Total_container">
        <Skeleton active />
        <Skeleton active />
        <Skeleton active />
      </div>
    );
  }

  return (
    <div className="Total_container">
      {totals.map((total) => (
        <TotalCard key={total.key} total={total} />
      ))}
    </div>
  );
}
