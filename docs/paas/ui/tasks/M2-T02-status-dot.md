# T02 — StatusDot

**Files to create:**

- `src/components/base/StatusDot.tsx`
- `src/components/base/StatusDot.module.css`

## What it looks like

A small filled circle (7×7px) whose color maps to a service/node status. When `pulse` is true and status is `running`,
it has a breathing CSS animation.

Colors:

- `running` → `var(--green)`
- `online`  → `var(--green)`
- `degraded` → `var(--amber)`
- `stopped`  → `var(--fg-faint)`
- `offline`  → `var(--red)`
- anything else → `var(--fg-dim)`

## Props interface

```ts
type StatusDotStatus = 'running' | 'degraded' | 'stopped' | 'online' | 'offline';

interface StatusDotProps {
    status: StatusDotStatus;
    pulse?: boolean;
}
```

## Component

```tsx
// src/components/base/StatusDot.tsx
import cls from '@/components/base/StatusDot.module.css';
import cn from 'classnames';

interface StatusDotProps {
    status: 'running' | 'degraded' | 'stopped' | 'online' | 'offline';
    pulse?: boolean;
}

export default function StatusDot({ status, pulse }: StatusDotProps) {
    return (
        <span
            className={cn(
                cls.StatusDotContainer,
                cls[status],
                { [cls.pulse]: pulse && status === 'running' }
            )}
        />
    );
}
```

## CSS

```css
/* src/components/base/StatusDot.module.css */
.StatusDotContainer {
    display: inline-block;
    width: 0.4375rem;   /* 7px */
    height: 0.4375rem;
    border-radius: 50%;
    flex-shrink: 0;
    background: var(--fg-dim);
}

.running  { background: var(--green); }
.online   { background: var(--green); }
.degraded { background: var(--amber); }
.stopped  { background: var(--fg-faint); }
.offline  { background: var(--red); }

.pulse {
    animation: pulse 2.2s ease-in-out infinite;
}
```

## Notes

- The `pulse` animation is defined globally in `src/index.module.css` via `@keyframes pulse`.
- Use `classnames` (already installed as `cn`) for conditional classes.
- Do not hardcode colors inline — always use CSS variables.
