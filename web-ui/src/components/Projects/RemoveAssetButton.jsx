import { DeleteOutlined, WarningOutlined } from '@ant-design/icons';
import { Button, Modal } from 'antd';

export default function RemoveAssetButton(props) {
  const [modal, contextHolder] = Modal.useModal();
  const { assetId, onClick } = props;

  return (<>
    <Button
      icon={<DeleteOutlined />}
      onClick={() => modal.confirm({
        title: 'Confirm asset removal',
        icon: <WarningOutlined />,
        content: 'Selected asset will be permanently removed. Are you sure?',
        onOk: () => onClick(assetId),
      })} />
      <div>{contextHolder}</div>
    </>
  );
}
