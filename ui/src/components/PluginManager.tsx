import { useState, useEffect } from 'react';
import type { Manifest, ConfigField } from '../types/plugin';

function ConfigFieldRow({ 
  pluginName, 
  field, 
  onSubmit 
}: { 
  pluginName: string, 
  field: ConfigField, 
  onSubmit: (pluginName: string, key: string, value: string) => Promise<void> 
}) {
  const [localValue, setLocalValue] = useState(field.value || '');

  return (
    <div className="flex flex-col space-y-1 mb-4">
      <label className="text-xs font-semibold text-gray-400">
        {field.key} {field.required && <span className="text-red-500">*</span>}
        <span className="ml-2 font-normal text-gray-500">- {field.description}</span>
      </label>
      <div className="flex space-x-2">
        <input
          type={field.type === 'secret' ? 'password' : 'text'}
          value={localValue}
          onChange={(e) => setLocalValue(e.target.value)}
          placeholder={field.type === 'secret' && field.value === '********' ? '********' : ''}
          className="flex-1 bg-gray-800 border border-gray-700 rounded px-3 py-1.5 text-sm text-white focus:outline-none focus:border-blue-500"
        />
        <button 
          onClick={() => onSubmit(pluginName, field.key, localValue)}
          disabled={localValue === field.value || localValue === ''}
          className="px-3 py-1.5 bg-purple-600 hover:bg-purple-500 disabled:bg-gray-700 disabled:text-gray-500 text-white text-sm rounded transition"
        >
          Save
        </button>
      </div>
    </div>
  );
}

export default function PluginManager() {
  const [plugins, setPlugins] = useState<Manifest[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchPlugins();
  }, []);

  const fetchPlugins = async () => {
    try {
      const res = await fetch(`http://${window.location.hostname}:8081/api/plugins`);
      if (res.ok) {
        const data = await res.json();
        setPlugins(data || []);
      }
    } catch (e) {
      console.error('Failed to fetch plugins', e);
    } finally {
      setLoading(false);
    }
  };

  const handleConfigSubmit = async (pluginName: string, key: string, value: string) => {
    if (value === '********') return; // Don't submit masked value
    try {
      await fetch(`http://${window.location.hostname}:8081/api/plugins/config`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ plugin_name: pluginName, key, value })
      });
      // Refresh to get the updated status
      await fetchPlugins();
    } catch (e) {
      console.error('Failed to save config', e);
    }
  };

  if (loading) return <div className="text-gray-500 bg-gray-800 p-6 rounded-lg border border-gray-700">Loading plugins...</div>;

  const pluginNames = plugins.map(p => p.name);

  return (
    <div className="bg-gray-800 rounded-lg shadow-lg border border-gray-700 h-fit flex flex-col max-h-[80vh] overflow-hidden">
      <div className="p-6 pb-2">
        <h2 className="text-xl font-semibold mb-2 text-gray-300">Plugin Manager & Vault</h2>
        <p className="text-xs text-gray-500 mb-4">Manage loaded WebAssembly plugins, system configuration, and dependency health.</p>
      </div>
      
      <div className="flex-1 overflow-y-auto p-6 pt-0 space-y-6">
        {plugins.map((plugin) => (
          <div key={plugin.name} className="p-4 bg-gray-900 border border-gray-700 rounded-lg">
            <div className="flex justify-between items-start mb-2">
              <div>
                <h3 className="text-md font-bold text-blue-400">{plugin.name}</h3>
                <p className="text-xs text-gray-400 mt-1">{plugin.description} (v{plugin.version})</p>
              </div>
              <span className="text-xs bg-gray-800 text-gray-400 border border-gray-700 px-2 py-1 rounded">
                {plugin.author}
              </span>
            </div>

            {/* Dependencies Section */}
            {plugin.dependencies && plugin.dependencies.length > 0 && (
              <div className="mt-4 mb-4 bg-gray-800 p-3 rounded border border-gray-700">
                <h4 className="text-xs uppercase tracking-wider font-semibold text-gray-400 mb-2">Dependencies</h4>
                <div className="space-y-2">
                  {plugin.dependencies.map(dep => {
                    const isMet = pluginNames.includes(dep.name);
                    return (
                      <div key={dep.name} className="flex items-center space-x-2 text-sm bg-gray-900 p-2 rounded border border-gray-700">
                        {isMet ? (
                          <span className="text-green-500">✓</span>
                        ) : (
                          <span className="text-yellow-500">⚠️</span>
                        )}
                        <span className={isMet ? "text-gray-300" : "text-yellow-400"}>
                          {dep.name} <span className="text-gray-500">({dep.version})</span>
                        </span>
                        {!isMet && dep.url && (
                          <a 
                            href={dep.url} 
                            target="_blank" 
                            rel="noreferrer"
                            className="ml-auto text-xs bg-yellow-600/20 hover:bg-yellow-600/40 text-yellow-500 border border-yellow-600/50 px-2 py-1 rounded transition"
                          >
                            Resolve Missing
                          </a>
                        )}
                      </div>
                    );
                  })}
                </div>
              </div>
            )}

            {/* Configuration Section */}
            {plugin.config && plugin.config.length > 0 && (
              <div className="mt-4">
                <h4 className="text-xs uppercase tracking-wider font-semibold text-gray-400 mb-3">Configuration</h4>
                <div>
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
              <p className="text-xs text-gray-600 italic mt-4">No configuration or dependencies required.</p>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
