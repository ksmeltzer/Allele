import { useState, useEffect } from 'react';
import { createPortal } from 'react-dom';
import { useWebSocket } from '../contexts/WebSocketContext';
import type { Manifest, ConfigField } from '../types/plugin';

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
      <label className="text-[10px] font-semibold text-gray-400 uppercase tracking-widest flex items-center">
        {field.key} {field.required && <span className="text-red-500 ml-1">*</span>}
        <span className="ml-2 font-normal text-gray-500 lowercase opacity-70">- {field.description}</span>
      </label>
      <div className="flex space-x-2">
        <input
          type={field.type === 'secret' ? 'password' : 'text'}
          value={localValue}
          onChange={(e) => setLocalValue(e.target.value)}
          placeholder={field.type === 'secret' && field.value === '********' ? '********' : 'Enter value...'}
          className={`flex-1 bg-[#050505] border ${needsConfig ? 'border-yellow-600/50 focus:border-yellow-500' : 'border-[#1f2937] focus:border-blue-500'} rounded px-3 py-1.5 text-xs text-gray-300 focus:outline-none focus:ring-1 focus:ring-blue-900/50 transition-all font-mono`}
        />
        <button 
          onClick={() => onSubmit(pluginName, field.key, localValue)}
          disabled={localValue === field.value || localValue === ''}
          className="px-3 py-1.5 bg-[#1f2937] hover:bg-[#374151] border border-gray-700 disabled:opacity-50 text-gray-200 text-xs font-bold uppercase tracking-wider rounded transition-all duration-200"
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
    <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
      <div className="bg-[#0a0a0c] border border-[#2d3748] rounded-md shadow-2xl max-w-2xl w-full max-h-[90vh] flex flex-col overflow-hidden animate-in fade-in zoom-in-95 duration-200">
        <div className="px-5 py-3 border-b border-[#1f2937] flex items-center justify-between bg-[#050505]">
          <div>
            <h3 className="text-sm font-bold text-gray-200">{plugin.name} Configuration</h3>
            <p className="text-[10px] text-gray-500 tracking-wide uppercase">{plugin.description}</p>
          </div>
          <button 
            onClick={onClose}
            className="text-gray-500 hover:text-gray-300 transition-colors p-1"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12"></path></svg>
          </button>
        </div>
        <div className="p-5 overflow-y-auto">
          {plugin.config && plugin.config.length > 0 ? (
            <div className="space-y-4">
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
            <p className="text-xs text-gray-500 italic">This plugin does not require any configuration.</p>
          )}
        </div>
        <div className="px-5 py-3 border-t border-[#1f2937] bg-[#050505] flex justify-end">
          <button 
            onClick={onClose}
            className="px-4 py-2 bg-[#1f2937] hover:bg-[#374151] border border-gray-700 text-gray-200 text-xs font-bold uppercase tracking-wider rounded transition-all duration-200"
          >
            Done
          </button>
        </div>
      </div>
    </div>,
    document.body
  );
}

