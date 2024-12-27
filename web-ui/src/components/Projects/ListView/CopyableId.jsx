import React from 'react';
import { Space, Typography } from 'antd';
import { CopyOutlined } from '@ant-design/icons';

const { Text } = Typography;

export function CopyableId({ id, label = 'ID' }) {
  return (
    <Space>
      <Text type="secondary">{label}:</Text>
      <Text copyable={{
        icon: <CopyOutlined />,
        text: id,
      }}>
        {`${id.slice(0, 6)}...${id.slice(-6)}`}
      </Text>
    </Space>
  );
}
