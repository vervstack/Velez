import cls from '@/widgets/sidebar/Sidebar.module.css';
import cn from 'classnames';
import SectionLabel from '@/components/base/SectionLabel';
import VelezIcon from '@/assets/icons/services/velez.svg';

type NavId = 'controlplane' | 'vcn' | 'deployments' | 'services' | 'search';
type ToolId = 'secrets' | 'config' | 'logs' | 'settings';

interface SidebarProps {
    activeNav: NavId;
    onNavChange: (id: NavId) => void;
    onToolNav?: (id: ToolId) => void;
}

const NAV_ITEMS: Array<{ id: NavId; label: string; icon: string }> = [
    {id: 'controlplane', label: 'Control Plane', icon: '⬡'},
    {id: 'vcn', label: 'VCN', icon: '◎'},
    {id: 'deployments', label: 'Deployments', icon: '⬕'},
    {id: 'services', label: 'Services', icon: '⬡'},
    {id: 'search', label: 'Search', icon: '⌕'},
];

const TOOL_ITEMS = [
    {id: 'secrets', label: 'Secrets', icon: '⊡'},
    {id: 'config', label: 'Config', icon: '≡'},
    {id: 'logs', label: 'Logs', icon: '≈'},
    {id: 'settings', label: 'Settings', icon: '◈'},
];

export default function Sidebar({activeNav, onNavChange, onToolNav}: SidebarProps) {
    return (
        <aside className={cls.SidebarContainer}>
            <Logo/>

            {/* Main nav */}
            <nav className={cls.nav}>
                <div className={cn(cls.sectionHeader, cls.navSectionHeader)}>
                    <SectionLabel>Services</SectionLabel>
                </div>
                {
                    NAV_ITEMS.map((n) =>
                        <div key={n.id}>
                            <NavItem
                                id={n.id}
                                label={n.label}
                                icon={n.icon}
                                isActive={activeNav === n.id}
                                onNavChange={onNavChange}
                            />
                        </div>)}

                <div className={cls.divider}/>
                <div className={cn(cls.sectionHeader, cls.navSectionHeader)}>
                    <SectionLabel>Tools</SectionLabel>
                </div>

                {TOOL_ITEMS.map(function renderToolItem(item) {
                    function handleToolClick() {
                        if (onToolNav) {
                            onToolNav(item.id as ToolId);
                        }
                    }

                    return (
                        <div
                            key={item.id}
                            className={cls.toolItem}
                            onClick={handleToolClick}
                        >
                            <span className={cls.toolIcon}>{item.icon}</span>
                            <span className={cls.toolLabel}>{item.label}</span>
                        </div>
                    );
                })}
            </nav>

            <UserBar/>
        </aside>
    );
}


function Logo() {
    return (
        <div className={cls.LogoContainer}>
            <img src={VelezIcon} alt="Velez" className={cls.logoIcon}/>
            <span className={cls.logoText}>
                Velez
                <span className={cls.logoSub}> / VervStack</span>
            </span>
        </div>
    )
}

interface NavItemProps {
    id: NavId;
    label: string;
    icon: string;
    isActive: boolean;
    onNavChange: (id: NavId) => void;
}

function NavItem({isActive, onNavChange, id, label, icon}: NavItemProps) {
    function handleClick() {
        onNavChange(id);
    }

    return (
        <div
            key={id}
            className={cn(cls.navItem, {[cls.navItemActive]: isActive})}
            onClick={handleClick}
        >
            <span className={cls.navIcon}>{icon}</span>
            <span className={cn(cls.navLabel, {[cls.navLabelActive]: isActive})}>
                {label}
            </span>
        </div>
    );
}

function UserBar() {
    return (
        <div className={cls.UserBarContainer}>
            <div className={cls.Avatar}>RS</div>

            <div>
                <div className={cls.UserName}>RedSock</div>
                <div className={cls.UserRole}>admin</div>
            </div>
        </div>)
}
