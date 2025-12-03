import { Server, Shield, Zap, Globe, Layers, Terminal, Cpu, Settings, Link, Link2, FolderOpen, FileText, AlertTriangle, Wrench, Download, Activity, Heart, Info } from 'lucide-react';
import { TerminalLine, Feature, CommandExample } from './types';

// CLI Commands Reference - Source of truth for all command documentation
export interface CommandDefinition {
  command: string;
  category: 'workspace' | 'projects' | 'php' | 'services' | 'system' | 'utilities';
  description: string;
  usage: string;
  keyFlags: CommandFlag[];
  examples: CommandExample[];
  notes?: string[];
}

export interface CommandFlag {
  flag: string;
  description: string;
  required?: boolean;
  default?: string;
}

// Complete CLI Commands Reference - Accurate as of current CLI implementation
// Organized by category with most essential commands first
export const CLI_COMMANDS: CommandDefinition[] = [
  // WORKSPACE - Most fundamental command first
  {
    command: 'chauf init',
    category: 'workspace',
    description: 'Initializes the Chauffeur workspace with default configuration and directory structure. Creates ~/.chauffeur/ with necessary subdirectories including config/, projects/, logs/, cache/, and service-specific directories.',
    usage: 'chauf init [--force] [--quiet]',
    keyFlags: [
      { flag: '--force', description: 'Overwrite existing configuration files', required: false },
      { flag: '--quiet, -q', description: 'Suppress verbose output', required: false }
    ],
    examples: [
      { name: 'Initialize workspace', description: 'Create Chauffeur workspace', command: 'chauf init', output: '✓ Creating workspace at ~/.chauffeur' }
    ],
    notes: ['Safe to run multiple times', 'Creates ~/.chauffeur directory structure', 'Automatically creates example project at ~/.chauffeur/projects/example-project']
  },

  // SERVICES - Service management commands
  {
    command: 'chauf install',
    category: 'services',
    description: 'Install Chauffeur-managed services with visual separators; supports multiple PHP versions in a single command.',
    usage: 'chauf install <service> [version...] [--force] [--local] [--no-cache]',
    keyFlags: [
      { flag: '--force', description: 'Reinstall even if already present', required: false, default: 'false' },
      { flag: '--local', description: 'Use local PHP tarball (PHP only)', required: false, default: 'false' },
      { flag: '--no-cache', description: 'Skip download caching', required: false, default: 'false' }
    ],
    examples: [
      { name: 'Install nginx', description: 'Install nginx web server', command: 'chauf install nginx', output: '✓ nginx installed successfully' },
      { name: 'Install composer', description: 'Install Composer dependency manager', command: 'chauf install composer', output: '✓ composer installed successfully' },
      { name: 'Install PHP version', description: 'Install specific PHP version', command: 'chauf install php 8.2', output: '✓ PHP 8.2 installed successfully' },
      { name: 'Install multiple PHP versions', description: 'Install multiple PHP versions with visual separators', command: 'chauf install php 8.3 php 7.4', output: '✓ PHP 8.3 installed successfully\n─────────────────────────────────────────────────\n✓ PHP 7.4 installed successfully' },
      { name: 'Install all services', description: 'Install nginx, multiple PHP versions, and composer', command: 'chauf install nginx php 8.3 php 7.4 composer', output: '✓ nginx installed successfully\n─────────────────────────────────────────────────\n✓ PHP 8.3 installed successfully\n─────────────────────────────────────────────────\n✓ PHP 7.4 installed successfully\n─────────────────────────────────────────────────\n✓ composer installed successfully' }
    ],
    notes: [
      'Service parameter is required. Available services: nginx, php, composer',
      'Supports multiple PHP versions in one command: chauf install php 8.3 php 7.4',
      'Visual separators clearly distinguish between different service installations',
      'Creates example project during chauf init, automatically links when nginx and php are installed',
      'Example project provides immediate testing environment with welcome page and phpinfo()',
      'Project location: ~/.chauffeur/projects/example-project'
    ]
  },
  {
    command: 'chauf start',
    category: 'services',
    description: 'Start Chauffeur services with port forwarding and dnsmasq validation.',
    usage: 'chauf start [service...] [--project <slug>] [--all] [--dry-run]',
    keyFlags: [
      { flag: '--project <slug>', description: 'Start services for specific project', required: false },
      { flag: '--all', description: 'Start all services (default behavior)', required: false, default: 'true' },
      { flag: '--dry-run', description: 'Show what would be started without executing', required: false }
    ],
    examples: [
      { name: 'Start all services', description: 'Start all Chauffeur services', command: 'chauf start', output: '✓ Started nginx (8080), php-fpm 8.3 (9000)' },
      { name: 'Start nginx only', description: 'Start only nginx service', command: 'chauf start nginx', output: '✓ nginx started successfully' },
      { name: 'Start project services', description: 'Start services for specific project', command: 'chauf start --project my-project', output: '✓ Started nginx, php-fpm for my-project' }
    ],
    notes: ['Valid services: nginx, php-fpm, or project slugs', 'Automatically configures port forwarding for privileged ports', 'Validates dnsmasq configuration for .test domains']
  },
  {
    command: 'chauf stop',
    category: 'services',
    description: 'Stop Chauffeur services and clean port-forward rules.',
    usage: 'chauf stop [service...] [--project <slug>] [--all] [--dry-run]',
    keyFlags: [
      { flag: '--project <slug>', description: 'Stop services for specific project', required: false },
      { flag: '--all', description: 'Stop all services (default behavior)', required: false, default: 'true' },
      { flag: '--dry-run', description: 'Show what would be stopped without executing', required: false }
    ],
    examples: [
      { name: 'Stop all services', description: 'Stop all Chauffeur services', command: 'chauf stop', output: '✓ Stopped nginx, php-fpm, port forwarding' }
    ],
    notes: ['Valid services: nginx, php-fpm, or project slugs', 'Cleans up port forwarding rules when nginx is stopped']
  },
  {
    command: 'chauf restart',
    category: 'services',
    description: 'Restart Chauffeur services (equivalent to stop then start, preserves configuration).',
    usage: 'chauf restart [service...] [--project <slug>] [--all] [--dry-run]',
    keyFlags: [
      { flag: '--project <slug>', description: 'Restart services for specific project', required: false },
      { flag: '--all', description: 'Restart all services (default behavior)', required: false, default: 'true' },
      { flag: '--dry-run', description: 'Show what would be restarted without executing', required: false }
    ],
    examples: [
      { name: 'Restart all services', description: 'Restart all services with preserved config', command: 'chauf restart', output: '✓ Restarted nginx (8080), php-fpm pools' }
    ],
    notes: ['Service-specific, project-specific, and all-service restart capabilities available']
  },
  {
    command: 'chauf status',
    category: 'services',
    description: 'Shows the status of Chauffeur services with chauf- prefix.',
    usage: 'chauf status [service-type] [--project <slug>] [--detail]',
    keyFlags: [
      { flag: '--project <slug>', description: 'Show status for specific project only', required: false },
      { flag: '--detail, -v', description: 'Show detailed status information', required: false }
    ],
    examples: [
      { name: 'Show all services', description: 'Display service status', command: 'chauf status', output: '[ status ] Checking status of 3 Chauffeur services...\n\n[ status ] Chauffeur Services (3 total)\n[ status ]      SERVICE             TYPE      PROJECT     UPTIME    MEMORY    \n[ status ] ───────────────────────────────────────────────────────────────────\n[ status ] ○    chauf-nginx         Global    —                               \n[ status ] ○    chauf-php-fpm-8.3   Global    8.3                             \n[ status ] ○    chauf-php-fpm-7.4   Global    7.4                             \n[ status ] \n[ status ] Legend: ● Running  ● Warning  ● Error  ○ Stopped\n[ status ] ✓ Status check complete (56ms)' }
    ],
    notes: ['Shows global services and project-specific services', 'Display includes process information, uptime, and resource usage']
  },

  // PROJECTS - Project management commands
  {
    command: 'chauf link',
    category: 'projects',
    description: 'Register current directory, detect template, generate configs with multi-domain support.',
    usage: 'chauf link [--site <domain>] [--secure] [--php <version>] [--http-port <port>] [--https-port <port>] [--alias <domain>] [--add-alias <domain>] [--dedicated-fpm] [--force]',
    keyFlags: [
      { flag: '--site <domain>', description: 'Site name for domain', required: false },
      { flag: '--secure', description: 'Enable HTTPS with SSL certificate', required: false },
      { flag: '--php <version>', description: 'PHP version for this project', required: false },
      { flag: '--http-port <port>', description: 'Custom HTTP port', required: false },
      { flag: '--https-port <port>', description: 'Custom HTTPS port', required: false },
      { flag: '--alias <domain>', description: 'Add domain alias', required: false },
      { flag: '--add-alias <domain>', description: 'Additional domain alias', required: false },
      { flag: '--dedicated-fpm', description: 'Use dedicated PHP-FPM pool', required: false },
      { flag: '--force', description: 'Overwrite existing configuration', required: false }
    ],
    examples: [
      { name: 'Link current directory', description: 'Link current directory as project', command: 'chauf link', output: '✓ Linked to http://my-app.test [PHP 8.3]' },
      { name: 'Link with SSL', description: 'Link project with HTTPS', command: 'chauf link --secure', output: '✓ Linked to https://my-app.test [PHP 8.3]' },
      { name: 'Link with custom PHP', description: 'Link with specific PHP version', command: 'chauf link --php 8.1', output: '✓ Linked to http://my-app.test [PHP 8.1 Dedicated]' },
      { name: 'Link with aliases', description: 'Link project with additional domains', command: 'chauf link --alias api.test --alias admin.test', output: '✓ Linked to http://my-app.test [aliases: api.test, admin.test]' }
    ],
    notes: ['Supports Laravel, WordPress, and general project detection', 'Creates nginx configuration and restarts services', 'Multi-domain SSL certificates supported']
  },
  {
    command: 'chauf category',
    category: 'projects',
    description: 'List all projects grouped by category with alphabetical sorting and "Uncategorized" at bottom.',
    usage: 'chauf category [--sort <method>] [--show-empty]',
    keyFlags: [
      { flag: '--sort <method>', description: 'Sorting method: alphabetical (default) or count', required: false },
      { flag: '--show-empty', description: 'Show empty categories', required: false }
    ],
    examples: [
      { name: 'List projects by category', description: 'Show projects grouped by category, sorted alphabetically with "Uncategorized" at bottom', command: 'chauf category', output: '[ category ] Projects by Category (3 categories)\n[ category ] \n[ category ] E-commerce (2)\n[ category ] └── shop.test\n[ category ] └── store.test\n[ category ] \n[ category ] Laravel (3)\n[ category ] └── blog.test\n[ category ] └── admin.test\n[ category ] └── api.test\n[ category ] \n[ category ] Uncategorized (1)\n[ category ] └── misc-project.test' },
      { name: 'Sort by project count', description: 'Show categories sorted by number of projects', command: 'chauf category --sort count', output: '[ category ] Projects by Category (3 categories)\n[ category ] \n[ category ] Laravel (3)\n[ category ] └── blog.test\n[ category ] └── admin.test\n[ category ] └── api.test\n[ category ] \n[ category ] E-commerce (2)\n[ category ] └── shop.test\n[ category ] └── store.test\n[ category ] \n[ category ] Uncategorized (1)\n[ category ] └── misc-project.test' },
      { name: 'Show empty categories', description: 'Include empty categories in output', command: 'chauf category --show-empty', output: '[ category ] Projects by Category (4 categories)\n[ category ] \n[ category ] E-commerce (2)\n[ category ] └── shop.test\n[ category ] └── store.test\n[ category ] \n[ category ] Laravel (3)\n[ category ] └── blog.test\n[ category ] └── admin.test\n[ category ] └── api.test\n[ category ] \n[ category ] WordPress (0)\n[ category ] \n[ category ] Uncategorized (1)\n[ category ] └── misc-project.test' }
    ],
    notes: [
      'Categories are sorted alphabetically by default with "Uncategorized" always at the bottom',
      'Use --sort count to order categories by number of projects (descending)',
      'Projects within each category are sorted alphabetically',
      'Empty categories are hidden by default unless --show-empty is used',
      'Uncategorized projects are automatically grouped at the bottom for better organization'
    ]
  },
  {
    command: 'chauf links',
    category: 'projects',
    description: 'List all registered projects in a formatted table with detailed view support.',
    usage: 'chauf links [--slug <slug>] [--site <domain>]',
    keyFlags: [
      { flag: '--slug <slug>', description: 'Display detailed information for a specific project by slug', required: false },
      { flag: '--site <domain>', description: 'Display detailed information for a specific project by site domain', required: false }
    ],
    examples: [
      { name: 'List projects', description: 'Show all linked projects sorted alphabetically by category', command: 'chauf links', output: '[ links ] Linked Projects (6)\n[ links ] SLUG             PATH                                             DOMAIN                ALIASES   SSL  PHP   CREATED\n[ links ] ---------------  -----------------------------------------------  --------------------  --------  ---  ----  -------------------\n[ links ] example-project 📁  /home/user/.chauffeur/projects/example-project  example-project.test  0              8.3   2025-11-23 03:02' },
      { name: 'View project details', description: 'Show detailed project information', command: 'chauf links --slug my-project', output: '[ links ] Project Details: my-project' }
    ],
    notes: ['--slug and --site flags are mutually exclusive', 'Shows SSL indicators (*) for secured sites', 'Displays primary domain and all aliases']
  },
  {
    command: 'chauf unlink',
    category: 'projects',
    description: 'Remove registrations or specific aliases. Defaults to current directory.',
    usage: 'chauf unlink [--slug <slug>] [--site <domain>] [--project <path>] [--alias <domain>] [--all] [--force]',
    keyFlags: [
      { flag: '--slug <slug>', description: 'Project slug to unlink', required: false },
      { flag: '--site <site>', description: 'Site name to unlink', required: false },
      { flag: '--project <path>', description: 'Project path to unlink', required: false },
      { flag: '--alias <domain>', description: 'Remove specific alias', required: false },
      { flag: '--all', description: 'Unlink all projects', required: false },
      { flag: '--force', description: 'Force unlink without confirmation', required: false }
    ],
    examples: [
      { name: 'Unlink current project', description: 'Unlink current directory', command: 'chauf unlink', output: '✓ Unlinked my-app.test' },
      { name: 'Unlink specific alias', description: 'Remove only a specific domain alias', command: 'chauf unlink --alias api.test', output: '✓ Removed alias api.test from my-project' }
    ],
    notes: ['Removes nginx configurations and restarts nginx', 'Shows confirmation with all domains and SSL status']
  },
  {
    command: 'chauf secure',
    category: 'projects',
    description: 'Add SSL certificate to current linked project',
    usage: 'chauf secure',
    keyFlags: [],
    examples: [
      {
        name: 'Add SSL to current project',
        description: 'Generate and install SSL certificate for the current project',
        command: 'chauf secure',
        output: `[ secure ] ✓ Multi-domain SSL certificate generated successfully with mkcert
[ secure ] ✓ SSL certificate added successfully
[ secure ]   Secure access: https://my-project.test`
      }
    ],
    notes: [
      'Must be run from a linked project directory',
      'Uses mkcert for trusted certificates if available',
      'Falls back to self-signed certificates if mkcert not available',
      'Automatically reloads nginx configuration',
      'Supports multi-domain certificates for all aliases'
    ]
  },
  {
    command: 'chauf unsecure',
    category: 'projects',
    description: 'Remove SSL certificate from current linked project',
    usage: 'chauf unsecure',
    keyFlags: [],
    examples: [
      {
        name: 'Remove SSL from current project',
        description: 'Remove SSL certificate from the current project',
        command: 'chauf unsecure',
        output: `[ unsecure ] ✓ SSL certificate removed successfully
[ unsecure ]   HTTP access: http://my-project.test`
      }
    ],
    notes: [
      'Must be run from a linked project directory',
      'Removes SSL certificate files and configuration',
      'Automatically reloads nginx configuration'
    ]
  },

  // PHP - PHP management commands
  {
    command: 'chauf php install',
    category: 'php',
    description: 'Build/install PHP runtime under workspace.',
    usage: 'chauf php install <version> [--force] [--no-ext] [--from <source>]',
    keyFlags: [
      { flag: '--force', description: 'Reinstall if already exists', required: false },
      { flag: '--no-ext', description: 'Skip extension compilation', required: false },
      { flag: '--from <source>', description: 'Build from specific source', required: false }
    ],
    examples: [
      { name: 'Install PHP 8.3', description: 'Install PHP 8.3 runtime', command: 'chauf php install 8.3', output: '✓ PHP 8.3 installed (extensions: 23)' }
    ],
    notes: ['First installation may take 10-15 minutes', 'Requires build-essential toolchain', 'Supports GD extension prompting for legacy PHP versions']
  },
  {
    command: 'chauf php use',
    category: 'php',
    description: 'Set global default PHP.',
    usage: 'chauf php use <version>',
    keyFlags: [],
    examples: [
      { name: 'Set global PHP', description: 'Set default PHP version', command: 'chauf php use 8.3', output: '✓ PHP 8.3 set as global default' }
    ]
  },
  {
    command: 'chauf php isolate',
    category: 'php',
    description: 'Pin current linked project to a PHP version.',
    usage: 'chauf php isolate <version>',
    keyFlags: [],
    examples: [
      { name: 'Isolate project PHP', description: 'Pin project to PHP version', command: 'chauf php isolate 8.1', output: '✓ Project isolated to PHP 8.1' }
    ]
  },
  {
    command: 'chauf php current',
    category: 'php',
    description: 'Show current PHP version for directory or global default',
    usage: 'chauf php current',
    keyFlags: [],
    examples: [
      {
        name: 'Show current PHP version',
        description: 'Display PHP version for current directory and global default',
        command: 'chauf php current',
        output: `[ php current ] Project: /home/user/my-project
[ php current ] Project PHP: 8.3 (isolated)
[ php current ] Global PHP: 8.2 (default)
[ php current ] PHP binary: /home/user/.chauffeur/php/8.3/bin/php`
      },
      {
        name: 'Show current PHP outside project',
        description: 'Display global PHP when not in a project directory',
        command: 'chauf php current',
        output: `[ php current ] No project detected in current directory
[ php current ] Global PHP: 8.2 (default)
[ php current ] PHP binary: /home/user/.chauffeur/php/8.2/bin/php`
      }
    ],
    notes: [
      'Shows project-specific PHP if directory is linked to Chauffeur',
      'Shows global PHP default when not in a project directory',
      'Displays PHP binary path for reference'
    ]
  },

  // SERVICE EXECUTION - Direct service commands
  {
    command: 'chauf nginx',
    category: 'services',
    description: 'Run the managed nginx binary with passthrough arguments.',
    usage: 'chauf nginx [args...]',
    keyFlags: [],
    examples: [
      { name: 'Check nginx version', description: 'Show nginx version', command: 'chauf nginx -v', output: 'nginx version: nginx/1.24.0' },
      { name: 'Test nginx configuration', description: 'Validate nginx configuration', command: 'chauf nginx -t', output: 'nginx: configuration file test is successful' },
      { name: 'Reload nginx configuration', description: 'Reload nginx without restart', command: 'chauf nginx -s reload', output: 'nginx: configuration reloaded' }
    ],
    notes: ['Only available after `chauf install nginx`', 'All nginx arguments are passed through directly']
  },
  {
    command: 'chauf php',
    category: 'php',
    description: 'Run the managed PHP binary with project-aware version isolation.',
    usage: 'chauf php [args...]',
    keyFlags: [],
    examples: [
      { name: 'Check PHP version', description: 'Show current PHP version', command: 'chauf php --version', output: 'PHP 8.3.10 (cli)' },
      { name: 'Run PHP script', description: 'Execute a PHP file', command: 'chauf php script.php', output: 'Script output here' },
      { name: 'Interactive PHP shell', description: 'Start interactive PHP shell', command: 'chauf php -a', output: 'Interactive shell' }
    ],
    notes: ['Automatically uses project-specific PHP if project is isolated', 'Falls back to global default PHP version', 'Only available after installing at least one PHP version']
  },
  {
    command: 'chauf composer',
    category: 'php',
    description: 'Run the managed Composer binary with Chauffeur PHP integration.',
    usage: 'chauf composer [args...]',
    keyFlags: [],
    examples: [
      { name: 'Check Composer version', description: 'Show Composer version', command: 'chauf composer --version', output: 'Composer version 2.7.1' },
      { name: 'Install dependencies', description: 'Install project dependencies', command: 'chauf composer install', output: 'Installing dependencies from lock file' }
    ],
    notes: ['Only available after `chauf install composer`', 'Automatically uses appropriate PHP version for projects', 'Integrates with Chauffeur PHP runtimes']
  },

  // SYSTEM - System commands
  {
    command: 'chauf self-update',
    category: 'system',
    description: 'Update binary from git or rebuild from current repo with dynamic version checking.',
    usage: 'chauf self-update [--dev]',
    keyFlags: [
      { flag: '--dev', description: 'Rebuild from current git repository', required: false }
    ],
    examples: [
      { name: 'Check for updates', description: 'Check if newer version is available', command: 'chauf self-update', output: 'You are already using the latest version (chauf 1.3.6)' },
      { name: 'Update to latest', description: 'Update when newer version available', command: 'chauf self-update', output: 'Update available: chauffeur 1.3.5 → 1.3.6' },
      { name: 'Development rebuild', description: 'Rebuild from current source', command: 'chauf self-update --dev', output: '✓ Dev rebuild complete (commit 90b3218)' }
    ],
    notes: [
      'Automatically checks for newer versions from GitHub before updating',
      'Shows version comparison when updates are available',
      'Development builds use branch-commit format (e.g., develop-90b3218)',
      'Production builds use semantic versioning (e.g., 1.3.6)'
    ]
  },
  {
    command: 'chauf uninstall',
    category: 'system',
    description: 'Remove workspace (and runtimes with --purge).',
    usage: 'chauf uninstall [--purge]',
    keyFlags: [
      { flag: '--purge', description: 'Remove all runtimes as well', required: false }
    ],
    examples: [
      { name: 'Uninstall Chauffeur', description: 'Remove Chauffeur workspace', command: 'chauf uninstall', output: '✓ Workspace removed' },
      { name: 'Complete uninstall', description: 'Remove workspace and all runtimes', command: 'chauf uninstall --purge', output: '✓ Workspace and runtimes removed' }
    ]
  },
  {
    command: 'chauf remove',
    category: 'system',
    description: 'Remove installed runtimes (php, nginx, composer).',
    usage: 'chauf remove <service> [version] [--force]',
    keyFlags: [
      { flag: '--force', description: 'Force removal without confirmation', required: false }
    ],
    examples: [
      { name: 'Remove PHP version', description: 'Remove specific PHP version', command: 'chauf remove php 8.1', output: '✓ Removed PHP 8.1' },
      { name: 'Remove nginx', description: 'Remove nginx installation', command: 'chauf remove nginx', output: '✓ Removed nginx' },
      { name: 'Remove composer', description: 'Remove composer installation', command: 'chauf remove composer', output: '✓ Removed composer' }
    ]
  },

  // UTILITIES - Enhanced utility commands
  {
    command: 'chauf info',
    category: 'utilities',
    description: 'Show workspace paths, installed services, versions, port configuration, and GitHub release status.',
    usage: 'chauf info',
    keyFlags: [
      { flag: '--help, -h', description: 'Show help message', required: false }
    ],
    examples: [
      {
        name: 'Show workspace information',
        description: 'Display comprehensive workspace and system information',
        command: 'chauf info',
        output: `[ info ] Environment
[ info ] Workspace: /home/user/.chauffeur
[ info ] Binary: /home/user/.chauffeur/bin/chauf
[ info ] Projects dir: /home/user/.chauffeur/projects
[ info ] Config file: /home/user/.chauffeur/config/chauffeur.yaml

[ info ] Versions
[ info ] Current CLI: 0.1.0
[ info ] Build timestamp: 2025-11-23T12:09:24Z
[ info ] Build commit: 4a20ddc
[ info ] Latest release: v1.3.4 (update available)
  ⚠ Warning: Remote branch has newer commits
  └── develop is 24 commit(s) ahead (latest 8323960)

[ info ] Managed Services
[ info ] Nginx: nginx version: nginx/1.29.3 (/home/user/.chauffeur/nginx/sbin/nginx)
[ info ] Composer: Composer version 2.9.2 2025-11-19 21:57:25 (/home/user/.chauffeur/bin/composer)
[ info ] PHP: 7.4, 8.3 (default)

[ info ] Port Configuration
[ info ] Nginx HTTP: 8080
[ info ] Nginx HTTPS: 8443
[ info ] Port range: 8080-8099 (PROMPT)
[ info ] PHP-FPM fallback: 9000`
      }
    ],
    notes: [
      'Shows GitHub release comparison with local build',
      'Reports commit drift (ahead/behind status)',
      'Displays workspace structure and paths',
      'Shows all installed service versions'
    ]
  },
  {
    command: 'chauf doctor',
    category: 'utilities',
    description: 'Performs comprehensive health checks on your Chauffeur installation, validates system dependencies, and provides guidance for resolving issues.',
    usage: 'chauf doctor [options]',
    keyFlags: [
      { flag: '--check-all, -a', description: 'Run all dependency checks (default behavior)', required: false },
      { flag: '--check-deps, -d', description: 'Check system dependencies (git, curl, tar, etc.)', required: false },
      { flag: '--check-php, -p', description: 'Check PHP build dependencies and headers', required: false },
      { flag: '--check-ssl, -s', description: 'Check SSL certificate dependencies', required: false },
      { flag: '--check-network, -n', description: 'Check network and port availability', required: false },
      { flag: '--check-dns', description: 'Check DNS resolution for .test domains', required: false },
      { flag: '--verbose, -v', description: 'Show detailed diagnostic information', required: false },
      { flag: '--fix, -f', description: 'Show fix suggestions for issues found', required: false },
      { flag: '--auto-fix', description: 'Attempt to automatically fix issues where safe', required: false },
      { flag: '--quiet, -q', description: 'Suppress non-error output', required: false },
      { flag: '--help, -h', description: 'Show this help message', required: false }
    ],
    examples: [
      { name: 'Full health check', description: 'Run all health checks', command: 'chauf doctor', output: '[ doctor ] 🩺 Chauffeur Doctor\n[ doctor ] ✓ Overall Status (✅ All systems healthy)\n[ doctor ] ✓ Doctor completed (All checks passed - system is healthy!)' },
      { name: 'Show fix suggestions', description: 'Check system and show fix suggestions', command: 'chauf doctor --fix', output: '[ doctor ] ⚠️ mkcert: Local trusted certificate authority\n[ doctor ]   └─ Fix: sudo dnf install -y mkcert' },
      { name: 'Auto-fix with confirmation', description: 'Review fix plan before execution', command: 'chauf doctor --auto-fix', output: '[ doctor ] 🔧 Fix Plan\n[ doctor ] Found 1 fixable issue(s): 0 errors, 1 warnings\n[ doctor ] SSL Certificate Dependencies:\n[ doctor ]   ⚠️ mkcert: Local trusted certificate authority\n[ doctor ]     Command: sudo dnf install -y mkcert\n[ doctor ] Do you want to proceed with these fixes? [y/N]' },
      { name: 'Check specific areas', description: 'Check only SSL and network dependencies', command: 'chauf doctor --check-ssl --check-network', output: '[ doctor ] ✓ SSL Certificate Dependencies: openssl available\n[ doctor ] ✓ Network Dependencies: iptables available' },
      { name: 'Verbose diagnostics', description: 'Run with detailed output', command: 'chauf doctor --verbose', output: '[ doctor ] ✅ Installed (2.51.1) (git)\n[ doctor ]   └─ Version: 2.51.1' }
    ],
    notes: [
      'Cross-platform support (Ubuntu/Debian, Arch, Fedora)',
      'Validates build dependencies for PHP compilation',
      'Checks SSL certificate setup (mkcert, OpenSSL)',
      'Validates PHP OpenSSL configuration with distribution-aware CA paths',
      'Validates network configuration and port availability',
      'Tests DNS resolution for .test domains',
      'Provides distro-specific fix commands'
    ]
  },
  {
    command: 'chauf logs',
    category: 'utilities',
    description: 'View and follow logs from Chauffeur services.',
    usage: 'chauf logs [service-name] [options]',
    keyFlags: [
      { flag: '--follow, -f', description: 'Follow log output in real-time (like tail -f)', required: false },
      { flag: '--lines <n>, -n', description: 'Show last N lines (default: 100)', required: false, default: '100' },
      { flag: '--since <time>', description: 'Show logs since specified time', required: false },
      { flag: '--until <time>', description: 'Show logs until specified time', required: false },
      { flag: '--level <level>', description: 'Filter logs by level (error, warning, info, debug)', required: false },
      { flag: '--context, -c', description: 'Show file context and metadata', required: false },
      { flag: '--verbose, -v', description: 'Show verbose output with service prefixes', required: false },
      { flag: '--quiet, -q', description: 'Show only log lines without additional formatting', required: false }
    ],
    examples: [
      { name: 'View available services', description: 'List all services with logs available', command: 'chauf logs', output: '📋 Available Services for Log Viewing' },
      { name: 'View nginx logs', description: 'Show nginx access and error logs', command: 'chauf logs nginx', output: 'nginx: 2024-01-01 12:00:00 [info] Server started' },
      { name: 'Follow PHP-FPM logs', description: 'Follow PHP-FPM logs in real-time', command: 'chauf logs php-fpm --follow', output: 'Following PHP-FPM logs...' },
      { name: 'Filter by log level', description: 'Show only error logs', command: 'chauf logs nginx --level error', output: 'nginx: 2024-01-01 12:00:00 [error] Connection failed' }
    ],
    notes: [
      'Automatic service discovery for global and project services',
      'Supports nginx and PHP-FPM log locations',
      'Interactive selection when multiple services match',
      'Smart log file detection across workspace',
      'Supports both real-time following and historical viewing'
    ]
  },
  {
    command: 'chauf clean',
    category: 'utilities',
    description: 'Clean workspace files with file size display and accurate reporting.',
    usage: 'chauf clean [target] [--dry-run] [--force] [--older-than <days>] [--keep-versions <num>] [--what]',
    keyFlags: [
      { flag: '--dry-run', description: 'Show what would be cleaned without executing', required: false },
      { flag: '--force', description: 'Clean without confirmation', required: false },
      { flag: '--older-than <days>', description: 'Clean files older than specified days', required: false },
      { flag: '--keep-versions <num>', description: 'Keep N most recent versions', required: false },
      { flag: '--what', description: 'Show what would be cleaned', required: false }
    ],
    examples: [
      { name: 'Clean workspace', description: 'Clean old files', command: 'chauf clean --dry-run', output: 'Would clean 2.3GB of old files' },
      { name: 'Force clean', description: 'Clean without confirmation', command: 'chauf clean --force', output: '✓ Cleaned 2.3GB of files' }
    ]
  },
  {
    command: 'chauf migrate',
    category: 'utilities',
    description: 'Migrate a project to a different Chauffeur workspace with backup support.',
    usage: 'chauf migrate <project-slug> <destination-workspace> [--backup] [--no-backup] [--dry-run] [--force] [--verbose]',
    keyFlags: [
      { flag: '--backup', description: 'Create backup before migration (default: true)', required: false },
      { flag: '--no-backup', description: 'Skip backup creation', required: false },
      { flag: '--dry-run, -n', description: 'Show what would be done without actually doing it', required: false },
      { flag: '--force, -f', description: 'Skip confirmation prompts', required: false },
      { flag: '--verbose, -v', description: 'Show detailed output', required: false }
    ],
    examples: [
      { name: 'Migrate project', description: 'Migrate project to different workspace', command: 'chauf migrate my-project /home/user/other-workspace', output: '✓ Migrated my-project to /home/user/other-workspace' },
      { name: 'Dry run migration', description: 'Preview migration without making changes', command: 'chauf migrate blog-site /backup/workspace --dry-run', output: 'DRY RUN: Would migrate blog-site to /backup/workspace' }
    ],
    notes: ['Creates automatic backups unless --no-backup is specified', 'Use --dry-run to preview changes before migration']
  },
  {
    command: 'chauf hello-world',
    category: 'utilities',
    description: 'Prints a friendly greeting message.',
    usage: 'chauf hello-world',
    keyFlags: [
      { flag: '-h, --help', description: 'Show this help message', required: false }
    ],
    examples: [
      {
        name: 'Print greeting',
        description: 'Display friendly welcome message',
        command: 'chauf hello-world',
        output: `[ hello-world ] ✓ Hello, World! (Chauffeur greeting)
[ hello-world ] Welcome to Chauffeur - your Linux PHP development environment`
      }
    ],
    notes: ['Simple utility command for testing Chauffeur installation']
  }
];

