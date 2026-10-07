import cls from "@/components/NetworkClusterBanner/NetworkClusterBanner.module.css"
import Button from "@/components/base/Button.tsx"

const BANNER_TEXT =
    "Cluster mode is on and the closed network (VCN) is not connected. " +
    "Traffic between services on different nodes may not be routed. Set up VCN to connect them."

interface Props {
    onOpenVcn(): void
}

export default function NetworkClusterBanner({onOpenVcn}: Props) {
    return (
        <div className={cls.NetworkClusterBannerContainer} role="alert">
            <span className={cls.Text}>{BANNER_TEXT}</span>
            <Button sm variant="warn" onClick={onOpenVcn}>Set up VCN</Button>
        </div>
    )
}
