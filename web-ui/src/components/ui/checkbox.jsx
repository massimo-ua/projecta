import React from 'react';
import PropTypes from 'prop-types';
import { Check } from 'lucide-react';
import { cn } from '@/lib/utils';

export function Checkbox({
  checked = false,
  onCheckedChange,
  disabled = false,
  className,
  id,
  'aria-label': ariaLabel,
}) {
  return (
    <button
      type="button"
      role="checkbox"
      id={id}
      aria-label={ariaLabel}
      aria-checked={checked}
      disabled={disabled}
      onClick={() => onCheckedChange && onCheckedChange(!checked)}
      className={cn(
        'peer h-4 w-4 shrink-0 rounded-md border ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 transition-colors flex items-center justify-center cursor-pointer',
        checked
          ? 'bg-primary border-primary text-primary-foreground'
          : 'bg-background border-border hover:border-primary/70',
        className,
      )}
    >
      {checked && <Check className="h-3 w-3 stroke-[3]" />}
    </button>
  );
}

Checkbox.propTypes = {
  checked: PropTypes.bool,
  onCheckedChange: PropTypes.func,
  disabled: PropTypes.bool,
  className: PropTypes.string,
  id: PropTypes.string,
  'aria-label': PropTypes.string,
};

Checkbox.defaultProps = {
  checked: false,
  onCheckedChange: undefined,
  disabled: false,
  className: '',
  id: undefined,
  'aria-label': undefined,
};

export default Checkbox;
