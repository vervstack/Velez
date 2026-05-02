# T08 — SectionLabel

**Files to create:**

- `src/components/base/SectionLabel.tsx`
- `src/components/base/SectionLabel.module.css`

## What it looks like

A small all-caps monospace label used as a section heading inside panels and sidebars. No surrounding elements.

Example:

```
NODES
NODE HEALTH
VERVSTACK PLUGINS
```

## Props interface

```ts
interface SectionLabelProps {
    children: string;
}
```

## Component

```tsx
// src/components/base/SectionLabel.tsx
import cls from '@/components/base/SectionLabel.module.css';

interface SectionLabelProps {
    children: string;
}

export default function SectionLabel({ children }: SectionLabelProps) {
    return <div className={cls.SectionLabelContainer}>{children}</div>;
}
```

## CSS

```css
/* src/components/base/SectionLabel.module.css */
.SectionLabelContainer {
    font-family: var(--font-mono);
    font-size: 0.5625rem;   /* 9px */
    color: var(--fg-dim);
    letter-spacing: 0.1em;
    text-transform: uppercase;
}
```
