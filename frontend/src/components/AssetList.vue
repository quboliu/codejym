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
        <span v-if="getExtension(asset.sourceName)" class="asset-badge">
          {{ getExtension(asset.sourceName).toUpperCase() }}
        </span>
      </div>
      <div class="asset-meta">
        <span class="meta-item">
          <svg width="12" height="12" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M4 3H10L13 6V17H4V3Z" stroke-linecap="round" stroke-linejoin="round"/>
            <path d="M10 3V6H13" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
          {{ asset.fileCount }}
        </span>
        <span class="meta-divider">·</span>
        <span class="meta-item">{{ formatBytes(asset.sizeBytes) }}</span>
      </div>
    </button>

    <div v-if="assets.length === 0" class="asset-empty">
      <p>暂无素材</p>
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
}>()

function getExtension(name: string) {
  const clean = (name ?? '').trim()
  if (!clean) return ''
  const lastDot = clean.lastIndexOf('.')
  if (lastDot <= 0 || lastDot === clean.length - 1) {
    return ''
  }
  return clean.slice(lastDot + 1).toLowerCase()
}

function formatBytes(bytes: number) {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  const exponent = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1)
  const value = bytes / Math.pow(1024, exponent)
  return `${value.toFixed(value >= 10 || exponent === 0 ? 0 : 1)} ${units[exponent]}`
}
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
  gap: var(--space-sm);
  padding: var(--space-md);
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

.asset-badge {
  flex-shrink: 0;
  font-size: var(--font-size-xs);
  font-weight: 600;
  padding: var(--space-xs) var(--space-sm);
  background: var(--color-bg-tertiary);
  border-radius: var(--radius-sm);
  letter-spacing: 0.05em;
}

.asset-card-active .asset-badge {
  background: rgba(255, 255, 255, 0.2);
}

.asset-meta {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
  font-size: var(--font-size-xs);
  color: var(--color-text-tertiary);
}

.asset-card-active .asset-meta {
  color: rgba(255, 255, 255, 0.8);
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 4px;
}

.meta-divider {
  opacity: 0.5;
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
