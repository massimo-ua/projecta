import {
  Button, Col, Divider, Form, Input, Row, message,
} from 'antd';
import { useNavigate } from 'react-router-dom';
import HomeLayout from '../Layout';
import { authProvider } from '../api';

export function Login() {
  const [messageApi, contextHolder] = message.useMessage();
  const navigate = useNavigate();

  const onFinish = async (values) => {
    authProvider.login(values.username, values.password).then(() => {
      navigate('/');
    });
  };

  const onFinishFailed = (error) => {
    messageApi.open({
      type: 'error',
      content: 'Login failed!',
    });
  };

  return (
    <HomeLayout>
      {contextHolder}
      <Divider orientation="center">Login</Divider>
      <Row>
        <Col flex={3} />
        <Col flex={3} style={{ padding: '10px' }}>
          <Form
            name="basic"
            labelCol={{
              span: 8,
            }}
            wrapperCol={{
              span: 16,
            }}
            style={{
              maxWidth: 600,
            }}
            initialValues={{
              remember: true,
            }}
            onFinish={onFinish}
            onFinishFailed={onFinishFailed}
            autoComplete="off"
          >
            <Form.Item
              label="Username"
              name="username"
              rules={[
                {
                  required: true,
                  message: 'Please input your username!',
                },
              ]}
            >
              <Input />
            </Form.Item>

            <Form.Item
              label="Password"
              name="password"
              rules={[
                {
                  required: true,
                  message: 'Please input your password!',
                },
              ]}
            >
              <Input.Password />
            </Form.Item>

            <Form.Item
              wrapperCol={{
                offset: 8,
                span: 16,
              }}
            >
              <Button type="primary" htmlType="submit">
                Submit
              </Button>
            </Form.Item>
          </Form>
        </Col>
        <Col flex={3} />
      </Row>
    </HomeLayout>
  );
}