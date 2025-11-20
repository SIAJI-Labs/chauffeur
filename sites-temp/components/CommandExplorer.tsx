import React, { useState } from 'react';
import { COMMAND_EXAMPLES } from '../constants';
import { ChevronRight, Terminal as TerminalIcon } from 'lucide-react';

export const CommandExplorer: React.FC = () => {
  const [activeIndex, setActiveIndex] = useState(0);

  return (
    <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 bg-surface rounded-2xl border border-slate-700 overflow-hidden shadow-xl">
      {/* Sidebar */}
      <div className="lg:col-span-4 bg-slate-900/50 border-r border-slate-700 p-4 flex flex-col gap-2">
        <div className="pb-4 mb-2 border-b border-slate-700/50">
          <h3 className="text-xs font-bold text-slate-400 uppercase tracking-wider px-3">Available Commands</h3>
        </div>
        {COMMAND_EXAMPLES.map((cmd, idx) => (
          <button
            key={idx}
            onClick={() => setActiveIndex(idx)}
            className={`w-full text-left px-4 py-3 rounded-lg text-sm font-medium transition-all flex items-center justify-between group ${
              activeIndex === idx 
                ? 'bg-primary/10 text-primary border border-primary/20' 
                : 'text-slate-400 hover:bg-slate-800 hover:text-slate-200'
            }`}
          >
            <span>{cmd.name}</span>
            {activeIndex === idx && <ChevronRight size={16} />}
          </button>
        ))}
      </div>

      {/* Main Display */}
      <div className="lg:col-span-8 p-8 flex flex-col justify-center min-w-0">
        <div className="mb-8">
          <div className="flex items-center gap-2 mb-2">
             <TerminalIcon className="text-slate-500" size={18} />
             <span className="text-sm text-slate-500 font-mono">Command Preview</span>
          </div>
          <div className="bg-slate-950 rounded-lg border border-slate-800 p-6 relative group overflow-x-auto scrollbar-thin scrollbar-thumb-slate-700 scrollbar-track-transparent">
            <div className="absolute top-3 right-3 opacity-0 group-hover:opacity-100 transition-opacity">
                <span className="text-xs text-slate-500 bg-slate-900 px-2 py-1 rounded">Click to copy</span>
            </div>
            <code className="font-mono text-lg text-emerald-400 whitespace-nowrap">
              <span className="text-slate-500">$</span> {COMMAND_EXAMPLES[activeIndex].command}
            </code>
          </div>
        </div>

        <div className="space-y-4">
           <div>
             <h4 className="text-sm font-semibold text-slate-200 mb-1">What it does</h4>
             <p className="text-slate-400">{COMMAND_EXAMPLES[activeIndex].description}</p>
           </div>
           <div>
             <h4 className="text-sm font-semibold text-slate-200 mb-1">Expected Output</h4>
             <div className="font-mono text-sm text-slate-500 bg-slate-900/50 p-3 rounded border border-slate-800">
                {COMMAND_EXAMPLES[activeIndex].output}
             </div>
           </div>
        </div>
      </div>
    </div>
  );
};