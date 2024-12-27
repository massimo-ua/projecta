import React from 'react';
import { Button, Modal } from 'antd';
import { WarningOutlined } from '@ant-design/icons';

export function RemoveButton({ onRemove }) {
  const [modal, contextHolder] = Modal.useModal();

  const handleClick = () => {
    modal.confirm({
      title: 'Confirm removal',
      icon: <WarningOutlined />,
      content: 'Are you sure you want to remove this item?',
      onOk: onRemove,
    });
  };

  return (
    <>
      <Button 
        color="danger" 
        variant="outlined" 
        icon={<WarningOutlined />} 
        onClick={handleClick}
      >
        Remove
      </Button>
      {contextHolder}
    </>
  );
}
