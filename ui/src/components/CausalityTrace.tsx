import { useState, useEffect } from 'react';
import { useWebSocket } from '../contexts/WebSocketContext';
import AccountTreeIcon from '@mui/icons-material/AccountTree';

export default function CausalityTrace() {
  const { subscribe } = useWebSocket();
  const [events, setEvents] = useState<any[]>([]);

  useEffect(() => {
    const unsub = subscribe('strategy_eval', payload => {
      setEvents(prev => [{ timestamp: new Date(), data: payload }, ...prev].slice(0, 20));
    });
    return unsub;
  }, [subscribe]);

  return (
    <div className="bg-[#1B2028] h-full flex flex-col overflow-hidden border-t border-[#2B3139]">
      <div className="flex-1 overflow-y-auto p-4 space-y-3 min-h-0 font-mono text-[11px]">
        {events.length === 0 ? (
          <div className="h-full flex flex-col items-center justify-center text-[#A6B0C3] opacity-50 space-y-4">
            <AccountTreeIcon sx={{ fontSize: 48 }} />
            <p className="uppercase tracking-widest text-xs text-center px-4">Awaiting Reasoning Trace...</p>
          </div>
        ) : (
          events.map((ev, index) => (
            <div key={index} className="p-3 rounded bg-[#2B3139]/30 border-l-2 border-[#4F46E5] flex flex-col space-y-2">
              <div className="text-[#A6B0C3] text-[10px] uppercase font-bold tracking-wider flex justify-between">
                <span>EVALUATION DAG</span>
                <span>{ev.timestamp.toLocaleTimeString()}</span>
              </div>
              <pre className="text-[#E2E8F0] overflow-x-auto whitespace-pre-wrap">
                {JSON.stringify(ev.data, (key, value) => {
                  // Keep it compact but show the new AssetName if it exists
                  if (key === 'AssetID' && ev.data.AssetName) {
                    return `${ev.data.AssetName} (${value})`;
                  }
                  return value;
                }, 2)}
              </pre>
            </div>
          ))
        )}
      </div>
    </div>
  );
}
