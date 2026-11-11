const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080';
const WS_BASE = import.meta.env.VITE_WS_URL || (API_BASE.replace(/^http/, 'ws') + '/ws/alerts');

export { API_BASE, WS_BASE };

async function handleResponse(res) {
  const json = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error(json.error || `HTTP error ${res.status}`);
  }
  return json;
}

export const api = {
  // Geofences
  async getGeofences(category = '') {
    const url = new URL(`${API_BASE}/geofences`);
    if (category) url.searchParams.set('category', category);
    const res = await fetch(url);
    return handleResponse(res);
  },

  async createGeofence(data) {
    const res = await fetch(`${API_BASE}/geofences`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });
    return handleResponse(res);
  },

  // Vehicles
  async getVehicles() {
    const res = await fetch(`${API_BASE}/vehicles`);
    return handleResponse(res);
  },

  async createVehicle(data) {
    const res = await fetch(`${API_BASE}/vehicles`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });
    return handleResponse(res);
  },

  // Location
  async updateLocation(vehicleId, latitude, longitude, timestamp) {
    const res = await fetch(`${API_BASE}/vehicles/location`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        vehicle_id: vehicleId,
        latitude: parseFloat(latitude),
        longitude: parseFloat(longitude),
        timestamp: timestamp || new Date().toISOString(),
      }),
    });
    return handleResponse(res);
  },

  async getVehicleLocation(vehicleId) {
    const res = await fetch(`${API_BASE}/vehicles/location/${vehicleId}`);
    return handleResponse(res);
  },

  // Alerts
  async configureAlert(data) {
    const res = await fetch(`${API_BASE}/alerts/configure`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });
    return handleResponse(res);
  },

  async getAlerts(geofenceId = '', vehicleId = '') {
    const url = new URL(`${API_BASE}/alerts`);
    if (geofenceId) url.searchParams.set('geofence_id', geofenceId);
    if (vehicleId) url.searchParams.set('vehicle_id', vehicleId);
    const res = await fetch(url);
    return handleResponse(res);
  },

  // Violations History
  async getViolations(params = {}) {
    const url = new URL(`${API_BASE}/violations/history`);
    if (params.vehicle_id) url.searchParams.set('vehicle_id', params.vehicle_id);
    if (params.geofence_id) url.searchParams.set('geofence_id', params.geofence_id);
    if (params.start_date) url.searchParams.set('start_date', params.start_date);
    if (params.end_date) url.searchParams.set('end_date', params.end_date);
    if (params.limit) url.searchParams.set('limit', params.limit);
    const res = await fetch(url);
    return handleResponse(res);
  },
};
