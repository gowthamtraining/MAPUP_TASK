import React, { useState } from 'react';
import { Plus, Filter, MapPin, X, Check, PenTool } from 'lucide-react';
import { api } from '../api';

const CATEGORIES = [
  { id: 'delivery_zone', label: 'Delivery Zone', color: 'bg-blue-500' },
  { id: 'restricted_zone', label: 'Restricted Zone', color: 'bg-rose-500' },
  { id: 'toll_zone', label: 'Toll Zone', color: 'bg-amber-500' },
  { id: 'customer_area', label: 'Customer Area', color: 'bg-emerald-500' },
];

export default function GeofenceManager({
  geofences = [],
  onGeofenceCreated,
  mapMode,
  setMapMode,
  drawnPoints,
  setDrawnPoints,
}) {
  const [showModal, setShowModal] = useState(false);
  const [filterCategory, setFilterCategory] = useState('');
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [category, setCategory] = useState('delivery_zone');
  const [coordText, setCoordText] = useState('');
  const [loading, setLoading] = useState(false);
  const [filterLoading, setFilterLoading] = useState(false);
  const [displayGeofences, setDisplayGeofences] = useState(geofences);
  const [error, setError] = useState('');

  // Fetch geofences from backend with optional ?category= filter query
  React.useEffect(() => {
    let isMounted = true;
    const fetchFiltered = async () => {
      setFilterLoading(true);
      try {
        const res = await api.getGeofences(filterCategory);
        if (isMounted) {
          setDisplayGeofences(res.geofences || []);
        }
      } catch (err) {
        console.error('Failed to filter geofences by category:', err);
      } finally {
        if (isMounted) setFilterLoading(false);
      }
    };
    fetchFiltered();
    return () => {
      isMounted = false;
    };
  }, [filterCategory, geofences]);

  // Handle open modal from drawn points
  const handleOpenModal = () => {
    setError('');
    if (drawnPoints.length >= 3) {
      // Auto close polygon if needed
      const closed = [...drawnPoints];
      const first = closed[0];
      const last = closed[closed.length - 1];
      if (first[0] !== last[0] || first[1] !== last[1]) {
        closed.push([first[0], first[1]]);
      }
      setCoordText(JSON.stringify(closed, null, 2));
    } else {
      // Default sample coordinates around SF
      setCoordText(
        JSON.stringify(
          [
            [37.7749, -122.4194],
            [37.7849, -122.4194],
            [37.7849, -122.4094],
            [37.7749, -122.4094],
            [37.7749, -122.4194],
          ],
          null,
          2
        )
      );
    }
    setShowModal(true);
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    try {
      let parsedCoords;
      try {
        parsedCoords = JSON.parse(coordText);
      } catch (err) {
        throw new Error('Coordinates must be valid JSON array of [lat, lon] pairs');
      }

      const res = await api.createGeofence({
        name,
        description,
        coordinates: parsedCoords,
        category,
      });

      setShowModal(false);
      setName('');
      setDescription('');
      setDrawnPoints([]);
      setMapMode('view');
      onGeofenceCreated && onGeofenceCreated(res);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const filteredGeofences = displayGeofences;

  return (
    <div className="bg-slate-900 border border-slate-800 rounded-xl p-4 shadow-xl flex flex-col h-full">
      <div className="flex items-center justify-between pb-3 border-b border-slate-800 mb-3">
        <div>
          <h2 className="text-sm font-bold text-white tracking-wide flex items-center gap-2">
            <MapPin className="w-4 h-4 text-blue-400" />
            Geofence Zones
          </h2>
          <p className="text-[11px] text-slate-400">Total Zones: {geofences.length}</p>
        </div>

        <div className="flex items-center gap-2">
          {mapMode === 'draw' ? (
            <div className="flex items-center gap-1.5">
              <button
                onClick={handleOpenModal}
                disabled={drawnPoints.length < 3}
                className="bg-emerald-600 hover:bg-emerald-500 disabled:opacity-50 text-white text-xs font-semibold px-2.5 py-1.5 rounded-lg shadow transition-colors flex items-center gap-1"
              >
                <Check className="w-3.5 h-3.5" />
                Save ({drawnPoints.length} pts)
              </button>
              <button
                onClick={() => {
                  setMapMode('view');
                  setDrawnPoints([]);
                }}
                className="bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs px-2 py-1.5 rounded-lg"
              >
                Cancel
              </button>
            </div>
          ) : (
            <div className="flex items-center gap-1.5">
              <button
                onClick={() => {
                  setDrawnPoints([]);
                  setMapMode('draw');
                }}
                className="bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-medium px-2.5 py-1.5 rounded-lg border border-slate-700 transition-colors flex items-center gap-1"
                title="Draw on map"
              >
                <PenTool className="w-3.5 h-3.5 text-rose-400" />
                Draw
              </button>
              <button
                onClick={handleOpenModal}
                className="bg-blue-600 hover:bg-blue-500 text-white text-xs font-semibold px-2.5 py-1.5 rounded-lg shadow transition-colors flex items-center gap-1"
              >
                <Plus className="w-3.5 h-3.5" />
                New Zone
              </button>
            </div>
          )}
        </div>
      </div>

      {/* Category Filter Pills */}
      <div className="flex items-center gap-1.5 pb-2 overflow-x-auto text-[11px]">
        <button
          onClick={() => setFilterCategory('')}
          className={`px-2.5 py-1 rounded-full whitespace-nowrap transition-colors ${
            filterCategory === ''
              ? 'bg-blue-600 text-white font-semibold'
              : 'bg-slate-800 text-slate-400 hover:bg-slate-700'
          }`}
        >
          All ({geofences.length})
        </button>
        {CATEGORIES.map((cat) => (
          <button
            key={cat.id}
            onClick={() => setFilterCategory(cat.id)}
            className={`px-2.5 py-1 rounded-full whitespace-nowrap transition-colors ${
              filterCategory === cat.id
                ? `${cat.color} text-white font-semibold`
                : 'bg-slate-800 text-slate-400 hover:bg-slate-700'
            }`}
          >
            {cat.label}
          </button>
        ))}
      </div>

      {/* Geofences List */}
      <div className="flex-1 overflow-y-auto space-y-2 mt-2 pr-1">
        {filteredGeofences.length === 0 ? (
          <div className="text-center py-8 text-xs text-slate-500">
            No geofences found for selected filter
          </div>
        ) : (
          filteredGeofences.map((g) => {
            const cat = CATEGORIES.find((c) => c.id === g.category);
            return (
              <div
                key={g.id}
                className="bg-slate-800/60 border border-slate-700/60 hover:border-slate-600 rounded-lg p-2.5 text-xs transition-all"
              >
                <div className="flex items-center justify-between mb-1">
                  <span className="font-semibold text-slate-100">{g.name}</span>
                  <span
                    className={`text-[9px] uppercase font-bold tracking-wider px-2 py-0.5 rounded text-white ${
                      cat ? cat.color : 'bg-slate-600'
                    }`}
                  >
                    {g.category.replace('_', ' ')}
                  </span>
                </div>
                {g.description && <p className="text-[11px] text-slate-400 mb-1">{g.description}</p>}
                <div className="flex items-center justify-between text-[10px] text-slate-500 font-mono mt-1 pt-1 border-t border-slate-700/40">
                  <span>ID: {g.id}</span>
                  <span>{g.coordinates?.length || 0} vertices</span>
                </div>
              </div>
            );
          })
        )}
      </div>

      {/* Create Modal */}
      {showModal && (
        <div className="fixed inset-0 z-[2000] bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-slate-900 border border-slate-700 rounded-xl w-full max-w-lg p-5 shadow-2xl relative">
            <button
              onClick={() => setShowModal(false)}
              className="absolute top-4 right-4 text-slate-400 hover:text-white"
            >
              <X className="w-5 h-5" />
            </button>

            <h3 className="text-base font-bold text-white mb-1">Create Geofence Boundary</h3>
            <p className="text-xs text-slate-400 mb-4">
              Define virtual polygon perimeter. Coordinates must be a closed polygon.
            </p>

            {error && (
              <div className="mb-3 p-2.5 bg-rose-950/60 border border-rose-500/50 rounded-lg text-rose-300 text-xs">
                {error}
              </div>
            )}

            <form onSubmit={handleSubmit} className="space-y-3 text-xs">
              <div>
                <label className="block text-slate-300 font-medium mb-1">Geofence Name *</label>
                <input
                  type="text"
                  required
                  placeholder="e.g. Downtown Delivery Zone"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-blue-500"
                />
              </div>

              <div>
                <label className="block text-slate-300 font-medium mb-1">Category *</label>
                <select
                  value={category}
                  onChange={(e) => setCategory(e.target.value)}
                  className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-blue-500"
                >
                  {CATEGORIES.map((cat) => (
                    <option key={cat.id} value={cat.id}>
                      {cat.label}
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-slate-300 font-medium mb-1">Description (Optional)</label>
                <input
                  type="text"
                  placeholder="e.g. High priority delivery zone for downtown orders"
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-blue-500"
                />
              </div>

              <div>
                <div className="flex items-center justify-between mb-1">
                  <label className="block text-slate-300 font-medium">
                    Coordinates (JSON [[lat, lon], ...]) *
                  </label>
                  <span className="text-[10px] text-slate-500">First point must equal last</span>
                </div>
                <textarea
                  rows={5}
                  required
                  value={coordText}
                  onChange={(e) => setCoordText(e.target.value)}
                  className="w-full font-mono text-[11px] bg-slate-800 border border-slate-700 rounded-lg p-2.5 text-slate-200 focus:outline-none focus:border-blue-500"
                />
              </div>

              <div className="flex items-center justify-end gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setShowModal(false)}
                  className="px-4 py-2 bg-slate-800 hover:bg-slate-700 text-slate-300 rounded-lg font-medium"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={loading}
                  className="px-4 py-2 bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white font-semibold rounded-lg shadow"
                >
                  {loading ? 'Creating...' : 'Create Geofence'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
