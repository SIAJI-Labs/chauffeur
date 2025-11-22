"use client";

// React & Next.js
import React from 'react';

// Third-party libraries
import {
  Search,
  Command
} from 'lucide-react';

interface SearchTriggerProps {
  onClick: () => void;
  className?: string;
}

export const SearchTrigger: React.FC<SearchTriggerProps> = ({
  onClick,
  className = ""
}) => {
  return (
    <button
      onClick={onClick}
      className={`
        w-full relative group
        ${className}
      `}
    >
      <Search className="absolute left-3 top-2.5 text-slate-500 group-hover:text-slate-400" size={16} />
      <div className="w-full bg-slate-800 border border-slate-700 text-slate-200 text-sm rounded-lg pl-9 pr-20 py-2 text-left focus:outline-none focus:ring-2 focus:ring-primary/50 group-hover:bg-slate-750 group-hover:border-slate-600 transition-colors">
        <span className="text-slate-400">Search docs...</span>
      </div>
      <div className="absolute right-2 top-2.5 flex items-center gap-1 pointer-events-none">
        <div className="flex items-center gap-1 px-1.5 py-0.5 rounded bg-slate-700 text-[10px] text-slate-400 font-medium">
          <Command size={10} />
          K
        </div>
      </div>
    </button>
  );
};