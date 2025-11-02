import { ReactNode, useEffect } from 'react';
import { theme } from '../../styles/theme';

interface ModalProps {
  isOpen: boolean;
  onClose: () => void;
  title?: string;
  children: ReactNode;
  maxWidth?: 'sm' | 'md' | 'lg' | 'xl' | '2xl' | '4xl';
  footer?: ReactNode;
}

const maxWidthClasses = {
  sm: 'max-w-sm',
  md: 'max-w-md',
  lg: 'max-w-lg',
  xl: 'max-w-xl',
  '2xl': 'max-w-2xl',
  '4xl': 'max-w-4xl',
};

export function Modal({
  isOpen,
  onClose,
  title,
  children,
  maxWidth = '2xl',
  footer,
}: ModalProps) {
  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = 'unset';
    }
    return () => {
      document.body.style.overflow = 'unset';
    };
  }, [isOpen]);

  if (!isOpen) return null;

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-4"
      style={{ background: 'rgba(0, 0, 0, 0.8)' }}
      onClick={onClose}
    >
      <div
        className={`${maxWidthClasses[maxWidth]} w-full max-h-[90vh] overflow-y-auto rounded-2xl p-6`}
        style={{
          background: theme.colors.background.secondary,
          border: `1px solid ${theme.colors.purple.border}`,
          boxShadow: theme.shadows.xl,
        }}
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        {title && (
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-2xl font-bold" style={{ color: theme.colors.text.primary }}>
              {title}
            </h2>
            <button
              onClick={onClose}
              className="text-2xl hover:scale-110 transition-transform"
              style={{ color: theme.colors.text.secondary }}
            >
              ✕
            </button>
          </div>
        )}

        {/* Content */}
        <div>{children}</div>

        {/* Footer */}
        {footer && <div className="mt-6">{footer}</div>}
      </div>
    </div>
  );
}
