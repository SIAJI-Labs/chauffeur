import { LucideIcon } from 'lucide-react';

export interface TerminalLine {
  text: string;
  type: 'command' | 'output' | 'success' | 'error' | 'info';
  delay?: number;
}

export interface Feature {
  title: string;
  description: string;
  icon: LucideIcon;
}

export interface CommandExample {
  name: string;
  description: string;
  command: string;
  output: string;
}

export enum InstallMethod {
  CURL = 'curl',
  WGET = 'wget',
  APT = 'apt',
}