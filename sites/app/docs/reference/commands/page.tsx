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
  Terminal
} from 'lucide-react';

// Page-specific components
import { TableOfContents } from '@/app/docs/_components/TableOfContents';
import { CodeBlock } from '@/app/docs/_components/CodeBlock';

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
            Comprehensive guide to the Chauffeur CLI, organized by category for easy navigation.
          </p>
        </div>

        {/* Quick Navigation */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-12">
          <a href="#project-management" className="block p-4 bg-surface border border-slate-800 rounded-lg hover:border-emerald-500/50 transition-colors">
            <Folder className="text-emerald-400 mb-2" size={20} />
            <h3 className="font-semibold text-white mb-1">Project Management</h3>
            <p className="text-sm text-slate-400">link, isolate, secure, unlink, status</p>
          </a>
          <a href="#service-control" className="block p-4 bg-surface border border-slate-800 rounded-lg hover:border-blue-500/50 transition-colors">
            <Server className="text-blue-400 mb-2" size={20} />
            <h3 className="font-semibold text-white mb-1">Service Control</h3>
            <p className="text-sm text-slate-400">start, stop, restart</p>
          </a>
          <a href="#php-management" className="block p-4 bg-surface border border-slate-800 rounded-lg hover:border-amber-500/50 transition-colors">
            <Code className="text-amber-400 mb-2" size={20} />
            <h3 className="font-semibold text-white mb-1">PHP Management</h3>
            <p className="text-sm text-slate-400">use, php install</p>
          </a>
          <a href="#utilities" className="block p-4 bg-surface border border-slate-800 rounded-lg hover:border-purple-500/50 transition-colors">
            <Wrench className="text-purple-400 mb-2" size={20} />
            <h3 className="font-semibold text-white mb-1">Utilities</h3>
            <p className="text-sm text-slate-400">logs, clean</p>
          </a>
        </div>

        {/* Project Management Commands */}
        <section id="project-management">
          <div className="flex items-center gap-3 mb-6 pb-4 border-b border-slate-800">
            <Folder className="text-emerald-400" size={24} />
            <h2 className="text-3xl font-bold text-white">Project Management</h2>
          </div>
          <p className="text-slate-400 mb-8">
            Commands for creating, configuring, and managing your Chauffeur projects.
          </p>

          <div className="space-y-12">
            <div>
              <h3 className="text-2xl font-bold text-white mb-4 font-mono group flex items-center gap-2" id="link">
                chauf link
                <Link href="#link" onClick={(e) => scrollToId(e, 'link')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
              </h3>
              <p className="text-slate-400 mb-4">Registers the current working directory as a Chauffeur site.</p>

              <h4 className="text-lg font-semibold text-slate-200 mb-2">Usage</h4>
              <CodeBlock code="chauf link [name] [flags]" />

              <h4 className="text-lg font-semibold text-slate-200 mb-2">Flags</h4>
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
                      <td className="p-3 font-mono text-slate-300">--php</td>
                      <td className="p-3 text-slate-400">Specify PHP version (e.g. 8.1, 8.3)</td>
                      <td className="p-3 text-slate-500">global</td>
                    </tr>
                    <tr>
                      <td className="p-3 font-mono text-slate-300">--secure</td>
                      <td className="p-3 text-slate-400">Generate SSL certificate for the site</td>
                      <td className="p-3 text-slate-500">false</td>
                    </tr>
                    <tr>
                      <td className="p-3 font-mono text-slate-300">--site</td>
                      <td className="p-3 text-slate-400">Set custom domain (e.g. myapp.test)</td>
                      <td className="p-3 text-slate-500">directory-name</td>
                    </tr>
                    <tr>
                      <td className="p-3 font-mono text-slate-300">--alias</td>
                      <td className="p-3 text-slate-400">Add domain alias (can be used multiple times)</td>
                      <td className="p-3 text-slate-500">none</td>
                    </tr>
                    <tr>
                      <td className="p-3 font-mono text-slate-300">--force</td>
                      <td className="p-3 text-slate-400">Overwrite existing project configuration</td>
                      <td className="p-3 text-slate-500">false</td>
                    </tr>
                    <tr>
                      <td className="p-3 font-mono text-slate-300">--dedicated-fpm</td>
                      <td className="p-3 text-slate-400">Use dedicated PHP-FPM pool for this project</td>
                      <td className="p-3 text-slate-500">false</td>
                    </tr>
                    <tr>
                      <td className="p-3 font-mono text-slate-300">--http-port</td>
                      <td className="p-3 text-slate-400">Set custom HTTP port (1-65535)</td>
                      <td className="p-3 text-slate-500">80</td>
                    </tr>
                    <tr>
                      <td className="p-3 font-mono text-slate-300">--https-port</td>
                      <td className="p-3 text-slate-400">Set custom HTTPS port (1-65535)</td>
                      <td className="p-3 text-slate-500">443</td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <h4 className="text-lg font-semibold text-slate-200 mb-2">Example</h4>
              <CodeBlock code={`$ cd ~/projects/my-laravel-app
$ chauf link --php=8.3 --secure --site=laravel.test --alias=api.laravel.test --alias=admin.laravel.test

✓ Linking project: my-laravel-app
✓ PHP 8.3 runtime configured
✓ SSL certificate generated for laravel.test
✓ Domain aliases configured: api.laravel.test, admin.laravel.test
✓ Nginx configuration created
✓ Site available at: https://laravel.test

🎉 Project successfully linked!`} />

              <div className="mt-6 p-4 bg-amber-500/10 border border-amber-500/20 rounded-lg">
                <h4 className="font-semibold text-amber-300 mb-2">🔐 SSL Certificate Requirements</h4>
                <div className="text-sm text-amber-100/80 space-y-2">
                  <p>
                    <strong>Self-signed certificates:</strong> Generated automatically without sudo privileges.
                    Works locally but browsers show security warnings.
                  </p>
                  <p>
                    <strong>Trusted certificates (mkcert):</strong> Requires sudo for system trust store installation.
                    Provides browser-trusted certificates without security warnings.
                  </p>
                  <div className="bg-slate-900/50 p-3 rounded mt-2">
                    <code className="text-amber-300"># Install mkcert for trusted certificates (requires sudo)</code><br/>
                    <code>go install -r filippo.io/mkcert@latest</code><br/>
                    <code>mkcert -install  # Prompts for sudo password</code>
                  </div>
                </div>
              </div>
            </div>

            <div>
              <h3 className="text-2xl font-bold text-white mb-4 font-mono group flex items-center gap-2" id="isolate">
                chauf isolate
                <Link href="#isolate" onClick={(e) => scrollToId(e, 'isolate')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
              </h3>
              <p className="text-slate-400 mb-4">Isolates the current site to a specific PHP version, separate from the global default.</p>
              <CodeBlock code="chauf php isolate <version>" />

              <h4 className="text-lg font-semibold text-slate-200 mb-2">Arguments</h4>
              <div className="overflow-x-auto">
                <table className="w-full text-left text-sm border-collapse">
                  <thead>
                    <tr className="border-b border-slate-700 bg-slate-800/50">
                      <th className="p-3 font-mono text-emerald-400">Argument</th>
                      <th className="p-3 text-slate-300">Description</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-800">
                    <tr>
                      <td className="p-3 font-mono text-slate-300">&lt;version&gt;</td>
                      <td className="p-3 text-slate-400">PHP version to isolate to (e.g. 7.4, 8.0, 8.1, 8.2, 8.3)</td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <h4 className="text-lg font-semibold text-slate-200 mb-2">Example</h4>
              <CodeBlock code={`$ chauf isolate 7.4

⚠️  Isolating current project to PHP 7.4
✓ Dedicated PHP-FPM pool created
✓ Configuration updated
✓ Services restarted

Project now uses PHP 7.4 (isolated from global PHP 8.3)

$ php -v
PHP 7.4.33 (cli) (built: Nov 15 2024 12:30:00)
Copyright (c) The PHP Group`} />
            </div>

            <div>
              <h3 className="text-2xl font-bold text-white mb-4 font-mono group flex items-center gap-2" id="unlink">
                chauf unlink
                <Link href="#unlink" onClick={(e) => scrollToId(e, 'unlink')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
              </h3>
              <p className="text-slate-400 mb-4">Remove registrations or specific aliases. Defaults to current dir.</p>
              <CodeBlock code="chauf unlink [--slug] [--site] [--project] [--alias] [--all] [--force]" />

              <h4 className="text-lg font-semibold text-slate-200 mb-2">Flags</h4>
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
                      <td className="p-3 font-mono text-slate-300">--slug</td>
                      <td className="p-3 text-slate-400">Remove project by slug identifier</td>
                    </tr>
                    <tr>
                      <td className="p-3 font-mono text-slate-300">--site</td>
                      <td className="p-3 text-slate-400">Remove site configuration only</td>
                    </tr>
                    <tr>
                      <td className="p-3 font-mono text-slate-300">--project</td>
                      <td className="p-3 text-slate-400">Remove entire project configuration</td>
                    </tr>
                    <tr>
                      <td className="p-3 font-mono text-slate-300">--alias</td>
                      <td className="p-3 text-slate-400">Remove specific domain alias</td>
                    </tr>
                    <tr>
                      <td className="p-3 font-mono text-slate-300">--all</td>
                      <td className="p-3 text-slate-400">Remove all registered projects</td>
                    </tr>
                    <tr>
                      <td className="p-3 font-mono text-slate-300">--force</td>
                      <td className="p-3 text-slate-400">Skip confirmation prompts</td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <h4 className="text-lg font-semibold text-slate-200 mb-2">Example</h4>
              <CodeBlock code={`$ chauf unlink old-project --force

⚠️  Unlinking project: old-project
✓ Nginx configuration removed
✓ PHP-FPM pool stopped
✓ SSL certificates backed up
✓ DNS rules cleaned up

🗑️  Project successfully unlinked!

# Remove specific domain alias
$ chauf unlink --alias=admin.my-app.test
✓ Domain alias removed: admin.my-app.test`} />
            </div>

            <div>
              <h3 className="text-2xl font-bold text-white mb-4 font-mono group flex items-center gap-2" id="status">
                chauf status
                <Link href="#status" onClick={(e) => scrollToId(e, 'status')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
              </h3>
              <p className="text-slate-400 mb-4">Show status for global or per-project services.</p>
              <CodeBlock code="chauf status [service-type] [--project] [--detail] [-v]" />

              <h4 className="text-lg font-semibold text-slate-200 mb-2">Arguments & Flags</h4>
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
                      <td className="p-3 font-mono text-slate-300">[service-type]</td>
                      <td className="p-3 text-slate-400">Filter by service type (nginx, php-fpm, dnsmasq)</td>
                    </tr>
                    <tr>
                      <td className="p-3 font-mono text-slate-300">--project</td>
                      <td className="p-3 text-slate-400">Show status for specific project only</td>
                    </tr>
                    <tr>
                      <td className="p-3 font-mono text-slate-300">--detail</td>
                      <td className="p-3 text-slate-400">Show detailed status information</td>
                    </tr>
                    <tr>
                      <td className="p-3 font-mono text-slate-300">-v, --verbose</td>
                      <td className="p-3 text-slate-400">Verbose output with additional details</td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <h4 className="text-lg font-semibold text-slate-200 mb-2">Example</h4>
              <CodeBlock code={`$ chauf status --detail

┌──────────────────────────────────────────────────────┐
│                  CHAUFFEUR STATUS                    │
├──────────────────────────────────────────────────────┤
│ Version: 2.1.0                                      │
│ Workspace: ~/.chauffeur                              │
├──────────────────────────────────────────────────────┤
│ GLOBAL SERVICES                                      │
│ ✓ dnsmasq    running  (pid: 1234)                   │
│ ✓ nginx      running  (pid: 5678)                   │
├──────────────────────────────────────────────────────┤
│ REGISTERED PROJECTS                                  │
│ ✓ my-app      PHP 8.3  https://my-app.test         │
│ ✓ blog        PHP 8.2  http://blog.test            │
│ ✓ api         PHP 8.1  https://api.test             │
└──────────────────────────────────────────────────────┘`} />
            </div>
          </div>
        </section>

        {/* Service Control Commands */}
        <section id="service-control">
          <div className="flex items-center gap-3 mb-6 pb-4 border-b border-slate-800">
            <Server className="text-blue-400" size={24} />
            <h2 className="text-3xl font-bold text-white">Service Control</h2>
          </div>
          <p className="text-slate-400 mb-8">
            Commands for starting, stopping, and managing Chauffeur services.
          </p>

          <div className="space-y-12">
            <div>
              <h3 className="text-2xl font-bold text-white mb-4 font-mono group flex items-center gap-2" id="start">
                chauf start
                <Link href="#start" onClick={(e) => scrollToId(e, 'start')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
              </h3>
              <p className="text-slate-400 mb-4">Start nginx/PHP-FPM plus dnsmasq validation.</p>
              <CodeBlock code="chauf start [--project <path>] [--all] [--dry-run]" />

              <h4 className="text-lg font-semibold text-slate-200 mb-2 mt-4">Flags</h4>
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

              <h4 className="text-lg font-semibold text-slate-200 mb-2">Example</h4>
              <CodeBlock code={`$ chauf start --all

🚀 Starting Chauffeur services...
✓ dnsmasq started (DNS resolution enabled)
✓ nginx started (port 80, 443)
✓ PHP 8.3 FPM pool started
✓ PHP 8.2 FPM pool started
✓ PHP 8.1 FPM pool started

🌐 All services running successfully!

# Start specific project
$ chauf start --project=~/projects/my-app
✓ PHP 8.3 FPM pool started for my-app
✓ Nginx configuration reloaded`} />
            </div>

            <div>
              <h3 className="text-2xl font-bold text-white mb-4 font-mono group flex items-center gap-2" id="stop">
                chauf stop
                <Link href="#stop" onClick={(e) => scrollToId(e, 'stop')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
              </h3>
              <p className="text-slate-400 mb-4">Stop services and clean port-forward rules.</p>
              <CodeBlock code="chauf stop [--project <path>] [--all] [--dry-run]" />

              <h4 className="text-lg font-semibold text-slate-200 mb-2">Example</h4>
              <CodeBlock code={`$ chauf stop --all

🛑 Stopping Chauffeur services...
✓ nginx stopped
✓ PHP 8.3 FPM pool stopped
✓ PHP 8.2 FPM pool stopped
✓ PHP 8.1 FPM pool stopped
✓ dnsmasq stopped

🔌 All services stopped gracefully!

# Stop specific project
$ chauf stop --project=~/projects/legacy-app
✓ PHP 7.4 FPM pool stopped for legacy-app
✓ Nginx configuration reloaded`} />
            </div>

            <div>
              <h3 className="text-2xl font-bold text-white mb-4 font-mono group flex items-center gap-2" id="restart">
                chauf restart
                <Link href="#restart" onClick={(e) => scrollToId(e, 'restart')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
              </h3>
              <p className="text-slate-400 mb-4">Restart services (equivalent to stop then start, preserves configuration).</p>
              <CodeBlock code="chauf restart [--project <slug>] [--all] [--dry-run]" />

              <h4 className="text-lg font-semibold text-slate-200 mb-2">Example</h4>
              <CodeBlock code={`$ chauf restart --all

🔄 Restarting Chauffeur services...
✓ nginx restarted gracefully
✓ PHP 8.3 FPM pool restarted
✓ PHP 8.2 FPM pool restarted
✓ PHP 8.1 FPM pool restarted
✓ dnsmasq restarted

🎯 All services restarted successfully!

# Restart after configuration changes
$ chauf restart
✓ Services restarted - applying new configuration`} />
            </div>
          </div>
        </section>

        {/* PHP Management Commands */}
        <section id="php-management">
          <div className="flex items-center gap-3 mb-6 pb-4 border-b border-slate-800">
            <Code className="text-amber-400" size={24} />
            <h2 className="text-3xl font-bold text-white">PHP Management</h2>
          </div>
          <p className="text-slate-400 mb-8">
            Commands for managing PHP versions and installations.
          </p>

          <div className="space-y-12">
            <div>
              <h3 className="text-2xl font-bold text-white mb-4 font-mono group flex items-center gap-2" id="php-use">
                chauf php use
                <Link href="#php-use" onClick={(e) => scrollToId(e, 'php-use')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
              </h3>
              <p className="text-slate-400 mb-4">Set the global default PHP version for new sites.</p>
              <CodeBlock code="chauf php use <version>" />

              <h4 className="text-lg font-semibold text-slate-200 mb-2">Arguments</h4>
              <div className="overflow-x-auto">
                <table className="w-full text-left text-sm border-collapse">
                  <thead>
                    <tr className="border-b border-slate-700 bg-slate-800/50">
                      <th className="p-3 font-mono text-emerald-400">Argument</th>
                      <th className="p-3 text-slate-300">Description</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-800">
                    <tr>
                      <td className="p-3 font-mono text-slate-300">&lt;version&gt;</td>
                      <td className="p-3 text-slate-400">PHP version to set as global default (e.g. 7.4, 8.0, 8.1, 8.2, 8.3)</td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <h4 className="text-lg font-semibold text-slate-200 mb-2">Arguments</h4>
              <div className="overflow-x-auto">
                <table className="w-full text-left text-sm border-collapse">
                  <thead>
                    <tr className="border-b border-slate-700 bg-slate-800/50">
                      <th className="p-3 font-mono text-emerald-400">Argument</th>
                      <th className="p-3 text-slate-300">Description</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-800">
                    <tr>
                      <td className="p-3 font-mono text-slate-300">&lt;version&gt;</td>
                      <td className="p-3 text-slate-400">PHP version to set as global default (e.g. 7.4, 8.0, 8.1, 8.2, 8.3, 8.4)</td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <h4 className="text-lg font-semibold text-slate-200 mb-2">Example</h4>
              <CodeBlock code={`$ chauf php use 8.4

🔄 Setting global PHP version to 8.4
✓ PHP 8.4 runtime already installed
✓ Global PHP version updated
✓ CLI PATH updated

✅ PHP 8.4 is now the global default
New projects will use PHP 8.4 unless specified otherwise

$ php -v
PHP 8.4.14 (cli) (built: Dec 15 2024 10:30:00)
Copyright (c) The PHP Group`} />
            </div>

            <div>
              <h3 className="text-2xl font-bold text-white mb-4 font-mono group flex items-center gap-2" id="php-isolate">
                chauf php isolate
                <Link href="#php-isolate" onClick={(e) => scrollToId(e, 'php-isolate')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
              </h3>
              <p className="text-slate-400 mb-4">Pin the current project to a specific PHP version.</p>
              <CodeBlock code="chauf php isolate <version>" />

              <h4 className="text-lg font-semibold text-slate-200 mb-2">Arguments</h4>
              <div className="overflow-x-auto">
                <table className="w-full text-left text-sm border-collapse">
                  <thead>
                    <tr className="border-b border-slate-700 bg-slate-800/50">
                      <th className="p-3 font-mono text-emerald-400">Argument</th>
                      <th className="p-3 text-slate-300">Description</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-800">
                    <tr>
                      <td className="p-3 font-mono text-slate-300">&lt;version&gt;</td>
                      <td className="p-3 text-slate-400">PHP version to pin to current project (e.g. 7.4, 8.0, 8.1, 8.2, 8.3, 8.4)</td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <h4 className="text-lg font-semibold text-slate-200 mb-2">Example</h4>
              <CodeBlock code={`$ cd ~/projects/legacy-app
$ chauf php isolate 7.4

🔧 Pinning project to PHP 7.4
✓ PHP 7.4 runtime already installed
✓ Project configuration updated
✓ FPM pool restarted

✅ legacy-app pinned to PHP 7.4
This project will use PHP 7.4 regardless of global settings`} />
            </div>

            <div>
              <h3 className="text-2xl font-bold text-white mb-4 font-mono group flex items-center gap-2" id="php-list">
                chauf php list
                <Link href="#php-list" onClick={(e) => scrollToId(e, 'php-list')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
              </h3>
              <p className="text-slate-400 mb-4">List all supported PHP versions and their installation status.</p>
              <CodeBlock code="chauf php list" />

              <h4 className="text-lg font-semibold text-slate-200 mb-2">Example</h4>
              <CodeBlock code={`$ chauf php list
✅ Supported PHP versions:
  PHP 8.4 ❌
  PHP 8.3 ✅ (active)
  PHP 8.2 ✅
  PHP 8.1 ✅
  PHP 8.0 ❌
  PHP 7.4 ❌`} />
            </div>

            <div>
              <h3 className="text-2xl font-bold text-white mb-4 font-mono group flex items-center gap-2" id="init">
                chauf init
                <Link href="#init" onClick={(e) => scrollToId(e, 'init')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
              </h3>
              <p className="text-slate-400 mb-4">Initialize Chauffeur workspace with default configuration and directory structure.</p>
              <CodeBlock code="chauf init [--force] [--quiet]" />

              <h4 className="text-lg font-semibold text-slate-200 mb-2">Flags</h4>
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
                      <td className="p-3 font-mono text-slate-300">--force</td>
                      <td className="p-3 text-slate-400">Overwrite existing configuration files</td>
                    </tr>
                    <tr>
                      <td className="p-3 font-mono text-slate-300">--quiet, -q</td>
                      <td className="p-3 text-slate-400">Suppress verbose output</td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <h4 className="text-lg font-semibold text-slate-200 mb-2 mt-4">What It Creates</h4>
              <div className="bg-slate-900 p-4 rounded-lg border border-slate-800">
                <pre className="text-sm text-slate-300 font-mono">
{`~/.chauffeur/
├── config/chauffeur.yaml     # Main configuration file
├── projects/                 # Project storage
├── logs/                     # Service logs
├── cache/                    # Download cache
├── php/                      # PHP installations
├── nginx/                    # nginx configurations
└── bin/                      # Chauffeur binaries and shims`}
                </pre>
              </div>

              <h4 className="text-lg font-semibold text-slate-200 mb-2 mt-4">Examples</h4>
              <CodeBlock code={`$ chauf init

[init] Initializing Chauffeur workspace...
[init] Workspace directory: /home/user/.chauffeur
[init] Creating directory: /home/user/.chauffeur/config
[init] Creating directory: /home/user/.chauffeur/projects
[init] Creating default configuration file...
✓ Configuration file created: /home/user/.chauffeur/config/chauffeur.yaml
[init] Creating .gitignore file
[init] Checking PATH configuration...
[init] Chauffeur bin directory: /home/user/.chauffeur/bin
[init] Add this to your shell profile (~/.bashrc or ~/.zshrc):
[init]   export PATH="/home/user/.chauffeur/bin:$PATH"
[init] Then restart your shell or run: source ~/.bashrc
✓ Chauffeur workspace initialized successfully at /home/user/.chauffeur

$ chauf init --force --quiet
✓ Workspace initialized /home/user/.chauffeur`} />
            </div>

            <div>
              <h3 className="text-2xl font-bold text-white mb-4 font-mono group flex items-center gap-2" id="install">
                chauf install
                <Link href="#install" onClick={(e) => scrollToId(e, 'install')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
              </h3>
              <p className="text-slate-400 mb-4">Install Chauffeur-managed services (nginx, php, composer).</p>
              <CodeBlock code="chauf install [--force] [--local] [--no-cache] <service> [version]" />

              <h4 className="text-lg font-semibold text-slate-200 mb-2">Services</h4>
              <div className="overflow-x-auto">
                <table className="w-full text-left text-sm border-collapse">
                  <thead>
                    <tr className="border-b border-slate-700 bg-slate-800/50">
                      <th className="p-3 font-mono text-emerald-400">Service</th>
                      <th className="p-3 text-slate-300">Description</th>
                      <th className="p-3 text-slate-300">Example</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-800">
                    <tr>
                      <td className="p-3 font-mono text-slate-300">nginx</td>
                      <td className="p-3 text-slate-400">Web server with Chauffeur configuration</td>
                      <td className="p-3 text-slate-500"><code>chauf install nginx</code></td>
                    </tr>
                    <tr>
                      <td className="p-3 font-mono text-slate-300">php</td>
                      <td className="p-3 text-slate-400">PHP runtime with development extensions</td>
                      <td className="p-3 text-slate-500"><code>chauf install php 8.3</code></td>
                    </tr>
                    <tr>
                      <td className="p-3 font-mono text-slate-300">composer</td>
                      <td className="p-3 text-slate-400">PHP dependency manager with version isolation</td>
                      <td className="p-3 text-slate-500"><code>chauf install composer</code></td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <h4 className="text-lg font-semibold text-slate-200 mb-2 mt-4">Flags</h4>
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
                      <td className="p-3 font-mono text-slate-300">--force</td>
                      <td className="p-3 text-slate-400">Reinstall even if already present</td>
                      <td className="p-3 text-slate-500">false</td>
                    </tr>
                    <tr>
                      <td className="p-3 font-mono text-slate-300">--local</td>
                      <td className="p-3 text-slate-400">Use local PHP tarball (PHP only)</td>
                      <td className="p-3 text-slate-500">false</td>
                    </tr>
                    <tr>
                      <td className="p-3 font-mono text-slate-300">--no-cache</td>
                      <td className="p-3 text-slate-400">Skip download caching</td>
                      <td className="p-3 text-slate-500">false</td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <h4 className="text-lg font-semibold text-slate-200 mb-2">Examples</h4>
              <CodeBlock code={`# Install nginx web server
$ chauf install nginx

# Install specific PHP version
$ chauf install php 8.3

# Install composer for dependency management
$ chauf install composer

# Force reinstall nginx
$ chauf install nginx --force

# Install PHP from local tarball
$ chauf install php 8.3 --local

# Install without caching downloads
$ chauf install php 8.4 --no-cache`} />

              <div className="mt-6 p-4 bg-blue-500/10 border border-blue-500/20 rounded-lg">
                <h4 className="font-semibold text-blue-300 mb-2">💡 Installation Order</h4>
                <p className="text-sm text-blue-100/80">
                  Recommended installation sequence: 1) nginx → 2) php → 3) composer.
                  Nginx provides the web server, PHP adds runtime support, and composer manages dependencies.
                </p>
              </div>
            </div>

            <div>
              <h3 className="text-2xl font-bold text-white mb-4 font-mono group flex items-center gap-2" id="php-install">
                chauf php install
                <Link href="#php-install" onClick={(e) => scrollToId(e, 'php-install')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
              </h3>
              <p className="text-slate-400 mb-4">Build/install PHP runtime under workspace.</p>
              <CodeBlock code="chauf php install <version> [--force] [--no-ext] [--from <source>]" />

              <h4 className="text-lg font-semibold text-slate-200 mb-2">Arguments & Flags</h4>
              <div className="overflow-x-auto">
                <table className="w-full text-left text-sm border-collapse">
                  <thead>
                    <tr className="border-b border-slate-700 bg-slate-800/50">
                      <th className="p-3 font-mono text-emerald-400">Argument/Flag</th>
                      <th className="p-3 text-slate-300">Description</th>
                      <th className="p-3 text-slate-300">Default</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-800">
                    <tr>
                      <td className="p-3 font-mono text-slate-300">&lt;version&gt;</td>
                      <td className="p-3 text-slate-400">PHP version to install (e.g. 7.4, 8.0, 8.1, 8.2, 8.3)</td>
                      <td className="p-3 text-slate-500">-</td>
                    </tr>
                    <tr>
                      <td className="p-3 font-mono text-slate-300">--force</td>
                      <td className="p-3 text-slate-400">Reinstall even if already exists</td>
                      <td className="p-3 text-slate-500">false</td>
                    </tr>
                    <tr>
                      <td className="p-3 font-mono text-slate-300">--no-ext</td>
                      <td className="p-3 text-slate-400">Skip extension compilation (faster install)</td>
                      <td className="p-3 text-slate-500">false</td>
                    </tr>
                    <tr>
                      <td className="p-3 font-mono text-slate-300">--from &lt;source&gt;</td>
                      <td className="p-3 text-slate-400">Install from specific source (package, source)</td>
                      <td className="p-3 text-slate-500">package</td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <h4 className="text-lg font-semibold text-slate-200 mb-2">Example</h4>
              <CodeBlock code={`$ chauf php install 8.3

🔧 Installing PHP 8.3...
📦 Downloading PHP 8.3 source...
⚙️  Configuring build options...
🏗️  Compiling PHP (this may take 10-15 minutes)...
📚 Installing extensions: curl, gd, intl, json, mbstring, mysqlnd, openssl, pdo, pdo_mysql, zip...
✗  Compilation failed

⚠️  Installation failed! Trying pre-compiled package...
📦 Downloading pre-compiled PHP 8.3 package...
✓ PHP 8.3 installed successfully
✓ Extensions loaded: 28
✓ CLI binary available: ~/.chauffeur/php/8.3/bin/php

🎉 PHP 8.3 installation complete!

# Quick install without extensions
$ chauf php install 8.2 --no-ext
✓ PHP 8.2 installed (basic runtime)`} />
            </div>
          </div>
        </section>

        {/* Utility Commands */}
        <section id="utilities">
          <div className="flex items-center gap-3 mb-6 pb-4 border-b border-slate-800">
            <Wrench className="text-purple-400" size={24} />
            <h2 className="text-3xl font-bold text-white">Utilities</h2>
          </div>
          <p className="text-slate-400 mb-8">
            Utility commands for debugging, maintenance, and system management.
          </p>

          <div className="space-y-12">
            <div>
              <h3 className="text-2xl font-bold text-white mb-4 font-mono group flex items-center gap-2" id="logs">
                chauf logs
                <Link href="#logs" onClick={(e) => scrollToId(e, 'logs')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
              </h3>
              <p className="text-slate-400 mb-4">View and follow service logs with interactive version selection.</p>
              <CodeBlock code="chauf logs [service] [version] [--follow] [-f] [--lines] [-n] [--level] [--context] [-c]" />

              <h4 className="text-lg font-semibold text-slate-200 mb-2">Arguments & Flags</h4>
              <div className="overflow-x-auto">
                <table className="w-full text-left text-sm border-collapse">
                  <thead>
                    <tr className="border-b border-slate-700 bg-slate-800/50">
                      <th className="p-3 font-mono text-emerald-400">Argument/Flag</th>
                      <th className="p-3 text-slate-300">Description</th>
                      <th className="p-3 text-slate-300">Default</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-800">
                    <tr>
                      <td className="p-3 font-mono text-slate-300">[service]</td>
                      <td className="p-3 text-slate-400">Service to monitor (nginx, php-fpm, dnsmasq, or all)</td>
                      <td className="p-3 text-slate-500">all</td>
                    </tr>
                    <tr>
                      <td className="p-3 font-mono text-slate-300">[version]</td>
                      <td className="p-3 text-slate-400">PHP version (8.1, 8.2, 8.3) for PHP-FPM logs</td>
                      <td className="p-3 text-slate-500">all</td>
                    </tr>
                    <tr>
                      <td className="p-3 font-mono text-slate-300">--follow, -f</td>
                      <td className="p-3 text-slate-400">Follow logs in real-time (like tail -f)</td>
                      <td className="p-3 text-slate-500">false</td>
                    </tr>
                    <tr>
                      <td className="p-3 font-mono text-slate-300">--lines, -n</td>
                      <td className="p-3 text-slate-400">Number of lines to show from end</td>
                      <td className="p-3 text-slate-500">50</td>
                    </tr>
                    <tr>
                      <td className="p-3 font-mono text-slate-300">--level</td>
                      <td className="p-3 text-slate-400">Filter by log level (error, warn, info, debug)</td>
                      <td className="p-3 text-slate-500">all</td>
                    </tr>
                    <tr>
                      <td className="p-3 font-mono text-slate-300">--context, -c</td>
                      <td className="p-3 text-slate-400">Show context lines around matches</td>
                      <td className="p-3 text-slate-500">3</td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <h4 className="text-lg font-semibold text-slate-200 mb-2">Example</h4>
              <CodeBlock code={`$ chauf logs --follow

📋 Following all service logs (Ctrl+C to stop)

[2025-01-20 14:30:15] nginx.access: 192.168.1.100 - GET /index.php HTTP/1.1 200
[2025-01-20 14:30:16] php-fpm.8.3: [15-Jan-2025 14:30:16] NOTICE: fpm is running, pid 1234
[2025-01-20 14:30:17] dnsmasq: read /etc/hosts - 2 addresses

# View specific service logs
$ chauf logs nginx --lines 20
📋 Nginx logs (last 20 lines)
[2025-01-20 14:25:10] nginx.error: 2025/01/20 14:25:10 [notice] 5678#5678: signal 17 (SIGCHLD) received
[2025-01-20 14:25:11] nginx.access: 127.0.0.1 - GET / HTTP/1.1 200 3124

# Filter by error level
$ chauf logs --level error
📋 Error logs from all services
[2025-01-20 14:15:32] php-fpm.8.2: [ERROR] Unable to connect to database
[2025-01-20 14:18:45] nginx.error: *1 connect() failed (111: Connection refused)`} />
            </div>

            <div>
              <h3 className="text-2xl font-bold text-white mb-4 font-mono group flex items-center gap-2" id="clean">
                chauf clean
                <Link href="#clean" onClick={(e) => scrollToId(e, 'clean')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
              </h3>
              <p className="text-slate-400 mb-4">Clean workspace files with file size display and accurate reporting.</p>
              <CodeBlock code="chauf clean [target] [--dry-run] [--force] [--older-than] [--keep-versions] [--what]" />

              <h4 className="text-lg font-semibold text-slate-200 mb-2">Arguments & Flags</h4>
              <div className="overflow-x-auto">
                <table className="w-full text-left text-sm border-collapse">
                  <thead>
                    <tr className="border-b border-slate-700 bg-slate-800/50">
                      <th className="p-3 font-mono text-emerald-400">Argument/Flag</th>
                      <th className="p-3 text-slate-300">Description</th>
                      <th className="p-3 text-slate-300">Default</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-800">
                    <tr>
                      <td className="p-3 font-mono text-slate-300">[target]</td>
                      <td className="p-3 text-slate-400">What to clean (logs, cache, temp, certs, all)</td>
                      <td className="p-3 text-slate-500">all</td>
                    </tr>
                    <tr>
                      <td className="p-3 font-mono text-slate-300">--dry-run</td>
                      <td className="p-3 text-slate-400">Show what would be deleted without actually deleting</td>
                      <td className="p-3 text-slate-500">false</td>
                    </tr>
                    <tr>
                      <td className="p-3 font-mono text-slate-300">--force</td>
                      <td className="p-3 text-slate-400">Skip confirmation prompts</td>
                      <td className="p-3 text-slate-500">false</td>
                    </tr>
                    <tr>
                      <td className="p-3 font-mono text-slate-300">--older-than</td>
                      <td className="p-3 text-slate-400">Delete files older than specified duration (e.g. 7d, 30d)</td>
                      <td className="p-3 text-slate-500">0d (all)</td>
                    </tr>
                    <tr>
                      <td className="p-3 font-mono text-slate-300">--keep-versions</td>
                      <td className="p-3 text-slate-400">Keep N most recent versions of each file</td>
                      <td className="p-3 text-slate-500">3</td>
                    </tr>
                    <tr>
                      <td className="p-3 font-mono text-slate-300">--what</td>
                      <td className="p-3 text-slate-400">Detailed breakdown of what would be cleaned</td>
                      <td className="p-3 text-slate-500">false</td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <h4 className="text-lg font-semibold text-slate-200 mb-2">Example</h4>
              <CodeBlock code={`$ chauf clean --dry-run

🧹 Workspace cleanup (dry run)
📊 Files that would be deleted:

📁 Logs (245 MB):
  ~/.chauffeur/logs/nginx-access.log (125 MB)
  ~/.chauffeur/logs/php-fpm.8.2.log (89 MB)
  ~/.chauffeur/logs/php-fpm.8.1.log (31 MB)

🗃️  Cache (512 MB):
  ~/.chauffeur/cache/nginx/ (312 MB)
  ~/.chauffeur/cache/opcache/ (200 MB)

📝 Temporary files (45 MB):
  ~/.chauffeur/tmp/ (45 MB)

💾 Total: 802 MB would be freed
Run without --dry-run to actually delete files

# Clean logs older than 7 days
$ chauf clean logs --older-than 7d --force
🧹 Cleaning logs older than 7 days...
✓ Deleted 142 log files (234 MB freed)
✓ Kept 28 recent log files

# Clean everything but keep recent versions
$ chauf clean all --keep-versions 5
⚠️  This will delete 1.2 GB of files
Proceed? (y/N): y
🧹 Cleaning workspace...
✓ Deleted old logs: 456 MB
✓ Cleaned cache: 678 MB
✓ Removed temp files: 67 MB

✨ Workspace cleanup complete! 1.2 GB freed`} />
            </div>
          </div>
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="doctor">
            chauf doctor
            <Link href="#doctor" onClick={(e) => scrollToId(e, 'doctor')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">Perform health checks and diagnose system issues.</p>
          <CodeBlock code="chauf doctor [options]" />

          <h4 className="text-lg font-semibold text-slate-200 mb-2">Flags</h4>
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
                  <td className="p-3 font-mono text-slate-300">--check-all, -a</td>
                  <td className="p-3 text-slate-400">Run all dependency checks (default)</td>
                </tr>
                <tr>
                  <td className="p-3 font-mono text-slate-300">--check-deps, -d</td>
                  <td className="p-3 text-slate-400">Check system dependencies (git, curl, tar)</td>
                </tr>
                <tr>
                  <td className="p-3 font-mono text-slate-300">--check-php, -p</td>
                  <td className="p-3 text-slate-400">Check PHP build dependencies and headers</td>
                </tr>
                <tr>
                  <td className="p-3 font-mono text-slate-300">--check-ssl, -s</td>
                  <td className="p-3 text-slate-400">Check SSL certificate dependencies</td>
                </tr>
                <tr>
                  <td className="p-3 font-mono text-slate-300">--check-network, -n</td>
                  <td className="p-3 text-slate-400">Check network and port availability</td>
                </tr>
                <tr>
                  <td className="p-3 font-mono text-slate-300">--check-dns</td>
                  <td className="p-3 text-slate-400">Check DNS resolution for .test domains</td>
                </tr>
                <tr>
                  <td className="p-3 font-mono text-slate-300">--verbose, -v</td>
                  <td className="p-3 text-slate-400">Show detailed diagnostic information</td>
                </tr>
                <tr>
                  <td className="p-3 font-mono text-slate-300">--fix, -f</td>
                  <td className="p-3 text-slate-400">Show fix suggestions for issues found</td>
                </tr>
                <tr>
                  <td className="p-3 font-mono text-slate-300">--auto-fix</td>
                  <td className="p-3 text-slate-400">Attempt to automatically fix issues where safe</td>
                </tr>
                <tr>
                  <td className="p-3 font-mono text-slate-300">--quiet, -q</td>
                  <td className="p-3 text-slate-400">Suppress non-error output</td>
                </tr>
              </tbody>
            </table>
          </div>

          <h4 className="text-lg font-semibold text-slate-200 mb-2">Examples</h4>
          <CodeBlock code={`# Run all health checks
$ chauf doctor

# Check only system dependencies
$ chauf doctor --check-deps

# Check PHP dependencies and show fixes
$ chauf doctor --check-php --fix

# Attempt automatic fixes
$ chauf doctor --auto-fix

# Verbose output for troubleshooting
$ chauf doctor --verbose`} />

          <div className="mt-6 p-4 bg-emerald-500/10 border border-emerald-500/20 rounded-lg">
            <h4 className="font-semibold text-emerald-300 mb-2">🩺 What Doctor Checks</h4>
            <div className="text-sm text-emerald-100/80 space-y-1">
              <p><strong>System:</strong> git, curl, tar, gcc, make, pkg-config</p>
              <p><strong>PHP Build:</strong> libzip, libjpeg, libpng, freetype, libxml2, libcurl, zlib</p>
              <p><strong>SSL:</strong> OpenSSL, mkcert availability</p>
              <p><strong>Network:</strong> Port availability, iptables for port forwarding</p>
              <p><strong>DNS:</strong> dnsmasq installation, .test domain resolution</p>
            </div>
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