// Group commands by category for easier reference
export const COMMANDS_BY_CATEGORY = {
  workspace: CLI_COMMANDS.filter(cmd => cmd.category === 'workspace'),
  projects: CLI_COMMANDS.filter(cmd => cmd.category === 'projects'),
  php: CLI_COMMANDS.filter(cmd => cmd.category === 'php'),
  services: CLI_COMMANDS.filter(cmd => cmd.category === 'services'),
  system: CLI_COMMANDS.filter(cmd => cmd.category === 'system'),
  utilities: CLI_COMMANDS.filter(cmd => cmd.category === 'utilities')
};

// Command categories with icons for documentation
export const COMMAND_CATEGORIES = [
  {
    id: 'workspace',
    title: 'Workspace Management',
    description: 'Initialize and manage Chauffeur workspace',
    icon: Settings
  },
  {
    id: 'services',
    title: 'Service Management',
    description: 'Install, start, stop, and manage Chauffeur services',
    icon: Server
  },
  {
    id: 'projects',
    title: 'Project Management',
    description: 'Link, unlink, and manage development projects',
    icon: FolderOpen
  },
  {
    id: 'php',
    title: 'PHP Management',
    description: 'Manage PHP versions, Composer, and extensions',
    icon: Cpu
  },
  {
    id: 'utilities',
    title: 'Utilities',
    description: 'Logs, cleanup, diagnostics, and helpful tools',
    icon: Wrench
  },
  {
    id: 'system',
    title: 'System',
    description: 'Installation updates and workspace management',
    icon: Terminal
  }
];

