import { Button, Form, Input, Modal } from 'antd';
import { useParams } from 'react-router-dom';
import { categoriesRepository } from '../../api';

const { TextArea } = Input;
const { useForm } = Form;

export default function AddCategoryModal(props) {
  const { projectId } = useParams();
  const { open, onSuccess, onCancel } = props;
  const [ form ] = useForm();

  const handleAdd = () => {
    const {
      name,
      description,
    } = form.getFieldsValue();
    categoriesRepository.addCategory(projectId, {
      name,
      description,
    }).then(() => {
      onSuccess();
    }).catch((e) => {
      console.error('Failed to add category', e.message);
    });
  };

  const handleCancel = () => onCancel();

  return (
    <Modal
      title="Add Category"
      open={ open }
      onCancel={ handleCancel }
      footer={ [
        <Button key="add" type="primary" onClick={ handleAdd }>
          Submit
        </Button>,
        <Button key="back" onClick={ handleCancel }>
          Cancel
        </Button>,
      ] }
    >
      <Form
        form={ form }
        labelCol={ {
          span: 6,
        } }
        wrapperCol={ {
          span: 18,
        } }
        style={ {
          maxWidth: 600,
        } }
        autoComplete="off"
      >
        <Form.Item label="Name" name="name">
          <Input/>
        </Form.Item>
        <Form.Item label="Description" name="description">
          <TextArea rows={ 4 }/>
        </Form.Item>
      </Form>
    </Modal>
  );
}
