import cls from "@/pages/transfer/TransferPage.module.css"
import IncomingTransferCard from "@/pages/transfer/components/IncomingTransferCard/IncomingTransferCard.tsx"
import MoveContainerForm from "@/pages/transfer/components/MoveContainerForm/MoveContainerForm.tsx"
import TransfersList from "@/pages/transfer/components/TransfersList/TransfersList.tsx"

export default function TransferPage() {
    return (
        <div className={cls.TransferPageContainer}>
            <div className={cls.Header}>
                <h1 className={cls.PageTitle}>Transfer</h1>
                <p className={cls.Subtitle}>
                    The source container is stopped during the move and stays stopped afterwards.
                    The copy on the target is created stopped.
                </p>
            </div>
            <MoveContainerForm/>
            <IncomingTransferCard/>
            <TransfersList/>
        </div>
    )
}