// Helper function to find command by name
export const findCommand = (commandName: string): CommandDefinition | undefined => {
  return CLI_COMMANDS.find(cmd => {
    // Handle command patterns like 'chauf php install'
    const cmdParts = commandName.split(' ');
    const thisCmdParts = cmd.command.split(' ');

    // Check if the command starts with the same pattern
    return thisCmdParts.every((part, index) => cmdParts[index] === part);
  });
};

// Get all available command names for autocomplete
export const COMMAND_NAMES = CLI_COMMANDS.map(cmd => cmd.command);

export const HERO_TERMINAL_LINES: TerminalLine[] = [
  { text: 'chauf init', type: 'command', delay: 0 },
  { text: '✓ Creating workspace at ~/.chauffeur', type: 'success', delay: 500 },
  { text: '✓ Configuration initialized', type: 'success', delay: 800 },
  { text: '', type: 'info', delay: 1000 },
  { text: 'chauf link my-app --secure --php 8.3', type: 'command', delay: 1200 },
  { text: '✓ Detected Laravel project', type: 'info', delay: 1800 },
  { text: '✓ Securing my-app.test with SSL', type: 'success', delay: 2200 },
  { text: '✓ Started PHP-FPM 8.3 (Isolated)', type: 'success', delay: 2600 },
  { text: '🎉 Ready! Visit https://my-app.test', type: 'output', delay: 3000 },
];

