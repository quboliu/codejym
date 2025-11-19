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

    <!-- 代码显示区域 -->
    <div class="code-display">
      <pre class="code-base" aria-hidden="true">{{ content.content }}</pre>
      <pre class="code-overlay">
        <span class="text-completed">{{ completed }}</span>
        <span class="text-cursor" :class="{ 'cursor-end': atEnd }">{{ atEnd ? '\u200b' : currentChar }}</span>
        <span class="text-remaining">{{ remaining }}</span>
      </pre>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { FileContent } from '../types'

const props = defineProps<{
  content: FileContent
  cursor: number
  errorFlash: boolean
}>()

const completed = computed(() => {
  return props.content.content.slice(0, props.cursor)
})

const currentChar = computed(() => {
  if (props.cursor >= props.content.content.length) return ''
  return props.content.content.slice(props.cursor, props.cursor + 1)
})

const remaining = computed(() => {
  if (props.cursor >= props.content.content.length) return ''
  return props.content.content.slice(props.cursor + 1)
})

const atEnd = computed(() => {
  return props.cursor >= props.content.content.length
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
.code-overlay {
  grid-area: 1 / 1 / 2 / 2;
  margin: 0;
  padding: var(--space-lg);
  font-family: var(--font-mono);
  font-size: 14px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
  tab-size: 2;
  -webkit-font-smoothing: subpixel-antialiased;
}

/* 基础层 - 参考文本 */
.code-base {
  color: var(--color-text-tertiary);
  opacity: 0.3;
}

/* 覆盖层 - 用户进度 */
.code-overlay {
  color: var(--color-text-primary);
  pointer-events: none;
  z-index: 1;
}

/* 已完成文本 */
.text-completed {
  color: var(--color-text-primary);
}

/* 光标 */
.text-cursor {
  position: relative;
  background: var(--color-accent);
  color: var(--color-text-inverse);
  border-radius: 2px;
  padding: 0 1px;
}

.text-cursor::after {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 2px;
  background: var(--color-accent);
  animation: blink 1s step-end infinite;
}

.cursor-end {
  display: inline-block;
  width: 8px;
  height: 1.4em;
  vertical-align: text-bottom;
  background: var(--color-accent);
  margin-left: 1px;
}

.cursor-end::after {
  display: none;
}

@keyframes blink {
  0%, 49% { opacity: 1; }
  50%, 100% { opacity: 0; }
}

/* 剩余文本 - 隐藏但占位 */
.text-remaining {
  visibility: hidden;
}

/* 响应式 */
@media (max-width: 768px) {
  .canvas-header {
    padding: var(--space-sm) var(--space-md);
  }

  .code-base,
  .code-overlay {
    padding: var(--space-md);
    font-size: 13px;
  }

  .code-display {
    min-height: 300px;
    max-height: calc(100vh - 350px);
  }
}
</style>
