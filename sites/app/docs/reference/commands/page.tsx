"use client";

import React, { useEffect } from 'react';
import Link from 'next/link';
import { ChevronRight } from 'lucide-react';
import { TableOfContents } from '@/components/docs/TableOfContents';
import { CodeBlock } from '@/components/docs/CodeBlock';

export default function CommandsPage() {
  const currentSlug = 'reference/commands';

  useEffect(() => {
    window.scrollTo(0, 0);
  }, []);

  // Helper for safe scrolling
  const scrollToId = (e: React.MouseEvent, id: string) => {
    e.preventDefault();
    const element = document.getElementById(id);
    if (element) {
      const yOffset = -100;
      const y = element.getBoundingClientRect().top + window.pageYOffset + yOffset;
      window.scrollTo({ top: y, behavior: 'smooth' });
    }
  };

  return (
    <>
      {/* Breadcrumbs */}
      <div className="flex items-center gap-2 text-sm text-slate-500 mb-8 overflow-x-auto whitespace-nowrap pb-2">
        <Link href="/" className="hover:text-primary transition-colors">Home</Link>
        <ChevronRight size={14} />
        <Link href="/docs" className="hover:text-primary transition-colors">Documentation</Link>
        <ChevronRight size={14} />
        <span className="text-slate-200 capitalize">{currentSlug.split('/')[0]}</span>
        <ChevronRight size={14} />
        <span className="text-slate-200 capitalize">{currentSlug.split('/')[1]?.replace('-', ' ')}</span>
      </div>

      {/* Content */}
      <div className="prose prose-invert prose-slate max-w-none space-y-8 animate-fade-in">
        <div>
          <h1 className="text-4xl font-bold text-white mb-4">Command Reference</h1>
          <p className="text-lg text-slate-400 leading-relaxed">
            Comprehensive guide to the Chauffeur CLI.
          </p>
        </div>

        <section id="link">
          <h2 className="text-2xl font-bold text-white mb-4 font-mono group flex items-center gap-2" id="link">
            chauf link
            <Link href="#link" onClick={(e) => scrollToId(e, 'link')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">Registers the current working directory as a Chauffeur site.</p>

          <h3 className="text-lg font-semibold text-slate-200 mb-2">Usage</h3>
          <CodeBlock code="chauf link [name] [flags]" />

          <h3 className="text-lg font-semibold text-slate-200 mb-2">Flags</h3>
          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm border-collapse">
              <thead>
                <tr className="border-b border-slate-700 bg-slate-800/50">
                  <th className="p-3 font-mono text-emerald-400">Flag</th>
                  <th className="p-3 text-slate-300">Description</th>
                  <th className="p-3 text-slate-300">Default</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-800">
                <tr>
                  <td className="p-3 font-mono text-slate-300">--secure</td>
                  <td className="p-3 text-slate-400">Generate SSL certificate</td>
                  <td className="p-3 text-slate-500">false</td>
                </tr>
                <tr>
                   <td className="p-3 font-mono text-slate-300">--php</td>
                   <td className="p-3 text-slate-400">Specify PHP version (e.g. 8.1)</td>
                   <td className="p-3 text-slate-500">global</td>
                </tr>
                 <tr>
                   <td className="p-3 font-mono text-slate-300">--force</td>
                   <td className="p-3 text-slate-400">Overwrite existing config</td>
                   <td className="p-3 text-slate-500">false</td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>

        <section id="isolate">
          <h2 className="text-2xl font-bold text-white mb-4 font-mono mt-12 group flex items-center gap-2" id="isolate">
            chauf isolate
            <Link href="#isolate" onClick={(e) => scrollToId(e, 'isolate')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">Isolates the current site to a specific PHP version, separate from the global default.</p>
          <CodeBlock code="chauf isolate 7.4" />
        </section>

        <section id="secure">
          <h2 className="text-2xl font-bold text-white mb-4 font-mono mt-12 group flex items-center gap-2" id="secure">
            chauf secure
            <Link href="#secure" onClick={(e) => scrollToId(e, 'secure')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">Generate and install SSL certificates for an existing site.</p>
          <CodeBlock code="chauf secure [site-name]" />
        </section>

        <section id="use">
          <h2 className="text-2xl font-bold text-white mb-4 font-mono mt-12 group flex items-center gap-2" id="use">
            chauf use
            <Link href="#use" onClick={(e) => scrollToId(e, 'use')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">Set the global default PHP version for new sites.</p>
          <CodeBlock code="chauf use 8.2" />
        </section>

        <section id="start">
          <h2 className="text-2xl font-bold text-white mb-4 font-mono mt-12 group flex items-center gap-2" id="start">
            chauf start
            <Link href="#start" onClick={(e) => scrollToId(e, 'start')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">Start nginx/PHP-FPM plus dnsmasq validation.</p>
          <CodeBlock code="chauf start [--project <path>] [--all] [--dry-run]" />
          <h3 className="text-lg font-semibold text-slate-200 mb-2 mt-4">Flags</h3>
          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm border-collapse">
              <thead>
                <tr className="border-b border-slate-700 bg-slate-800/50">
                  <th className="p-3 font-mono text-emerald-400">Flag</th>
                  <th className="p-3 text-slate-300">Description</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-800">
                <tr>
                  <td className="p-3 font-mono text-slate-300">--project</td>
                  <td className="p-3 text-slate-400">Start specific project services</td>
                </tr>
                <tr>
                  <td className="p-3 font-mono text-slate-300">--all</td>
                  <td className="p-3 text-slate-400">Start all registered projects</td>
                </tr>
                <tr>
                  <td className="p-3 font-mono text-slate-300">--dry-run</td>
                  <td className="p-3 text-slate-400">Show what would be started</td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>

        <section id="stop">
          <h2 className="text-2xl font-bold text-white mb-4 font-mono mt-12 group flex items-center gap-2" id="stop">
            chauf stop
            <Link href="#stop" onClick={(e) => scrollToId(e, 'stop')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">Stop services and clean port-forward rules.</p>
          <CodeBlock code="chauf stop [--project <path>] [--all] [--dry-run]" />
        </section>

        <section id="restart">
          <h2 className="text-2xl font-bold text-white mb-4 font-mono mt-12 group flex items-center gap-2" id="restart">
            chauf restart
            <Link href="#restart" onClick={(e) => scrollToId(e, 'restart')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">Restart services (equivalent to stop then start, preserves configuration).</p>
          <CodeBlock code="chauf restart [--project <slug>] [--all] [--dry-run]" />
        </section>

        <section id="status">
          <h2 className="text-2xl font-bold text-white mb-4 font-mono mt-12 group flex items-center gap-2" id="status">
            chauf status
            <Link href="#status" onClick={(e) => scrollToId(e, 'status')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">Show status for global or per-project services.</p>
          <CodeBlock code="chauf status [service-type] [--project] [--detail] [-v]" />
        </section>

        <section id="unlink">
          <h2 className="text-2xl font-bold text-white mb-4 font-mono mt-12 group flex items-center gap-2" id="unlink">
            chauf unlink
            <Link href="#unlink" onClick={(e) => scrollToId(e, 'unlink')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">Remove registrations or specific aliases. Defaults to current dir.</p>
          <CodeBlock code="chauf unlink [--slug] [--site] [--project] [--alias] [--all] [--force]" />
        </section>

        <section id="php-install">
          <h2 className="text-2xl font-bold text-white mb-4 font-mono mt-12 group flex items-center gap-2" id="php-install">
            chauf php install
            <Link href="#php-install" onClick={(e) => scrollToId(e, 'php-install')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">Build/install PHP runtime under workspace.</p>
          <CodeBlock code="chauf php install <version> [--force] [--no-ext] [--from <source>]" />
        </section>

        <section id="logs">
          <h2 className="text-2xl font-bold text-white mb-4 font-mono mt-12 group flex items-center gap-2" id="logs">
            chauf logs
            <Link href="#logs" onClick={(e) => scrollToId(e, 'logs')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">View and follow service logs with interactive version selection.</p>
          <CodeBlock code="chauf logs [service] [version] [--follow] [-f] [--lines] [-n] [--level] [--context] [-c]" />
        </section>

        <section id="clean">
          <h2 className="text-2xl font-bold text-white mb-4 font-mono mt-12 group flex items-center gap-2" id="clean">
            chauf clean
            <Link href="#clean" onClick={(e) => scrollToId(e, 'clean')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">Clean workspace files with file size display and accurate reporting.</p>
          <CodeBlock code="chauf clean [target] [--dry-run] [--force] [--older-than] [--keep-versions] [--what]" />
        </section>

        {/* Page Footer Navigation */}
        <div className="mt-16 pt-8 border-t border-slate-800 flex justify-between">
           <div className="text-left">
             <div className="text-xs text-slate-500 mb-1">Previous</div>
             <Link href="/docs/getting-started/architecture" className="text-primary hover:underline">Architecture</Link>
           </div>
           <div className="text-right">
             <div className="text-xs text-slate-500 mb-1">Next</div>
             <Link href="/docs" className="text-primary hover:underline">Documentation Home</Link>
           </div>
        </div>
      </div>

      {/* Right TOC (Desktop Only) */}
      <div className="hidden xl:block">
        <TableOfContents />
      </div>
    </>
  );
}