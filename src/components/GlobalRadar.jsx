import React, { useEffect, useState } from 'react';

const mockLogs = [
  { id: 1, type: 'SENSOR', msg: 'Detected volatility spike in sector 7.', time: '10:42:01', color: 'text-blue-400' },
  { id: 2, type: 'STRATEGY', msg: 'Re-evaluating aggressive parameters...', time: '10:42:05', color: 'text-purple-400' },
  { id: 3, type: 'VAULT', msg: 'Secured 500 units of capital allocation.', time: '10:42:12', color: 'text-amber-500' },
  { id: 4, type: 'EXCHANGE', msg: 'Executed trade: BUY 100 @ 1.24.', time: '10:42:15', color: 'text-emerald-400' },
  { id: 5, type: 'SENSOR', msg: 'Nominal conditions restored. Monitoring...', time: '10:42:45', color: 'text-blue-400 opacity-75' },
];

export default function GlobalRadar() {
  const [logs, setLogs] = useState(mockLogs);

  // In a real app, this would stream via WebSocket or polling
  
  return (
    <div className="bg-black border border-emerald-900/50 rounded-lg h-96 w-full max-w-xl flex flex-col overflow-hidden shadow-2xl relative">
      {/* Radar sweeping effect overlay placeholder */}
      <div className="absolute inset-0 bg-gradient-to-b from-transparent to-emerald-900/10 pointer-events-none"></div>
      
      <div className="p-4 border-b border-emerald-900/30 bg-gray-950 flex justify-between items-center z-10">
        <h2 className="text-xl font-bold text-emerald-500 uppercase tracking-widest flex items-center gap-2">
          <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></span>
          Global Radar
        </h2>
        <span className="text-xs font-mono text-emerald-700">SYS_LOG_ACTIVE</span>
      </div>

      <div className="flex-1 overflow-y-auto p-4 font-mono text-sm space-y-2 flex flex-col-reverse z-10">
        {/* Render bottom-up for scrolling effect */}
        {[...logs].reverse().map((log) => (
          <div key={log.id} className="flex gap-3 hover:bg-gray-900/50 p-1 rounded">
            <span className="text-gray-600 min-w-[70px] shrink-0">{log.time}</span>
            <span className={`font-semibold min-w-[80px] shrink-0 ${log.color}`}>[{log.type}]</span>
            <span className="text-gray-300 break-words">{log.msg}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
