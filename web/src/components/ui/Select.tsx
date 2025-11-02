import { SelectHTMLAttributes } from 'react';
import { theme } from '../../styles/theme';

interface SelectProps extends SelectHTMLAttributes<HTMLSelectElement> {
  label?: string;
  hint?: string;
  error?: string;
  fullWidth?: boolean;
  options?: { value: string; label: string }[];
}

export function Select({
  label,
  hint,
  error,
  fullWidth = false,
  options = [],
  className = '',
  style,
  children,
  ...props
}: SelectProps) {
  const selectStyle: React.CSSProperties = {
    background: theme.colors.background.primary,
    border: `1px solid ${error ? theme.colors.error.main : theme.colors.border.primary}`,
    color: theme.colors.text.primary,
    borderRadius: theme.radius.lg,
    padding: '0.5rem 1rem',
    width: fullWidth ? '100%' : 'auto',
    outline: 'none',
    cursor: 'pointer',
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
      <select
        className="focus:ring-2"
        style={selectStyle}
        {...props}
      >
        {options.length > 0
          ? options.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))
          : children}
      </select>
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
