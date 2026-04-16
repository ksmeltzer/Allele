import React, { useState, useEffect } from 'react';
import { useWebSocket } from '../contexts/WebSocketContext';
import AccountBalanceWalletIcon from '@mui/icons-material/AccountBalanceWallet';
import ContentCopyIcon from '@mui/icons-material/ContentCopy';
import LocalGasStationIcon from '@mui/icons-material/LocalGasStation';
import AttachMoneyIcon from '@mui/icons-material/AttachMoney';

interface WalletPayload {
  address: string;
  network: string;
  matic: number;
  usdc: number;
}

const WalletInfo: React.FC = () => {
  const { subscribe } = useWebSocket();
  const [wallet, setWallet] = useState<WalletPayload | null>(null);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    const unsubscribe = subscribe('wallet_balance', (payload: WalletPayload) => {
      setWallet(payload);
    });
    return () => unsubscribe();
  }, [subscribe]);

  if (!wallet) return null;

  const truncateAddress = (address: string) => {
    if (!address || address.length < 10) return address;
    return `${address.slice(0, 6)}...${address.slice(-4)}`;
  };

  const formatCurrency = (val: number) => {
    return val.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 });
  };

  const copyAddress = () => {
    navigator.clipboard.writeText(wallet.address);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const isSimulation = wallet.network.toLowerCase().includes('amoy') || wallet.network.toLowerCase().includes('testnet');
  const needsFunding = isSimulation && wallet.matic < 0.01 && wallet.usdc < 1;

  return (
    <>
      <div className="flex items-center space-x-3 text-[12px] font-medium text-[#A6B0C3] bg-[#1B2028] border border-[#2B3139] rounded px-3 py-1.5 mr-2 transition-all duration-300">
        <AccountBalanceWalletIcon fontSize="small" sx={{ color: isSimulation ? '#A855F7' : '#4F46E5' }} />
        
        <div className="flex flex-col border-r border-[#2B3139] pr-3 mr-1">
          <span className="text-[10px] text-[#A6B0C3] uppercase tracking-wider">
            {isSimulation ? 'Simulation' : wallet.network}
          </span>
          <span className="font-mono text-white tracking-tight">{truncateAddress(wallet.address)}</span>
        </div>
        
        <div className="flex space-x-4">
          <div className="flex flex-col">
            <span className="text-[10px] text-[#A6B0C3] uppercase tracking-wider">USDC</span>
            <span className="font-mono text-[#00C087]">{formatCurrency(wallet.usdc)}</span>
          </div>
          <div className="flex flex-col">
            <span className="text-[10px] text-[#A6B0C3] uppercase tracking-wider">MATIC</span>
            <span className="font-mono text-white">{formatCurrency(wallet.matic)}</span>
          </div>
        </div>
      </div>

      {needsFunding && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/60 backdrop-blur-sm modal-enter p-4">
          <div className="bg-[#0B0E11] border border-[#2B3139] rounded-xl shadow-2xl max-w-lg w-full overflow-hidden flex flex-col relative">
            
            {/* Subtle brand tint background at top */}
            <div className="absolute top-0 inset-x-0 h-32 bg-[#A855F7]/10 pointer-events-none opacity-50 blur-3xl"></div>

            <div className="p-8 flex flex-col gap-6 relative z-10">
              <div className="flex flex-col gap-2">
                <span className="text-[11px] font-mono text-[#A855F7] uppercase tracking-widest font-bold">
                  Action Required
                </span>
                <h2 className="text-2xl font-bold text-white tracking-tight leading-tight">
                  Fund Your Simulation Engine
                </h2>
                <p className="text-sm text-[#A6B0C3] leading-relaxed max-w-[45ch]">
                  Your AI trading agent is running in Simulation Mode. To begin executing paper trades safely, you need to request free Practice Credits to bootstrap its digital identity.
                </p>
              </div>

              <div className="flex flex-col gap-4">
                <div className="flex flex-col gap-1.5">
                  <label className="text-[10px] text-[#5B616E] uppercase tracking-widest font-bold ml-1">
                    Agent Identity
                  </label>
                  <div className="flex items-center justify-between bg-[#1B2028] border border-[#2B3139] rounded-lg p-1 pl-4">
                    <span className="font-mono text-sm text-[#E2E8F0] tracking-tight truncate">
                      {wallet.address}
                    </span>
                    <button 
                      onClick={copyAddress}
                      className="flex items-center gap-2 px-4 py-2 bg-[#2B3139] hover:bg-[#3A414C] text-white text-xs font-semibold rounded-md transition-colors"
                    >
                      <ContentCopyIcon fontSize="inherit" />
                      {copied ? 'Copied' : 'Copy Identity'}
                    </button>
                  </div>
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 mt-2">
                  <a 
                    href="https://faucet.polygon.technology/" 
                    target="_blank" 
                    rel="noreferrer"
                    className="flex items-center gap-3 p-4 bg-[#1B2028] hover:bg-[#2B3139] border border-[#2B3139] rounded-lg transition-all group"
                  >
                    <div className="h-8 w-8 rounded-full bg-[#0B0E11] flex items-center justify-center text-[#A6B0C3] group-hover:text-white transition-colors">
                      <LocalGasStationIcon fontSize="small" />
                    </div>
                    <div className="flex flex-col">
                      <span className="text-sm font-semibold text-white">Network Gas</span>
                      <span className="text-[10px] text-[#A6B0C3] uppercase tracking-wide">Amoy Faucet</span>
                    </div>
                  </a>
                  
                  <a 
                    href="https://faucet.circle.com/" 
                    target="_blank" 
                    rel="noreferrer"
                    className="flex items-center gap-3 p-4 bg-[#1B2028] hover:bg-[#2B3139] border border-[#2B3139] rounded-lg transition-all group"
                  >
                    <div className="h-8 w-8 rounded-full bg-[#00C087]/10 flex items-center justify-center text-[#00C087] group-hover:bg-[#00C087]/20 transition-colors">
                      <AttachMoneyIcon fontSize="small" />
                    </div>
                    <div className="flex flex-col">
                      <span className="text-sm font-semibold text-white">Trading Capital</span>
                      <span className="text-[10px] text-[#A6B0C3] uppercase tracking-wide">Circle USDC Faucet</span>
                    </div>
                  </a>
                </div>
              </div>

              <div className="pt-2 border-t border-[#2B3139]/50 mt-2">
                <p className="text-[11px] text-[#5B616E] flex items-center justify-center gap-2">
                  <span className="w-1.5 h-1.5 rounded-full bg-yellow-500 animate-pulse"></span>
                  Waiting for incoming practice credits...
                </p>
              </div>
            </div>
          </div>
        </div>
      )}
    </>
  );
};

export default WalletInfo;
