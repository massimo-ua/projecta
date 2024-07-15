import { useEffect } from 'react';
import { Button, Form, Modal, Select, Input, InputNumber, DatePicker } from 'antd';
import useTypes from '../../hooks/types';
import { useParams } from 'react-router-dom';
import { expensesRepository } from '../../api';

const { TextArea } = Input;
const { useForm } = Form;

export default function AddExpenseModal(props) {
  const { projectId } = useParams();
  const { open, onSuccess, onCancel } = props;

  const [, types, , setTypesFilter] = useTypes();
  const [form] = useForm();

  const handleAdd = () => {
    const {
      typeId,
      amount,
      currency,
      expenseDate,
      description,
    } = form.getFieldsValue();
    expensesRepository.addExpense(projectId, {
      typeId,
      amount,
      currency,
      expenseDate: expenseDate.toDate(),
      description,
    }).then(() => {
      onSuccess();
    }).catch((e) => {
      console.error('Failed to add expense', e.message);
    });
  };

  const handleCancel = () => onCancel();

  useEffect(() => {
    setTypesFilter({
      projectId,
      limit: 100,
      offset: 0,
    });
  }, [projectId]);

  return (
    <Modal
        title="Add Expense"
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
          initialValues={{
            currency: 'UAH',
          }}
        >
          <Form.Item label="Type" name="typeId">
            <Select>
              {types.map((type) => (
                <Select.Option key={type.key} value={type.id}>
                  {`${type.name} [${type.category}]`}
                </Select.Option>
              ))}
            </Select>
          </Form.Item>
          <Form.Item label="Amount" name="amount">
            <InputNumber />
          </Form.Item>
          <Form.Item label="Currency" name="currency">
            <Select>
            {['UAH'].map((code) => (
              <Select.Option key={code} value={code}>
                {code}
              </Select.Option>
            ))}
            </Select>
          </Form.Item>
          <Form.Item label="Expense Date" name="expenseDate"><DatePicker /></Form.Item>

          <Form.Item label="Description" name="description">
            <TextArea rows={4} />
          </Form.Item>
        </Form>
      </Modal>
  );
}
