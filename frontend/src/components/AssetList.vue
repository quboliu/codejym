<template>
  <div class="asset-list">
    <button
      v-for="asset in assets"
      :key="asset.id"
      class="asset-card"
      :class="{ 'asset-card-active': selectedId === asset.id }"
      @click="$emit('select', asset.id)"
    >
      <div class="asset-header">
        <span class="asset-name">{{ asset.name }}</span>
        <div class="asset-actions">
          <button
            class="btn-icon-xs"
            @click.stop="$emit('rename', asset.id)"
            title="重命名"
          >
            <svg width="14" height="14" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M11 3L17 3L17 9M16 4L9 11L6 14L3 14L3 11L6 8L13 1" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
          </button>
          <button
            class="btn-icon-xs btn-icon-danger"
            @click.stop="$emit('delete', asset.id)"
            title="删除"
          >
            <svg width="14" height="14" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M3 5H17M8 5V3H12V5M8 9V15M12 9V15M7 5H13L14 17H6L7 5Z" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
          </button>
        </div>
      </div>
    </button>

    <div v-if="assets.length === 0" class="asset-empty">
      <p>暂无训练组</p>
      <p class="text-secondary">上传文件或粘贴代码开始练习</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Asset } from '../types'

defineProps<{
  assets: Asset[]
  selectedId?: string | null
}>()

defineEmits<{
  select: [assetId: string]
  rename: [assetId: string]
  delete: [assetId: string]
}>()
</script>

<style scoped>
.asset-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
  max-height: 320px;
  overflow-y: auto;
}

.asset-card {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
  padding: var(--space-sm) var(--space-md);
  background: var(--color-bg-secondary);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  text-align: left;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.asset-card:hover {
  background: var(--color-bg-tertiary);
  border-color: var(--color-border-hover);
  transform: translateY(-1px);
}

.asset-card-active {
  background: var(--color-accent);
  border-color: var(--color-accent);
  color: var(--color-text-inverse);
}

.asset-card-active:hover {
  background: var(--color-accent-hover);
  border-color: var(--color-accent-hover);
}

.asset-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-sm);
}

.asset-name {
  flex: 1;
  min-width: 0;
  font-size: var(--font-size-sm);
  font-weight: 600;
  color: inherit;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.asset-actions {
  display: flex;
  align-items: center;
  gap: 4px;
  opacity: 0;
  transition: opacity var(--transition-fast);
}

.asset-card:hover .asset-actions,
.asset-card:focus-within .asset-actions {
  opacity: 1;
}

.btn-icon-xs {
  padding: 4px;
  background: transparent;
  border: none;
  border-radius: var(--radius-sm);
  cursor: pointer;
  color: inherit;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all var(--transition-fast);
}

.btn-icon-xs:hover {
  background: rgba(0, 0, 0, 0.1);
}

.asset-card-active .btn-icon-xs:hover {
  background: rgba(255, 255, 255, 0.2);
}

.btn-icon-danger:hover {
  background: var(--color-error) !important;
  color: white !important;
}

.asset-empty {
  padding: var(--space-xl) var(--space-lg);
  text-align: center;
  color: var(--color-text-tertiary);
}

.asset-empty p {
  font-size: var(--font-size-sm);
  margin-bottom: var(--space-xs);
}

.asset-empty p:first-child {
  font-weight: 500;
  color: var(--color-text-secondary);
}
</style>
