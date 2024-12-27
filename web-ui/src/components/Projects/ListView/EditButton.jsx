import React from 'react';
import { Button } from 'antd';
import { EditOutlined } from '@ant-design/icons';

export function EditButton({ onClick }) {
  return (
    <Button 
      icon={<EditOutlined />} 
      onClick={onClick}
    >
      Edit
    </Button>
  );
}
