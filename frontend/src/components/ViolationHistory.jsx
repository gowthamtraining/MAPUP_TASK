import React, { useState, useEffect } from 'react';
import { History, Filter, RefreshCw, Calendar, ChevronLeft, ChevronRight } from 'lucide-react';
import { api } from '../api';

export default function ViolationHistory({ vehicles = [], geofences = [] }) {
  const [violations, setViolations] = useState([]);
  const [totalCount, setTotalCount] = useState(0);
  const [loading, setLoading] = useState(false);
  const [vehicleId, setVehicleId] = useState('');
  const [geofenceId, setGeofenceId] = useState('');
  const [limit, setLimit] = useState('50');

  const fetchHistory = async () => {
    setLoading(true);
    try {
      const res = await api.getViolations({
        vehicle_id: vehicleId,
        geofence_id: geofenceId,
        limit,
      });
      setViolations(res.violations || []);
      setTotalCount(res.total_count || 0);
    } catch (err) {
      console.error('Failed to load violation history:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchHistory();
  }, [vehicleId, geofenceId, limit]);

  return (
    <div className="bg-slate-900 border border-slate-800 rounded-xl p-5 shadow-xl flex flex-col h-full">
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 pb-4 border-b border-slate-800 mb-4">
        <div>
          <h2 className="text-base font-bold text-white tracking-wide flex items-center gap-2">
            <History className="w-5 h-5 text-rose-400" />
            Violation & Event History Log
          </h2>
          <p className="text-xs text-slate-400 mt-0.5">
            Audit record of all boundary entries and exits. Total Events: {totalCount}
          </p>
        </div>

        <button
          onClick={fetchHistory}
          disabled={loading}
          className="bg-slate-800 hover:bg-slate-700 disabled:opacity-50 text-slate-300 hover:text-white text-xs font-medium px-3 py-1.5 rounded-lg border border-slate-700 flex items-center gap-1.5 transition-colors self-end sm:self-auto"
        >
          <RefreshCw className={`w-3.5 h-3.5 ${loading ? 'animate-spin' : ''}`} />
          Refresh
        </button>
      </div>

      {/* Filters Bar */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 mb-4 text-xs">
        <div>
          <label className="block text-slate-400 mb-1 font-medium">Filter by Vehicle</label>
          <select
            value={vehicleId}
            onChange={(e) => setVehicleId(e.target.value)}
            className="w-full bg-slate-800 border border-slate-700 rounded-lg px-2.5 py-1.5 text-slate-200 focus:outline-none focus:border-rose-500"
          >
            <option value="">All Vehicles</option>
            {vehicles.map((v) => (
              <option key={v.id} value={v.id}>
                {v.vehicle_number} ({v.driver_name})
              </option>
            ))}
          </select>
        </div>

        <div>
          <label className="block text-slate-400 mb-1 font-medium">Filter by Geofence</label>
          <select
            value={geofenceId}
            onChange={(e) => setGeofenceId(e.target.value)}
            className="w-full bg-slate-800 border border-slate-700 rounded-lg px-2.5 py-1.5 text-slate-200 focus:outline-none focus:border-rose-500"
          >
            <option value="">All Geofences</option>
            {geofences.map((g) => (
              <option key={g.id} value={g.id}>
                {g.name}
              </option>
            ))}
          </select>
        </div>

        <div>
          <label className="block text-slate-400 mb-1 font-medium">Records Limit</label>
          <select
            value={limit}
            onChange={(e) => setLimit(e.target.value)}
            className="w-full bg-slate-800 border border-slate-700 rounded-lg px-2.5 py-1.5 text-slate-200 focus:outline-none focus:border-rose-500"
          >
            <option value="25">25 records</option>
            <option value="50">50 records</option>
            <option value="100">100 records</option>
            <option value="250">250 records</option>
          </select>
        </div>
      </div>

      {/* Table view */}
      <div className="flex-1 overflow-x-auto overflow-y-auto rounded-lg border border-slate-800">
        <table className="w-full text-left text-xs text-slate-300">
          <thead className="bg-slate-800/80 text-[11px] uppercase tracking-wider text-slate-400 sticky top-0">
            <tr>
              <th className="py-2.5 px-3">Event</th>
              <th className="py-2.5 px-3">Vehicle</th>
              <th className="py-2.5 px-3">Geofence</th>
              <th className="py-2.5 px-3">Coordinates</th>
              <th className="py-2.5 px-3">Timestamp</th>
              <th className="py-2.5 px-3 text-right">Event ID</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-800 font-normal">
            {violations.length === 0 ? (
              <tr>
                <td colSpan={6} className="text-center py-10 text-slate-500">
                  {loading ? 'Loading violation history...' : 'No violation events found matching criteria'}
                </td>
              </tr>
            ) : (
              violations.map((viol) => {
                const isEntry = viol.event_type?.toLowerCase() === 'entry';
                return (
                  <tr key={viol.id} className="hover:bg-slate-800/40 transition-colors">
                    <td className="py-2.5 px-3">
                      <span
                        className={`inline-block text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded ${
                          isEntry
                            ? 'bg-emerald-950 text-emerald-400 border border-emerald-500/30'
                            : 'bg-amber-950 text-amber-400 border border-amber-500/30'
                        }`}
                      >
                        {viol.event_type}
                      </span>
                    </td>
                    <td className="py-2.5 px-3 font-semibold text-slate-200">
                      {viol.vehicle_number}
                    </td>
                    <td className="py-2.5 px-3 text-slate-300">{viol.geofence_name}</td>
                    <td className="py-2.5 px-3 font-mono text-[11px] text-slate-400">
                      {viol.latitude?.toFixed(4)}, {viol.longitude?.toFixed(4)}
                    </td>
                    <td className="py-2.5 px-3 text-slate-400 whitespace-nowrap">
                      {new Date(viol.timestamp).toLocaleString()}
                    </td>
                    <td className="py-2.5 px-3 text-right font-mono text-[10px] text-slate-500">
                      {viol.id}
                    </td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
