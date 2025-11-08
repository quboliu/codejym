import { useEffect, useMemo, useRef, useState } from 'react';
import './App.css';
import { AssetList } from './components/AssetList';
import { FileTree } from './components/FileTree';
import { PracticeCanvas } from './components/PracticeCanvas';
import {
  createSession,
  fetchFileContent,
  fetchFileTree,
  fetchSession,
  listAssets,
  patchSession,
  uploadAsset,
} from './api';
import type { Asset, FileNode, FileContent, Session } from './types';

function App() {
  const [assets, setAssets] = useState<Asset[]>([]);
  const [assetLoading, setAssetLoading] = useState(false);
  const [selectedAsset, setSelectedAsset] = useState<string | null>(null);
  const [tree, setTree] = useState<FileNode[]>([]);
  const [treeLoading, setTreeLoading] = useState(false);
  const [selectedPath, setSelectedPath] = useState<string | null>(null);
  const [fileContent, setFileContent] = useState<FileContent | null>(null);
  const [session, setSession] = useState<Session | null>(null);
  const [cursor, setCursor] = useState(0);
  const [errors, setErrors] = useState(0);
  const [elapsedSeconds, setElapsedSeconds] = useState(0);
  const [message, setMessage] = useState<string | null>(null);
  const [uploading, setUploading] = useState(false);
  const [flashError, setFlashError] = useState(false);
  const errorTimer = useRef<number | null>(null);

  useEffect(() => {
    refreshAssets();
  }, []);

  useEffect(() => {
    if (!session || !fileContent) {
      return;
    }
    const timer = window.setInterval(() => {
      setElapsedSeconds((prev) => prev + 1);
    }, 1000);
    return () => window.clearInterval(timer);
  }, [session?.id, fileContent]);

  useEffect(() => {
    if (!session) {
      return;
    }
    const timeout = window.setTimeout(() => {
      patchSession(session.id, {
        cursor,
        errors,
        durationSeconds: Math.round(elapsedSeconds),
      }).catch((err) => console.warn('session sync failed', err));
    }, 1200);
    return () => window.clearTimeout(timeout);
  }, [session?.id, cursor, errors, elapsedSeconds]);

  useEffect(() => {
    const handler = (event: KeyboardEvent) => {
      if (!fileContent) return;
      if (['INPUT', 'TEXTAREA'].includes(document.activeElement?.tagName ?? '')) {
        return;
      }
      if (event.metaKey || event.ctrlKey || event.altKey) {
        return;
      }
      if (event.key === 'Backspace') {
        event.preventDefault();
        setCursor((prev) => Math.max(0, prev - 1));
        return;
      }
      if (cursor >= fileContent.content.length) {
        return;
      }
      const char = mapKeyToChar(event);
      if (char === null) {
        return;
      }
      event.preventDefault();
      const expected = fileContent.content.charAt(cursor);
      if (expected === char) {
        setCursor((prev) => Math.min(fileContent.content.length, prev + 1));
      } else {
        setErrors((prev) => prev + 1);
        setFlashError(true);
        if (errorTimer.current) {
          window.clearTimeout(errorTimer.current);
        }
        errorTimer.current = window.setTimeout(() => setFlashError(false), 200);
      }
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [fileContent, cursor]);

  const progress = useMemo(() => {
    if (!fileContent) return 0;
    if (fileContent.content.length === 0) return 0;
    return Math.round((cursor / fileContent.content.length) * 100);
  }, [cursor, fileContent]);

  const accuracy = useMemo(() => {
    if (cursor + errors === 0) return 100;
    return Math.max(0, Math.round((cursor / (cursor + errors)) * 100));
  }, [cursor, errors]);

  async function refreshAssets() {
    setAssetLoading(true);
    try {
      const data = await listAssets();
      setAssets(data);
      if (data.length && !selectedAsset) {
        setSelectedAsset(data[0].id);
        loadTree(data[0].id);
      }
    } catch (err) {
      setMessage((err as Error).message);
    } finally {
      setAssetLoading(false);
    }
  }

  async function handleUpload(event: React.ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0];
    if (!file) return;
    setUploading(true);
    try {
      const created = await uploadAsset(file);
      setMessage('上传成功');
      await refreshAssets();
      await handleSelectAsset(created.id);
    } catch (err) {
      setMessage((err as Error).message);
    } finally {
      setUploading(false);
      event.target.value = '';
    }
  }

  async function loadTree(assetId: string) {
    setTreeLoading(true);
    try {
      const nodes = await fetchFileTree(assetId);
      setTree(nodes);
    } catch (err) {
      setMessage((err as Error).message);
    } finally {
      setTreeLoading(false);
    }
  }

  async function handleSelectAsset(id: string) {
    setSelectedAsset(id);
    setSelectedPath(null);
    setFileContent(null);
    setSession(null);
    setCursor(0);
    setErrors(0);
    setElapsedSeconds(0);
    await loadTree(id);
  }

  async function handleSelectFile(path: string) {
    if (!selectedAsset) return;
    setSelectedPath(path);
    try {
      const content = await fetchFileContent(selectedAsset, path);
      const sessionData = await ensureSession(selectedAsset, path);
      setSession(sessionData);
      setCursor(Math.min(sessionData.cursor ?? 0, content.content.length));
      setErrors(sessionData.errors ?? 0);
      setElapsedSeconds(sessionData.durationSeconds ?? 0);
      setFileContent(content);
    } catch (err) {
      setMessage((err as Error).message);
    }
  }

  async function ensureSession(assetId: string, filePath: string) {
    const storageKey = sessionKey(assetId, filePath);
    let sessionData: Session | null = null;
    const existingId = localStorage.getItem(storageKey);
    if (existingId) {
      try {
        sessionData = await fetchSession(existingId);
      } catch {
        localStorage.removeItem(storageKey);
      }
    }
    if (!sessionData) {
      sessionData = await createSession(assetId, filePath);
      localStorage.setItem(storageKey, sessionData.id);
    }
    return sessionData;
  }

  return (
    <div className="app-shell">
      <header className="app-header">
        <div>
          <h1>Code Copy Book</h1>
          <p className="muted">上传源码 → 选择文件 → 键入临摹，逐字加深。</p>
        </div>
        <label className={`upload-button ${uploading ? 'loading' : ''}`}>
          <input type="file" onChange={handleUpload} accept=".zip,.go,.ts,.tsx,.js,.jsx,.py,.java,.rs,.c,.cpp,.cs,.rb,.php,.swift,.kt,.txt" />
          {uploading ? '上传中…' : '上传文件 / 压缩包'}
        </label>
      </header>

      {message && (
        <div className="toast" onClick={() => setMessage(null)}>
          {message}
        </div>
      )}

      <main className="app-main">
        <section className="panel">
          <div className="panel-header">
            <h2>素材</h2>
            {assetLoading && <span className="spinner" />}
          </div>
          <AssetList assets={assets} selectedId={selectedAsset} onSelect={handleSelectAsset} />
          <p className="muted tip">上传文件夹时，请先打包为 .zip。</p>
        </section>

        <section className="panel">
          <div className="panel-header">
            <h2>文件树</h2>
            {treeLoading && <span className="spinner" />}
          </div>
          {selectedAsset ? (
            <FileTree nodes={tree} activePath={selectedPath} onSelect={handleSelectFile} />
          ) : (
            <p className="muted">先选择一个素材，然后展开文件。</p>
          )}
        </section>

        <section className="panel practice">
          <div className="panel-header">
            <h2>临摹区域</h2>
            {fileContent && (
              <div className="stats">
                <span>进度 {progress}%</span>
                <span>准确率 {accuracy}%</span>
                <span>时间 {formatDuration(elapsedSeconds)}</span>
              </div>
            )}
          </div>
          <PracticeCanvas content={fileContent} cursor={cursor} errorFlash={flashError} />
        </section>
      </main>
    </div>
  );
}

function mapKeyToChar(event: KeyboardEvent): string | null {
  if (event.key === 'Enter') {
    return '\n';
  }
  if (event.key === 'Tab') {
    return '\t';
  }
  if (event.key.length === 1) {
    return event.key;
  }
  return null;
}

function sessionKey(assetId: string, path: string) {
  return `ccb:${assetId}:${path}`;
}

function formatDuration(seconds: number) {
  const mins = Math.floor(seconds / 60);
  const secs = seconds % 60;
  return `${String(mins).padStart(2, '0')}:${String(secs).padStart(2, '0')}`;
}

export default App;
