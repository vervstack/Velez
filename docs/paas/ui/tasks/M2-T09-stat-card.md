# T09 — StatCard

**Files to create:**

- `src/components/base/StatCard.tsx`
- `src/components/base/StatCard.module.css`

## What it looks like

A tile showing a large metric value above a label. Used in 4-column summary grids on the Control Plane and VCN pages.

```
┌────────────────┐
│  3             │  ← large number in accent color
│  Total nodes   │  ← small sans-serif label
└────────────────┘
```

## Props interface

```ts
interface StatCardProps {
    value: string | number;
    label: string;
    color?: string; // CSS color, e.g. "var(--cyan)" or "var(--green)". Default: "var(--cyan)"
}
```

## Component

```tsx
// src/components/base/StatCard.tsx
import cls from '@/components/base/StatCard.module.css';

interface StatCardProps {
    value: string | number;
    label: string;
    color?: string;
}

export default function StatCard({ value, label, color = 'var(--cyan)' }: StatCardProps) {
    return (
        <div className={cls.StatCardContainer}>
            <div className={cls.value} style={{ color }}>
                {value}
            </div>
            <div className={cls.label}>{label}</div>
        </div>
    );
}
```

## CSS

```css
/* src/components/base/StatCard.module.css */
.StatCardContainer {
    background: var(--bg2);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 1.125rem 1.25rem; /* 18px 20px */
}

.value {
    font-family: var(--font-mono);
    font-size: 1.625rem;  /* 26px */
    font-weight: 600;
    margin-bottom: 0.25rem;
}

.label {
    font-family: var(--font-sans);
    font-size: 0.75rem;  /* 12px */
    color: var(--fg-dim);
}
```

## Notes

- `color` is the one inline style here because it's a runtime-chosen accent — StatCard is used with different colors for
  different metrics.
- The component has no interaction state; it is purely presentational.
