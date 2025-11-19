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

  // 使用 useRef 存储最新状态，避免频繁重新创建 interval
  const cursorRef = useRef(cursor);
  const errorsRef = useRef(errors);
  const elapsedSecondsRef = useRef(elapsedSeconds);

  // 同步状态到 ref
  useEffect(() => {
    cursorRef.current = cursor;
  }, [cursor]);

  useEffect(() => {
    errorsRef.current = errors;
  }, [errors]);

  useEffect(() => {
    elapsedSecondsRef.current = elapsedSeconds;
  }, [elapsedSeconds]);

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
    // 使用 setInterval 每 1.2 秒保存一次进度
    const interval = window.setInterval(() => {
      patchSession(session.id, {
        cursor: cursorRef.current,
        errors: errorsRef.current,
        durationSeconds: Math.round(elapsedSecondsRef.current),
      }).catch((err) => console.warn('session sync failed', err));
    }, 1200);
    return () => window.clearInterval(interval);
  }, [session?.id]); // 只依赖 session，避免频繁重建

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

  async function resetProgress() {
    if (!session || !fileContent) return;
    if (!window.confirm('确定要重置当前文档的进度吗？此操作不可撤销。')) {
      return;
    }
    setCursor(0);
    setErrors(0);
    setElapsedSeconds(0);
    try {
      await patchSession(session.id, {
        cursor: 0,
        errors: 0,
        durationSeconds: 0,
      });
      setMessage('进度已重置');
    } catch (err) {
      setMessage((err as Error).message);
    }
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
      <div className="app-layout">
        {message && (
          <div className="alert floating" onClick={() => setMessage(null)}>
            {message}
          </div>
        )}

        {/* 左侧文档树区域 */}
        <aside className="sidebar">
          <div className="sidebar-header">
            <div className="user-info">
              <p className="eyebrow">当前用户</p>
              <h4>{user.name}</h4>
              <p className="user-email">{user.email}</p>
            </div>
            <div className="import-buttons">
              <button className="back-button" onClick={exitPractice}>
                ← 返回素材库
              </button>
            </div>
          </div>

          <div className="sidebar-content">
            <div className="card">
              <div className="card-header">
                <h3>📚 素材库</h3>
              </div>
              <AssetList assets={assets} selectedId={selectedAsset} onSelect={handleSelectAsset} />
            </div>

            <div className="card">
              <div className="card-header">
                <h3>🗂️ 文件 / 模块</h3>
              </div>
              {selectedAsset ? (
                <FileTree nodes={tree} activePath={selectedPath} onSelect={handleSelectFile} />
              ) : (
                <div className="empty-card">先选择一个素材查看文件结构。</div>
              )}
            </div>
          </div>
        </aside>

        {/* 中间临摹区域 */}
        <main className="practice-workspace">
          <div className="practice-container">
            {/* 固定的状态栏 */}
            <header className="practice-status-bar">
              <div className="status-left">
                <h2>{selectedPath ?? '未选择文件'}</h2>
                {fileContent && <p className="subtitle">语言：{fileContent.language.toUpperCase()}</p>}
              </div>
              <div className="status-right">
                <div className="stat-item">
                  <span>进度</span>
                  <strong>{progress}%</strong>
                </div>
                <div className="stat-item">
                  <span>准确率</span>
                  <strong>{accuracy}%</strong>
                </div>
                <div className="stat-item">
                  <span>用时</span>
                  <strong>{formatDuration(elapsedSeconds)}</strong>
                </div>
                <div className="stat-item">
                  <span>错误</span>
                  <strong>{errors}</strong>
                </div>
              </div>
            </header>

            {/* 可滚动的临摹区域 */}
            <div className="practice-scroll-area">
              <div className="practice-focus">
                <div className="practice-progress">
                  <div className="progress-thumb" style={{ width: `${progress}%` }} />
                </div>
                <PracticeCanvas content={fileContent} cursor={cursor} errorFlash={flashError} />
              </div>
            </div>

            {/* 固定的功能栏 */}
            <footer className="practice-action-bar">
              <div className="action-bar-content">
                <div className="action-left">
                  <span className="action-hint">
                    💡 提示：按 Backspace 可以回到上一个字符 | 估算速度：{computeWPM(cursor, elapsedSeconds)} WPM
                  </span>
                </div>
                <div className="action-right">
                  <button
                    className="action-button skip"
                    onClick={skipCurrentLine}
                    disabled={!canSkipLine}
                  >
                    ⏭️ 跳过当前行
                  </button>
                  <button
                    className="action-button reset"
                    onClick={resetProgress}
                    disabled={!session}
                  >
                    🔄 重置进度
                  </button>
                </div>
              </div>
            </footer>
          </div>
        </main>
      </div>
    );
  }

  return (
    <div className="app-layout">
      {message && (
        <div className="alert" onClick={() => setMessage(null)}>
          {message}
        </div>
      )}

      {/* 左侧文档树区域 */}
      <aside className="sidebar">
        <div className="sidebar-header">
          <div className="user-info">
            <p className="eyebrow">当前用户</p>
            <h4>{user.name}</h4>
            <p className="user-email">{user.email}</p>
          </div>
          <div className="import-buttons">
            <label className={`import-button ${uploading || pasting ? 'loading' : ''}`}>
              <input
                type="file"
                onChange={handleUpload}
                accept=".zip,.go,.ts,.tsx,.js,.jsx,.py,.java,.rs,.c,.cpp,.cs,.rb,.php,.swift,.kt,.txt,.sh,.bash,.yaml,.yml,.json,.md,.toml,.conf,.cfg"
              />
              📁 导入文件 / 压缩包
            </label>
            <button
              className="import-button secondary"
              onClick={() => {
                const filename = window.prompt('请输入文件名:', 'snippet.txt');
                if (filename !== null) {
                  setPasteFilename(filename);
                  // 显示悬浮弹窗
                  const content = window.prompt('请粘贴内容:');
                  if (content !== null) {
                    setPasteContent(content);
                    handlePasteSubmit({ preventDefault: () => {} } as any);
                  }
                }
              }}
              disabled={uploading || pasting}
            >
              📋 粘贴导入
            </button>
          </div>
        </div>

        <div className="sidebar-content">
          <div className="card">
            <div className="card-header">
              <h3>📚 素材库</h3>
              <div className="header-actions">
                {assetLoading && <span className="spinner" />}
                <button className="refresh-button" onClick={() => refreshAssets()} disabled={assetLoading}>
                  🔄
                </button>
                <button className="logout-button" onClick={handleLogout}>🚪</button>
              </div>
            </div>
            <AssetList assets={assets} selectedId={selectedAsset} onSelect={handleSelectAsset} />
            <p className="muted tip">支持单文件或 ZIP，大小建议 &lt; 20MB。</p>
          </div>

          <div className="card">
            <div className="card-header">
              <h3>🗂️ 文件 / 模块</h3>
              {treeLoading && <span className="spinner" />}
            </div>
            {selectedAsset ? (
              <FileTree nodes={tree} activePath={selectedPath} onSelect={handleSelectFile} />
            ) : (
              <div className="empty-card">先选择一个素材查看文件结构。</div>
            )}
          </div>
        </div>
      </aside>

      {/* 中间临摹区域 */}
      <main className="practice-workspace">
        <div className="practice-container">
          {/* 固定的状态栏 */}
          <header className="practice-status-bar">
            <div className="status-left">
              <h2>{selectedPath ?? '等待选择文件'}</h2>
              {fileContent && <p className="subtitle">语言：{fileContent.language.toUpperCase()}</p>}
            </div>
            <div className="status-right">
              <div className="stat-item">
                <span>进度</span>
                <strong>{progress}%</strong>
              </div>
              <div className="stat-item">
                <span>准确率</span>
                <strong>{accuracy}%</strong>
              </div>
              <div className="stat-item">
                <span>用时</span>
                <strong>{formatDuration(elapsedSeconds)}</strong>
              </div>
              {selectedPath && (
                <button className="start-practice-button" onClick={() => setView('practice')} disabled={!fileContent}>
                  进入临摹 →
                </button>
              )}
            </div>
          </header>

          {/* 可滚动的临摹区域 */}
          <div className="practice-scroll-area">
            {selectedPath ? (
              <div className="practice-preview">
                <p className="preview-text">准备就绪，点击上方"进入临摹"按钮开始练习。</p>
                {fileContent && <p className="preview-info">文件大小：{fileContent.content.length} 字符</p>}
              </div>
            ) : (
              <div className="practice-placeholder">
                <p>选择一个文件开始临摹。</p>
              </div>
            )}
          </div>

          {/* 固定的功能栏 */}
          <footer className="practice-action-bar">
            <div className="action-bar-content">
              <div className="action-left">
                <span className="action-hint">💡 提示：使用键盘输入进行临摹，按 Backspace 可回退</span>
              </div>
              <div className="action-right">
                <button
                  className="action-button skip"
                  onClick={skipCurrentLine}
                  disabled={!fileContent || !canSkipLine}
                >
                  ⏭️ 跳过当前行
                </button>
                <button
                  className="action-button reset"
                  onClick={resetProgress}
                  disabled={!session}
                >
                  🔄 重置进度
                </button>
              </div>
            </div>
          </footer>
        </div>
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
