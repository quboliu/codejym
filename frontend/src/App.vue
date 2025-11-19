<template>
  <div id="app">
    <!-- Toast 通知 -->
    <Toast ref="toastComponentRef" />

    <!-- 登录页 -->
    <div v-if="!user" class="auth-page">
      <div class="auth-container">
        <div class="auth-header">
          <h1>CodeJYM</h1>
          <p class="text-secondary">代码临摹，刻意练习</p>
        </div>

        <form class="auth-form" @submit.prevent="handleAuthSubmit">
          <div class="form-group">
            <label for="email">邮箱</label>
            <input
              id="email"
              type="email"
              class="input"
              v-model="authEmail"
              placeholder="your@email.com"
              required
              autocomplete="email"
            />
          </div>

          <div v-if="authMode === 'signup'" class="form-group">
            <label for="name">昵称</label>
            <input
              id="name"
              type="text"
              class="input"
              v-model="authName"
              placeholder="你的昵称"
              required
              autocomplete="name"
            />
          </div>

          <div class="form-group">
            <label for="password">密码</label>
            <input
              id="password"
              type="password"
              class="input"
              v-model="authPassword"
              placeholder="••••••••"
              required
              autocomplete="current-password"
            />
          </div>

          <button type="submit" class="btn btn-primary btn-lg" :disabled="authLoading" style="width: 100%">
            {{ authLoading ? '处理中…' : authMode === 'login' ? '登录' : '注册' }}
          </button>
        </form>

        <div class="auth-footer">
          <button type="button" class="btn-link" @click="toggleAuthMode">
            {{ authMode === 'login' ? '还没有账号？立即注册' : '已有账号？返回登录' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 主应用 -->
    <div v-else class="app-layout">
      <!-- 顶部导航栏 -->
      <header class="app-header">
        <div class="header-left">
          <button class="btn-icon" @click="sidebarCollapsed = !sidebarCollapsed" aria-label="切换侧边栏">
            <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
              <path d="M3 7H17M3 13H17"/>
            </svg>
          </button>
          <h1 class="app-title">CodeJYM</h1>
        </div>

        <div class="header-right">
          <!-- 主题切换按钮 -->
          <button class="btn-icon" @click="handleToggleTheme" :title="`当前主题: ${theme}`">
            <svg v-if="isDark" width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M10 3V2M10 18V17M17 10H18M2 10H3M15.5 4.5L16.2 3.8M3.8 16.2L4.5 15.5M15.5 15.5L16.2 16.2M3.8 3.8L4.5 4.5" stroke-linecap="round"/>
              <circle cx="10" cy="10" r="4"/>
            </svg>
            <svg v-else width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M17 10.5C16.1 13.9 12.9 16.4 9.2 16.4C5.5 16.4 2.5 13.4 2.5 9.7C2.5 6 5 2.8 8.4 1.9C8.1 2.6 8 3.3 8 4C8 7.3 10.7 10 14 10C14.7 10 15.4 9.9 16.1 9.6L17 10.5Z" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
          </button>

          <!-- 用户菜单 -->
          <div class="user-menu">
            <button class="btn-icon">
              <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="10" cy="7" r="3.5"/>
                <path d="M4 17C4 13.7 6.7 11 10 11C13.3 11 16 13.7 16 17" stroke-linecap="round"/>
              </svg>
            </button>
            <div class="user-dropdown">
              <div class="user-info">
                <div class="user-name">{{ user.name }}</div>
                <div class="user-email">{{ user.email }}</div>
              </div>
              <div class="dropdown-divider"></div>
              <button class="dropdown-item" @click="handleLogout">
                <svg width="16" height="16" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M8 17H3V3H8M13 13L17 9L13 5M17 9H8" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
                退出登录
              </button>
            </div>
          </div>
        </div>
      </header>

      <div class="app-content">
        <!-- 侧边栏 -->
        <aside class="sidebar" :class="{ collapsed: sidebarCollapsed }">
          <div class="sidebar-section">
            <div class="section-header">
              <h3 class="section-title">导入素材</h3>
            </div>
            <div class="import-actions">
              <label class="btn btn-primary" style="width: 100%">
                <input
                  type="file"
                  style="display: none"
                  @change="handleUpload"
                  accept=".zip,.go,.ts,.tsx,.js,.jsx,.py,.java,.rs,.c,.cpp,.cs,.rb,.php,.swift,.kt,.txt,.sh,.bash,.yaml,.yml,.json,.md,.toml,.conf,.cfg"
                />
                <svg width="16" height="16" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M10 3V13M6 9L10 13L14 9" stroke-linecap="round" stroke-linejoin="round"/>
                  <path d="M17 13V16C17 16.5 16.5 17 16 17H4C3.5 17 3 16.5 3 16V13" stroke-linecap="round"/>
                </svg>
                上传文件
              </label>
              <button class="btn btn-secondary" style="width: 100%" @click="showPasteModal = true" :disabled="uploading || pasting">
                <svg width="16" height="16" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M7 3H5C4 3 3 4 3 5V16C3 17 4 18 5 18H15C16 18 17 17 17 16V5C17 4 16 3 15 3H13" stroke-linecap="round" stroke-linejoin="round"/>
                  <path d="M7 3C7 2 8 2 8 2H12C12 2 13 2 13 3V4C13 5 12 5 12 5H8C8 5 7 5 7 4V3Z" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
                粘贴代码
              </button>
            </div>
          </div>

          <div class="sidebar-section">
            <div class="section-header">
              <h3 class="section-title">素材库</h3>
              <button class="btn-icon btn-icon-sm" @click="refreshAssets" :disabled="assetLoading">
                <svg width="16" height="16" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="2" :class="{ 'animate-spin': assetLoading }">
                  <path d="M3 10C3 6.1 6.1 3 10 3C13.9 3 17 6.1 17 10M17 10L14 7M17 10L20 7M17 10C17 13.9 13.9 17 10 17C6.1 17 3 13.9 3 10M3 10L6 13M3 10L0 13" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
              </button>
            </div>
            <AssetList
              :assets="assets"
              :selected-id="selectedAsset"
              @select="handleSelectAsset"
            />
          </div>

          <div v-if="selectedAsset" class="sidebar-section">
            <div class="section-header">
              <h3 class="section-title">文件列表</h3>
            </div>
            <FileTree
              :nodes="tree"
              :active-path="selectedPath"
              @select="handleSelectFile"
            />
          </div>
        </aside>

        <!-- 主工作区 -->
        <main class="workspace">
          <div v-if="!fileContent" class="workspace-empty">
            <div class="empty-icon">
              <svg width="64" height="64" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5">
                <path d="M3 3H11L15 7V17H3V3Z" stroke-linecap="round" stroke-linejoin="round"/>
                <path d="M11 3V7H15" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
            </div>
            <h3>选择一个文件开始练习</h3>
            <p class="text-secondary">从左侧选择一个素材和文件</p>
          </div>

          <div v-else class="workspace-content">
            <!-- 进度栏 -->
            <div class="progress-bar">
              <div class="progress-fill" :style="{ width: progress + '%' }"></div>
            </div>

            <!-- 统计栏 -->
            <div class="stats-bar">
              <div class="stat">
                <span class="stat-label">进度</span>
                <span class="stat-value">{{ progress }}%</span>
              </div>
              <div class="stat">
                <span class="stat-label">准确率</span>
                <span class="stat-value">{{ accuracy }}%</span>
              </div>
              <div class="stat">
                <span class="stat-label">用时</span>
                <span class="stat-value">{{ formatDuration(elapsedSeconds) }}</span>
              </div>
              <div class="stat">
                <span class="stat-label">速度</span>
                <span class="stat-value">{{ computeWPM(cursor, elapsedSeconds) }} WPM</span>
              </div>
              <div class="stat">
                <span class="stat-label">错误</span>
                <span class="stat-value" :class="{ 'text-error': errors > 0 }">{{ errors }}</span>
              </div>
            </div>

            <!-- 临摹区域 -->
            <div class="practice-area">
              <PracticeCanvas
                :content="fileContent"
                :cursor="cursor"
                :error-flash="flashError"
              />
            </div>

            <!-- 操作栏 -->
            <div class="action-bar">
              <div class="action-hint">
                <kbd>Backspace</kbd> 回退
                <span class="divider">|</span>
                <kbd>?</kbd> 帮助
              </div>
              <div class="action-buttons">
                <button class="btn btn-ghost btn-sm" @click="skipCurrentLine" :disabled="!canSkipLine">
                  跳过当前行
                </button>
                <button class="btn btn-ghost btn-sm" @click="handleResetProgress" :disabled="!session">
                  重置进度
                </button>
              </div>
            </div>
          </div>
        </main>
      </div>
    </div>

    <!-- 粘贴代码 Modal -->
    <Modal
      v-model="showPasteModal"
      title="粘贴代码"
      @confirm="handlePasteSubmit"
      @cancel="resetPasteForm"
      :confirm-disabled="!pasteContent.trim()"
    >
      <div class="form-group">
        <label for="paste-filename">文件名</label>
        <input
          id="paste-filename"
          type="text"
          class="input"
          v-model="pasteFilename"
          placeholder="example.js"
        />
      </div>
      <div class="form-group">
        <label for="paste-content">代码内容</label>
        <textarea
          id="paste-content"
          class="textarea"
          v-model="pasteContent"
          placeholder="粘贴你的代码..."
          rows="10"
        ></textarea>
      </div>
    </Modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import AssetList from './components/AssetList.vue'
import FileTree from './components/FileTree.vue'
import PracticeCanvas from './components/PracticeCanvas.vue'
import Modal from './components/common/Modal.vue'
import Toast from './components/common/Toast.vue'
import { useToast, toastRef as globalToastRef } from './composables/useToast'
import { useTheme } from './composables/useTheme'
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
} from './api'
import type { Asset, FileNode, FileContent, Session, User } from './types'

const AUTH_TOKEN_KEY = 'codecopybook_token'

// Composables
const toast = useToast()
const { theme, isDark, toggleTheme } = useTheme()

// 主题切换
function handleToggleTheme() {
  toggleTheme()
}

// 认证状态
const user = ref<User | null>(null)
const authMode = ref<'login' | 'signup'>('login')
const authEmail = ref('')
const authPassword = ref('')
const authName = ref('')
const authLoading = ref(false)

// UI状态
const sidebarCollapsed = ref(false)
const showPasteModal = ref(false)

// 数据状态
const assets = ref<Asset[]>([])
const assetLoading = ref(false)
const selectedAsset = ref<string | null>(null)
const tree = ref<FileNode[]>([])
const treeLoading = ref(false)
const selectedPath = ref<string | null>(null)
const fileContent = ref<FileContent | null>(null)
const session = ref<Session | null>(null)

// 练习状态
const cursor = ref(0)
const errors = ref(0)
const elapsedSeconds = ref(0)
const flashError = ref(false)
const uploading = ref(false)
const pasting = ref(false)
const pasteFilename = ref('')
const pasteContent = ref('')

// 计算属性
const progress = computed(() => {
  if (!fileContent.value) return 0
  if (fileContent.value.content.length === 0) return 0
  return Math.round((cursor.value / fileContent.value.content.length) * 100)
})

const accuracy = computed(() => {
  if (cursor.value + errors.value === 0) return 100
  return Math.max(0, Math.round((cursor.value / (cursor.value + errors.value)) * 100))
})

const canSkipLine = computed(() => {
  return !!(fileContent.value && cursor.value < fileContent.value.content.length)
})

// Toast组件引用
const toastComponentRef = ref<InstanceType<typeof Toast> | null>(null)

// 生命周期
onMounted(async () => {
  // 设置Toast全局引用
  globalToastRef.value = toastComponentRef.value

  // 尝试恢复登录状态
  const stored = localStorage.getItem(AUTH_TOKEN_KEY)
  if (stored) {
    setAuthToken(stored)
    try {
      user.value = await fetchCurrentUser()
    } catch {
      localStorage.removeItem(AUTH_TOKEN_KEY)
      setAuthToken(null)
    }
  }

  // 监听键盘事件
  window.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
  if (elapsedTimer) clearInterval(elapsedTimer)
  if (sessionTimer) clearInterval(sessionTimer)
})

// 监听用户变化
watch(user, (newUser) => {
  if (newUser) {
    refreshAssets()
  } else {
    resetState()
  }
})

// 监听session和fileContent变化，启动计时器
let elapsedTimer: number | null = null
let sessionTimer: number | null = null

watch([() => session.value, () => fileContent.value], () => {
  if (session.value && fileContent.value) {
    // 启动计时器
    if (elapsedTimer) clearInterval(elapsedTimer)
    elapsedTimer = window.setInterval(() => {
      elapsedSeconds.value += 1
    }, 1000)

    // 启动进度同步
    if (sessionTimer) clearInterval(sessionTimer)
    sessionTimer = window.setInterval(() => {
      if (session.value) {
        patchSession(session.value.id, {
          cursor: cursor.value,
          errors: errors.value,
          durationSeconds: Math.round(elapsedSeconds.value),
        }).catch(err => console.warn('session sync failed', err))
      }
    }, 1200)
  }
})

// 认证方法
function toggleAuthMode() {
  authMode.value = authMode.value === 'login' ? 'signup' : 'login'
}

async function handleAuthSubmit() {
  authLoading.value = true
  try {
    const email = authEmail.value.trim()
    const password = authPassword.value
    const name = authName.value.trim()

    if (!email || !password) {
      toast.error('请输入邮箱和密码')
      return
    }

    if (authMode.value === 'signup' && !name) {
      toast.error('请输入昵称')
      return
    }

    const response = authMode.value === 'login'
      ? await login(email, password)
      : await signup(email, password, name)

    localStorage.setItem(AUTH_TOKEN_KEY, response.token)
    setAuthToken(response.token)
    user.value = response.user

    toast.success(authMode.value === 'login' ? '登录成功' : '注册成功')

    // 清空表单
    authPassword.value = ''
    if (authMode.value === 'signup') {
      authName.value = ''
    }
  } catch (err) {
    toast.error((err as Error).message)
  } finally {
    authLoading.value = false
  }
}

function handleLogout() {
  localStorage.removeItem(AUTH_TOKEN_KEY)
  setAuthToken(null)
  user.value = null
  toast.info('已退出登录')
}

function resetState() {
  assets.value = []
  selectedAsset.value = null
  tree.value = []
  selectedPath.value = null
  fileContent.value = null
  session.value = null
  cursor.value = 0
  errors.value = 0
  elapsedSeconds.value = 0
}

// 素材管理
async function refreshAssets() {
  if (!user.value) return
  assetLoading.value = true
  try {
    const data = await listAssets()
    assets.value = data
    if (data.length && !selectedAsset.value) {
      const first = data[0]
      await handleSelectAsset(first.id)
    }
  } catch (err) {
    toast.error((err as Error).message)
  } finally {
    assetLoading.value = false
  }
}

async function handleUpload(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  uploading.value = true
  try {
    const created = await uploadAsset(file)
    toast.success('上传成功')
    await refreshAssets()
    await handleSelectAsset(created.id)
  } catch (err) {
    toast.error((err as Error).message)
  } finally {
    uploading.value = false
    target.value = ''
  }
}

function resetPasteForm() {
  pasteFilename.value = ''
  pasteContent.value = ''
}

async function handlePasteSubmit() {
  if (!user.value) {
    toast.error('请先登录')
    return
  }

  if (!pasteContent.value.trim()) {
    toast.error('请输入代码内容')
    return
  }

  const filename = pasteFilename.value.trim() || 'snippet.txt'

  pasting.value = true
  try {
    const created = await uploadPastedAsset(filename, pasteContent.value)
    toast.success('代码已保存')
    showPasteModal.value = false
    resetPasteForm()
    await refreshAssets()
    await handleSelectAsset(created.id)
  } catch (err) {
    toast.error((err as Error).message)
  } finally {
    pasting.value = false
  }
}

async function handleSelectAsset(id: string) {
  if (!user.value) return
  selectedAsset.value = id
  selectedPath.value = null
  fileContent.value = null
  session.value = null
  cursor.value = 0
  errors.value = 0
  elapsedSeconds.value = 0

  treeLoading.value = true
  try {
    const nodes = await fetchFileTree(id)
    tree.value = nodes
  } catch (err) {
    toast.error((err as Error).message)
  } finally {
    treeLoading.value = false
  }
}

async function handleSelectFile(path: string) {
  if (!selectedAsset.value || !user.value) return
  selectedPath.value = path

  try {
    const content = await fetchFileContent(selectedAsset.value, path)
    const sessionData = await ensureSession(selectedAsset.value, path)

    session.value = sessionData
    cursor.value = Math.min(sessionData.cursor ?? 0, content.content.length)
    errors.value = sessionData.errors ?? 0
    elapsedSeconds.value = sessionData.durationSeconds ?? 0
    fileContent.value = content
  } catch (err) {
    toast.error((err as Error).message)
  }
}

async function ensureSession(assetId: string, filePath: string) {
  const storageKey = sessionKey(user.value?.id ?? 'anon', assetId, filePath)
  let sessionData: Session | null = null

  const existingId = localStorage.getItem(storageKey)
  if (existingId) {
    try {
      sessionData = await fetchSession(existingId)
    } catch {
      localStorage.removeItem(storageKey)
    }
  }

  if (!sessionData) {
    sessionData = await createSession(assetId, filePath)
    localStorage.setItem(storageKey, sessionData.id)
  }

  return sessionData
}

// 练习控制
function skipCurrentLine() {
  if (!fileContent.value) return
  if (cursor.value >= fileContent.value.content.length) return

  const newlineIndex = fileContent.value.content.indexOf('\n', cursor.value)
  const nextCursor = newlineIndex === -1 ? fileContent.value.content.length : newlineIndex + 1
  cursor.value = nextCursor
}

async function handleResetProgress() {
  if (!session.value || !fileContent.value) return

  // 使用原生confirm（后续可以改成自定义Modal）
  if (!window.confirm('确定要重置当前文档的进度吗？此操作不可撤销。')) {
    return
  }

  cursor.value = 0
  errors.value = 0
  elapsedSeconds.value = 0

  try {
    await patchSession(session.value.id, {
      cursor: 0,
      errors: 0,
      durationSeconds: 0,
    })
    toast.success('进度已重置')
  } catch (err) {
    toast.error((err as Error).message)
  }
}

// 键盘事件处理
function handleKeydown(event: KeyboardEvent) {
  if (!fileContent.value) return
  if (['INPUT', 'TEXTAREA'].includes(document.activeElement?.tagName ?? '')) return
  if (event.metaKey || event.ctrlKey || event.altKey) return

  if (event.key === 'Backspace') {
    event.preventDefault()
    cursor.value = Math.max(0, cursor.value - 1)
    return
  }

  if (cursor.value >= fileContent.value.content.length) return

  const char = mapKeyToChar(event)
  if (char === null) return

  event.preventDefault()
  const expected = fileContent.value.content.charAt(cursor.value)

  if (expected === char) {
    cursor.value = Math.min(fileContent.value.content.length, cursor.value + 1)
  } else {
    errors.value += 1
    flashError.value = true
    setTimeout(() => {
      flashError.value = false
    }, 200)
  }
}

function mapKeyToChar(event: KeyboardEvent): string | null {
  if (event.key === 'Enter') return '\n'
  if (event.key === 'Tab') return '\t'
  if (event.key.length === 1) return event.key
  return null
}

// 工具函数
function sessionKey(userId: string | null, assetId: string, path: string) {
  return `ccb:${userId ?? 'anon'}:${assetId}:${path}`
}

function formatDuration(seconds: number) {
  const mins = Math.floor(seconds / 60)
  const secs = seconds % 60
  return `${String(mins).padStart(2, '0')}:${String(secs).padStart(2, '0')}`
}

function computeWPM(chars: number, seconds: number) {
  if (seconds === 0) return 0
  const words = chars / 5
  return Math.max(0, Math.round((words / seconds) * 60))
}
</script>

<style scoped>
/* ==================== 登录页 ==================== */
.auth-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-lg);
  background: var(--color-bg-secondary);
}

