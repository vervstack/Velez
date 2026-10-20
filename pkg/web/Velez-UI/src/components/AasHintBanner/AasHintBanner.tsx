import cls from "@/components/AasHintBanner/AasHintBanner.module.css"
import Button from "@/components/base/Button.tsx"

interface Props {
    title: string
    onOpen(): void
    onDismiss(): void
}

export default function AasHintBanner({title, onOpen, onDismiss}: Props) {
    function handleOpen() {
        onOpen()
    }

    function handleDismiss() {
        onDismiss()
    }

    return (
        <div className={cls.AasHintBannerContainer} role="note">
            <span className={cls.Text}>There is a managed option for this image: {title}</span>
            <Button sm variant="primary" onClick={handleOpen}>Open</Button>
            <Button sm variant="ghost" onClick={handleDismiss} ariaLabel="Dismiss">✕</Button>
        </div>
    )
}
