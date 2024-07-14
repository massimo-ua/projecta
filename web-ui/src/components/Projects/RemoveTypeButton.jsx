import { DeleteOutlined, WarningOutlined } from '@ant-design/icons';
import { Button, Modal } from 'antd';

export default function RemoveTypeButton(props) {
  const [modal, contextHolder] = Modal.useModal();
  const { typeId, onClick } = props;

  return (<>
    <Button
      icon={<DeleteOutlined />}
      onClick={() => modal.confirm({
        title: 'Confirm type removal',
        icon: <WarningOutlined />,
        content: 'Confirm type removal',
        onOk: () => onClick(typeId),
      })} />
      <div>{contextHolder}</div>
    </>
  );
}
