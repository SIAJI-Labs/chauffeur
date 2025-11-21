"use client";

// React & Next.js
import React, { useEffect } from 'react';
import Link from 'next/link';

// Third-party libraries
import {
  ChevronRight,
  Folder,
  Server,
  Code,
  Wrench,
  Terminal,
  Settings
} from 'lucide-react';

// Page-specific components
import { TableOfContents } from '@/app/docs/_components/TableOfContents';
import { CodeBlock } from '@/app/docs/_components/CodeBlock';

// Import command constants
import { CLI_COMMANDS, COMMANDS_BY_CATEGORY, COMMAND_CATEGORIES, findCommand } from '@/constants';

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

  // Helper to render command details
  const renderCommand = (command: any) => {
    const commandId = command.command.replace(/\s+/g, '-');

    return (
      <div key={command.command} className="space-y-6">
        <h3 className="text-2xl font-bold text-white mb-4 font-mono group flex items-center gap-2" id={commandId}>
          {command.command}
          <Link href={`#${commandId}`} onClick={(e) => scrollToId(e, commandId)} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
        </h3>

        <p className="text-slate-400 mb-4">{command.description}</p>

        <div className="bg-slate-900 p-4 rounded-lg border border-slate-800">
          <code className="text-emerald-300 text-sm font-mono">{command.usage}</code>
        </div>

        {command.keyFlags && command.keyFlags.length > 0 && (
          <div>
            <h4 className="text-lg font-semibold text-slate-200 mb-3">Flags</h4>
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
                  {command.keyFlags.map((flag: any, index: number) => (
                    <tr key={index}>
                      <td className="p-3 font-mono text-slate-300">{flag.flag}</td>
                      <td className="p-3 text-slate-400">{flag.description}</td>
                      <td className="p-3 text-slate-500">{flag.default || '-'}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}

        {command.examples && command.examples.length > 0 && (
          <div>
            <h4 className="text-lg font-semibold text-slate-200 mb-3">
              Examples{command.examples.length > 1 ? 's' : ''}
            </h4>
            <div className="space-y-4">
              {command.examples.map((example: any, index: number) => (
                <div key={index}>
                  <p className="text-sm text-slate-500 mb-2">
                    <strong>{example.name}</strong>: {example.description}
                  </p>
                  <div className="bg-slate-900 p-4 rounded-lg border border-slate-800">
                    <div className="text-sm">
                      <div className="text-emerald-300 font-mono mb-2">$ {example.command}</div>
                      <pre className="text-slate-300 font-mono whitespace-pre-wrap">{example.output}</pre>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {command.notes && command.notes.length > 0 && (
          <div className="mt-6 p-4 bg-blue-500/10 border border-blue-500/20 rounded-lg">
            <h4 className="font-semibold text-blue-300 mb-2">📝 Notes</h4>
            <ul className="text-sm text-blue-100/80 space-y-1 list-disc list-inside">
              {command.notes.map((note: string, index: number) => (
                <li key={index}>{note}</li>
              ))}
            </ul>
          </div>
        )}
      </div>
    );
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
            Comprehensive guide to the Chauffeur CLI, organized by category for easy navigation.
          </p>
        </div>

        {/* Quick Navigation */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mb-12">
          {COMMAND_CATEGORIES.map((category) => (
            <a
              key={category.id}
              href={`#${category.id}`}
              className="block p-4 bg-surface border border-slate-800 rounded-lg hover:border-emerald-500/50 transition-colors"
            >
              <category.icon className="text-emerald-400 mb-2" size={20} />
              <h3 className="font-semibold text-white mb-1">{category.title}</h3>
              <p className="text-sm text-slate-400">{category.description}</p>
            </a>
          ))}
        </div>

        {/* Commands by Category */}
        {COMMAND_CATEGORIES.map((category) => (
          <section key={category.id} id={category.id}>
            <div className="flex items-center gap-3 mb-6 pb-4 border-b border-slate-800">
              <category.icon className="text-emerald-400" size={24} />
              <h2 className="text-3xl font-bold text-white">{category.title}</h2>
            </div>
            <p className="text-slate-400 mb-8">
              {category.description}
            </p>

            <div className="space-y-12">
              {COMMANDS_BY_CATEGORY[category.id as keyof typeof COMMANDS_BY_CATEGORY].map(renderCommand)}
            </div>
          </section>
        ))}

        {/* CLI Commands Summary Table */}
        <section id="command-summary">
          <div className="flex items-center gap-3 mb-6 pb-4 border-b border-slate-800">
            <Settings className="text-emerald-400" size={24} />
            <h2 className="text-3xl font-bold text-white">Command Summary</h2>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm border-collapse">
              <thead>
                <tr className="border-b border-slate-700 bg-slate-800/50">
                  <th className="p-3 text-left text-emerald-400">Command</th>
                  <th className="p-3 text-left text-slate-300">Description</th>
                  <th className="p-3 text-left text-slate-300">Category</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-800">
                {CLI_COMMANDS.map((command) => (
                  <tr key={command.command}>
                    <td className="p-3 font-mono text-slate-300">{command.command}</td>
                    <td className="p-3 text-slate-400">{command.description}</td>
                    <td className="p-3 text-slate-500 capitalize">{command.category}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </section>

        {/* Page Footer Navigation */}
        <div className="mt-16 pt-8 border-t border-slate-800 flex justify-between">
           <div className="text-left">
             <div className="text-xs text-slate-500 mb-1">Previous</div>
             <Link href="/docs/getting-started/architecture" className="text-primary hover:underline">Architecture</Link>
           </div>
           <div className="text-right">
             <div className="text-xs text-slate-500 mb-1">Next</div>
             <Link href="/docs/reference/configuration" className="text-primary hover:underline">Configuration</Link>
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