import { DeleteOutlined, WarningOutlined } from '@ant-design/icons';
import { Button, Modal } from 'antd';

export default function RemoveExpenseButton(props) {
  const [modal, contextHolder] = Modal.useModal();
  const { expenseId, onClick } = props;

  return (<>
    <Button
      icon={<DeleteOutlined />}
      onClick={() => modal.confirm({
        title: 'Confirm expense removal',
        icon: <WarningOutlined />,
        content: 'Confirm expense removal',
        onOk: () => onClick(expenseId),
      })} />
      <div>{contextHolder}</div>
    </>
  );
}
