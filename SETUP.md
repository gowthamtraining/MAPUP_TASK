# 🧭 MapUp Geofencing & Real-Time Alert System - Technical Guide

A production-grade, full-stack geospatial tracking and notification platform built for MapUp's Full-Stack Developer Assessment.

---

## 🏗️ System Architecture & Design Choices

```
                        ┌───────────────────────────────────────────────┐
                        │      React 19 Frontend Dashboard (Vite)       │
                        │   - Interactive Leaflet Polygon Visualizer    │
                        │   - Live Real-Time WebSocket Alerts Feed      │
                        │   - Click-to-Move & Polygon Drawing Tools     │
                        └───────────────────────┬───────────────────────┘
                                                │
                          REST (JSON) + time_ns │ WebSocket (/ws/alerts)
                                                ▼
                        ┌───────────────────────────────────────────────┐
                        │              GoLang Backend API               │
                        │   - Chi Router + High-Precision Nano Timers   │
                        │   - Stateful Transition Engine (Entry/Exit)   │
                        │   - Non-Blocking Asynchronous WS Hub          │
                        │   - Geospatial Ray-Casting & Validation       │
                        └───────────────────────┬───────────────────────┘
                                                │
                                    SQL / Spatial Queries
                                                ▼
                        ┌───────────────────────────────────────────────┐
                        │          PostgreSQL + PostGIS 16              │
                        │   - ST_Covers / ST_Contains Spatial Filtering │
                        │   - GiST Spatial Index on Polygon Geometries  │
                        │   - Graceful Fallback In-Memory Engine        │
                        └───────────────────────────────────────────────┘
```

### 1. Geospatial Logic & Point-in-Polygon (PIP)
- **Ray-Casting Algorithm ($O(N)$ with $O(1)$ Bounding Box Pre-filter)**:
  - First tests point coordinates against an $O(1)$ pre-computed axis-aligned bounding box (`MinLat`, `MaxLat`, `MinLon`, `MaxLon`). Points outside the bounding box are instantly rejected with zero trigonometric/intersection operations.
  - For points within the bounding box, a horizontal ray is cast eastward from the test point $(lat, lon)$.
- **Boundary Handling (`ST_Covers` Semantics)**:
  - Explicit collinearity and cross-product tests verify whether the point lies exactly on an edge segment.
  - Points on the perimeter boundary are considered **INSIDE** (following PostGIS `ST_Covers` semantics), avoiding dropped alerts when vehicles travel along boundary perimeters.
- **Coordinate Order Standard**:
  - The API specification specifies coordinates as `[latitude, longitude]`.
  - When storing and querying geometries in PostGIS (which expects `[longitude, latitude]` / X, Y order), the repository correctly constructs WKT as `SRID=4326;POLYGON((lon lat, ...))` and `ST_SetSRID(ST_MakePoint(lon, lat), 4326)`.

### 2. State-Based Entry & Exit Detection
- Entry and exit are **transitions**, not static containment states.
- The system stores each vehicle's set of previously occupied geofence IDs (`previousGeofenceSet`).
- On every location update:
  1. Computes the `currentGeofenceSet` of geofences containing the new location.
  2. **Entry events**: Geofences present in `currentGeofenceSet` that were not in `previousGeofenceSet`.
  3. **Exit events**: Geofences present in `previousGeofenceSet` that are no longer in `currentGeofenceSet`.
  4. Updates the stored state in the database.
- Any entry into a `restricted_zone` automatically triggers an alert.
- Other events match against configured alert rules (`entry`, `exit`, or `both`, for specific vehicles or all fleet vehicles).

### 3. Non-Blocking Real-Time Alert Delivery
- HTTP response latency is protected: alert generation and WebSocket broadcasting are executed asynchronously via Go channels and goroutines.
- The `websocket.Hub` utilizes a channel-driven event loop (`register`, `unregister`, `broadcast`) to prevent race conditions without lock contention.
- Client write pumps use heartbeats (ping/pong) and write deadlines to gracefully drop dead connections and reclaim resources.

### 4. Nanosecond Precision (`time_ns`)
- Every HTTP response (including 400 and 500 error responses) measures execution duration using `time.Now()` at the beginning and end of each request handler, injecting `"time_ns"` as a string at the root of the JSON payload.

---

## 🚀 Quick Start with Docker Compose

Ensure Docker and Docker Compose are installed.

```bash
# Clone the repository
git clone <your-repo-url>
cd mapup

# Launch PostGIS database, Go backend, and React frontend
docker-compose up --build
```

