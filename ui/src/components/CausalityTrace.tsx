import React from 'react';

type SideliningReason = {
  reason: string;
};

type CausalityTraceProps = {
  sideliningReasons: SideliningReason[];
};

const CausalityTrace: React.FC<CausalityTraceProps> = ({sideliningReasons}) => {
  return (
    <div className="bg-[#050505] h-full flex flex-col overflow-hidden border border-[#1f2937]">
      <div className="flex-1 overflow-y-auto p-4 space-y-3 min-h-0">
        {sideliningReasons.length === 0 ? (
          <div className="h-full flex flex-col items-center justify-center text-gray-600 opacity-50 space-y-4">
            <svg className="w-12 h-12" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
            </svg>
            <p className="text-[10px] uppercase tracking-[0.2em] text-center px-4">All Active Organisms Healthy<br/>No Sidelining Events Detected</p>
          </div>
        ) : (
          sideliningReasons.map((reason, index) => (
            <div key={index} className="p-3 rounded bg-red-950/20 border border-red-900/30 flex items-start space-x-3">
              <span className="text-red-500 mt-0.5">⚠️</span>
              <p className="text-gray-400 text-xs font-mono leading-relaxed">{reason.reason}</p>
            </div>
          ))
        )}
      </div>
    </div>
  );
};

export default CausalityTrace;