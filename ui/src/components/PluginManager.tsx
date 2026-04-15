import { useState, useEffect, useMemo } from 'react';
import { createPortal } from 'react-dom';
import { useWebSocket } from '../contexts/WebSocketContext';
import type { Manifest, ConfigField } from '../types/plugin';

// Material UI Icons
import SettingsIcon from '@mui/icons-material/Settings';
import CloseIcon from '@mui/icons-material/Close';
import InfoIcon from '@mui/icons-material/Info';
import PulseIcon from '@mui/icons-material/WifiTethering';
import WarningIcon from '@mui/icons-material/WarningAmber';
import DownloadIcon from '@mui/icons-material/FileDownload';

function ConfigFieldRow({ 
  pluginName, 
  field, 
  onSubmit 
}: { 
  pluginName: string, 
  field: ConfigField, 
  onSubmit: (pluginName: string, key: string, value: string) => void 
}) {
  const [localValue, setLocalValue] = useState(field.value || '');
  const needsConfig = field.required && (!field.value || field.value === '');

  return (
    <div className="flex flex-col space-y-1.5 mb-4">
      <label className="text-[12px] font-medium text-[#A6B0C3] flex items-center">
        {field.key} {field.required && <span className="text-[#FB3836] ml-1">*</span>}
        <span className="ml-2 font-normal text-[#5B616E]">- {field.description}</span>
      </label>
      <div className="flex space-x-2">
        <input
          type={field.type === 'secret' ? 'password' : 'text'}
          value={localValue}
          onChange={(e) => setLocalValue(e.target.value)}
          placeholder={field.type === 'secret' && field.value === '********' ? '********' : 'Enter value...'}
          className={`flex-1 bg-[#0B0E11] border ${needsConfig ? 'border-yellow-600 focus:border-yellow-500' : 'border-[#2B3139] focus:border-[#4F46E5]'} rounded px-3 py-1.5 text-sm text-white focus:outline-none focus:ring-1 focus:ring-[#4F46E5]/50 transition-all font-mono`}
        />
        <button 
          onClick={() => onSubmit(pluginName, field.key, localValue)}
          disabled={localValue === field.value || localValue === ''}
          className="px-4 py-1.5 bg-[#4F46E5] hover:bg-[#4338CA] disabled:bg-[#2B3139] disabled:text-[#5B616E] text-white text-sm font-medium rounded transition-all duration-200"
        >
          Save
        </button>
      </div>
    </div>
  );
}

function ConfigModal({ 
  plugin, 
  onClose, 
  onSubmit 
}: { 
  plugin: Manifest, 
  onClose: () => void, 
  onSubmit: (pluginName: string, key: string, value: string) => void 
}) {
  return createPortal(
    <div className="fixed inset-0 z-[9999] flex items-center justify-center bg-black/80 backdrop-blur-sm p-4">
      <div className="bg-[#1B2028] border border-[#2B3139] rounded-xl shadow-2xl max-w-2xl w-full max-h-[90vh] flex flex-col overflow-hidden animate-in fade-in zoom-in-95 duration-200">
        <div className="px-6 py-4 border-b border-[#2B3139] flex items-center justify-between bg-[#1B2028]">
          <div className="flex items-center space-x-3">
            <SettingsIcon sx={{ color: '#A6B0C3' }} fontSize="small" />
            <div>
              <h3 className="text-[16px] font-bold text-white">{plugin.name} Settings</h3>
              <p className="text-[12px] text-[#A6B0C3]">{plugin.description}</p>
            </div>
          </div>
          <button 
            onClick={onClose}
            className="text-[#A6B0C3] hover:text-white transition-colors p-1 rounded-full hover:bg-[#2B3139] flex items-center justify-center"
          >
            <CloseIcon fontSize="small" />
          </button>
        </div>
        <div className="p-6 overflow-y-auto bg-[#11141A]">
          {plugin.config && plugin.config.length > 0 ? (
            <div className="space-y-5">
              {plugin.config.map(field => (
                <ConfigFieldRow 
                  key={field.key} 
                  pluginName={plugin.name} 
                  field={field} 
                  onSubmit={onSubmit} 
                />
              ))}
            </div>
          ) : (
            <p className="text-sm text-[#A6B0C3] italic flex items-center">
              <InfoIcon fontSize="small" sx={{ mr: 1 }} /> This plugin does not require any configuration.
            </p>
          )}
        </div>
        <div className="px-6 py-4 border-t border-[#2B3139] bg-[#1B2028] flex justify-end">
          <button 
            onClick={onClose}
            className="px-6 py-2 bg-[#2B3139] hover:bg-[#3A414A] text-white text-sm font-medium rounded transition-all duration-200"
          >
            Done
          </button>
        </div>
      </div>
    </div>,
    document.body
  );
}

