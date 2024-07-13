import { Button, Form, Modal, Input } from 'antd';
import { useParams } from 'react-router-dom';
import { typesRepository } from '../../api';

const { TextArea } = Input;
const { useForm } = Form;

export default function AddTypeModal(props) {
  const { projectId } = useParams();
  const { open, onSuccess, onCancel } = props;
  const [form] = useForm();

  const handleAdd = () => {
    const {
      name,
      description,
    } = form.getFieldsValue();
    typesRepository.addType(projectId, {
      name,
      description,
    }).then(() => {
      onSuccess();
    }).catch((e) => {
      console.error('Failed to add expense', e.message);
    });
  };

  const handleCancel = () => onCancel();

  return (
    <Modal
        title="Add Type"
        open={open}
        onCancel={handleCancel}
        footer={[
          <Button key="add" type="primary" onClick={handleAdd}>
            Submit
          </Button>,
          <Button key="back" onClick={handleCancel}>
            Cancel
          </Button>,
        ]}
      >
        <Form
          form={form}
          labelCol={{
            span: 6,
          }}
          wrapperCol={{
            span: 18,
          }}
          style={{
            maxWidth: 600,
          }}
          autoComplete="off"
        >
          <Form.Item label="Name" name="name">
            <Input />
          </Form.Item>
          <Form.Item label="Description" name="description">
            <TextArea rows={4} />
          </Form.Item>
        </Form>
      </Modal>
  );
}
