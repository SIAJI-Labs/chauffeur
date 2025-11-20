"use client";

import React, { useEffect, useState } from 'react';
import Link from 'next/link';
import { Terminal as TerminalIcon, Menu, X, ChevronRight, Github } from 'lucide-react';
import { DocSidebar } from '@/components/docs/DocSidebar';
import { TableOfContents } from '@/components/docs/TableOfContents';
import { CodeBlock } from '@/components/docs/CodeBlock';
import { Button } from '@/components/ui/Button';

const FirstProjectPage: React.FC = () => {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const currentSlug = 'getting-started/first-project';

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
              <ChevronRight size={14} />
              <span className="text-slate-200 capitalize">{currentSlug.split('/').pop()?.replace('-', ' ')}</span>
            </div>

            {/* Content */}
            <div className="prose prose-invert prose-slate max-w-none space-y-8 animate-fade-in">
               <div>
                <h1 className="text-4xl font-bold text-white mb-4">Your First Project</h1>
                <p className="text-lg text-slate-400 leading-relaxed">
                  Let's get a simple PHP project running with SSL and a custom domain in under a minute.
                </p>
              </div>

              <section id="create-project">
                <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="create-project">
                  1. Create a Project
                  <Link href="#create-project" onClick={(e) => scrollToId(e, 'create-project')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
                </h2>
                <p className="text-slate-400 mb-4">Create a directory for your new project and add an index.php file.</p>
                <CodeBlock code={`mkdir my-website\ncd my-website\necho "<?php phpinfo();" > index.php`} />
              </section>

              <section id="link-project">
                <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="link-project">
                  2. Link the Project
                  <Link href="#link-project" onClick={(e) => scrollToId(e, 'link-project')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
                </h2>
                <p className="text-slate-400 mb-4">Run the link command to tell Chauffeur to serve this directory.</p>
                <CodeBlock code="chauf link" />
                <p className="text-slate-400">
                  By default, this will create a site at <code>http://my-website.test</code>.
                </p>
              </section>

               <section id="add-ssl">
                <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="add-ssl">
                  3. Add SSL
                  <Link href="#add-ssl" onClick={(e) => scrollToId(e, 'add-ssl')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
                </h2>
                <p className="text-slate-400 mb-4">Secure your local site with a trusted certificate.</p>
                <CodeBlock code="chauf secure" />
                <p className="text-slate-400">
                  Now visit <Link href="#" onClick={(e) => e.preventDefault()} className="text-primary hover:underline">https://my-website.test</Link> in your browser.
                </p>
              </section>

              {/* Page Footer Navigation */}
              <div className="mt-16 pt-8 border-t border-slate-800 flex justify-between">
                 <div className="text-left">
                   <div className="text-xs text-slate-500 mb-1">Previous</div>
                   <Link href="/docs/getting-started/installation" className="text-primary hover:underline">Installation</Link>
                 </div>
                 <div className="text-right">
                   <div className="text-xs text-slate-500 mb-1">Next</div>
                   <Link href="/docs/getting-started/architecture" className="text-primary hover:underline">Architecture</Link>
                 </div>
              </div>
            </div>
          </div>
        </main>

        {/* Right TOC (Desktop Only) */}
        <TableOfContents />
      </div>
    </div>
  );
};

export default FirstProjectPage;