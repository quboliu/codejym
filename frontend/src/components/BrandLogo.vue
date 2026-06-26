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
defineProps({
  size: {
    type: Number,
    default: 24,
  },
  wordmarkSize: {
    type: String,
    default: '',
  },
})
</script>

<style scoped>
.brand-logo {
  display: inline-flex;
  align-items: center;
  gap: 0.6em;
  color: var(--color-text-primary);
  user-select: none;
  line-height: 1;
  transition: opacity var(--transition-fast);
}

.brand-logo:hover {
  opacity: 0.8;
}

.brand-seal {
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-accent); /* Seal red */
  transform: rotate(-3deg); /* Slight handcrafted rotation */
}

.brand-seal svg {
  width: 100%;
  height: 100%;
}

.brand-name {
  font-family: var(--font-display);
  font-weight: 500;
  letter-spacing: -0.02em;
  color: var(--color-text-primary);
  white-space: nowrap;
}
</style>
