import { useState, useEffect, useRef } from 'react';
import { useWebSocket } from './contexts/WebSocketContext';
import { Layout, Model, TabNode } from 'flexlayout-react';
import type { IJsonModel } from 'flexlayout-react';
import type { Manifest } from './types/plugin';
import 'flexlayout-react/style/dark.css';

// Material UI components
import IconButton from '@mui/material/IconButton';

// Material UI Icons
import SettingsIcon from '@mui/icons-material/Settings';
import PulseIcon from '@mui/icons-material/WifiTethering'; // Custom stand-in for pulse
import NotificationsIcon from '@mui/icons-material/Notifications';
import WarningIcon from '@mui/icons-material/WarningAmber';

import CausalityTrace from './components/CausalityTrace';
import PluginManager from './components/PluginManager';
import TradeSignals from './components/TradeSignals';
import RiskConstraints from './components/RiskConstraints';
import WalletInfo from './components/WalletInfo';
import HelpModal from './components/HelpModal';

const GlobalStatusIndicators = () => {
  const { subscribe, sendEvent, connected } = useWebSocket();
  const [plugins, setPlugins] = useState<Manifest[]>([]);
  const [open, setOpen] = useState(false);

  useEffect(() => {
    const unsubscribe = subscribe('manifests_updated', (payload: Manifest[]) => {
      setPlugins(payload || []);
    });

    if (connected) {
      sendEvent('request_manifests');
    }

    return () => unsubscribe();
  }, [connected, subscribe, sendEvent]);

  const pluginsNeedingConfig = plugins.filter(p => p.config?.some(c => c.required && (!c.value || c.value === '')));

  if (pluginsNeedingConfig.length === 0) return null;

  return (
    <div className="relative flex items-center mr-4">
      <IconButton 
        onClick={() => setOpen(!open)}
        sx={{ color: '#A6B0C3', '&:hover': { color: 'white' } }}
      >
        <div className="relative">
          <NotificationsIcon />
          <span className="absolute -top-1 -right-1 flex h-3 w-3">
            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-yellow-400 opacity-75"></span>
            <span className="relative inline-flex rounded-full h-3 w-3 bg-yellow-500 border border-[#0B0E11]"></span>
          </span>
        </div>
      </IconButton>

      {open && (
        <div className="absolute top-full right-0 mt-2 w-64 bg-[#1B2028] border border-[#2B3139] rounded shadow-xl z-50 overflow-hidden">
          <div className="px-4 py-2 bg-[#11141A] border-b border-[#2B3139] text-xs font-bold text-[#A6B0C3] uppercase tracking-wider">
            Action Required
          </div>
          <div className="max-h-64 overflow-y-auto">
            {pluginsNeedingConfig.map(p => (
              <button
                key={`alert-${p.name}`}
                onClick={() => {
                  document.dispatchEvent(new CustomEvent('open-plugin-config', { detail: p }));
                  setOpen(false);
                }}
                className="w-full flex items-start space-x-3 px-4 py-3 hover:bg-[#2B3139] transition-colors text-left"
              >
                <WarningIcon sx={{ color: '#EAB308', fontSize: 20 }} />
                <div>
                  <div className="text-sm font-medium text-white">{p.name}</div>
                  <div className="text-xs text-[#A6B0C3] mt-0.5">Needs configuration before it can run.</div>
                </div>
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};

const ComingSoonPanel = ({ title, IconComponent, helpText }: { title: string, IconComponent: any, helpText?: React.ReactNode }) => (
  <div className="h-full flex items-center justify-center bg-[#1B2028] text-[#A6B0C3] font-mono text-xs uppercase tracking-widest p-4 relative">
    {helpText && (
      <div className="absolute top-2 right-4 z-10 flex items-center bg-[#11141A] rounded px-2 py-0.5 border border-[#2B3139]">
        <HelpModal title={title} iconColor="#A6B0C3" size="small">
          {helpText}
        </HelpModal>
      </div>
    )}
    <div className="flex flex-col items-center space-y-4 opacity-50">
      <IconComponent sx={{ fontSize: 48 }} />
      <span>{title} - Coming Soon</span>
    </div>
  </div>
);

const FirehosePanel = () => {
  const { subscribe } = useWebSocket();
  const [logs, setLogs] = useState<any[]>([]);
  const logsEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const pushLog = (type: string, payload: any) => {
      setLogs(prev => [...prev.slice(-99), { type, timestamp: new Date(), ...payload }]);
    };
    
    const unsubs = [
      subscribe('tick', p => pushLog('tick', p)),
      subscribe('strategy_eval', p => pushLog('strategy_eval', p)),
      subscribe('health', p => pushLog('health', p)),
      subscribe('plugin_status', p => pushLog('plugin_status', p))
    ];

    return () => unsubs.forEach(fn => fn());
  }, [subscribe]);

  useEffect(() => {
    logsEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [logs]);

  return (
    <div className="bg-[#1B2028] h-full flex flex-col relative overflow-hidden">
      <div className="absolute top-2 right-4 z-10 flex items-center bg-[#11141A] rounded px-2 py-0.5 border border-[#2B3139]">
        <HelpModal title="System Telemetry" iconColor="#A6B0C3" size="small">
          <p>
            The raw central nervous system of the <strong>Allele Engine</strong>.
          </p>
          <p className="mt-2">
            This stream displays all unprocessed JSON payloads routing through the central event bus, including plugin health pings, market data (ticks), strategy decisions, and error logs.
          </p>
        </HelpModal>
      </div>

      <div className="flex-1 overflow-y-auto p-4 space-y-2 font-mono text-[11px] leading-relaxed min-h-0 pt-10">
        {logs.length === 0 ? (
          <div className="text-[#A6B0C3] flex items-center space-x-2">
            <PulseIcon fontSize="small" sx={{ animation: 'pulse 2s infinite' }} />
            <span className="uppercase tracking-widest">Awaiting telemetry...</span>
          </div>
        ) : (
          logs.map((log, i) => (
            <div key={i} className="pl-3 border-l-2 py-1 border-[#2B3139] hover:bg-[#2B3139]/30 text-[#E2E8F0] transition-colors rounded-r">
              <span className="text-[#5B616E] mr-3">[{log.timestamp.toLocaleTimeString()}]</span>
              <span className={`font-bold mr-2 ${log.type === 'tick' ? 'text-[#00C087]' : log.type === 'health' ? 'text-[#4F46E5]' : 'text-[#FB3836]'}`}>
                [{log.type.toUpperCase()}]
              </span>
              <span className="break-all opacity-80">{JSON.stringify(log, (k,v) => k === 'type' || k === 'timestamp' ? undefined : v)}</span>
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
        type: "column",
        weight: 75,
        children: [
          {
            type: "tabset",
            weight: 70,
            children: [
              { type: "tab", name: "Trade Signals", component: "trades" },
              { type: "tab", name: "Risk Constraints", component: "risk" },
              { type: "tab", name: "Causality Trace", component: "causality" }
            ]
          },
          {
            type: "tabset",
            weight: 30,
            children: [
              { type: "tab", name: "System Log", component: "firehose" }
            ]
          }
        ]
      },
      {
        type: "column",
        weight: 25,
        children: [
          {
            type: "tabset",
            weight: 33,
            children: [
              { type: "tab", name: "Exchanges", component: "plugins_exchanges" }
            ]
          },
          {
            type: "tabset",
            weight: 33,
            children: [
              { type: "tab", name: "Strategies", component: "plugins_strategies" }
            ]
          },
          {
            type: "tabset",
            weight: 34,
            children: [
              { type: "tab", name: "Data & Core", component: "plugins_core" }
            ]
          }
        ]
      }
    ]
  }
};

function App() {
  const [model, setModel] = useState<Model>(() => {
    const savedLayout = localStorage.getItem('allele-layout-v4');
    if (savedLayout) {
      try { return Model.fromJson(JSON.parse(savedLayout)); } 
      catch (e) { return Model.fromJson(DEFAULT_LAYOUT); }
    }
    return Model.fromJson(DEFAULT_LAYOUT);
  });

  const onModelChange = (newModel: Model) => {
    setModel(newModel);
    localStorage.setItem('allele-layout-v4', JSON.stringify(newModel.toJson()));
  };

  const factory = (node: TabNode) => {
    const component = node.getComponent();
    switch (component) {
      case "firehose": return <FirehosePanel />;
      case "plugins_exchanges": return <PluginManager allowedCategories={['Exchanges & Markets']} showInstallBar={false} />;
      case "plugins_strategies": return <PluginManager allowedCategories={['Trading Strategies']} showInstallBar={false} />;
      case "plugins_core": return <PluginManager allowedCategories={['Core & Utilities', 'Data Sensors', 'Risk & Portfolio']} showInstallBar={true} />;
      case "trades": return <TradeSignals />;
      case "risk": return <RiskConstraints />;
      case "causality": return <CausalityTrace />;
      default: return <ComingSoonPanel title={component || "Unknown"} IconComponent={SettingsIcon} />;
    }
  };

  return (
    <div className="h-screen w-screen bg-[#0B0E11] text-[#E2E8F0] font-sans flex flex-col overflow-hidden">
      {/* Global Header (Coinbase Advanced Style) */}
      <header className="bg-[#0B0E11] border-b border-[#2B3139] px-6 py-3 flex items-center justify-between shrink-0 z-50">
        <div className="flex items-center space-x-6">
          <div className="flex items-center space-x-3">
            <PulseIcon sx={{ color: '#4F46E5' }} />
            <h1 className="text-[15px] font-bold tracking-wide text-white">Allele <span className="font-normal text-[#A6B0C3]">Advanced</span></h1>
          </div>
          
          <nav className="hidden md:flex items-center space-x-6 text-[13px] font-medium text-[#A6B0C3]">
            <a href="#" className="text-white border-b-2 border-[#4F46E5] pb-1 hover:text-white transition-colors">Trade Engine</a>
            <a href="#" className="hover:text-white transition-colors pb-1 border-b-2 border-transparent">
              Portfolios
              <HelpModal title="Portfolios (Coming Soon)" iconColor="#A6B0C3" size="small">
                <p>
                  The Portfolios view tracks positions that plugins (Organisms) have successfully entered.
                </p>
                <p className="mt-2">
                  Once an organism expresses an Action (buy/sell) via a <span className="text-[#4F46E5] font-bold">Strategy Eval</span> event, the trade is sent to the target exchange. The resulting real-world holdings are monitored here.
                </p>
                <p className="mt-2 text-yellow-500 italic">
                  Note: The portfolio data integration is currently being mapped to the Wallet RPC providers.
                </p>
              </HelpModal>
            </a>
            <a href="#" className="hover:text-white transition-colors pb-1 border-b-2 border-transparent">Risk Limits</a>
          </nav>
        </div>
        
        <div className="flex items-center space-x-4">
          <WalletInfo />
          <div id="global-status-indicators" className="flex items-center mr-2">
            <GlobalStatusIndicators />
          </div>
          <div className="text-[12px] font-medium text-[#A6B0C3] flex items-center px-3 py-1.5 bg-[#1B2028] rounded border border-[#2B3139]">
            <span className="w-2 h-2 bg-[#00C087] rounded-full mr-2 shadow-[0_0_8px_rgba(0,192,135,0.6)]"></span> 
            Engine Connected
          </div>
          <IconButton size="small" sx={{ color: '#A6B0C3', '&:hover': { color: 'white', backgroundColor: '#1B2028' } }}>
            <SettingsIcon fontSize="small" />
          </IconButton>
        </div>
      </header>

      {/* FlexLayout Workspace */}
      <div className="flex-1 relative isolate layout-override">
        <Layout 
          model={model} 
          factory={factory} 
          onModelChange={onModelChange}
          realtimeResize={true}
        />
      </div>
    </div>
  );
}

export default App;
