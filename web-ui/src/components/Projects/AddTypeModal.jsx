import { Button, Form, Modal, Input, Select } from 'antd';
import { useParams } from 'react-router-dom';
import { typesRepository } from '../../api';
import useCategories from '../../hooks/categories.js';
import { useEffect } from 'react';

const { TextArea } = Input;
const { useForm } = Form;

export default function AddTypeModal(props) {
  const { projectId } = useParams();
  const { open, onSuccess, onCancel } = props;
  const [form] = useForm();
  const [, categories, , setCategoriesFilter] = useCategories();

  const handleAdd = () => {
    const {
      categoryId,
      name,
      description,
    } = form.getFieldsValue();
    typesRepository.addType(projectId, {
      categoryId,
      name,
      description,
    }).then(() => {
      onSuccess();
    }).catch((e) => {
      console.error('Failed to add expense', e.message);
    });
  };

  const handleCancel = () => onCancel();

  useEffect(() => {
    setCategoriesFilter({
      projectId,
      limit: 100,
      offset: 0,
    });
  }, []);

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
          <Form.Item label="Category" name="categoryId">
            <Select>
              {categories.map((category) => (
                <Select.Option key={category.key} value={category.id}>
                  {category.name}
                </Select.Option>
              ))}
            </Select>
          </Form.Item>
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
