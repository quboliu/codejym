import { useState } from 'react';
import type { FileNode } from '../types';

interface FileTreeProps {
  nodes: FileNode[];
  activePath?: string | null;
  onSelect: (path: string) => void;
}

export function FileTree({ nodes, activePath, onSelect }: FileTreeProps) {
  if (!nodes.length) {
    return <p className="muted">选择一个文件夹以查看树状结构。</p>;
  }
  return (
    <div className="file-tree">
      {nodes.map((node) => (
        <TreeItem key={node.path} node={node} depth={0} activePath={activePath} onSelect={onSelect} />
      ))}
    </div>
  );
}

interface TreeItemProps {
  node: FileNode;
  depth: number;
  activePath?: string | null;
  onSelect: (path: string) => void;
}

function TreeItem({ node, depth, activePath, onSelect }: TreeItemProps) {
  const [open, setOpen] = useState(true);
  const paddingLeft = 12 * depth;
  if (node.isDir) {
    return (
      <div className="tree-item dir" style={{ paddingLeft }}>
        <button className="tree-toggle" onClick={() => setOpen((prev) => !prev)}>
          {open ? '▾' : '▸'}
        </button>
        <span className="tree-name">{node.name}</span>
        {open && node.children && node.children.length > 0 && (
          <div className="tree-children">
            {node.children.map((child) => (
              <TreeItem key={child.path} node={child} depth={depth + 1} activePath={activePath} onSelect={onSelect} />
            ))}
          </div>
        )}
      </div>
    );
  }
  const isActive = activePath === node.path;
  return (
    <button
      className={`tree-item file ${isActive ? 'active' : ''}`}
      style={{ paddingLeft }}
      onClick={() => onSelect(node.path)}
    >
      <span className="tree-name">{node.name}</span>
    </button>
  );
}
