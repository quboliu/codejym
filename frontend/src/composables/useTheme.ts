import { ref, onMounted } from 'vue'

type Theme = 'light' | 'dark' | 'auto'

const THEME_KEY = 'codecopybook_theme'

export function useTheme() {
  const theme = ref<Theme>('auto')
  const isDark = ref(false)

  // 检测系统主题偏好
  function detectSystemTheme(): 'light' | 'dark' {
    if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
      return 'dark'
    }
    return 'light'
  }

  // 应用主题
  function applyTheme(t: Theme) {
    const root = document.documentElement

    if (t === 'auto') {
      const systemTheme = detectSystemTheme()
      root.setAttribute('data-theme', systemTheme)
      isDark.value = systemTheme === 'dark'
    } else {
      root.setAttribute('data-theme', t)
      isDark.value = t === 'dark'
    }
  }

  // 设置主题
  function setTheme(t: Theme) {
    theme.value = t
    applyTheme(t)
    localStorage.setItem(THEME_KEY, t)
  }

  // 切换主题（在 light/dark 之间切换）
  function toggleTheme() {
    if (theme.value === 'auto') {
      // 从 auto 切换到相反的系统主题
      const systemTheme = detectSystemTheme()
      setTheme(systemTheme === 'dark' ? 'light' : 'dark')
    } else {
      // 在 light/dark 之间切换
      setTheme(theme.value === 'dark' ? 'light' : 'dark')
    }
  }

  // 初始化
  onMounted(() => {
    // 从本地存储加载主题
    const stored = localStorage.getItem(THEME_KEY) as Theme | null
    if (stored && ['light', 'dark', 'auto'].includes(stored)) {
      theme.value = stored
    }

    applyTheme(theme.value)

    // 监听系统主题变化
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    const handleChange = () => {
      if (theme.value === 'auto') {
        applyTheme('auto')
      }
    }

    if (mediaQuery.addEventListener) {
      mediaQuery.addEventListener('change', handleChange)
    } else {
      // 旧版浏览器兼容
      mediaQuery.addListener(handleChange)
    }
  })

  return {
    theme,
    isDark,
    setTheme,
    toggleTheme,
  }
}
