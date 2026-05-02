# T10 — ThreeDotMenu

**Files to create:**

- `src/components/complex/ThreeDotMenu/ThreeDotMenu.tsx`
- `src/components/complex/ThreeDotMenu/ThreeDotMenu.module.css`

## What it looks like

A 26×26 button showing `···` that toggles a dropdown menu. The menu appears below-right of the trigger. Clicking outside
closes it.

```
[···]
     ┌──────────────┐
     │ ↺ Restart    │
     │ ◼ Stop       │  ← red
     │ ≈ View logs  │
     │ $ Shell      │
     └──────────────┘
```

## Props interface

```ts
interface MenuAction {
    label: string;
    icon: string;
    danger?: boolean;
    onClick?: () => void;
}

interface ThreeDotMenuProps {
    actions: MenuAction[];
}
```

## Component

```tsx
// src/components/complex/ThreeDotMenu/ThreeDotMenu.tsx
import { useState, useEffect, useRef } from 'react';
import cls from '@/components/complex/ThreeDotMenu/ThreeDotMenu.module.css';
import cn from 'classnames';

interface MenuAction {
    label: string;
    icon: string;
    danger?: boolean;
    onClick?: () => void;
}

interface ThreeDotMenuProps {
    actions: MenuAction[];
}

export default function ThreeDotMenu({ actions }: ThreeDotMenuProps) {
    const [open, setOpen] = useState(false);
    const ref = useRef<HTMLDivElement>(null);

    useEffect(function bindOutsideClick() {
        if (!open) return;

        function handleClickOutside(e: MouseEvent) {
            if (ref.current && !ref.current.contains(e.target as Node)) {
                setOpen(false);
            }
        }

        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, [open]);

    function handleTriggerClick(e: React.MouseEvent) {
        e.stopPropagation();
        setOpen(prev => !prev);
    }

    return (
        <div ref={ref} className={cls.ThreeDotMenuContainer}>
            <button
                className={cn(cls.trigger, { [cls.triggerOpen]: open })}
                onClick={handleTriggerClick}
                aria-label="Open menu"
            >
                ···
            </button>

            {open && (
                <div className={cls.dropdown}>
                    {actions.map(function renderAction(action, i) {
                        function handleActionClick() {
                            setOpen(false);
                            action.onClick?.();
                        }

                        return (
                            <button
                                key={i}
                                className={cn(cls.action, { [cls.actionDanger]: action.danger })}
                                onClick={handleActionClick}
                            >
                                <span className={cls.actionIcon}>{action.icon}</span>
                                {action.label}
                            </button>
                        );
                    })}
                </div>
            )}
        </div>
    );
}
```

## CSS

```css
/* src/components/complex/ThreeDotMenu/ThreeDotMenu.module.css */
.ThreeDotMenuContainer {
    position: relative;
    flex-shrink: 0;
}

.trigger {
    background: transparent;
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    color: var(--fg-dim);
    width: 1.625rem;  /* 26px */
    height: 1.625rem;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    font-size: 0.875rem;
    letter-spacing: -0.0625rem;
    font-family: serif;
    transition: background 0.12s, border-color 0.12s;

    &:hover,
    &.triggerOpen {
        background: var(--bg4);
        border-color: var(--border-m);
    }
}

.dropdown {
    position: absolute;
    top: calc(100% + 0.25rem);
    right: 0;
    background: var(--bg3);
    border: 1px solid var(--border-m);
    border-radius: 0.5625rem; /* 9px */
    padding: 0.25rem;
    min-width: 8.75rem; /* 140px */
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.5);
    animation: fadeUp 0.12s ease both;
}

.action {
    display: flex;
    align-items: center;
    gap: 0.5625rem; /* 9px */
    width: 100%;
    background: transparent;
    border: none;
    border-radius: var(--radius-sm);
    padding: 0.4375rem 0.625rem; /* 7px 10px */
    cursor: pointer;
    color: var(--fg-mid);
    font-family: var(--font-mono);
    font-size: 0.71875rem; /* 11.5px */
    text-align: left;
    transition: background 0.1s;

    &:hover {
        background: var(--bg4);
    }
}

.actionDanger {
    color: var(--red);

    &:hover {
        background: var(--red-dim);
    }
}

.actionIcon {
    font-size: 0.75rem;
    opacity: 0.7;
    width: 0.875rem;
    text-align: center;
    flex-shrink: 0;
}
```

## Notes

- The `fadeUp` keyframe animation is defined globally in `src/index.module.css`.
- `e.stopPropagation()` on the trigger prevents cards from reacting to the menu click.
- Actions are rendered with a `key={i}` here since labels are stable, but callers should ensure uniqueness.
