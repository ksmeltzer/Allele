import React, { useState } from 'react';

const PluginManager = () => {
  const [activeTab, setActiveTab] = useState('exchanges');

  const tabs = [
    { id: 'exchanges', name: 'Exchanges' },
    { id: 'sensors', name: 'Sensors' },
    { id: 'strategies', name: 'Strategies' },
  ];

  return (
    <div className="bg-white p-6 rounded-lg shadow-md w-full max-w-4xl mx-auto mt-8">
      <h2 className="text-2xl font-bold mb-6 text-gray-800">Plugin Manager</h2>
      
      {/* Tabs */}
      <div className="border-b border-gray-200 mb-6">
        <nav className="-mb-px flex space-x-8">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`${
                activeTab === tab.id
                  ? 'border-indigo-500 text-indigo-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              } whitespace-nowrap py-4 px-1 border-b-2 font-medium text-sm transition-colors`}
            >
              {tab.name}
            </button>
          ))}
        </nav>
      </div>

      {/* Content Area */}
      <div className="bg-gray-50 p-6 rounded-md min-h-[300px] flex items-center justify-center border border-gray-100">
        <div className="text-center text-gray-500">
          <p className="text-lg">Manage {tabs.find(t => t.id === activeTab)?.name} Plugins</p>
          <p className="text-sm mt-2">Activate, configure, or disable available plugins.</p>
          
          <div className="mt-6 flex gap-4 justify-center">
             <div className="bg-white p-4 border rounded shadow-sm w-48 text-left">
                <div className="font-semibold text-gray-700">Example Plugin A</div>
                <div className="text-xs text-gray-400 mb-3">Status: Active</div>
                <button className="text-xs bg-red-100 text-red-600 px-3 py-1 rounded">Disable</button>
             </div>
             <div className="bg-white p-4 border rounded shadow-sm w-48 text-left">
                <div className="font-semibold text-gray-700">Example Plugin B</div>
                <div className="text-xs text-gray-400 mb-3">Status: Inactive</div>
                <button className="text-xs bg-green-100 text-green-700 px-3 py-1 rounded">Enable</button>
             </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default PluginManager;
