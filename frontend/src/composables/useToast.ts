import { ref } from 'vue'

// 定义 Toast 类型
export type ToastType = 'success' | 'error' | 'warning' | 'info'

interface ToastOptions {
  title?: string
  message: string
  type?: ToastType
  duration?: number
}

// 全局Toast实例引用
export const toastRef = ref<{ addToast: (toast: any) => void } | null>(null)

export function useToast() {
  function showToast(options: ToastOptions) {
    if (!toastRef.value) {
      console.warn('Toast component not mounted')
      return
    }

    toastRef.value.addToast({
      type: options.type ?? 'info',
      title: options.title,
      message: options.message,
      duration: options.duration ?? 3000,
    })
  }

  function success(message: string, title?: string) {
    showToast({ type: 'success', message, title })
  }

  function error(message: string, title?: string) {
    showToast({ type: 'error', message, title })
  }

  function warning(message: string, title?: string) {
    showToast({ type: 'warning', message, title })
  }

  function info(message: string, title?: string) {
    showToast({ type: 'info', message, title })
  }

  return {
    showToast,
    success,
    error,
    warning,
    info,
  }
}
