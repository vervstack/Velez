import cn from "classnames";

import cls from "@/pages/service/widgets/TabButton.module.css";
import {ServiceTab, TabInfo} from "@/pages/service/widgets/tabs.ts";

interface Props {
    tab: TabInfo;
    isActive: boolean;
    setActiveTab: (s: ServiceTab) => void;
}

export default function TabButton({tab, isActive, setActiveTab}: Props) {
    function handleClick() {
        setActiveTab(tab.id);
    }

    return (
        <button
            className={cn(cls.TabButton, {[cls.tabActive]: isActive})}
            onClick={handleClick}
        >
            {tab.label}
        </button>
    );
}
