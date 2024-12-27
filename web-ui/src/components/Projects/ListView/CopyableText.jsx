import React from 'react';
import { Space, Typography } from 'antd';
import { CopyOutlined } from '@ant-design/icons';

const { Text } = Typography;

export function CopyableText({ text, label, truncate }) {
  const displayText = truncate ? `${text.slice(0, 6)}...${text.slice(-6)}` : text;

  return (
    <Space>
      {label && <Text type="secondary">{label}</Text>}
      <Text copyable={{
        icon: <CopyOutlined />,
        text: text,
      }}>
        {displayText}
      </Text>
    </Space>
  );
}