// Helper to categorize plugins by their name pattern
function getPluginCategory(name: string): string {
  if (name.includes('-exchange-') || name.includes('-market-')) return 'Exchanges & Markets';
  if (name.includes('-strategy-')) return 'Trading Strategies';
  if (name.includes('-sensor-')) return 'Data Sensors';
  if (name.includes('-risk-') || name.includes('-portfolio-')) return 'Risk & Portfolio';
  return 'Core & Utilities';
}

// Ordered categories for display
const CATEGORY_ORDER = [
  'Exchanges & Markets',
  'Data Sensors',
  'Trading Strategies',
  'Risk & Portfolio',
  'Core & Utilities'
];

interface PluginManagerProps {
  title?: string;
  allowedCategories?: string[];
  showInstallBar?: boolean;
}

export default function PluginManager({ 
  allowedCategories, 
  showInstallBar = true 
}: PluginManagerProps) {
  const { connected, subscribe, sendEvent } = useWebSocket();
  const [plugins, setPlugins] = useState<Manifest[]>([]);
  const [loading, setLoading] = useState(true);
  const [installUri, setInstallUri] = useState('');
  const [configuringPlugin, setConfiguringPlugin] = useState<Manifest | null>(null);

  useEffect(() => {
    const unsubscribe = subscribe('manifests_updated', (payload: Manifest[]) => {
      setPlugins(payload || []);
      setLoading(false);
      setConfiguringPlugin(prev => {
        if (!prev) return null;
        const updated = (payload || []).find(p => p.name === prev.name);
        return updated || prev;
      });
    });

    if (connected) {
      sendEvent('request_manifests');
    }

    return () => unsubscribe();
  }, [connected, subscribe, sendEvent]);

  const handleConfigSubmit = (pluginName: string, key: string, value: string) => {
    if (value === '********') return;
    sendEvent('update_config', { plugin_name: pluginName, key, value });
  };

  const handleInstall = (e: React.FormEvent) => {
    e.preventDefault();
    if (!installUri.trim()) return;
    sendEvent('install_plugin', { uri: installUri.trim() });
    setInstallUri('');
  };

  const groupedPlugins = useMemo(() => {
    const groups: Record<string, Manifest[]> = {};
    plugins.forEach(p => {
      const cat = getPluginCategory(p.name);
      if (!groups[cat]) groups[cat] = [];
      groups[cat].push(p);
    });
    return groups;
  }, [plugins]);

  if (loading || !connected) {
    return (
      <div className="bg-[#1B2028] p-6 text-[#A6B0C3] text-sm flex flex-col items-center justify-center h-full">
        <PulseIcon sx={{ fontSize: 32, mb: 1, animation: 'pulse 2s infinite' }} />
        <span className="font-medium">{connected ? 'Loading plugins...' : 'Connecting to engine...'}</span>
      </div>
    );
  }

  const pluginNames = plugins.map(p => p.name);
  const pluginsNeedingConfig = plugins.filter(p => p.config?.some(c => c.required && (!c.value || c.value === '')));

  return (
    <div className="bg-[#1B2028] h-full flex flex-col relative overflow-hidden">
      {configuringPlugin && (
        <ConfigModal plugin={configuringPlugin} onClose={() => setConfiguringPlugin(null)} onSubmit={handleConfigSubmit} />
      )}

      {pluginsNeedingConfig.length > 0 && createPortal(
        <div className="flex items-center space-x-2 mr-4">
          {pluginsNeedingConfig.map(p => (
            <button
              key={`alert-${p.name}`}
              onClick={() => setConfiguringPlugin(p)}
              className="flex items-center space-x-1.5 px-3 py-1.5 bg-yellow-900/40 border border-yellow-700/50 rounded hover:bg-yellow-900/60 transition-colors text-yellow-500 text-xs font-bold"
            >
              <WarningIcon fontSize="inherit" />
              <span>{p.name} NEEDS CONFIG</span>
            </button>
          ))}
        </div>,
        document.getElementById('global-status-indicators') || document.body
      )}

      <div className="flex-1 overflow-y-auto p-4 min-h-0">
        {/* Install Bar */}
        {showInstallBar && (
          <div className="bg-[#0B0E11] rounded p-1.5 flex items-center border border-[#2B3139] mb-4">
            <DownloadIcon sx={{ color: '#5B616E', ml: 1, mr: 0.5 }} fontSize="small" />
            <form onSubmit={handleInstall} className="flex flex-1">
              <input
                type="text"
                value={installUri}
                onChange={(e) => setInstallUri(e.target.value)}
                placeholder="Install .wasm plugin via URI..."
                className="flex-1 bg-transparent border-none text-[13px] text-white px-2 focus:outline-none focus:ring-0 placeholder-[#5B616E]"
              />
              <button 
                type="submit"
                disabled={!installUri.trim()}
                className="px-4 py-1 bg-[#2B3139] hover:bg-[#3A414A] disabled:opacity-50 text-white text-[12px] font-medium rounded transition-colors"
              >
                Install
              </button>
            </form>
          </div>
        )}

        {plugins.length === 0 && (
          <div className="text-[#5B616E] text-center text-sm mt-8 font-medium">
            No plugins loaded.
          </div>
        )}

        <div className="space-y-6">
          {CATEGORY_ORDER.map(category => {
            if (allowedCategories && !allowedCategories.includes(category)) return null;
            const catPlugins = groupedPlugins[category];
            if (!catPlugins || catPlugins.length === 0) return null;

            return (
              <div key={category} className="space-y-1">
                <h4 className="text-[10px] font-bold text-[#5B616E] uppercase tracking-wider mb-2 pl-1 border-b border-[#2B3139] pb-1">
                  {category}
                </h4>
                {catPlugins.map((plugin) => {
                  const needsConfig = plugin.config?.some(c => c.required && (!c.value || c.value === ''));
                  const missingDeps = plugin.dependencies?.filter(dep => !pluginNames.includes(dep.name)) || [];
                  const hasError = needsConfig || missingDeps.length > 0;
                  
                  return (
                    <div 
                      key={plugin.name} 
                      onClick={(e) => {
                        e.stopPropagation();
                        setConfiguringPlugin(plugin);
                      }}
                      className="group flex items-center p-2 rounded hover:bg-[#2B3139] cursor-pointer transition-colors relative z-10"
                      title={hasError ? "Plugin needs configuration or dependencies" : "Plugin is ready"}
                    >
                      {/* Left-aligned status indicator */}
                      <div className="w-6 flex items-center justify-center shrink-0">
                        <div 
                          className={`w-2 h-2 rounded-full ${hasError ? 'bg-[#FB3836] shadow-[0_0_8px_rgba(251,56,54,0.8)] animate-pulse' : 'bg-[#00C087] shadow-[0_0_8px_rgba(0,192,135,0.5)]'}`}
                        ></div>
                      </div>
                      
                      <span className={`text-[13px] font-medium ml-1 truncate transition-colors ${hasError ? 'text-[#E2E8F0] group-hover:text-white' : 'text-[#A6B0C3] group-hover:text-[#E2E8F0]'}`}>
                        {plugin.name}
                      </span>
                      
                      {missingDeps.length > 0 && (
                        <span className="ml-3 text-[10px] text-[#FB3836] bg-[#FB3836]/10 px-1.5 py-0.5 rounded uppercase tracking-wider">Missing Deps</span>
                      )}
                    </div>
                  );
                })}
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
