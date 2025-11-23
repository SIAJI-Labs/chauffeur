"use client";

// React & Next.js
import React, { useEffect } from 'react';
import Link from 'next/link';

// Third-party libraries
import {
  ChevronRight,
  AlertTriangle
} from 'lucide-react';

// Page-specific components
import { TableOfContents } from '@/app/docs/_components/TableOfContents';
import { CodeBlock } from '@/app/docs/_components/CodeBlock';

export default function InstallationPage() {
  const currentSlug = 'getting-started/installation';

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
        <span className="text-slate-200 capitalize">{currentSlug.split('/').pop()?.replace('-', ' ')}</span>
      </div>

      {/* Content */}
      <div className="prose prose-invert prose-slate max-w-none space-y-8 animate-fade-in">
        <div>
          <h1 className="text-4xl font-bold text-white mb-4">Installation</h1>
          <p className="text-lg text-slate-400 leading-relaxed">
            Chauffeur is designed to be installed in seconds. It currently supports Debian-based (Ubuntu, Pop!_OS) and Arch-based distributions.
          </p>
        </div>

        <div className="p-4 bg-amber-500/10 border border-amber-500/20 rounded-lg flex gap-3">
           <AlertTriangle className="text-amber-400 shrink-0" />
           <div className="text-sm text-amber-100/80">
             <strong>Prerequisite:</strong> Ensure you have <code>curl</code>, <code>git</code>, and <code>unzip</code> installed on your system before proceeding.
           </div>
        </div>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="quick-install">
            Quick Install
            <Link href="#quick-install" onClick={(e) => scrollToId(e, 'quick-install')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">The easiest way to install Chauffeur is via our installer script. This will detect your distribution, install dependencies (Nginx, PHP, dnsmasq), and configure the environment.</p>
          <CodeBlock code="curl -sL https://chauffeur.siaji.com/install | bash" />
          <p className="text-slate-400 mt-4">Once completed, the installer will ask you to restart your session to add the binary to your path.</p>
        </section>

  
        <section>
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="verify">
            Verification
            <Link href="#verify" onClick={(e) => scrollToId(e, 'verify')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">Verify the installation by checking the version:</p>
          <CodeBlock code="$ chauf --version\nChauffeur v0.1.0-beta (linux/amd64)" />
        </section>

        {/* Page Footer Navigation */}
        <div className="mt-16 pt-8 border-t border-slate-800 flex justify-between">
           <div className="text-left">
             <div className="text-xs text-slate-500 mb-1">Previous</div>
             <Link href="/docs" className="text-primary hover:underline">Documentation Home</Link>
           </div>
           <div className="text-right">
             <div className="text-xs text-slate-500 mb-1">Next</div>
             <Link href="/docs/getting-started/first-project" className="text-primary hover:underline">First Project</Link>
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