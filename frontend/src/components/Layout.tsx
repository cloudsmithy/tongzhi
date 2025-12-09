import { Link, Outlet, useNavigate } from 'react-router-dom';
import { logout } from '../services/api';

export function Layout() {
  const navigate = useNavigate();

  const handleLogout = async () => {
    try {
      await logout();
      navigate('/login');
    } catch (error) {
      console.error('Logout failed:', error);
      navigate('/login');
    }
  };

  return (
    <div className="app-container">
      <nav className="navbar">
        <Link to="/" className="nav-brand">微信通知系统</Link>
        <div className="nav-links">
          <Link to="/" className="nav-link">主页</Link>
          <Link to="/settings" className="nav-link">设置</Link>
          <Link to="/webhook" className="nav-link">Webhook</Link>
          <button onClick={handleLogout} className="nav-link logout-btn">
            登出
          </button>
        </div>
      </nav>
      <main className="main-content">
        <Outlet />
      </main>
    </div>
  );
}
