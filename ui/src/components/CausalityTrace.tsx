import React from 'react';

type SideliningReason = {
  reason: string;
};

type CausalityTraceProps = {
  sideliningReasons: SideliningReason[];
};

const CausalityTrace: React.FC<CausalityTraceProps> = ({sideliningReasons}) => {
  return (
    <div className="bg-gray-800 rounded-lg p-6 shadow-lg border border-gray-700 h-[80vh] overflow-auto">
      <h2 className="text-xl font-semibold mb-4 text-gray-300">Causality Trace</h2>
      <div className="space-y-2">
        {sideliningReasons.length === 0 ? (
          <p className="text-gray-500 italic">No sidelining reasons to display...</p>
        ) : (
          sideliningReasons.map((reason, index) => (
            <div key={index} className="p-3 rounded bg-gray-900 border border-gray-700">
              <p className="text-gray-300 text-sm">{reason.reason}</p>
            </div>
          ))
        )}
      </div>
    </div>
  );
};

export default CausalityTrace;