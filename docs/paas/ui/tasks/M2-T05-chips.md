# T05 — Status Chips (EnvChip, IncidentChip, FreezeChip)

**Files to create:**

- `src/components/base/chips/EnvChip.tsx`
- `src/components/base/chips/IncidentChip.tsx`
- `src/components/base/chips/FreezeChip.tsx`
- `src/components/base/chips/chips.module.css`

## What they look like

Pill-shaped labels with a small icon or dot on the left. Used inside service cards to communicate env, incident state,
and release-freeze state.

### EnvChip

- `prod` → cyan border + text, small cyan dot
- `stage` → dim border + text, small dim dot

### IncidentChip

- Red border + text
- Warning triangle SVG icon on the left

### FreezeChip

- Blue border + text
- Snowflake SVG icon on the left (4 lines + 4 endpoints circles)

---

## EnvChip

```tsx
// src/components/base/chips/EnvChip.tsx
import cls from '@/components/base/chips/chips.module.css';
import cn from 'classnames';

interface EnvChipProps {
    env: string;
}

export default function EnvChip({ env }: EnvChipProps) {
    const isProd = env === 'prod';
    return (
        <span className={cn(cls.ChipContainer, isProd ? cls.envProd : cls.envStage)}>
            <span className={cn(cls.dot, isProd ? cls.dotProd : cls.dotStage)} />
            {env}
        </span>
    );
}
```

## IncidentChip

```tsx
// src/components/base/chips/IncidentChip.tsx
import cls from '@/components/base/chips/chips.module.css';

export default function IncidentChip() {
    return (
        <span className={cls.ChipContainer + ' ' + cls.incident}>
            <svg width="10" height="10" viewBox="0 0 10 10" fill="none" className={cls.icon}>
                <path
                    d="M5 1.2L9.2 8.8H0.8L5 1.2Z"
                    stroke="var(--red)"
                    strokeWidth="1.2"
                    strokeLinejoin="round"
                    fill="rgba(237,47,50,0.18)"
                />
                <line x1="5" y1="4.2" x2="5" y2="6.4" stroke="var(--red)" strokeWidth="1.1" strokeLinecap="round"/>
                <circle cx="5" cy="7.5" r="0.55" fill="var(--red)"/>
            </svg>
            incident
        </span>
    );
}
```

## FreezeChip

```tsx
// src/components/base/chips/FreezeChip.tsx
import cls from '@/components/base/chips/chips.module.css';

export default function FreezeChip() {
    return (
        <span className={cls.ChipContainer + ' ' + cls.freeze}>
            <svg width="10" height="10" viewBox="0 0 10 10" fill="none" className={cls.icon}>
                <line x1="5" y1="1" x2="5" y2="9" stroke="var(--blue)" strokeWidth="1.2" strokeLinecap="round"/>
                <line x1="1" y1="5" x2="9" y2="5" stroke="var(--blue)" strokeWidth="1.2" strokeLinecap="round"/>
                <line x1="2.05" y1="2.05" x2="7.95" y2="7.95" stroke="var(--blue)" strokeWidth="1.2" strokeLinecap="round"/>
                <line x1="7.95" y1="2.05" x2="2.05" y2="7.95" stroke="var(--blue)" strokeWidth="1.2" strokeLinecap="round"/>
                <circle cx="5" cy="1" r="0.9" fill="var(--blue)"/>
                <circle cx="5" cy="9" r="0.9" fill="var(--blue)"/>
                <circle cx="1" cy="5" r="0.9" fill="var(--blue)"/>
                <circle cx="9" cy="5" r="0.9" fill="var(--blue)"/>
            </svg>
            freeze
        </span>
    );
}
```

## CSS

```css
/* src/components/base/chips/chips.module.css */

.ChipContainer {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    font-family: var(--font-mono);
    font-size: 0.625rem;   /* 10px */
    font-weight: 500;
    padding: 0.125rem 0.4375rem; /* 2px 7px */
    border-radius: 99px;
    border: 1px solid;
    letter-spacing: 0.03em;
    white-space: nowrap;
    flex-shrink: 0;
}

/* EnvChip variants */
.envProd {
    background: var(--cyan-dim);
    border-color: rgba(10, 183, 238, 0.30);
    color: var(--cyan);
}

.envStage {
    background: rgba(114, 114, 138, 0.12);
    border-color: rgba(114, 114, 138, 0.25);
    color: var(--fg-dim);
}

.dot {
    width: 0.3125rem;  /* 5px */
    height: 0.3125rem;
    border-radius: 50%;
    display: block;
    opacity: 0.7;
}

.dotProd  { background: var(--cyan); }
.dotStage { background: var(--fg-dim); }

/* IncidentChip */
.incident {
    background: var(--red-dim);
    border-color: rgba(237, 47, 50, 0.30);
    color: var(--red);
}

/* FreezeChip */
.freeze {
    background: var(--blue-dim);
    border-color: rgba(96, 165, 250, 0.28);
    color: var(--blue);
}

.icon {
    flex-shrink: 0;
}
```

## Notes

- All three chips share the same CSS module to avoid repetition.
- SVG strokes use inline CSS variable references (`var(--red)`, `var(--blue)`) because SVG attributes don't support CSS
  Modules class names.
- `--blue` and `--blue-dim` must be defined in T01 (design tokens).
