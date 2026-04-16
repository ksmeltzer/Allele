import { useState, useEffect } from 'react';
import { useWebSocket } from '../contexts/WebSocketContext';
import AccountBalanceWalletIcon from '@mui/icons-material/AccountBalanceWallet';
import HelpModal from './HelpModal';

export default function PortfolioView() {
  const { subscribe } = useWebSocket();
  const [positions, setPositions] = useState<any[]>([]);

  useEffect(() => {
    const unsub = subscribe('portfolio_update', payload => {
      setPositions(payload || []);
    });
    return unsub;
  }, [subscribe]);

  return (
    <div className="bg-[#1B2028] h-full flex flex-col overflow-hidden relative">
      <div className="absolute top-2 right-4 z-10 flex items-center bg-[#11141A] rounded px-2 py-0.5 border border-[#2B3139]">
        <HelpModal title="Organism Portfolio" iconColor="#A6B0C3" size="small">
          <p>
            When an organism successfully expresses a trait (triggers a trade) and the trade executes on the Exchange, the resulting positions are tracked here.
          </p>
          <p className="mt-2">
            This represents the collective holdings accumulated by the system's evolutionary agents.
          </p>
        </HelpModal>
      </div>

      <div className="flex-1 overflow-y-auto p-4 pt-10">
        {positions.length === 0 ? (
          <div className="h-full flex flex-col items-center justify-center text-[#A6B0C3] opacity-50 space-y-4">
            <AccountBalanceWalletIcon sx={{ fontSize: 48 }} />
            <span className="uppercase tracking-widest text-xs font-mono">No Active Positions</span>
          </div>
        ) : (
          <table className="w-full text-left font-mono text-[11px] text-[#E2E8F0]">
            <thead>
              <tr className="text-[#A6B0C3] border-b border-[#2B3139]">
                <th className="pb-2 font-medium tracking-wider">ASSET</th>
                <th className="pb-2 font-medium tracking-wider text-right">SIZE</th>
                <th className="pb-2 font-medium tracking-wider text-right">AVG ENTRY (¢)</th>
                <th className="pb-2 font-medium tracking-wider text-right">TOTAL COST ($)</th>
              </tr>
            </thead>
            <tbody>
              {positions.map((pos, idx) => {
                const displayName = pos.asset_name || pos.asset_id;
                const totalCost = (pos.size * pos.avg_price).toFixed(2);
                
                return (
                  <tr key={idx} className="border-b border-[#2B3139]/50 hover:bg-[#2B3139]/30 transition-colors">
                    <td className="py-3 font-semibold truncate max-w-[200px]" title={displayName}>
                      {displayName}
                    </td>
                    <td className="py-3 text-right text-[#00C087] font-bold">
                      {pos.size.toLocaleString()}
                    </td>
                    <td className="py-3 text-right">
                      {pos.avg_price.toFixed(4)} ¢
                    </td>
                    <td className="py-3 text-right text-[#A6B0C3]">
                      ${totalCost}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}