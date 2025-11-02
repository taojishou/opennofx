import { InputHTMLAttributes } from 'react';
import { theme } from '../../styles/theme';

interface SwitchProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'type'> {
  label?: string;
  description?: string;
}

export function Switch({ label, description, checked, className = '', ...props }: SwitchProps) {
  return (
    <label className={`flex items-center gap-3 cursor-pointer ${className}`}>
      <div className="relative">
        <input type="checkbox" checked={checked} className="sr-only peer" {...props} />
        <div
          className="w-11 h-6 rounded-full peer peer-focus:ring-2 peer-focus:ring-blue-300 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-0.5 after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all"
          style={{
            background: checked ? theme.colors.success.main : theme.colors.border.secondary,
          }}
        />
      </div>
      {(label || description) && (
        <div className="flex-1">
          {label && (
            <div className="font-semibold" style={{ color: theme.colors.text.primary }}>
              {label}
            </div>
          )}
          {description && (
            <div className="text-sm" style={{ color: theme.colors.text.secondary }}>
              {description}
            </div>
          )}
        </div>
      )}
    </label>
  );
}
