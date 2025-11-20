import React from 'react';

export const ArchitectureDiagram: React.FC = () => {
  return (
    <div className="relative w-full aspect-[16/9] bg-slate-900 rounded-xl border border-slate-700 overflow-hidden p-8 flex items-center justify-center">
      {/* Background Pattern */}
      <div className="absolute inset-0 opacity-20" 
           style={{backgroundImage: 'radial-gradient(#6366F1 1px, transparent 1px)', backgroundSize: '20px 20px'}}></div>

      <svg className="w-full h-full max-w-2xl" viewBox="0 0 800 400" fill="none" xmlns="http://www.w3.org/2000/svg">
        {/* Host System Layer */}
        <rect x="50" y="50" width="200" height="300" rx="10" fill="#1E293B" stroke="#334155" strokeWidth="2" />
        <text x="150" y="90" textAnchor="middle" fill="#94A3B8" fontSize="14" fontWeight="bold">HOST (LINUX)</text>
        
        {/* Incoming Request */}
        <path d="M 0 200 L 50 200" stroke="#10B981" strokeWidth="2" strokeDasharray="4 4">
           <animate attributeName="stroke-dashoffset" from="8" to="0" dur="1s" repeatCount="indefinite" />
        </path>

        {/* Chauffeur Core */}
        <rect x="80" y="150" width="140" height="100" rx="8" fill="#0F172A" stroke="#10B981" strokeWidth="2" />
        <text x="150" y="190" textAnchor="middle" fill="#10B981" fontSize="16" fontWeight="bold">Nginx</text>
        <text x="150" y="210" textAnchor="middle" fill="#64748B" fontSize="12">(Reverse Proxy)</text>

        {/* Connection Lines */}
        <path d="M 220 200 L 350 120" stroke="#64748B" strokeWidth="2" />
        <path d="M 220 200 L 350 280" stroke="#64748B" strokeWidth="2" />

        {/* Project A */}
        <g transform="translate(350, 70)">
           <rect width="200" height="100" rx="8" fill="#1E293B" stroke="#6366F1" strokeWidth="2" />
           <text x="20" y="30" fill="#F1F5F9" fontSize="14" fontWeight="bold">Project A</text>
           <text x="20" y="50" fill="#94A3B8" fontSize="12">shop.test</text>
           <rect x="120" y="60" width="60" height="24" rx="4" fill="#6366F1" fillOpacity="0.2" />
           <text x="150" y="76" textAnchor="middle" fill="#818CF8" fontSize="11" fontWeight="bold">PHP 8.3</text>
        </g>

        {/* Project B */}
        <g transform="translate(350, 230)">
           <rect width="200" height="100" rx="8" fill="#1E293B" stroke="#F59E0B" strokeWidth="2" />
           <text x="20" y="30" fill="#F1F5F9" fontSize="14" fontWeight="bold">Project B</text>
           <text x="20" y="50" fill="#94A3B8" fontSize="12">legacy.test</text>
           <rect x="120" y="60" width="60" height="24" rx="4" fill="#F59E0B" fillOpacity="0.2" />
           <text x="150" y="76" textAnchor="middle" fill="#FBBF24" fontSize="11" fontWeight="bold">PHP 7.4</text>
        </g>
      </svg>

      {/* Floating labels */}
      <div className="absolute top-4 right-4 bg-slate-800 px-3 py-1 rounded-full border border-slate-600 text-xs text-slate-300">
        ~/.chauffeur
      </div>
    </div>
  );
};