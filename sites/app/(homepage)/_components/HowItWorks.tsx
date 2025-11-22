// React
import React, { useState, useEffect } from 'react';

// Third-party libraries
import {
  FolderTree,
  Shield,
  Cpu,
  Network,
  Lock,
  FileJson,
  CheckCircle2,
  XCircle,
  Server,
  ArrowRight,
  AlertTriangle
} from 'lucide-react';

type TabId = 'isolation' | 'ssl' | 'performance' | 'ports';

const TABS: { id: TabId; label: string; icon: React.ElementType }[] = [
  { id: 'isolation', label: 'Workspace Isolation', icon: FolderTree },
  { id: 'ssl', label: 'SSL & Domains', icon: Shield },
  { id: 'performance', label: 'Resource Efficiency', icon: Cpu },
  { id: 'ports', label: 'Port Management', icon: Network },
];

export const HowItWorks: React.FC = () => {
  const [activeTab, setActiveTab] = useState<TabId>('isolation');
  const [demoDomain, setDemoDomain] = useState('my-app');
  const [isScanning, setIsScanning] = useState(false);

  // Reset scanning state when switching tabs
  useEffect(() => {
    if (activeTab === 'ports') {
      setIsScanning(true);
      const timer = setTimeout(() => setIsScanning(false), 1500);
      return () => clearTimeout(timer);
    }
  }, [activeTab]);

  return (
    <section id="how-it-works" className="py-24 bg-background relative overflow-hidden border-t border-slate-800/50">
      <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_bottom,_var(--tw-gradient-stops))] from-indigo-900/10 via-transparent to-transparent pointer-events-none" />

      <div className="container mx-auto px-4 md:px-6 relative z-10">
        {/* Header */}
        <div className="text-center max-w-3xl mx-auto mb-16">
          <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-indigo-500/10 border border-indigo-500/20 text-indigo-400 text-xs font-medium mb-4">
            <span className="w-2 h-2 rounded-full bg-indigo-500 animate-pulse" />
            System Architecture
          </div>
          <h2 className="text-3xl md:text-4xl font-bold text-white mb-6">
            Linux Development, <span className="text-indigo-400">Reimagined</span>
          </h2>
          <p className="text-slate-400 text-lg leading-relaxed">
            Built by developers who understand Linux architecture. Chauffeur combines the simplicity
            of Valet with the power of native Linux system design—without the pollution.
          </p>
        </div>

        {/* Interactive Tabs */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-12 max-w-4xl mx-auto">
          {TABS.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`
                flex flex-col items-center justify-center p-4 rounded-xl border transition-all duration-300 hover:cursor-pointer
                ${activeTab === tab.id
                  ? 'bg-surface border-indigo-500/50 shadow-lg shadow-indigo-500/10 scale-105'
                  : 'bg-surface/50 border-slate-800 hover:border-slate-700 hover:bg-surface/80 text-slate-500'}
              `}
            >
              <tab.icon className={`mb-2 ${activeTab === tab.id ? 'text-indigo-400' : 'text-slate-500'}`} size={24} />
              <span className={`text-sm font-medium ${activeTab === tab.id ? 'text-white' : 'text-slate-500'}`}>
                {tab.label}
              </span>
            </button>
          ))}
        </div>

        {/* Visualization Area */}
        <div className="max-w-5xl mx-auto bg-surface border border-slate-800 rounded-2xl overflow-hidden shadow-2xl min-h-[500px] flex flex-col md:flex-row">

          {/* Left: Explanation */}
          <div className="w-full md:w-1/3 p-8 border-b md:border-b-0 md:border-r bg-slate-900/50">
            {activeTab === 'isolation' && (
              <div className="space-y-6 animate-fade-in">
                <h3 className="text-xl font-bold text-white">Complete Isolation</h3>
                <p className="text-slate-400 text-sm leading-relaxed">
                  Stop polluting your system folders. Chauffeur creates a self-contained environment in your home directory.
                </p>
                <ul className="space-y-4">
                  <li className="flex gap-3 text-sm text-slate-300">
                    <CheckCircle2 className="text-emerald-500 shrink-0" size={18} />
                    <span>No root permissions required for daily use</span>
                  </li>
                  <li className="flex gap-3 text-sm text-slate-300">
                    <CheckCircle2 className="text-emerald-500 shrink-0" size={18} />
                    <span>Self-contained binaries and configs</span>
                  </li>
                  <li className="flex gap-3 text-sm text-slate-300">
                    <CheckCircle2 className="text-emerald-500 shrink-0" size={18} />
                    <span>Zero impact on system-installed packages</span>
                  </li>
                </ul>
              </div>
            )}

            {activeTab === 'ssl' && (
              <div className="space-y-6 animate-fade-in">
                <h3 className="text-xl font-bold text-white">Smart SSL & Routing</h3>
                <p className="text-slate-400 text-sm leading-relaxed">
                  Automatic certificate generation for main domains and all aliases. Nginx is configured dynamically.
                </p>
                <div className="space-y-4">
                  <div className="bg-slate-800 p-4 rounded-lg border border-slate-700">
                    <label className="text-xs text-slate-500 mb-2 block">Project Name Input</label>
                    <input 
                      type="text" 
                      value={demoDomain}
                      onChange={(e) => setDemoDomain(e.target.value)}
                      className="w-full bg-slate-900 border border-slate-600 rounded px-3 py-2 text-sm text-white focus:outline-none focus:border-indigo-500"
                    />
                    <p className="text-xs text-slate-500 mt-2">Try typing a name!</p>
                  </div>
                </div>
              </div>
            )}

            {activeTab === 'performance' && (
              <div className="space-y-6 animate-fade-in">
                <h3 className="text-xl font-bold text-white">Resource Efficiency</h3>
                <p className="text-slate-400 text-sm leading-relaxed">
                  Forget the 4GB overhead of Laravel Homestead. Chauffeur uses native processes with smart shared pools.
                </p>
                <ul className="space-y-4">
                   <li className="flex gap-3 text-sm text-slate-300">
                    <CheckCircle2 className="text-emerald-500 shrink-0" size={18} />
                    <span>Shared PHP-FPM pools by default</span>
                  </li>
                  <li className="flex gap-3 text-sm text-slate-300">
                    <CheckCircle2 className="text-emerald-500 shrink-0" size={18} />
                    <span>Optional isolation for specific apps</span>
                  </li>
                   <li className="flex gap-3 text-sm text-slate-300">
                    <CheckCircle2 className="text-emerald-500 shrink-0" size={18} />
                    <span>Idle processes sleep automatically</span>
                  </li>
                </ul>
              </div>
            )}

            {activeTab === 'ports' && (
              <div className="space-y-6 animate-fade-in">
                <h3 className="text-xl font-bold text-white">Conflict Resolution</h3>
                <p className="text-slate-400 text-sm leading-relaxed">
                  Port 80 busy? No problem. Chauffeur automatically scans for available ports and maps them.
                </p>
                <div className="bg-slate-800/50 p-4 rounded-lg border border-slate-700 text-xs font-mono space-y-2">
                  <div className="flex justify-between">
                    <span className="text-slate-400">Standard HTTP</span>
                    <span className="text-slate-200">80 → {isScanning ? '...' : '8080'}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-slate-400">Standard HTTPS</span>
                    <span className="text-slate-200">443 → {isScanning ? '...' : '8443'}</span>
                  </div>
                  <div className="flex justify-between border-t border-slate-700 pt-2 mt-2">
                    <span className="text-emerald-400">Strategy</span>
                    <span className="text-emerald-400">iptables REDIRECT</span>
                  </div>
                </div>
              </div>
            )}
          </div>

          {/* Right: Visual Interactive Demo */}
          <div className="w-full md:w-2/3 bg-[#0F172A] p-8 flex items-center justify-center relative">
            
            {/* VIEW: ISOLATION */}
            {activeTab === 'isolation' && (
              <div className="w-full grid grid-cols-2 gap-8 animate-fade-in">
                {/* Traditional Way */}
                <div className="space-y-4 opacity-50">
                  <div className="flex items-center gap-2 text-red-400 mb-4">
                    <XCircle size={16} />
                    <span className="text-xs font-bold uppercase tracking-wider">System Pollution</span>
                  </div>
                  <div className="font-mono text-xs space-y-2 text-slate-500 border-l-2 border-red-900/30 pl-4">
                    <div className="flex items-center gap-2">
                      <FolderTree size={14} /> /etc/nginx <span className="text-red-900 text-[10px] px-1 rounded bg-red-900/20 border border-red-900/30">ROOT</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <FolderTree size={14} /> /etc/php/8.1 <span className="text-red-900 text-[10px] px-1 rounded bg-red-900/20 border border-red-900/30">ROOT</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <FolderTree size={14} /> /var/www/html <span className="text-red-900 text-[10px] px-1 rounded bg-red-900/20 border border-red-900/30">ROOT</span>
                    </div>
                  </div>
                </div>

                {/* Chauffeur Way */}
                <div className="space-y-4">
                  <div className="flex items-center gap-2 text-emerald-400 mb-4">
                    <CheckCircle2 size={16} />
                    <span className="text-xs font-bold uppercase tracking-wider">Chauffeur Workspace</span>
                  </div>
                   <div className="font-mono text-xs space-y-3 text-slate-300 border-l-2 border-emerald-500/30 pl-4">
                    <div className="flex items-center gap-2 text-emerald-200">
                      <FolderTree size={14} /> ~/.chauffeur/
                    </div>
                    <div className="pl-4 space-y-2">
                       <div className="flex items-center gap-2">
                         <span className="text-slate-600">├──</span> <FileJson size={12} className="text-indigo-400"/> nginx/
                       </div>
                       <div className="flex items-center gap-2">
                         <span className="text-slate-600">├──</span> <FileJson size={12} className="text-indigo-400"/> php/
                         <span className="text-slate-500 text-[10px]">(8.1, 8.2, 8.3)</span>
                       </div>
                       <div className="flex items-center gap-2">
                         <span className="text-slate-600">├──</span> <FileJson size={12} className="text-indigo-400"/> projects/
                         <span className="text-emerald-400 text-[10px] px-1 rounded bg-emerald-500/10 border border-emerald-500/20">USER</span>
                       </div>
                       <div className="flex items-center gap-2">
                         <span className="text-slate-600">└──</span> <FileJson size={12} className="text-indigo-400"/> ssl/
                       </div>
                    </div>
                  </div>
                </div>
              </div>
            )}

            {/* VIEW: SSL */}
            {activeTab === 'ssl' && (
              <div className="w-full space-y-6 animate-fade-in">
                <div className="flex items-center justify-center gap-8">
                   {/* Project File */}
                   <div className="w-32 bg-slate-800 border border-slate-700 rounded p-3">
                      <div className="text-[10px] text-slate-500 mb-2 flex items-center gap-1">
                        <FileJson size={10} /> project.yaml
                      </div>
                      <div className="text-[9px] font-mono text-slate-300 leading-relaxed">
                        site: <span className="text-indigo-400">{demoDomain || 'app'}.test</span><br/>
                        ssl: <span className="text-emerald-400">true</span>
                      </div>
                   </div>

                   <ArrowRight className="text-slate-600" />

                   {/* Certificate */}
                   <div className="w-40 bg-slate-800 border border-slate-700 rounded p-3 relative overflow-hidden">
                      <div className="absolute top-0 right-0 p-1">
                         <Lock size={12} className="text-emerald-400" />
                      </div>
                      <div className="text-[10px] text-slate-500 mb-2 flex items-center gap-1">
                        <Shield size={10} /> Certificate
                      </div>
                      <div className="text-[9px] font-mono text-slate-300 leading-relaxed">
                        CN: {demoDomain || 'app'}.test<br/>
                        SANs: *.{demoDomain || 'app'}.test
                        <div className="mt-1 text-emerald-500 text-[8px] flex items-center gap-1">
                          <CheckCircle2 size={8} /> Trusted
                        </div>
                      </div>
                   </div>
                </div>

                {/* Nginx Preview */}
                <div className="bg-[#1e1e1e] rounded-lg border border-slate-800 p-4 font-mono text-xs text-slate-400 relative">
                   <div className="absolute top-2 right-2 text-[10px] text-slate-600">generated nginx.conf</div>
                   <p>server {'{'}</p>
                   <p className="pl-4">listen 443 ssl http2;</p>
                   <p className="pl-4">server_name <span className="text-indigo-400">{demoDomain || 'app'}.test</span>;</p>
                   <p className="pl-4">ssl_certificate <span className="text-emerald-400">.../{demoDomain || 'app'}.crt</span>;</p>
                   <p className="pl-4">ssl_certificate_key <span className="text-emerald-400">.../{demoDomain || 'app'}.key</span>;</p>
                   <p className="pl-4">...</p>
                   <p>{'}'}</p>
                </div>
              </div>
            )}

            {/* VIEW: PERFORMANCE */}
            {activeTab === 'performance' && (
              <div className="w-full max-w-md space-y-6 animate-fade-in">
                 {/* Docker */}
                 <div>
                    <div className="flex justify-between text-xs mb-1">
                      <span className="text-slate-400">Laravel Homestead (5 Projects)</span>
                      <span className="text-red-400">~2.4 GB</span>
                    </div>
                    <div className="h-2 bg-slate-800 rounded-full overflow-hidden">
                       <div className="h-full w-full bg-red-500/50"></div>
                    </div>
                 </div>

                 {/* Manual */}
                 <div>
                    <div className="flex justify-between text-xs mb-1">
                      <span className="text-slate-400">Manual Setup (Apache/Nginx)</span>
                      <span className="text-amber-400">~450 MB</span>
                    </div>
                    <div className="h-2 bg-slate-800 rounded-full overflow-hidden">
                       <div className="h-full w-[20%] bg-amber-500/50"></div>
                    </div>
                 </div>

                 {/* Chauffeur */}
                 <div>
                    <div className="flex justify-between text-xs mb-1">
                      <span className="text-white font-medium">Chauffeur (Shared FPM)</span>
                      <span className="text-emerald-400 font-bold">~65 MB</span>
                    </div>
                    <div className="h-2 bg-slate-800 rounded-full overflow-hidden relative">
                       <div className="h-full w-[3%] bg-emerald-500 shadow-[0_0_10px_rgba(16,185,129,0.5)]"></div>
                    </div>
                    <p className="text-[10px] text-slate-500 mt-2">
                      * Chauffeur shares the PHP-FPM process pool across compatible projects, creating new pools only when version isolation is explicitly requested.
                    </p>
                 </div>
              </div>
            )}

            {/* VIEW: PORTS */}
            {activeTab === 'ports' && (
              <div className="w-full max-w-md animate-fade-in">
                 <div className="relative border border-slate-700 rounded-xl bg-slate-900 p-6">
                    <div className="absolute -top-3 left-4 bg-slate-900 px-2 text-xs text-slate-400">
                      Port Resolution Logic
                    </div>
                    
                    <div className="space-y-4">
                       {/* Port 80 Check */}
                       <div className="flex items-center justify-between p-3 bg-slate-800/50 rounded-lg">
                          <div className="flex items-center gap-3">
                             <div className={`w-2 h-2 rounded-full ${isScanning ? 'bg-yellow-400 animate-pulse' : 'bg-red-500'}`} />
                             <span className="text-sm text-slate-300">Port 80 (System)</span>
                          </div>
                          {isScanning ? (
                             <span className="text-xs text-slate-500">Checking...</span>
                          ) : (
                             <span className="text-xs text-red-400 flex items-center gap-1"><AlertTriangle size={12}/> Occupied</span>
                          )}
                       </div>

                       <div className="flex justify-center">
                          <ArrowRight className="rotate-90 text-slate-600" size={16} />
                       </div>

                       {/* Port 8080 Check */}
                       <div className="flex items-center justify-between p-3 bg-slate-800/50 rounded-lg border border-emerald-500/30">
                          <div className="flex items-center gap-3">
                             <div className={`w-2 h-2 rounded-full ${isScanning ? 'bg-yellow-400 animate-pulse' : 'bg-emerald-500'}`} />
                             <span className="text-sm text-white">Port 8080 (Chauffeur)</span>
                          </div>
                           {isScanning ? (
                             <span className="text-xs text-slate-500">Checking...</span>
                          ) : (
                             <span className="text-xs text-emerald-400 flex items-center gap-1"><CheckCircle2 size={12}/> Available</span>
                          )}
                       </div>

                       <div className="mt-4 pt-4 border-t border-slate-700">
                          <div className="text-xs font-mono text-slate-500 mb-2">$ sudo iptables -t nat ...</div>
                          <div className="text-xs text-slate-400">
                             Redirecting <span className="text-white">80 → 8080</span> and <span className="text-white">443 → 8443</span>
                          </div>
                       </div>
                    </div>
                 </div>
              </div>
            )}

          </div>
        </div>
      </div>
    </section>
  );
};