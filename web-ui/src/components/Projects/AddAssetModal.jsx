import { useEffect } from 'react';
import { Button, DatePicker, Form, Input, InputNumber, Modal, Select, Switch } from 'antd';
import useTypes from '../../hooks/types';
import { useParams } from 'react-router-dom';
import { assetRepository } from '../../api';

const { TextArea } = Input;
const { useForm } = Form;

export default function AddAssetModal(props) {
  const { projectId } = useParams();
  const { open, onSuccess, onCancel } = props;

  const [ , types, , setTypesFilter ] = useTypes();
  const [ form ] = useForm();

  const handleAdd = () => {
    const {
      typeId,
      price,
      currency,
      acquiredAt,
      name,
      description,
      withPayment,
    } = form.getFieldsValue();
    assetRepository.addAsset(projectId, {
      typeId,
      price,
      currency,
      acquiredAt: acquiredAt.toDate(),
      name,
      description,
      withPayment,
    }).then(() => {
      onSuccess();
    }).catch((e) => {
      console.error('Failed to add asset', e.message);
    });
  };

  const handleCancel = () => onCancel();

  useEffect(() => {
    setTypesFilter({
      projectId,
      limit: 100,
      offset: 0,
    });
  }, [ projectId ]);

  return (
    <Modal
      title="Add Asset"
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
        initialValues={ {
          currency: 'UAH',
        } }
      >
        <Form.Item label="Type" name="typeId">
          <Select>
            { types.map((type) => (
              <Select.Option key={ type.key } value={ type.id }>
                { `${ type.name } [${ type.category }]` }
              </Select.Option>
            )) }
          </Select>
        </Form.Item>
        <Form.Item
          name="withPayment"
          label="Create Payment"
          valuePropName="withPayment"
        >
          <Switch/>
        </Form.Item>
        <Form.Item label="Price" name="price">
          <InputNumber/>
        </Form.Item>
        <Form.Item label="Currency" name="currency">
          <Select>
            { [ 'UAH' ].map((code) => (
              <Select.Option key={ code } value={ code }>
                { code }
              </Select.Option>
            )) }
          </Select>
        </Form.Item>
        <Form.Item label="Acquision Date" name="acquiredAt"><DatePicker/></Form.Item>
        <Form.Item label="Name" name="name">
          <Input />
        </Form.Item>
        <Form.Item label="Description" name="description">
          <TextArea rows={ 4 }/>
        </Form.Item>
      </Form>
    </Modal>
  );
}
