# T04 — MiniBar (progress bar)

**Files to create:**

- `src/components/base/MiniBar.tsx`
- `src/components/base/MiniBar.module.css`

## What it looks like

A thin 3px-tall bar with a filled segment. Color changes automatically by threshold:

- Above 80% → red (`var(--red)`)
- Above 60% → amber (`var(--amber)`)
- Otherwise → the passed `color` prop (default `var(--cyan)`)

The fill animates smoothly on value change.

## Props interface

```ts
interface MiniBarProps {
    val: number;
    max?: number;   // default 100
    color?: string; // default "var(--cyan)"
}
```

## Component

```tsx
// src/components/base/MiniBar.tsx
import cls from '@/components/base/MiniBar.module.css';

interface MiniBarProps {
    val: number;
    max?: number;
    color?: string;
}

export default function MiniBar({ val, max = 100, color = 'var(--cyan)' }: MiniBarProps) {
    const pct = Math.min(100, (val / max) * 100);
    const fillColor = pct > 80 ? 'var(--red)' : pct > 60 ? 'var(--amber)' : color;

    return (
        <div className={cls.MiniBarContainer}>
            <div
                className={cls.fill}
                style={{ width: `${pct}%`, background: fillColor }}
            />
        </div>
    );
}
```

## CSS

```css
/* src/components/base/MiniBar.module.css */
.MiniBarContainer {
    height: 0.1875rem; /* 3px */
    background: var(--bg4);
    border-radius: 99px;
    overflow: hidden;
    width: 100%;
}

.fill {
    height: 100%;
    border-radius: 99px;
    transition: width 0.6s ease;
}
```

## Notes

- The `background` on `.fill` is inline because it's a runtime-computed color (threshold logic). Width is also inline (
  computed from `val`).
- `max` defaults to 100, but for memory bars the caller should pass `max={2048}` (MiB).
