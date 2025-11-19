<template>
  <div class="file-tree">
    <TreeItem
      v-for="node in nodes"
      :key="node.path"
      :node="node"
      :active-path="activePath"
      :depth="0"
      @select="handleSelect"
    />
    <div v-if="nodes.length === 0" class="tree-empty">
      <p>暂无文件</p>
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
}>()

function handleSelect(path: string) {
  emit('select', path)
}
</script>

<style scoped>
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
