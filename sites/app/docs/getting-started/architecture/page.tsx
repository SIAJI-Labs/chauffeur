"use client";

import React, { useEffect, useState } from 'react';
import Link from 'next/link';
import { Terminal as TerminalIcon, Menu, X, ChevronRight, Github } from 'lucide-react';
import { DocSidebar } from '@/components/docs/DocSidebar';
import { TableOfContents } from '@/components/docs/TableOfContents';
import { CodeBlock } from '@/components/docs/CodeBlock';
import { Button } from '@/components/ui/Button';

const ArchitecturePage: React.FC = () => {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const currentSlug = 'getting-started/architecture';

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
                <h1 className="text-4xl font-bold text-white mb-4">Chauffeur Architecture</h1>
                <p className="text-lg text-slate-400 leading-relaxed">
                  Understanding the core design principles and architecture of Chauffeur.
                </p>
              </div>

              <section id="core-principles">
                <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="core-principles">
                  Core Principles
                  <Link href="#core-principles" onClick={(e) => scrollToId(e, 'core-principles')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
                </h2>
                <p className="text-slate-400 mb-4">
                  Chauffeur is built upon a set of core principles that prioritize workspace isolation, minimal host impact, and developer ergonomics.
                </p>
                <ul className="list-disc list-inside text-slate-400 space-y-2">
                    <li><strong>Workspace-first:</strong> All binaries, configs, sockets, and logs live under <code>~/.chauffeur</code>.</li>
                    <li><strong>Minimal host impact:</strong> Prefer workspace changes. Host-level changes (e.g., dnsmasq configs) are avoided or require explicit user action.</li>
                    <li><strong>Manual project registration:</strong> Projects are registered explicitly using <code>chauf link</code>.</li>
                    <li><strong>Idempotent operations:</strong> Commands like <code>init</code>, <code>install</code>, <code>link</code> are safe to re-run.</li>
                    <li><strong>Linux-focused:</strong> Designed for Arch/Ubuntu/Debian, no macOS/Windows assumptions.</li>
                    <li><strong>No external env managers:</strong> Chauffeur ships its own runtimes.</li>
                    <li><strong>Documentation parity:</strong> `README.md`, `docs/TODO_STATUS.md`, and `AGENTS.md` must be kept in sync with the codebase.</li>
                    <li><strong>Host edits only as last resort:</strong> System modifications are opt-in and include cleanup guidance.</li>
                </ul>
              </section>

              <section id="workspace-layout">
                <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="workspace-layout">
                  Workspace Layout & Config Contracts
                  <Link href="#workspace-layout" onClick={(e) => scrollToId(e, 'workspace-layout')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
                </h2>
                <p className="text-slate-400 mb-4">
                  All Chauffeur managed files and configurations reside within the <code>~/.chauffeur/</code> directory.
                </p>
                <CodeBlock code={`~/.chauffeur/\n  config/\n    chauffeur.yaml          # global config\n  projects/\n    <slug>/\n      project.yaml          # per-project state\n      runtime/php-fpm/\n      logs/\n  php/<version>/            # installed runtimes\n  nginx/\n    bin/\n    etc/\n    sites-available/\n    sites-enabled/\n    conf.d/\n    certs/\n  bin/\n    chauf                   # optional self-managed binary\n    shims/                  # php, composer, php-<ver>\n  cli/templates/nginx/`}/>
                <h3 className="text-xl font-bold text-white mb-2"><code>chauffeur.yaml</code> (Global Config)</h3>
                <CodeBlock code={`version: 1\ntelemetry: false\nworkspace_dir: ~/.chauffeur\nnginx:\n  enable: true\n  http_port: 8080\n  https_port: 8443\nphp:\n  default: 8.3\nports:\n  start_range: 8080\n  end_range: 8099\n  conflict_resolution: prompt   # prompt|auto|fail\n  nginx_http_fallback: 8080\n  nginx_https_fallback: 8443\n  php_fpm_fallback: 9000\nprojects_dir: ~/.chauffeur/projects`}/>
                <h3 className="text-xl font-bold text-white mb-2"><code>project.yaml</code> (Per-Project State)</h3>
                <CodeBlock code={`version: 1\npath: /absolute/path/to/project\nphp: 8.3\nsite:\n  domain: slug.test\n  ssl: false\ndomains:\n  aliases:\n    - domain: admin.test\n      ssl: true\n    - domain: api.test\n      ssl: false\nruntime:\n  php_fpm_socket: ~/.chauffeur/projects/<slug>/runtime/php-fpm/php-fpm.sock\ncreated_at: 2025-10-30T12:00:00+07:00`}/>
              </section>

              <section id="php-fpm-architecture">
                <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="php-fpm-architecture">
                  PHP-FPM Architecture
                  <Link href="#php-fpm-architecture" onClick={(e) => scrollToId(e, 'php-fpm-architecture')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
                </h2>
                <p className="text-slate-400 mb-4">
                  Chauffeur offers flexible PHP-FPM control, balancing resource efficiency with isolation needs.
                </p>
                <h3 className="text-xl font-bold text-white mb-2">Shared FPM (Default)</h3>
                <ul className="list-disc list-inside text-slate-400 space-y-2">
                    <li>Resource efficient: Multiple projects share the same PHP-FPM pool per PHP version.</li>
                    <li>Default behavior: <code>chauf link</code> uses shared FPM unless <code>--dedicated-fpm</code> is specified.</li>
                    <li>Socket path example: <code>~/.chauffeur/php/8.3/runtime/php-fpm/php-fpm.sock</code></li>
                </ul>
                <h3 className="text-xl font-bold text-white mb-2 mt-4">Dedicated FPM (Optional)</h3>
                <ul className="list-disc list-inside text-slate-400 space-y-2">
                    <li>Maximum isolation: Each project gets its own PHP-FPM pool.</li>
                    <li>Usage: <code>chauf link --dedicated-fpm</code> for critical projects.</li>
                    <li>Socket path example: <code>~/.chauffeur/projects/&lt;slug&gt;/runtime/php-fpm/php-fpm.sock</code></li>
                </ul>
                <p className="text-slate-400 mt-4">
                  Chauffeur supports a mixed strategy, automatically routing Nginx to the correct socket based on project configuration.
                </p>
              </section>

              <section id="multi-domain-architecture">
                <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="multi-domain-architecture">
                  Multi-Domain Architecture
                  <Link href="#multi-domain-architecture" onClick={(e) => scrollToId(e, 'multi-domain-architecture')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
                </h2>
                <p className="text-slate-400 mb-4">
                  Chauffeur supports multiple domains per project, featuring isolated SSL certificate management and alias-specific configuration.
                </p>
                <h3 className="text-xl font-bold text-white mb-2">SSL Certificate Management</h3>
                <ul className="list-disc list-inside text-slate-400 space-y-2">
                    <li>Multi-domain SAN certificates cover all SSL-enabled domains.</li>
                    <li>Supports mkcert (trusted) and self-signed certificates.</li>
                    <li>Certificates are automatically regenerated upon alias changes.</li>
                    <li>File naming uses the base domain (e.g., <code>hja-cms.test.crt</code>).</li>
                </ul>
                <h3 className="text-xl font-bold text-white mb-2 mt-4">Domain Resolution</h3>
                <ul className="list-disc list-inside text-slate-400 space-y-2">
                    <li>Primary domain from <code>proj.Site.Domain</code>.</li>
                    <li>Alias domains from <code>proj.Domains.Aliases</code> array.</li>
                    <li>Helper methods provide <code>GetAllDomains()</code>, <code>GetServerNames()</code>, <code>HasSSLEnabled()</code> for Nginx integration.</li>
                </ul>
              </section>


              {/* Page Footer Navigation */}
              <div className="mt-16 pt-8 border-t border-slate-800 flex justify-between">
                 <div className="text-left">
                   <div className="text-xs text-slate-500 mb-1">Previous</div>
                   <Link href="/docs/getting-started/first-project" className="text-primary hover:underline">First Project</Link>
                 </div>
                 <div className="text-right">
                   <div className="text-xs text-slate-500 mb-1">Next</div>
                   <Link href="/docs/core/linking" className="text-primary hover:underline">Project Linking</Link>
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

export default ArchitecturePage;