Services will be available at:
- **Frontend Dashboard**: [http://localhost:3000](http://localhost:3000)
- **Backend API**: [http://localhost:8080](http://localhost:8080)
- **WebSocket Feed**: `ws://localhost:8080/ws/alerts`
- **PostGIS Database**: `localhost:5432` (user: `geouser`, pass: `geopass`, db: `geotracker`)

To stop the containers:
```bash
docker-compose down
```

---

## 💻 Local Setup (Without Docker)

### Prerequisites
- **GoLang 1.23+**
- **Node.js 18+ & npm**
- Optional: PostgreSQL with PostGIS extension (if not running, backend automatically uses high-performance thread-safe In-Memory storage).

### 1. Run Backend
```bash
cd backend

# Build and start server (uses in-memory repository by default)
go run ./cmd/server

# Or with PostgreSQL/PostGIS:
# export DATABASE_URL="postgres://geouser:geopass@localhost:5432/geotracker?sslmode=disable"
# go run ./cmd/server
```
The backend starts on `http://localhost:8080`.

### 2. Run Frontend
```bash
cd frontend

# Install dependencies
npm install

# Start Vite development server
npm run dev
```
The frontend starts on `http://localhost:3000`.

---

## 🧪 Comprehensive API Testing Guide (curl)

### 1. Create a Geofence
```bash
curl -X POST http://localhost:8080/geofences \
  -H "Content-Type: application/json" \
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

### 2. Retrieve All Geofences (with optional category filter)
```bash
# All geofences
curl http://localhost:8080/geofences

# Filtered by category
curl "http://localhost:8080/geofences?category=delivery_zone"
```

### 3. Register a Vehicle
```bash
curl -X POST http://localhost:8080/vehicles \
  -H "Content-Type: application/json" \
  -d '{
    "vehicle_number": "KA-01-AB-1234",
    "driver_name": "John Doe",
    "vehicle_type": "truck",
    "phone": "+1234567890"
  }'
```

### 4. Retrieve All Vehicles
```bash
curl http://localhost:8080/vehicles
```

### 5. Update Vehicle Location (Triggers Entry/Exit Events & Alerts)
```bash
# Update position inside geofence
curl -X POST http://localhost:8080/vehicles/location \
  -H "Content-Type: application/json" \
  -d '{
    "vehicle_id": "veh_truck_101",
    "latitude": 37.7800,
    "longitude": -122.4150,
    "timestamp": "2025-01-15T10:35:00Z"
  }'
```

### 6. Get Current Vehicle Location & Geofence Status
```bash
curl http://localhost:8080/vehicles/location/veh_truck_101
```

### 7. Configure Alert Rule
```bash
curl -X POST http://localhost:8080/alerts/configure \
  -H "Content-Type: application/json" \
  -d '{
    "geofence_id": "geo_downtown_01",
    "vehicle_id": "veh_truck_101",
    "event_type": "entry"
  }'
