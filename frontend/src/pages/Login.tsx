import { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { getLoginUrl, checkAuthStatus } from '../services/api';

export function Login() {
  const [searchParams] = useSearchParams();
  const [error, setError] = useState<string | null>(null);
  const [checking, setChecking] = useState(true);
  const navigate = useNavigate();

  useEffect(() => {
    // Check for error in URL params (from OIDC callback)
    const errorParam = searchParams.get('error');
    if (errorParam) {
      setError(decodeURIComponent(errorParam));
      setChecking(false);
      return;
    }

    // Check if already authenticated
    checkAuthStatus()
      .then((status) => {
        if (status.authenticated) {
          navigate('/');
        }
      })
      .catch(() => {
        // Not authenticated, stay on login page
      })
      .finally(() => {
        setChecking(false);
      });
  }, [navigate, searchParams]);

  const handleLogin = () => {
    window.location.href = getLoginUrl();
  };

  if (checking) {
    return (
      <div className="login-container">
        <div className="login-card">
          <div className="loading">检查登录状态...</div>
        </div>
      </div>
    );
  }

  return (
    <div className="login-container">
      <div className="login-card">
        <h1 className="login-title">微信通知系统</h1>
        <p className="login-subtitle">请登录以继续</p>
        
        {error && (
          <div className="error-message">
            登录失败: {error}
          </div>
        )}
        
        <button onClick={handleLogin} className="login-btn">
          登录
        </button>
      </div>
    </div>
  );
}
