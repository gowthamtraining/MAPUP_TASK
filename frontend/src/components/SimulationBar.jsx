import React, { useState } from 'react';
import { Play, ShieldAlert, Sparkles, Navigation, RotateCcw } from 'lucide-react';
import { api } from '../api';

export default function SimulationBar({ vehicles = [], onSimulationTriggered }) {
  const [running, setRunning] = useState(false);
  const [statusMsg, setStatusMsg] = useState('');

  const triggerStep = async (vehicleId, lat, lon, stepName) => {
    setStatusMsg(`Simulating: ${stepName}...`);
    await api.updateLocation(vehicleId, lat, lon);
    onSimulationTriggered && onSimulationTriggered();
    await new Promise((r) => setTimeout(r, 1200));
  };

  // Full Trip Simulation: moves a vehicle through outside -> Downtown Delivery -> Exit -> Restricted Breach
  const runFullTripSimulation = async () => {
    if (vehicles.length === 0) return;
    const vehicle = vehicles[0];
    setRunning(true);

    try {
      // 1. Outside
      await triggerStep(vehicle.id, 37.7650, -122.4300, `${vehicle.vehicle_number} starting trip (outside)`);
      // 2. Enter Downtown Delivery Zone (Entry alert)
      await triggerStep(vehicle.id, 37.7780, -122.4150, `${vehicle.vehicle_number} entering Downtown Delivery Zone`);
      // 3. Move within delivery zone
      await triggerStep(vehicle.id, 37.7810, -122.4120, `${vehicle.vehicle_number} moving in Delivery Zone`);
      // 4. Exit delivery zone
      await triggerStep(vehicle.id, 37.7880, -122.4050, `${vehicle.vehicle_number} exiting Delivery Zone`);
      // 5. Enter Restricted Security Harbor (Immediate Breach Alert!)
      await triggerStep(vehicle.id, 37.7950, -122.3950, `🚨 ${vehicle.vehicle_number} ENTERING RESTRICTED ZONE!`);
      setStatusMsg('Trip Simulation Completed Successfully!');
    } catch (err) {
      console.error(err);
      setStatusMsg('Simulation encountered an error: ' + err.message);
    } finally {
      setRunning(false);
      setTimeout(() => setStatusMsg(''), 4000);
    }
  };

  // Quick Restricted Zone breach trigger
  const triggerRestrictedBreach = async () => {
    if (vehicles.length === 0) return;
    const vehicle = vehicles[0];
    setRunning(true);
    setStatusMsg(`Breaching Restricted Zone with ${vehicle.vehicle_number}...`);
    try {
      // First move outside to ensure transition
      await api.updateLocation(vehicle.id, 37.7850, -122.4050);
      await new Promise((r) => setTimeout(r, 600));
      // Now enter restricted zone
      await api.updateLocation(vehicle.id, 37.7950, -122.3950);
      onSimulationTriggered && onSimulationTriggered();
      setStatusMsg('🚨 Restricted Zone breach triggered live via WebSocket!');
    } catch (err) {
      setStatusMsg('Error: ' + err.message);
    } finally {
      setRunning(false);
      setTimeout(() => setStatusMsg(''), 4000);
    }
  };

  // Quick Delivery Entry trigger
  const triggerDeliveryEntry = async () => {
    if (vehicles.length === 0) return;
    const vehicle = vehicles[0];
    setRunning(true);
    setStatusMsg(`Entering Delivery Zone with ${vehicle.vehicle_number}...`);
    try {
      // Move outside first
      await api.updateLocation(vehicle.id, 37.7650, -122.4300);
      await new Promise((r) => setTimeout(r, 600));
      // Move inside
      await api.updateLocation(vehicle.id, 37.7780, -122.4150);
      onSimulationTriggered && onSimulationTriggered();
      setStatusMsg(`Entered Downtown Delivery Zone!`);
    } catch (err) {
      setStatusMsg('Error: ' + err.message);
    } finally {
      setRunning(false);
      setTimeout(() => setStatusMsg(''), 4000);
    }
  };

  return (
    <div className="bg-slate-900/90 border border-slate-800 rounded-xl px-4 py-2.5 shadow-xl flex flex-wrap items-center justify-between gap-3 text-xs">
      <div className="flex items-center gap-2">
        <Sparkles className="w-4 h-4 text-purple-400" />
        <span className="font-bold text-slate-200">Live Assessment Demo Simulator:</span>
        {statusMsg && (
          <span className="text-purple-300 font-medium animate-pulse ml-2">
            {statusMsg}
          </span>
        )}
      </div>

      <div className="flex items-center gap-2">
        <button
          onClick={triggerDeliveryEntry}
          disabled={running || vehicles.length === 0}
          className="bg-blue-600/80 hover:bg-blue-500 disabled:opacity-50 text-white font-medium px-3 py-1.5 rounded-lg border border-blue-400/30 transition-colors flex items-center gap-1.5"
        >
          <Navigation className="w-3.5 h-3.5" />
          Test Delivery Entry
        </button>

        <button
          onClick={triggerRestrictedBreach}
          disabled={running || vehicles.length === 0}
          className="bg-rose-600/90 hover:bg-rose-500 disabled:opacity-50 text-white font-medium px-3 py-1.5 rounded-lg border border-rose-400/40 shadow-sm transition-colors flex items-center gap-1.5"
        >
          <ShieldAlert className="w-3.5 h-3.5" />
          Trigger Restricted Breach
        </button>

        <button
          onClick={runFullTripSimulation}
          disabled={running || vehicles.length === 0}
          className="bg-gradient-to-r from-purple-600 to-indigo-600 hover:from-purple-500 hover:to-indigo-500 disabled:opacity-50 text-white font-semibold px-3.5 py-1.5 rounded-lg shadow-md transition-all flex items-center gap-1.5"
        >
          <Play className={`w-3.5 h-3.5 ${running ? 'animate-spin' : ''}`} />
          Run Full Trip Simulation
        </button>
      </div>
    </div>
  );
}
