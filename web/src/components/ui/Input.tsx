import { InputHTMLAttributes } from 'react';
import { theme } from '../../styles/theme';

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  hint?: string;
  error?: string;
  fullWidth?: boolean;
}

export function Input({
  label,
  hint,
  error,
  fullWidth = false,
  className = '',
  style,
  ...props
}: InputProps) {
  const inputStyle: React.CSSProperties = {
    background: theme.colors.background.primary,
    border: `1px solid ${error ? theme.colors.error.main : theme.colors.border.primary}`,
    color: theme.colors.text.primary,
    borderRadius: theme.radius.lg,
    padding: '0.5rem 1rem',
    width: fullWidth ? '100%' : 'auto',
    outline: 'none',
    transition: theme.transitions.base,
    ...style,
  };

  return (
    <div className={`${fullWidth ? 'w-full' : ''} ${className}`}>
      {label && (
        <label
          className="block text-sm mb-2"
          style={{ color: theme.colors.text.secondary }}
        >
          {label}
        </label>
      )}
      <input
        className="focus:ring-2"
        style={inputStyle}
        {...props}
      />
      {hint && !error && (
        <div
          className="text-xs mt-1"
          style={{ color: theme.colors.text.secondary }}
        >
          {hint}
        </div>
      )}
      {error && (
        <div className="text-xs mt-1" style={{ color: theme.colors.error.main }}>
          {error}
        </div>
      )}
    </div>
  );
}

interface TextAreaProps extends InputHTMLAttributes<HTMLTextAreaElement> {
  label?: string;
  hint?: string;
  error?: string;
  fullWidth?: boolean;
  rows?: number;
}

export function TextArea({
  label,
  hint,
  error,
  fullWidth = false,
  rows = 4,
  className = '',
  style,
  ...props
}: TextAreaProps) {
  const textareaStyle: React.CSSProperties = {
    background: theme.colors.background.primary,
    border: `1px solid ${error ? theme.colors.error.main : theme.colors.border.primary}`,
    color: theme.colors.text.primary,
    borderRadius: theme.radius.lg,
    padding: '0.75rem 1rem',
    width: fullWidth ? '100%' : 'auto',
    outline: 'none',
    resize: 'vertical' as const,
    fontFamily: 'monospace',
    fontSize: '0.875rem',
    lineHeight: '1.5',
    transition: theme.transitions.base,
    ...style,
  };

  return (
    <div className={`${fullWidth ? 'w-full' : ''} ${className}`}>
      {label && (
        <label
          className="block text-sm mb-2"
          style={{ color: theme.colors.text.secondary }}
        >
          {label}
        </label>
      )}
      <textarea
        className="focus:ring-2"
        style={textareaStyle}
        rows={rows}
        {...props}
      />
      {hint && !error && (
        <div
          className="text-xs mt-1"
          style={{ color: theme.colors.text.secondary }}
        >
          {hint}
        </div>
      )}
      {error && (
        <div className="text-xs mt-1" style={{ color: theme.colors.error.main }}>
          {error}
        </div>
      )}
    </div>
  );
}
