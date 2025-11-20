import React, { useState } from 'react';
import { Check, Copy, Terminal } from 'lucide-react';

interface CodeBlockProps {
  code: string;
  language?: string;
  showLineNumbers?: boolean;
  className?: string;
}

export const CodeBlock: React.FC<CodeBlockProps> = ({ 
  code, 
  language = 'bash', 
  showLineNumbers = false,
  className = '' 
}) => {
  const [copied, setCopied] = useState(false);

  const handleCopy = () => {
    navigator.clipboard.writeText(code);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const lines = code.trim().split('\n');

  return (
    <div className={`relative group rounded-lg overflow-hidden border border-slate-700 bg-[#0F172A] my-6 ${className}`}>
      <div className="flex items-center justify-between px-4 py-2 bg-slate-800/50 border-b border-slate-700">
        <div className="flex items-center gap-2">
          <Terminal size={14} className="text-slate-500" />
          <span className="text-xs font-mono text-slate-400 lowercase">{language}</span>
        </div>
        <button 
          onClick={handleCopy}
          className="p-1.5 rounded-md hover:bg-slate-700 text-slate-400 hover:text-white transition-colors"
          aria-label="Copy code"
        >
          {copied ? <Check size={14} className="text-emerald-400" /> : <Copy size={14} />}
        </button>
      </div>
      <div className="p-4 overflow-x-auto">
        <pre className="font-mono text-sm leading-relaxed">
          {lines.map((line, i) => (
            <div key={i} className="table-row">
              {showLineNumbers && (
                <span className="table-cell text-right pr-4 text-slate-600 select-none w-8">{i + 1}</span>
              )}
              <span className="table-cell text-slate-300">
                {language === 'bash' && line.startsWith('$') ? (
                  <>
                    <span className="text-emerald-500 select-none mr-2">$</span>
                    {line.substring(1)}
                  </>
                ) : language === 'bash' && line.startsWith('#') ? (
                  <span className="text-slate-500 italic">{line}</span>
                ) : (
                  line
                )}
              </span>
            </div>
          ))}
        </pre>
      </div>
    </div>
  );
};