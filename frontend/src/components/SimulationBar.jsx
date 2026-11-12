import React, { useState } from 'react';
import { Play, ShieldAlert, Sparkles, Navigation, RotateCcw } from 'lucide-react';
import { api } from '../api';

export default function SimulationBar({ vehicles = [], selectedVehicleId, onSimulationTriggered }) {
  const [running, setRunning] = useState(false);
  const [statusMsg, setStatusMsg] = useState('');

  const targetVehicle = vehicles.find((v) => v.id === selectedVehicleId) || vehicles[0];

  const triggerStep = async (vehicleId, lat, lon, stepName) => {
    setStatusMsg(`${stepName}`);
    await api.updateLocation(vehicleId, lat, lon);
    onSimulationTriggered && onSimulationTriggered();
    await new Promise((r) => setTimeout(r, 1200));
  };

  // Full Trip Simulation: moves the vehicle through outside -> Downtown Delivery -> Exit -> Restricted Breach
  const runFullTripSimulation = async () => {
    if (!targetVehicle) return;
    setRunning(true);

    try {
      // Step 1: Start outside all geofences
      await triggerStep(targetVehicle.id, 37.7650, -122.4300, `Step 1/5: ${targetVehicle.vehicle_number} starting outside all zones`);
      // Step 2: Enter Downtown Delivery Zone (triggers delivery_zone entry alert)
      await triggerStep(targetVehicle.id, 37.7780, -122.4150, `Step 2/5: ${targetVehicle.vehicle_number} entered Downtown Delivery Zone`);
      // Step 3: Move inside delivery zone (no duplicate alert)
      await triggerStep(targetVehicle.id, 37.7810, -122.4120, `Step 3/5: ${targetVehicle.vehicle_number} navigating within Delivery Zone`);
      // Step 4: Exit delivery zone (triggers delivery_zone exit alert)
      await triggerStep(targetVehicle.id, 37.7880, -122.4050, `Step 4/5: ${targetVehicle.vehicle_number} exited Delivery Zone`);
      // Step 5: Breaches Port Security Restricted Zone (triggers instant HIGH-PRIORITY restricted alert!)
      await triggerStep(targetVehicle.id, 37.7950, -122.3950, `Step 5/5: 🚨 ${targetVehicle.vehicle_number} BREACHED RESTRICTED ZONE!`);
      setStatusMsg(`Trip Simulation Complete for ${targetVehicle.vehicle_number}!`);
    } catch (err) {
      console.error(err);
      setStatusMsg('Simulation encountered an error: ' + err.message);
    } finally {
      setRunning(false);
      setTimeout(() => setStatusMsg(''), 5000);
    }
  };

  // Quick Restricted Zone breach trigger
  const triggerRestrictedBreach = async () => {
    if (!targetVehicle) return;
    setRunning(true);
    setStatusMsg(`Breaching Restricted Zone with ${targetVehicle.vehicle_number}...`);
    try {
      // Move outside first to ensure transition
      await api.updateLocation(targetVehicle.id, 37.7850, -122.4050);
      await new Promise((r) => setTimeout(r, 600));
      // Enter restricted zone (fires WebSocket alert)
      await api.updateLocation(targetVehicle.id, 37.7950, -122.3950);
      onSimulationTriggered && onSimulationTriggered();
      setStatusMsg(`🚨 Restricted Zone breach triggered with ${targetVehicle.vehicle_number}!`);
    } catch (err) {
      setStatusMsg('Error: ' + err.message);
    } finally {
      setRunning(false);
      setTimeout(() => setStatusMsg(''), 5000);
    }
  };

  // Quick Delivery Entry trigger
  const triggerDeliveryEntry = async () => {
    if (!targetVehicle) return;
    setRunning(true);
    setStatusMsg(`Entering Delivery Zone with ${targetVehicle.vehicle_number}...`);
    try {
      // Move outside first
      await api.updateLocation(targetVehicle.id, 37.7650, -122.4300);
      await new Promise((r) => setTimeout(r, 600));
      // Enter delivery zone
      await api.updateLocation(targetVehicle.id, 37.7780, -122.4150);
      onSimulationTriggered && onSimulationTriggered();
      setStatusMsg(`Entered Downtown Delivery Zone with ${targetVehicle.vehicle_number}!`);
    } catch (err) {
      setStatusMsg('Error: ' + err.message);
    } finally {
      setRunning(false);
      setTimeout(() => setStatusMsg(''), 5000);
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
