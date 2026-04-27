import { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { api } from '../lib/api';
import { useAuth } from '../contexts/AuthContext';
import { ArrowLeft, Users, Clock, MapPin, Calendar, UserPlus, Trash2 } from 'lucide-react';
import mapboxgl from 'mapbox-gl';
import polyline from '@mapbox/polyline';
import 'mapbox-gl/dist/mapbox-gl.css';

mapboxgl.accessToken = import.meta.env.VITE_MAPBOX_TOKEN;

export default function RouteDetail() {
  const { id } = useParams();
  const navigate = useNavigate();
  const { user } = useAuth();
  const [route, setRoute] = useState(null);
  const [participants, setParticipants] = useState([]);
  const [loading, setLoading] = useState(true);
  const [joining, setJoining] = useState(false);
  const mapContainer = useRef(null);
  const map = useRef(null);

  useEffect(() => {
    loadRouteData();
  }, [id]);

  // Initialize map when route data is loaded
  useEffect(() => {
    if (!route || !route.waypoints || map.current) return;

    // Initialize map
    map.current = new mapboxgl.Map({
      container: mapContainer.current,
      style: 'mapbox://styles/mapbox/outdoors-v12',
      center: [route.waypoints[0].longitude, route.waypoints[0].latitude],
      zoom: 12,
    });

    map.current.addControl(new mapboxgl.NavigationControl(), 'top-right');

    map.current.on('load', () => {
      // Add waypoint markers
      route.waypoints.forEach((wp, index) => {
        const el = document.createElement('div');
        el.className = 'waypoint-marker';
        el.style.backgroundColor = index === 0 ? '#22c55e' : index === route.waypoints.length - 1 ? '#ef4444' : '#3b82f6';
        el.style.width = '30px';
        el.style.height = '30px';
        el.style.borderRadius = '50%';
        el.style.border = '3px solid white';
        el.style.boxShadow = '0 2px 4px rgba(0,0,0,0.3)';
        el.style.display = 'flex';
        el.style.alignItems = 'center';
        el.style.justifyContent = 'center';
        el.style.color = 'white';
        el.style.fontWeight = 'bold';
        el.style.fontSize = '14px';
        el.textContent = index + 1;

        new mapboxgl.Marker({ element: el })
          .setLngLat([wp.longitude, wp.latitude])
          .setPopup(new mapboxgl.Popup().setHTML(`<strong>${wp.name || `Point ${index + 1}`}</strong>`))
          .addTo(map.current);
      });

      // Add route line
      let coordinates;
      
      // Use Mapbox geometry if available (actual road route)
      if (route.geometry) {
        // Decode polyline to coordinates [lat, lng] then convert to [lng, lat]
        const decoded = polyline.decode(route.geometry);
        coordinates = decoded.map(coord => [coord[1], coord[0]]); // [lat, lng] -> [lng, lat]
      } else {
        // Fallback to straight lines between waypoints
        coordinates = route.waypoints.map(wp => [wp.longitude, wp.latitude]);
      }
      
      map.current.addSource('route', {
        type: 'geojson',
        data: {
          type: 'Feature',
          properties: {},
          geometry: {
            type: 'LineString',
            coordinates: coordinates,
          },
        },
      });

      map.current.addLayer({
        id: 'route',
        type: 'line',
        source: 'route',
        layout: {
          'line-join': 'round',
          'line-cap': 'round',
        },
        paint: {
          'line-color': '#3b82f6',
          'line-width': 4,
        },
      });

      // Fit map to show all waypoints
      const bounds = new mapboxgl.LngLatBounds();
      coordinates.forEach(coord => bounds.extend(coord));
      map.current.fitBounds(bounds, { padding: 50 });
    });

    return () => {
      if (map.current) {
        map.current.remove();
        map.current = null;
      }
    };
  }, [route]);

  async function loadRouteData() {
    try {
      const [routeData, participantsData] = await Promise.all([
        api.getRoute(id),
        api.getRouteParticipants(id),
      ]);
      setRoute(routeData);
      setParticipants(participantsData || []);
    } catch (error) {
      console.error('Failed to load route:', error);
      alert('Failed to load route details');
      navigate('/dashboard');
    } finally {
      setLoading(false);
    }
  }

  async function handleJoinRoute() {
    setJoining(true);
    try {
      await api.joinRoute(id);
      await loadRouteData();
    } catch (error) {
      alert(error.message || 'Failed to join route');
    } finally {
      setJoining(false);
    }
  }

  async function handleLeaveRoute() {
    if (!confirm('Are you sure you want to leave this route?')) return;
    
    setJoining(true);
    try {
      await api.leaveRoute(id);
      await loadRouteData();
    } catch (error) {
      alert(error.message || 'Failed to leave route');
    } finally {
      setJoining(false);
    }
  }

  async function handleRemoveParticipant(userId, displayName) {
    if (!confirm(`Remove ${displayName} from this route?`)) return;
    
    try {
      await api.removeParticipant(id, userId);
      await loadRouteData();
    } catch (error) {
      alert(error.message || 'Failed to remove participant');
    }
  }

  function formatDate(dateString) {
    if (!dateString) return 'Not scheduled';
    return new Date(dateString).toLocaleDateString('en-US', {
      weekday: 'long',
      month: 'long',
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

  const isParticipant = participants.some(p => p.user_id === user?.id);
  const isCreator = route?.creator_id === user?.id;

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="inline-block w-8 h-8 border-4 border-primary-600 border-t-transparent rounded-full animate-spin"></div>
          <p className="mt-4 text-gray-600">Loading route...</p>
        </div>
      </div>
    );
  }

  if (!route) return null;

  return (
    <div className="min-h-screen bg-gray-900">
      {/* Header */}
      <div className="bg-gray-800 border-b border-gray-700">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <button
            onClick={() => navigate('/dashboard')}
            className="flex items-center gap-2 text-gray-400 hover:text-white mb-4"
          >
            <ArrowLeft className="w-5 h-5" />
            <span>Back to Dashboard</span>
          </button>
        </div>
      </div>

      {/* Main Content */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Route Details */}
          <div className="lg:col-span-2 space-y-6">
            <div className="bg-gray-800 rounded-2xl shadow-sm border border-gray-700 p-8">
              <div className="flex items-start justify-between mb-6">
                <div className="flex items-center gap-4">
                  <span className="text-5xl">{getRouteTypeIcon(route.route_type)}</span>
                  <div>
                    <h1 className="text-3xl font-bold text-white">{route.name}</h1>
                    <p className="text-gray-400 capitalize mt-1">{route.route_type}</p>
                  </div>
                </div>
                <span className={`px-4 py-2 rounded-full text-sm font-medium border ${
                  route.status === 'planned' ? 'bg-blue-950 text-blue-300 border-blue-800' :
                  route.status === 'in_progress' ? 'bg-green-950 text-green-300 border-green-800' :
                  route.status === 'completed' ? 'bg-gray-700 text-gray-300 border-gray-600' :
                  'bg-red-950 text-red-300 border-red-800'
                }`}>
                  {route.status}
                </span>
              </div>

              {route.description && (
                <div className="mb-6">
                  <h2 className="text-lg font-semibold text-white mb-2">Description</h2>
                  <p className="text-gray-400">{route.description}</p>
                </div>
              )}

              <div className="grid grid-cols-2 gap-4 mb-6">
                <div className="flex items-center gap-3 p-4 bg-gray-900 rounded-lg">
                  <Calendar className="w-5 h-5 text-primary-600" />
                  <div>
                    <p className="text-sm text-gray-400">Start Time</p>
                    <p className="font-medium text-white">{formatDate(route.start_time)}</p>
                  </div>
                </div>

                <div className="flex items-center gap-3 p-4 bg-gray-900 rounded-lg">
                  <Users className="w-5 h-5 text-primary-600" />
                  <div>
                    <p className="text-sm text-gray-400">Participants</p>
                    <p className="font-medium text-white">
                      {participants.length}
                      {route.max_participants ? ` / ${route.max_participants}` : ''}
                    </p>
                  </div>
                </div>

                {route.difficulty && (
                  <div className="flex items-center gap-3 p-4 bg-gray-900 rounded-lg">
                    <MapPin className="w-5 h-5 text-primary-600" />
                    <div>
                      <p className="text-sm text-gray-400">Difficulty</p>
                      <p className="font-medium text-white capitalize">{route.difficulty}</p>
                    </div>
                  </div>
                )}

                <div className="flex items-center gap-3 p-4 bg-gray-900 rounded-lg">
                  <Clock className="w-5 h-5 text-primary-600" />
                  <div>
                    <p className="text-sm text-gray-400">Visibility</p>
                    <p className="font-medium text-white capitalize">{route.visibility.replace('_', ' ')}</p>
                  </div>
                </div>
              </div>

              {!isCreator && !isParticipant && route.status === 'planned' && (
                <button
                  onClick={handleJoinRoute}
                  disabled={joining || (route.max_participants && participants.length >= route.max_participants)}
                  className="w-full flex items-center justify-center gap-2 px-6 py-3 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition font-medium disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  <UserPlus className="w-5 h-5" />
                  <span>{joining ? 'Joining...' : 'Join Route'}</span>
                </button>
              )}

              {isParticipant && !isCreator && (
                <div className="space-y-3">
                  <div className="p-4 bg-green-950 border border-green-800 rounded-lg text-center">
                    <p className="text-green-300 font-medium">✓ You're participating in this route</p>
                  </div>
                  {route.status === 'planned' && (
                    <button
                      onClick={handleLeaveRoute}
                      disabled={joining}
                      className="w-full flex items-center justify-center gap-2 px-6 py-3 bg-red-600 text-white rounded-lg hover:bg-red-700 transition font-medium disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      <span>{joining ? 'Leaving...' : 'Leave Route'}</span>
                    </button>
                  )}
                </div>
              )}

              {isCreator && (
                <div className="p-4 bg-blue-950 border border-blue-800 rounded-lg text-center">
                  <p className="text-blue-300 font-medium">You created this route</p>
                </div>
              )}
            </div>

            {/* Route Map */}
            <div className="bg-gray-800 rounded-2xl shadow-sm border border-gray-700 overflow-hidden">
              <div className="p-6 border-b border-gray-700">
                <h2 className="text-xl font-bold text-white flex items-center gap-2">
                  <MapPin className="w-6 h-6 text-primary-600" />
                  Route Map
                </h2>
                {route.distance && (
                  <p className="text-sm text-gray-400 mt-1">
                    {formatDistance(route.distance)} • {formatDuration(route.duration)}
                  </p>
                )}
              </div>
              <div ref={mapContainer} className="w-full h-[500px]" />
            </div>
          </div>

          {/* Participants Sidebar */}
          <div className="lg:col-span-1">
            <div className="bg-gray-800 rounded-2xl shadow-sm border border-gray-700 p-6 sticky top-8">
              <h2 className="text-xl font-bold text-white mb-4">
                Participants ({participants.length})
              </h2>
              {participants.length === 0 ? (
                <p className="text-gray-400 text-center py-8">No participants yet</p>
              ) : (
                <div className="space-y-3">
                  {participants.map((participant) => (
                    <div
                      key={participant.id}
                      className="flex items-center gap-3 p-3 bg-gray-900 rounded-lg"
                    >
                      <div className="w-10 h-10 bg-primary-600 rounded-full flex items-center justify-center text-white font-medium">
                        {participant.user?.display_name?.[0]?.toUpperCase() || '?'}
                      </div>
                      <div className="flex-1">
                        <p className="font-medium text-white">
                          {participant.user?.display_name || 'Unknown User'}
                          {participant.user_id === route.creator_id && (
                            <span className="ml-2 text-xs bg-primary-950 text-primary-300 border border-primary-800 px-2 py-1 rounded">
                              Creator
                            </span>
                          )}
                        </p>
                        <p className="text-sm text-gray-400">
                          Joined {new Date(participant.joined_at).toLocaleDateString()}
                        </p>
                      </div>
                      {isCreator && participant.user_id !== route.creator_id && route.status === 'planned' && (
                        <button
                          onClick={() => handleRemoveParticipant(participant.user_id, participant.user?.display_name)}
                          className="p-2 text-red-400 hover:text-red-300 hover:bg-red-950 rounded transition"
                          title="Remove participant"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      )}
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
