<template>
  <div class="file-tree-container">
    <div class="file-tree-toolbar">
      <button class="btn-icon btn-icon-sm" @click="$emit('create-folder')" title="新建文件夹">
        <svg width="16" height="16" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M3 5H8L10 7H17V17H3V5Z" stroke-linecap="round" stroke-linejoin="round"/>
          <path d="M10 10V14M8 12H12" stroke-linecap="round" stroke-linejoin="round"/>
        </svg>
      </button>
    </div>
    <div class="file-tree">
      <TreeItem
        v-for="node in nodes"
        :key="node.path"
        :node="node"
        :active-path="activePath"
        :depth="0"
        @select="handleSelect"
        @context-menu="handleContextMenu"
      />
      <div v-if="nodes.length === 0" class="tree-empty">
        <p>暂无文件</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import TreeItem from './TreeItem.vue'
import type { FileNode } from '../types'

defineProps<{
  nodes: FileNode[]
  activePath: string | null
}>()

const emit = defineEmits<{
  select: [path: string]
  'context-menu': [event: { node: FileNode; clientX: number; clientY: number }]
  'create-folder': []
}>()

function handleSelect(path: string) {
  emit('select', path)
}

function handleContextMenu(event: { node: FileNode; clientX: number; clientY: number }) {
  emit('context-menu', event)
}
</script>

<style scoped>
.file-tree-container {
  display: flex;
  flex-direction: column;
}

.file-tree-toolbar {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
  padding: var(--space-xs) 0;
  margin-bottom: var(--space-xs);
  border-bottom: 1px solid var(--color-border);
}

.file-tree {
  display: flex;
  flex-direction: column;
  gap: 2px;
  max-height: 400px;
  overflow-y: auto;
}

.tree-empty {
  padding: var(--space-lg);
  text-align: center;
  color: var(--color-text-tertiary);
  font-size: var(--font-size-sm);
}
</style>
