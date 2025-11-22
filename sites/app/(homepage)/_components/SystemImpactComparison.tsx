import React from 'react';
import { Server, Cpu, HardDrive, Zap, Shield, Package } from 'lucide-react';

export const SystemImpactComparison: React.FC = () => {
  return (
    <section className="py-24 bg-slate-900 border-y border-slate-800/50">
      <div className="container mx-auto px-4 md:px-6">
        {/* Header */}
        <div className="text-center max-w-3xl mx-auto mb-16">
          <h2 className="text-3xl md:text-4xl font-bold text-white mb-4">
            System Impact Comparison
          </h2>
          <p className="text-lg text-slate-400">
            See how Chauffeur compares to traditional development environments in resource usage and performance.
          </p>
        </div>

        {/* Comparison Table */}
        <div className="max-w-6xl mx-auto">
          <div className="overflow-x-auto">
            <table className="w-full bg-surface border border-slate-800 rounded-2xl overflow-hidden">
              <thead>
                <tr className="bg-slate-900/50 border-b border-slate-700">
                  <th className="text-left px-6 py-4 text-sm font-semibold text-slate-200">Aspect</th>
                  <th className="text-center px-6 py-4 text-sm font-semibold text-emerald-400">Chauffeur</th>
                  <th className="text-center px-6 py-4 text-sm font-semibold text-slate-400">Laravel Homestead</th>
                  <th className="text-center px-6 py-4 text-sm font-semibold text-slate-400">Manual Setup</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-700">
                <tr className="hover:bg-slate-800/50 transition-colors">
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-3">
                      <Cpu className="text-slate-500" size={20} />
                      <div>
                        <div className="text-sm font-medium text-white">Memory Usage</div>
                        <div className="text-xs text-slate-500">5 active projects</div>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4 text-center">
                    <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                      ~65 MB
                    </span>
                  </td>
                  <td className="px-6 py-4 text-center">
                    <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-red-500/10 text-red-400 border border-red-500/20">
                      ~4 GB
                    </span>
                  </td>
                  <td className="px-6 py-4 text-center">
                    <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-amber-500/10 text-amber-400 border border-amber-500/20">
                      ~450 MB
                    </span>
                  </td>
                </tr>

                <tr className="hover:bg-slate-800/50 transition-colors">
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-3">
                      <Zap className="text-slate-500" size={20} />
                      <div>
                        <div className="text-sm font-medium text-white">Startup Time</div>
                        <div className="text-xs text-slate-500">Cold boot</div>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4 text-center">
                    <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                      &lt;2s
                    </span>
                  </td>
                  <td className="px-6 py-4 text-center">
                    <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-red-500/10 text-red-400 border border-red-500/20">
                      90-180s
                    </span>
                  </td>
                  <td className="px-6 py-4 text-center">
                    <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-amber-500/10 text-amber-400 border border-amber-500/20">
                      ~5s
                    </span>
                  </td>
                </tr>

                <tr className="hover:bg-slate-800/50 transition-colors">
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-3">
                      <HardDrive className="text-slate-500" size={20} />
                      <div>
                        <div className="text-sm font-medium text-white">Disk Space</div>
                        <div className="text-xs text-slate-500">Base installation</div>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4 text-center">
                    <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                      ~50 MB
                    </span>
                  </td>
                  <td className="px-6 py-4 text-center">
                    <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-red-500/10 text-red-400 border border-red-500/20">
                      ~2 GB
                    </span>
                  </td>
                  <td className="px-6 py-4 text-center">
                    <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-amber-500/10 text-amber-400 border border-amber-500/20">
                      ~100 MB
                    </span>
                  </td>
                </tr>

                <tr className="hover:bg-slate-800/50 transition-colors">
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-3">
                      <Package className="text-slate-500" size={20} />
                      <div>
                        <div className="text-sm font-medium text-white">Isolation</div>
                        <div className="text-xs text-slate-500">Process separation</div>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4 text-center">
                    <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                      Native
                    </span>
                  </td>
                  <td className="px-6 py-4 text-center">
                    <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                      Container
                    </span>
                  </td>
                  <td className="px-6 py-4 text-center">
                    <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-red-500/10 text-red-400 border border-red-500/20">
                      None
                    </span>
                  </td>
                </tr>

                <tr className="hover:bg-slate-800/50 transition-colors">
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-3">
                      <Server className="text-slate-500" size={20} />
                      <div>
                        <div className="text-sm font-medium text-white">System Access</div>
                        <div className="text-xs text-slate-500">Root required</div>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4 text-center">
                    <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                      No
                    </span>
                  </td>
                  <td className="px-6 py-4 text-center">
                    <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-amber-500/10 text-amber-400 border border-amber-500/20">
                      Setup only
                    </span>
                  </td>
                  <td className="px-6 py-4 text-center">
                    <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-red-500/10 text-red-400 border border-red-500/20">
                      Yes
                    </span>
                  </td>
                </tr>

                <tr className="hover:bg-slate-800/50 transition-colors">
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-3">
                      <Shield className="text-slate-500" size={20} />
                      <div>
                        <div className="text-sm font-medium text-white">SSL Management</div>
                        <div className="text-xs text-slate-500">Automatic HTTPS</div>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4 text-center">
                    <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                      Built-in
                    </span>
                  </td>
                  <td className="px-6 py-4 text-center">
                    <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-red-500/10 text-red-400 border border-red-500/20">
                      Manual
                    </span>
                  </td>
                  <td className="px-6 py-4 text-center">
                    <span className="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-red-500/10 text-red-400 border border-red-500/20">
                      Manual
                    </span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          {/* Key Benefits Summary */}
          <div className="mt-16 grid grid-cols-1 md:grid-cols-3 gap-6">
            <div className="bg-surface p-6 rounded-2xl border border-slate-800 text-center">
              <div className="w-16 h-16 bg-emerald-500/10 rounded-xl flex items-center justify-center mx-auto mb-4">
                <Cpu className="text-emerald-400" size={32} />
              </div>
              <h3 className="text-lg font-semibold text-white mb-2">97% Less Memory</h3>
              <p className="text-slate-400 text-sm">
                Shared FPM pools and native processes eliminate container overhead
              </p>
            </div>

            <div className="bg-surface p-6 rounded-2xl border border-slate-800 text-center">
              <div className="w-16 h-16 bg-indigo-500/10 rounded-xl flex items-center justify-center mx-auto mb-4">
                <Zap className="text-indigo-400" size={32} />
              </div>
              <h3 className="text-lg font-semibold text-white mb-2">15x Faster Startup</h3>
              <p className="text-slate-400 text-sm">
                No container initialization or virtualization layers to slow you down
              </p>
            </div>

            <div className="bg-surface p-6 rounded-2xl border border-slate-800 text-center">
              <div className="w-16 h-16 bg-amber-500/10 rounded-xl flex items-center justify-center mx-auto mb-4">
                <Shield className="text-amber-400" size={32} />
              </div>
              <h3 className="text-lg font-semibold text-white mb-2">Zero Configuration</h3>
              <p className="text-slate-400 text-sm">
                Automatic SSL, DNS, and routing. Just focus on your code.
              </p>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
};