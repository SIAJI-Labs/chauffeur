"use client";

// React & Next.js
import React, { useEffect } from 'react';
import Link from 'next/link';

// Third-party libraries
import {
  ChevronRight,
  Layers,
  Zap,
  Cpu,
  Package,
  Terminal,
  Download,
  AlertTriangle,
  CheckCircle2
} from 'lucide-react';

// Page-specific components
import { TableOfContents } from '@/app/docs/_components/TableOfContents';
import { CodeBlock } from '@/app/docs/_components/CodeBlock';

export default function PHPVersionsPage() {
  const currentSlug = 'core/php-versions';

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
        <span className="text-slate-200 capitalize">PHP Versions</span>
      </div>

      {/* Content */}
      <div className="prose prose-invert prose-slate max-w-none space-y-8 animate-fade-in">
        <div>
          <h1 className="text-4xl font-bold text-white mb-4">PHP Versions</h1>
          <p className="text-lg text-slate-400 leading-relaxed">
            Chauffeur provides seamless multi-version PHP support, allowing you to run different PHP versions simultaneously without conflicts.
          </p>
        </div>

        <div className="p-4 bg-blue-500/10 border border-blue-500/20 rounded-lg flex gap-3">
          <Layers className="text-blue-400 shrink-0" />
          <div className="text-sm text-blue-100/80">
            <strong>Key Feature:</strong> Chauffeur ships with its own PHP runtimes, completely isolated from your system PHP. You can have PHP 8.4 globally, 8.3 for one project, and 7.4 for legacy applications - all running simultaneously.
          </div>
        </div>

        <section id="supported-versions">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="supported-versions">
            Supported PHP Versions
            <Link href="#supported-versions" onClick={(e) => scrollToId(e, 'supported-versions')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">Chauffeur supports the following PHP versions:</p>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            <div className="bg-surface p-4 rounded-lg border border-slate-800">
              <div className="flex items-center justify-between mb-2">
                <h3 className="font-semibold text-white">PHP 8.4</h3>
                <span className="text-xs bg-emerald-500/20 text-emerald-300 px-2 py-1 rounded border border-emerald-500/30">Latest</span>
              </div>
              <p className="text-slate-400 text-sm">Latest stable release with performance improvements and new features.</p>
            </div>

            <div className="bg-surface p-4 rounded-lg border border-slate-800">
              <div className="flex items-center justify-between mb-2">
                <h3 className="font-semibold text-white">PHP 8.3</h3>
                <span className="text-xs bg-emerald-500/20 text-emerald-300 px-2 py-1 rounded border border-emerald-500/30">LTS</span>
              </div>
              <p className="text-slate-400 text-sm">Stable version with performance improvements and modern features.</p>
            </div>

            <div className="bg-surface p-4 rounded-lg border border-slate-800">
              <div className="flex items-center justify-between mb-2">
                <h3 className="font-semibold text-white">PHP 8.2</h3>
                <span className="text-xs bg-emerald-500/20 text-emerald-300 px-2 py-1 rounded border border-emerald-500/30">LTS</span>
              </div>
              <p className="text-slate-400 text-sm">Stable version with modern language features and security updates.</p>
            </div>

            <div className="bg-surface p-4 rounded-lg border border-slate-800">
              <div className="flex items-center justify-between mb-2">
                <h3 className="font-semibold text-white">PHP 8.1</h3>
                <span className="text-xs bg-emerald-500/20 text-emerald-300 px-2 py-1 rounded border border-emerald-500/30">LTS</span>
              </div>
              <p className="text-slate-400 text-sm">Actively supported version with modern language features.</p>
            </div>

            <div className="bg-surface p-4 rounded-lg border border-slate-800">
              <div className="flex items-center justify-between mb-2">
                <h3 className="font-semibold text-white">PHP 8.0</h3>
                <span className="text-xs bg-amber-500/20 text-amber-300 px-2 py-1 rounded border border-amber-500/30">EOL</span>
              </div>
              <p className="text-slate-400 text-sm">End of life. Available for compatibility with legacy applications.</p>
            </div>

            <div className="bg-surface p-4 rounded-lg border border-slate-800">
              <div className="flex items-center justify-between mb-2">
                <h3 className="font-semibold text-white">PHP 7.4</h3>
                <span className="text-xs bg-red-500/20 text-red-300 px-2 py-1 rounded border border-red-500/30">EOL</span>
              </div>
              <p className="text-slate-400 text-sm">End of life. Available for compatibility with legacy applications.</p>
            </div>
          </div>
        </section>

        <section id="installation">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="installation">
            Installing PHP Versions
            <Link href="#installation" onClick={(e) => scrollToId(e, 'installation')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">Chauffeur downloads and compiles PHP versions on demand:</p>

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Install a Specific Version</h3>
          <CodeBlock code="chauf php install 8.3" />

          <div className="bg-amber-500/10 border border-amber-500/20 rounded-lg p-4 my-4 flex gap-3">
            <Download className="text-amber-400 shrink-0" />
            <div className="text-sm text-amber-100/80">
              <strong>First Installation:</strong> The initial PHP installation may take 10-15 minutes as Chauffeur compiles from source. Subsequent installations are much faster.
            </div>
          </div>

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Installation Options</h3>
          <CodeBlock code="chauf php install 8.2 --force        # Reinstall if exists
chauf php install 8.1 --no-ext        # Skip extensions
chauf php install 8.3 --from source   # Force compilation from source" />

          <h3 className="text-xl font-semibold text-slate-200 mb-3">List Installed Versions</h3>
          <CodeBlock code="$ chauf php list
┌─────────────┬─────────────────┬───────────┬─────────────┐
│ Version     │ Path            │ Status    │ Size        │
├─────────────┼─────────────────┼───────────┼─────────────┤
│ 8.3         │ ~/.chauffeur/... │ active    │ 145 MB      │
│ 8.2         │ ~/.chauffeur/... │ active    │ 138 MB      │
│ 8.1         │ ~/.chauffeur/... │ active    │ 132 MB      │
└─────────────┴─────────────────┴───────────┴─────────────┘" />
        </section>

        <section id="global-version">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="global-version">
            Global PHP Version
            <Link href="#global-version" onClick={(e) => scrollToId(e, 'global-version')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">Set the default PHP version for new projects:</p>

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Set Global Version</h3>
          <CodeBlock code="chauf php isolate 8.3" />

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Check Current Global Version</h3>
          <CodeBlock code="$ chauf php current
Global PHP version: 8.3
Path: ~/.chauffeur/php/8.3/bin/php" />

          <h3 className="text-xl font-semibold text-slate-200 mb-3">CLI Integration</h3>
          <p className="text-slate-400 mb-4">Chauffeur adds PHP binaries to your PATH:</p>
          <CodeBlock code="$ php --version
PHP 8.3.15 (cli) (built: Dec 15 2024 10:30:00)
Copyright (c) The PHP Group
Zend Engine v4.3.15, Copyright (c) Zend Technologies

$ composer --version
Composer version 2.7.7 2024-12-15 10:30:00" />
        </section>

        <section id="project-specific">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="project-specific">
            Project-Specific PHP Versions
            <Link href="#project-specific" onClick={(e) => scrollToId(e, 'project-specific')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">Override the global PHP version for specific projects:</p>

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Link with Specific PHP Version</h3>
          <CodeBlock code="chauf link --php=8.1" />

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Change Existing Project's PHP Version</h3>
          <CodeBlock code="cd ~/projects/legacy-app
chauf isolate 7.4" />

          <div className="bg-emerald-500/10 border border-emerald-500/20 rounded-lg p-4 my-4 flex gap-3">
            <CheckCircle2 className="text-emerald-400 shrink-0" />
            <div className="text-sm text-emerald-100/80">
              <strong>Isolation Benefits:</strong> Each project gets its own PHP-FPM pool with the specified version, ensuring complete isolation from other projects.
            </div>
          </div>
        </section>

        <section id="fpm-architecture">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="fpm-architecture">
            PHP-FPM Architecture
            <Link href="#fpm-architecture" onClick={(e) => scrollToId(e, 'fpm-architecture')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">Chauffeur uses PHP-FPM for optimal performance with flexible isolation options:</p>

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Shared PHP-FPM (Default)</h3>
          <div className="bg-surface p-4 rounded-lg border border-slate-800 mb-4">
            <h4 className="font-semibold text-white mb-2 flex items-center gap-2">
              <Zap size={16} className="text-emerald-400" />
              Resource Efficient
            </h4>
            <p className="text-slate-400 text-sm mb-2">Multiple projects share the same PHP-FPM pool per PHP version:</p>
            <CodeBlock code="~/.chauffeur/php/8.3/runtime/php-fpm/php-fpm.sock  ← All 8.3 projects
~/.chauffeur/php/8.2/runtime/php-fpm/php-fpm.sock  ← All 8.2 projects
~/.chauffeur/php/8.1/runtime/php-fpm/php-fpm.sock  ← All 8.1 projects" />
          </div>

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Dedicated PHP-FPM (Optional)</h3>
          <div className="bg-surface p-4 rounded-lg border border-slate-800 mb-4">
            <h4 className="font-semibold text-white mb-2 flex items-center gap-2">
              <Package size={16} className="text-blue-400" />
              Maximum Isolation
            </h4>
            <p className="text-slate-400 text-sm mb-2">Each project gets its own PHP-FPM pool:</p>
            <CodeBlock code="~/.chauffeur/projects/blog/runtime/php-fpm/php-fpm.sock
~/.chauffeur/projects/api/runtime/php-fpm/php-fpm.sock
~/.chauffeur/projects/admin/runtime/php-fpm/php-fpm.sock" />
          </div>

          <h3 className="text-xl font-semibold text-slate-200 mb-3">When to Use Dedicated FPM</h3>
          <ul className="list-disc list-inside text-slate-400 space-y-2">
            <li>Applications requiring custom php.ini settings</li>
            <li>Memory-intensive applications that need dedicated resources</li>
            <li>Security-sensitive applications requiring isolation</li>
            <li>Applications with conflicting extension requirements</li>
          </ul>
        </section>

        <section id="extensions">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="extensions">
            PHP Extensions
            <Link href="#extensions" onClick={(e) => scrollToId(e, 'extensions')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">Chauffeur includes common PHP extensions by default:</p>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
            <div className="bg-surface p-3 rounded-lg border border-slate-800">
              <h4 className="font-semibold text-white text-sm mb-2">Core Extensions</h4>
              <ul className="text-xs text-slate-400 space-y-1">
                <li>• curl, gd, intl, json, mbstring</li>
                <li>• mysqlnd, mysqli, pdo, pdo_mysql</li>
                <li>• opcache, openssl, tokenizer</li>
                <li>• xml, xmlwriter, xmlreader</li>
              </ul>
            </div>

            <div className="bg-surface p-3 rounded-lg border border-slate-800">
              <h4 className="font-semibold text-white text-sm mb-2">Development Extensions</h4>
              <ul className="text-xs text-slate-400 space-y-1">
                <li>• xdebug (development only)</li>
                <li>• blackfire (performance profiling)</li>
                <li>• redis, memcached (caching)</li>
                <li>• zip, bcmath, calendar</li>
              </ul>
            </div>
          </div>

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Check Loaded Extensions</h3>
          <CodeBlock code="$ php -m
[PHP Modules]
Core, date, libxml, openssl, pcre, zlib, filter, hash, json,
apcu, bcmath, bz2, calendar, cgi-fcgi, ctype, curl, dom, fileinfo,
ftp, gd, gettext, iconv, igbinary, imagick, imap, intl, mbstring,
mcrypt, memcached, mongodb, msgpack, mysqli, pdo, pdo_mysql,
pdo_pgsql, pdo_sqlite, pgsql, phalcon, pspell, redis, session,
shmop, simplexml, soap, sockets, sqlite3, sysvmsg, sysvsem,
sysvshm, tidy, tokenizer, wddx, xml, xmlreader, xmlrpc,
xmlwriter, xsl, yaml, zip, zlib, xdebug" />

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Custom Extensions</h3>
          <p className="text-slate-400 mb-4">For additional extensions, you can modify the PHP configuration:</p>
          <CodeBlock code="# Edit PHP configuration
~/.chauffeur/php/8.3/etc/php.ini

# Reload services after changes
chauf restart" />
        </section>

        <section id="troubleshooting">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="troubleshooting">
            Troubleshooting PHP Issues
            <Link href="#troubleshooting" onClick={(e) => scrollToId(e, 'troubleshooting')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>

          <div className="space-y-4">
            <div className="bg-red-500/10 border border-red-500/20 rounded-lg p-4">
              <h4 className="font-semibold text-red-300 mb-2 flex items-center gap-2">
                <AlertTriangle size={16} />
                PHP Version Not Available
              </h4>
              <CodeBlock code="# Install missing version
chauf php install 8.3

# Check what's installed
chauf php list" />
            </div>

            <div className="bg-amber-500/10 border border-amber-500/20 rounded-lg p-4">
              <h4 className="font-semibold text-amber-300 mb-2 flex items-center gap-2">
                <AlertTriangle size={16} />
                Extension Missing
              </h4>
              <CodeBlock code="# Check if extension is loaded
php -m | grep extension_name

# Reinstall PHP with all extensions
chauf php install 8.3 --force" />
            </div>

            <div className="bg-blue-500/10 border border-blue-500/20 rounded-lg p-4">
              <h4 className="font-semibold text-blue-300 mb-2 flex items-center gap-2">
                <Terminal size={16} />
                Performance Issues
              </h4>
              <CodeBlock code="# Check PHP-FPM status
chauf status

# Check process memory usage
ps aux | grep php-fpm

# Enable/disable OPcache
php -d opcache.enable=1 -m | grep opcache" />
            </div>
          </div>
        </section>

        {/* Page Footer Navigation */}
        <div className="mt-16 pt-8 border-t border-slate-800 flex justify-between">
          <div className="text-left">
            <div className="text-xs text-slate-500 mb-1">Previous</div>
            <Link href="/docs/core/linking" className="text-primary hover:underline">Project Linking</Link>
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