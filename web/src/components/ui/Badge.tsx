import { ReactNode } from 'react';
import { theme } from '../../styles/theme';

type BadgeVariant = 'success' | 'error' | 'warning' | 'info' | 'default' | 'purple';
type BadgeSize = 'sm' | 'md';

interface BadgeProps {
  children: ReactNode;
  variant?: BadgeVariant;
  size?: BadgeSize;
  className?: string;
}

const variantStyles: Record<BadgeVariant, React.CSSProperties> = {
  success: {
    background: theme.colors.success.light,
    color: theme.colors.success.main,
    border: `1px solid ${theme.colors.success.border}`,
  },
  error: {
    background: theme.colors.error.light,
    color: theme.colors.error.main,
    border: `1px solid ${theme.colors.error.border}`,
  },
  warning: {
    background: theme.colors.warning.light,
    color: theme.colors.warning.main,
    border: `1px solid ${theme.colors.warning.border}`,
  },
  info: {
    background: theme.colors.info.light,
    color: theme.colors.info.main,
    border: `1px solid ${theme.colors.info.border}`,
  },
  purple: {
    background: theme.colors.purple.light,
    color: theme.colors.purple.secondary,
    border: `1px solid ${theme.colors.purple.border}`,
  },
  default: {
    background: theme.colors.background.tertiary,
    color: theme.colors.text.secondary,
    border: `1px solid ${theme.colors.border.primary}`,
  },
};

const sizeStyles: Record<BadgeSize, React.CSSProperties> = {
  sm: {
    padding: '0.125rem 0.5rem',
    fontSize: '0.75rem',
  },
  md: {
    padding: '0.25rem 0.75rem',
    fontSize: '0.875rem',
  },
};

export function Badge({ children, variant = 'default', size = 'sm', className = '' }: BadgeProps) {
  const style: React.CSSProperties = {
    ...variantStyles[variant],
    ...sizeStyles[size],
    borderRadius: theme.radius.md,
    fontWeight: 'bold',
    display: 'inline-block',
  };

  return (
    <span className={className} style={style}>
      {children}
    </span>
  );
}
