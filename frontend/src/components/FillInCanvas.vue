<template>
  <div class="fillin-canvas">
    <div class="fillin-header">
      <div class="fillin-meta">
        <span class="fillin-path">{{ source.path }}</span>
        <span class="fillin-badge">{{ template.difficulty }}</span>
        <span class="fillin-badge">{{ template.provider || 'local' }}</span>
      </div>
      <div class="fillin-intent">{{ template.intent }}</div>
    </div>

    <pre class="fillin-code hljs"><template v-for="part in parts" :key="part.key"><span v-if="part.type === 'text'" v-html="part.html"></span><span
      v-else
      class="blank-shell"
      :class="`blank-${part.blank.status}`"
    ><input
      class="blank-input"
      :style="{ width: `${inputWidth(part.blank)}ch` }"
      :value="inputs[part.blank.id] ?? part.blank.currentInput"
      :disabled="part.blank.status === 'correct' || part.blank.status === 'revealed'"
      :aria-label="`填空 ${part.blank.lineStart} 行`"
      spellcheck="false"
      autocomplete="off"
      @input="$emit('update-input', part.blank.id, ($event.target as HTMLInputElement).value)"
      @keydown.enter.prevent="$emit('submit', part.blank.id)"
    /><span class="blank-actions"><button
      class="blank-action"
      type="button"
      title="提交"
      :disabled="part.blank.status === 'correct' || part.blank.status === 'revealed'"
      @click="$emit('submit', part.blank.id)"
    >✓</button><button
      class="blank-action"
      type="button"
      title="显示答案"
      :disabled="part.blank.status === 'revealed'"
      @click="$emit('reveal', part.blank.id)"
    >?</button></span><span v-if="part.blank.hint" class="blank-hint">{{ part.blank.hint }}</span><span v-if="part.blank.status === 'revealed' && part.blank.answer" class="blank-answer">{{ part.blank.answer }}</span></span></template></pre>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import hljs from 'highlight.js'
import 'highlight.js/styles/github-dark.css'
import type { FileContent, FillInBlank, FillInTemplate } from '../types'

const props = defineProps<{
  source: FileContent
  template: FillInTemplate
  blanks: FillInBlank[]
  inputs: Record<string, string>
}>()

defineEmits<{
  'update-input': [blankId: string, value: string]
  submit: [blankId: string]
  reveal: [blankId: string]
}>()

type TextPart = {
  key: string
  type: 'text'
  html: string
}

type BlankPart = {
  key: string
  type: 'blank'
  blank: FillInBlank
}

type HighlightRun = {
  start: number
  end: number
  classes: string
  text: string
}

const highlightedRuns = computed(() => {
  const content = props.source.content
  if (!content) return []
  const language = props.source.language || 'plaintext'
  let highlighted = ''
  try {
    highlighted = hljs.highlight(content, { language, ignoreIllegals: true }).value
  } catch {
    highlighted = hljs.highlight(content, { language: 'plaintext' }).value
  }
  return flattenHighlightedHTML(highlighted, content)
})

const parts = computed<Array<TextPart | BlankPart>>(() => {
  const sorted = [...props.blanks].sort((a, b) => a.startOffset - b.startOffset)
  const out: Array<TextPart | BlankPart> = []
  let cursor = 0
  for (const blank of sorted) {
    const start = Math.max(0, Math.min(props.source.content.length, blank.startOffset))
    const end = Math.max(start, Math.min(props.source.content.length, blank.endOffset))
    if (start > cursor) {
      out.push({
        key: `text-${cursor}-${start}`,
        type: 'text',
        html: renderHighlightedSlice(cursor, start),
      })
    }
    out.push({
      key: `blank-${blank.id}`,
      type: 'blank',
      blank,
    })
    cursor = end
  }
  if (cursor < props.source.content.length) {
    out.push({
      key: `text-${cursor}-end`,
      type: 'text',
      html: renderHighlightedSlice(cursor, props.source.content.length),
    })
  }
  return out
})

function renderHighlightedSlice(start: number, end: number) {
  const runs = highlightedRuns.value
  if (runs.length === 0) {
    return escapeHTML(props.source.content.slice(start, end))
  }

  let html = ''
  let cursor = start
  for (const run of runs) {
    if (run.end <= start) continue
    if (run.start >= end) break
    const sliceStart = Math.max(start, run.start)
    const sliceEnd = Math.min(end, run.end)
    if (sliceStart > cursor) {
      html += escapeHTML(props.source.content.slice(cursor, sliceStart))
    }
    const text = run.text.slice(sliceStart - run.start, sliceEnd - run.start)
    html += run.classes
      ? `<span class="${run.classes}">${escapeHTML(text)}</span>`
      : escapeHTML(text)
    cursor = sliceEnd
  }
  if (cursor < end) {
    html += escapeHTML(props.source.content.slice(cursor, end))
  }
  return html
}

