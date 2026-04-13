export default function GeneticArena() {
  return (
    <div className="bg-gray-900 text-white rounded-lg shadow-xl p-6 w-full max-w-2xl border border-gray-800">
      <h2 className="text-2xl font-bold mb-6 text-emerald-400 border-b border-gray-700 pb-2">
        Genetic Arena Leaderboard
      </h2>
      
      <div className="space-y-4">
        {/* Header */}
        <div className="grid grid-cols-4 gap-4 text-xs font-semibold text-gray-400 uppercase tracking-wider mb-2">
          <div className="col-span-1">Organism</div>
          <div className="col-span-1 text-right">Fitness Score</div>
          <div className="col-span-2 text-center">Capital Allocation</div>
        </div>

        {/* Organism Row 1 */}
        <div className="grid grid-cols-4 gap-4 items-center bg-gray-800 p-3 rounded-md hover:bg-gray-750 transition-colors">
          <div className="col-span-1 font-mono text-blue-300 font-medium text-sm">Alpha-7X</div>
          <div className="col-span-1 text-right font-mono text-amber-400">98.24</div>
          <div className="col-span-2 flex items-center space-x-2">
            <div className="flex-1 bg-gray-700 h-2.5 rounded-full overflow-hidden">
              <div className="bg-emerald-500 h-full rounded-full" style={{ width: '80%' }}></div>
            </div>
            <span className="text-xs text-gray-500 w-8">80%</span>
          </div>
        </div>

        {/* Organism Row 2 */}
        <div className="grid grid-cols-4 gap-4 items-center bg-gray-800 p-3 rounded-md hover:bg-gray-750 transition-colors">
          <div className="col-span-1 font-mono text-purple-300 font-medium text-sm">Beta-9Q</div>
          <div className="col-span-1 text-right font-mono text-amber-400">85.41</div>
          <div className="col-span-2 flex items-center space-x-2">
            <div className="flex-1 bg-gray-700 h-2.5 rounded-full overflow-hidden">
              <div className="bg-emerald-500 h-full rounded-full" style={{ width: '45%' }}></div>
            </div>
            <span className="text-xs text-gray-500 w-8">45%</span>
          </div>
        </div>

        {/* Organism Row 3 */}
        <div className="grid grid-cols-4 gap-4 items-center bg-gray-800 p-3 rounded-md hover:bg-gray-750 transition-colors opacity-75">
          <div className="col-span-1 font-mono text-rose-300 font-medium text-sm">Gamma-1V</div>
          <div className="col-span-1 text-right font-mono text-amber-400">42.10</div>
          <div className="col-span-2 flex items-center space-x-2">
            <div className="flex-1 bg-gray-700 h-2.5 rounded-full overflow-hidden">
              <div className="bg-rose-500 h-full rounded-full" style={{ width: '15%' }}></div>
            </div>
            <span className="text-xs text-gray-500 w-8">15%</span>
          </div>
        </div>
      </div>
    </div>
  );
}
