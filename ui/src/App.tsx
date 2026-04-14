import { useState, useEffect, useRef } from 'react';
import { useWebSocket } from './contexts/WebSocketContext';
import { Layout, Model, TabNode } from 'flexlayout-react';
import type { IJsonModel } from 'flexlayout-react';
import 'flexlayout-react/style/dark.css'; // Using the standard dark theme
import CausalityTrace from './components/CausalityTrace';
import PluginManager from './components/PluginManager';

// Temporary dummy component for the Order Book / other unbuilt panels
const ComingSoonPanel = ({ title }: { title: string }) => (
  <div className="h-full flex items-center justify-center bg-[#050505] text-gray-500 font-mono text-xs uppercase tracking-widest border border-[#1f2937] p-4">
    [{title} - Development Pending]
  </div>
);

const FirehosePanel = () => {
  const { subscribe } = useWebSocket();
  const [logs, setLogs] = useState<any[]>([]);
  const logsEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const unsubscribeTick = subscribe('tick', (payload) => {
      setLogs(prev => [...prev.slice(-99), { type: 'tick', ...payload }]);
    });
    const unsubscribeEval = subscribe('strategy_eval', (payload) => {
      setLogs(prev => [...prev.slice(-99), { type: 'strategy_eval', ...payload }]);
    });
    const unsubscribeHealth = subscribe('health', (payload) => {
      setLogs(prev => [...prev.slice(-99), { type: 'health', ...payload }]);
    });
    const unsubscribePlugin = subscribe('plugin_status', (payload) => {
      setLogs(prev => [...prev.slice(-99), { type: 'plugin_status', ...payload }]);
    });

    return () => {
      unsubscribeTick();
      unsubscribeEval();
      unsubscribeHealth();
      unsubscribePlugin();
    };
  }, [subscribe]);

  useEffect(() => {
    logsEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [logs]);

  return (
    <div className="bg-[#050505] h-full flex flex-col relative overflow-hidden border border-[#1f2937]">
      <div className="flex-1 overflow-y-auto p-4 space-y-2 font-mono text-[11px] leading-relaxed min-h-0">
        {logs.length === 0 ? (
          <div className="flex flex-col space-y-4 mt-2">
            <div className="flex items-center space-x-3 text-gray-600">
              <p className="uppercase tracking-widest">Waiting for events...</p>
            </div>
          </div>
        ) : (
          logs.map((log, i) => (
            <div 
              key={i} 
              className={`pl-2 border-l-2 py-1 border-gray-700 text-gray-300`}
            >
              <span className="text-gray-500 mr-3">[{new Date().toLocaleTimeString()}]</span>
              <span className="font-bold text-gray-400 mr-2">[{log.type}]</span>
              <span className="break-all">{JSON.stringify(log)}</span>
            </div>
          ))
        )}
        <div ref={logsEndRef} />
      </div>
    </div>
  );
};

const DEFAULT_LAYOUT: IJsonModel = {
  global: {
    tabEnableClose: false,
    splitterSize: 4,
  },
  borders: [],
  layout: {
    type: "row",
    weight: 100,
    children: [
      {
        type: "tabset",
        weight: 50,
        children: [
          {
            type: "tab",
            name: "Firehose / Logs",
            component: "firehose"
          }
        ]
      },
      {
        type: "tabset",
        weight: 25,
        children: [
          {
            type: "tab",
            name: "Causality Trace",
            component: "causality"
          }
        ]
      },
      {
        type: "tabset",
        weight: 25,
        children: [
          {
            type: "tab",
            name: "Plugin Vault",
            component: "plugins"
          }
        ]
      }
    ]
  }
};

function App() {
  const { connected } = useWebSocket();
  const [model] = useState<Model>(() => {
    const saved = localStorage.getItem('allele-layout');
    if (saved) {
      try {
        return Model.fromJson(JSON.parse(saved));
      } catch (e) {
        console.error("Failed to parse saved layout", e);
      }
    }
    return Model.fromJson(DEFAULT_LAYOUT);
  });

  // Save layout changes to localStorage
  const handleModelChange = () => {
    localStorage.setItem('allele-layout', JSON.stringify(model.toJson()));
  };

  const factory = (node: TabNode) => {
    const component = node.getComponent();
    switch (component) {
      case "firehose":
        return <FirehosePanel />;
      case "causality":
        return <CausalityTrace sideliningReasons={[]} />;
      case "plugins":
        return <PluginManager />;
      default:
        return <ComingSoonPanel title={component || 'Unknown'} />;
    }
  };

  return (
    <div className="bg-[#0a0a0c] text-gray-300 h-screen font-mono flex flex-col">
      <header className="shrink-0 flex items-center justify-between border-b border-[#1f2937] px-4 py-2 bg-[#050505]">
        <div className="flex items-center space-x-3">
          <div>
            <h1 className="text-sm font-bold text-gray-200 uppercase tracking-widest">Allele Terminal</h1>
          </div>
        </div>
        
        <div className="flex items-center space-x-2">
          <div className={`w-2 h-2 rounded-full ${connected ? 'bg-green-500' : 'bg-red-500'}`}></div>
          <span className={`text-[10px] uppercase tracking-wider font-bold ${connected ? 'text-green-500' : 'text-red-500'}`}>
            {connected ? 'Connected' : 'Offline'}
          </span>
        </div>
      </header>

      <div className="flex-1 min-h-0 relative flexlayout-theme-overrides">
        <Layout 
          model={model} 
          factory={factory} 
          onModelChange={handleModelChange}
        />
      </div>
    </div>
  );
}

export default App;
