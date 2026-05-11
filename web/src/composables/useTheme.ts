import { ref, watch, onMounted, type InjectionKey } from 'vue'
import { inject } from 'vue'

export type Theme = 'dark' | 'light'

export interface ThemeComposable {
  theme: Theme
  toggleTheme: () => void
  setTheme: (t: Theme) => void
}

export const ThemeKey: InjectionKey<ThemeComposable> = Symbol('theme')

const STORAGE_KEY = 'ai-graph-theme'

export function useTheme(): ThemeComposable {
  const theme = ref<Theme>('dark')

  function loadTheme() {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored === 'light' || stored === 'dark') {
      theme.value = stored
    } else if (window.matchMedia('(prefers-color-scheme: light)').matches) {
      theme.value = 'light'
    }
    applyTheme(theme.value)
  }

  function applyTheme(t: Theme) {
    document.documentElement.setAttribute('data-theme', t)
    if (t === 'light') {
      document.body.classList.add('light-theme')
    } else {
      document.body.classList.remove('light-theme')
    }
  }

  function toggleTheme() {
    theme.value = theme.value === 'dark' ? 'light' : 'dark'
  }

  function setTheme(t: Theme) {
    theme.value = t
  }

  watch(theme, (newTheme) => {
    localStorage.setItem(STORAGE_KEY, newTheme)
    applyTheme(newTheme)
  })

  onMounted(() => {
    loadTheme()
  })

  return {
    theme,
    toggleTheme,
    setTheme,
  }
}

export function provideTheme() {
  const themeComposable = useTheme()
  return themeComposable
}

export function injectTheme(): ThemeComposable {
  const theme = inject(ThemeKey)
  if (!theme) {
    throw new Error('useTheme must be used within a provider')
  }
  return theme
}