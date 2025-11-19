<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="modelValue" class="modal-overlay" @click="handleOverlayClick">
        <div class="modal-container" :class="sizeClass" @click.stop>
          <!-- 头部 -->
          <div v-if="!hideHeader" class="modal-header">
            <h3 class="modal-title">{{ title }}</h3>
            <button
              v-if="!hideClose"
              class="modal-close"
              @click="handleClose"
              aria-label="关闭"
            >
              <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
                <path d="M5 5L15 15M15 5L5 15"/>
              </svg>
            </button>
          </div>

          <!-- 内容 -->
          <div class="modal-body">
            <slot></slot>
          </div>

          <!-- 底部操作栏 -->
          <div v-if="!hideFooter" class="modal-footer">
            <slot name="footer">
              <button class="btn btn-ghost" @click="handleCancel">
                {{ cancelText }}
              </button>
              <button class="btn btn-primary" @click="handleConfirm" :disabled="confirmDisabled">
                {{ confirmText }}
              </button>
            </slot>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    modelValue: boolean
    title?: string
    size?: 'sm' | 'md' | 'lg'
    hideHeader?: boolean
    hideFooter?: boolean
    hideClose?: boolean
    closeOnOverlay?: boolean
    confirmText?: string
    cancelText?: string
    confirmDisabled?: boolean
  }>(),
  {
    title: '提示',
    size: 'md',
    hideHeader: false,
    hideFooter: false,
    hideClose: false,
    closeOnOverlay: true,
    confirmText: '确定',
    cancelText: '取消',
    confirmDisabled: false,
  }
)

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'confirm': []
  'cancel': []
}>()

const sizeClass = computed(() => `modal-${props.size}`)

function handleClose() {
  emit('update:modelValue', false)
  emit('cancel')
}

function handleOverlayClick() {
  if (props.closeOnOverlay) {
    handleClose()
  }
}

function handleConfirm() {
  emit('confirm')
}

function handleCancel() {
  handleClose()
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-lg);
  z-index: 1000;
}

.modal-container {
  background: var(--color-bg-elevated);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-xl);
  display: flex;
  flex-direction: column;
  max-height: 90vh;
  width: 100%;
}

.modal-sm { max-width: 400px; }
.modal-md { max-width: 600px; }
.modal-lg { max-width: 800px; }

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-lg);
  border-bottom: 1px solid var(--color-border);
}

.modal-title {
  font-size: var(--font-size-lg);
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0;
}

.modal-close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  padding: 0;
  background: transparent;
  border: none;
  border-radius: var(--radius-md);
  color: var(--color-text-tertiary);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.modal-close:hover {
  background: var(--color-accent-subtle);
  color: var(--color-text-primary);
}

.modal-body {
  flex: 1;
  overflow-y: auto;
  padding: var(--space-lg);
  color: var(--color-text-secondary);
  line-height: 1.6;
}

.modal-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: var(--space-sm);
  padding: var(--space-lg);
  border-top: 1px solid var(--color-border);
}

/* 动画 */
.modal-enter-active {
  transition: opacity var(--transition-base);
}

.modal-leave-active {
  transition: opacity var(--transition-fast);
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-active .modal-container {
  animation: modalSlideUp var(--transition-base);
}

.modal-leave-active .modal-container {
  animation: modalSlideDown var(--transition-fast);
}

@keyframes modalSlideUp {
  from {
    transform: translateY(20px);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

@keyframes modalSlideDown {
  from {
    transform: translateY(0);
    opacity: 1;
  }
  to {
    transform: translateY(20px);
    opacity: 0;
  }
}

/* 响应式 */
@media (max-width: 768px) {
  .modal-overlay {
    padding: var(--space-md);
  }

  .modal-container {
    max-height: 95vh;
  }

  .modal-header,
  .modal-body,
  .modal-footer {
    padding: var(--space-md);
  }
}
</style>
