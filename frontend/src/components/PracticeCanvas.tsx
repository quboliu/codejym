import { useMemo } from 'react';
import type { FileContent } from '../types';

interface PracticeCanvasProps {
  content: FileContent | null;
  cursor: number;
  errorFlash: boolean;
}

export function PracticeCanvas({ content, cursor, errorFlash }: PracticeCanvasProps) {
  const { completed, currentChar, remaining } = useMemo(() => {
    if (!content) {
      return { completed: '', currentChar: '', remaining: '' };
    }
    const completedText = content.content.slice(0, cursor);
    const current = content.content.slice(cursor, cursor + 1);
    const rest = content.content.slice(cursor + 1);
    return { completed: completedText, currentChar: current, remaining: rest };
  }, [content, cursor]);

  if (!content) {
    return (
      <div className="practice-empty">
        <p>选择一个源码文件开始临摹。</p>
      </div>
    );
  }

  return (
    <div className={`practice-canvas ${errorFlash ? 'shake' : ''}`}>
      <div className="canvas-header">
        <div>
          <div className="canvas-path">{content.path}</div>
          <div className="canvas-language">{content.language.toUpperCase()}</div>
        </div>
        <div className="next-char">
          下一字符：<kbd>{displayChar(currentChar)}</kbd>
        </div>
      </div>
      <div className="code-wrapper">
        <pre className="code-layer base">{content.content}</pre>
        <pre className="code-layer overlay" aria-hidden="true">
          <span className="code-completed">{completed}</span>
          <span className="code-rest">{remaining}</span>
        </pre>
      </div>
    </div>
  );
}

function displayChar(char: string) {
  if (!char) return '完成';
  if (char === '\n') return '↵';
  if (char === '\t') return '⇥';
  if (char === ' ') return '␠';
  return char;
}
