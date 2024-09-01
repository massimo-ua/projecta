import { useEffect } from 'react';
import { Button, DatePicker, Form, Input, InputNumber, Modal, Select } from 'antd';
import useTypes from '../../hooks/types';
import { useParams } from 'react-router-dom';
import { assetRepository } from '../../api';

const { TextArea } = Input;
const { useForm } = Form;

export default function AddAssetModal(props) {
  const { projectId } = useParams();
  const { onSuccess, onCancel, assetId } = props;

  const [ , types, , setTypesFilter ] = useTypes();
  const [ form ] = useForm();

  const handleUpdate = () => {
    const {
      typeId,
      price,
      currency,
      acquiredAt,
      name,
      description,
    } = form.getFieldsValue();
    assetRepository.updateAsset(projectId, {
      id: assetId,
      typeId,
      price,
      currency,
      acquiredAt: acquiredAt.toDate(),
      name,
      description,
    }).then(() => {
      onSuccess();
    }).catch((e) => {
      console.error('Failed to update asset', e.message);
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

  useEffect(() => {
    if (!assetId) {
      return;
    }

    assetRepository
      .getAsset(projectId, assetId)
      .then((asset) => {
      form.setFieldsValue({
        typeId: asset.typeId,
        price: asset.price,
        currency: asset.currency,
        acquiredAt: asset.acquiredAt,
        name: asset.name,
        description: asset.description,
      });
  })}, [ assetId ]);

  return (
    <Modal
      title="Edit Asset"
      open={ Boolean(assetId) }
      onCancel={ handleCancel }
      footer={ [
        <Button key="edit" type="primary" onClick={ handleUpdate }>
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
        <Form.Item label="Type" name="typeId">
          <Select>
            { types.map((type) => (
              <Select.Option key={ type.key } value={ type.id }>
                { `${ type.name } [${ type.category }]` }
              </Select.Option>
            )) }
          </Select>
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
