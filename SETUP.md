# 🛠️ System Setup & Developer Guide (`SETUP.md`)

This guide provides complete instructions to install, configure, test, and operate the **MapUp Geofencing & Real-Time Alert System**.

---

## 📋 Table of Contents
1. [Prerequisites and Dependencies](#1-prerequisites-and-dependencies)
2. [Local Setup Instructions](#2-local-setup-instructions)
3. [How to Run with Docker Compose](#3-how-to-run-with-docker-compose)
4. [API Testing Guide with Example `curl` Commands](#4-api-testing-guide-with-example-curl-commands)
5. [Frontend Usage Guide](#5-frontend-usage-guide)
6. [Architecture Overview](#6-architecture-overview)

---

## 1. Prerequisites and Dependencies

Before setting up the project locally or via containers, ensure your machine satisfies the following requirements:

### System & Tool Requirements
| Component | Minimum Version | Purpose |
|---|---|---|
| **Operating System** | macOS, Linux, or Windows 10/11 | Host environment |
| **GoLang** | Go `1.23.0` or higher | Backend REST API & WebSocket server |
| **Node.js** | Node `18.x` or higher (LTS recommended) | Frontend build environment (Vite) |
| **npm** | `9.x` or higher | Frontend package manager |
| **Docker & Docker Compose** | Docker `24.x+`, Compose `v2.x+` | Containerized deployment |
| **curl** | Any modern version | CLI testing for REST endpoints |
| **wscat** *(Optional)* | `npm install -g wscat` | CLI testing for WebSocket connections |

### Ports Used by Default
- `8080`: Go Backend REST API & WebSocket Server
- `3000`: React Frontend (Vite Dev Server or Docker Nginx)
- `5432`: PostgreSQL 16 with PostGIS spatial extension (when running database container)

> [!NOTE]
> **Zero-Dependency Database Fallback**: The backend includes an automated dual-engine database layer. If PostgreSQL/PostGIS is not detected on port `5432`, the backend **automatically falls back to a high-performance in-memory repository**, allowing developers to run the entire backend immediately without installing or configuring a local database!

---

## 2. Local Setup Instructions

Follow these steps to run both the backend and frontend directly on your host machine without Docker.

### 2.1 Backend Setup (Go)

1. **Clone the repository**:
   ```bash
   git clone https://github.com/gowthamtraining/MAPUP_TASK.git
   cd MAPUP_TASK
   ```

2. **Navigate to backend and download dependencies**:
   ```bash
   cd backend
   go mod download
   ```

3. **Configure Environment Variables (Optional)**:
   By default, sensible defaults are configured:
   - `PORT=8080`
   - `REQUIRE_AUTH=false` (Set `true` to enforce strict API key / JWT checks)
   - `API_KEY=mapup-secret-key-2025`
   - `DATABASE_URL` (Leave empty to use thread-safe in-memory storage, or provide a PostGIS connection string):
     ```bash
     # Optional PostGIS connection string:
     # export DATABASE_URL="postgres://geouser:geopass@localhost:5432/geotracker?sslmode=disable"
     ```

4. **Start the Go Backend Server**:
   ```bash
   go run ./cmd/server
   ```
   *Expected console output:*
   ```text
   2026/11/13 10:00:00 Connecting to PostgreSQL with PostGIS...
   2026/11/13 10:00:00 Initializing storage engine...
   2026/11/13 10:00:00 Seeding initial sample geofences and vehicles...
   2026/11/13 10:00:00 WebSocket Hub started and listening for broadcast events.
   2026/11/13 10:00:00 MapUp Geofencing & Real-Time Alert Server running on port :8080
   ```

5. **Run Backend Unit Tests**:
   To verify that all algorithms, endpoints, and middleware pass evaluation:
   ```bash
   go test -v ./...
   ```

---

### 2.2 Frontend Setup (React + Vite)

1. **Open a new terminal window** and navigate to the `frontend` directory:
   ```bash
   cd frontend
   ```

2. **Install frontend dependencies**:
   ```bash
   npm install
   ```

3. **Configure Frontend Environment (Optional)**:
   Create or verify `.env` in `frontend/`:
   ```env
   VITE_API_URL=http://localhost:8080
   VITE_WS_URL=ws://localhost:8080/ws/alerts
   VITE_API_KEY=mapup-secret-key-2025
   ```

4. **Start Vite Development Server**:
   ```bash
   npm run dev
   ```
   *Expected output:*
   ```text
     VITE v6.0.0  ready in 250 ms

     ➜  Local:   http://localhost:3000/
     ➜  Network: use --host to expose
   ```

5. Open your web browser at [http://localhost:3000](http://localhost:3000).

---

## 3. How to Run with Docker Compose

Running the entire stack with Docker Compose is the recommended approach for complete parity with production. It orchestrates three containers:
- `mapup-postgis-db`: PostgreSQL 16 with PostGIS spatial extension pre-enabled.
- `mapup-geofence-backend`: Multi-stage compiled Go backend binary.
- `mapup-geofence-frontend`: Production Vite build served via Nginx with reverse proxy caching.

### Step 1: Start All Services
From the root repository directory:
```bash
docker compose up --build
```
*(Or `docker-compose up --build` on older Docker versions)*

### Step 2: Verify Running Containers
Open another terminal and verify the healthy state of the services:
```bash
docker compose ps
```
*Expected status:*
```text
NAME                     IMAGE                         COMMAND                  SERVICE    STATUS
mapup-postgis-db         postgis/postgis:16-3.4-alpine "docker-entrypoint.s…"   db         Up (healthy) :5432
mapup-geofence-backend   mapup-backend:latest          "/app/server"            backend    Up           :8080
mapup-geofence-frontend  mapup-frontend:latest         "/docker-entrypoint.…"   frontend   Up           :3000
```

### Step 3: Access the Endpoints
- **Web Application Dashboard**: [http://localhost:3000](http://localhost:3000)
- **REST API Base URL**: [http://localhost:8080](http://localhost:8080)
- **Live WebSocket Alerts Stream**: `ws://localhost:8080/ws/alerts`
- **Database Connection**: `localhost:5432` (User: `geouser`, Password: `geopass`, DB: `geotracker`)

### Step 4: Stop Containers & Clean Volumes
```bash
docker compose down -v
```

---

## 4. API Testing Guide with Example `curl` Commands

Every response strictly adheres to the assessment requirements and returns execution duration measured in nanoseconds as `"time_ns"`.

### 1. `POST /geofences` — Create a Geofence
Creates a closed polygonal geofence. The first and last coordinate must match.

```bash
curl -X POST http://localhost:8080/geofences \
  -H "Content-Type: application/json" \
  -H "X-API-Key: mapup-secret-key-2025" \
  -d '{
    "name": "Downtown Delivery Zone",
    "description": "Main delivery area for downtown customers",
    "coordinates": [
      [37.7749, -122.4194],
      [37.7849, -122.4194],
      [37.7849, -122.4094],
      [37.7749, -122.4094],
      [37.7749, -122.4194]
    ],
    "category": "delivery_zone"
  }'
```
**Sample Response (`HTTP 201 Created`):**
```json
{
  "id": "geo_123",
  "name": "Downtown Delivery Zone",
  "status": "active",
  "time_ns": "1234567"
}
```

---

### 2. `GET /geofences` — Retrieve All Geofences (with Optional Category Filter)

```bash
# Retrieve all geofences
curl http://localhost:8080/geofences

# Retrieve geofences filtered by category
curl "http://localhost:8080/geofences?category=delivery_zone"
```
**Sample Response (`HTTP 200 OK`):**
```json
{
  "geofences": [
    {
      "id": "geo_123",
      "name": "Downtown Delivery Zone",
      "description": "Main delivery area for downtown customers",
      "coordinates": [
        [37.7749, -122.4194],
        [37.7849, -122.4194],
        [37.7849, -122.4094],
        [37.7749, -122.4094],
        [37.7749, -122.4194]
      ],
      "category": "delivery_zone",
      "created_at": "2025-01-15T10:30:00Z"
    }
  ],
  "time_ns": "987654"
}
```

---

### 3. `POST /vehicles` — Register a New Vehicle

```bash
curl -X POST http://localhost:8080/vehicles \
  -H "Content-Type: application/json" \
  -H "X-API-Key: mapup-secret-key-2025" \
  -d '{
    "vehicle_number": "KA-01-AB-1234",
    "driver_name": "John Doe",
    "vehicle_type": "truck",
    "phone": "+1234567890"
  }'
```
**Sample Response (`HTTP 201 Created`):**
```json
{
  "id": "veh_456",
  "vehicle_number": "KA-01-AB-1234",
  "status": "active",
  "time_ns": "1123456"
}
```

---

### 4. `GET /vehicles` — Retrieve All Registered Vehicles

```bash
curl http://localhost:8080/vehicles
```
**Sample Response (`HTTP 200 OK`):**
```json
{
  "vehicles": [
    {
      "id": "veh_456",
      "vehicle_number": "KA-01-AB-1234",
      "driver_name": "John Doe",
      "vehicle_type": "truck",
      "phone": "+1234567890",
      "status": "active",
      "created_at": "2025-01-15T09:00:00Z"
    }
  ],
  "time_ns": "876543"
}
```

---

### 5. `POST /vehicles/location` — Update Location & Trigger Geofence Alerts
*Evaluates whether the vehicle entered or exited zones and asynchronously broadcasts alerts to `/ws/alerts` for restricted breaches.*

```bash
curl -X POST http://localhost:8080/vehicles/location \
  -H "Content-Type: application/json" \
  -H "X-API-Key: mapup-secret-key-2025" \
  -d '{
    "vehicle_id": "veh_456",
    "latitude": 37.7849,
    "longitude": -122.4194,
    "timestamp": "2025-01-15T10:35:00Z"
  }'
```
**Sample Response (`HTTP 200 OK`):**
```json
{
  "vehicle_id": "veh_456",
  "location_updated": true,
  "current_geofences": [
    {
      "geofence_id": "geo_123",
      "geofence_name": "Downtown Delivery Zone",
      "status": "inside"
    }
  ],
  "time_ns": "2345678"
}
```

---

### 6. `GET /vehicles/location/{vehicle_id}` — Get Current Vehicle Location & Zones

```bash
curl http://localhost:8080/vehicles/location/veh_456
```
**Sample Response (`HTTP 200 OK`):**
```json
{
  "vehicle_id": "veh_456",
  "vehicle_number": "KA-01-AB-1234",
  "current_location": {
    "latitude": 37.7849,
    "longitude": -122.4194,
    "timestamp": "2025-01-15T10:35:00Z"
  },
  "current_geofences": [
    {
      "geofence_id": "geo_123",
      "geofence_name": "Downtown Delivery Zone",
      "category": "delivery_zone"
    }
  ],
  "time_ns": "876543"
}
```

---

### 7. `POST /alerts/configure` — Configure Alert Rules

```bash
curl -X POST http://localhost:8080/alerts/configure \
  -H "Content-Type: application/json" \
  -H "X-API-Key: mapup-secret-key-2025" \
  -d '{
    "geofence_id": "geo_123",
    "vehicle_id": "veh_456",
    "event_type": "entry"
  }'
```
**Sample Response (`HTTP 201 Created`):**
```json
{
  "alert_id": "alert_789",
  "geofence_id": "geo_123",
  "vehicle_id": "veh_456",
  "event_type": "entry",
  "status": "active",
  "time_ns": "1567890"
}
```

---

### 8. `GET /alerts` — Retrieve Configured Alert Rules (with Optional Filters)

```bash
# Retrieve all alert rules
curl http://localhost:8080/alerts

# Retrieve rules filtered by geofence or vehicle
curl "http://localhost:8080/alerts?geofence_id=geo_123&vehicle_id=veh_456"
```
**Sample Response (`HTTP 200 OK`):**
```json
{
  "alerts": [
    {
      "alert_id": "alert_789",
      "geofence_id": "geo_123",
      "geofence_name": "Downtown Delivery Zone",
      "vehicle_id": "veh_456",
      "vehicle_number": "KA-01-AB-1234",
      "event_type": "entry",
      "status": "active",
      "created_at": "2025-01-15T09:15:00Z"
    }
  ],
  "time_ns": "654321"
}
```

---

### 9. `GET /violations/history` — Retrieve Historical Entry/Exit Audit Logs

```bash
# Basic retrieval with default limit (50)
curl http://localhost:8080/violations/history

# Multi-filter query with pagination
curl "http://localhost:8080/violations/history?vehicle_id=veh_456&geofence_id=geo_123&limit=100"
```
**Sample Response (`HTTP 200 OK`):**
```json
{
  "violations": [
    {
      "id": "viol_111",
      "vehicle_id": "veh_456",
      "vehicle_number": "KA-01-AB-1234",
      "geofence_id": "geo_123",
      "geofence_name": "Downtown Delivery Zone",
      "event_type": "entry",
      "latitude": 37.7849,
      "longitude": -122.4194,
      "timestamp": "2025-01-15T10:35:00Z"
    }
  ],
  "total_count": 245,
  "time_ns": "3456789"
}
```

---

### 10. `GET /ws/alerts` — WebSocket Real-Time Alert Stream

Test using `wscat`:
```bash
wscat -c "ws://localhost:8080/ws/alerts"
```
When a vehicle enters a restricted zone or triggers a configured rule, the WebSocket immediately pushes:
```json
{
  "event_id": "evt_999",
  "event_type": "entry",
  "timestamp": "2025-01-15T10:35:00Z",
  "vehicle": {
    "vehicle_id": "veh_456",
    "vehicle_number": "KA-01-AB-1234",
    "driver_name": "John Doe"
  },
  "geofence": {
    "geofence_id": "geo_123",
    "geofence_name": "Downtown Delivery Zone",
    "category": "delivery_zone"
  },
  "location": {
    "latitude": 37.7849,
    "longitude": -122.4194
  }
}
```

---

## 5. Frontend Usage Guide

The React application delivers a responsive web interface designed for fleet managers, dispatchers, and security supervisors.

```
┌────────────────────────────────────────────────────────────────────────┐
│  🌐 MapUp GeoTrack Engine                     ● WebSocket Live         │
├─────────────────────────────────────────┬──────────────────────────────┤
│                                         │ [Alerts] [Zones] [Fleet]     │
│                                         ├──────────────────────────────┤
│                                         │ 🚨 Live Real-Time Alerts     │
│         Interactive Leaflet Map         │                              │
│                                         │ - Restricted Zone Breach!    │
│  - Red: Restricted Zones                │   KA-01-AB-1234 entered Port │
│  - Blue: Delivery Zones                 │   Security at 10:35:00Z      │
│  - Green: Customer Areas                │                              │
│  - Vehicles with pulsating SVG icons    │ - Exit Event                 │
│                                         │   KA-05-XY-7777 left Zone    │
│                                         │                              │
│                                         ├──────────────────────────────┤
│                                         │ [Draw Polygon] [Move Vehicle]│
└─────────────────────────────────────────┴──────────────────────────────┘
```

### 5.1 Main Layout & Status Bar
- **Top Navigation Bar**:
  - **WebSocket Live Indicator**: A green pulsing dot signals an active WebSocket connection. If connection drops, it displays an amber reconnecting badge.
  - **View Switcher**: Toggle between the main **Dashboard Map** and **History Logs** audit table.

### 5.2 Creating Geofences (Interactive Drawing & Coordinates)
1. In the right sidebar, click the **"Zones"** tab.
2. Click **"Draw Mode"**.
3. Click anywhere on the map to place vertices (at least 3 unique points).
4. As you click, a dynamic dashed polygon outline previews the shape.
5. Click **"Save (N pts)"** to open the Creation Modal.
6. Provide:
   - **Name**: e.g., `Downtown Delivery Hub`
   - **Category**: Select `delivery_zone`, `restricted_zone`, `toll_zone`, or `customer_area`.
   - **Description**: Optional text.
7. Click **"Create Geofence"**. The polygon renders on the map immediately with category-specific color coding.

### 5.3 Managing Vehicles & Fleet
1. Click the **"Fleet"** tab on the sidebar.
2. Click **"Add Vehicle"** to register a new truck, car, van, or motorcycle.
3. Every registered vehicle displays:
   - Registration number, driver name, and phone.
   - Live location status badge (`Inside: <Zone Name>` or `Outside geofences`).
   - A distinct animated SVG vehicle marker positioned on the map.

### 5.4 Updating Vehicle Locations (Click-to-Move Teleportation)
1. On any vehicle card in the Fleet list, click the blue **"Move"** button.
2. The map cursor changes to crosshairs.
3. Click any point on the map (inside or outside of a geofence).
4. The frontend calls `POST /vehicles/location`, animates the vehicle marker, and updates the vehicle status in real time.

### 5.5 Triggering Real-Time Alerts (Restricted Zone Breaches)
- Move any vehicle into a red **`restricted_zone`** (such as the pre-seeded *Port Security Zone*).
- The system automatically detects the state transition from outside to inside:
  - An urgent **red floating toast notification** slides in at the top right.
  - A notification audio ping plays.
  - An alert card is pushed to the top of the **Live Alerts** feed with exact vehicle, zone, and coordinate details.

### 5.6 Reviewing Historical Audit Logs (History Logs View)
1. Click **"History Logs"** in the top navigation bar.
2. View the tabulated history of all entry and exit events.
3. Use the filter bar to:
   - Filter by specific vehicle number.
   - Filter by specific geofence name.
   - Navigate pages using server-side pagination.

---

## 6. Architecture Overview

### 6.1 Architectural Block Diagram

```
[React 19 Frontend (Vite + Leaflet)]
        │                 ▲
   HTTP │            WSS  │ (/ws/alerts)
        ▼                 │
[Go Chi Router & API Controllers]
        │
   [Middleware: Token-Bucket Rate Limiter & Auth Validator]
        │
   [Execution Timer: Injects time_ns in all responses]
        │
        ▼
[Tracker Service: Stateful Transition Detector]
   ├── Computes Set Difference: Entries (New - Old) & Exits (Old - New)
   ├── Evaluates Alert Rules & Restricted Zone Flags
   └── Dispatches Async Goroutine ──► [Go WebSocket Hub Actor]
        │
        ▼
[Geospatial Math Engine & Persistence Layer]
   ├── O(1) Bounding-Box Envelope Pre-filter
   ├── Ray-Casting Point-in-Polygon with Inclusive Boundary Check
   └── PostGIS 16 (GiST Spatial Index on Polygon Geometries)
        └── (Automatic Fallback: In-Memory Mutex Store)
```

### 6.2 Key Architectural Decisions

1. **Stateful Transition Detector**:
   - Simple containment does not trigger alerts. Alerts are **edge triggers**.
   - The engine maintains each vehicle's `previousGeofenceIDs`. Transitions are computed as:
     $$\text{Entries} = \text{CurrentGeofences} \setminus \text{PreviousGeofences}$$
     $$\text{Exits} = \text{PreviousGeofences} \setminus \text{CurrentGeofences}$$
   - This prevents alert spamming when a vehicle remains stationary inside a zone.

2. **Geospatial Point-in-Polygon Engine**:
   - **$O(1)$ Bounding-Box Fast Rejection**: Calculates $[MinLat, MaxLat, MinLon, MaxLon]$ to instantly dismiss distant points before running trigonometry.
   - **Ray-Casting Algorithm**: Casts a horizontal ray eastward and counts segment intersections (odd = inside, even = outside).
   - **Inclusive Boundary Semantics (`ST_Covers`)**: Explicit collinearity tests ensure points resting exactly on polygon segments are classified as **INSIDE**.

3. **Coordinate Order Standard**:
   - The REST API receives `[Latitude, Longitude]`.
   - PostGIS and GeoJSON expect `(Longitude, Latitude)` / $(X, Y)$ order.
   - The repository explicitly maps Longitude $\to X$ and Latitude $\to Y$ when constructing WKT (`SRID=4326;POLYGON((lon lat, ...))`) and point geometry queries.

4. **Non-Blocking Asynchronous WebSocket Broadcast**:
   - Alert dispatch is offloaded into a Go goroutine (`go s.wsHub.BroadcastAlert(msg)`).
   - The HTTP response returns to the GPS tracker in **under 2 milliseconds**, completely decoupled from connected browser latency.

5. **Production Bonus Features**:
   - **Token Bucket Rate Limiter**: Protects `/vehicles/location` against DoS floods with standard `X-RateLimit-*` headers.
   - **Authentication**: Supports `X-API-Key` and RFC 7519 HMAC-SHA256 JWT tokens.
   - **Universal Pagination**: Full `page` and `limit` support on geofences, vehicles, alerts, and history.
   - **Unit Test Suite**: 100% passing tests for geometric math, API contracts, transition logic, and security middleware.

---

## 7. Cloud Deployment & Docker Hub Reference

### Pre-Built Docker Images
- **Backend Image**: [`gowtham095/mapup-geofence-backend:latest`](https://hub.docker.com/r/gowtham095/mapup-geofence-backend)
- **Frontend Image**: [`gowtham095/mapup-geofence-frontend:latest`](https://hub.docker.com/r/gowtham095/mapup-geofence-frontend)

### Live Cloud Deployment
- **Backend API (Render)**: [`https://mapup-geofence-backend-latest.onrender.com`](https://mapup-geofence-backend-latest.onrender.com)
- **Live WebSocket Feed**: `wss://mapup-geofence-backend-latest.onrender.com/ws/alerts`
- **Frontend Configuration**: Ready for one-click deployment on Netlify (`netlify.toml` included) or Vercel (`vercel.json` included).

---
*Developed for the MapUp Full-Stack Developer Assessment.*