.auth-container {
  width: 100%;
  max-width: 400px;
  background: var(--color-bg-elevated);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-xl);
  padding: var(--space-2xl);
  box-shadow: var(--shadow-xl);
}

.auth-header {
  text-align: center;
  margin-bottom: var(--space-xl);
}

.auth-header h1 {
  font-size: var(--font-size-3xl);
  margin-bottom: var(--space-sm);
  letter-spacing: -0.02em;
}

.auth-form {
  display: flex;
  flex-direction: column;
  gap: var(--space-lg);
  margin-bottom: var(--space-lg);
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
}

.form-group label {
  font-size: var(--font-size-sm);
  font-weight: 500;
  color: var(--color-text-primary);
}

.auth-footer {
  text-align: center;
  padding-top: var(--space-lg);
  border-top: 1px solid var(--color-border);
}

.btn-link {
  background: none;
  border: none;
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
  cursor: pointer;
  padding: var(--space-xs) 0;
  transition: color var(--transition-fast);
}

.btn-link:hover {
  color: var(--color-text-primary);
}

/* ==================== 主应用布局 ==================== */
.app-layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--color-bg-primary);
}

/* 顶部导航栏 */
.app-header {
  height: var(--header-height);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 var(--space-lg);
  background: var(--color-bg-elevated);
  border-bottom: 1px solid var(--color-border);
  box-shadow: var(--shadow-sm);
  flex-shrink: 0;
}

