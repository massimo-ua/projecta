import React from 'react';
import { Typography } from 'antd';
import './DetailItem.css';

const { Text } = Typography;

export function DetailItem({ label, children }) {
  return (
    <div className="detail-item">
      <Text type="secondary" className="detail-item-label">{label}</Text>
      {children}
    </div>
  );
}
