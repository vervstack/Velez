# T07 — Button (rebuild)

**Files to modify:**

- `src/components/base/Button.tsx`
- `src/components/base/Button.module.css`

## What it looks like

A general-purpose button with three variants:

- **primary** — solid red background, white text. Used for the main "Deploy" CTA.
- **secondary** — transparent with dim border; on hover gets `--bg3` background.
- **danger** — solid red-dim background, red text and border.

## Props interface

```ts
interface ButtonProps {
    children: React.ReactNode;
    variant?: 'primary' | 'secondary' | 'danger';
    onClick?: () => void;
    disabled?: boolean;
    type?: 'button' | 'submit' | 'reset';
}
```

## Component

```tsx
// src/components/base/Button.tsx
import cls from '@/components/base/Button.module.css';
import cn from 'classnames';

interface ButtonProps {
    children: React.ReactNode;
    variant?: 'primary' | 'secondary' | 'danger';
    onClick?: () => void;
    disabled?: boolean;
    type?: 'button' | 'submit' | 'reset';
}

export default function Button({
    children,
    variant = 'secondary',
    onClick,
    disabled,
    type = 'button',
}: ButtonProps) {
    return (
        <button
            className={cn(cls.ButtonContainer, cls[variant])}
            onClick={onClick}
            disabled={disabled}
            type={type}
        >
            {children}
        </button>
    );
}
```

## CSS

```css
/* src/components/base/Button.module.css */
.ButtonContainer {
    font-family: var(--font-sans);
    font-weight: 700;
    font-size: 0.75rem;  /* 12px */
    padding: 0.4375rem 1rem; /* 7px 16px */
    border-radius: var(--radius-sm);
    border: none;
    cursor: pointer;
    transition: opacity 0.15s, background 0.15s, border-color 0.15s;
    white-space: nowrap;

    &:disabled {
        opacity: 0.4;
        cursor: not-allowed;
    }
}

/* Primary — solid red CTA */
.primary {
    background: var(--red);
    color: #fff;
    border: none;

    &:hover:not(:disabled) {
        opacity: 0.82;
    }
}

/* Secondary — ghost */
.secondary {
    background: transparent;
    color: var(--fg-mid);
    border: 1px solid var(--border);

    &:hover:not(:disabled) {
        background: var(--bg3);
        border-color: var(--border-m);
    }
}

/* Danger */
.danger {
    background: var(--red-dim);
    color: var(--red);
    border: 1px solid rgba(237, 47, 50, 0.30);

    &:hover:not(:disabled) {
        background: rgba(237, 47, 50, 0.20);
        border-color: rgba(237, 47, 50, 0.50);
    }
}
```

## Notes

- Keep the existing props if they exist — add `variant` without removing anything in use.
- If the current Button.tsx has a different API, add the new variant system on top rather than replacing the whole prop
  set (to avoid breaking callers).
