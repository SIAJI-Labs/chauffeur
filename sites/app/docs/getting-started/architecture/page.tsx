"use client";

// React & Next.js
import React, { useEffect } from 'react';
import Link from 'next/link';

// Third-party libraries
import { ChevronRight } from 'lucide-react';

// Page-specific components
import { TableOfContents } from '@/app/docs/_components/TableOfContents';
import { CodeBlock } from '@/app/docs/_components/CodeBlock';

export default function ArchitecturePage() {
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
          <h1 className="text-4xl font-bold text-white mb-4">Chauffeur Architecture</h1>
          <p className="text-lg text-slate-400 leading-relaxed">
            Understanding the core design principles and architecture of Chauffeur.
          </p>
        </div>

        <section>
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

        <section>
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

        <section>
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

        <section>
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

        <section>
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="nginx-architecture">
            Nginx Architecture
            <Link href="#nginx-architecture" onClick={(e) => scrollToId(e, 'nginx-architecture')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">
            Chauffeur manages its own Nginx installation with optimized configurations for local development.
          </p>

          <h3 className="text-xl font-bold text-white mb-2">Self-Contained Nginx</h3>
          <ul className="list-disc list-inside text-slate-400 space-y-2">
            <li><strong>Workspace isolation:</strong> Nginx binary and configurations reside under <code>~/.chauffeur/nginx/</code></li>
            <li><strong>Non-conflicting service:</strong> Runs as <code>chauf-nginx</code> to avoid conflicts with system nginx</li>
            <li><strong>Port forwarding:</strong> Default ports 8080 (HTTP) and 8443 (HTTPS) with optional system port forwarding</li>
            <li><strong>Template-based configs:</strong> Nginx configurations generated from templates for each project</li>
          </ul>

          <h3 className="text-xl font-bold text-white mb-2 mt-4">Configuration Management</h3>
          <ul className="list-disc list-inside text-slate-400 space-y-2">
            <li><strong>sites-available/:</strong> Template configurations for each linked project</li>
            <li><strong>sites-enabled/:</strong> Active configurations (symlinks to sites-available)</li>
            <li><strong>Dynamic generation:</strong> Configurations auto-generated from project state</li>
            <li><strong>SSL integration:</strong> Automatic SSL certificate configuration and reload</li>
          </ul>

          <h3 className="text-xl font-bold text-white mb-2 mt-4">Port Management</h3>
          <CodeBlock code={`# Default workspace ports (no sudo required)
HTTP:  8080  → Nginx serves sites
HTTPS: 8443  → Nginx serves SSL sites

# Optional system port forwarding (sudo required once)
80  → 8080  (HTTP)
443 → 8443  (HTTPS)

# Dynamic port allocation for conflicts
Range: 8080-8099
Auto-assigns free port if 8080/8443 occupied`}/>

          <div className="mt-6 p-4 bg-blue-500/10 border border-blue-500/20 rounded-lg">
            <h4 className="font-semibold text-blue-300 mb-2">🌐 Nginx Benefits</h4>
            <div className="text-sm text-blue-100/80">
              <p>• <strong>Zero system impact:</strong> Won't interfere with existing web servers</p>
              <p>• <strong>Consistent environment:</strong> Same Nginx version and modules across projects</p>
              <p>• <strong>Easy configuration:</strong> Optimized templates for common frameworks</p>
              <p>• <strong>SSL handling:</strong> Automatic certificate management and HTTPS redirection</p>
            </div>
          </div>
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="composer-architecture">
            Composer Architecture
            <Link href="#composer-architecture" onClick={(e) => scrollToId(e, 'composer-architecture')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">
            Chauffeur provides Composer with intelligent PHP version isolation and workspace integration.
          </p>

          <h3 className="text-xl font-bold text-white mb-2">Version-Isolated Composer</h3>
          <ul className="list-disc list-inside text-slate-400 space-y-2">
            <li><strong>Project-aware Composer:</strong> Automatically uses the project's configured PHP version</li>
            <li><strong>Global availability:</strong> Composer command available system-wide via PATH shims</li>
            <li><strong>Version switching:</strong> Composer respects per-project PHP version settings</li>
            <li><strong>Workspace isolation:</strong> Composer installation and cache managed within workspace</li>
          </ul>

          <h3 className="text-xl font-bold text-white mb-2 mt-4">PHP Version Integration</h3>
          <CodeBlock code={`# Project with PHP 8.3
$ composer install
→ Uses PHP 8.3 runtime from ~/.chauffeur/php/8.3/

# Project with PHP 8.1
$ composer update
→ Uses PHP 8.1 runtime from ~/.chauffeur/php/8.1/

# Global default PHP 8.2
$ composer global require laravel/installer
→ Uses ~/.chauffeur/php/8.2/ as configured globally`}/>

          <h3 className="text-xl font-bold text-white mb-2 mt-4">Cache and Configuration</h3>
          <ul className="list-disc list-inside text-slate-400 space-y-2">
            <li><strong>Workspace cache:</strong> Composer cache isolated under <code>~/.chauffeur/composer/</code></li>
            <li><strong>Global packages:</strong> Composer global packages managed with workspace PHP version</li>
            <li><strong>Authentication:</strong> Composer authentication tokens stored in workspace config</li>
            <li><strong>Performance:</strong> Intelligent caching across projects with same PHP version</li>
          </ul>

          <h3 className="text-xl font-bold text-white mb-2 mt-4">Shims and PATH Integration</h3>
          <CodeBlock code={`# PATH shim locations
~/.chauffeur/bin/composer      # Main Composer shim
~/.chauffeur/bin/php           # PHP version shim
~/.chauffeur/bin/php-8.3      # Version-specific shim
~/.chauffeur/bin/composer-8.3 # PHP 8.3 Composer shim`}/>

          <div className="mt-6 p-4 bg-emerald-500/10 border border-emerald-500/20 rounded-lg">
            <h4 className="font-semibold text-emerald-300 mb-2">🎼 Composer Advantages</h4>
            <div className="text-sm text-emerald-100/80">
              <p>• <strong>No PHP conflicts:</strong> Each project uses correct PHP version automatically</p>
              <p>• <strong>Dependency isolation:</strong> Composer cache isolated from system PHP</p>
              <p>• <strong>Seamless workflow:</strong> Standard Composer commands work as expected</p>
              <p>• <strong>Global packages:</strong> Composer global available with workspace isolation</p>
            </div>
          </div>
        </section>

        <section>
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="service-management">
            Service Management & Lifecycle
            <Link href="#service-management" onClick={(e) => scrollToId(e, 'service-management')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">
            Chauffeur manages services with systemd-style naming and comprehensive lifecycle operations.
          </p>

          <h3 className="text-xl font-bold text-white mb-2">Service Naming Convention</h3>
          <CodeBlock code={`# Global services
chauf-nginx              # Main web server
chauf-dnsmasq           # DNS resolution (if enabled)

# Project-specific services
chauf-php-fpm-myapp     # PHP-FPM for myapp project
chauf-php-fpm-blog     # PHP-FPM for blog project`}/>

          <h3 className="text-xl font-bold text-white mb-2 mt-4">Service Operations</h3>
          <ul className="list-disc list-inside text-slate-400 space-y-2">
            <li><strong>Start services:</strong> <code>chauf start</code> launches required services for all projects</li>
            <li><strong>Stop services:</strong> <code>chauf stop</code> gracefully shuts down services</li>
            <li><strong>Restart services:</strong> <code>chauf restart</code> stops and starts services</li>
            <li><strong>Service status:</strong> <code>chauf status</code> shows running state and configuration</li>
            <li><strong>Log access:</strong> <code>chauf logs</code> provides access to service logs</li>
          </ul>

          <h3 className="text-xl font-bold text-white mb-2 mt-4">System Integration</h3>
          <ul className="list-disc list-inside text-slate-400 space-y-2">
            <li><strong>Optional systemd:</strong> Services can be registered with systemd for auto-start</li>
            <li><strong>Port forwarding:</strong> iptables rules for privileged port access (80/443)</li>
            <li><strong>DNS integration:</strong> Optional dnsmasq configuration for .test domains</li>
            <li><strong>SSL certificates:</strong> Integration with system trust stores for mkcert</li>
          </ul>

          <div className="mt-6 p-4 bg-purple-500/10 border border-purple-500/20 rounded-lg">
            <h4 className="font-semibold text-purple-300 mb-2">⚙️ Service Design Philosophy</h4>
            <div className="text-sm text-purple-100/80">
              <p>• <strong>Non-intrusive:</strong> Services don't conflict with system packages</p>
              <p>• <strong>Self-contained:</strong> All dependencies managed within workspace</p>
              <p>• <strong>Portable:</strong> Workspace can be moved between systems</p>
              <p>• <strong>Observable:</strong> Comprehensive logging and status monitoring</p>
            </div>
          </div>
        </section>

        {/* Page Footer Navigation */}
        <div className="mt-16 pt-8 border-t border-slate-800 flex justify-between">
           <div className="text-left">
             <div className="text-xs text-slate-500 mb-1">Previous</div>
             <Link href="/docs/getting-started/first-project" className="text-primary hover:underline">First Project</Link>
           </div>
           <div className="text-right">
             <div className="text-xs text-slate-500 mb-1">Next</div>
             <Link href="/docs/reference/commands" className="text-primary hover:underline">Command Reference</Link>
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