import { DeleteOutlined, WarningOutlined } from '@ant-design/icons';
import { Button, Modal } from 'antd';

export default function RemovePaymentButton(props) {
  const [modal, contextHolder] = Modal.useModal();
  const { paymentId, onClick } = props;

  return (<>
    <Button
      icon={<DeleteOutlined />}
      onClick={() => modal.confirm({
        title: 'Confirm payment removal',
        icon: <WarningOutlined />,
        content: 'Confirm payment removal',
        onOk: () => onClick(paymentId),
      })} />
      <div>{contextHolder}</div>
    </>
  );
}
