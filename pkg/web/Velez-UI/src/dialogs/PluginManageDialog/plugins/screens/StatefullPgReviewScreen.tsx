import {StatefullPgContext} from '@/dialogs/PluginManageDialog/plugins/StatefullPgContext.ts';
import cls from '@/dialogs/PluginManageDialog/plugins/screens/StatefullPgReviewScreen.module.css';

type StatefullPgReviewScreenProps = Pick<StatefullPgContext, 'exposePort' | 'portNumber'>;

export default function StatefullPgReviewScreen({exposePort, portNumber}: StatefullPgReviewScreenProps) {
    return (
        <div className={cls.ReviewContainer}>
            <div className={cls.ReviewRow}>
                <span className={cls.ReviewLabel}>Expose port:</span>
                <span className={cls.ReviewValue}>{exposePort ? 'Yes' : 'No'}</span>
            </div>

            {exposePort && (
                <div className={cls.ReviewRow}>
                    <span className={cls.ReviewLabel}>Port number:</span>
                    <span className={cls.ReviewValue}>{portNumber}</span>
                </div>
            )}
        </div>
    );
}