.header-left,
.header-right {
  display: flex;
  align-items: center;
  gap: var(--space-md);
}

.app-title {
  font-size: var(--font-size-xl);
  font-weight: 600;
  letter-spacing: -0.02em;
}

.user-menu {
  position: relative;
}

.user-dropdown {
  position: absolute;
  top: calc(100% + var(--space-sm));
  right: 0;
  min-width: 200px;
  background: var(--color-bg-elevated);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  padding: var(--space-sm);
  opacity: 0;
  visibility: hidden;
  transform: translateY(-4px);
  transition: all var(--transition-fast);
  z-index: 100;
}

.user-menu:hover .user-dropdown {
  opacity: 1;
  visibility: visible;
  transform: translateY(0);
}

.user-info {
  padding: var(--space-sm);
}

.user-name {
  font-size: var(--font-size-sm);
  font-weight: 600;
  color: var(--color-text-primary);
  margin-bottom: var(--space-xs);
}

.user-email {
  font-size: var(--font-size-xs);
  color: var(--color-text-secondary);
}

.dropdown-divider {
  height: 1px;
  background: var(--color-border);
  margin: var(--space-sm) 0;
}

.dropdown-item {
  width: 100%;
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  padding: var(--space-sm);
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
  background: none;
  border: none;
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all var(--transition-fast);
  text-align: left;
}

