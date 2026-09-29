import cls from "@/dialogs/CreateServiceDialog/screens/PickerScreen/PickerScreen.module.css"
import Badge from "@/components/base/Badge.tsx"
import Choice from "@/components/base/Choice.tsx"
import {
    PRODUCT_CARDS,
    ProductCard,
    ProductScreen,
    ServiceScreen,
} from "@/dialogs/CreateServiceDialog/processes/serviceScreen.ts"

interface Props {
    suggestedScreen?: ServiceScreen
    onSelect(screen: ProductScreen): void
}

export default function PickerScreen({suggestedScreen, onSelect}: Props) {
    function renderCard(card: ProductCard) {
        const isSuggested = card.screen === suggestedScreen

        function handleClick() {
            onSelect(card.screen)
        }

        return (
            <div key={card.screen} className={cls.CardWrapper}>
                {isSuggested && <Badge label="Suggested" dim="var(--cyan-dim)"/>}
                <Choice title={card.title} sub={card.description} active={isSuggested} onClick={handleClick}/>
            </div>
        )
    }

    return (
        <div className={cls.PickerScreenContainer}>
            {PRODUCT_CARDS.map(renderCard)}
        </div>
    )
}
