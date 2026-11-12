import React, { useEffect, useMemo } from 'react';
import { MapContainer, TileLayer, Polygon, Marker, Popup, Tooltip, useMapEvents, useMap } from 'react-leaflet';
import L from 'leaflet';
import { Truck, Car, AlertTriangle, ShieldAlert } from 'lucide-react';

const CATEGORY_COLORS = {
  restricted_zone: { stroke: '#ef4444', fill: '#ef4444', label: 'Restricted Zone' },
  delivery_zone: { stroke: '#3b82f6', fill: '#3b82f6', label: 'Delivery Zone' },
  toll_zone: { stroke: '#f59e0b', fill: '#f59e0b', label: 'Toll Zone' },
  customer_area: { stroke: '#10b981', fill: '#10b981', label: 'Customer Area' },
};

// Create custom DOM Marker with SVG icon
function createVehicleIcon(vehicle, isSelected) {
  const isTruck = vehicle.vehicle_type?.toLowerCase() === 'truck';
  const color = isSelected ? '#a855f7' : '#38bdf8';
  const html = `
    <div style="display:flex; flex-direction:column; align-items:center; cursor:pointer; width:80px;">
      <div style="background:#0f172a; color:#fff; border: 1.5px solid ${color}; border-radius: 6px; padding: 1px 5px; font-size: 10px; font-weight: 700; white-space:nowrap; box-shadow: 0 4px 6px rgba(0,0,0,0.5); margin-bottom: 2px;">
        ${vehicle.vehicle_number}
      </div>
      <div style="position:relative; width: 34px; height: 34px; background: ${color}; border-radius: 50%; display:flex; align-items:center; justify-content:center; box-shadow: 0 0 12px ${color}88;">
        <span style="color:#0f172a; font-weight:900; font-size:16px; line-height:1;">
          ${isTruck ? '🚛' : '🚗'}
        </span>
        <div style="position:absolute; width: 44px; height: 44px; border: 2px solid ${color}; border-radius:50%; animation: ping 1.5s cubic-bezier(0, 0, 0.2, 1) infinite; opacity: 0.75;"></div>
      </div>
    </div>
  `;
  return L.divIcon({
    html,
    className: 'custom-vehicle-marker',
    iconSize: [80, 56],
    iconAnchor: [40, 38], // Exact center of the 34px circular icon
  });
}

function MapClickHandler({ mode, onMapClick }) {
  useMapEvents({
    click(e) {
      if (mode === 'draw' || mode === 'teleport') {
        onMapClick(e.latlng.lat, e.latlng.lng);
      }
    },
  });
  return null;
}

// Center map view on vehicle selection
function MapViewController({ center }) {
  const map = useMap();
  useEffect(() => {
    if (center && center[0] && center[1]) {
      map.setView(center, map.getZoom(), { animate: true });
    }
  }, [center, map]);
  return null;
}

