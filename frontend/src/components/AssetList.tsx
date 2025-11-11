import type { Asset } from '../types';

interface AssetListProps {
  assets: Asset[];
  selectedId?: string | null;
  onSelect: (assetId: string) => void;
}

export function AssetList({ assets, selectedId, onSelect }: AssetListProps) {
  if (assets.length === 0) {
    return <p className="muted">暂时没有字帖素材，上传一个源码文件或压缩包开始。</p>;
  }
  return (
    <div className="asset-grid">
      {assets.map((asset) => {
        const active = selectedId === asset.id;
        const extension = getExtension(asset.sourceName);
        return (
          <button key={asset.id} className={`asset-card ${active ? 'active' : ''}`} onClick={() => onSelect(asset.id)}>
            <div className="asset-name">
              {asset.name}
              {extension && <span className="asset-ext">.{extension}</span>}
            </div>
            <div className="asset-meta">
              <span>{formatBytes(asset.sizeBytes)}</span>
              <span>·{asset.fileCount} 文件</span>
            </div>
            <div className="asset-date">{new Date(asset.createdAt).toLocaleString()}</div>
          </button>
        );
      })}
    </div>
  );
}

function getExtension(name: string) {
  const clean = (name ?? '').trim();
  if (!clean) return '';
  const lastDot = clean.lastIndexOf('.');
  if (lastDot <= 0 || lastDot === clean.length - 1) {
    return '';
  }
  return clean.slice(lastDot + 1).toLowerCase();
}

function formatBytes(bytes: number) {
  if (!bytes) {
    return '0 B';
  }
  const units = ['B', 'KB', 'MB', 'GB'];
  const exponent = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1);
  const value = bytes / Math.pow(1024, exponent);
  return `${value.toFixed(value >= 10 || exponent === 0 ? 0 : 1)} ${units[exponent]}`;
}
