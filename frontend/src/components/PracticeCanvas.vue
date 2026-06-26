<template>
  <div class="practice-canvas" :class="{ 'canvas-error': errorFlash }">
    <!-- 文件信息头 -->
    <div class="canvas-header">
      <div class="file-info">
        <span class="file-path">{{ content.path }}</span>
        <span class="file-lang">{{ content.language.toUpperCase() }}</span>
      </div>
      <div class="next-char-hint">
        下一个字符：<kbd>{{ displayChar(currentChar) }}</kbd>
      </div>
    </div>

    <!-- 代码显示区域 - 单层渲染 -->
    <div class="code-display">
      <!-- 背景参考层 -->
      <pre class="code-base hljs" :style="{ opacity: bgOpacity }" v-html="highlightedCode"></pre>
      <!-- 前景进度层 -->
      <pre class="code-foreground hljs" v-html="displayContent"></pre>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import hljs from 'highlight.js'
import 'highlight.js/styles/github-dark.css'
import type { FileContent } from '../types'

const props = defineProps<{
  content: FileContent
  cursor: number
  errorFlash: boolean
  backgroundOpacity?: number
}>()

const currentChar = computed(() => {
  if (props.cursor >= props.content.content.length) return ''
  return props.content.content.slice(props.cursor, props.cursor + 1)
})

const bgOpacity = computed(() => {
  return props.backgroundOpacity ?? 0.5
})

// 背景层：完整的语法高亮代码
const highlightedCode = computed(() => {
  const language = props.content.language || 'plaintext'
  try {
    const result = hljs.highlight(props.content.content, { language, ignoreIllegals: true })
    return result.value
  } catch {
    return hljs.highlight(props.content.content, { language: 'plaintext' }).value
  }
})

// 前景层：在语法高亮的HTML中插入光标和标记completed/remaining
const displayContent = computed(() => {
  const language = props.content.language || 'plaintext'
  let highlighted = ''
  try {
    const result = hljs.highlight(props.content.content, { language, ignoreIllegals: true })
    highlighted = result.value
  } catch {
    highlighted = hljs.highlight(props.content.content, { language: 'plaintext' }).value
  }

  // 创建临时DOM来解析HTML
  const temp = document.createElement('div')
  temp.innerHTML = highlighted

  // 遍历所有文本节点，只按光标进度标记 completed / remaining 并插入光标。
  let charCount = 0
  let cursorInserted = false

  // 提取父元素的 hljs 相关 class
  function getParentHljsClasses(node: Node): string {
    const parent = node.parentNode as Element
    if (!parent || !parent.className) return ''

    // 只提取 hljs- 开头的 class
    const classes = parent.className.split(' ')
    const hljsClasses = classes.filter(cls => cls.startsWith('hljs-'))
    return hljsClasses.join(' ')
  }

  function processNode(node: Node): void {
    // 优先处理元素节点
    if (node.nodeType === Node.ELEMENT_NODE) {
      const element = node as Element
      const elementText = element.textContent || ''
      const nodeStart = charCount
      const nodeEnd = charCount + elementText.length

      if (nodeEnd <= props.cursor) {
        // 整个元素都是已完成 - 直接添加 class，保留原有高亮
        element.classList.add('code-completed')
        charCount += elementText.length
        return  // 不递归子节点
      } else if (nodeStart > props.cursor) {
        // 整个元素都是未完成 - 直接添加 class，保留原有高亮
        element.classList.add('code-remaining')
        charCount += elementText.length
        return  // 不递归子节点
      }

      // 光标在元素中间，递归处理子节点
      const children = Array.from(element.childNodes)
      children.forEach(child => processNode(child))
      return
    }

    // 处理文本节点
    if (node.nodeType === Node.TEXT_NODE) {
      const text = node.textContent || ''
      const nodeStart = charCount
      const nodeEnd = charCount + text.length

      if (nodeEnd <= props.cursor) {
        // 整个节点都是已完成
        const span = document.createElement('span')
        const hljsClasses = getParentHljsClasses(node)
        span.className = hljsClasses ? `${hljsClasses} code-completed` : 'code-completed'
        span.textContent = text
        node.parentNode!.replaceChild(span, node)
      } else if (nodeStart > props.cursor) {
        // 整个节点都是未完成（光标在节点之前）
        const span = document.createElement('span')
        const hljsClasses = getParentHljsClasses(node)
        span.className = hljsClasses ? `${hljsClasses} code-remaining` : 'code-remaining'
        span.textContent = text
        node.parentNode!.replaceChild(span, node)
      } else if (!cursorInserted) {
        // 光标在这个节点中（包括开头、中间）
        const offset = props.cursor - nodeStart
        const before = text.slice(0, offset)
        const after = text.slice(offset)

        const parent = node.parentNode!
        const hljsClasses = getParentHljsClasses(node)

        // 创建新节点 - 保留语法高亮 class
        const completedSpan = document.createElement('span')
        completedSpan.className = hljsClasses ? `${hljsClasses} code-completed` : 'code-completed'
        completedSpan.textContent = before

        const cursorSpan = document.createElement('span')
        cursorSpan.className = 'cursor-line'

        const remainingSpan = document.createElement('span')
        remainingSpan.className = hljsClasses ? `${hljsClasses} code-remaining` : 'code-remaining'
        remainingSpan.textContent = after

        // 替换原节点
        parent.replaceChild(remainingSpan, node)
        parent.insertBefore(cursorSpan, remainingSpan)
        parent.insertBefore(completedSpan, cursorSpan)

        cursorInserted = true
      }

      charCount += text.length
    }
  }

  processNode(temp)

  // 如果还没插入光标（cursor在末尾），在最后添加
  if (!cursorInserted) {
    const cursorSpan = document.createElement('span')
    cursorSpan.className = 'cursor-line'
    temp.appendChild(cursorSpan)
  }

  return temp.innerHTML
})

