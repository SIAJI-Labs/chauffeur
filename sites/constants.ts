import { Server, Shield, Zap, Globe, Layers, Terminal, Cpu, Settings, Link, Link2, FolderOpen, FileText, AlertTriangle, Wrench, Download } from 'lucide-react';
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

// Complete CLI Commands Reference - Matches AGENTS.md specification
// Organized by category with most essential commands first
export const CLI_COMMANDS: CommandDefinition[] = [
  // WORKSPACE - Most fundamental command first
  {
    command: 'chauf init',
    category: 'workspace',
    description: 'Bootstrap workspace under ~/.chauffeur/. Idempotent operation.',
    usage: 'chauf init [--force] [--quiet]',
    keyFlags: [
      { flag: '--force', description: 'Reinitialize existing workspace', required: false },
      { flag: '--quiet', description: 'Minimal output during initialization', required: false }
    ],
    examples: [
      { name: 'Initialize workspace', description: 'Create Chauffeur workspace', command: 'chauf init', output: '✓ Creating workspace at ~/.chauffeur' }
    ],
    notes: ['Safe to run multiple times', 'Creates ~/.chauffeur directory structure']
  },

  // PROJECTS - Right after chauf init
  {
    command: 'chauf link',
    category: 'projects',
    description: 'Register current directory, detect template, generate configs with multi-domain support.',
    usage: 'chauf link [--site] [--ssl] [--php <version>] [--http-port <port>] [--https-port <port>] [--alias <domain>] [--add-alias <domain>] [--force]',
    keyFlags: [
      { flag: '--site', description: 'Site name for domain', required: false },
      { flag: '--ssl', description: 'Enable HTTPS with SSL certificate', required: false },
      { flag: '--php <version>', description: 'PHP version for this project', required: false },
      { flag: '--http-port <port>', description: 'Custom HTTP port', required: false },
      { flag: '--https-port <port>', description: 'Custom HTTPS port', required: false },
      { flag: '--alias <domain>', description: 'Add domain alias', required: false },
      { flag: '--add-alias <domain>', description: 'Additional domain alias', required: false },
      { flag: '--dedicated-fpm', description: 'Use dedicated PHP-FPM pool', required: false },
      { flag: '--force', description: 'Overwrite existing configuration', required: false }
    ],
    examples: [
      { name: 'Link with SSL', description: 'Link project with HTTPS', command: 'chauf link --ssl', output: '✓ Linked to https://my-app.test [PHP 8.3]' },
      { name: 'Link with custom PHP', description: 'Link with specific PHP version', command: 'chauf link --php 8.1', output: '✓ Linked to http://my-app.test [PHP 8.1 Dedicated]' }
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
      { name: 'List projects', description: 'Show all linked projects', command: 'chauf links', output: '📋 Found 3 linked projects' },
      { name: 'View project details', description: 'Show detailed project information', command: 'chauf links --slug my-project', output: '📋 Project Details: my-project' }
    ],
    notes: ['--slug and --site flags are mutually exclusive']
  },
  {
    command: 'chauf unlink',
    category: 'projects',
    description: 'Remove registrations or specific aliases. Defaults to current dir.',
    usage: 'chauf unlink [--slug] [--site] [--project] [--alias] [--all] [--force]',
    keyFlags: [
      { flag: '--slug <slug>', description: 'Project slug to unlink', required: false },
      { flag: '--site <site>', description: 'Site name to unlink', required: false },
      { flag: '--project <path>', description: 'Project path to unlink', required: false },
      { flag: '--alias <domain>', description: 'Remove specific alias', required: false },
      { flag: '--all', description: 'Unlink all projects', required: false },
      { flag: '--force', description: 'Force unlink without confirmation', required: false }
    ],
    examples: [
      { name: 'Unlink current project', description: 'Unlink current directory', command: 'chauf unlink', output: '✓ Unlinked my-app.test' }
    ]
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
        output: `[ secure ] Checking for mkcert availability...
[ secure ] mkcert found - will generate trusted certificates
[ secure ] Generating SSL certificates ⠋[ secure ] Generating multi-domain certificate for:
[ secure ]   - my-project.test (HTTPS)
[ secure ] Generating templates ⠙100.8ms[ secure ] Generating trusted multi-domain SSL certificate using mkcert
[ secure ] Generating SSL certificates ⠇801.2ms...[ secure ] ✓ Multi-domain SSL certificate generated successfully with mkcert (domains: my-project.test)
[ secure ] Generating SSL certificates ✓ (891.2ms)
  └── ✓ Multi-domain SSL certificates generated
[ secure ] ✓ SSL certificate added successfully
[ secure ] ✓ Trusted SSL certificate generated (mkcert certificate is automatically trusted by browsers)
[ secure ]   Secure access: https://my-project.test`
      }
    ],
    notes: [
      'Must be run from a linked project directory',
      'Uses mkcert for trusted certificates if available',
      'Falls back to self-signed certificates if mkcert not available',
      'Automatically reloads nginx configuration'
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
        output: `[ unsecure ] Removing SSL certificate...
[ unsecure ] ✓ SSL certificate removed successfully
[ unsecure ]   HTTP access: http://my-project.test`
      }
    ],
    notes: [
      'Must be run from a linked project directory',
      'Removes SSL certificate files and configuration',
      'Automatically reloads nginx configuration'
    ]
  },

  // SERVICES - Service management commands
  {
    command: 'chauf install',
    category: 'services',
    description: 'Install Chauffeur-managed services (nginx, php, composer).',
    usage: 'chauf install <service> [--force] [--local] [--no-cache] [version]',
    keyFlags: [
      { flag: '--force', description: 'Reinstall even if already present', required: false, default: 'false' },
      { flag: '--local', description: 'Use local PHP tarball (PHP only)', required: false, default: 'false' },
      { flag: '--no-cache', description: 'Skip download caching', required: false, default: 'false' }
    ],
    examples: [
      { name: 'Install nginx', description: 'Install nginx web server', command: 'chauf install nginx', output: '✓ nginx installed successfully' },
      { name: 'Install composer', description: 'Install Composer dependency manager', command: 'chauf install composer', output: '✓ composer installed successfully' },
      { name: 'Install PHP version', description: 'Install specific PHP version', command: 'chauf install php 8.2', output: '✓ PHP 8.2 installed successfully' }
    ],
    notes: ['Service parameter is required. Available services: nginx, php, composer']
  },
  {
    command: 'chauf start',
    category: 'services',
    description: 'Start nginx/PHP-FPM plus dnsmasq validation.',
    usage: 'chauf start [--project <path>] [--all] [--dry-run]',
    keyFlags: [
      { flag: '--project <path>', description: 'Start specific project only', required: false },
      { flag: '--all', description: 'Start all services', required: false, default: 'true' },
      { flag: '--dry-run', description: 'Show what would be started without executing', required: false }
    ],
    examples: [
      { name: 'Start all services', description: 'Start all Chauffeur services', command: 'chauf start', output: '✓ Started nginx (8080), php-fpm 8.3 (9000)' }
    ]
  },
  {
    command: 'chauf stop',
    category: 'services',
    description: 'Stop services and clean port-forward rules.',
    usage: 'chauf stop [--project <path>] [--all] [--dry-run]',
    keyFlags: [
      { flag: '--project <path>', description: 'Stop specific project only', required: false },
      { flag: '--all', description: 'Stop all services', required: false, default: 'true' },
      { flag: '--dry-run', description: 'Show what would be stopped without executing', required: false }
    ],
    examples: [
      { name: 'Stop all services', description: 'Stop all Chauffeur services', command: 'chauf stop', output: '✓ Stopped nginx, php-fpm, port forwarding' }
    ]
  },
  {
    command: 'chauf restart',
    category: 'services',
    description: 'Restart services (equivalent to stop then start, preserves configuration).',
    usage: 'chauf restart [--project <slug>] [--all] [--dry-run]',
    keyFlags: [
      { flag: '--project <slug>', description: 'Restart specific project', required: false },
      { flag: '--all', description: 'Restart all services', required: false, default: 'true' },
      { flag: '--dry-run', description: 'Show what would be restarted without executing', required: false }
    ],
    examples: [
      { name: 'Restart all services', description: 'Restart all services with preserved config', command: 'chauf restart', output: '✓ Restarted nginx (8080), php-fpm pools' }
    ]
  },
  {
    command: 'chauf status',
    category: 'services',
    description: 'Show status for global or per-project services.',
    usage: 'chauf status [service-type] [--project] [--detail] [-v]',
    keyFlags: [
      { flag: '--project', description: 'Show project-specific status', required: false },
      { flag: '--detail', description: 'Show detailed service information', required: false },
      { flag: '-v', description: 'Verbose output', required: false }
    ],
    examples: [
      { name: 'Show all services', description: 'Display service status', command: 'chauf status', output: 'nginx: running (8080), php-fpm: running (2 pools)' }
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
    notes: ['First installation may take 10-15 minutes', 'Requires build-essential toolchain']
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
    description: 'Pin current linked project to a version.',
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
[ php current ] PHP binary: /home/user/.chauffeur/runtimes/php/8.3/bin/php`
      },
      {
        name: 'Show current PHP outside project',
        description: 'Display global PHP when not in a project directory',
        command: 'chauf php current',
        output: `[ php current ] No project detected in current directory
[ php current ] Global PHP: 8.2 (default)
[ php current ] PHP binary: /home/user/.chauffeur/runtimes/php/8.2/bin/php`
      }
    ],
    notes: [
      'Shows project-specific PHP if directory is linked to Chauffeur',
      'Shows global PHP default when not in a project directory',
      'Displays PHP binary path for reference'
    ]
  },
  {
    command: 'chauf install composer',
    category: 'php',
    description: 'Fetch verified composer PHAR and shims.',
    usage: 'chauf install composer',
    keyFlags: [],
    examples: [
      { name: 'Install Composer', description: 'Install Composer dependency manager', command: 'chauf install composer', output: '✓ Composer installed' }
    ]
  },
  {
    command: 'chauf install php',
    category: 'php',
    description: 'Build/install specific PHP runtime version under workspace.',
    usage: 'chauf install php <version> [--force] [--no-ext] [--from <source>]',
    keyFlags: [
      { flag: '--force', description: 'Reinstall even if already exists', required: false, default: 'false' },
      { flag: '--no-ext', description: 'Skip extension compilation', required: false, default: 'false' },
      { flag: '--from <source>', description: 'Install from specific source', required: false, default: 'package' }
    ],
    examples: [
      { name: 'Install PHP version', description: 'Install specific PHP version', command: 'chauf install php 8.2', output: '✓ PHP 8.2 installed successfully' }
    ]
  },

  // SYSTEM - System commands
  {
    command: 'chauf self-update',
    category: 'system',
    description: 'Update binary from git or rebuild from current repo.',
    usage: 'chauf self-update [--dev]',
    keyFlags: [
      { flag: '--dev', description: 'Rebuild from current repo', required: false }
    ],
    examples: [
      { name: 'Update Chauffeur', description: 'Update to latest version', command: 'chauf self-update', output: '✓ Updated to v1.2.0' }
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
      { name: 'Uninstall Chauffeur', description: 'Remove Chauffeur workspace', command: 'chauf uninstall', output: '✓ Workspace removed' }
    ]
  },
  {
    command: 'chauf remove',
    category: 'php',
    description: 'Remove installed runtimes (php, nginx, composer).',
    usage: 'chauf remove <service> [version] [--force]',
    keyFlags: [
      { flag: '--force', description: 'Force removal without confirmation', required: false }
    ],
    examples: [
      { name: 'Remove PHP version', description: 'Remove specific PHP version', command: 'chauf remove php 8.1', output: '✓ Removed PHP 8.1' }
    ]
  },

  // UTILITIES - Finally utilities
  {
    command: 'chauf doctor',
    category: 'utilities',
    description: 'Perform health checks and diagnose system issues.',
    usage: 'chauf doctor [options]',
    keyFlags: [
      { flag: '--check-all, -a', description: 'Run all dependency checks (default)', required: false },
      { flag: '--check-deps, -d', description: 'Check system dependencies (git, curl, tar, etc.)', required: false },
      { flag: '--check-php, -p', description: 'Check PHP build dependencies and headers', required: false },
      { flag: '--check-ssl, -s', description: 'Check SSL certificate dependencies', required: false },
      { flag: '--check-network, -n', description: 'Check network and port availability', required: false },
      { flag: '--check-dns', description: 'Check DNS resolution for .test domains', required: false },
      { flag: '--verbose, -v', description: 'Show detailed diagnostic information', required: false },
      { flag: '--fix, -f', description: 'Show fix suggestions for issues found', required: false },
      { flag: '--auto-fix', description: 'Attempt to automatically fix issues where safe', required: false },
      { flag: '--quiet, -q', description: 'Suppress non-error output', required: false }
    ],
    examples: [
      { name: 'Full health check', description: 'Run all health checks', command: 'chauf doctor', output: '✓ System health check completed' },
      { name: 'Check dependencies only', description: 'Check only system dependencies', command: 'chauf doctor --check-deps', output: '✓ System dependencies OK' }
    ],
    notes: ['Provides comprehensive validation and guidance for resolving issues']
  },
  {
    command: 'chauf info',
    category: 'utilities',
    description: 'Show workspace paths, installed services, versions, port config.',
    usage: 'chauf info',
    keyFlags: [],
    examples: [
      { name: 'Show info', description: 'Display workspace information', command: 'chauf info', output: 'Workspace: ~/.chauffeur, PHP: 8.3, nginx: running' }
    ]
  },
  {
    command: 'chauf logs',
    category: 'utilities',
    description: 'View and follow service logs with interactive version selection.',
    usage: 'chauf logs [service] [version] [--follow] [-f] [--lines] [-n] [--level] [--context] [-c] [--verbose] [-v] [--quiet] [-q]',
    keyFlags: [
      { flag: '--follow, -f', description: 'Follow log output', required: false },
      { flag: '--lines, -n <num>', description: 'Number of lines to show', required: false, default: '50' },
      { flag: '--level <level>', description: 'Filter by log level', required: false },
      { flag: '--context, -c', description: 'Show context around matches', required: false },
      { flag: '--verbose, -v', description: 'Verbose output', required: false },
      { flag: '--quiet, -q', description: 'Quiet output', required: false }
    ],
    examples: [
      { name: 'View logs', description: 'Show service logs', command: 'chauf logs nginx', output: 'nginx: 2024-01-01 12:00:00 [info] Server started' }
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
      { flag: '--older-than <days>', description: 'Clean files older than days', required: false },
      { flag: '--keep-versions <num>', description: 'Keep N versions', required: false },
      { flag: '--what', description: 'Show what would be cleaned', required: false }
    ],
    examples: [
      { name: 'Clean workspace', description: 'Clean old files', command: 'chauf clean --dry-run', output: 'Would clean 2.3GB of old files' }
    ]
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
    id: 'services',
    title: 'Service Management',
    description: 'Install, start, stop, and manage Chauffeur services',
    icon: Server
  },
  {
    id: 'utilities',
    title: 'Utilities',
    description: 'Logs, cleanup, and diagnostic tools',
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
  { text: 'chauf link my-app --ssl --php 8.3', type: 'command', delay: 1200 },
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
      { title: "Troubleshooting", slug: "reference/troubleshooting" }
    ]
  }
];