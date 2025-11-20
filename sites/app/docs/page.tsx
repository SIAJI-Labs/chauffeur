"use client";

import React, { useState, useEffect } from 'react';
import { usePathname } from 'next/navigation';
import Link from 'next/link';
import { Terminal as TerminalIcon, Menu, X, ChevronRight, AlertTriangle, Github, ArrowLeft } from 'lucide-react';
import { DocSidebar } from '../../components/docs/DocSidebar';
import { TableOfContents } from '../../components/docs/TableOfContents';
import { CodeBlock } from '../../components/docs/CodeBlock';
import { Button } from '../../components/ui/Button';

const DocsPage: React.FC = () => {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const pathname = usePathname();
  const currentSlug = pathname.replace('/docs/', '') || 'getting-started/installation';

  // Scroll to top on route change
  useEffect(() => {
    window.scrollTo(0, 0);
  }, [pathname]);

  // Content Renderer (Simulated based on slug)
  const renderContent = () => {
    switch (true) {
      case currentSlug.includes('installation'):
        return (
          <div>
            <h1 className="text-4xl font-bold mb-4">Installation</h1>
            <p className="text-lg mb-6">Install Chauffeur on your Linux system with a single command:</p>
            <CodeBlock code="curl -sL https://chauffeur.dev/get | bash" />
            <p className="text-lg mt-6 mb-6">After installation, initialize the workspace:</p>
            <CodeBlock code="chauf init" />
          </div>
        );
      case currentSlug.includes('first-project'):
        return (
          <div>
            <h1 className="text-4xl font-bold mb-4">First Project</h1>
            <p className="text-lg mb-6">Create your first Chauffeur project:</p>
            <CodeBlock code="chauf link --ssl" />
            <p className="text-lg mt-6 mb-6">Start the development services:</p>
            <CodeBlock code="chauf start" />
          </div>
        );
      case currentSlug.includes('link'):
        return (
          <div>
            <h1 className="text-4xl font-bold mb-4">chauf link</h1>
            <p className="text-lg mb-6">Registers the current working directory as a Chauffeur site with automatic .test domain routing.</p>

            <h2 className="text-2xl font-semibold mb-3">Usage</h2>
            <CodeBlock code="chauf link [name] [flags]" />

            <h2 className="text-2xl font-semibold mb-3">Common Flags</h2>
            <div className="bg-card rounded-lg p-4 overflow-x-auto">
              <table className="w-full text-left">
                <thead>
                  <tr className="border-b">
                    <th className="p-2 text-left font-mono">Flag</th>
                    <th className="p-2 text-left">Description</th>
                    <th className="p-2 text-left">Default</th>
                  </tr>
                </thead>
                <tbody>
                  <tr className="border-b">
                    <td className="p-2 font-mono">--ssl</td>
                    <td className="p-2">Generate SSL certificate</td>
                    <td className="p-2">false</td>
                  </tr>
                  <tr className="border-b">
                    <td className="p-2 font-mono">--php</td>
                    <td className="p-2">Specify PHP version (e.g. 8.1)</td>
                    <td className="p-2">global</td>
                  </tr>
                  <tr className="border-b">
                    <td className="p-2 font-mono">--force</td>
                    <td className="p-2">Overwrite existing config</td>
                    <td className="p-2">false</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        );
      default:
        return (
          <div>
            <h1 className="text-4xl font-bold mb-4">Documentation</h1>
            <p className="text-lg mb-6">Select a section from the sidebar to get started.</p>
          </div>
        );
    }
  };

  return (
    <div className="min-h-screen bg-background flex">
      {/* Mobile Menu Button */}
      <button
        onClick={() => setSidebarOpen(!sidebarOpen)}
        className="md:hidden fixed top-4 left-4 z-50 p-2 bg-card rounded-lg border"
      >
        {sidebarOpen ? <X size={24} /> : <Menu size={24} />}
      </button>

      {/* Sidebar */}
      <div className={`
        fixed top-0 left-0 h-full w-72 bg-card border-r transform transition-transform duration-300 ease-in-out z-40
        ${sidebarOpen ? 'translate-x-0' : '-translate-x-full'}
        md:translate-x-0
      `}>
        <div className="h-full overflow-y-auto">
          <div className="p-6">
            <div className="flex items-center justify-between mb-6">
              <Link href="/" className="flex items-center gap-2 text-foreground hover:text-primary">
                <ArrowLeft size={16} />
                <span className="font-bold">Chauffeur</span>
              </Link>
              <button
                onClick={() => setSidebarOpen(false)}
                className="md:hidden p-1 text-muted-foreground hover:text-foreground"
              >
                <X size={20} />
              </button>
            </div>
          </div>

          <div className="px-6 pb-6">
            <DocSidebar isOpen={sidebarOpen} onClose={() => setSidebarOpen(false)} />
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 md:ml-72">
        {/* Top Bar */}
        <div className="sticky top-0 z-30 bg-card border-b px-6 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <TerminalIcon className="w-5 h-5 text-muted-foreground" />
              <h1 className="text-lg font-semibold text-foreground">
                Documentation
              </h1>
            </div>
            <div className="flex items-center gap-4">
              <a
                href="https://github.com/siaji/chauffeur"
                target="_blank"
                rel="noopener noreferrer"
                className="text-muted-foreground hover:text-foreground"
              >
                <Github size={20} />
              </a>
            </div>
          </div>
        </div>

        {/* Content */}
        <main className="px-6 py-8">
          <div className="max-w-4xl mx-auto">
            {renderContent()}
          </div>
        </main>
      </div>

      {/* Table of Contents */}
      <TableOfContents />

      {/* Sidebar Overlay */}
      {sidebarOpen && (
        <div
          className="fixed inset-0 bg-black/50 z-30 md:hidden"
          onClick={() => setSidebarOpen(false)}
        />
      )}
    </div>
  );
};

export default DocsPage;