import { useState, useEffect } from 'react';
import { useWebSocket } from '../contexts/WebSocketContext';
import InsightsIcon from '@mui/icons-material/Insights';
import HelpModal from './HelpModal';

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

  const renderTick = (tick: any) => {
    const displayName = tick.AssetName || tick.AssetID;
    return (
      <div className="flex flex-col">
        <span className="font-semibold text-white truncate" title={displayName}>
          {displayName}
        </span>
        <div className="mt-1 flex justify-between opacity-80">
          <span className={tick.IsBid ? 'text-[#00C087]' : 'text-[#F6465D]'}>
            {tick.IsBid ? 'BID' : 'ASK'}
          </span>
          <span>
            {tick.Price?.toFixed(4)} ¢ <span className="opacity-50">x</span> {tick.Size?.toFixed(0)}
          </span>
        </div>
      </div>
    );
  };

  const renderEval = (evalData: any) => {
    const { strategy_id, actions } = evalData;
    return (
      <div className="flex flex-col space-y-1">
        <div className="font-semibold text-[#4F46E5] truncate">
          Agent: {strategy_id}
        </div>
        {actions && actions.length > 0 ? (
          <div className="space-y-1 mt-1 border-t border-[#4F46E5]/20 pt-1">
            {actions.map((act: any, idx: number) => {
              const displayName = act.AssetName || act.AssetID;
              return (
                <div key={idx} className="flex justify-between text-[#E2E8F0] opacity-90">
                  <span className="truncate pr-2" title={displayName}>
                    <span className={act.Side === 'BUY' ? 'text-[#00C087] font-bold mr-1' : 'text-[#F6465D] font-bold mr-1'}>
                      {act.Side}
                    </span>
                    {displayName}
                  </span>
                  <span className="whitespace-nowrap">
                    {act.Price?.toFixed(4)} ¢
                  </span>
                </div>
              );
            })}
          </div>
        ) : (
          <span className="text-[#A6B0C3] italic">No actions generated</span>
        )}
      </div>
    );
  };

  return (
    <div className="bg-[#1B2028] h-full flex flex-col overflow-hidden relative">
      <div className="absolute top-2 right-4 z-10 flex items-center bg-[#11141A] rounded px-2 py-0.5 border border-[#2B3139]">
        <HelpModal title="Trade Signals & Traits" iconColor="#A6B0C3" size="small">
          <p>
            In the <strong>Allele metaphor</strong>, this stream represents how organisms react to their environment.
          </p>
          <ul className="list-disc pl-4 space-y-2 mt-2">
            <li><span className="text-[#00C087] font-bold">Market Ticks</span>: The raw environmental stimuli (price movements and book updates) from prediction markets.</li>
            <li><span className="text-[#4F46E5] font-bold">Strategy Evals</span>: This is the critical indicator! It shows when an organism (a WASM plugin) successfully found what it was looking for in the data and <em>expressed a phenotypic trait</em> by generating an Action (like a Buy/Sell order).</li>
          </ul>
        </HelpModal>
      </div>

      <div className="flex-1 overflow-y-auto p-4 space-y-2 font-mono text-[11px] min-h-0 pt-10">
        {signals.length === 0 ? (
          <div className="h-full flex flex-col items-center justify-center text-[#A6B0C3] opacity-50 space-y-4">
            <InsightsIcon sx={{ fontSize: 48 }} />
            <span className="uppercase tracking-widest text-xs">Awaiting Trade Signals...</span>
          </div>
        ) : (
          signals.map((sig, i) => (
            <div key={i} className={`p-2 rounded border-l-2 flex flex-col space-y-1 ${
              sig.type === 'tick' ? 'bg-[#00C087]/10 border-[#00C087]' : 'bg-[#4F46E5]/10 border-[#4F46E5]'
            }`}>
              <div className="flex justify-between items-center opacity-80 text-[10px] mb-1">
                <span className="font-bold uppercase" style={{ color: sig.type === 'tick' ? '#00C087' : '#4F46E5' }}>
                  {sig.type === 'tick' ? 'Market Tick' : 'Strategy Eval'}
                </span>
                <span className="text-[#A6B0C3]">{sig.timestamp.toLocaleTimeString()}</span>
              </div>
              <div className="text-[#E2E8F0]">
                {sig.type === 'tick' ? renderTick(sig.data) : renderEval(sig.data)}
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
}
