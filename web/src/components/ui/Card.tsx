import { ReactNode, CSSProperties } from 'react';
import { theme } from '../../styles/theme';

interface CardProps {
  children?: ReactNode;
  title?: string;
  subtitle?: string;
  icon?: string;
  variant?: 'default' | 'elevated' | 'gradient' | 'purple';
  className?: string;
  style?: CSSProperties;
  onClick?: () => void;
}

const variantStyles: Record<string, CSSProperties> = {
  default: {
    background: theme.colors.background.secondary,
    border: `1px solid ${theme.colors.border.primary}`,
  },
  elevated: {
    background: theme.colors.background.elevated,
    border: `1px solid ${theme.colors.border.primary}`,
  },
  gradient: {
    background: 'linear-gradient(135deg, rgba(240, 185, 11, 0.15) 0%, rgba(252, 213, 53, 0.05) 100%)',
    border: '1px solid rgba(240, 185, 11, 0.2)',
    boxShadow: '0 0 30px rgba(240, 185, 11, 0.15)',
  },
  purple: {
    background: 'linear-gradient(135deg, rgba(139, 92, 246, 0.15) 0%, rgba(99, 102, 241, 0.1) 50%, rgba(30, 35, 41, 0.8) 100%)',
    border: '1px solid rgba(139, 92, 246, 0.3)',
    boxShadow: '0 8px 32px rgba(139, 92, 246, 0.2)',
  },
};

export function Card({
  children,
  title,
  subtitle,
  icon,
  variant = 'default',
  className = '',
  style,
  onClick,
}: CardProps) {
  const cardStyle: CSSProperties = {
    ...variantStyles[variant],
    borderRadius: theme.radius['2xl'],
    padding: theme.spacing.xl,
    ...style,
  };

  return (
    <div
      className={`${onClick ? 'cursor-pointer hover:scale-[1.01]' : ''} transition-all ${className}`}
      style={cardStyle}
      onClick={onClick}
    >
      {(title || icon) && (
        <div className="mb-4">
          {icon && title && (
            <div className="flex items-center gap-3 mb-2">
              <div
                className="flex items-center justify-center text-2xl"
                style={{
                  width: '3rem',
                  height: '3rem',
                  borderRadius: theme.radius.xl,
                  background: theme.colors.brand.gradient,
                  boxShadow: theme.shadows.brand,
                  border: '2px solid rgba(255, 255, 255, 0.1)',
                }}
              >
                {icon}
              </div>
              <div>
                <h3
                  className="text-2xl font-bold"
                  style={{
                    color: theme.colors.text.primary,
                    textShadow: '0 2px 8px rgba(240, 185, 11, 0.3)',
                  }}
                >
                  {title}
                </h3>
                {subtitle && (
                  <p style={{ color: theme.colors.brand.primary }}>{subtitle}</p>
                )}
              </div>
            </div>
          )}
          {!icon && title && (
            <h3 className="text-xl font-bold" style={{ color: theme.colors.text.primary }}>
              {title}
            </h3>
          )}
        </div>
      )}
      {children}
    </div>
  );
}
