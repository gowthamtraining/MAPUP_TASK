import React, { useState } from 'react';
import { Truck, Plus, X, Navigation, Phone, User, CheckCircle } from 'lucide-react';
import { api } from '../api';

const VEHICLE_TYPES = ['truck', 'van', 'car', 'motorcycle'];

export default function VehicleManager({
  vehicles = [],
  vehicleLocations = {},
  selectedVehicleId,
  onSelectVehicle,
  onVehicleRegistered,
  mapMode,
  setMapMode,
  onLocationUpdated,
}) {
  const [showModal, setShowModal] = useState(false);
  const [vehicleNumber, setVehicleNumber] = useState('');
  const [driverName, setDriverName] = useState('');
  const [vehicleType, setVehicleType] = useState('truck');
  const [phone, setPhone] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  // Manual coordinate update modal state
  const [showMoveModal, setShowMoveModal] = useState(false);
  const [moveLat, setMoveLat] = useState('37.7849');
  const [moveLon, setMoveLon] = useState('-122.4194');

  const handleRegister = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    try {
      const res = await api.createVehicle({
        vehicle_number: vehicleNumber,
        driver_name: driverName,
        vehicle_type: vehicleType,
        phone,
      });

      setShowModal(false);
      setVehicleNumber('');
      setDriverName('');
      setPhone('');
      onVehicleRegistered && onVehicleRegistered(res);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleManualMove = async (e) => {
    e.preventDefault();
    if (!selectedVehicleId) return;
    setLoading(true);
    try {
      await api.updateLocation(selectedVehicleId, moveLat, moveLon);
      setShowMoveModal(false);
      onLocationUpdated && onLocationUpdated();
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
            <Truck className="w-4 h-4 text-purple-400" />
            Fleet Vehicles
          </h2>
          <p className="text-[11px] text-slate-400">Total Registered: {vehicles.length}</p>
        </div>

        <button
          onClick={() => {
            setError('');
            setShowModal(true);
          }}
          className="bg-purple-600 hover:bg-purple-500 text-white text-xs font-semibold px-2.5 py-1.5 rounded-lg shadow transition-colors flex items-center gap-1"
        >
          <Plus className="w-3.5 h-3.5" />
          Add Vehicle
        </button>
      </div>

      {/* Vehicle Cards List */}
      <div className="flex-1 overflow-y-auto space-y-2 pr-1">
        {vehicles.length === 0 ? (
          <div className="text-center py-8 text-xs text-slate-500">
            No vehicles registered yet
          </div>
        ) : (
          vehicles.map((v) => {
            const isSelected = selectedVehicleId === v.id;
            const locData = vehicleLocations[v.id];
            const currentGeofences = locData?.current_geofences || [];

            return (
              <div
                key={v.id}
                onClick={() => onSelectVehicle && onSelectVehicle(v.id)}
                className={`p-3 rounded-lg border text-xs cursor-pointer transition-all ${
                  isSelected
                    ? 'bg-purple-950/40 border-purple-500/80 shadow-md shadow-purple-950/30'
                    : 'bg-slate-800/60 border-slate-700/60 hover:border-slate-600'
                }`}
              >
                <div className="flex items-center justify-between mb-1">
                  <div className="flex items-center gap-2">
                    <span className="font-bold text-slate-100">{v.vehicle_number}</span>
                    <span className="text-[10px] bg-slate-700/80 text-slate-300 px-1.5 py-0.5 rounded capitalize">
                      {v.vehicle_type}
                    </span>
                  </div>
                  <span className="text-[10px] text-emerald-400 flex items-center gap-1">
                    <CheckCircle className="w-3 h-3" />
                    {v.status}
                  </span>
                </div>

                <div className="text-[11px] text-slate-400 space-y-0.5 mb-2">
                  <div className="flex items-center gap-1.5">
                    <User className="w-3 h-3 text-slate-500" />
                    <span>{v.driver_name}</span>
                  </div>
                  <div className="flex items-center gap-1.5">
                    <Phone className="w-3 h-3 text-slate-500" />
                    <span>{v.phone}</span>
                  </div>
                </div>

                {/* Geofence Status */}
                <div className="pt-1.5 border-t border-slate-700/40 flex items-center justify-between">
                  <div className="text-[10px]">
                    {currentGeofences.length > 0 ? (
                      <span className="text-blue-400 font-medium">
                        Inside: {currentGeofences.map((cg) => cg.geofence_name).join(', ')}
                      </span>
                    ) : (
                      <span className="text-slate-500 italic">Outside geofences</span>
                    )}
                  </div>

                  {isSelected && (
                    <div className="flex items-center gap-1">
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          setMapMode(mapMode === 'teleport' ? 'view' : 'teleport');
                        }}
                        className={`text-[10px] px-2 py-0.5 rounded font-medium border flex items-center gap-1 transition-colors ${
                          mapMode === 'teleport'
                            ? 'bg-purple-600 text-white border-purple-400'
                            : 'bg-slate-700 text-purple-300 border-purple-500/40 hover:bg-slate-600'
                        }`}
                        title="Click map to move vehicle"
                      >
                        <Navigation className="w-2.5 h-2.5" />
                        {mapMode === 'teleport' ? 'Click Map' : 'Move'}
                      </button>
                    </div>
                  )}
                </div>
              </div>
            );
          })
        )}
      </div>

      {/* Register Vehicle Modal */}
      {showModal && (
        <div className="fixed inset-0 z-[2000] bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-slate-900 border border-slate-700 rounded-xl w-full max-w-md p-5 shadow-2xl relative">
            <button
              onClick={() => setShowModal(false)}
              className="absolute top-4 right-4 text-slate-400 hover:text-white"
            >
              <X className="w-5 h-5" />
            </button>

            <h3 className="text-base font-bold text-white mb-1">Register New Vehicle</h3>
            <p className="text-xs text-slate-400 mb-4">
              Add vehicle to the tracking fleet and assign a driver.
            </p>

            {error && (
              <div className="mb-3 p-2 bg-rose-950/60 border border-rose-500/50 rounded-lg text-rose-300 text-xs">
                {error}
              </div>
            )}

            <form onSubmit={handleRegister} className="space-y-3 text-xs">
              <div>
                <label className="block text-slate-300 font-medium mb-1">Vehicle Number *</label>
                <input
                  type="text"
                  required
                  placeholder="e.g. KA-01-AB-1234"
                  value={vehicleNumber}
                  onChange={(e) => setVehicleNumber(e.target.value)}
                  className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-purple-500"
                />
              </div>

              <div>
                <label className="block text-slate-300 font-medium mb-1">Driver Name *</label>
                <input
                  type="text"
                  required
                  placeholder="e.g. John Doe"
                  value={driverName}
                  onChange={(e) => setDriverName(e.target.value)}
                  className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-purple-500"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-slate-300 font-medium mb-1">Vehicle Type *</label>
                  <select
                    value={vehicleType}
                    onChange={(e) => setVehicleType(e.target.value)}
                    className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-purple-500"
                  >
                    {VEHICLE_TYPES.map((t) => (
                      <option key={t} value={t}>
                        {t.toUpperCase()}
                      </option>
                    ))}
                  </select>
                </div>

                <div>
                  <label className="block text-slate-300 font-medium mb-1">Phone *</label>
                  <input
                    type="text"
                    required
                    placeholder="+1234567890"
                    value={phone}
                    onChange={(e) => setPhone(e.target.value)}
                    className="w-full bg-slate-800 border border-slate-700 rounded-lg px-3 py-2 text-white focus:outline-none focus:border-purple-500"
                  />
                </div>
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
                  className="px-4 py-2 bg-purple-600 hover:bg-purple-500 disabled:opacity-50 text-white font-semibold rounded-lg shadow"
                >
                  {loading ? 'Registering...' : 'Register Vehicle'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
