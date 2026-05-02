# T06 — IconButton

**Files to create/replace:**

- `src/components/base/IconButton.tsx`
- `src/components/base/IconButton.module.css`

> Note: The existing `src/components/base/ActionButton.tsx` serves a different purpose (icon-only). This is a new small
> text button used for node/service actions like "shell", "drain", "restart". Do not delete ActionButton.

## What it looks like

A compact bordered button with a short text label (e.g. "shell", "drain", "logs"). On hover it gets a dim background.
Danger variant uses red colors.

Design reference:

```
┌──────────┐
│  shell   │   ← default: dim border, fgMid text; on hover: cyan-dim bg, cyan-dim border
└──────────┘

┌──────────┐
│  drain   │   ← danger: dim red border, red text; on hover: red-dim bg
└──────────┘
```

## Props interface

```ts
interface IconButtonProps {
    label: string;
    title?: string;
    onClick?: () => void;
    danger?: boolean;
    disabled?: boolean;
}
```

## Component

```tsx
// src/components/base/IconButton.tsx
import cls from '@/components/base/IconButton.module.css';
import cn from 'classnames';

interface IconButtonProps {
    label: string;
    title?: string;
    onClick?: () => void;
    danger?: boolean;
    disabled?: boolean;
}

export default function IconButton({ label, title, onClick, danger, disabled }: IconButtonProps) {
    return (
        <button
            className={cn(cls.IconButtonContainer, { [cls.danger]: danger })}
            title={title}
            onClick={onClick}
            disabled={disabled}
        >
            {label}
        </button>
    );
}
```

## CSS

```css
/* src/components/base/IconButton.module.css */
.IconButtonContainer {
    background: transparent;
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    color: var(--fg-mid);
    font-family: var(--font-mono);
    font-size: 0.6875rem; /* 11px */
    padding: 0.25rem 0.5rem; /* 4px 8px */
    cursor: pointer;
    transition: background 0.15s, border-color 0.15s, color 0.15s;

    &:hover {
        background: var(--cyan-dim);
        border-color: var(--border-m);
    }

    &:disabled {
        opacity: 0.4;
        cursor: not-allowed;
    }
}

.danger {
    color: var(--red);

    &:hover {
        background: var(--red-dim);
        border-color: rgba(237, 47, 50, 0.44);
    }
}
```
