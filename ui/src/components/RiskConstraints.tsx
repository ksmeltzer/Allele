import { useState, useEffect } from 'react';
import { useWebSocket } from '../contexts/WebSocketContext';
import SecurityIcon from '@mui/icons-material/Security';
import HelpModal from './HelpModal';

export default function RiskConstraints() {
  const { subscribe } = useWebSocket();
  const [constraints, setConstraints] = useState<any[]>([]);

  useEffect(() => {
    const unsub = subscribe('health', payload => {
      // Only show warnings or errors in the Risk Constraint pane
      if (payload.level === 'warning' || payload.level === 'error' || payload.type === 'violation') {
        setConstraints(prev => [{ timestamp: new Date(), data: payload }, ...prev].slice(0, 50));
      }
    });
    return unsub;
  }, [subscribe]);

  return (
    <div className="bg-[#1B2028] h-full flex flex-col overflow-hidden relative">
      <div className="absolute top-2 right-4 z-10 flex items-center bg-[#11141A] rounded px-2 py-0.5 border border-[#2B3139]">
        <HelpModal title="Risk Constraints (The Boundary)" iconColor="#A6B0C3" size="small">
          <p>
            If the Engine is the "Arena" where organisms act, the Risk Constraints represent the <strong>absolute boundaries</strong> of that environment.
          </p>
          <p className="mt-2">
            Even if a Strategy (Organism) generates a <span className="text-[#4F46E5] font-bold">Strategy Eval</span> that commands a massive trade, it will be intercepted and nullified here if it violates max position sizing, drawdown limits, or total allowed leverage. This pane alerts you to any active health, limit, or boundary violations.
          </p>
        </HelpModal>
      </div>

      <div className="flex-1 overflow-y-auto p-4 space-y-2 font-mono text-[11px] min-h-0 pt-10">
        {constraints.length === 0 ? (
          <div className="h-full flex flex-col items-center justify-center text-[#A6B0C3] opacity-50 space-y-4">
            <SecurityIcon sx={{ fontSize: 48 }} />
            <span className="uppercase tracking-widest text-xs">All Systems Healthy</span>
          </div>
        ) : (
          constraints.map((c, i) => {
            const isError = c.data.level === 'error' || c.data.type === 'violation';
            const colorHex = isError ? '#FB3836' : '#EAB308'; // Red for errors, Yellow for warnings
            return (
              <div key={i} className={`p-2 rounded border-l-2 bg-opacity-10 flex flex-col space-y-1`} 
                   style={{ borderColor: colorHex, backgroundColor: `${colorHex}1A`, color: colorHex }}>
                <div className="flex justify-between items-center opacity-80 text-[10px]">
                  <span className="font-bold uppercase">
                    {c.data.level === 'warning' ? 'BOUNDARY WARNING' : 'RISK VIOLATION'}
                  </span>
                  <span>{c.timestamp.toLocaleTimeString()}</span>
                </div>
                <div className="text-[#E2E8F0] break-all">
                  {c.data.message || JSON.stringify(c.data)}
                </div>
              </div>
            );
          })
        )}
      </div>
    </div>
  );
}
