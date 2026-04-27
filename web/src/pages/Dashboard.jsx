import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { api } from '../lib/api';
import { Plus, LogOut, Map, Users, Clock, Trash2, Edit } from 'lucide-react';
import CreateRouteModal from '../components/CreateRouteModal';

export default function Dashboard() {
  const [routes, setRoutes] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    // Only load routes if user is authenticated
    if (user) {
      loadRoutes();
    } else {
      setLoading(false);
    }
  }, [user]);

  async function loadRoutes() {
    try {
      const data = await api.getRoutes();
      setRoutes(data || []);
    } catch (error) {
      console.error('Failed to load routes:', error);
    } finally {
      setLoading(false);
    }
  }

  async function handleDeleteRoute(id) {
    if (!confirm('Are you sure you want to delete this route?')) return;

    try {
      await api.deleteRoute(id);
      setRoutes(routes.filter(r => r.id !== id));
    } catch (error) {
      alert('Failed to delete route');
    }
  }

  function handleLogout() {
    logout();
    navigate('/login');
  }

  function formatDate(dateString) {
    if (!dateString) return 'Not scheduled';
    return new Date(dateString).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  }

  function formatDuration(seconds) {
    if (!seconds) return '0 min';
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    
    if (hours > 0) {
      return minutes > 0 ? `${hours}h ${minutes}m` : `${hours}h`;
    }
    return `${minutes}m`;
  }

  function formatDistance(meters) {
    if (!meters) return '0 ft';
    const miles = meters * 0.000621371; // meters to miles
    
    if (miles >= 0.1) {
      return `${miles.toFixed(2)} mi`;
    }
    // Show feet for short distances
    const feet = meters * 3.28084;
    return `${Math.round(feet)} ft`;
  }

  function getRouteTypeIcon(type) {
    const icons = {
      driving: '🚗',
      cycling: '🚴',
      walking: '🚶',
      running: '🏃',
    };
    return icons[type] || '📍';
  }

  return (
    <div className="min-h-screen bg-gray-900">
      {/* Header */}
      <header className="bg-gray-800 border-b border-gray-700">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center py-4">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-primary-600 rounded-lg flex items-center justify-center shadow-lg shadow-primary-900/50">
                <Map className="w-6 h-6 text-white" />
              </div>
              <div>
                <h1 className="text-2xl font-bold text-white">Convoy</h1>
                <p className="text-sm text-gray-400">Welcome, {user?.display_name}</p>
              </div>
            </div>
            <button
              onClick={handleLogout}
              className="flex items-center gap-2 px-4 py-2 text-gray-300 hover:text-white hover:bg-gray-700 rounded-lg transition"
            >
              <LogOut className="w-5 h-5" />
              <span>Logout</span>
            </button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="flex justify-between items-center mb-8">
          <div>
            <h2 className="text-3xl font-bold text-white">My Routes</h2>
            <p className="text-gray-400 mt-1">Plan and manage your journeys</p>
          </div>
          <button
            onClick={() => setShowCreateModal(true)}
            className="flex items-center gap-2 px-6 py-3 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition shadow-lg shadow-primary-900/50"
          >
            <Plus className="w-5 h-5" />
            <span>Create Route</span>
          </button>
        </div>

        {loading ? (
          <div className="text-center py-12">
            <div className="inline-block w-8 h-8 border-4 border-primary-600 border-t-transparent rounded-full animate-spin"></div>
            <p className="mt-4 text-gray-400">Loading routes...</p>
          </div>
        ) : routes.length === 0 ? (
          <div className="text-center py-16 bg-gray-800 rounded-2xl border-2 border-dashed border-gray-700">
            <Map className="w-16 h-16 text-gray-600 mx-auto mb-4" />
            <h3 className="text-xl font-semibold text-white mb-2">No routes yet</h3>
            <p className="text-gray-400 mb-6">Create your first route to get started</p>
            <button
              onClick={() => setShowCreateModal(true)}
              className="inline-flex items-center gap-2 px-6 py-3 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition"
            >
              <Plus className="w-5 h-5" />
              <span>Create Route</span>
            </button>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {routes.map((route) => (
              <div
                key={route.id}
                className="bg-gray-800 rounded-xl shadow-sm hover:shadow-lg transition border border-gray-700 overflow-hidden"
              >
                <div className="p-6">
                  <div className="flex items-start justify-between mb-4">
                    <div className="flex items-center gap-3">
                      <span className="text-3xl">{getRouteTypeIcon(route.route_type)}</span>
                      <div>
                        <h3 className="font-semibold text-lg text-white">{route.name}</h3>
                        <span className="text-sm text-gray-400 capitalize">{route.route_type}</span>
                      </div>
                    </div>
                    <span className={`px-3 py-1 rounded-full text-xs font-medium ${
                      route.status === 'planned' ? 'bg-blue-950 text-blue-300 border border-blue-800' :
                      route.status === 'in_progress' ? 'bg-green-950 text-green-300 border border-green-800' :
                      route.status === 'completed' ? 'bg-gray-700 text-gray-300 border border-gray-600' :
                      'bg-red-950 text-red-300 border border-red-800'
                    }`}>
                      {route.status}
                    </span>
                  </div>

                  {route.description && (
                    <p className="text-gray-400 text-sm mb-4 line-clamp-2">{route.description}</p>
                  )}

                  <div className="space-y-2 mb-4">
                    <div className="flex items-center gap-2 text-sm text-gray-400">
                      <Clock className="w-4 h-4" />
                      <span>{formatDate(route.start_time)}</span>
                    </div>
                    {route.distance > 0 && (
                      <div className="flex items-center gap-2 text-sm text-gray-400">
                        <Map className="w-4 h-4" />
                        <span>{formatDistance(route.distance)} • {formatDuration(route.duration)}</span>
                      </div>
                    )}
                    <div className="flex items-center gap-2 text-sm text-gray-400">
                      <Users className="w-4 h-4" />
                      <span>{route.max_participants ? `Max ${route.max_participants} participants` : 'Unlimited'}</span>
                    </div>
                  </div>

                  <div className="flex gap-2 pt-4 border-t border-gray-700">
                    <button
                      onClick={() => navigate(`/routes/${route.id}`)}
                      className="flex-1 px-4 py-2 bg-primary-950 text-primary-300 border border-primary-800 rounded-lg hover:bg-primary-900 transition text-sm font-medium"
                    >
                      View Details
                    </button>
                    {route.creator_id === user?.id && (
                      <button
                        onClick={() => handleDeleteRoute(route.id)}
                        className="px-4 py-2 bg-red-950 text-red-300 border border-red-800 rounded-lg hover:bg-red-900 transition"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </main>

      {showCreateModal && (
        <CreateRouteModal
          onClose={() => setShowCreateModal(false)}
          onSuccess={() => {
            setShowCreateModal(false);
            loadRoutes();
          }}
        />
      )}
    </div>
  );
}
