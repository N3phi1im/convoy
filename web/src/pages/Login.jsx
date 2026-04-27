import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { LogIn, Mail, Lock, AlertCircle } from 'lucide-react';

export default function Login() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const { login } = useAuth();
  const navigate = useNavigate();

  async function handleSubmit(e) {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      await login(email, password);
      navigate('/dashboard');
    } catch (err) {
      setError(err.message || 'Failed to login');
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900 px-4">
      <div className="max-w-md w-full">
        <div className="bg-gray-800 border border-gray-700 rounded-2xl shadow-2xl p-8">
          <div className="text-center mb-8">
            <div className="inline-flex items-center justify-center w-16 h-16 bg-primary-600 rounded-full mb-4 shadow-lg shadow-primary-900/50">
              <LogIn className="w-8 h-8 text-white" />
            </div>
            <h1 className="text-3xl font-bold text-white">Welcome Back</h1>
            <p className="text-gray-400 mt-2">Sign in to your Convoy account</p>
          </div>

          {error && (
            <div className="mb-6 p-4 bg-red-950 border border-red-800 rounded-lg flex items-start gap-3">
              <AlertCircle className="w-5 h-5 text-red-400 flex-shrink-0 mt-0.5" />
              <p className="text-sm text-red-200">{error}</p>
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-6">
            <div>
              <label htmlFor="email" className="block text-sm font-medium text-gray-300 mb-2">
                Email Address
              </label>
              <div className="relative">
                <Mail className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-500" />
                <input
                  id="email"
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                  className="w-full pl-10 pr-4 py-3 bg-gray-900 border border-gray-600 text-white rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none transition placeholder-gray-500"
                  placeholder="you@example.com"
                />
              </div>
            </div>

            <div>
              <label htmlFor="password" className="block text-sm font-medium text-gray-300 mb-2">
                Password
              </label>
              <div className="relative">
                <Lock className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-500" />
                <input
                  id="password"
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                  className="w-full pl-10 pr-4 py-3 bg-gray-900 border border-gray-600 text-white rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none transition placeholder-gray-500"
                  placeholder="••••••••"
                />
              </div>
            </div>

            <button
              type="submit"
              disabled={loading}
              className="w-full bg-primary-600 text-white py-3 rounded-lg font-medium hover:bg-primary-700 focus:ring-4 focus:ring-primary-900/50 transition disabled:opacity-50 disabled:cursor-not-allowed shadow-lg shadow-primary-900/30"
            >
              {loading ? 'Signing in...' : 'Sign In'}
            </button>
          </form>

          <div className="mt-6 text-center">
            <p className="text-gray-400">
              Don't have an account?{' '}
              <Link to="/register" className="text-primary-400 hover:text-primary-300 font-medium">
                Sign up
              </Link>
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
