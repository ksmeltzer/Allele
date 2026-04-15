import { useState, useEffect } from 'react';
import { useWebSocket } from '../contexts/WebSocketContext';
import SecurityIcon from '@mui/icons-material/Security';

export default function RiskConstraints() {
  const { subscribe } = useWebSocket();
  const [constraints, setConstraints] = useState<any[]>([]);

  useEffect(() => {
    const unsub = subscribe('health', payload => {
      setConstraints(prev => [{ timestamp: new Date(), data: payload }, ...prev].slice(0, 50));
    });
    return unsub;
  }, [subscribe]);

  return (
    <div className="bg-[#1B2028] h-full flex flex-col overflow-hidden relative">
      <div className="flex-1 overflow-y-auto p-4 space-y-2 font-mono text-[11px] min-h-0">
        {constraints.length === 0 ? (
          <div className="h-full flex flex-col items-center justify-center text-[#A6B0C3] opacity-50 space-y-4">
            <SecurityIcon sx={{ fontSize: 48 }} />
            <span className="uppercase tracking-widest text-xs">All Systems Healthy</span>
          </div>
        ) : (
          constraints.map((c, i) => (
            <div key={i} className="p-2 rounded border-l-2 border-[#FB3836] bg-[#FB3836]/10 text-[#FB3836] flex flex-col space-y-1">
              <div className="flex justify-between items-center opacity-80 text-[10px]">
                <span className="font-bold uppercase">HEALTH WARNING</span>
                <span>{c.timestamp.toLocaleTimeString()}</span>
              </div>
              <div className="text-[#E2E8F0] break-all">
                {JSON.stringify(c.data)}
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
}
