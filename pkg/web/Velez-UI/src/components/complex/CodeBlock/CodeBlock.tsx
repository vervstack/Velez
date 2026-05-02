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
