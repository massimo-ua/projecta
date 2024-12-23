import { useNavigate } from 'react-router-dom';
import { message } from 'antd';
import { authProvider } from '../api';
import { GoogleLogin } from '@react-oauth/google';


export function GoogleLoginBtn() {
  const [ messageApi, contextHolder ] = message.useMessage();
  const navigate = useNavigate();

  const responseMessage = (response) => {
    return authProvider.loginSocial(
      response.credential,
      'GOOGLE',
    ).then(() => {
      navigate('/');
    });
  };

  const errorMessage = (error) => {
    messageApi.open({
      type: 'error',
      content: `Login failed: ${error.message}`,
    });
  };

  return (
    <>
      { contextHolder }
      <GoogleLogin
        onSuccess={ responseMessage }
        onError={ errorMessage }
        type="icon"
        shape="circle"
        theme="outline"
      />
    </>
  );
}
