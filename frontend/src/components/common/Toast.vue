<template>
  <Teleport to="body">
    <TransitionGroup name="toast" tag="div" class="toast-container">
      <div
        v-for="toast in toasts"
        :key="toast.id"
        class="toast"
        :class="['toast-' + toast.type]"
        @click="removeToast(toast.id)"
      >
        <div class="toast-icon">
          <svg v-if="toast.type === 'success'" width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M5 10L8 13L15 6" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
          <svg v-else-if="toast.type === 'error'" width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M5 5L15 15M15 5L5 15" stroke-linecap="round"/>
          </svg>
          <svg v-else-if="toast.type === 'warning'" width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M10 6V11M10 14H10.01" stroke-linecap="round"/>
            <path d="M9 2L3 17H17L11 2H9Z" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
          <svg v-else width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="10" cy="10" r="8"/>
            <path d="M10 6V11M10 14H10.01" stroke-linecap="round"/>
          </svg>
        </div>
        <div class="toast-content">
          <div v-if="toast.title" class="toast-title">{{ toast.title }}</div>
          <div class="toast-message">{{ toast.message }}</div>
        </div>
      </div>
    </TransitionGroup>
  </Teleport>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { ToastType } from '../../composables/useToast'

defineOptions({
  name: 'CommonToast',
})

export interface Toast {
  id: string
  type: ToastType
  title?: string
  message: string
  duration?: number
}

const toasts = ref<Toast[]>([])

function addToast(toast: Omit<Toast, 'id'>) {
  const id = Date.now().toString()
  const newToast = { ...toast, id }
  toasts.value.push(newToast)

  // 自动移除
  const duration = toast.duration ?? 3000
  if (duration > 0) {
    setTimeout(() => {
      removeToast(id)
    }, duration)
  }

  return id
}

function removeToast(id: string) {
  const index = toasts.value.findIndex(t => t.id === id)
  if (index > -1) {
    toasts.value.splice(index, 1)
  }
}

defineExpose({
  addToast,
  removeToast,
})
</script>

<style scoped>
.toast-container {
  position: fixed;
  top: var(--space-lg);
  right: var(--space-lg);
  z-index: 2000;
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
  pointer-events: none;
}

.toast {
  display: flex;
  align-items: flex-start;
  gap: var(--space-md);
  min-width: 320px;
  max-width: 400px;
  padding: var(--space-md);
  background: var(--color-bg-elevated);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  cursor: pointer;
  pointer-events: auto;
  transition: all var(--transition-fast);
}

.toast:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-xl);
}

.toast-icon {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: var(--radius-md);
}

.toast-success .toast-icon {
  color: var(--color-success);
  background: rgba(34, 197, 94, 0.1);
}

.toast-error .toast-icon {
  color: var(--color-error);
  background: rgba(239, 68, 68, 0.1);
}

.toast-warning .toast-icon {
  color: var(--color-warning);
  background: rgba(245, 158, 11, 0.1);
}

.toast-info .toast-icon {
  color: var(--color-info);
  background: rgba(59, 130, 246, 0.1);
}

.toast-content {
  flex: 1;
  min-width: 0;
}

.toast-title {
  font-size: var(--font-size-sm);
  font-weight: 600;
  color: var(--color-text-primary);
  margin-bottom: var(--space-xs);
}

.toast-message {
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
  line-height: 1.5;
}

/* 动画 */
.toast-enter-active {
  animation: toastSlideIn var(--transition-base);
}

.toast-leave-active {
  animation: toastSlideOut var(--transition-fast);
}

@keyframes toastSlideIn {
  from {
    transform: translateX(100%);
    opacity: 0;
  }
  to {
    transform: translateX(0);
    opacity: 1;
  }
}

@keyframes toastSlideOut {
  from {
    transform: translateX(0);
    opacity: 1;
  }
  to {
    transform: translateX(100%);
    opacity: 0;
  }
}

/* 响应式 */
@media (max-width: 768px) {
  .toast-container {
    top: var(--space-md);
    right: var(--space-md);
    left: var(--space-md);
  }

  .toast {
    min-width: 0;
    max-width: none;
  }
}
</style>