function flattenHighlightedHTML(highlighted: string, source: string): HighlightRun[] {
  const temp = document.createElement('div')
  temp.innerHTML = highlighted
  const runs: HighlightRun[] = []
  let offset = 0

  function walk(node: Node, inheritedClasses: string[]): void {
    if (node.nodeType === Node.TEXT_NODE) {
      const text = node.textContent || ''
      if (!text) return
      runs.push({
        start: offset,
        end: offset + text.length,
        classes: inheritedClasses.join(' '),
        text,
      })
      offset += text.length
      return
    }

    if (node.nodeType !== Node.ELEMENT_NODE && node.nodeType !== Node.DOCUMENT_FRAGMENT_NODE) {
      return
    }

    const element = node as Element
    const ownClasses = typeof element.className === 'string'
      ? element.className.split(/\s+/).filter(cls => cls.startsWith('hljs-'))
      : []
    const nextClasses = ownClasses.length > 0
      ? Array.from(new Set([...inheritedClasses, ...ownClasses]))
      : inheritedClasses
    Array.from(node.childNodes).forEach(child => walk(child, nextClasses))
  }

  walk(temp, [])
  if (offset !== source.length) {
    return []
  }
  return runs
}

function escapeHTML(value: string) {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}

function inputWidth(blank: FillInBlank) {
  const typed = props.inputs[blank.id] ?? blank.currentInput
  const targetWidth = Math.max(6, blank.endOffset - blank.startOffset)
  return Math.min(48, Math.max(targetWidth, typed.length + 1))
}
</script>

<style scoped>
.fillin-canvas {
  background: var(--color-bg-elevated);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  overflow: hidden;
}

.fillin-header {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
  padding: var(--space-md) var(--space-lg);
  background: var(--color-bg-secondary);
  border-bottom: 1px solid var(--color-border);
}

.fillin-meta {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  min-width: 0;
}

.fillin-path {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-family: var(--font-mono);
  font-size: var(--font-size-sm);
  color: var(--color-text-primary);
}

.fillin-badge {
  flex-shrink: 0;
  padding: 2px var(--space-sm);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  color: var(--color-text-tertiary);
  font-size: var(--font-size-xs);
  text-transform: uppercase;
}

.fillin-intent {
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
}

.fillin-code {
  margin: 0;
  padding: var(--space-lg);
  min-height: 400px;
  max-height: calc(100vh - 400px);
  overflow: auto;
  white-space: pre;
  min-width: max-content;
  font-family: var(--font-mono);
  font-size: 14px;
  line-height: 1.8;
  color: var(--color-text-primary);
  background: transparent;
}

.fillin-code.hljs {
  background: transparent !important;
}

.fillin-code :deep([class*="hljs-"]) {
  filter: saturate(1.15) contrast(1.05);
}

:global([data-theme="light"] .fillin-code [class*="hljs-"]) {
  filter: saturate(1.45) contrast(1.12) brightness(0.88);
}

.blank-shell {
  display: inline-flex;
  align-items: baseline;
  gap: 4px;
  margin: 0 1px;
  padding: 1px 3px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--color-border);
  background: var(--color-bg-secondary);
  vertical-align: baseline;
}

.blank-input {
  height: 1.7em;
  min-width: 6ch;
  padding: 0 4px;
  border: 0;
  border-bottom: 1px solid var(--color-text-tertiary);
  outline: none;
  background: transparent;
  color: var(--color-text-primary);
  font: inherit;
}

.blank-input:focus {
  border-bottom-color: var(--color-accent);
}

.blank-correct {
  border-color: rgba(34, 197, 94, 0.45);
  background: rgba(34, 197, 94, 0.08);
}

.blank-incorrect {
  border-color: rgba(239, 68, 68, 0.5);
  background: rgba(239, 68, 68, 0.08);
}

.blank-revealed {
  border-color: rgba(245, 158, 11, 0.45);
  background: rgba(245, 158, 11, 0.08);
}

.blank-actions {
  display: inline-flex;
  gap: 2px;
}

.blank-action {
  width: 20px;
  height: 20px;
  padding: 0;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  background: var(--color-bg-elevated);
  color: var(--color-text-secondary);
  cursor: pointer;
  font-size: 11px;
  line-height: 1;
}

.blank-action:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.blank-hint {
  color: var(--color-text-tertiary);
  font-size: 11px;
}

.blank-answer {
  color: var(--color-warning);
  font-weight: 600;
}

@media (max-width: 768px) {
  .fillin-code {
    padding: var(--space-md);
    font-size: 13px;
    min-height: 300px;
    max-height: calc(100vh - 350px);
  }
}
</style>
