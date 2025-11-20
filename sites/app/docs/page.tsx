"use client";

import React, { useEffect } from 'react';
import Link from 'next/link';
import { ChevronRight, Terminal as TerminalIcon } from 'lucide-react';
import { TableOfContents } from '@/components/docs/TableOfContents';

export default function DocsHomePage() {
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
        <span>Documentation</span>
      </div>

      {/* Content */}
      <div className="prose prose-invert prose-slate max-w-none space-y-8 animate-fade-in">
        <div>
          <h1 className="text-4xl font-bold text-white mb-4">Documentation</h1>
          <p className="text-lg text-slate-400 leading-relaxed">
            Welcome to the Chauffeur documentation. Here you'll find comprehensive guides to get you started with Chauffeur, from installation to advanced configuration.
          </p>
        </div>

        <div className="p-4 bg-blue-500/10 border border-blue-500/20 rounded-lg flex gap-3">
           <TerminalIcon className="text-blue-400 shrink-0" />
           <div className="text-sm text-blue-100/80">
             <strong>Quick Start:</strong> New to Chauffeur? Start with the <Link href="/docs/getting-started/installation" className="text-primary hover:underline">Installation guide</Link> and then create your <Link href="/docs/getting-started/first-project" className="text-primary hover:underline">First Project</Link>.
           </div>
        </div>

        <section id="getting-started">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="getting-started">
            Getting Started
            <Link href="#getting-started" onClick={(e) => scrollToId(e, 'getting-started')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">Everything you need to know to get Chauffeur up and running on your system.</p>
          <ul className="space-y-3">
            <li className="flex items-center gap-3">
              <ChevronRight size={16} className="text-slate-500" />
              <Link href="/docs/getting-started/installation" className="text-primary hover:underline">Installation</Link>
              <span className="text-slate-500">- Install Chauffeur on your Linux system</span>
            </li>
            <li className="flex items-center gap-3">
              <ChevronRight size={16} className="text-slate-500" />
              <Link href="/docs/getting-started/first-project" className="text-primary hover:underline">First Project</Link>
              <span className="text-slate-500">- Create and serve your first PHP project</span>
            </li>
            <li className="flex items-center gap-3">
              <ChevronRight size={16} className="text-slate-500" />
              <Link href="/docs/getting-started/architecture" className="text-primary hover:underline">Architecture</Link>
              <span className="text-slate-500">- Understand how Chauffeur works</span>
            </li>
          </ul>
        </section>

        <section id="reference">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="reference">
            Reference
            <Link href="#reference" onClick={(e) => scrollToId(e, 'reference')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">Comprehensive reference material for all Chauffeur features.</p>
          <ul className="space-y-3">
            <li className="flex items-center gap-3">
              <ChevronRight size={16} className="text-slate-500" />
              <Link href="/docs/reference/commands" className="text-primary hover:underline">Command Reference</Link>
              <span className="text-slate-500">- Complete CLI command documentation</span>
            </li>
          </ul>
        </section>
      </div>

      {/* Right TOC (Desktop Only) */}
      <div className="hidden xl:block">
        <TableOfContents />
      </div>
    </>
  );
}