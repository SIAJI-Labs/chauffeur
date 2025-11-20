"use client";

// React & Next.js
import React, { useEffect } from 'react';
import Link from 'next/link';

// Third-party libraries
import {
  ChevronRight,
  Package,
  GitBranch,
  Zap,
  Layers,
  Shield,
  Settings
} from 'lucide-react';

// Page-specific components
import { TableOfContents } from '@/app/docs/_components/TableOfContents';
import { CodeBlock } from '@/app/docs/_components/CodeBlock';

export default function ComposerPage() {
  const currentSlug = 'core/composer';

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
        <span className="text-slate-200 capitalize">Composer</span>
      </div>

      {/* Content */}
      <div className="prose prose-invert prose-slate max-w-none space-y-8 animate-fade-in">
        <div>
          <h1 className="text-4xl font-bold text-white mb-4">Composer</h1>
          <p className="text-lg text-slate-400 leading-relaxed">
            Chauffeur provides Composer with intelligent PHP version isolation, ensuring each project uses the correct PHP version automatically.
          </p>
        </div>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="version-isolation">
            Version-Aware Dependency Management
            <Link href="#version-isolation" onClick={(e) => scrollToId(e, 'version-isolation')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">
            Composer automatically respects your project's PHP version, eliminating version conflicts and dependency management issues.
          </p>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mt-6">
            <div className="p-6 bg-surface rounded-xl border border-slate-800">
              <div className="flex items-center gap-3 mb-4">
                <div className="p-2 bg-blue-500/10 rounded-lg">
                  <Layers className="text-blue-400" size={20} />
                </div>
                <h3 className="font-semibold text-white">Project-Aware Composer</h3>
              </div>
              <p className="text-slate-400">
                Automatically detects and uses the PHP version configured for each project, ensuring compatibility.
              </p>
            </div>

            <div className="p-6 bg-surface rounded-xl border border-slate-800">
              <div className="flex items-center gap-3 mb-4">
                <div className="p-2 bg-emerald-500/10 rounded-lg">
                  <Package className="text-emerald-400" size={20} />
                </div>
                <h3 className="font-semibold text-white">Global Availability</h3>
              </div>
              <p className="text-slate-400">
                Composer commands available system-wide via PATH shims while maintaining workspace isolation.
              </p>
            </div>

            <div className="p-6 bg-surface rounded-xl border border-slate-800">
              <div className="flex items-center gap-3 mb-4">
                <div className="p-2 bg-purple-500/10 rounded-lg">
                  <GitBranch className="text-purple-400" size={20} />
                </div>
                <h3 className="font-semibold text-white">Version Switching</h3>
              </div>
              <p className="text-slate-400">
                Seamlessly switches PHP versions based on project context without manual intervention.
              </p>
            </div>

            <div className="p-6 bg-surface rounded-xl border border-slate-800">
              <div className="flex items-center gap-3 mb-4">
                <div className="p-2 bg-amber-500/10 rounded-lg">
                  <Shield className="text-amber-400" size={20} />
                </div>
                <h3 className="font-semibold text-white">Workspace Isolation</h3>
              </div>
              <p className="text-slate-400">
                Composer installation, cache, and configuration managed within Chauffeur workspace.
              </p>
            </div>
          </div>
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="automatic-version-detection">
            Automatic PHP Version Detection
            <Link href="#automatic-version-detection" onClick={(e) => scrollToId(e, 'automatic-version-detection')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">
            Chauffeur's Composer intelligently selects the correct PHP runtime based on your project configuration.
          </p>

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Smart Version Selection</h3>
          <CodeBlock code={`# Project with PHP 8.3 configured
my-laravel-app/
├── composer.json
├── .chauffeur/
│   └── project.yaml     # Contains "php: 8.3"
└── ...

# When you run Composer, Chauffeur automatically:
$ composer install
→ Detects project PHP: 8.3
→ Uses: ~/.chauffeur/php/8.3/bin/php
→ Installs packages compatible with PHP 8.3

$ composer update
→ Maintains PHP 8.3 compatibility
→ Updates packages within 8.3 constraints`} />

          <h3 className="text-xl font-semibold text-slate-200 mb-3 mt-6">Multi-Project Workflow</h3>
          <CodeBlock code={`# Work on different projects with different PHP versions
$ cd ~/projects/legacy-app      # PHP 7.4
$ composer install
→ Uses PHP 7.4 runtime
→ Installs packages compatible with PHP 7.4

$ cd ~/projects/modern-app      # PHP 8.3
$ composer install
→ Uses PHP 8.3 runtime
→ Installs packages compatible with PHP 8.3

$ cd ~/projects/api-app         # PHP 8.2
$ composer install
→ Uses PHP 8.2 runtime
→ Installs packages compatible with PHP 8.2`} />

          <h3 className="text-xl font-semibold text-slate-200 mb-3 mt-6">Global Composer Commands</h3>
          <CodeBlock code={`# Global Composer uses your global PHP version
$ chauf php use 8.2                    # Set global PHP to 8.2

$ composer global require laravel/installer
→ Uses PHP 8.2 from ~/.chauffeur/php/8.2/
→ Installs Laravel installer globally

$ composer global update
→ Updates all global packages with PHP 8.2
→ Maintains compatibility with global PHP version`} />
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="cache-management">
            Cache & Performance
            <Link href="#cache-management" onClick={(e) => scrollToId(e, 'cache-management')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">
            Chauffeur optimizes Composer performance with intelligent caching and workspace isolation.
          </p>

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Workspace Cache Structure</h3>
          <CodeBlock code={`~/.chauffeur/composer/
├── cache/
│   ├── repo/               # Package repository cache
│   ├── files/              # Downloaded package files
│   └── vcs/                # Version control cache
├── home/
│   ├── .composer/          # Composer home directory
│   └── auth.json           # Authentication tokens
├── global/
│   ├── composer.json       # Global dependencies
│   ├── composer.lock       # Global lock file
│   └── vendor/             # Global packages
└── shims/
    ├── composer            # Main Composer shim
    └── composer-*          # Version-specific shims`} />

          <h3 className="text-xl font-semibold text-slate-200 mb-3 mt-6">Performance Benefits</h3>
          <div className="space-y-4">
            <div className="p-4 bg-blue-500/10 border border-blue-500/20 rounded-lg">
              <h4 className="font-semibold text-blue-300 mb-2">⚡ Intelligent Caching</h4>
              <p className="text-sm text-blue-100/80">
                Packages cached once per PHP version, shared across all projects using that version.
              </p>
            </div>

            <div className="p-4 bg-emerald-500/10 border border-emerald-500/20 rounded-lg">
              <h4 className="font-semibold text-emerald-300 mb-2">🔒 Isolation Benefits</h4>
              <p className="text-sm text-emerald-100/80">
                No conflicts with system Composer or other PHP installations on the same machine.
              </p>
            </div>

            <div className="p-4 bg-purple-500/10 border border-purple-500/20 rounded-lg">
              <h4 className="font-semibold text-purple-300 mb-2">💾 Persistent Cache</h4>
              <p className="text-sm text-purple-100/80">
                Cache persists between workspace resets, dramatically speeding up subsequent installs.
              </p>
            </div>
          </div>
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="path-shims">
            PATH Shims & Integration
            <Link href="#path-shims" onClick={(e) => scrollToId(e, 'path-shims')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">
            Chauffeur creates smart shims that intelligently route Composer commands to the correct PHP version.
          </p>

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Shim Architecture</h3>
          <CodeBlock code={`# Main shims in ~/.chauffeur/bin/
~/.chauffeur/bin/
├── composer              # Intelligent Composer shim
├── php                   # Global PHP version shim
├── php-7.4              # PHP 7.4 specific shim
├── php-8.1              # PHP 8.1 specific shim
├── php-8.2              # PHP 8.2 specific shim
├── php-8.3              # PHP 8.3 specific shim
└── composer-8.3         # PHP 8.3 Composer shim (fallback)

# How shims work:
1. Detect current project context
2. Read project's PHP version from ~/.chauffeur/projects/<slug>/project.yaml
3. Route to appropriate PHP runtime
4. Execute Composer with correct PHP version`} />

          <h3 className="text-xl font-semibold text-slate-200 mb-3 mt-6">Environment Integration</h3>
          <CodeBlock code={`# Shell integration (automatic)
echo $PATH
# Output includes: ~/.chauffeur/bin

# Standard Composer commands work everywhere
$ composer install           # Uses project's PHP version
$ composer update           # Maintains compatibility
$ composer require package  # Version-aware installation
$ composer global require   # Uses global PHP version
$ composer show             # Shows package info
$ composer outdated          # Checks for updates`} />

          <h3 className="text-xl font-semibold text-slate-200 mb-3 mt-6">IDE Integration</h3>
          <CodeBlock code={`# VSCode integration example
{
    "php.executablePath": "~/.chauffeur/bin/php",
    "composer.executablePath": "~/.chauffeur/bin/composer"
}

# PhpStorm integration
# PHP Interpreter: ~/.chauffeur/bin/php
# Composer: ~/.chauffeur/bin/composer

# Automatic IDE benefits:
✅ Correct PHP version per project
✅ Accurate code completion
✅ Proper syntax highlighting
✅ Compatible debugging configuration`} />
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="authentication">
            Authentication & Private Packages
            <Link href="#authentication" onClick={(e) => scrollToId(e, 'authentication')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">
            Chauffeur manages Composer authentication securely within the workspace, supporting private packages and repositories.
          </p>

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Workspace Authentication</h3>
          <CodeBlock code={`# Authentication stored in workspace (isolated)
~/.chauffeur/composer/home/auth.json
{
    "github-oauth": {
        "github.com": "your-github-token"
    },
    "http-basic": {
        "repo.example.com": {
            "username": "your-username",
            "password": "your-password"
        }
    },
    "gitlab-oauth": {
        "gitlab.com": "your-gitlab-token"
    }
}

# Setup commands work normally
$ composer config --global github-oauth.github.com your-token
$ composer config --global http-basic.repo.example.com username password
$ composer config --global gitlab-oauth.gitlab.com your-token`} />

          <h3 className="text-xl font-semibold text-slate-200 mb-3 mt-6">Private Repository Integration</h3>
          <CodeBlock code={`# composer.json with private repositories
{
    "require": {
        "your-private/package": "^1.0"
    },
    "repositories": [
        {
            "type": "vcs",
            "url": "git@github.com:your-org/private-repo.git"
        },
        {
            "type": "composer",
            "url": "https://repo.packagist.com/your-org"
        }
    ],
    "config": {
        "allow-plugins": {
            "private/plugin": true
        }
    }
}

# Chauffeur handles authentication automatically
$ composer install
→ Uses stored credentials for private repos
→ Downloads packages with correct PHP version
→ Maintains workspace isolation`} />
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="vs-system-composer">
            vs System Composer
            <Link href="#vs-system-composer" onClick={(e) => scrollToId(e, 'vs-system-composer')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">
            Chauffeur's Composer integration provides significant advantages over using system Composer for multi-version PHP development.
          </p>

          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm border-collapse">
              <thead>
                <tr className="border-b border-slate-700 bg-slate-800/50">
                  <th className="p-3 font-mono text-emerald-400">Feature</th>
                  <th className="p-3 text-slate-300">Chauffeur Composer</th>
                  <th className="p-3 text-slate-300">System Composer</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-800">
                <tr>
                  <td className="p-3 font-mono text-slate-300">PHP Version</td>
                  <td className="p-3 text-slate-400">✅ Per-project automatic</td>
                  <td className="p-3 text-slate-400">❌ System-wide single</td>
                </tr>
                <tr>
                  <td className="p-3 font-mono text-slate-300">Cache Isolation</td>
                  <td className="p-3 text-slate-400">✅ PHP version isolated</td>
                  <td className="p-3 text-slate-400">❌ Shared cache</td>
                </tr>
                <tr>
                  <td className="p-3 font-mono text-slate-300">Global Packages</td>
                  <td className="p-3 text-slate-400">✅ Workspace isolated</td>
                  <td className="p-3 text-slate-400">❌ System-wide</td>
                </tr>
                <tr>
                  <td className="p-3 font-mono text-slate-300">Authentication</td>
                  <td className="p-3 text-slate-400">✅ Workspace secure</td>
                  <td className="p-3 text-slate-400">⚠️ System-wide</td>
                </tr>
                <tr>
                  <td className="p-3 font-mono text-slate-300">Version Conflicts</td>
                  <td className="p-3 text-slate-400">✅ Zero conflicts</td>
                  <td className="p-3 text-slate-400">❌ Common issue</td>
                </tr>
                <tr>
                  <td className="p-3 font-mono text-slate-300">IDE Integration</td>
                  <td className="p-3 text-slate-400">✅ Per-project aware</td>
                  <td className="p-3 text-slate-400">❌ Single version</td>
                </tr>
              </tbody>
            </table>
          </div>

          <div className="mt-6 p-4 bg-amber-500/10 border border-amber-500/20 rounded-lg">
            <h4 className="font-semibold text-amber-300 mb-2">🎯 Key Advantage</h4>
            <p className="text-sm text-amber-100/80">
              Chauffeur enables <strong>seamless multi-version PHP development</strong> where Composer automatically adapts to each project's PHP version, eliminating the classic "dependency hell" problems.
            </p>
          </div>
        </section>

        {/* Page Footer Navigation */}
        <div className="mt-16 pt-8 border-t border-slate-800 flex justify-between">
          <div className="text-left">
            <div className="text-xs text-slate-500 mb-1">Previous</div>
            <Link href="/docs/core/nginx" className="text-primary hover:underline">Nginx</Link>
          </div>
          <div className="text-right">
            <div className="text-xs text-slate-500 mb-1">Next</div>
            <Link href="/docs/core/ssl-domains" className="text-primary hover:underline">SSL & Domains</Link>
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