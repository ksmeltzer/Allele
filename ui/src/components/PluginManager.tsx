import { useState, useEffect, useMemo } from 'react';
import { createPortal } from 'react-dom';
import { useWebSocket } from '../contexts/WebSocketContext';
import type { Manifest, ConfigField } from '../types/plugin';

// Material UI Icons
import SettingsIcon from '@mui/icons-material/Settings';
import CloseIcon from '@mui/icons-material/Close';
import InfoIcon from '@mui/icons-material/Info';
import PulseIcon from '@mui/icons-material/WifiTethering';
import DownloadIcon from '@mui/icons-material/FileDownload';

function ConfigFieldRow({ 
  field, 
  value,
  onChange 
}: { 
  field: ConfigField, 
  value: string,
  onChange: (val: string) => void 
}) {
  const needsConfig = field.required && (!value || value === '');

  return (
    <div className="flex flex-col space-y-1.5 mb-4">
      <label className="text-[12px] font-medium text-[#A6B0C3] flex items-center">
        {field.key} {field.required && <span className="text-[#FB3836] ml-1">*</span>}
        <span className="ml-2 font-normal text-[#5B616E]">- {field.description}</span>
      </label>
      <input
        type={field.type === 'secret' ? 'password' : 'text'}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={field.type === 'secret' && field.value === '********' ? '********' : 'Enter value...'}
        className={`w-full bg-[#0B0E11] border ${needsConfig ? 'border-yellow-600 focus:border-yellow-500' : 'border-[#2B3139] focus:border-[#4F46E5]'} rounded px-3 py-1.5 text-sm text-white focus:outline-none focus:ring-1 focus:ring-[#4F46E5]/50 transition-all font-mono`}
      />
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
  onSubmit: (pluginName: string, configs: Record<string, string>) => void 
}) {
  const [localConfigs, setLocalConfigs] = useState<Record<string, string>>(() => {
    const init: Record<string, string> = {};
    plugin.config?.forEach(c => {
      init[c.key] = c.value || '';
    });
    return init;
  });

  const handleSave = () => {
    onSubmit(plugin.name, localConfigs);
    onClose();
  };

  return createPortal(
    <div className="fixed inset-0 z-[9999] flex items-center justify-center bg-black/80 backdrop-blur-sm p-4">
      <div className="bg-[#1B2028] border border-[#2B3139] rounded-xl shadow-2xl max-w-2xl w-full max-h-[90vh] flex flex-col overflow-hidden modal-enter">
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
                  field={field} 
                  value={localConfigs[field.key] || ''}
                  onChange={(val) => setLocalConfigs(prev => ({ ...prev, [field.key]: val }))}
                />
              ))}
            </div>
          ) : (
            <p className="text-sm text-[#A6B0C3] italic flex items-center">
              <InfoIcon fontSize="small" sx={{ mr: 1 }} /> This plugin does not require any configuration.
            </p>
          )}
        </div>
        <div className="px-6 py-4 border-t border-[#2B3139] bg-[#1B2028] flex justify-end space-x-3">
          <button 
            onClick={onClose}
            className="px-6 py-2 bg-[#2B3139] hover:bg-[#3A414A] text-white text-sm font-medium rounded transition-all duration-200"
          >
            Cancel
          </button>
          <button 
            onClick={handleSave}
            className="px-6 py-2 bg-[#4F46E5] hover:bg-[#4338CA] text-white text-sm font-medium rounded transition-all duration-200"
          >
            Save
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
  const [pluginStatuses, setPluginStatuses] = useState<Record<string, { level: string; message: string }>>({});

  useEffect(() => {
    const handleOpenConfig = (e: Event) => {
      const customEvent = e as CustomEvent<Manifest>;
      // Only open if this PluginManager contains the plugin category
      const plugin = customEvent.detail;
      const cat = getPluginCategory(plugin.name);
      if (!allowedCategories || allowedCategories.includes(cat)) {
        setConfiguringPlugin(plugin);
      }
    };
    
    document.addEventListener('open-plugin-config', handleOpenConfig);
    return () => document.removeEventListener('open-plugin-config', handleOpenConfig);
  }, [allowedCategories]);

  useEffect(() => {
    const unsubscribeManifests = subscribe('manifests_updated', (payload: Manifest[]) => {
      setPlugins(payload || []);
      setLoading(false);
      setConfiguringPlugin(prev => {
        if (!prev) return null;
        const updated = (payload || []).find(p => p.name === prev.name);
        return updated || prev;
      });
    });

    const unsubscribeAlerts = subscribe('system_alert', (payload: { source: string; level: string; message: string }) => {
      if (payload && payload.source) {
        setPluginStatuses(prev => {
          // If it's an info alert, we assume the error condition is resolved.
          if (payload.level === 'info') {
            const newStatuses = { ...prev };
            delete newStatuses[payload.source];
            return newStatuses;
          }
          return {
            ...prev,
            [payload.source]: { level: payload.level, message: payload.message }
          };
        });
      }
    });

    if (connected) {
      sendEvent('request_manifests');
    }

    return () => {
      unsubscribeManifests();
      unsubscribeAlerts();
    };
  }, [connected, subscribe, sendEvent]);

  const handleConfigSubmit = (pluginName: string, configs: Record<string, string>) => {
    Object.entries(configs).forEach(([key, value]) => {
      if (value === '********') return;
      // Find original value
      const original = plugins.find(p => p.name === pluginName)?.config?.find(c => c.key === key)?.value;
      if (value !== original) {
        sendEvent('update_config', { plugin_name: pluginName, key, value });
      }
    });
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

  return (
    <div className="bg-[#1B2028] h-full flex flex-col relative overflow-hidden">
      {configuringPlugin && (
        <ConfigModal plugin={configuringPlugin} onClose={() => setConfiguringPlugin(null)} onSubmit={handleConfigSubmit} />
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
                  const runtimeAlert = pluginStatuses[plugin.name];
                  const hasError = needsConfig || missingDeps.length > 0 || (runtimeAlert && (runtimeAlert.level === 'error' || runtimeAlert.level === 'warning'));
                  
                  let tooltip = "Plugin is ready";
                  if (needsConfig) tooltip = "Plugin needs configuration";
                  else if (missingDeps.length > 0) tooltip = "Missing dependencies";
                  else if (runtimeAlert) tooltip = runtimeAlert.message;
                  
                  return (
                    <div 
                      key={plugin.name} 
                      onClick={(e) => {
                        e.stopPropagation();
                        setConfiguringPlugin(plugin);
                      }}
                      className="group flex items-center p-2 rounded hover:bg-[#2B3139] cursor-pointer transition-colors relative z-10"
                      title={tooltip}
                    >
                      {/* Left-aligned status indicator */}
                      <div className="w-6 flex items-center justify-center shrink-0">
                        <div className="relative flex h-2 w-2">
                          {hasError && (
                            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-red-500 opacity-75"></span>
                          )}
                          <span className={`relative inline-flex rounded-full h-2 w-2 ${hasError ? 'bg-red-500 shadow-[0_0_8px_rgba(239,68,68,0.8)]' : 'bg-[#00C087] shadow-[0_0_8px_rgba(0,192,135,0.5)]'}`}></span>
                        </div>
                      </div>
                      
                      <span className={`text-[13px] font-medium ml-1 truncate transition-colors ${hasError ? 'text-[#E2E8F0] group-hover:text-white' : 'text-[#A6B0C3] group-hover:text-[#E2E8F0]'}`}>
                        {plugin.name}
                      </span>
                      
                      {missingDeps.length > 0 && (
                        <span className="ml-3 text-[10px] text-[#FB3836] bg-[#FB3836]/10 px-1.5 py-0.5 rounded uppercase tracking-wider">Missing Deps</span>
                      )}
                      
                      {runtimeAlert && (runtimeAlert.level === 'warning' || runtimeAlert.level === 'error') && !missingDeps.length && !needsConfig && (
                        <span className="ml-3 text-[10px] text-[#FB3836] bg-[#FB3836]/10 px-1.5 py-0.5 rounded uppercase tracking-wider">System Error</span>
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
