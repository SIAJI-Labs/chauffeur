// React
import React, { useState, useEffect, useRef } from 'react';

// Third-party libraries
import { Copy, Check } from 'lucide-react';

// Internal types
import { TerminalLine } from '@/types';

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
    <div className={`rounded-xl overflow-hidden shadow-2xl border bg-chauffeur-background font-mono text-sm ${className}`}>
      {/* Terminal Header */}
      <div className="flex items-center justify-between px-4 py-3 bg-chauffeur-surface/50 border-b">
        <div className="flex items-center gap-2">
          <div className="w-3 h-3 rounded-full bg-red-500/80" />
          <div className="w-3 h-3 rounded-full bg-yellow-500/80" />
          <div className="w-3 h-3 rounded-full bg-green-500/80" />
        </div>
        <div className="text-chauffeur-muted text-xs font-medium">user@linux:~/.chauffeur</div>
        <button onClick={handleCopy} className="text-muted-foreground hover:text-foreground transition-colors">
          {isCopied ? <Check size={14} /> : <Copy size={14} />}
        </button>
      </div>

      {/* Terminal Body */}
      <div
        ref={terminalBodyRef}
        className="p-6 h-[320px] overflow-y-auto text-chauffeur-muted space-y-2 scrollbar-thin"
      >
        {displayedLines.map((line, idx) => (
          <div key={idx} className={`${line.type === 'command' ? 'text-foreground font-bold' : ''} ${line.type === 'success' ? 'text-primary' : ''} ${line.type === 'error' ? 'text-destructive' : ''} animate-fade-in`}>
            {line.type === 'command' && <span className="text-primary mr-2">$</span>}
            {line.text}
          </div>
        ))}

        {/* Current Typing Line */}
        {currentLineIndex < lines.length && lines[currentLineIndex].type === 'command' && (
          <div className="text-foreground font-bold">
            <span className="text-primary mr-2">$</span>
            {currentText}
            <span className="inline-block w-2 h-4 bg-muted-foreground ml-1 animate-cursor-blink align-middle"></span>
          </div>
        )}

        {/* Final prompt when done */}
        {currentLineIndex >= lines.length && (
           <div className="text-foreground font-bold mt-4">
             <span className="text-primary mr-2">$</span>
             <span className="inline-block w-2 h-4 bg-muted-foreground animate-cursor-blink align-middle"></span>
           </div>
        )}
      </div>
    </div>
  );
};