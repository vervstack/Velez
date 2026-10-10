import cls from "@/dialogs/CreateServiceDialog/screens/PickerScreen/PickerScreen.module.css"
import Badge from "@/components/base/Badge.tsx"
import Choice from "@/components/base/Choice.tsx"
import {
    PRODUCT_CARDS,
    ProductCard,
    ProductScreen,
    ServiceScreen,
} from "@/dialogs/CreateServiceDialog/processes/serviceScreen.ts"

const DISABLED_NOTE = "Not available for existing containers yet"

interface Props {
    suggestedScreen?: ServiceScreen
    enabledScreens?: ProductScreen[]
    onSelect(screen: ProductScreen): void
}

export default function PickerScreen({suggestedScreen, enabledScreens, onSelect}: Props) {
    function renderCard(card: ProductCard) {
        const isSuggested = card.screen === suggestedScreen
        const isDisabled = enabledScreens !== undefined && !enabledScreens.includes(card.screen)

        function handleClick() {
            onSelect(card.screen)
        }

        const CardIcon = card.icon

        return (
            <div key={card.screen} className={cls.CardWrapper}>
                {isSuggested && <Badge label="Suggested" dim="var(--cyan-dim)"/>}
                <Choice
                    title={card.title}
                    sub={card.description}
                    icon={<CardIcon/>}
                    active={false}
                    isSuggested={isSuggested}
                    disabled={isDisabled}
                    onClick={handleClick}
                />
                {isDisabled && <span className={cls.DisabledNote}>{DISABLED_NOTE}</span>}
            </div>
        )
    }

    return (
        <div className={cls.PickerScreenContainer}>
            {PRODUCT_CARDS.map(renderCard)}
        </div>
    )
}