export const FEATURES: Feature[] = [
  {
    title: "Linux Native",
    description: "Built specifically for Linux distributions. No containers, no virtual machines, just pure native performance.",
    icon: Terminal
  },
  {
    title: "Zero Config",
    description: "Forget Nginx configurations. Chauffeur handles routing, FPM sockets, and DNS automatically.",
    icon: Zap
  },
  {
    title: "Easy SSL",
    description: "Generate locally trusted SSL certificates with a simple flag. Secure your sites with HTTPS instantly.",
    icon: Shield
  },
  {
    title: "Multi-Version PHP",
    description: "Run PHP 7.4 through 8.4 simultaneously. Project-level isolation means no version conflicts.",
    icon: Layers
  },
  {
    title: "Resource Efficient",
    description: "Smart process management sleeps idle workers. Uses 98% less RAM than Laravel Homestead.",
    icon: Cpu
  },
  {
    title: ".test Domains",
    description: "Automatic local DNS resolution. Just create a folder and visit foldername.test.",
    icon: Globe
  }
];

export const COMMAND_EXAMPLES: CommandExample[] = [
  {
    name: "Link Project",
    description: "Serve the current directory",
    command: "chauf link --secure",
    output: "Linking ./ to https://my-project.test [PHP 8.4]"
  },
  {
    name: "Project PHP Version",
    description: "Set project-specific PHP version",
    command: "chauf link --php=8.1",
    output: "Project pinned to PHP 8.1. FPM pool restarted."
  },
  {
    name: "Global PHP Version",
    description: "Set global default PHP version",
    command: "chauf php use 8.3",
    output: "PHP 8.3 set as global default"
  },
  {
    name: "List Projects",
    description: "Show all linked projects",
    command: "chauf links",
    output: "📋 Found 2 linked projects"
  },
  {
    name: "Health Check",
    description: "Run comprehensive health checks",
    command: "chauf doctor",
    output: "✅ All systems healthy"
  },
  {
    name: "Auto-Fix Issues",
    description: "Fix system issues with confirmation",
    command: "chauf doctor --auto-fix",
    output: "🔧 Fix plan shown, waiting for confirmation"
  },
  {
    name: "View Logs",
    description: "View nginx logs",
    command: "chauf logs nginx --follow",
    output: "Following nginx logs..."
  }
];

export const DOCS_NAVIGATION = [
  {
    title: "Getting Started",
    items: [
      { title: "Installation", slug: "getting-started/installation" },
      { title: "First Project", slug: "getting-started/first-project" },
      { title: "Architecture", slug: "getting-started/architecture" }
    ]
  },
  {
    title: "Core Concepts",
    items: [
      { title: "Project Linking", slug: "core/linking" },
      { title: "PHP Versions", slug: "core/php-versions" },
      { title: "Nginx", slug: "core/nginx" },
      { title: "Composer", slug: "core/composer" },
      { title: "SSL & Domains", slug: "core/ssl-domains" }
    ]
  },
  {
    title: "Reference",
    items: [
      { title: "CLI Commands", slug: "reference/commands", badge: "Updated" },
      { title: "Configuration", slug: "reference/configuration" },
      { title: "Dependencies", slug: "reference/dependencies", badge: "New" },
      { title: "Troubleshooting", slug: "reference/troubleshooting" }
    ]
  }
];