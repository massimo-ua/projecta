import { useNavigate } from 'react-router-dom';
import { Button, message, Tooltip } from 'antd';
import { authProvider } from '../api';
import { useGoogleLogin } from '@react-oauth/google';
import { useState } from 'react';
import { GoogleOutlined, LoadingOutlined } from '@ant-design/icons';

export function GoogleLoginBtn() {
  const [ loading, setLoading ] = useState(false);
  const [ messageApi, contextHolder ] = message.useMessage();
  const navigate = useNavigate();

  const onSuccess = (response) => {
    return authProvider.loginSocial(
      response.code,
      'GOOGLE',
    ).then(() => {
      setLoading(false);
      navigate('/');
    });
  };

  const onError = (error) => {
    setLoading(false);
    messageApi.open({
      type: 'error',
      content: `Login failed: ${ error.message }`,
    });
  };

  const onClick = () => {
    setLoading(true);
    login();
  };

  const login = useGoogleLogin({
    flow: 'auth-code',
    onSuccess,
    onError,
  });


  return (
    <>
      { contextHolder }
      { loading ? <LoadingOutlined/> : <Tooltip title="Login with Google">
        <Button shape="circle" icon={ <GoogleOutlined/> } onClick={ onClick }/>
      </Tooltip> }
    </>
  );
}
