"use client";

import React, { useState, useEffect } from 'react';
import Link from 'next/link';
import { Terminal as TerminalIcon, Menu, X, ChevronRight, AlertTriangle, Github } from 'lucide-react';
import { DocSidebar } from '@/components/docs/DocSidebar';
import { TableOfContents } from '@/components/docs/TableOfContents';
import { CodeBlock } from '@/components/docs/CodeBlock';
import { Button } from '@/components/ui/Button';

const DocsHomePage: React.FC = () => {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const currentSlug = 'docs';

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
    <div className="min-h-screen bg-background text-slate-300">
      {/* Docs Navbar */}
      <nav className="sticky top-0 z-50 w-full bg-slate-900/95 backdrop-blur border-b border-slate-800">
        <div className="w-full px-4 h-16 flex items-center justify-between">
          <div className="flex items-center gap-4">
            <button
              onClick={() => setSidebarOpen(!sidebarOpen)}
              className="lg:hidden p-2 text-slate-400 hover:text-white"
            >
              {sidebarOpen ? <X /> : <Menu />}
            </button>
            <Link href="/" className="flex items-center gap-2 hover:opacity-80 transition-opacity">
               <div className="w-8 h-8 bg-primary rounded-lg flex items-center justify-center">
                 <TerminalIcon className="text-slate-900" size={20} />
               </div>
               <span className="text-xl font-bold tracking-tight text-slate-100 hidden sm:inline">Chauffeur</span>
               <span className="text-xs bg-slate-800 text-slate-400 px-2 py-0.5 rounded-full ml-2 border border-slate-700">Docs</span>
            </Link>
          </div>
          <div className="flex items-center gap-3">
             <a href="https://github.com" target="_blank" rel="noreferrer" className="text-slate-400 hover:text-white transition-colors">
                <Github size={20} />
             </a>
             <Button size="sm" variant="outline" className="hidden sm:flex">v0.1.0-beta</Button>
          </div>
        </div>
      </nav>

      <div className="flex max-w-[1440px] mx-auto">
        {/* Left Sidebar */}
        <DocSidebar isOpen={sidebarOpen} onClose={() => setSidebarOpen(false)} />

        {/* Main Content Area */}
        <main className="flex-1 min-w-0 lg:pl-[280px] xl:pr-[280px]">
          <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-10">
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
          </div>
        </main>

        {/* Right TOC (Desktop Only) */}
        <TableOfContents />
      </div>
    </div>
  );
};

export default DocsHomePage;