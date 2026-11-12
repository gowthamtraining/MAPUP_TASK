import React, { useState, useEffect, useCallback, useRef } from 'react';
import {
  MapPin,
  Truck,
  Bell,
  History,
  Shield,
  Layers,
  Radio,
  ExternalLink,
  ChevronRight,
  Info,
} from 'lucide-react';
import { api, WS_BASE } from './api';
import MapView from './components/MapView';
import AlertFeed from './components/AlertFeed';
import GeofenceManager from './components/GeofenceManager';
import VehicleManager from './components/VehicleManager';
import AlertRuleManager from './components/AlertRuleManager';
import ViolationHistory from './components/ViolationHistory';
import SimulationBar from './components/SimulationBar';

export default function App() {
  const [activeTab, setActiveTab] = useState('dashboard'); // 'dashboard' | 'history'
  const [activeSidebarTab, setActiveSidebarTab] = useState('alerts'); // 'alerts' | 'geofences' | 'vehicles' | 'rules'

  const [geofences, setGeofences] = useState([]);
  const [vehicles, setVehicles] = useState([]);
  const [vehicleLocations, setVehicleLocations] = useState({});
  const [alertRules, setAlertRules] = useState([]);
  const [liveAlerts, setLiveAlerts] = useState([]);

  const [selectedVehicleId, setSelectedVehicleId] = useState('');
  const [mapMode, setMapMode] = useState('view'); // 'view' | 'draw' | 'teleport'
  const [drawnPoints, setDrawnPoints] = useState([]);

  const [wsConnected, setWsConnected] = useState(false);
  const [toastAlert, setToastAlert] = useState(null);

  const wsRef = useRef(null);

  // Load initial data
  const loadGeofences = useCallback(async () => {
    try {
      const res = await api.getGeofences();
      setGeofences(res.geofences || []);
    } catch (err) {
      console.error('Failed to load geofences:', err);
    }
  }, []);

  const loadVehicles = useCallback(async () => {
    try {
      const res = await api.getVehicles();
      const list = res.vehicles || [];
      setVehicles(list);
      if (list.length > 0 && !selectedVehicleId) {
        setSelectedVehicleId(list[0].id);
      }
      // Load locations for all vehicles
      list.forEach(async (v) => {
        try {
          const locRes = await api.getVehicleLocation(v.id);
          setVehicleLocations((prev) => ({ ...prev, [v.id]: locRes }));
        } catch {
          // No location yet
        }
      });
    } catch (err) {
      console.error('Failed to load vehicles:', err);
    }
  }, [selectedVehicleId]);

  const loadAlertRules = useCallback(async () => {
    try {
      const res = await api.getAlerts();
      setAlertRules(res.alerts || []);
    } catch (err) {
      console.error('Failed to load alert rules:', err);
    }
  }, []);

  const refreshVehicleLocations = useCallback(async () => {
    vehicles.forEach(async (v) => {
      try {
        const locRes = await api.getVehicleLocation(v.id);
        setVehicleLocations((prev) => ({ ...prev, [v.id]: locRes }));
      } catch {
        // vehicle may not have location yet
      }
    });
  }, [vehicles]);

  useEffect(() => {
    loadGeofences();
    loadVehicles();
    loadAlertRules();
  }, [loadGeofences, loadVehicles, loadAlertRules]);

  // WebSocket connection lifecycle with auto-reconnect
  useEffect(() => {
    let reconnectTimeout = null;

    const connectWebSocket = () => {
      try {
        const ws = new WebSocket(WS_BASE);
        wsRef.current = ws;

        ws.onopen = () => {
          console.log('[WebSocket] Connected to', WS_BASE);
          setWsConnected(true);
        };

        ws.onmessage = (event) => {
          try {
            const data = JSON.parse(event.data);
            console.log('[WebSocket Alert Received]:', data);
            setLiveAlerts((prev) => [data, ...prev]);

            // Show active toast banner
            setToastAlert(data);
            setTimeout(() => {
              setToastAlert((current) => (current?.event_id === data.event_id ? null : current));
            }, 6000);

            // Automatically refresh vehicle state to show updated geofences on map
            refreshVehicleLocations();
          } catch (err) {
            console.error('Error parsing WS message:', err);
          }
        };

        ws.onclose = () => {
          console.log('[WebSocket] Disconnected. Reconnecting in 3s...');
          setWsConnected(false);
          reconnectTimeout = setTimeout(connectWebSocket, 3000);
        };

        ws.onerror = (err) => {
          console.error('[WebSocket Error]:', err);
          ws.close();
        };
      } catch (err) {
        console.error('WebSocket connection error:', err);
        reconnectTimeout = setTimeout(connectWebSocket, 3000);
      }
    };

    connectWebSocket();

    return () => {
      if (reconnectTimeout) clearTimeout(reconnectTimeout);
      if (wsRef.current) wsRef.current.close();
    };
  }, [refreshVehicleLocations]);

  // Handle map clicks
  const handleMapClick = async (lat, lng) => {
    if (mapMode === 'draw') {
      setDrawnPoints((prev) => [...prev, [lat, lng]]);
    } else if (mapMode === 'teleport' && selectedVehicleId) {
      try {
        const res = await api.updateLocation(selectedVehicleId, lat, lng);
        setMapMode('view');
        // Instantly update the vehicle location and current geofences in local state
        setVehicleLocations((prev) => ({
          ...prev,
          [selectedVehicleId]: {
            vehicle_id: selectedVehicleId,
            current_location: {
              latitude: lat,
              longitude: lng,
              timestamp: new Date().toISOString(),
            },
            current_geofences: (res.current_geofences || []).map((cg) => ({
              geofence_id: cg.geofence_id,
              geofence_name: cg.geofence_name,
              category: cg.category || '',
            })),
          },
        }));
        refreshVehicleLocations();
      } catch (err) {
        console.error('Failed to update vehicle location:', err);
      }
    }
  };

  return (
    <div className="flex flex-col h-screen w-screen bg-slate-950 text-slate-100 overflow-hidden font-sans">
      {/* Top Navigation Bar */}
      <header className="h-14 bg-slate-900 border-b border-slate-800 px-4 flex items-center justify-between shrink-0 shadow-md">
        <div className="flex items-center gap-3">
          <div className="w-9 h-9 rounded-lg bg-gradient-to-tr from-blue-600 via-indigo-600 to-purple-600 flex items-center justify-center shadow-lg shadow-blue-900/30">
            <Radio className="w-5 h-5 text-white animate-pulse" />
          </div>
          <div>
            <h1 className="text-sm font-extrabold tracking-wide text-white flex items-center gap-2">
              MapUp GeoTrack Engine
              <span className="text-[10px] bg-blue-950 border border-blue-500/40 text-blue-300 px-2 py-0.5 rounded-full uppercase tracking-wider font-semibold">
                Real-Time v2.0
              </span>
            </h1>
            <p className="text-[10px] text-slate-400">
              Autonomous Polygon Geofencing & Instant Breach Alerts
            </p>
          </div>
        </div>

        {/* Status Indicators & Navigation Tabs */}
        <div className="flex items-center gap-3">
          {/* WebSocket Live Badge */}
          <div
            className={`flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold border ${
              wsConnected
                ? 'bg-emerald-950/60 border-emerald-500/50 text-emerald-400'
                : 'bg-rose-950/60 border-rose-500/50 text-rose-400'
            }`}
          >
            <span
              className={`w-2 h-2 rounded-full ${
                wsConnected ? 'bg-emerald-400 animate-ping' : 'bg-rose-500'
              }`}
            />
            {wsConnected ? 'WebSocket Live' : 'Connecting...'}
          </div>

          {/* Quick Metrics */}
          <div className="hidden md:flex items-center gap-3 px-3 py-1 bg-slate-800/80 rounded-lg border border-slate-700/60 text-xs text-slate-300">
            <div>
              <span className="text-slate-500">Vehicles:</span>{' '}
              <strong className="text-white">{vehicles.length}</strong>
            </div>
            <div className="w-px h-3 bg-slate-700" />
            <div>
              <span className="text-slate-500">Geofences:</span>{' '}
              <strong className="text-white">{geofences.length}</strong>
            </div>
            <div className="w-px h-3 bg-slate-700" />
            <div>
              <span className="text-slate-500">Rules:</span>{' '}
              <strong className="text-white">{alertRules.length}</strong>
            </div>
          </div>

          {/* Tab Switcher */}
          <div className="flex items-center bg-slate-800 p-0.5 rounded-lg border border-slate-700 text-xs">
            <button
              onClick={() => setActiveTab('dashboard')}
              className={`px-3 py-1.5 rounded-md font-semibold transition-all flex items-center gap-1.5 ${
                activeTab === 'dashboard'
                  ? 'bg-blue-600 text-white shadow'
                  : 'text-slate-400 hover:text-white'
              }`}
            >
              <Layers className="w-3.5 h-3.5" />
              Live Map
            </button>
            <button
              onClick={() => setActiveTab('history')}
              className={`px-3 py-1.5 rounded-md font-semibold transition-all flex items-center gap-1.5 ${
                activeTab === 'history'
                  ? 'bg-blue-600 text-white shadow'
                  : 'text-slate-400 hover:text-white'
              }`}
            >
              <History className="w-3.5 h-3.5" />
              History Logs
            </button>
          </div>
        </div>
      </header>

      {/* Main Content Area */}
      <main className="flex-1 flex overflow-hidden p-3 gap-3">
        {activeTab === 'history' ? (
          <div className="w-full h-full">
            <ViolationHistory vehicles={vehicles} geofences={geofences} />
          </div>
        ) : (
          <>
            {/* Left/Center Interactive Map View + Simulation Bar */}
            <div className="flex-1 flex flex-col gap-2.5 min-w-0">
              <SimulationBar
                vehicles={vehicles}
                selectedVehicleId={selectedVehicleId}
                onSimulationTriggered={refreshVehicleLocations}
              />
              <div className="flex-1 relative min-h-0">
                <MapView
                  geofences={geofences}
                  vehicles={vehicles}
                  vehicleLocations={vehicleLocations}
                  selectedVehicleId={selectedVehicleId}
                  onSelectVehicle={(id) => {
                    setSelectedVehicleId(id);
                    setActiveSidebarTab('vehicles');
                  }}
                  mapMode={mapMode}
                  drawnPoints={drawnPoints}
                  onMapClick={handleMapClick}
                />

                {/* Floating Map Legend */}
                <div className="absolute bottom-4 left-4 z-[900] bg-slate-900/90 backdrop-blur-md border border-slate-700/80 rounded-lg p-2.5 shadow-xl text-[11px] text-slate-300">
                  <div className="font-bold text-slate-200 mb-1.5 uppercase tracking-wider text-[10px]">
                    Zone Legend
                  </div>
                  <div className="space-y-1">
                    <div className="flex items-center gap-2">
                      <span className="w-3 h-3 rounded bg-rose-500 border border-rose-300" />
                      <span>Restricted Zone (Instant Breach Alert)</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <span className="w-3 h-3 rounded bg-blue-500 border border-blue-300" />
                      <span>Delivery Zone</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <span className="w-3 h-3 rounded bg-amber-500 border border-amber-300" />
                      <span>Toll Zone</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <span className="w-3 h-3 rounded bg-emerald-500 border border-emerald-300" />
                      <span>Customer Area</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            {/* Right Control Panels Drawer */}
            <div className="w-80 md:w-96 flex flex-col gap-2 shrink-0">
              {/* Sidebar Tabs */}
              <div className="grid grid-cols-4 bg-slate-900 border border-slate-800 p-1 rounded-xl text-xs font-semibold text-slate-400">
                <button
                  onClick={() => setActiveSidebarTab('alerts')}
                  className={`py-1.5 rounded-lg transition-all flex items-center justify-center gap-1 ${
                    activeSidebarTab === 'alerts'
                      ? 'bg-slate-800 text-amber-400 shadow'
                      : 'hover:text-slate-200'
                  }`}
                  title="Real-time Alerts Feed"
                >
                  <Bell className="w-3.5 h-3.5" />
                  Alerts
                </button>

                <button
                  onClick={() => setActiveSidebarTab('geofences')}
                  className={`py-1.5 rounded-lg transition-all flex items-center justify-center gap-1 ${
                    activeSidebarTab === 'geofences'
                      ? 'bg-slate-800 text-blue-400 shadow'
                      : 'hover:text-slate-200'
                  }`}
                  title="Geofence Zones"
                >
                  <MapPin className="w-3.5 h-3.5" />
                  Zones
                </button>

                <button
                  onClick={() => setActiveSidebarTab('vehicles')}
                  className={`py-1.5 rounded-lg transition-all flex items-center justify-center gap-1 ${
                    activeSidebarTab === 'vehicles'
                      ? 'bg-slate-800 text-purple-400 shadow'
                      : 'hover:text-slate-200'
                  }`}
                  title="Fleet Vehicles"
                >
                  <Truck className="w-3.5 h-3.5" />
                  Fleet
                </button>

                <button
                  onClick={() => setActiveSidebarTab('rules')}
                  className={`py-1.5 rounded-lg transition-all flex items-center justify-center gap-1 ${
                    activeSidebarTab === 'rules'
                      ? 'bg-slate-800 text-amber-500 shadow'
                      : 'hover:text-slate-200'
                  }`}
                  title="Alert Policies"
                >
                  <Shield className="w-3.5 h-3.5" />
                  Rules
                </button>
              </div>

              {/* Active Sidebar Component */}
              <div className="flex-1 min-h-0">
                {activeSidebarTab === 'alerts' && (
                  <AlertFeed
                    alerts={liveAlerts}
                    onClear={() => setLiveAlerts([])}
                    isConnected={wsConnected}
                  />
                )}
                {activeSidebarTab === 'geofences' && (
                  <GeofenceManager
                    geofences={geofences}
                    onGeofenceCreated={() => {
                      loadGeofences();
                      refreshVehicleLocations();
                    }}
                    mapMode={mapMode}
                    setMapMode={setMapMode}
                    drawnPoints={drawnPoints}
                    setDrawnPoints={setDrawnPoints}
                  />
                )}
                {activeSidebarTab === 'vehicles' && (
                  <VehicleManager
                    vehicles={vehicles}
                    vehicleLocations={vehicleLocations}
                    selectedVehicleId={selectedVehicleId}
                    onSelectVehicle={setSelectedVehicleId}
                    onVehicleRegistered={loadVehicles}
                    mapMode={mapMode}
                    setMapMode={setMapMode}
                    onLocationUpdated={refreshVehicleLocations}
                  />
                )}
                {activeSidebarTab === 'rules' && (
                  <AlertRuleManager
                    alertRules={alertRules}
                    geofences={geofences}
                    vehicles={vehicles}
                    onRuleCreated={loadAlertRules}
                  />
                )}
              </div>
            </div>
          </>
        )}
      </main>

      {/* Floating Real-Time Toast Alert Banner */}
      {toastAlert && (
        <div className="fixed top-16 right-6 z-[3000] alert-pop max-w-sm w-full bg-slate-900 border-2 border-rose-500 rounded-xl p-4 shadow-2xl text-slate-100 flex items-start gap-3">
          <div className="w-8 h-8 rounded-full bg-rose-500/20 text-rose-400 flex items-center justify-center shrink-0">
            <Radio className="w-5 h-5 animate-pulse" />
          </div>
          <div className="flex-1 text-xs">
            <div className="flex items-center justify-between mb-1">
              <span className="font-extrabold uppercase tracking-wider text-rose-400">
                {toastAlert.event_type} Alert: {toastAlert.geofence?.geofence_name}
              </span>
              <button
                onClick={() => setToastAlert(null)}
                className="text-slate-400 hover:text-white"
              >
                ✕
              </button>
            </div>
            <p className="text-slate-300">
              Vehicle <strong className="text-white">{toastAlert.vehicle?.vehicle_number}</strong> (
              {toastAlert.vehicle?.driver_name}) has {toastAlert.event_type}ed{' '}
              {toastAlert.geofence?.category?.replace('_', ' ')}.
            </p>
          </div>
        </div>
      )}
    </div>
  );
}
