import { useState, useEffect, useRef } from 'react';
import { X, MapPin, Trash2 } from 'lucide-react';
import { api } from '../lib/api';
import mapboxgl from 'mapbox-gl';
import 'mapbox-gl/dist/mapbox-gl.css';

mapboxgl.accessToken = import.meta.env.VITE_MAPBOX_TOKEN;

export default function CreateRouteModal({ onClose, onSuccess }) {
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    route_type: 'driving',
    visibility: 'public',
    start_time: '',
    max_participants: '',
    difficulty: '',
  });
  const [waypoints, setWaypoints] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  
  const mapContainer = useRef(null);
  const map = useRef(null);
  const markers = useRef([]);

  // Initialize map
  useEffect(() => {
    if (map.current) return; // Initialize map only once

    map.current = new mapboxgl.Map({
      container: mapContainer.current,
      style: 'mapbox://styles/mapbox/outdoors-v12',
      center: [-98.5795, 39.8283], // Center of USA
      zoom: 4,
    });

    map.current.addControl(new mapboxgl.NavigationControl(), 'top-right');

    // Add click handler to add waypoints
    map.current.on('click', (e) => {
      addWaypoint(e.lngLat.lng, e.lngLat.lat);
    });

    return () => {
      if (map.current) {
        map.current.remove();
        map.current = null;
      }
    };
  }, []);

  // Update route line when waypoints change
  useEffect(() => {
    if (!map.current || waypoints.length < 2) {
      // Remove route line if less than 2 waypoints
      if (map.current && map.current.getSource('route')) {
        map.current.removeLayer('route');
        map.current.removeSource('route');
      }
      return;
    }

    const coordinates = waypoints.map(wp => [wp.longitude, wp.latitude]);

    if (map.current.getSource('route')) {
      map.current.getSource('route').setData({
        type: 'Feature',
        properties: {},
        geometry: {
          type: 'LineString',
          coordinates: coordinates,
        },
      });
    } else {
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
    }
  }, [waypoints]);

  function addWaypoint(lng, lat) {
    const waypointIndex = waypoints.length; // Capture the index before state update
    const newWaypoint = {
      latitude: lat,
      longitude: lng,
      order: waypointIndex,
      name: `Point ${waypointIndex + 1}`,
    };

    setWaypoints(prev => [...prev, newWaypoint]);

    // Add marker
    const el = document.createElement('div');
    el.className = 'waypoint-marker';
    el.style.backgroundColor = waypointIndex === 0 ? '#22c55e' : '#3b82f6';
    el.style.width = '24px';
    el.style.height = '24px';
    el.style.borderRadius = '50%';
    el.style.border = '3px solid white';
    el.style.cursor = 'pointer';
    el.style.boxShadow = '0 2px 4px rgba(0,0,0,0.3)';
    el.innerHTML = `<div style="color: white; font-size: 12px; font-weight: bold; text-align: center; line-height: 18px;">${waypointIndex + 1}</div>`;

    const marker = new mapboxgl.Marker({ element: el, draggable: true })
      .setLngLat([lng, lat])
      .addTo(map.current);

    // Capture the index in the closure
    marker.on('dragend', () => {
      const lngLat = marker.getLngLat();
      updateWaypointPosition(waypointIndex, lngLat.lng, lngLat.lat);
    });

    markers.current.push(marker);
  }

  function updateWaypointPosition(index, lng, lat) {
    setWaypoints(prev => {
      const updated = [...prev];
      updated[index] = {
        ...updated[index],
        latitude: lat,
        longitude: lng,
      };
      return updated;
    });
  }

  function removeWaypoint(index) {
    setWaypoints(prev => prev.filter((_, i) => i !== index).map((wp, i) => ({ ...wp, order: i })));
    
    // Remove marker
    if (markers.current[index]) {
      markers.current[index].remove();
      markers.current.splice(index, 1);
    }
  }

  function clearWaypoints() {
    setWaypoints([]);
    markers.current.forEach(marker => marker.remove());
    markers.current = [];
  }

  function handleChange(e) {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value,
    }));
  }

  async function handleSubmit(e) {
    e.preventDefault();
    setError('');

    // Check if user is authenticated
    const token = api.getToken();
    if (!token) {
      setError('You must be logged in to create a route');
      return;
    }

    if (waypoints.length < 2) {
      setError('Please add at least 2 waypoints by clicking on the map');
      return;
    }

    setLoading(true);

    try {
      // Convert datetime-local format to RFC3339 format
      let startTime = null;
      if (formData.start_time) {
        // datetime-local gives us "2026-04-26T10:00"
        // We need to convert it to "2026-04-26T10:00:00Z"
        startTime = new Date(formData.start_time).toISOString();
      }

      // Normalize waypoint order to ensure sequential values (0, 1, 2, ...)
      const normalizedWaypoints = waypoints.map((wp, index) => ({
        ...wp,
        order: index,
      }));

      const payload = {
        ...formData,
        max_participants: formData.max_participants ? parseInt(formData.max_participants) : null,
        start_time: startTime,
        waypoints: normalizedWaypoints,
      };

      await api.createRoute(payload);
      onSuccess();
    } catch (err) {
      console.error('Failed to create route:', err);
      setError(err.message || 'Failed to create route');
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="fixed inset-0 bg-black bg-opacity-75 flex items-center justify-center p-4 z-50">
      <div className="bg-gray-800 border border-gray-700 rounded-2xl shadow-2xl max-w-6xl w-full max-h-[90vh] overflow-hidden flex flex-col">
        <div className="bg-gray-800 border-b border-gray-700 px-6 py-4 flex items-center justify-between">
          <h2 className="text-2xl font-bold text-white">Create New Route</h2>
          <button
            onClick={onClose}
            className="p-2 hover:bg-gray-700 rounded-lg transition text-gray-400 hover:text-white"
          >
            <X className="w-6 h-6" />
          </button>
        </div>

        <div className="flex-1 overflow-hidden flex">
          {/* Map Section */}
          <div className="w-1/2 relative">
            <div ref={mapContainer} className="w-full h-full" />
            <div className="absolute top-4 left-4 bg-gray-900 bg-opacity-90 rounded-lg p-3 text-sm text-gray-300 max-w-xs">
              <MapPin className="w-4 h-4 inline mr-2" />
              Click on the map to add waypoints. Drag markers to adjust.
            </div>
            {waypoints.length > 0 && (
              <div className="absolute bottom-4 left-4 right-4 bg-gray-900 bg-opacity-90 rounded-lg p-3">
                <div className="flex items-center justify-between mb-2">
                  <span className="text-sm font-medium text-white">Waypoints ({waypoints.length})</span>
                  <button
                    type="button"
                    onClick={clearWaypoints}
                    className="text-xs text-red-400 hover:text-red-300"
                  >
                    Clear All
                  </button>
                </div>
                <div className="space-y-1 max-h-32 overflow-y-auto">
                  {waypoints.map((wp, index) => (
                    <div key={index} className="flex items-center justify-between text-xs text-gray-300 bg-gray-800 rounded px-2 py-1">
                      <span className="flex items-center gap-2">
                        <span className="w-5 h-5 rounded-full bg-blue-600 text-white flex items-center justify-center text-[10px] font-bold">
                          {index + 1}
                        </span>
                        {wp.latitude.toFixed(4)}, {wp.longitude.toFixed(4)}
                      </span>
                      <button
                        type="button"
                        onClick={() => removeWaypoint(index)}
                        className="text-red-400 hover:text-red-300"
                      >
                        <Trash2 className="w-3 h-3" />
                      </button>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>

          {/* Form Section */}
          <div className="w-1/2 overflow-y-auto">
            <form onSubmit={handleSubmit} className="p-6 space-y-6">
              {error && (
                <div className="p-4 bg-red-950 border border-red-800 rounded-lg text-sm text-red-200">
                  {error}
                </div>
              )}

          <div>
            <label htmlFor="name" className="block text-sm font-medium text-gray-300 mb-2">
              Route Name *
            </label>
            <input
              id="name"
              name="name"
              type="text"
              value={formData.name}
              onChange={handleChange}
              required
              className="w-full px-4 py-3 bg-gray-900 border border-gray-600 text-white rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none placeholder-gray-500"
              placeholder="Sunday Morning Ride"
            />
          </div>

          <div>
            <label htmlFor="description" className="block text-sm font-medium text-gray-300 mb-2">
              Description
            </label>
            <textarea
              id="description"
              name="description"
              value={formData.description}
              onChange={handleChange}
              rows={3}
              className="w-full px-4 py-3 bg-gray-900 border border-gray-600 text-white rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none resize-none placeholder-gray-500"
              placeholder="A scenic route through the countryside..."
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label htmlFor="route_type" className="block text-sm font-medium text-gray-300 mb-2">
                Route Type *
              </label>
              <select
                id="route_type"
                name="route_type"
                value={formData.route_type}
                onChange={handleChange}
                required
                className="w-full px-4 py-3 bg-gray-900 border border-gray-600 text-white rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none"
              >
                <option value="driving">🚗 Driving</option>
                <option value="cycling">🚴 Cycling</option>
                <option value="walking">🚶 Walking</option>
                <option value="running">🏃 Running</option>
              </select>
            </div>

            <div>
              <label htmlFor="visibility" className="block text-sm font-medium text-gray-300 mb-2">
                Visibility *
              </label>
              <select
                id="visibility"
                name="visibility"
                value={formData.visibility}
                onChange={handleChange}
                required
                className="w-full px-4 py-3 bg-gray-900 border border-gray-600 text-white rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none"
              >
                <option value="public">Public</option>
                <option value="private">Private</option>
                <option value="invite_only">Invite Only</option>
              </select>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label htmlFor="start_time" className="block text-sm font-medium text-gray-300 mb-2">
                Start Time
              </label>
              <input
                id="start_time"
                name="start_time"
                type="datetime-local"
                value={formData.start_time}
                onChange={handleChange}
                className="w-full px-4 py-3 bg-gray-900 border border-gray-600 text-white rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none"
              />
            </div>

            <div>
              <label htmlFor="max_participants" className="block text-sm font-medium text-gray-300 mb-2">
                Max Participants
              </label>
              <input
                id="max_participants"
                name="max_participants"
                type="number"
                min="1"
                value={formData.max_participants}
                onChange={handleChange}
                className="w-full px-4 py-3 bg-gray-900 border border-gray-600 text-white rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none placeholder-gray-500"
                placeholder="Unlimited"
              />
            </div>
          </div>

          <div>
            <label htmlFor="difficulty" className="block text-sm font-medium text-gray-300 mb-2">
              Difficulty
            </label>
            <select
              id="difficulty"
              name="difficulty"
              value={formData.difficulty}
              onChange={handleChange}
              className="w-full px-4 py-3 bg-gray-900 border border-gray-600 text-white rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 outline-none"
            >
              <option value="">Not specified</option>
              <option value="easy">Easy</option>
              <option value="moderate">Moderate</option>
              <option value="hard">Hard</option>
              <option value="expert">Expert</option>
            </select>
          </div>

              <div className="flex gap-3 pt-4">
                <button
                  type="button"
                  onClick={onClose}
                  className="flex-1 px-6 py-3 border border-gray-600 text-gray-300 rounded-lg hover:bg-gray-700 transition font-medium"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={loading}
                  className="flex-1 px-6 py-3 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition font-medium disabled:opacity-50 disabled:cursor-not-allowed shadow-lg shadow-primary-900/30"
                >
                  {loading ? 'Creating...' : 'Create Route'}
                </button>
              </div>
            </form>
          </div>
        </div>
      </div>
    </div>
  );
}
