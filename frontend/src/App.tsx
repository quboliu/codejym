import { useEffect, useMemo, useRef, useState } from 'react';
import type { FormEvent } from 'react';
import './App.css';
import { AssetList } from './components/AssetList';
import { FileTree } from './components/FileTree';
import { PracticeCanvas } from './components/PracticeCanvas';
import {
  createSession,
  fetchCurrentUser,
  fetchFileContent,
  fetchFileTree,
  fetchSession,
  listAssets,
  login,
  patchSession,
  setAuthToken,
  signup,
  uploadAsset,
  uploadPastedAsset,
} from './api';
import type { Asset, FileNode, FileContent, Session, User } from './types';

type View = 'dashboard' | 'practice';

const AUTH_TOKEN_KEY = 'codecopybook_token';

function App() {
  const [user, setUser] = useState<User | null>(null);
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
  const [pasting, setPasting] = useState(false);
  const [pasteFilename, setPasteFilename] = useState('');
  const [pasteContent, setPasteContent] = useState('');
  const [authMode, setAuthMode] = useState<'login' | 'signup'>('login');
  const [authEmail, setAuthEmail] = useState('');
  const [authPassword, setAuthPassword] = useState('');
  const [authName, setAuthName] = useState('');
  const [authLoading, setAuthLoading] = useState(false);
  const [flashError, setFlashError] = useState(false);
  const [view, setView] = useState<View>('dashboard');
  const errorTimer = useRef<number | null>(null);

  useEffect(() => {
    const stored = localStorage.getItem(AUTH_TOKEN_KEY);
    if (!stored) {
      return;
    }
    setAuthToken(stored);
    fetchCurrentUser()
      .then((current) => {
        setUser(current);
      })
      .catch(() => {
        localStorage.removeItem(AUTH_TOKEN_KEY);
        setAuthToken(null);
      });
  }, []);

  useEffect(() => {
    if (!user) {
      setAssets([]);
      setSelectedAsset(null);
      setTree([]);
      setSelectedPath(null);
      setFileContent(null);
      setSession(null);
      setCursor(0);
      setErrors(0);
      setElapsedSeconds(0);
      setView('dashboard');
      return;
    }
    refreshAssets(user);
  }, [user?.id]);

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
      if (view !== 'practice') return;
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
  }, [fileContent, cursor, view]);

  const progress = useMemo(() => {
    if (!fileContent) return 0;
    if (fileContent.content.length === 0) return 0;
    return Math.round((cursor / fileContent.content.length) * 100);
  }, [cursor, fileContent]);

  const accuracy = useMemo(() => {
    if (cursor + errors === 0) return 100;
    return Math.max(0, Math.round((cursor / (cursor + errors)) * 100));
  }, [cursor, errors]);

  const canSkipLine = !!(fileContent && cursor < fileContent.content.length);

  async function refreshAssets(targetUser?: User | null) {
    const activeUser = targetUser ?? user;
    if (!activeUser) {
      return;
    }
    setAssetLoading(true);
    try {
      const data = await listAssets();
      setAssets(data);
      if (data.length && !selectedAsset) {
        const first = data[0];
        setSelectedAsset(first.id);
        await loadTree(first.id);
      }
    } catch (err) {
      setMessage((err as Error).message);
    } finally {
      setAssetLoading(false);
    }
  }

  async function handleAuthSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setAuthLoading(true);
    try {
      const email = authEmail.trim();
      const password = authPassword;
      if (!email || !password) {
        setMessage('请输入邮箱和密码');
        return;
      }
      const name = authName.trim();
      if (authMode === 'signup' && !name) {
        setMessage('请输入昵称');
        return;
      }
      const response =
        authMode === 'login' ? await login(email, password) : await signup(email, password, name);
      localStorage.setItem(AUTH_TOKEN_KEY, response.token);
      setAuthToken(response.token);
      setUser(response.user);
      setAuthPassword('');
      if (authMode === 'signup') {
        setAuthName('');
      }
      setMessage(authMode === 'login' ? '登录成功' : '注册成功，已自动登录');
    } catch (err) {
      setMessage((err as Error).message);
    } finally {
      setAuthLoading(false);
    }
  }

  function handleLogout() {
    localStorage.removeItem(AUTH_TOKEN_KEY);
    setAuthToken(null);
    setUser(null);
    setAssets([]);
    setSelectedAsset(null);
    setTree([]);
    setSelectedPath(null);
    setFileContent(null);
    setSession(null);
    setCursor(0);
    setErrors(0);
    setElapsedSeconds(0);
    setMessage('已退出登录');
  }

  async function handleUpload(event: React.ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0];
    if (!file) return;
    if (!user) {
      setMessage('请先登录');
      return;
    }
    setUploading(true);
    try {
      const created = await uploadAsset(file);
      setMessage('上传成功');
      await refreshAssets(user);
      await handleSelectAsset(created.id);
    } catch (err) {
      setMessage((err as Error).message);
    } finally {
      setUploading(false);
      event.target.value = '';
    }
  }

  async function handlePasteSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!user) {
      setMessage('请先登录');
      return;
    }
    if (!pasteContent.trim()) {
      setMessage('请粘贴一些内容再上传');
      return;
    }
    const filename = pasteFilename.trim() || 'snippet.txt';
    setPasting(true);
    try {
      const created = await uploadPastedAsset(filename, pasteContent);
      setMessage('粘贴内容已保存');
      setPasteContent('');
      setPasteFilename('');
      await refreshAssets(user);
      await handleSelectAsset(created.id);
    } catch (err) {
      setMessage((err as Error).message);
    } finally {
      setPasting(false);
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
    if (!user) return;
    setSelectedAsset(id);
    setSelectedPath(null);
    setFileContent(null);
    setSession(null);
    setCursor(0);
    setErrors(0);
    setElapsedSeconds(0);
    setView('dashboard');
    await loadTree(id);
  }

  async function handleSelectFile(path: string) {
    if (!selectedAsset || !user) return;
    setSelectedPath(path);
    try {
      const content = await fetchFileContent(selectedAsset, path);
      const sessionData = await ensureSession(selectedAsset, path);
      setSession(sessionData);
      setCursor(Math.min(sessionData.cursor ?? 0, content.content.length));
      setErrors(sessionData.errors ?? 0);
      setElapsedSeconds(sessionData.durationSeconds ?? 0);
      setFileContent(content);
      setView('practice');
    } catch (err) {
      setMessage((err as Error).message);
    }
  }

  async function ensureSession(assetId: string, filePath: string) {
    const storageKey = sessionKey(user?.id ?? 'anon', assetId, filePath);
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

  function exitPractice() {
    setView('dashboard');
  }

  function skipCurrentLine() {
    if (!fileContent) return;
    if (cursor >= fileContent.content.length) {
      return;
    }
    const newlineIndex = fileContent.content.indexOf('\n', cursor);
    const nextCursor = newlineIndex === -1 ? fileContent.content.length : newlineIndex + 1;
    setCursor(nextCursor);
  }

  if (!user) {
    return (
      <div className="auth-page">
        <div className="auth-card">
          <h1>代码临摹工作室</h1>
          <p className="auth-subtitle">登录后即可上传素材并同步练习进度。</p>
          {message && (
            <div className="alert" onClick={() => setMessage(null)}>
              {message}
            </div>
          )}
          <form className="auth-form" onSubmit={handleAuthSubmit}>
            <input
              type="email"
              placeholder="邮箱"
              value={authEmail}
              onChange={(e) => setAuthEmail(e.target.value)}
              required
            />
            {authMode === 'signup' && (
              <input
                type="text"
                placeholder="昵称 / 显示名称"
                value={authName}
                onChange={(e) => setAuthName(e.target.value)}
                required
              />
            )}
            <input
              type="password"
              placeholder="密码"
              value={authPassword}
              onChange={(e) => setAuthPassword(e.target.value)}
              required
            />
            <button type="submit" className="primary" disabled={authLoading}>
              {authLoading ? '处理中…' : authMode === 'login' ? '登录' : '注册并登录'}
            </button>
          </form>
          <button
            type="button"
            className="ghost-link"
            onClick={() => {
              setAuthMode((prev) => (prev === 'login' ? 'signup' : 'login'));
              setMessage(null);
            }}
          >
            {authMode === 'login' ? '没有账户？注册一个' : '已有账户？直接登录'}
          </button>
        </div>
      </div>
    );
  }

  if (view === 'practice') {
    return (
      <div className="practice-page">
        <header className="practice-topbar">
          <div className="topbar-left">
            <button className="ghost" onClick={exitPractice}>
              ← 返回素材库
            </button>
            <div>
              <p className="eyebrow">当前文件</p>
              <h2>{selectedPath ?? '未选择文件'}</h2>
              {fileContent && <p className="subtitle">语言：{fileContent.language.toUpperCase()}</p>}
            </div>
          </div>
          <div className="topbar-stats">
            <div>
              <span>进度</span>
              <strong>{progress}%</strong>
            </div>
            <div>
              <span>准确率</span>
              <strong>{accuracy}%</strong>
            </div>
            <div>
              <span>用时</span>
              <strong>{formatDuration(elapsedSeconds)}</strong>
            </div>
          </div>
        </header>
        {message && (
          <div className="alert floating" onClick={() => setMessage(null)}>
            {message}
          </div>
        )}
        <div className="practice-main">
          <div className="practice-focus">
            <div className="practice-toolbar">
              <p className="toolbar-hint">光标实时标记当前位置，如需跳过整行可使用该按钮。</p>
              <button className="skip-line-button" onClick={skipCurrentLine} disabled={!canSkipLine}>
                跳过当前行
              </button>
            </div>
            <div className="practice-progress">
              <div className="progress-thumb" style={{ width: `${progress}%` }} />
            </div>
            <PracticeCanvas content={fileContent} cursor={cursor} errorFlash={flashError} />
          </div>
          <aside className="practice-side">
            <div className="side-card">
              <h4>会话信息</h4>
              <p>错误次数：{errors}</p>
              <p>估算 WPM：{computeWPM(cursor, elapsedSeconds)}</p>
              <p>Session ID：{session?.id}</p>
            </div>
            <div className="side-card">
              <h4>提示</h4>
              <ul>
                <li>按 Backspace 可以回到上一个字符。</li>
                <li>完成所有字符后自动保持 100% 进度。</li>
                <li>每 1.2 秒自动同步进度，刷新不会丢失。</li>
              </ul>
            </div>
          </aside>
        </div>
        <button
          className="skip-line-fab"
          onClick={skipCurrentLine}
          disabled={!canSkipLine}
          aria-label="跳过当前行"
        >
          跳过当前行
        </button>
      </div>
    );
  }

  return (
    <div className="page">
      <header className="hero">
        <div>
          <p className="eyebrow">代码临摹工作室</p>
          <h1>挑选最爱的源码，逐字临摹</h1>
          <p className="subtitle">上传文件或项目压缩包，选择文件后在右侧工作区跟随浅色字帖逐个字符键入。</p>
          <div className="hero-actions">
            <label className={`upload-button ${uploading || pasting ? 'loading' : ''}`}>
              <input
                type="file"
                onChange={handleUpload}
                accept=".zip,.go,.ts,.tsx,.js,.jsx,.py,.java,.rs,.c,.cpp,.cs,.rb,.php,.swift,.kt,.txt,.sh,.bash,.yaml,.yml,.json,.md,.toml,.conf,.cfg"
              />
              {uploading || pasting ? '上传中…' : '上传文件 / 压缩包'}
            </label>
            <button className="secondary" onClick={() => refreshAssets()} disabled={assetLoading}>
              {assetLoading ? '同步中…' : '刷新素材'}
            </button>
            <button className="secondary logout" onClick={handleLogout}>
              退出登录
            </button>
          </div>
          <div className="user-pill">
            <div>
              <p className="eyebrow">当前用户</p>
              <h4>{user.name}</h4>
              <p className="user-email">{user.email}</p>
            </div>
          </div>
          <div className="paste-upload">
            <div>
              <p className="eyebrow">快速粘贴</p>
              <p className="subtitle">没有文件？直接输入文件名并粘贴文本，立即生成素材。</p>
            </div>
            <form className="paste-upload-form" onSubmit={handlePasteSubmit}>
              <input
                type="text"
                placeholder="文件名（例如 script.sh）"
                value={pasteFilename}
                onChange={(e) => setPasteFilename(e.target.value)}
                disabled={uploading || pasting}
              />
              <textarea
                placeholder="在这里粘贴内容..."
                value={pasteContent}
                onChange={(e) => setPasteContent(e.target.value)}
                disabled={uploading || pasting}
              />
              <div className="paste-upload-actions">
                <button
                  type="submit"
                  className="primary"
                  disabled={pasting || uploading || !pasteContent.trim()}
                >
                  {pasting ? '生成中…' : '粘贴生成素材'}
                </button>
              </div>
            </form>
          </div>
        </div>
        <div className="hero-card">
          <p>当前状态</p>
          <h3>{fileContent ? selectedPath : '尚未选择文件'}</h3>
          <ul>
            <li>进度 <strong>{progress}%</strong></li>
            <li>准确率 <strong>{accuracy}%</strong></li>
            <li>用时 <strong>{formatDuration(elapsedSeconds)}</strong></li>
          </ul>
        </div>
      </header>

      {message && (
        <div className="alert" onClick={() => setMessage(null)}>
          {message}
        </div>
      )}

      <section className="stats-row">
        <div className="stat-card">
          <span>当前进度</span>
          <strong>{progress}%</strong>
          <div className="progress-track">
            <div className="progress-thumb" style={{ width: `${progress}%` }} />
          </div>
        </div>
        <div className="stat-card">
          <span>准确率</span>
          <strong>{accuracy}%</strong>
          <p className="stat-detail">错误次数 {errors}</p>
        </div>
        <div className="stat-card">
          <span>键入时间</span>
          <strong>{formatDuration(elapsedSeconds)}</strong>
          <p className="stat-detail">约 {computeWPM(cursor, elapsedSeconds)} WPM</p>
        </div>
      </section>

      <div className="workspace">
        <aside className="sidebar">
          <div className="card">
            <div className="card-header">
              <h3>素材库</h3>
              {assetLoading && <span className="spinner" />}
            </div>
            <AssetList assets={assets} selectedId={selectedAsset} onSelect={handleSelectAsset} />
            <p className="muted tip">支持单文件或 ZIP，大小建议 &lt; 20MB。</p>
          </div>

          <div className="card">
            <div className="card-header">
              <h3>文件 / 模块</h3>
              {treeLoading && <span className="spinner" />}
            </div>
            {selectedAsset ? (
              <FileTree nodes={tree} activePath={selectedPath} onSelect={handleSelectFile} />
            ) : (
              <div className="empty-card">先选择一个素材查看文件结构。</div>
            )}
          </div>
        </aside>

        <section className="practice-area placeholder">
          <div className="practice-head">
            <div>
              <p className="eyebrow">临摹区域</p>
              <h2>{selectedPath ?? '等待选择文件'}</h2>
              {fileContent && <p className="subtitle">语言：{fileContent.language.toUpperCase()}</p>}
            </div>
            <div className="session-meta">
              <span>进度 {progress}%</span>
              <span>准确率 {accuracy}%</span>
              <span>时间 {formatDuration(elapsedSeconds)}</span>
            </div>
          </div>
          <div className="practice-placeholder">
            {selectedPath ? (
              <div>
                <p>准备就绪，点击下方按钮进入临摹页面。</p>
                <button className="primary" onClick={() => setView('practice')} disabled={!fileContent}>
                  进入临摹
                </button>
              </div>
            ) : (
              <p>选择一个文件开始临摹。</p>
            )}
          </div>
        </section>
      </div>
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

function sessionKey(userId: string | null, assetId: string, path: string) {
  return `ccb:${userId ?? 'anon'}:${assetId}:${path}`;
}

function formatDuration(seconds: number) {
  const mins = Math.floor(seconds / 60);
  const secs = seconds % 60;
  return `${String(mins).padStart(2, '0')}:${String(secs).padStart(2, '0')}`;
}

function computeWPM(chars: number, seconds: number) {
  if (seconds === 0) return 0;
  const words = chars / 5;
  return Math.max(0, Math.round((words / seconds) * 60));
}

export default App;
