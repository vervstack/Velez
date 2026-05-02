# T03 — Badge

**Files to create:**

- `src/components/base/Badge.tsx`
- `src/components/base/Badge.module.css`

## What it looks like

A small mono-font label with a colored border and a semi-transparent background. Used for status text like "degraded", "
enabled", "disabled", "active".

Example from the design:

```
┌──────────────┐
│  degraded    │   ← amber border + text, dim amber bg
└──────────────┘
```

## Props interface

```ts
interface BadgeProps {
    label: string;
    /** CSS color value, e.g. "var(--green)" or "#28c840" */
    color?: string;
    /** CSS background value for the fill. Defaults to transparent. */
    dim?: string;
}
```

## Component

```tsx
// src/components/base/Badge.tsx
import cls from '@/components/base/Badge.module.css';

interface BadgeProps {
    label: string;
    color?: string;
    dim?: string;
}

export default function Badge({ label, color, dim }: BadgeProps) {
    return (
        <span
            className={cls.BadgeContainer}
            style={{
                color: color ?? 'var(--fg-dim)',
                borderColor: color ? color + '44' : 'var(--border)',
                background: dim ?? 'transparent',
            }}
        >
            {label}
        </span>
    );
}
```

## CSS

```css
/* src/components/base/Badge.module.css */
.BadgeContainer {
    font-family: var(--font-mono);
    font-size: 0.625rem;   /* 10px */
    border: 1px solid;
    border-radius: 0.25rem;
    padding: 0.0625rem 0.4375rem; /* 1px 7px */
    flex-shrink: 0;
    letter-spacing: 0.04em;
    white-space: nowrap;
}
```

## Notes

- `color` and `dim` are intentionally inline here because Badge is a generic color-parameterized atom — CSS Modules
  alone can't express arbitrary runtime color values. This is the one allowed exception per the project rules (pure
  atoms without business logic may take color params).
- The `44` appended to `color` for `borderColor` produces a ~27% opacity hex suffix that works for any 6-digit hex
  color.
