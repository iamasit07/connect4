import { Link } from 'react-router-dom';
import { LoginForm } from '../components/LoginForm';

const LoginPage = () => {
  return (
    <div className="min-h-screen bg-background flex flex-col items-center justify-center p-4">
      <Link to="/" className="mb-8">
        <h1 className="text-3xl font-bold text-primary">Connect 4</h1>
      </Link>
      <LoginForm />
    </div>
  );
};

export default LoginPage;