```

### 8. Retrieve Configured Alert Rules
```bash
curl http://localhost:8080/alerts
```

### 9. Retrieve Violation History
```bash
curl "http://localhost:8080/violations/history?limit=50"
```

---

## 📹 Video Presentation Roadmap (30-45 Minutes)

For your video submission, follow this structured agenda:

### Section 1: Live Application Demo (10–12 Minutes)
1. **Introduction**: Introduce yourself and present the high-level architecture overview.
2. **Dashboard Tour**:
   - Show the interactive map with color-coded polygons (Red = Restricted, Blue = Delivery, Amber = Toll, Green = Customer).
   - Point out real-time WebSocket connection status badge.
3. **Geofence Creation**:
   - Use the "Draw" tool on the map to define vertices and save a new zone.
   - Show the polygon rendered on the map and appearing in the zones list.
4. **Vehicle Registration & Live Tracking**:
   - Register a new vehicle with assigned driver.
   - Use the "Move" tool to click on the map and teleport the vehicle.
5. **Real-time Alert Demonstration**:
   - Move a vehicle from outside a zone to inside a **Restricted Zone**.
   - Show the immediate pop-up toast and live alert streaming into the WebSocket feed with zero latency.
6. **Audit & Violation History**:
   - Open the "History Logs" tab.
   - Filter by vehicle, geofence, and record limit to inspect historical entries and exits.
7. **Docker Compose Demonstration**:
   - Show `docker-compose up` running the database, backend, and frontend containers locally.

### Section 2: Implementation Deep Dive (12–18 Minutes)
1. **Project Structure**:
   - Explain clean architectural boundaries: `domain`, `geospatial`, `repository`, `service`, `handler`, `websocket`.
2. **Point-in-Polygon (PIP) Algorithm**:
   - Walk through `internal/geospatial/polygon.go`.
   - Explain the $O(1)$ Bounding Box optimization for instant rejection.
   - Explain the Ray-Casting algorithm and how edge boundaries are handled using `ST_Covers` inclusive semantics.
   - Highlight coordinate ordering: API uses `[lat, lon]`, whereas PostGIS uses `[lon, lat]` (X, Y).
3. **Stateful Transition Detection**:
   - Walk through `internal/service/tracker.go`.
   - Explain comparing `previousGeofenceSet` vs `currentGeofenceSet` to calculate entries and exits.
   - Show how alerts are evaluated and dispatched into an asynchronous goroutine.
4. **WebSocket Concurrency & Hub Architecture**:
   - Walk through `internal/websocket/hub.go`.
   - Discuss how the channel-based loop prevents race conditions on `clients map[*Client]bool`.
   - Explain dead-connection drops and heartbeat handling.
5. **Execution Timing (`time_ns`)**:
   - Show `internal/handler/response.go` and explain how every response (success and error) records nanosecond duration.
6. **Database Layer (Dual Engine)**:
   - Walk through `internal/repository/postgres.go` (PostGIS `ST_Covers`, GiST index).
   - Explain the graceful fallback to `internal/repository/memory.go`.

### Section 3: Gaps, Production Scaling & Roadmap (10–15 Minutes)
1. **Edge Cases & Nuances**:
   - Complex concave/self-intersecting polygons (validating simple polygons with Jordan Curve Theorem).
   - GPS drift/jitter: How to add Kalman filtering or speed thresholds so a vehicle idling on a boundary does not flutter alerts.
2. **Horizontal Scaling to 100,000+ Vehicles**:
   - **Distributed Message Broker**: Introduce Redis Pub/Sub or Apache Kafka to distribute WebSocket alerts across multiple backend replicas.
   - **Spatial Indexing & Sharding**: PostGIS table partitioning by region or geohash.
   - **In-Memory Spatial Grid / R-Tree Cache**: Caching active geofences in an in-memory R-Tree (e.g., `tidwall/rtree`) so containment checks run in sub-millisecond memory without querying the database for every ping.
3. **Security & Production Hardening**:
   - JWT authentication for drivers and fleet managers.
   - Redis-backed rate limiting on `/vehicles/location` per vehicle API key.
   - Mutual TLS (mTLS) for IoT telemetry devices.

---

## 🐳 Docker Hub Push & Cloud Deployment Guide

### 1. Build & Push Images to Docker Hub

Both services are containerized with production-ready multi-stage Dockerfiles.

```bash
# Push Backend
docker tag mapup-backend:latest gowtham095/mapup-geofence-backend:latest
docker push gowtham095/mapup-geofence-backend:latest

# Push Frontend
docker tag mapup-frontend:latest gowtham095/mapup-geofence-frontend:latest
docker push gowtham095/mapup-geofence-frontend:latest
```

### 2. Cloud Deployment (Render / Railway / Koyeb)

**Option A: Deploy via Docker Images on Render (Free Tier & WebSocket Support)**
- **Backend Service**:
  - Image: `docker.io/gowtham095/mapup-geofence-backend:latest`
  - Port: `8080`
  - Environment: `PORT=8080`, `API_KEY=mapup-secret-key-2025`
- **Frontend Service**:
  - Image: `docker.io/gowtham095/mapup-geofence-frontend:latest`
  - Port: `80` (Nginx)

**Option A: Deploy on Render (Recommended - Free Tier & WebSocket Support)**
1. Go to [render.com](https://render.com) and create a **New Web Service**.
2. Connect your GitHub repository or enter your Docker Hub image:
   `docker.io/<your-dockerhub-username>/mapup-geofence-backend:latest`
3. Set Environment Variables:
   - `PORT`: `8080`
   - `DATABASE_URL`: `postgres://...` (or leave blank to use fast In-Memory mode)
   - `API_KEY`: `mapup-secret-key-2025`
4. Click **Deploy**. Render will generate a public HTTPS URL (e.g., `https://mapup-backend.onrender.com`).

**Option B: Deploy on Railway**
1. Go to [railway.app](https://railway.app) and create a **New Project**.
2. Choose **Deploy from Docker Hub** $\rightarrow$ `<your-dockerhub-username>/mapup-geofence-backend:latest`.
3. Add variable `PORT=8080`.
4. Generate Domain $\rightarrow$ gives public URL.

### 3. Frontend Deployment (Vercel / Netlify)

Deploy the React application to [Vercel](https://vercel.com):
1. Import your GitHub repository in Vercel.
2. Set **Root Directory** to `frontend`.
3. Framework Preset: **Vite**.
4. Configure Environment Variables:
   - `VITE_API_URL`: `https://<your-deployed-backend-url>`
   - `VITE_WS_URL`: `wss://<your-deployed-backend-host>/ws/alerts`
5. Click **Deploy**. Vercel will build and serve your app globally.