function displayChar(char: string) {
  if (!char) return '✓ 完成'
  if (char === '\n') return '↵ 换行'
  if (char === '\t') return '⇥ Tab'
  if (char === ' ') return '␣ 空格'
  return char
}
</script>

<style scoped>
.practice-canvas {
  background: var(--color-bg-elevated);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  overflow: hidden;
  transition: all var(--transition-fast);
}

.canvas-error {
  animation: shake 0.3s cubic-bezier(.36,.07,.19,.97);
}

@keyframes shake {
  0%, 100% { transform: translateX(0); }
  10%, 30%, 50%, 70%, 90% { transform: translateX(-4px); }
  20%, 40%, 60%, 80% { transform: translateX(4px); }
}

/* 文件信息头 */
.canvas-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-md) var(--space-lg);
  background: var(--color-bg-secondary);
  border-bottom: 1px solid var(--color-border);
  gap: var(--space-md);
  flex-wrap: wrap;
}

.file-info {
  display: flex;
  align-items: center;
  gap: var(--space-md);
  min-width: 0;
  flex: 1;
}

.file-path {
  font-size: var(--font-size-sm);
  font-weight: 500;
  color: var(--color-text-primary);
  font-family: var(--font-mono);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.file-lang {
  flex-shrink: 0;
  font-size: var(--font-size-xs);
  font-weight: 600;
  color: var(--color-text-tertiary);
  background: var(--color-bg-tertiary);
  padding: var(--space-xs) var(--space-sm);
  border-radius: var(--radius-sm);
  letter-spacing: 0.05em;
}

.next-char-hint {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
}

.next-char-hint kbd {
  padding: var(--space-xs) var(--space-sm);
  background: var(--color-bg-elevated);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  font-family: var(--font-mono);
  font-size: var(--font-size-xs);
  font-weight: 600;
  color: var(--color-text-primary);
  min-width: 40px;
  text-align: center;
}

/* 代码显示区域 */
.code-display {
  position: relative;
  display: grid;
  grid-template-columns: 1fr;
  grid-template-rows: 1fr;
  overflow: auto;
  min-height: 400px;
  max-height: calc(100vh - 400px);
}

.code-base,
.code-foreground {
  grid-area: 1 / 1 / 2 / 2;
  margin: 0;
  padding: var(--space-lg);
  font-family: var(--font-mono);
  font-size: 14px;
  line-height: 1.6;
  white-space: pre;
  min-width: max-content;
  word-wrap: normal;
  overflow-wrap: normal;
  word-break: normal;
  tab-size: 2;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-rendering: optimizeLegibility;
}

/* 背景层 - 参考文本 */
.code-base {
  /* opacity 通过行内样式动态设置 */
}

.code-base.hljs {
  background: transparent !important;
  color: var(--color-text-tertiary) !important;
}

.code-base :deep(*) {
  opacity: inherit;
}

/* 前景层 - 用户进度 */
.code-foreground {
  pointer-events: none;
  z-index: 1;
}

.code-foreground.hljs {
  background: transparent !important;
}

/* 已完成的代码 - 保留语法高亮，正常显示 */
.code-foreground :deep(.code-completed) {
  /* 不设置 color，让 hljs class 的颜色生效 */
}

/* 已完成的代码中的所有语法高亮元素 - 使用通配符匹配所有 hljs- class */
.code-foreground :deep(.code-completed[class*="hljs-"]) {
  /* 保持 hljs 原有的颜色，稍微降低亮度表示已完成 */
  filter: brightness(0.9) !important;
}

/* 浅色主题下，让已完成的语法色更饱和、更深一些 */
:global([data-theme="light"] .code-foreground .code-completed[class*="hljs-"]) {
  filter: saturate(1.55) contrast(1.12) brightness(0.86) !important;
}

/* 已完成但没有高亮的普通文本 */
.code-foreground :deep(.code-completed:not([class*="hljs-"])) {
  color: var(--color-text-primary);
}

/* 未完成的代码 - 隐藏 */
.code-foreground :deep(.code-remaining) {
  visibility: hidden;
}

/* 光标 */
.code-foreground :deep(.cursor-line) {
  display: inline-block;
  width: 0;
  height: 1em;
  border-left: 2px solid var(--color-accent);
  animation: blink 1s step-end infinite;
  vertical-align: baseline;
  margin: 0;
  padding: 0;
}

/* 闪烁动画 */
@keyframes blink {
  0%, 49% { opacity: 1; }
  50%, 100% { opacity: 0; }
}

/* 响应式 */
@media (max-width: 768px) {
  .canvas-header {
    padding: var(--space-sm) var(--space-md);
  }

  .code-base,
  .code-foreground {
    padding: var(--space-md);
    font-size: 13px;
    line-height: 1.6;
  }

  .code-display {
    min-height: 300px;
    max-height: calc(100vh - 350px);
  }
}
</style>
