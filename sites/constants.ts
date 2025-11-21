import { Server, Shield, Zap, Globe, Layers, Terminal, Cpu } from 'lucide-react';
import { TerminalLine, Feature, CommandExample } from './types';

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