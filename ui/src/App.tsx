import { useState, useEffect } from 'react';
import CausalityTrace from './components/CausalityTrace';
import PluginManager from './components/PluginManager';

function App() {
  const [logs, setLogs] = useState<any[]>([]);
  const [connected, setConnected] = useState(false);

  useEffect(() => {
    const ws = new WebSocket(`ws://${window.location.hostname}:8081/ws`);

    ws.onopen = () => {
      setConnected(true);
    };

    ws.onclose = () => {
      setConnected(false);
    };

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        setLogs((prev) => {
          const newLogs = [data, ...prev];
          return newLogs.slice(0, 50); // Keep last 50
        });
      } catch (e) {
        console.error('Failed to parse WebSocket message:', e);
      }
    };

    return () => {
      ws.close();
    };
  }, []);

  return (
    <div className="bg-gray-900 text-white min-h-screen p-8 font-mono">
      <header className="mb-8 border-b border-gray-700 pb-4">
        <h1 className="text-3xl font-bold text-purple-400">🧬 .allele Strategy Arena</h1>
      </header>

      <div className="flex flex-col lg:flex-row gap-8">
        {/* Left Side: Allele Radar */}
        <div className="flex-1 bg-gray-800 rounded-lg p-6 shadow-lg border border-gray-700 h-[80vh] overflow-hidden flex flex-col">
          <h2 className="text-xl font-semibold mb-4 text-gray-300">Radar Logs</h2>
          <div className="flex-1 overflow-y-auto space-y-2 pr-2">
            {logs.length === 0 ? (
              <p className="text-gray-500 italic">Waiting for signals...</p>
            ) : (
              logs.map((log, i) => (
                <div 
                  key={i} 
                  className={`p-3 rounded bg-gray-900 border ${log.type === 'arbitrage' ? 'border-green-500 animate-pulse' : 'border-gray-700'}`}
                >
                  <pre className="text-xs whitespace-pre-wrap break-words">
                    {JSON.stringify(log, null, 2)}
                  </pre>
                </div>
              ))
            )}
          </div>
        </div>

        {/* Center: Causality Trace */}
        <CausalityTrace sideliningReasons={[]} />

        {/* Right Side: System Health & Plugins */}
        <div className="w-full lg:w-[400px] flex flex-col space-y-8">
          <div className="bg-gray-800 rounded-lg p-6 shadow-lg border border-gray-700 h-fit">
            <h2 className="text-xl font-semibold mb-4 text-gray-300">System Health</h2>
            
            <div className="flex items-center space-x-3 mb-6 p-4 bg-gray-900 rounded border border-gray-700">
              <div className="flex-1">
                <span className="text-sm text-gray-400 block">WebSocket Status</span>
                <span className={`font-semibold ${connected ? 'text-green-400' : 'text-red-400'}`}>
                  {connected ? 'Connected' : 'Disconnected'}
                </span>
              </div>
              <div className={`w-3 h-3 rounded-full ${connected ? 'bg-green-500 animate-pulse' : 'bg-red-500'}`}></div>
            </div>
          </div>

          {/* Plugin Manager injected here */}
          <PluginManager />
        </div>
      </div>
    </div>
  );
}

export default App;
