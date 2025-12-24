import cn from 'classnames';
import {useState, useRef, useEffect} from "react";

import cls from "@/widgets/settings/Settings.Widget.module.css";

import Input from "@/components/base/Input.tsx";
import useSettings from "@/app/settings/state.ts";
import {useCredentialsStore} from "@/app/settings/creds.ts";

export default function SettingsWidget() {
    const settings = useSettings();

    const [isHidden, setIsHidden] = useState(true);

    const containerRef = useRef(null);

    useEffect(() => {
        const handleKeyDown = (event: KeyboardEvent) => {
            if (event.key === "Escape") {
                event.preventDefault();
                setIsHidden(true);
            }
        };

        const handleClickOutside = (event: MouseEvent) => {
            if (containerRef.current &&
                // @ts-ignore
                !containerRef.current.contains(event.target as Node)) {
                setIsHidden(true);
            }
        };

        document.addEventListener("keydown", handleKeyDown);
        document.addEventListener("mousedown", handleClickOutside);

        return () => {
            document.removeEventListener("keydown", handleKeyDown);
            document.removeEventListener("mousedown", handleClickOutside);
        };
    }, []);

    // Toggle visibility with Ctrl+. or Cmd+.
    useEffect(() => {
        const handleKeyDown = (event: KeyboardEvent) => {
            if ((event.ctrlKey || event.metaKey) && event.key === ".") {
                event.preventDefault();
                setIsHidden(prev => !prev);
            }
        };

        document.addEventListener("keydown", handleKeyDown);
        return () => {
            document.removeEventListener("keydown", handleKeyDown);
        };
    }, []);

    const credStore = useCredentialsStore();
    return (
        <div
            ref={containerRef}
            className={cn(cls.SettingsContainer, {
            [cls.hidden]: isHidden,
        })}>
            <div className={cls.SettingsClause}>
                <p>Backend url:</p>
                <Input
                    onChange={(v) => {
                        credStore.setUrl(v)
                        settings.setBackendUrl(v)
                    }}
                    inputValue={credStore.url}
                />
            </div>

            <div className={cls.SettingsClause}>
                <p>Auth header:</p>
                <Input
                    onChange={(v) => {
                        credStore.setToken(v)
                        settings.setAuthHeader(v)
                    }}
                    inputValue={credStore.token}
                />
            </div>
        </div>
    )
}