export default function PluginManager() {
  const { connected, subscribe, sendEvent } = useWebSocket();
  const [plugins, setPlugins] = useState<Manifest[]>([]);
  const [loading, setLoading] = useState(true);
  const [installUri, setInstallUri] = useState('');
  const [configuringPlugin, setConfiguringPlugin] = useState<Manifest | null>(null);

  useEffect(() => {
    const unsubscribe = subscribe('manifests_updated', (payload: Manifest[]) => {
      setPlugins(payload || []);
      setLoading(false);
      
      // Update configuring plugin if it's open so it gets the new state
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
    if (value === '********') return; // Don't submit masked value
    sendEvent('update_config', { plugin_name: pluginName, key, value });
  };

  const handleInstall = (e: React.FormEvent) => {
    e.preventDefault();
    if (!installUri.trim()) return;
    sendEvent('install_plugin', { uri: installUri.trim() });
    setInstallUri('');
  };

  if (loading || !connected) {
    return (
      <div className="bg-[#050505] border border-[#1f2937] p-6 text-gray-500 text-sm flex flex-col items-center justify-center h-full">
        <span className="uppercase tracking-widest text-[10px] mb-2">{connected ? 'Requesting manifests...' : 'Waiting for connection...'}</span>
      </div>
    );
  }

  const pluginNames = plugins.map(p => p.name);

  // Find plugins that need configuration
  const pluginsNeedingConfig = plugins.filter(p => 
    p.config?.some(c => c.required && (!c.value || c.value === ''))
  );

  return (
    <div className="bg-[#0a0a0c] h-full flex flex-col relative overflow-hidden">
      {configuringPlugin && (
        <ConfigModal 
          plugin={configuringPlugin} 
          onClose={() => setConfiguringPlugin(null)} 
          onSubmit={handleConfigSubmit} 
        />
      )}

      {/* Global Status Injector - Injects into the main header */}
      {pluginsNeedingConfig.length > 0 && createPortal(
        <div className="flex items-center space-x-2 animate-pulse">
          {pluginsNeedingConfig.map(p => (
            <button
              key={`alert-${p.name}`}
              onClick={() => setConfiguringPlugin(p)}
              className="flex items-center space-x-1.5 px-2 py-1 bg-yellow-900/30 border border-yellow-700 rounded text-yellow-500 text-[10px] uppercase font-bold tracking-widest hover:bg-yellow-900/50 transition-colors"
            >
              <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path></svg>
              <span>{p.name} CONFIG REQUIRED</span>
            </button>
          ))}
        </div>,
        document.getElementById('global-status-indicators') || document.body
      )}

      <div className="bg-[#050505] px-4 py-2 border-b border-[#1f2937] flex items-center justify-between z-10 shrink-0">
        <h2 className="text-xs text-gray-500 uppercase tracking-widest font-semibold">Plugin Vault</h2>
        <span className="text-[10px] text-gray-400 font-bold px-2 py-0.5 rounded bg-[#1f2937] uppercase tracking-wider">{plugins.length} Active</span>
      </div>
      
      <div className="flex-1 overflow-y-auto p-4 space-y-3 min-h-0">
        <div className="p-3 bg-[#050505] border border-[#1f2937] rounded-sm relative overflow-hidden group">
          <form onSubmit={handleInstall} className="flex space-x-2 relative z-10">
            <input
              type="text"
              value={installUri}
              onChange={(e) => setInstallUri(e.target.value)}
              placeholder="Enter plugin URI to install..."
              className="flex-1 bg-[#0a0a0c] border border-[#1f2937] rounded px-3 py-1.5 text-xs text-gray-300 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-900/50 transition-all font-mono"
            />
            <button 
              type="submit"
              disabled={!installUri.trim()}
              className="px-3 py-1.5 bg-[#1f2937] hover:bg-[#374151] border border-gray-700 disabled:opacity-50 text-gray-200 text-xs font-bold uppercase tracking-wider rounded transition-all duration-200"
            >
              Install
            </button>
          </form>
        </div>

        {plugins.length === 0 && (
          <div className="text-gray-600 text-center text-xs mt-10 uppercase tracking-widest">
            No modules loaded in registry.
          </div>
        )}

        {plugins.map((plugin) => {
          const needsConfig = plugin.config?.some(c => c.required && (!c.value || c.value === ''));
          const missingDeps = plugin.dependencies?.filter(dep => !pluginNames.includes(dep.name)) || [];
          
          return (
            <div key={plugin.name} className={`p-3 bg-[#050505] border ${needsConfig || missingDeps.length > 0 ? 'border-yellow-900/50' : 'border-[#1f2937]'} rounded-sm relative overflow-hidden flex items-center justify-between`}>
              <div className="flex items-center space-x-4">
                {/* Status Indicator */}
                <div className="relative flex items-center justify-center w-6 h-6">
                  {needsConfig || missingDeps.length > 0 ? (
                    <div className="absolute inset-0 bg-yellow-500/20 rounded-full animate-pulse"></div>
                  ) : null}
                  <div className={`w-2.5 h-2.5 rounded-full ${needsConfig || missingDeps.length > 0 ? 'bg-yellow-500' : 'bg-green-500'}`}></div>
                </div>

                <div>
                  <div className="flex items-center space-x-2">
                    <h3 className="text-sm font-bold text-gray-200">{plugin.name}</h3>
                    <span className="text-[9px] text-gray-500 uppercase tracking-wider bg-[#1f2937] px-1.5 py-0.5 rounded">v{plugin.version}</span>
                  </div>
                  
                  {/* Compact Status Messages */}
                  {missingDeps.length > 0 && (
                    <p className="text-[10px] text-red-400 mt-1 uppercase tracking-widest">Missing deps: {missingDeps.map(d => d.name).join(', ')}</p>
                  )}
                  {needsConfig && (
                    <p className="text-[10px] text-yellow-500 mt-1 uppercase tracking-widest">Configuration Required</p>
                  )}
                </div>
              </div>

              <div className="flex items-center space-x-3">
                <span className="text-[9px] text-gray-600 uppercase tracking-widest hidden sm:block">{plugin.author}</span>
                {(plugin.config && plugin.config.length > 0) && (
                  <button 
                    onClick={() => setConfiguringPlugin(plugin)}
                    className="p-1.5 text-gray-400 hover:text-white bg-[#1f2937] hover:bg-gray-700 rounded transition-colors"
                    title="Configure Plugin"
                  >
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"></path>
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
                    </svg>
                  </button>
                )}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