.dropdown-item:hover {
  background: var(--color-accent-subtle);
  color: var(--color-text-primary);
}

/* 内容区域 */
.app-content {
  flex: 1;
  display: flex;
  overflow: hidden;
}

/* 侧边栏 */
.sidebar {
  width: var(--sidebar-width);
  background: var(--color-bg-elevated);
  border-right: 1px solid var(--color-border);
  display: flex;
  flex-direction: column;
  overflow-y: auto;
  transition: transform var(--transition-base);
}

.sidebar.collapsed {
  transform: translateX(-100%);
}

.sidebar-section {
  padding: var(--space-lg);
  border-bottom: 1px solid var(--color-border);
}

.sidebar-section:last-child {
  border-bottom: none;
  flex: 1;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-md);
}

.section-title {
  font-size: var(--font-size-sm);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--color-text-tertiary);
  margin: 0;
}

.import-actions {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
}

.btn-icon-sm {
  padding: var(--space-xs);
  min-width: 24px;
  min-height: 24px;
}

/* 工作区 */
.workspace {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: var(--color-bg-secondary);
}

.workspace-empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--space-md);
  text-align: center;
  color: var(--color-text-tertiary);
}

.empty-icon {
  opacity: 0.3;
}

.workspace-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* 进度栏 */
.progress-bar {
  height: 3px;
  background: var(--color-border);
  position: relative;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: var(--color-accent);
  transition: width var(--transition-base);
  position: relative;
}

