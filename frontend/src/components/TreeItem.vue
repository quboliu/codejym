<template>
  <div>
    <div
      class="tree-item"
      :class="{
        'tree-item-active': !node.isDir && node.path === activePath,
        'tree-item-dir': node.isDir,
      }"
      :style="{ paddingLeft: `${depth * 16 + 8}px` }"
      @click="handleClick"
      @contextmenu.prevent="handleContextMenu"
    >
      <!-- 文件夹图标 -->
      <svg
        v-if="node.isDir"
        class="tree-icon"
        width="14"
        height="14"
        viewBox="0 0 20 20"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        :class="{ 'icon-rotated': isExpanded }"
      >
        <path d="M6 8L10 12L14 8" stroke-linecap="round" stroke-linejoin="round"/>
      </svg>

      <!-- 文件图标 -->
      <svg
        v-else
        class="tree-icon"
        width="14"
        height="14"
        viewBox="0 0 20 20"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
      >
        <path d="M4 3H10L13 6V17H4V3Z" stroke-linecap="round" stroke-linejoin="round"/>
        <path d="M10 3V6H13" stroke-linecap="round" stroke-linejoin="round"/>
      </svg>

      <span class="tree-label">{{ node.name }}</span>
    </div>

    <!-- 子节点 -->
    <div v-if="node.isDir && isExpanded && node.children" class="tree-children">
      <TreeItem
        v-for="child in node.children"
        :key="child.name"
        :node="child"
        :active-path="activePath"
        :depth="depth + 1"
        @select="$emit('select', $event)"
        @context-menu="$emit('context-menu', $event)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { FileNode } from '../types'

const props = defineProps<{
  node: FileNode
  activePath: string | null
  depth: number
}>()

const emit = defineEmits<{
  select: [path: string]
  'context-menu': [event: { node: FileNode; clientX: number; clientY: number }]
}>()

const isExpanded = ref(props.depth === 0) // 第一层默认展开

function handleClick() {
  if (props.node.isDir) {
    isExpanded.value = !isExpanded.value
  } else {
    emit('select', props.node.path)
  }
}

function handleContextMenu(event: MouseEvent) {
  emit('context-menu', {
    node: props.node,
    clientX: event.clientX,
    clientY: event.clientY,
  })
}
</script>

<style scoped>
.tree-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px var(--space-sm);
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
  user-select: none;
}

.tree-item:hover {
  background: var(--color-accent-subtle);
  color: var(--color-text-primary);
}

.tree-item-active {
  background: var(--color-accent);
  color: var(--color-text-inverse);
}

.tree-item-active:hover {
  background: var(--color-accent-hover);
}

.tree-item-dir {
  font-weight: 500;
}

.tree-icon {
  flex-shrink: 0;
  transition: transform var(--transition-fast);
}

.icon-rotated {
  transform: rotate(0deg);
}

.tree-item-dir .tree-icon:not(.icon-rotated) {
  transform: rotate(-90deg);
}

.tree-label {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tree-children {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
</style>