export default function MapView({
  geofences = [],
  vehicles = [],
  vehicleLocations = {},
  selectedVehicleId,
  onSelectVehicle,
  mapMode = 'view', // 'view' | 'draw' | 'teleport'
  drawnPoints = [],
  onMapClick,
}) {
  const defaultCenter = [37.7749, -122.4194];

  // Selected vehicle's coordinates if any
  const selectedVehicleCoord = useMemo(() => {
    if (selectedVehicleId && vehicleLocations[selectedVehicleId]) {
      const loc = vehicleLocations[selectedVehicleId].current_location;
      if (loc) return [loc.latitude, loc.longitude];
    }
    return null;
  }, [selectedVehicleId, vehicleLocations]);

  return (
    <div className="relative w-full h-full rounded-xl overflow-hidden shadow-2xl border border-slate-700">
      <MapContainer
        center={defaultCenter}
        zoom={13}
        scrollWheelZoom={true}
        className="w-full h-full"
      >
        <TileLayer
          attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        />

        <MapClickHandler mode={mapMode} onMapClick={onMapClick} />
        {selectedVehicleCoord && <MapViewController center={selectedVehicleCoord} />}

        {/* Render Existing Geofences */}
        {geofences.map((g) => {
          const style = CATEGORY_COLORS[g.category] || { stroke: '#94a3b8', fill: '#94a3b8' };
          const positions = g.coordinates.map((c) => [c[0], c[1]]);
          const isInteractive = mapMode === 'view';

          return (
            <Polygon
              key={`${g.id}-${mapMode}`}
              positions={positions}
              pathOptions={{
                color: style.stroke,
                fillColor: style.fill,
                fillOpacity: 0.25,
                weight: 2.5,
                dashArray: g.category === 'restricted_zone' ? '6, 6' : undefined,
              }}
              eventHandlers={{
                click: (e) => {
                  if (mapMode === 'draw' || mapMode === 'teleport') {
                    if (e.originalEvent) {
                      e.originalEvent.stopPropagation();
                    }
                    onMapClick(e.latlng.lat, e.latlng.lng);
                  }
                },
              }}
            >
              {isInteractive && (
                <>
                  <Tooltip sticky>
                    <div className="p-1">
                      <div className="font-bold text-slate-900">{g.name}</div>
                      <div className="text-xs font-semibold uppercase tracking-wider text-slate-500">
                        {g.category.replace('_', ' ')}
                      </div>
                    </div>
                  </Tooltip>
                  <Popup>
                    <div className="text-slate-900 p-1 min-w-[180px]">
                      <div className="flex items-center gap-2 mb-1">
                        {g.category === 'restricted_zone' && <ShieldAlert className="w-5 h-5 text-red-500" />}
                        <h3 className="font-bold text-sm">{g.name}</h3>
                      </div>
                      <span
                        className="inline-block text-[11px] font-semibold uppercase tracking-wider px-2 py-0.5 rounded text-white mb-2"
                        style={{ backgroundColor: style.stroke }}
                      >
                        {g.category.replace('_', ' ')}
                      </span>
                      <p className="text-xs text-slate-600 mb-2">{g.description || 'No description provided'}</p>
                      <div className="text-[10px] text-slate-400">ID: {g.id}</div>
                    </div>
                  </Popup>
                </>
              )}
            </Polygon>
          );
        })}

        {/* Render Polygon Currently Being Drawn */}
        {drawnPoints.length > 0 && (
          <Polygon
            positions={drawnPoints}
            pathOptions={{
              color: '#f43f5e',
              fillColor: '#f43f5e',
              fillOpacity: 0.2,
              weight: 2,
              dashArray: '4, 4',
            }}
          />
        )}

        {/* Render Temporary Markers for Points in Drawing Mode */}
        {mapMode === 'draw' &&
          drawnPoints.map((pt, i) => (
            <Marker
              key={`draw-pt-${i}`}
              position={pt}
              icon={L.divIcon({
                html: `<div style="width:12px; height:12px; background:#f43f5e; border:2px solid white; border-radius:50%; box-shadow:0 0 6px #f43f5e;"></div>`,
                className: '',
                iconSize: [12, 12],
                iconAnchor: [6, 6],
              })}
            />
          ))}

        {/* Render Vehicle Markers */}
        {vehicles.map((v) => {
          const locData = vehicleLocations[v.id];
          if (!locData || !locData.current_location) return null;
          const pos = [locData.current_location.latitude, locData.current_location.longitude];
          const isSelected = selectedVehicleId === v.id;
          const isInteractive = mapMode === 'view';

          return (
            <Marker
              key={`${v.id}-${mapMode}`}
              position={pos}
              icon={createVehicleIcon(v, isSelected)}
              eventHandlers={{
                click: (e) => {
                  if (mapMode === 'draw' || mapMode === 'teleport') {
                    if (e.originalEvent) {
                      e.originalEvent.stopPropagation();
                    }
                    onMapClick(pos[0], pos[1]);
                  } else {
                    onSelectVehicle && onSelectVehicle(v.id);
                  }
                },
              }}
            >
              {isInteractive && (
                <Popup>
                  <div className="text-slate-900 p-1 min-w-[200px]">
                    <div className="flex items-center justify-between border-b pb-1 mb-2">
                      <span className="font-bold text-sm text-slate-800">{v.vehicle_number}</span>
                      <span className="text-xs bg-slate-100 text-slate-700 px-1.5 py-0.5 rounded capitalize">
                        {v.vehicle_type}
                      </span>
                    </div>
                    <div className="text-xs space-y-1 mb-2">
                      <div>
                        <span className="text-slate-500">Driver:</span> <strong className="text-slate-700">{v.driver_name}</strong>
                      </div>
                      <div>
                        <span className="text-slate-500">Phone:</span> {v.phone}
                      </div>
                      <div>
                        <span className="text-slate-500">Coordinates:</span>{' '}
                        {pos[0].toFixed(5)}, {pos[1].toFixed(5)}
                      </div>
                    </div>

                    <div className="border-t pt-1.5">
                      <div className="text-[11px] font-semibold text-slate-500 uppercase tracking-wider mb-1">
                        Current Zones:
                      </div>
                      {locData.current_geofences && locData.current_geofences.length > 0 ? (
                        <div className="flex flex-wrap gap-1">
                          {locData.current_geofences.map((cg) => (
                            <span
                              key={cg.geofence_id}
                              className="text-[10px] bg-blue-100 text-blue-800 font-medium px-1.5 py-0.5 rounded"
                            >
                              {cg.geofence_name}
                            </span>
                          ))}
                        </div>
                      ) : (
                        <span className="text-[11px] text-slate-400 italic">Outside all geofences</span>
                      )}
                    </div>
                  </div>
                </Popup>
              )}
            </Marker>
          );
        })}
      </MapContainer>

      {/* Floating Instructions Banner when Draw or Teleport mode is active */}
      {mapMode === 'draw' && (
        <div className="absolute top-4 left-1/2 -translate-x-1/2 z-[1000] bg-rose-600/95 text-white px-4 py-2 rounded-lg shadow-xl backdrop-blur-sm border border-rose-400/30 flex items-center gap-3 text-xs font-semibold tracking-wide">
          <span className="w-2.5 h-2.5 rounded-full bg-white animate-pulse" />
          Click on the map to place polygon points. At least 3 points needed.
        </div>
      )}
      {mapMode === 'teleport' && (
        <div className="absolute top-4 left-1/2 -translate-x-1/2 z-[1000] bg-purple-600/95 text-white px-4 py-2 rounded-lg shadow-xl backdrop-blur-sm border border-purple-400/30 flex items-center gap-3 text-xs font-semibold tracking-wide">
          <span className="w-2.5 h-2.5 rounded-full bg-white animate-pulse" />
          Click anywhere on the map to move vehicle {vehicles.find((v) => v.id === selectedVehicleId)?.vehicle_number}
        </div>
      )}
    </div>
  );
}
