"use client";

import React, { useEffect } from 'react';
import { addToSearchIndex } from './SearchData';

interface SearchInitializerProps {
  children?: React.ReactNode;
}

export const SearchInitializer: React.FC<SearchInitializerProps> = ({ children }) => {
  useEffect(() => {
    // Clear existing index to avoid duplicates
    // Note: In a real implementation, you might want to build this at build time
    // For now, we'll populate it on client side

    // Getting Started Pages
    addToSearchIndex({
      title: "Installation",
      slug: "getting-started/installation",
      content: "Install Chauffeur on Linux. Chauffeur provides an installer script that automatically downloads, builds, and installs the binary. The installer detects your system and installs Chauffeur to ~/.chauffeur. Quick install: curl -fsSL https://chauffeur.siaji.com/install.sh | bash. Manual install: git clone https://github.com/SIAJI-Labs/chauffeur.git && cd chauffeur && go build -o ~/.chauffeur/bin/chauf ./cli/main.go.",
      section: "Getting Started",
      category: "Getting Started"
    });

    addToSearchIndex({
      title: "First Project",
      slug: "getting-started/first-project",
      content: "Create your first project with Chauffeur. Initialize workspace, install PHP versions, link projects, and start developing. Commands: chauf init, chauf install php 8.3, chauf link --secure, chauf start. Chauffeur automatically detects project types like Laravel, Symfony, WordPress.",
      section: "Getting Started",
      category: "Getting Started"
    });

    addToSearchIndex({
      title: "Architecture",
      slug: "getting-started/architecture",
      content: "Chauffeur architecture and how it works. Linux native application using systemd for service management. Nginx reverse proxy with PHP-FPM pools. Automatic DNS resolution for .test domains. SSL certificate generation with mkcert. Resource efficient with 98% less RAM usage than containers.",
      section: "Getting Started",
      category: "Getting Started"
    });

    // Core Concepts Pages
    addToSearchIndex({
      title: "Project Linking",
      slug: "core/linking",
      content: "Link projects to Chauffeur for development. Commands: chauf link, chauf links, chauf unlink. Project linking automatically detects project type and configures Nginx and PHP-FPM. Support for Laravel, Symfony, WordPress, and custom PHP applications.",
      section: "Core Concepts",
      category: "Core Concepts"
    });

    addToSearchIndex({
      title: "PHP Versions",
      slug: "core/php-versions",
      content: "Multi-version PHP support from 7.4 to 8.4. Install multiple PHP versions: chauf php install 8.3. Set global version: chauf php use 8.3. Project-specific versions: chauf php isolate 8.1. List installed versions: chauf php list. Each project gets isolated PHP-FPM pools.",
      section: "Core Concepts",
      category: "Core Concepts"
    });

    addToSearchIndex({
      title: "Nginx",
      slug: "core/nginx",
      content: "Nginx reverse proxy configuration. Chauffeur automatically configures Nginx for each linked project with SSL support, custom domains, and PHP-FPM integration. Rewrite rules for Laravel routes. Static file serving. Gzip compression enabled.",
      section: "Core Concepts",
      category: "Core Concepts"
    });

    addToSearchIndex({
      title: "Composer",
      slug: "core/composer",
      content: "Composer integration for PHP package management. Auto-installed with PHP versions. Global Composer available. Project-specific composer.phar support. Authentication for private packages. Cache optimization for faster installs.",
      section: "Core Concepts",
      category: "Core Concepts"
    });

    addToSearchIndex({
      title: "SSL & Domains",
      slug: "core/ssl-domains",
      content: "SSL certificate management and custom domains. Generate trusted SSL certificates with --secure flag. Custom domain support with chauf link --domain=example.com. Automatic .test domain resolution. Local development with HTTPS.",
      section: "Core Concepts",
      category: "Core Concepts"
    });

    // Reference Pages
    addToSearchIndex({
      title: "CLI Commands",
      slug: "reference/commands",
      content: "Complete CLI command reference. chauf init, chauf install, chauf link, chauf links, chauf unlink, chauf start, chauf stop, chauf restart, chauf status, chauf logs, chauf php, chauf doctor, chauf clean, chauf self-update. All command options and flags documented.",
      section: "Reference",
      category: "Reference"
    });

    addToSearchIndex({
      title: "Configuration",
      slug: "reference/configuration",
      content: "Chauffeur configuration files and options. Main config at ~/.chauffeur/config/chauffeur.yaml. Project-specific configs. Nginx template customization. PHP-FPM settings. SSL certificate paths. Port configuration.",
      section: "Reference",
      category: "Reference"
    });

    addToSearchIndex({
      title: "Troubleshooting",
      slug: "reference/troubleshooting",
      content: "Common issues and solutions. Installation problems, PHP version conflicts, Nginx configuration errors, SSL certificate issues, service management problems. Debug commands: chauf status, chauf logs, chauf doctor.",
      section: "Reference",
      category: "Reference"
    });

  }, []);

  return <>{children}</>;
};