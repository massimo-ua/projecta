import { DeleteOutlined } from '@ant-design/icons';
import { Button } from 'antd';

export default function RemoveTypeButton(props) {
  const { typeId, onClick } = props;

  return (
    <Button
      icon={<DeleteOutlined />}
      onClick={() => onClick(typeId)}
    />
  );
}
