import { useParams } from 'react-router-dom';
import React, { useEffect, useState } from 'react';
import { Button, Card, Collapse, Grid, List, Modal, Row, Skeleton, Space, Tag, Typography } from 'antd';
import { CopyOutlined, DollarOutlined, EditOutlined, WarningOutlined, } from '@ant-design/icons';
import usePayments from '../../hooks/payments';
import AddPaymentModal from './AddPaymentModal';
import { paymentRepository } from '../../api';
import { DEFAULT_OFFSET, PAGE_SIZE } from '../../constants';
import EditPaymentModal from './EditPaymentModal';

const { useBreakpoint } = Grid;
const { Text } = Typography;
const { Panel } = Collapse;

export function PaymentsListView() {
  const { projectId } = useParams();
  const screens = useBreakpoint();
  const [modal, contextHolder] = Modal.useModal();
  const [ loading, payments, total, setFilter ] = usePayments();
  const [ addModalOpened, setAddModalOpen ] = useState(false);
  const [ paymentIdToEdit, setPaymentIdToEdit ] = useState('');
  const [ currentPage, setCurrentPage ] = useState(1);

  const onPaginationChange = (nextPage) => {
    setCurrentPage(nextPage);
  };

  const onAddButtonClick = () => {
    if (!addModalOpened) {
      setAddModalOpen(true);
    }
  };

  const onEditButtonClick = (paymentId) => {
    if (!paymentIdToEdit) {
      setPaymentIdToEdit(paymentId);
    }
  };

  useEffect(() => {
    setFilter({
      projectId,
      limit: PAGE_SIZE,
      offset: (currentPage - 1) * PAGE_SIZE,
    });
  }, [ currentPage, projectId, setFilter ]);

  const onAddCancel = () => setAddModalOpen(false);
  const onAddSuccess = () => {
    setAddModalOpen(false);
    setFilter({
      projectId,
      limit: PAGE_SIZE,
      offset: DEFAULT_OFFSET,
    });
  };

  const onRemoveButtonClick = (paymentId) => {
    paymentRepository.removePayment(projectId, paymentId)
      .then(() => {
        setFilter({
          projectId,
          limit: PAGE_SIZE,
          offset: DEFAULT_OFFSET,
        });
      })
      .catch((error) => {
        console.error(error);
      });
  };

  const onEditSuccess = () => {
    setPaymentIdToEdit('');
    setFilter({
      projectId,
      limit: PAGE_SIZE,
      offset: DEFAULT_OFFSET,
    });
  };

  const onEditCancel = () => setPaymentIdToEdit('');

  const PaymentCard = ({ payment }) => (
    <Card
      size="small"
      style={ { marginBottom: '8px' } }
    >
      <Space direction="vertical" style={ { width: '100%' } } size="small">
        <Row justify="space-between" align="top">
          <Space direction="vertical" size={ 0 } style={ { flex: 1 } }>
            <Text strong>{ payment.description }</Text>
            <Space size={ 4 }>
              <Text type="secondary" style={ { fontSize: '12px' } }>
                { payment.paymentDate }
              </Text>
              <Tag>{ payment.category }</Tag>
            </Space>
          </Space>
          <Text
            strong
            style={ {
              color: payment.kind === 'DOWN_PAYMENT' ? 'red' : 'green',
              fontSize: '16px',
              marginLeft: '8px'
            } }
          >
            { payment.amount } { payment.currency }
          </Text>
        </Row>

        <Collapse ghost style={ { marginLeft: -12, marginRight: -12 } }>
          <Panel
            header="Details"
            key="1"
            style={ { padding: '0 12px' } }
          >
            <Space direction="vertical" style={ { width: '100%' } } size="small">
              <Space>
                <Text type="secondary">ID:</Text>
                <Text copyable={{
                  icon: <CopyOutlined/>,
                  text: payment.id,
                }}>
                  { `${ payment.id.slice(0, 6) }...${ payment.id.slice(-6) }` }
                </Text>
              </Space>
              <Space>
                <Text type="secondary">Type:</Text>
                <Text>
                  { payment.type }
                </Text>
              </Space>
              <Space>
                <Row style={ { gap: '5px' } }>
                  <Button icon={ <EditOutlined/> } onClick={ () => onEditButtonClick(payment.id) }>Edit</Button>
                  <>
                    <Button color="danger" variant="outlined" icon={ <WarningOutlined/> } onClick={ () => modal.confirm({
                      title: 'Confirm payment removal',
                      icon: <WarningOutlined/>,
                      content: 'Confirm payment removal',
                      onOk: () => onRemoveButtonClick(payment.id),
                    }) }>Remove</Button>
                    <div>{ contextHolder }</div>
                  </>
                </Row>
              </Space>
            </Space>
          </Panel>
        </Collapse>
      </Space>
    </Card>
  );

  if (loading) {
    return (
      <div style={ { padding: '16px' } }>
        <Skeleton active/>
        <Skeleton active/>
        <Skeleton active/>
      </div>
    );
  }

  return (
    <div style={ { padding: '16px' } }>
      <Button
        disabled={ addModalOpened }
        style={ {
          marginBottom: '16px',
          width: screens.xs ? '100%' : 'auto'
        } }
        icon={ <DollarOutlined/> }
        type="primary"
        onClick={ onAddButtonClick }
      >
        Add Payment
      </Button>

      <List
        dataSource={ payments }
        renderItem={ (payment) => <PaymentCard payment={ payment }/> }
        pagination={ total > PAGE_SIZE ? {
          current: currentPage,
          pageSize: PAGE_SIZE,
          total: total,
          onChange: onPaginationChange,
          style: { textAlign: 'center', marginTop: '16px' }
        } : false }
      />

      <AddPaymentModal
        open={ addModalOpened }
        onCancel={ onAddCancel }
        onSuccess={ onAddSuccess }
      />
      <EditPaymentModal
        paymentId={ paymentIdToEdit }
        onCancel={ onEditCancel }
        onSuccess={ onEditSuccess }
      />
    </div>
  );
}
