import React from 'react';
import { Button, Card, Collapse, Grid, List, Row, Skeleton, Space, Typography, theme } from 'antd';
import { CaretRightOutlined } from '@ant-design/icons';

const { useBreakpoint } = Grid;
const { Text } = Typography;

export function ListView({
  loading,
  items,
  total,
  currentPage,
  pageSize,
  onPaginationChange,
  onAddButtonClick,
  addButtonIcon,
  addButtonText,
  addButtonDisabled,
  renderItemMainContent,
  renderItemAmount,
  renderItemDetails,
  renderItemActions,
}) {
  const screens = useBreakpoint();
  const { token } = theme.useToken();

  const ListCard = ({ item }) => {
    const items = [{
      key: '1',
      label: 'Details',
      children: (
        <Space direction="vertical" style={{ width: '100%' }} size="small">
          {renderItemDetails(item)}
          <Space>
            <Row style={{ gap: '5px' }}>
              {renderItemActions(item)}
            </Row>
          </Space>
        </Space>
      ),
    }];

    return (
      <Card
        size="small"
        style={{ marginBottom: '8px' }}
      >
        <Space direction="vertical" style={{ width: '100%' }} size="small">
          <Row justify="space-between" align="top">
            <Space direction="vertical" size={0} style={{ flex: 1 }}>
              {renderItemMainContent(item)}
            </Space>
            {renderItemAmount && (
              <Text
                strong
                style={{
                  fontSize: '16px',
                  marginLeft: '8px'
                }}
              >
                {renderItemAmount(item)}
              </Text>
            )}
          </Row>

          <Collapse
            ghost
            bordered={false}
            expandIcon={({ isActive }) => (
              <CaretRightOutlined rotate={isActive ? 90 : 0} />
            )}
            style={{
              marginLeft: -12,
              marginRight: -12,
              background: 'transparent',
            }}
            items={items}
          />
        </Space>
      </Card>
    );
  };

  if (loading) {
    return (
      <div style={{ padding: '16px' }}>
        <Skeleton active />
        <Skeleton active />
        <Skeleton active />
      </div>
    );
  }

  return (
    <div style={{ padding: '16px' }}>
      <Button
        disabled={addButtonDisabled}
        style={{
          marginBottom: '16px',
          width: screens.xs ? '100%' : 'auto'
        }}
        icon={addButtonIcon}
        type="primary"
        onClick={onAddButtonClick}
      >
        {addButtonText}
      </Button>

      <List
        dataSource={items}
        renderItem={(item) => <ListCard item={item} />}
        pagination={total > pageSize ? {
          current: currentPage,
          pageSize: pageSize,
          total: total,
          onChange: onPaginationChange,
          style: { textAlign: 'center', marginTop: '16px' }
        } : false}
      />
    </div>
  );
}
