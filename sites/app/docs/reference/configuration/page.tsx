"use client";

// React & Next.js
import React, { useEffect } from 'react';
import Link from 'next/link';

// Third-party libraries
import {
  ChevronRight,
  Settings,
  FileText,
  Server,
  Globe,
  Shield,
  Cpu,
  Zap,
  AlertTriangle,
  CheckCircle2,
  Copy,
  Terminal,
  FolderOpen
} from 'lucide-react';

// Page-specific components
import { TableOfContents } from '@/app/docs/_components/TableOfContents';
import { CodeBlock } from '@/app/docs/_components/CodeBlock';

export default function ConfigurationPage() {
  const currentSlug = 'reference/configuration';

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
        <Link href="/docs/reference" className="hover:text-primary transition-colors">Reference</Link>
        <ChevronRight size={14} />
        <span className="text-slate-200 capitalize">Configuration</span>
      </div>

      {/* Content */}
      <div className="prose prose-invert prose-slate max-w-none space-y-8 animate-fade-in">
        <div>
          <h1 className="text-4xl font-bold text-white mb-4">Configuration</h1>
          <p className="text-lg text-slate-400 leading-relaxed">
            Complete guide to configuring Chauffeur for your development environment, from global settings to project-specific overrides.
          </p>
        </div>

        <div className="p-4 bg-blue-500/10 border border-blue-500/20 rounded-lg flex gap-3">
          <Settings className="text-blue-400 shrink-0" />
          <div className="text-sm text-blue-100/80">
            <strong>Configuration Philosophy:</strong> Chauffeur uses a hierarchical configuration system. Global settings provide defaults, while project configurations override them for specific needs.
          </div>
        </div>

        <section id="global-configuration">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="global-configuration">
            Global Configuration
            <Link href="#global-configuration" onClick={(e) => scrollToId(e, 'global-configuration')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">The main configuration file controls Chauffeur's global behavior:</p>

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Configuration File Location</h3>
          <CodeBlock code="~/.chauffeur/config/chauffeur.yaml" />

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Default Configuration</h3>
          <CodeBlock code="version: 1
telemetry: false
workspace_dir: ~/.chauffeur

# Web Server Settings
nginx:
  enable: true
  http_port: 8080
  https_port: 8443
  worker_processes: auto
  worker_connections: 1024

# PHP Settings
php:
  default: 8.3
  memory_limit: 256M
  max_execution_time: 300
  upload_max_filesize: 64M
  post_max_size: 64M

# Port Management
ports:
  start_range: 8080
  end_range: 8099
  conflict_resolution: prompt
  nginx_http_fallback: 8080
  nginx_https_fallback: 8443
  php_fpm_fallback: 9000

# Project Management
projects_dir: ~/.chauffeur/projects
auto_discovery: true
default_ssl: false

# SSL Settings
ssl:
  provider: mkcert
  certificate_duration: 365
  auto_renewal: true

# Logging
logging:
  level: info
  rotate: true
  max_files: 7
  max_size: 100MB

# Performance
performance:
  shared_fpm: true
  process_idle_timeout: 300
  memory_optimization: true" />

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Editing Global Configuration</h3>
          <CodeBlock code="# Open configuration in your editor
chauf config edit

# Or edit directly
nano ~/.chauffeur/config/chauffeur.yaml

# Validate configuration
chauf config validate

# Restart to apply changes
chauf restart" />
        </section>

        <section id="nginx-settings">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="nginx-settings">
            Nginx Settings
            <Link href="#nginx-settings" onClick={(e) => scrollToId(e, 'nginx-settings')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
            <div className="bg-surface p-4 rounded-lg border border-slate-800">
              <h4 className="font-semibold text-white mb-2 flex items-center gap-2">
                <Server className="text-emerald-400" />
                Basic Settings
              </h4>
              <ul className="text-sm text-slate-400 space-y-2">
                <li><code>enable</code>: Enable/disable Nginx server</li>
                <li><code>http_port</code>: Default HTTP port</li>
                <li><code>https_port</code>: Default HTTPS port</li>
                <li><code>worker_processes</code>: Number of worker processes</li>
              </ul>
            </div>

            <div className="bg-surface p-4 rounded-lg border border-slate-800">
              <h4 className="font-semibold text-white mb-2 flex items-center gap-2">
                <Zap className="text-blue-400" />
                Performance Settings
              </h4>
              <ul className="text-sm text-slate-400 space-y-2">
                <li><code>worker_connections</code>: Max connections per worker</li>
                <li><code>keepalive_timeout</code>: Keep-alive timeout</li>
                <li><code>gzip</code>: Enable gzip compression</li>
                <li><code>cache</code>: Static file caching settings</li>
              </ul>
            </div>
          </div>

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Custom Nginx Directives</h3>
          <CodeBlock code={`nginx:
  custom_directives:
    - "add_header X-Powered-By \"Chauffeur\""
    - "add_header X-Frame-Options DENY"
    - "client_max_body_size 100M"`} />
        </section>

        <section id="php-settings">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="php-settings">
            PHP Settings
            <Link href="#php-settings" onClick={(e) => scrollToId(e, 'php-settings')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>

          <div className="bg-surface p-4 rounded-lg border border-slate-800 mb-4">
            <h4 className="font-semibold text-white mb-2">PHP Runtime Configuration</h4>
            <CodeBlock code="php:
  default: 8.3                    # Default PHP version
  memory_limit: 256M             # Memory limit per process
  max_execution_time: 300        # Max execution time (seconds)
  upload_max_filesize: 64M       # Max upload file size
  post_max_size: 64M             # Max POST request size
  max_input_vars: 3000           # Max input variables
  expose_php: off                # Hide PHP version in headers

  # OPcache Settings
  opcache:
    enable: true
    memory_consumption: 128
    max_accelerated_files: 10000
    validate_timestamps: off
    revalidate_freq: 0

  # Error Reporting
  error_reporting: E_ALL & ~E_DEPRECATED
  display_errors: off
  display_startup_errors: off" />
          </div>

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Version-Specific Settings</h3>
          <CodeBlock code={`php:
  versions:
    "8.3":
      memory_limit: 512M
      opcache_memory_consumption: 256
    "8.2":
      memory_limit: 256M
      opcache_memory_consumption: 128
    "8.1":
      memory_limit: 128M
      opcache_memory_consumption: 64`} />
        </section>

        <section id="port-management">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="port-management">
            Port Management
            <Link href="#port-management" onClick={(e) => scrollToId(e, 'port-management')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">Configure how Chauffeur handles port conflicts:</p>

          <div className="bg-surface p-4 rounded-lg border border-slate-800">
            <CodeBlock code="ports:
  start_range: 8080              # Start of port range
  end_range: 8099                # End of port range
  conflict_resolution: prompt     # prompt|auto|fail

  # Fallback ports when defaults are occupied
  nginx_http_fallback: 8080
  nginx_https_fallback: 8443
  php_fpm_fallback: 9000

  # Port Assignment Strategy
  strategy: sequential             # sequential|random|least_used

  # Port Persistence
  persistent_assignments: true    # Remember port assignments
  assignment_timeout: 30           # Timeout for port discovery (seconds)" />
          </div>

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Port Resolution Strategies</h3>
          <ul className="list-disc list-inside text-slate-400 space-y-2">
            <li><strong>Sequential:</strong> Try ports in order (8080, 8081, 8082...)</li>
            <li><strong>Random:</strong> Choose random available ports within range</li>
            <li><strong>Least Used:</strong> Choose ports with the fewest current connections</li>
          </ul>
        </section>

        <section id="project-configuration">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="project-configuration">
            Project Configuration
            <Link href="#project-configuration" onClick={(e) => scrollToId(e, 'project-configuration')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">Each project has its own configuration file for project-specific settings:</p>

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Project Configuration File</h3>
          <CodeBlock code="~/.chauffeur/projects/[project-slug]/project.yaml" />

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Project Configuration Example</h3>
          <CodeBlock code={`version: 1
path: /home/user/projects/my-laravel-app
php: 8.3

# Domain Configuration
site:
  domain: my-app.test
  ssl: true
  document_root: public
  index_files: [index.php, index.html]

# Domain Aliases
domains:
  aliases:
    - domain: admin.my-app.test
      ssl: true
      document_root: public/admin
    - domain: api.my-app.test
      ssl: true
      document_root: public/api
    - domain: "*.my-app.test"
      ssl: false
      wildcard: true

# PHP-FPM Configuration
php_fpm:
  pool_name: my-app
  dedicated_pool: false
  max_children: 50
  start_servers: 5
  min_spare_servers: 5
  max_spare_servers: 35
  max_requests: 500

  # Custom PHP Settings
  php_values:
    memory_limit: 512M
    max_execution_time: 600
    upload_max_filesize: 128M

# Environment Variables
environment:
  APP_ENV: local
  APP_DEBUG: true
  DB_CONNECTION: mysql
  DB_HOST: 127.0.0.1
  DB_DATABASE: my_app_local
  DB_USERNAME: root
  DB_PASSWORD: ""

# Custom Nginx Configuration
nginx:
  custom_config: |
    # Custom headers
    add_header X-Custom-Header "MyApp";

    # Custom rewrite rules
    location /admin/ {
      try_files $uri $uri/ /admin/index.php?$query_string;
    }

    # Proxy specific paths
    location /proxy/ {
      proxy_pass http://localhost:3000/;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
    }

# Runtime Information
runtime:
  php_fpm_socket: ~/.chauffeur/projects/my-app/runtime/php-fpm/php-fpm.sock
  nginx_socket: ~/.chauffeur/projects/my-app/runtime/nginx/nginx.sock
created_at: 2025-01-15T10:30:00+07:00
updated_at: 2025-01-20T14:45:00+07:00`} />

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Managing Project Configuration</h3>
          <CodeBlock code="# View project configuration
chauf config show my-app

# Edit project configuration
chauf config edit --project my-app

# Validate project configuration
chauf config validate --project my-app

# Reset project configuration to defaults
chauf config reset --project my-app" />
        </section>

        <section id="environment-variables">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="environment-variables">
            Environment Variables
            <Link href="#environment-variables" onClick={(e) => scrollToId(e, 'environment-variables')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>
          <p className="text-slate-400 mb-4">Chauffeur supports multiple ways to manage environment variables:</p>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
            <div className="bg-surface p-4 rounded-lg border border-slate-800">
              <h4 className="font-semibold text-white mb-2">Project Configuration</h4>
              <CodeBlock code="environment:
  APP_ENV: local
  DB_HOST: localhost
  API_KEY: secret-key" />
            </div>

            <div className="bg-surface p-4 rounded-lg border border-slate-800">
              <h4 className="font-semibold text-white mb-2">.env Files</h4>
              <CodeBlock code="# .env.local (loaded automatically)
APP_ENV=local
DB_HOST=localhost
API_KEY=secret-key" />
            </div>
          </div>

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Environment Variable Precedence</h3>
          <ol className="list-decimal list-inside text-slate-400 space-y-2">
            <li>System environment variables</li>
            <li>Global Chauffeur configuration</li>
            <li>Project configuration file</li>
            <li><code>.env</code> files (in project directory)</li>
            <li><code>.env.local</code> files (git ignored)</li>
          </ol>

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Environment File Templates</h3>
          <CodeBlock code="# Create .env.example template
chauf env init

# Create .env.local from template
chauf env copy my-app

# Show environment variables
chauf env show my-app" />
        </section>

        <section id="logging-configuration">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="logging-configuration">
            Logging Configuration
            <Link href="#logging-configuration" onClick={(e) => scrollToId(e, 'logging-configuration')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>

          <div className="bg-surface p-4 rounded-lg border border-slate-800">
            <CodeBlock code="logging:
  level: info                      # debug|info|warn|error
  format: json                     # json|text
  rotate: true                     # Enable log rotation

  # File Rotation
  max_files: 7                    # Number of log files to keep
  max_size: 100MB                  # Max size per file

  # Log Channels
  channels:
    nginx:
      enabled: true
      file: nginx/error.log
      level: error
    php:
      enabled: true
      file: php/error.log
      level: error
    fpm:
      enabled: true
      file: fpm/error.log
      level: warning
    access:
      enabled: true
      file: access.log
      format: combined

  # Structured Logging
  structured:
    enabled: true
    fields: [timestamp, level, service, message, context]

  # Remote Logging (optional)
  remote:
    enabled: false
    endpoint: https://logs.example.com/api/logs
    api_key: your-api-key" />
          </div>

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Log File Locations</h3>
          <CodeBlock code="~/.chauffeur/logs/
├── nginx/
│   ├── error.log
│   ├── access.log
│   └── ssl-request.log
├── php/
│   ├── error.log
│   ├── fpm.log
│   └── slow.log
├── projects/
│   ├── my-app/
│   │   ├── error.log
│   │   └── access.log
│   └── api/
│       ├── error.log
│       └── access.log
└── chauffeur.log" />

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Viewing Logs</h3>
          <CodeBlock code="# View all logs
chauf logs

# View specific service logs
chauf logs nginx
chauf logs php
chauf logs fpm

# View project logs
chauf logs --project=my-app

# Follow logs (tail -f)
chauf logs --follow
chauf logs nginx --follow

# Filter by log level
chauf logs --level=error
chauf logs --level=warning" />
        </section>

        <section id="security-settings">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="security-settings">
            Security Settings
            <Link href="#security-settings" onClick={(e) => scrollToId(e, 'security-settings')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>

          <div className="bg-surface p-4 rounded-lg border border-slate-800">
            <CodeBlock code={`security:
  # SSL/TLS Settings
  ssl:
    protocols: [TLSv1.2, TLSv1.3]
    ciphers: HIGH:!aNULL:!MD5
    prefer_server_ciphers: true

  # Security Headers
  headers:
    x_frame_options: DENY
    x_content_type_options: nosniff
    x_xss_protection: "1; mode=block"
    strict_transport_security: "max-age=63072000; includeSubDomains"
    referrer_policy: strict-origin-when-cross-origin

  # Access Control
  access_control:
    allowed_hosts: ["*.test"]
    ip_whitelist: []
    ip_blacklist: []

  # File Security
  file_security:
    hidden_files: [".env", ".git", "*.log"]
    upload_restrictions:
      max_file_size: 100MB
      allowed_extensions: [".php", ".html", ".css", ".js", ".png", ".jpg", ".gif"]

  # PHP Security
  php_security:
    expose_php: off
    allow_url_fopen: off
    allow_url_include: off
    disable_functions: ["exec", "shell_exec", "passthru", "system"]
    open_basedir: "\${PROJECT_ROOT}:\${UPLOADS_DIR}"`} />
          </div>

          <h3 className="text-xl font-semibold text-slate-200 mb-3">Project-Specific Security</h3>
          <CodeBlock code={`security:
  # Override global security settings
  override_globals: false

  # Project-specific restrictions
  restrictions:
    ip_whitelist: ["192.168.1.0/24"]
    require_auth: true
    auth_users: ["admin"]

  # SSL Requirements
  ssl_required: true
  redirect_to_https: true`} />
        </section>

        <section id="configuration-commands">
          <h2 className="text-2xl font-bold text-white mb-4 group flex items-center gap-2" id="configuration-commands">
            Configuration Commands
            <Link href="#configuration-commands" onClick={(e) => scrollToId(e, 'configuration-commands')} className="opacity-0 group-hover:opacity-100 text-slate-500 hover:text-primary cursor-pointer">#</Link>
          </h2>

          <div className="space-y-4">
            <div className="bg-surface p-4 rounded-lg border border-slate-800">
              <h4 className="font-semibold text-white mb-2">Global Configuration</h4>
              <CodeBlock code="# Show global configuration
chauf config show

# Edit global configuration
chauf config edit

# Validate configuration
chauf config validate

# Reset to defaults
chauf config reset

# Export configuration
chauf config export > chauffeur-backup.yaml

# Import configuration
chauf config import chauffeur-backup.yaml" />
            </div>

            <div className="bg-surface p-4 rounded-lg border border-slate-800">
              <h4 className="font-semibold text-white mb-2">Project Configuration</h4>
              <CodeBlock code="# Show project configuration
chauf config show --project=my-app

# Edit project configuration
chauf config edit --project=my-app

# Validate project configuration
chauf config validate --project=my-app

# Reset project to defaults
chauf config reset --project=my-app

# Copy configuration to another project
chauf config copy my-app --to=new-project" />
            </div>

            <div className="bg-surface p-4 rounded-lg border border-slate-800">
              <h4 className="font-semibold text-white mb-2">Environment Management</h4>
              <CodeBlock code="# Create environment template
chauf env init

# Copy environment file
chauf env copy my-app

# Show environment variables
chauf env show --project=my-app

# Set environment variable
chauf env set --project=my-app APP_KEY=secret-key

# Remove environment variable
chauf env unset --project=my-app APP_KEY" />
            </div>
          </div>
        </section>

        {/* Page Footer Navigation */}
        <div className="mt-16 pt-8 border-t border-slate-800 flex justify-between">
          <div className="text-left">
            <div className="text-xs text-slate-500 mb-1">Previous</div>
            <Link href="/docs/core/ssl-domains" className="text-primary hover:underline">SSL & Domains</Link>
          </div>
          <div className="text-right">
            <div className="text-xs text-slate-500 mb-1">Next</div>
            <Link href="/docs/reference/troubleshooting" className="text-primary hover:underline">Troubleshooting</Link>
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