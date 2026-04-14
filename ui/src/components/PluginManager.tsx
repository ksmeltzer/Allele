import { useState, useEffect } from 'react';
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
          className="flex-1 bg-[#050505] border border-[#1f2937] rounded px-3 py-1.5 text-xs text-gray-300 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-900/50 transition-all font-mono"
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

export default function PluginManager() {
  const { connected, subscribe, sendEvent } = useWebSocket();
  const [plugins, setPlugins] = useState<Manifest[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const unsubscribe = subscribe('manifests_updated', (payload: Manifest[]) => {
      setPlugins(payload || []);
      setLoading(false);
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

  if (loading || !connected) {
    return (
      <div className="bg-[#050505] border border-[#1f2937] p-6 text-gray-500 text-sm flex flex-col items-center justify-center h-full">
        <span className="uppercase tracking-widest text-[10px] mb-2">{connected ? 'Requesting manifests...' : 'Waiting for connection...'}</span>
      </div>
    );
  }

  const pluginNames = plugins.map(p => p.name);

  return (
    <div className="bg-[#0a0a0c] h-full flex flex-col relative overflow-hidden">
      <div className="bg-[#050505] px-4 py-2 border-b border-[#1f2937] flex items-center justify-between z-10 shrink-0">
        <h2 className="text-xs text-gray-500 uppercase tracking-widest font-semibold">Plugin Vault</h2>
        <span className="text-[10px] text-gray-400 font-bold px-2 py-0.5 rounded bg-[#1f2937] uppercase tracking-wider">{plugins.length} Active</span>
      </div>
      
      <div className="flex-1 overflow-y-auto p-4 space-y-6 min-h-0">
        {plugins.length === 0 && (
          <div className="text-gray-600 text-center text-xs mt-10 uppercase tracking-widest">
            No modules loaded in registry.
          </div>
        )}

        {plugins.map((plugin) => (
          <div key={plugin.name} className="p-4 bg-[#050505] border border-[#1f2937] rounded-sm relative overflow-hidden group">
            <div className="flex justify-between items-start mb-3 relative z-10">
              <div>
                <h3 className="text-sm font-bold text-gray-200 mb-1">{plugin.name}</h3>
                <p className="text-[10px] text-gray-500 tracking-wide">{plugin.description} <span className="text-gray-700 ml-1">v{plugin.version}</span></p>
              </div>
              <span className="text-[9px] bg-[#1f2937] text-gray-400 px-2 py-1 rounded uppercase tracking-wider">
                {plugin.author}
              </span>
            </div>

            {/* Dependencies Section */}
            {plugin.dependencies && plugin.dependencies.length > 0 && (
              <div className="mt-4 mb-4 p-3 rounded-sm border border-[#1f2937] relative z-10">
                <h4 className="text-[9px] uppercase tracking-[0.2em] font-bold text-gray-600 mb-3 border-b border-gray-800 pb-2">Dependencies</h4>
                <div className="space-y-2">
                  {plugin.dependencies.map(dep => {
                    const isMet = pluginNames.includes(dep.name);
                    return (
                      <div key={dep.name} className="flex items-center space-x-3 text-xs p-2 rounded-sm border border-gray-800/50">
                        {isMet ? (
                          <div className="w-1.5 h-1.5 rounded-full bg-green-500"></div>
                        ) : (
                          <div className="w-1.5 h-1.5 rounded-full bg-red-500"></div>
                        )}
                        <span className={isMet ? "text-gray-400 font-mono" : "text-red-400 font-mono"}>
                          {dep.name} <span className="text-gray-600 ml-1 opacity-50">({dep.version})</span>
                        </span>
                      </div>
                    );
                  })}
                </div>
              </div>
            )}

            {/* Configuration Section */}
            {plugin.config && plugin.config.length > 0 && (
              <div className="mt-4 relative z-10">
                <h4 className="text-[9px] uppercase tracking-[0.2em] font-bold text-gray-600 mb-4 border-b border-gray-800 pb-2">Configuration Parameters</h4>
                <div className="space-y-4">
                  {plugin.config.map(field => (
                    <ConfigFieldRow 
                      key={field.key} 
                      pluginName={plugin.name} 
                      field={field} 
                      onSubmit={handleConfigSubmit} 
                    />
                  ))}
                </div>
              </div>
            )}
            
            {(!plugin.config || plugin.config.length === 0) && (!plugin.dependencies || plugin.dependencies.length === 0) && (
              <p className="text-[10px] text-gray-600/50 italic mt-4 relative z-10">No configuration or dependencies required.</p>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