.progress-fill::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.3), transparent);
  animation: shimmer 1.5s infinite;
}

@keyframes shimmer {
  0% { transform: translateX(-100%); }
  100% { transform: translateX(100%); }
}

/* 统计栏 */
.stats-bar {
  display: flex;
  align-items: center;
  gap: var(--space-xl);
  padding: var(--space-lg) var(--space-xl);
  background: var(--color-bg-elevated);
  border-bottom: 1px solid var(--color-border);
}

.stat {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
}

.stat-label {
  font-size: var(--font-size-xs);
  color: var(--color-text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.stat-value {
  font-size: var(--font-size-xl);
  font-weight: 600;
  color: var(--color-text-primary);
  font-variant-numeric: tabular-nums;
}

.text-error {
  color: var(--color-error);
}

/* 临摹区域 */
.practice-area {
  flex: 1;
  overflow-y: auto;
  padding: var(--space-xl);
}

/* 操作栏 */
.action-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-md) var(--space-xl);
  background: var(--color-bg-elevated);
  border-top: 1px solid var(--color-border);
  gap: var(--space-md);
}

.action-hint {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  font-size: var(--font-size-sm);
  color: var(--color-text-tertiary);
}

.action-hint kbd {
  padding: var(--space-xs) var(--space-sm);
  background: var(--color-bg-secondary);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  font-family: var(--font-mono);
  font-size: var(--font-size-xs);
  color: var(--color-text-secondary);
}

.divider {
  color: var(--color-border);
}

.action-buttons {
  display: flex;
  gap: var(--space-sm);
}

/* 响应式 */
@media (max-width: 1024px) {
  .stats-bar {
    gap: var(--space-lg);
    padding: var(--space-md) var(--space-lg);
  }

  .stat-value {
    font-size: var(--font-size-lg);
  }
}

@media (max-width: 768px) {
  .app-header {
    padding: 0 var(--space-md);
  }

  .sidebar {
    position: absolute;
    left: 0;
    top: var(--header-height);
    bottom: 0;
    z-index: 50;
    box-shadow: var(--shadow-lg);
  }

  .sidebar:not(.collapsed) {
    transform: translateX(0);
  }

  .stats-bar {
    flex-wrap: wrap;
    gap: var(--space-md);
  }

  .practice-area {
    padding: var(--space-md);
  }

  .action-bar {
    flex-direction: column;
    align-items: stretch;
  }

  .action-buttons {
    width: 100%;
  }

  .action-buttons .btn {
    flex: 1;
  }
}
</style>
