import { DeleteOutlined, WarningOutlined } from '@ant-design/icons';
import { Modal } from 'antd';

export default function RemoveTypeButton(props) {
  const [modal, contextHolder] = Modal.useModal();
  const { typeId, onClick } = props;

  return (<>
    <DeleteOutlined
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
