const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080';
const API_KEY = import.meta.env.VITE_API_KEY || 'mapup-secret-key-2025';
const WS_BASE = import.meta.env.VITE_WS_URL || (API_BASE.replace(/^http/, 'ws') + '/ws/alerts?api_key=' + API_KEY);

export { API_BASE, WS_BASE, API_KEY };

const authHeaders = {
  'Content-Type': 'application/json',
  'X-API-Key': API_KEY,
};

async function handleResponse(res) {
  const json = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error(json.error || `HTTP error ${res.status}`);
  }
  return json;
}

export const api = {
  // Authentication
  async getAuthToken(username = 'admin', password = 'password') {
    const res = await fetch(`${API_BASE}/auth/token`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password }),
    });
    return handleResponse(res);
  },

  async verifyAuth() {
    const res = await fetch(`${API_BASE}/auth/verify`, {
      headers: authHeaders,
    });
    return handleResponse(res);
  },

  // Geofences with pagination support
  async getGeofences(category = '', page = 0, limit = 0) {
    const url = new URL(`${API_BASE}/geofences`);
    if (category) url.searchParams.set('category', category);
    if (page > 0) url.searchParams.set('page', page);
    if (limit > 0) url.searchParams.set('limit', limit);
    const res = await fetch(url, { headers: authHeaders });
    return handleResponse(res);
  },

  async createGeofence(data) {
    const res = await fetch(`${API_BASE}/geofences`, {
      method: 'POST',
      headers: authHeaders,
      body: JSON.stringify(data),
    });
    return handleResponse(res);
  },

  // Vehicles with pagination support
  async getVehicles(page = 0, limit = 0) {
    const url = new URL(`${API_BASE}/vehicles`);
    if (page > 0) url.searchParams.set('page', page);
    if (limit > 0) url.searchParams.set('limit', limit);
    const res = await fetch(url, { headers: authHeaders });
    return handleResponse(res);
  },

  async createVehicle(data) {
    const res = await fetch(`${API_BASE}/vehicles`, {
      method: 'POST',
      headers: authHeaders,
      body: JSON.stringify(data),
    });
    return handleResponse(res);
  },

  // Location Updates (Rate Limited)
  async updateLocation(vehicleId, latitude, longitude, timestamp) {
    const res = await fetch(`${API_BASE}/vehicles/location`, {
      method: 'POST',
      headers: {
        ...authHeaders,
        'X-Vehicle-ID': vehicleId,
      },
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
    const res = await fetch(`${API_BASE}/vehicles/location/${vehicleId}`, {
      headers: authHeaders,
    });
    return handleResponse(res);
  },

  // Alerts with pagination support
  async configureAlert(data) {
    const res = await fetch(`${API_BASE}/alerts/configure`, {
      method: 'POST',
      headers: authHeaders,
      body: JSON.stringify(data),
    });
    return handleResponse(res);
  },

  async getAlerts(geofenceId = '', vehicleId = '', page = 0, limit = 0) {
    const url = new URL(`${API_BASE}/alerts`);
    if (geofenceId) url.searchParams.set('geofence_id', geofenceId);
    if (vehicleId) url.searchParams.set('vehicle_id', vehicleId);
    if (page > 0) url.searchParams.set('page', page);
    if (limit > 0) url.searchParams.set('limit', limit);
    const res = await fetch(url, { headers: authHeaders });
    return handleResponse(res);
  },

  // Violations History with pagination support
  async getViolations(params = {}) {
    const url = new URL(`${API_BASE}/violations/history`);
    if (params.vehicle_id) url.searchParams.set('vehicle_id', params.vehicle_id);
    if (params.geofence_id) url.searchParams.set('geofence_id', params.geofence_id);
    if (params.start_date) url.searchParams.set('start_date', params.start_date);
    if (params.end_date) url.searchParams.set('end_date', params.end_date);
    if (params.limit) url.searchParams.set('limit', params.limit);
    if (params.page) url.searchParams.set('page', params.page);
    const res = await fetch(url, { headers: authHeaders });
    return handleResponse(res);
  },
};
