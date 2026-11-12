# 🌐 Geofencing & Real-Time Alert System

**Full-Stack Developer Assessment Submission for MapUp**

[![Go](https://img.shields.io/badge/GoLang-1.23+-00ADD8?style=flat&logo=go)](https://golang.org)
[![React](https://img.shields.io/badge/React-19-61DAFB?style=flat&logo=react)](https://react.dev)
[![PostGIS](https://img.shields.io/badge/PostgreSQL-PostGIS-336791?style=flat&logo=postgresql)](https://postgis.net)
[![WebSocket](https://img.shields.io/badge/Real--Time-WebSocket-brightgreen?style=flat)]()
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=flat&logo=docker)](https://www.docker.com)

---

## 📺 Video Walkthrough Submission
> **Video Link**: [Insert Your YouTube / Loom / Google Drive Video Link Here]  
> *(Video duration: 30–45 minutes walking through Live App Demo, Architecture Deep Dive, and Production Scaling Roadmap)*

---

## 🐳 Docker Hub Image
- **Repository**: [`gowtham095/mapup-geofence-backend`](https://hub.docker.com/r/gowtham095/mapup-geofence-backend)
- **Tag**: `latest`
- **Docker Pull**:
  ```bash
  docker pull gowtham095/mapup-geofence-backend:latest
  ```

---

## 👥 Assessment Collaborators
When creating the private GitHub repository, ensure the following MapUp collaborators are invited:
- `vedantp@mapup.ai`
- `ajayap@mapup.ai`
- `jyotis@mapup.ai`
- `tanishqs@mapup.ai`
- `naveenn@mapup.ai`
- `asijitp@mapup.ai`
- `dewanshp@mapup.ai`
- `karkuvelpandip@mapup.ai`
- `parask@mapup.ai`

---

## ⚡ Core Highlights & Capabilities

- **Strict API Compliance**: Exact REST endpoints, method signatures, parameter names, and payload structures matching the MapUp specification.
- **Microsecond/Nanosecond Precision**: Every API response (including error responses) returns `"time_ns"`.
- **Advanced Geospatial Engine**:
  - $O(1)$ Bounding-Box rejection + Ray-Casting Point-in-Polygon calculation.
  - Boundary inclusive check (`ST_Covers` semantics).
  - Handles PostGIS `[longitude, latitude]` vs API `[latitude, longitude]` coordinate translation seamlessly.
- **Stateful Transition Detector**:
  - Evaluates transitions (`entry` and `exit`), not merely static containment.
  - Automatically fires critical security alerts upon entry into `restricted_zone` areas.
  - Evaluates configurable vehicle-specific and global alert rules (`entry`, `exit`, `both`).
- **Concurrent Non-Blocking Real-Time Alerts**:
  - WebSocket hub (`/ws/alerts`) with mutex-free channel-based event loop.
  - Ping/pong heartbeat dead-connection cleanup.
  - Alert delivery is offloaded asynchronously to ensure minimal HTTP latency for incoming location pings.
- **Interactive React Frontend**:
  - Interactive Leaflet map with color-coded polygons, pulsing vehicle markers, and popup details.
  - Drawing tool to create polygon boundaries directly by clicking on the map.
  - Interactive location teleport/move tool.
  - Live streaming WebSocket alert feed with unread counter badges and toast notifications.
  - Violation audit history with multi-filter query builder (vehicle, zone, limit).
  - One-click Simulation Bar for live demonstration of route traversal and restricted zone breaches.
- **Dual Persistence Architecture**:
  - Production-ready PostgreSQL + PostGIS with GiST spatial indexing and automatic schema migration.
  - High-performance in-memory repository fallback for instant local execution without external dependencies.

---

## 🚀 Quick Run with Docker Compose

```bash
docker-compose up --build
```

Access the applications:
- **Frontend Dashboard**: [http://localhost:3000](http://localhost:3000)
- **Backend API**: [http://localhost:8080](http://localhost:8080)
- **WebSocket Endpoint**: `ws://localhost:8080/ws/alerts`

---

## 🛠️ Local Development (Standalone)

### 1. Backend (Go)
```bash
cd backend
go run ./cmd/server
```

### 2. Frontend (React + Vite)
```bash
cd frontend
npm install
npm run dev
```

For detailed documentation, architectural diagrams, curl commands, and video presentation guide, please refer to [SETUP.md](./SETUP.md).
