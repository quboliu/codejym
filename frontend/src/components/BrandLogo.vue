<template>
  <span class="brand-logo">
    <span class="brand-mark-wrap" :class="{ badge }" :style="badge ? badgeStyle : undefined">
      <svg
        class="brand-mark"
        :width="size"
        :height="size"
        viewBox="0 0 24 24"
        fill="none"
        :stroke="badge ? 'var(--color-bg-elevated)' : 'currentColor'"
        :stroke-width="strokeWidth"
        stroke-linecap="round"
        stroke-linejoin="round"
        aria-hidden="true"
      >
        <path d="M9 6 H6 V18 H9" />
        <path d="M15 6 H18 V18 H15" />
        <path d="M7.5 12 H16.5" />
      </svg>
    </span>
    <span v-if="withWordmark" class="brand-name" :style="{ fontSize: wordmarkSize }">CodeJYM</span>
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    /** Pixel size of the logo mark. */
    size?: number
    /** Show the "CodeJYM" wordmark next to the mark. */
    withWordmark?: boolean
    /** CSS font-size for the wordmark. */
    wordmarkSize?: string
    /** Stroke width of the mark. */
    strokeWidth?: number
    /** Render the mark inside a filled rounded badge (inverse colors). */
    badge?: boolean
  }>(),
  {
    size: 24,
    withWordmark: true,
    wordmarkSize: '1.25rem',
    strokeWidth: 2.4,
    badge: false,
  }
)

const badgeStyle = computed(() => {
  const pad = Math.round(props.size * 0.42)
  return {
    padding: `${pad}px`,
    borderRadius: `${Math.round((props.size + pad * 2) * 0.28)}px`,
  }
})
</script>

<style scoped>
.brand-logo {
  display: inline-flex;
  align-items: center;
  gap: 0.55em;
  color: var(--color-text-primary);
  user-select: none;
  line-height: 1;
}

.brand-mark-wrap {
  display: inline-flex;
  flex-shrink: 0;
}

.brand-mark-wrap.badge {
  background: var(--color-accent);
  box-shadow: var(--shadow-md);
}

.brand-mark {
  display: block;
}

.brand-name {
  font-family: var(--font-sans);
  font-weight: 700;
  letter-spacing: -0.035em;
  color: var(--color-text-primary);
  white-space: nowrap;
}
</style>
