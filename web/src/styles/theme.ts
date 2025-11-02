// Binance风格设计系统

export const theme = {
  colors: {
    // 背景色
    background: {
      primary: '#0B0E11',
      secondary: '#1E2329',
      tertiary: '#2B3139',
      elevated: '#181A20',
    },
    // 文本色
    text: {
      primary: '#EAECEF',
      secondary: '#848E9C',
      tertiary: '#5E6673',
      disabled: '#474D57',
    },
    // 品牌色
    brand: {
      primary: '#F0B90B',
      secondary: '#FCD535',
      light: 'rgba(240, 185, 11, 0.1)',
      gradient: 'linear-gradient(135deg, #F0B90B 0%, #FCD535 100%)',
    },
    // 功能色
    success: {
      main: '#0ECB81',
      light: 'rgba(14, 203, 129, 0.1)',
      border: 'rgba(14, 203, 129, 0.2)',
    },
    error: {
      main: '#F6465D',
      light: 'rgba(246, 70, 93, 0.1)',
      border: 'rgba(246, 70, 93, 0.2)',
    },
    warning: {
      main: '#F0B90B',
      light: 'rgba(240, 185, 11, 0.1)',
      border: 'rgba(240, 185, 11, 0.2)',
    },
    info: {
      main: '#60a5fa',
      light: 'rgba(96, 165, 250, 0.1)',
      border: 'rgba(96, 165, 250, 0.2)',
    },
    // 紫色系（AI/高级功能）
    purple: {
      main: '#8B5CF6',
      secondary: '#A78BFA',
      light: 'rgba(139, 92, 246, 0.1)',
      border: 'rgba(139, 92, 246, 0.3)',
      gradient: 'linear-gradient(135deg, #8B5CF6 0%, #6366F1 100%)',
    },
    // 边框色
    border: {
      primary: '#2B3139',
      secondary: '#474D57',
    },
  },
  spacing: {
    xs: '0.25rem',    // 4px
    sm: '0.5rem',     // 8px
    md: '0.75rem',    // 12px
    lg: '1rem',       // 16px
    xl: '1.5rem',     // 24px
    '2xl': '2rem',    // 32px
    '3xl': '3rem',    // 48px
  },
  radius: {
    sm: '0.375rem',   // 6px
    md: '0.5rem',     // 8px
    lg: '0.75rem',    // 12px
    xl: '1rem',       // 16px
    '2xl': '1.5rem',  // 24px
    full: '9999px',
  },
  shadows: {
    sm: '0 2px 8px rgba(0, 0, 0, 0.3)',
    md: '0 4px 16px rgba(0, 0, 0, 0.4)',
    lg: '0 8px 32px rgba(0, 0, 0, 0.5)',
    xl: '0 20px 60px rgba(0, 0, 0, 0.5)',
    brand: '0 4px 16px rgba(240, 185, 11, 0.3)',
    purple: '0 4px 16px rgba(139, 92, 246, 0.3)',
    success: '0 4px 16px rgba(16, 185, 129, 0.3)',
  },
  transitions: {
    fast: '150ms ease',
    base: '200ms ease',
    slow: '300ms ease',
  },
} as const;

export type Theme = typeof theme;
