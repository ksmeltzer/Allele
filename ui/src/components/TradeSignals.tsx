import { useState, useEffect } from 'react';
import { useWebSocket } from '../contexts/WebSocketContext';
import InsightsIcon from '@mui/icons-material/Insights';

export default function TradeSignals() {
  const { subscribe } = useWebSocket();
  const [signals, setSignals] = useState<any[]>([]);

  useEffect(() => {
    const unsubs = [
      subscribe('tick', payload => {
        setSignals(prev => [{ type: 'tick', timestamp: new Date(), data: payload }, ...prev].slice(0, 50));
      }),
      subscribe('strategy_eval', payload => {
        setSignals(prev => [{ type: 'eval', timestamp: new Date(), data: payload }, ...prev].slice(0, 50));
      })
    ];
    return () => unsubs.forEach(fn => fn());
  }, [subscribe]);

  return (
    <div className="bg-[#1B2028] h-full flex flex-col overflow-hidden relative">
      <div className="flex-1 overflow-y-auto p-4 space-y-2 font-mono text-[11px] min-h-0">
        {signals.length === 0 ? (
          <div className="h-full flex flex-col items-center justify-center text-[#A6B0C3] opacity-50 space-y-4">
            <InsightsIcon sx={{ fontSize: 48 }} />
            <span className="uppercase tracking-widest text-xs">Awaiting Trade Signals...</span>
          </div>
        ) : (
          signals.map((sig, i) => (
            <div key={i} className={`p-2 rounded border-l-2 flex flex-col space-y-1 ${
              sig.type === 'tick' ? 'bg-[#00C087]/10 border-[#00C087] text-[#00C087]' : 'bg-[#0052FF]/10 border-[#0052FF] text-[#0052FF]'
            }`}>
              <div className="flex justify-between items-center opacity-80 text-[10px]">
                <span className="font-bold uppercase">{sig.type}</span>
                <span>{sig.timestamp.toLocaleTimeString()}</span>
              </div>
              <div className="text-[#E2E8F0] break-all">
                {JSON.stringify(sig.data)}
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
}
