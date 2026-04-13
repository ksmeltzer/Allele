import React, { useState } from 'react';

const SecureVault = () => {
  const [exchange, setExchange] = useState('');
  const [apiKey, setApiKey] = useState('');
  const [apiSecret, setApiSecret] = useState('');
  const [globalLimit, setGlobalLimit] = useState('');

  const handleSaveConfig = (e) => {
    e.preventDefault();
    console.log('SecureVault Config Saved', { exchange, apiKey, apiSecret, globalLimit });
    alert('Configuration saved securely.');
  };

  return (
    <div className="bg-white p-6 rounded-lg shadow-md w-full max-w-4xl mx-auto mt-8">
      <h2 className="text-2xl font-bold mb-6 text-gray-800">Secure Vault Configuration</h2>
      
      <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
        
        {/* API Keys Configuration */}
        <div>
          <h3 className="text-lg font-semibold mb-4 text-indigo-600 border-b pb-2">API Keys</h3>
          <form className="space-y-4" onSubmit={handleSaveConfig}>
            <div>
              <label className="block text-sm font-medium text-gray-700">Exchange</label>
              <select
                value={exchange}
                onChange={(e) => setExchange(e.target.value)}
                className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md shadow-sm border"
              >
                <option value="">Select Exchange...</option>
                <option value="binance">Binance</option>
                <option value="coinbase">Coinbase</option>
                <option value="kraken">Kraken</option>
              </select>
            </div>
            
            <div>
              <label className="block text-sm font-medium text-gray-700">API Key</label>
              <input
                type="text"
                value={apiKey}
                onChange={(e) => setApiKey(e.target.value)}
                className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                placeholder="Enter API Key"
              />
            </div>
            
            <div>
              <label className="block text-sm font-medium text-gray-700">API Secret</label>
              <input
                type="password"
                value={apiSecret}
                onChange={(e) => setApiSecret(e.target.value)}
                className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                placeholder="Enter API Secret"
              />
            </div>
            
            <button
              type="submit"
              className="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
            >
              Save Credentials
            </button>
          </form>
        </div>

        {/* Global Limits Configuration */}
        <div>
          <h3 className="text-lg font-semibold mb-4 text-emerald-600 border-b pb-2">Global Limits</h3>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700">Max Portfolio Exposure (%)</label>
              <input
                type="number"
                value={globalLimit}
                onChange={(e) => setGlobalLimit(e.target.value)}
                className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm py-2 px-3 focus:outline-none focus:ring-emerald-500 focus:border-emerald-500 sm:text-sm"
                placeholder="e.g., 50"
              />
              <p className="mt-2 text-xs text-gray-500">
                Set a hard limit across all active strategies to prevent catastrophic loss.
              </p>
            </div>

            <div className="bg-red-50 border-l-4 border-red-400 p-4 mt-6">
              <div className="flex">
                <div className="ml-3">
                  <p className="text-sm text-red-700 font-semibold">
                    Kill Switch
                  </p>
                  <p className="mt-2 text-xs text-red-600">
                    Immediately halt all trading activities and cancel open orders across all connected exchanges.
                  </p>
                  <button className="mt-3 bg-red-600 text-white px-4 py-2 text-sm rounded shadow hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-1">
                    Emergency Halt
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
        
      </div>
    </div>
  );
};

export default SecureVault;
