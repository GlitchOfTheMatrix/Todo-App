import { Navigate, Outlet } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

// Protected route wrapper — redirects to /login if not authenticated
function ProtectedRoute() {
  const { user, loading } = useAuth();

  if (loading) {
    return (
      <div className="loading-screen">
        <span className="spinner spinner-lg" />
      </div>
    );
  }

  return user ? <Outlet /> : <Navigate to="/login" replace />;
}

// Guest route wrapper — redirects to / if already authenticated
function GuestRoute() {
  const { user, loading } = useAuth();

  if (loading) return null;

  return user ? <Navigate to="/" replace /> : <Outlet />;
}

export { GuestRoute, ProtectedRoute };
