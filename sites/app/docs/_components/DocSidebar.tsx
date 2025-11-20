"use client";

// React & Next.js
import React from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';

// Third-party libraries
import {
  ChevronRight,
  Search
} from 'lucide-react';

// Constants
import { DOCS_NAVIGATION } from '@/constants';

interface DocSidebarProps {
  isOpen: boolean;
  onClose: () => void;
}

export const DocSidebar: React.FC<DocSidebarProps> = ({ isOpen, onClose }) => {
  const pathname = usePathname();
  const currentPath = pathname.replace('/docs/', '') || 'getting-started/installation';

  return (
    <>
      {/* Mobile Backdrop */}
      {isOpen && (
        <div 
          className="fixed inset-0 bg-black/60 z-40 lg:hidden backdrop-blur-sm"
          onClick={onClose}
        />
      )}

      {/* Sidebar Container */}
      <aside className={`
        fixed top-0 left-0 bottom-0 w-[280px] bg-slate-900 border-r border-slate-800 z-50
        transform transition-transform duration-300 ease-in-out overflow-y-auto
        ${isOpen ? 'translate-x-0' : '-translate-x-full'}
        lg:translate-x-0 lg:top-16 lg:h-[calc(100vh-4rem)]
      `}>
        <div className="p-4 sticky top-0 bg-slate-900 z-10">
           <div className="relative">
             <Search className="absolute left-3 top-2.5 text-slate-500" size={16} />
             <input 
               type="text" 
               placeholder="Search docs..." 
               className="w-full bg-slate-800 border border-slate-700 text-slate-200 text-sm rounded-lg pl-9 pr-4 py-2 focus:outline-none focus:ring-2 focus:ring-primary/50"
             />
             <div className="absolute right-2 top-2.5 px-1.5 py-0.5 rounded bg-slate-700 text-[10px] text-slate-400 font-medium">⌘K</div>
           </div>
        </div>

        <nav className="px-4 pb-8">
          {DOCS_NAVIGATION.map((section, idx) => (
            <div key={idx} className="mb-6">
              <h3 className="text-xs font-bold text-slate-500 uppercase tracking-wider mb-3 px-2">
                {section.title}
              </h3>
              <ul className="space-y-1">
                {section.items.map((item, itemIdx) => {
                  const isActive = currentPath === item.slug;
                  return (
                    <li key={itemIdx}>
                      <Link
                        href={`/docs/${item.slug}`}
                        onClick={onClose}
                        className={`
                          group flex items-center justify-between px-2 py-1.5 rounded-md text-sm font-medium transition-colors
                          ${isActive
                            ? 'bg-primary/10 text-emerald-400'
                            : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800'}
                        `}
                      >
                        {item.title}
                        {item.badge && (
                          <span className="text-[10px] bg-indigo-500/20 text-indigo-300 px-1.5 py-0.5 rounded border border-indigo-500/30">
                            {item.badge}
                          </span>
                        )}
                      </Link>
                    </li>
                  );
                })}
              </ul>
            </div>
          ))}
        </nav>
      </aside>
    </>
  );
};