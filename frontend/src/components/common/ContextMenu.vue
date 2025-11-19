<template>
  <Teleport to="body">
    <div
      v-if="visible"
      class="context-menu-backdrop"
      @click="emit('close')"
      @contextmenu.prevent
    >
      <div
        class="context-menu"
        :style="{ top: `${position.y}px`, left: `${position.x}px` }"
        @click.stop
      >
        <button
          v-for="item in items"
          :key="item.label"
          class="context-menu-item"
          :class="{ 'context-menu-item-danger': item.danger }"
          @click="handleItemClick(item)"
          :disabled="item.disabled"
        >
          <span class="context-menu-icon" v-if="item.icon" v-html="item.icon"></span>
          <span class="context-menu-label">{{ item.label }}</span>
        </button>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
export interface ContextMenuItem {
  label: string
  icon?: string
  action: () => void
  danger?: boolean
  disabled?: boolean
}

interface Props {
  visible: boolean
  position: { x: number; y: number }
  items: ContextMenuItem[]
}

defineProps<Props>()

const emit = defineEmits<{
  close: []
}>()

function handleItemClick(item: ContextMenuItem) {
  if (!item.disabled) {
    item.action()
    emit('close')
  }
}
</script>

<style scoped>
.context-menu-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 9999;
  background: transparent;
}

.context-menu {
  position: fixed;
  min-width: 180px;
  background: var(--color-bg-elevated);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  padding: var(--space-xs);
  z-index: 10000;
}

.context-menu-item {
  width: 100%;
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  padding: var(--space-sm) var(--space-md);
  background: transparent;
  border: none;
  border-radius: var(--radius-sm);
  text-align: left;
  cursor: pointer;
  font-size: var(--font-size-sm);
  color: var(--color-text-primary);
  transition: background-color var(--transition-fast);
}

.context-menu-item:hover:not(:disabled) {
  background: var(--color-bg-secondary);
}

.context-menu-item:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.context-menu-item-danger {
  color: var(--color-error);
}

.context-menu-item-danger:hover:not(:disabled) {
  background: var(--color-error);
  color: white;
}

.context-menu-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
}

.context-menu-icon :deep(svg) {
  width: 16px;
  height: 16px;
  stroke: currentColor;
}

.context-menu-label {
  flex: 1;
}
</style>
