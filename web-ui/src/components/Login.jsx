import { Button, Form, Input, message } from 'antd';
import { useNavigate } from 'react-router-dom';
import HomeLayout from '../Layout';
import { authProvider } from '../api';
import { GoogleLoginBtn } from './GoogleLoginBtn';
import './Login.css';

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
      content: `Login failed: ${error.message}`,
    });
  };

  return (
    <HomeLayout>
      {contextHolder}
      <div className="login-container">
        <div className="login-image-section" />

        <div className="login-form-section">
          <div className="login-form-container">
            <div className="login-divider-section">
              <div className="login-divider-text">Social</div>
              <div className="login-social-section">
                <GoogleLoginBtn />
              </div>

              <div className="login-divider-text">Login</div>
            </div>

            <Form
              name="basic"
              labelCol={{
                span: 8,
              }}
              wrapperCol={{
                span: 16,
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
          </div>
        </div>
      </div>
    </HomeLayout>
  );
}
