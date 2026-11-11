import React from 'react';
import { AlertCircle, ArrowRightCircle, ArrowLeftCircle, ShieldAlert, Clock, Bell, Trash2 } from 'lucide-react';

export default function AlertFeed({ alerts = [], onClear, isConnected }) {
  return (
    <div className="bg-slate-900 border border-slate-800 rounded-xl p-4 shadow-xl flex flex-col h-full">
      <div className="flex items-center justify-between pb-3 border-b border-slate-800 mb-3">
        <div className="flex items-center gap-2">
          <div className="relative">
            <Bell className="w-5 h-5 text-amber-400" />
            {alerts.length > 0 && (
              <span className="absolute -top-1.5 -right-1.5 bg-rose-500 text-white text-[10px] font-extrabold w-4 h-4 rounded-full flex items-center justify-center animate-bounce">
                {alerts.length > 99 ? '99+' : alerts.length}
              </span>
            )}
          </div>
          <div>
            <h2 className="text-sm font-bold text-white tracking-wide">Live Real-Time Alerts</h2>
            <div className="flex items-center gap-1.5 text-[11px] text-slate-400">
              <span className={`w-2 h-2 rounded-full ${isConnected ? 'bg-emerald-400' : 'bg-rose-500 animate-pulse'}`} />
              {isConnected ? 'WebSocket Connected' : 'Reconnecting...'}
            </div>
          </div>
        </div>

        {alerts.length > 0 && (
          <button
            onClick={onClear}
            className="text-xs text-slate-400 hover:text-slate-200 hover:bg-slate-800 px-2.5 py-1 rounded-md transition-colors flex items-center gap-1"
            title="Clear Feed"
          >
            <Trash2 className="w-3.5 h-3.5" />
            Clear
          </button>
        )}
      </div>

      <div className="flex-1 overflow-y-auto space-y-2.5 pr-1">
        {alerts.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-48 text-center p-4 text-slate-500">
            <AlertCircle className="w-10 h-10 mb-2 stroke-1 text-slate-600" />
            <p className="text-xs font-medium">No real-time alerts yet</p>
            <p className="text-[11px] text-slate-600 mt-1 max-w-[200px]">
              Alerts stream live when vehicles enter or exit monitored or restricted zones
            </p>
          </div>
        ) : (
          alerts.map((alert, idx) => {
            const isEntry = alert.event_type?.toLowerCase() === 'entry';
            const isRestricted = alert.geofence?.category === 'restricted_zone';

            return (
              <div
                key={alert.event_id || idx}
                className={`alert-pop p-3 rounded-lg border text-xs transition-all ${
                  isRestricted && isEntry
                    ? 'bg-rose-950/40 border-rose-500/60 shadow-lg shadow-rose-950/30'
                    : isEntry
                    ? 'bg-emerald-950/30 border-emerald-500/40'
                    : 'bg-amber-950/30 border-amber-500/40'
                }`}
              >
                <div className="flex items-center justify-between mb-1.5">
                  <div className="flex items-center gap-1.5">
                    {isRestricted ? (
                      <span className="flex items-center gap-1 text-[10px] font-bold uppercase tracking-wider bg-rose-600 text-white px-2 py-0.5 rounded shadow-sm">
                        <ShieldAlert className="w-3 h-3" />
                        RESTRICTED BREACH
                      </span>
                    ) : (
                      <span
                        className={`flex items-center gap-1 text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded ${
                          isEntry
                            ? 'bg-emerald-600/90 text-white'
                            : 'bg-amber-600/90 text-white'
                        }`}
                      >
                        {isEntry ? <ArrowRightCircle className="w-3 h-3" /> : <ArrowLeftCircle className="w-3 h-3" />}
                        {alert.event_type}
                      </span>
                    )}
                  </div>
                  <div className="flex items-center gap-1 text-[10px] text-slate-400">
                    <Clock className="w-3 h-3" />
                    {new Date(alert.timestamp).toLocaleTimeString()}
                  </div>
                </div>

                <div className="mt-1">
                  <div className="font-semibold text-slate-100 flex items-center justify-between">
                    <span>{alert.vehicle?.vehicle_number}</span>
                    <span className="text-[11px] font-normal text-slate-400">
                      Driver: {alert.vehicle?.driver_name}
                    </span>
                  </div>

                  <div className="text-[11px] text-slate-300 mt-1 flex items-center gap-1.5">
                    <span className="text-slate-400">Zone:</span>
                    <span className="font-medium text-slate-200">{alert.geofence?.geofence_name}</span>
                    <span className="text-[9px] bg-slate-800 text-slate-400 px-1 rounded uppercase">
                      {alert.geofence?.category?.replace('_', ' ')}
                    </span>
                  </div>

                  {alert.location && (
                    <div className="text-[10px] text-slate-500 mt-1 font-mono">
                      Location: {alert.location.latitude?.toFixed(4)}, {alert.location.longitude?.toFixed(4)}
                    </div>
                  )}
                </div>
              </div>
            );
          })
        )}
      </div>
    </div>
  );
}
