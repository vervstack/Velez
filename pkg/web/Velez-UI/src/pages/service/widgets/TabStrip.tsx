import cls from "@/pages/service/widgets/TabStrip.module.css";
import {ServiceTab, TABS} from "@/pages/service/widgets/tabs.ts";
import TabButton from "@/pages/service/widgets/TabButton.tsx";

interface Props {
    activeTab: ServiceTab;
    setActiveTab: (s: ServiceTab) => void;
}

export default function TabStrip({activeTab, setActiveTab}: Props) {
    return (
        <div className={cls.TabStripContainer}>
            {TABS.map(function renderTab(t) {
                return (
                    <TabButton
                        key={t.id}
                        tab={t}
                        isActive={activeTab === t.id}
                        setActiveTab={setActiveTab}
                    />
                );
            })}
        </div>
    );
}
