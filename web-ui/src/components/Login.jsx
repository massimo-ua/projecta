import { Button, Col, Divider, Form, Input, message, Row, } from 'antd';
import { useNavigate } from 'react-router-dom';
import HomeLayout from '../Layout';
import { authProvider } from '../api';
import { GoogleLoginBtn } from './GoogleLoginBtn';

export function Login() {
  const [ messageApi, contextHolder ] = message.useMessage();
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
      { contextHolder }
      <Row>
        <Col
          span={ 12 }
          style={ {
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            backgroundImage: "url('/traxer-vjMZIuTozXc-unsplash.jpg')",
            backgroundSize: 'cover',
            backgroundPosition: 'center',
            backgroundRepeat: 'no-repeat',
            height: '92vh',
            width: '100%',
          } }>
        </Col>
        <Col
          span={ 12 }
          style={ {
            display: 'flex',
            alignItems: 'flex-start',
            justifyContent: 'center',
          } }
        >
          <Row>
            <Col>
              <Row style={{ paddingTop: 30 }}><Divider orientation="center">Social</Divider></Row>
              <Row>
                <Col flex={ 3 }
                     style={ { padding: '10px', display: 'flex', justifyContent: 'center', alignItems: 'center' } }>
                  <GoogleLoginBtn/>
                </Col>
              </Row>
              <Row><Divider orientation="center">Login</Divider></Row>
              <Row>
                <Col flex={ 5 }/>
                <Col flex={ 1 }
                     style={ { padding: '10px', display: 'flex', justifyContent: 'center', alignItems: 'center' } }>
                  <Form
                    name="basic"
                    labelCol={ {
                      span: 8,
                    } }
                    wrapperCol={ {
                      span: 16,
                    } }
                    style={ {
                      maxWidth: 500,
                      minWidth: 300,
                    } }
                    initialValues={ {
                      remember: true,
                    } }
                    onFinish={ onFinish }
                    onFinishFailed={ onFinishFailed }
                    autoComplete="off"
                  >
                    <Form.Item
                      label="Username"
                      name="username"
                      rules={ [
                        {
                          required: true,
                          message: 'Please input your username!',
                        },
                      ] }
                    >
                      <Input/>
                    </Form.Item>

                    <Form.Item
                      label="Password"
                      name="password"
                      rules={ [
                        {
                          required: true,
                          message: 'Please input your password!',
                        },
                      ] }
                    >
                      <Input.Password/>
                    </Form.Item>

                    <Form.Item
                      wrapperCol={ {
                        offset: 8,
                        span: 16,
                      } }
                    >
                      <Button type="primary" htmlType="submit">
                        Submit
                      </Button>
                    </Form.Item>
                  </Form>
                </Col>
                <Col flex={ 6 }/>
              </Row>
            </Col>
          </Row>
        </Col>
      </Row>
    </HomeLayout>
  );
}
