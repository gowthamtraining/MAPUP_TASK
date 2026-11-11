import React, { useState } from 'react';
import { BellRing, Plus, X, Shield, CheckCircle2 } from 'lucide-react';
import { api } from '../api';

export default function AlertRuleManager({
  alertRules = [],
  geofences = [],
  vehicles = [],
  onRuleCreated,
}) {
  const [showModal, setShowModal] = useState(false);
  const [geofenceId, setGeofenceId] = useState('');
  const [vehicleId, setVehicleId] = useState('');
  const [eventType, setEventType] = useState('entry');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleOpen = () => {
    setError('');
    if (geofences.length > 0 && !geofenceId) {
      setGeofenceId(geofences[0].id);
    }
    setShowModal(true);
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    try {
      const payload = {
        geofence_id: geofenceId,
        event_type: eventType,
      };
      if (vehicleId) {
        payload.vehicle_id = vehicleId;
      }

      const res = await api.configureAlert(payload);
      setShowModal(false);
      onRuleCreated && onRuleCreated(res);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="bg-slate-900 border border-slate-800 rounded-xl p-4 shadow-xl flex flex-col h-full">
      <div className="flex items-center justify-between pb-3 border-b border-slate-800 mb-3">
        <div>
          <h2 className="text-sm font-bold text-white tracking-wide flex items-center gap-2">
            <BellRing className="w-4 h-4 text-amber-400" />
            Alert Rule Policies
          </h2>
          <p className="text-[11px] text-slate-400">Active Alert Rules: {alertRules.length}</p>
        </div>

        <button
          onClick={handleOpen}
          className="bg-amber-600 hover:bg-amber-500 text-white text-xs font-semibold px-2.5 py-1.5 rounded-lg shadow transition-colors flex items-center gap-1"
        >
          <Plus className="w-3.5 h-3.5" />
          Add Rule
        </button>
      </div>

      <div className="flex-1 overflow-y-auto space-y-2 pr-1">
        {alertRules.length === 0 ? (
          <div className="text-center py-8 text-xs text-slate-500">
            No custom rules configured yet (Restricted zone breach alerts trigger automatically)
          </div>
        ) : (
          alertRules.map((rule) => (
            <div
              key={rule.alert_id}
              className="bg-slate-800/60 border border-slate-700/60 rounded-lg p-2.5 text-xs transition-all"
            >
              <div className="flex items-center justify-between mb-1">
                <span className="font-semibold text-slate-200">
                  {rule.geofence_name || rule.geofence_id}
                </span>
                <span className="text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded bg-amber-950/60 border border-amber-500/40 text-amber-300">
                  {rule.event_type}
                </span>
              </div>

              <div className="text-[11px] text-slate-400 flex items-center justify-between mt-1">
                <span>
                  Target:{' '}
                  <strong className="text-slate-300">
                    {rule.vehicle_number || (rule.vehicle_id ? rule.vehicle_id : 'All Fleet Vehicles')}
                  </strong>
                </span>
                <span className="text-emerald-400 text-[10px] flex items-center gap-1">
                  <CheckCircle2 className="w-3 h-3" />
                  active
                </span>
              </div>
            </div>
          ))
        )}
      </div>

      {showModal && (
        <div className="fixed inset-0 z-[2000] bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-slate-900 border border-slate-700 rounded-xl w-full max-w-md p-5 shadow-2xl relative">
            <button
              onClick={() => setShowModal(false)}
              className="absolute top-4 right-4 text-slate-400 hover:text-white"
            >
              <X className="w-5 h-5" />
            </button>

            <h3 className="text-base font-bold text-white mb-1">Configure Geofence Alert Rule</h3>
            <p className="text-xs text-slate-400 mb-4">
              Trigger real-time notifications on entry, exit, or both for selected zone.
            </p>

            {error && (
              <div className="mb-3 p-2 bg-rose-950/60 border border-rose-500/50 rounded-lg text-rose-300 text-xs">
                {error}
              </div>
            )}

            <form onSubmit={handleSubmit} className="space-y-3 text-xs">
              <div>
                <label className="block text-slate-300 font-medium mb-1">Monitored Geofence *</label>
                <select
                  required
                  value={geofenceId}
                  onChange={(e) => setGeofenceId(e.target.value)}
                  className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-amber-500"
                >
                  <option value="" disabled>
                    Select a Geofence
                  </option>
                  {geofences.map((g) => (
                    <option key={g.id} value={g.id}>
                      {g.name} ({g.category})
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-slate-300 font-medium mb-1">Vehicle Scope</label>
                <select
                  value={vehicleId}
                  onChange={(e) => setVehicleId(e.target.value)}
                  className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-amber-500"
                >
                  <option value="">All Vehicles (Global Rule)</option>
                  {vehicles.map((v) => (
                    <option key={v.id} value={v.id}>
                      {v.vehicle_number} ({v.driver_name})
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-slate-300 font-medium mb-1">Event Type *</label>
                <select
                  value={eventType}
                  onChange={(e) => setEventType(e.target.value)}
                  className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-amber-500"
                >
                  <option value="entry">Entry Only</option>
                  <option value="exit">Exit Only</option>
                  <option value="both">Both Entry & Exit</option>
                </select>
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
                  className="px-4 py-2 bg-amber-600 hover:bg-amber-500 disabled:opacity-50 text-white font-semibold rounded-lg shadow"
                >
                  {loading ? 'Saving...' : 'Save Rule'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
