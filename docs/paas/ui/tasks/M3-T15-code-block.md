# T15 — CodeBlock

**Files to create:**

- `src/components/complex/CodeBlock/CodeBlock.tsx`
- `src/components/complex/CodeBlock/CodeBlock.module.css`

## What it looks like

A panel with an optional title bar and a monospaced code body. The title bar has a "copy" button on the right. Used on
the VCN page to display the network config snippet.

```
┌──────────────────────────────────────────────┐
│  vcn config — node01                  [copy] │  ← title bar (bg3)
├──────────────────────────────────────────────┤
│  # /etc/velez/vcn.yaml                       │
│  network:                                    │
│    mode: mesh                                │
│    subnet: 10.100.0.0/24                     │
└──────────────────────────────────────────────┘
```

Basic syntax coloring (no library):

- Lines starting with `#` → dim fg (comment)
- Keys before `:` → cyan
- Values after `:` → fg (default) or green for bare words (`mesh`, `wireguard`)

## Props interface

```ts
interface CodeBlockProps {
    title?: string;
    code: string;       // raw multi-line string
}
```

## Component

Use a simple line-by-line renderer with regex to colorize without a syntax library.

```tsx
// src/components/complex/CodeBlock/CodeBlock.tsx
import { useState } from 'react';
import cls from '@/components/complex/CodeBlock/CodeBlock.module.css';

interface CodeBlockProps {
    title?: string;
    code: string;
}

function tokenizeLine(line: string): React.ReactNode {
    if (line.trim().startsWith('#')) {
        return <span className={cls.comment}>{line}</span>;
    }
    const colonIdx = line.indexOf(':');
    if (colonIdx === -1) return <span>{line}</span>;

    const key = line.slice(0, colonIdx);
    const rest = line.slice(colonIdx + 1);

    const bareWordMatch = rest.match(/^\s+([a-z][a-z0-9-]+)\s*$/);
    const valueNode = bareWordMatch
        ? <><span>: </span><span className={cls.valueWord}>{bareWordMatch[1]}</span></>
        : <span>{':' + rest}</span>;

    return <><span className={cls.key}>{key}</span>{valueNode}</>;
}

export default function CodeBlock({ title, code }: CodeBlockProps) {
    const [copied, setCopied] = useState(false);

    function handleCopy() {
        navigator.clipboard.writeText(code).then(function onCopied() {
            setCopied(true);
            setTimeout(function resetCopied() { setCopied(false); }, 1500);
        });
    }

    const lines = code.split('\n');

    return (
        <div className={cls.CodeBlockContainer}>
            {title && (
                <div className={cls.titleBar}>
                    <span className={cls.titleText}>{title}</span>
                    <button className={cls.copyBtn} onClick={handleCopy}>
                        {copied ? 'copied' : 'copy'}
                    </button>
                </div>
            )}
            <div className={cls.body}>
                {lines.map(function renderLine(line, i) {
                    return (
                        <div key={i} className={cls.line}>
                            {tokenizeLine(line)}
                        </div>
                    );
                })}
            </div>
        </div>
    );
}
```

## CSS

```css
/* src/components/complex/CodeBlock/CodeBlock.module.css */
.CodeBlockContainer {
    background: var(--bg2);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    overflow: hidden;
}

.titleBar {
    display: flex;
    align-items: center;
    padding: 0.75rem 1.125rem;
    border-bottom: 1px solid var(--border);
    background: var(--bg3);
}

.titleText {
    font-family: var(--font-mono);
    font-size: 0.625rem;
    color: var(--fg-dim);
    flex: 1;
}

.copyBtn {
    font-family: var(--font-mono);
    font-size: 0.625rem;
    color: var(--cyan);
    background: transparent;
    border: none;
    cursor: pointer;
    transition: opacity 0.12s;

    &:hover {
        opacity: 0.7;
    }
}

.body {
    padding: 1rem 1.25rem;
}

.line {
    font-family: var(--font-mono);
    font-size: 0.71875rem; /* 11.5px */
    color: var(--fg-mid);
    line-height: 1.9;
    white-space: pre;
}

/* Token colors */
.comment    { color: var(--fg-dim); }
.key        { color: var(--cyan); }
.valueWord  { color: var(--green); }
```
