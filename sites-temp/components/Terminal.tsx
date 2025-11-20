import React, { useState, useEffect, useRef } from 'react';
import { TerminalLine } from '../types';
import { Copy, Check } from 'lucide-react';

interface TerminalProps {
  lines: TerminalLine[];
  className?: string;
  autoPlay?: boolean;
}

export const Terminal: React.FC<TerminalProps> = ({ lines, className = "", autoPlay = true }) => {
  const [displayedLines, setDisplayedLines] = useState<TerminalLine[]>([]);
  const [currentLineIndex, setCurrentLineIndex] = useState(0);
  const [currentText, setCurrentText] = useState('');
  const [isTyping, setIsTyping] = useState(false);
  const [isCopied, setIsCopied] = useState(false);
  const terminalBodyRef = useRef<HTMLDivElement>(null);

  // Handle Copy
  const handleCopy = () => {
    navigator.clipboard.writeText("curl -sL https://chauffeur.dev/install | bash");
    setIsCopied(true);
    setTimeout(() => setIsCopied(false), 2000);
  };

  // Typing Effect Logic
  useEffect(() => {
    if (!autoPlay) return;

    if (currentLineIndex >= lines.length) return;

    const line = lines[currentLineIndex];
    
    if (line.type === 'command' && currentText.length < line.text.length) {
      setIsTyping(true);
      const timeout = setTimeout(() => {
        setCurrentText(line.text.slice(0, currentText.length + 1));
      }, 50 + Math.random() * 50); // Random typing speed
      return () => clearTimeout(timeout);
    } 
    
    if (line.type === 'command' && currentText.length === line.text.length) {
       setIsTyping(false);
       const timeout = setTimeout(() => {
         setDisplayedLines(prev => [...prev, line]);
         setCurrentText('');
         setCurrentLineIndex(prev => prev + 1);
       }, 400);
       return () => clearTimeout(timeout);
    }

    if (line.type !== 'command') {
      const timeout = setTimeout(() => {
        setDisplayedLines(prev => [...prev, line]);
        setCurrentLineIndex(prev => prev + 1);
      }, line.delay || 100);
      return () => clearTimeout(timeout);
    }

  }, [currentLineIndex, currentText, lines, autoPlay]);

  // Auto scroll
  useEffect(() => {
    if (terminalBodyRef.current) {
      terminalBodyRef.current.scrollTop = terminalBodyRef.current.scrollHeight;
    }
  }, [displayedLines, currentText]);

  return (
    <div className={`rounded-xl overflow-hidden shadow-2xl border border-slate-700 bg-[#0F172A] font-mono text-sm ${className}`}>
      {/* Terminal Header */}
      <div className="flex items-center justify-between px-4 py-3 bg-slate-800/50 border-b border-slate-700">
        <div className="flex items-center gap-2">
          <div className="w-3 h-3 rounded-full bg-red-500/80" />
          <div className="w-3 h-3 rounded-full bg-yellow-500/80" />
          <div className="w-3 h-3 rounded-full bg-green-500/80" />
        </div>
        <div className="text-slate-400 text-xs font-medium">user@linux:~/.chauffeur</div>
        <button onClick={handleCopy} className="text-slate-500 hover:text-slate-300 transition-colors">
          {isCopied ? <Check size={14} /> : <Copy size={14} />}
        </button>
      </div>

      {/* Terminal Body */}
      <div 
        ref={terminalBodyRef}
        className="p-6 h-[320px] overflow-y-auto text-slate-300 space-y-2 scrollbar-thin scrollbar-thumb-slate-700 scrollbar-track-transparent"
      >
        {displayedLines.map((line, idx) => (
          <div key={idx} className={`${line.type === 'command' ? 'text-slate-100 font-bold' : ''} ${line.type === 'success' ? 'text-emerald-400' : ''} ${line.type === 'error' ? 'text-red-400' : ''} animate-fade-in`}>
            {line.type === 'command' && <span className="text-emerald-500 mr-2">$</span>}
            {line.text}
          </div>
        ))}
        
        {/* Current Typing Line */}
        {currentLineIndex < lines.length && lines[currentLineIndex].type === 'command' && (
          <div className="text-slate-100 font-bold">
            <span className="text-emerald-500 mr-2">$</span>
            {currentText}
            <span className="inline-block w-2 h-4 bg-slate-400 ml-1 animate-cursor-blink align-middle"></span>
          </div>
        )}

        {/* Final prompt when done */}
        {currentLineIndex >= lines.length && (
           <div className="text-slate-100 font-bold mt-4">
             <span className="text-emerald-500 mr-2">$</span>
             <span className="inline-block w-2 h-4 bg-slate-400 animate-cursor-blink align-middle"></span>
           </div>
        )}
      </div>
    </div>
  );
};