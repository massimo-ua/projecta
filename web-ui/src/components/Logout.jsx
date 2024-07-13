import { authProvider } from '../api';
import { useNavigate } from 'react-router-dom';
import { Button } from 'antd';
import { LogoutOutlined } from '@ant-design/icons';

export default function Logout() {
  const navigate = useNavigate();
  const onClick = () => {
    authProvider.logout();
    navigate('/login');
  }

  return (authProvider.isAuthenticated()
    ? <Button icon={<LogoutOutlined />} shape="circle" onClick={onClick} />
    : null);
}
