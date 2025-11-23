"use client";

// React & Next.js
import React, { useEffect } from 'react';
import Link from 'next/link';

// Third-party libraries
import {
  ChevronRight,
  FolderOpen,
  Link2,
  Shield,
  Globe,
  Terminal
} from 'lucide-react';

// Page-specific components
import { TableOfContents } from '@/app/docs/_components/TableOfContents';
import { CodeBlock } from '@/app/docs/_components/CodeBlock';

export default function ProjectLinkingPage() {
  const currentSlug = 'core/linking';

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
        <Link href="/docs/core" className="hover:text-primary transition-colors">Core Concepts</Link>
        <ChevronRight size={14} />
        <span className="text-slate-200 capitalize">Project Linking</span>
      </div>

      {/* Content */}
      <div className="prose prose-invert prose-slate max-w-none space-y-8 animate-fade-in">
        <div>
          <h1 className="text-4xl font-bold text-white mb-4">Project Linking</h1>
          <p className="text-lg text-slate-400 leading-relaxed">
            Understanding how Chauffeur connects your local directories to web domains and manages project configurations.
          </p>
        </div>

        <div className="p-4 bg-blue-500/10 border border-blue-500/20 rounded-lg flex gap-3">
          <FolderOpen className="text-blue-400 shrink-0" />
          <div className="text-sm text-blue-100/80">
            <strong>Core Concept:</strong> Project linking is Chauffeur's way of mapping directories to domains. Unlike traditional setups that require manual Nginx configuration, Chauffeur handles this automatically.
          </div>
        </div>

        <section id="what-is-linking">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="what-is-linking">
            What is Project Linking?
            <Link href="#what-is-linking" onClick={(e) => scrollToId(e, 'what-is-linking')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">
            Project linking is the process of registering a directory with Chauffeur so it can be served through a web browser. When you link a project, Chauffeur:
          </p>
          <ul className="list-disc list-inside text-slate-400 space-y-2">
            <li>Creates a domain mapping (e.g., <code>my-project.test</code>)</li>
            <li>Configures Nginx to serve the directory</li>
            <li>Sets up PHP-FPM processing for PHP files</li>
            <li>Creates a project configuration file</li>
            <li>Optional: Generates SSL certificates for HTTPS</li>
          </ul>
        </section>

        <section id="basic-linking">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="basic-linking">
            Basic Linking
            <Link href="#basic-linking" onClick={(e) => scrollToId(e, 'basic-linking')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">The simplest way to link a project is to navigate to your project directory and run:</p>
          <CodeBlock code="cd ~/projects/my-app
chauf link" />
          <p className="text-slate-400 mt-4 mb-4">This will:</p>
          <ul className="list-disc list-inside text-slate-400 space-y-2">
            <li>Use the directory name as the domain (<code>my-app.test</code>)</li>
            <li>Use the global PHP version</li>
            <li>Create HTTP access only (no SSL by default)</li>
            <li>Use shared PHP-FPM for efficiency</li>
          </ul>
        </section>

        <section id="advanced-linking">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="advanced-linking">
            Advanced Linking Options
            <Link href="#advanced-linking" onClick={(e) => scrollToId(e, 'advanced-linking')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Custom Domain Name</h3>
          <CodeBlock code="chauf link --name=admin-dashboard" />
          <p className="text-slate-400 text-sm mt-2">Creates: <code>http://admin-dashboard.test</code></p>

          <h3 className="text-xl font-semibold text-slate-200 mb-3 mt-6">Automatic SSL</h3>
          <CodeBlock code="chauf link --secure" />
          <p className="text-slate-400 text-sm mt-2">Generates trusted SSL certificate for HTTPS access</p>

          <h3 className="text-xl font-semibold text-slate-200 mb-3 mt-6">Specific PHP Version</h3>
          <CodeBlock code="chauf link --php=8.1" />
          <p className="text-slate-400 text-sm mt-2">Isolates this project to PHP 8.1, regardless of global version</p>

          <h3 className="text-xl font-semibold text-slate-200 mb-3 mt-6">Dedicated PHP-FPM Pool</h3>
          <CodeBlock code="chauf link --dedicated-fpm" />
          <p className="text-slate-400 text-sm mt-2">Creates an isolated PHP-FPM pool for maximum separation</p>

          <h3 className="text-xl font-semibold text-slate-200 mb-3 mt-6">Combined Options</h3>
          <CodeBlock code="chauf link --name=api --secure --php=8.2 --dedicated-fpm" />
        </section>

        <section id="project-structure">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="project-structure">
            Project Structure Requirements
            <Link href="#project-structure" onClick={(e) => scrollToId(e, 'project-structure')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">Chauffeur works with any project structure, but here are the common patterns:</p>

          <div className="space-y-4">
            <div className="bg-surface p-4 rounded-lg border border-slate-800">
              <h4 className="font-semibold text-white mb-2 flex items-center gap-2">
                <Terminal size={16} className="text-emerald-400" />
                Laravel/Symfony Projects
              </h4>
              <CodeBlock code="my-laravel-app/
├── public/          ← Document root
├── app/
├── config/
└── .env" />
            </div>

            <div className="bg-surface p-4 rounded-lg border border-slate-800">
              <h4 className="font-semibold text-white mb-2 flex items-center gap-2">
                <Terminal size={16} className="text-emerald-400" />
                WordPress Projects
              </h4>
              <CodeBlock code="my-wp-site/
├── index.php       ← Document root
├── wp-content/
└── wp-config.php" />
            </div>

            <div className="bg-surface p-4 rounded-lg border border-slate-800">
              <h4 className="font-semibold text-white mb-2 flex items-center gap-2">
                <Terminal size={16} className="text-emerald-400" />
                Static/SPA Projects
              </h4>
              <CodeBlock code="my-spa/
├── index.html
├── dist/
└── assets/" />
            </div>
          </div>
        </section>

        <section id="managing-links">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="managing-links">
            Managing Linked Projects
            <Link href="#managing-links" onClick={(e) => scrollToId(e, 'managing-links')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>

          <h3 className="text-xl font-semibold text-slate-200 mb-3">List All Projects</h3>
          <CodeBlock code="$ chauf status
┌─────────────────────┬─────────────┬──────┬───────┬─────────────┐
│ Project             │ Domain      │ PHP  │ SSL   │ FPM         │
├─────────────────────┼─────────────┼──────┼───────┼─────────────┤
│ ~/projects/blog     │ blog.test   │ 8.3  │ false │ shared      │
│ ~/projects/api      │ api.test    │ 8.1  │ true  │ dedicated   │
│ ~/projects/admin    │ admin.test  │ 8.2  │ true  │ shared      │
└─────────────────────┴─────────────┴──────┴───────┴─────────────┘" />

          <h3 className="text-xl font-semibold text-slate-200 mb-3 mt-6">Unlink Projects</h3>
          <CodeBlock code="# Unlink current directory
chauf unlink

# Unlink specific project by domain
chauf unlink --site=blog.test

# Unlink all projects
chauf unlink --all" />

          <h3 className="text-xl font-semibold text-slate-200 mb-3 mt-6">Project Configuration</h3>
          <p className="text-slate-400 mb-4">Each linked project has a configuration file at:</p>
          <CodeBlock code="~/.chauffeur/projects/[project-slug]/project.yaml" />
          <p className="text-slate-400 mt-2">Example configuration:</p>
          <CodeBlock code="version: 1
path: /home/user/projects/my-app
php: 8.3
site:
  domain: my-app.test
  ssl: true
domains:
  aliases:
    - domain: api.my-app.test
      ssl: true
    - domain: admin.my-app.test
      ssl: false
runtime:
  php_fpm_socket: ~/.chauffeur/projects/my-app/runtime/php-fpm/php-fpm.sock
created_at: 2025-01-15T10:30:00+07:00" />
        </section>

        <section id="domain-resolution">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="domain-resolution">
            Domain Resolution
            <Link href="#domain-resolution" onClick={(e) => scrollToId(e, 'domain-resolution')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">Chauffeur automatically configures local DNS resolution for <code>.test</code> domains:</p>

          <div className="bg-amber-500/10 border border-amber-500/20 rounded-lg p-4 mb-4">
            <h4 className="font-semibold text-amber-300 mb-2">🔧 Automatic Setup</h4>
            <p className="text-amber-100/80 text-sm">
              Chauffeur configures <code>dnsmasq</code> to resolve all <code>*.test</code> domains to <code>127.0.0.1</code>. This happens automatically during installation.
            </p>
          </div>

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Port Management</h3>
          <p className="text-slate-400 mb-4">If standard ports (80, 443) are occupied, Chauffeur automatically:</p>
          <ul className="list-disc list-inside text-slate-400 space-y-2">
            <li>Detects port conflicts</li>
            <li>Maps to available ports (e.g., 8080, 8443)</li>
            <li>Sets up iptables rules for transparent redirection</li>
            <li>Users still access via standard ports</li>
          </ul>
        </section>

        {/* Page Footer Navigation */}
        <div className="mt-16 pt-8 border-t border-slate-800 flex justify-between">
          <div className="text-left">
            <div className="text-xs text-slate-500 mb-1">Previous</div>
            <Link href="/docs/getting-started/architecture" className="text-primary hover:underline">Architecture</Link>
          </div>
          <div className="text-right">
            <div className="text-xs text-slate-500 mb-1">Next</div>
            <Link href="/docs/core/php-versions" className="text-primary hover:underline">PHP Versions</Link>
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