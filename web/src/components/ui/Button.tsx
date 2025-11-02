import { ButtonHTMLAttributes, ReactNode } from 'react';
import { theme } from '../../styles/theme';

type ButtonVariant = 'primary' | 'secondary' | 'success' | 'danger' | 'ghost' | 'purple';
type ButtonSize = 'sm' | 'md' | 'lg';

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  size?: ButtonSize;
  children: ReactNode;
  isLoading?: boolean;
  fullWidth?: boolean;
}

const variantStyles: Record<ButtonVariant, React.CSSProperties> = {
  primary: {
    background: theme.colors.brand.gradient,
    color: '#000',
    border: 'none',
    boxShadow: theme.shadows.brand,
  },
  secondary: {
    background: theme.colors.background.secondary,
    color: theme.colors.text.secondary,
    border: `1px solid ${theme.colors.border.primary}`,
  },
  success: {
    background: 'linear-gradient(135deg, #10B981 0%, #0ECB81 100%)',
    color: '#FFFFFF',
    border: 'none',
    boxShadow: theme.shadows.success,
  },
  danger: {
    background: theme.colors.error.light,
    color: '#FCA5A5',
    border: `1px solid ${theme.colors.error.border}`,
  },
  ghost: {
    background: 'transparent',
    color: theme.colors.text.secondary,
    border: 'none',
  },
  purple: {
    background: theme.colors.purple.gradient,
    color: '#FFF',
    border: 'none',
    boxShadow: theme.shadows.purple,
  },
};

const sizeStyles: Record<ButtonSize, React.CSSProperties> = {
  sm: {
    padding: '0.375rem 0.75rem',
    fontSize: '0.875rem',
  },
  md: {
    padding: '0.5rem 1.5rem',
    fontSize: '1rem',
  },
  lg: {
    padding: '0.75rem 2rem',
    fontSize: '1.125rem',
  },
};

export function Button({
  variant = 'primary',
  size = 'md',
  children,
  isLoading = false,
  fullWidth = false,
  disabled,
  style,
  className = '',
  ...props
}: ButtonProps) {
  const buttonStyle: React.CSSProperties = {
    ...variantStyles[variant],
    ...sizeStyles[size],
    fontWeight: 'bold',
    borderRadius: theme.radius.xl,
    cursor: disabled || isLoading ? 'not-allowed' : 'pointer',
    opacity: disabled || isLoading ? 0.5 : 1,
    transition: theme.transitions.base,
    width: fullWidth ? '100%' : 'auto',
    ...style,
  };

  return (
    <button
      className={`transition-all hover:scale-105 active:scale-95 ${className}`}
      style={buttonStyle}
      disabled={disabled || isLoading}
      {...props}
    >
      {isLoading ? '⏳ 加载中...' : children}
    </button>
  );
}
