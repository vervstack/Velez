import styles from './card.module.css'
import {Tooltip} from "react-tooltip";
import {Service} from "@/model/services/Services";

export default function ServiceCard(props: Service) {
    return (
        <div className={styles.Card}>
            <div className={styles.CardTop}>
                <div className={styles.ServiceIcon}>
                    {props.icon}
                </div>
                <div className={styles.Tittle}>
                    {props.tittle}
                </div>

                <div
                    className={styles.ExternalLink}
                    data-tooltip-id={"open-external-service-link-" + props.tittle}
                    data-tooltip-content="Open in new window"
                    data-tooltip-place="left"
                >
                        <span
                            className="material-symbols-outlined"
                            children={"open_in_new"}/>
                </div>
            </div>
            <div className={styles.CardBottom}>
                <div className={styles.Content}>
                    {'Surely hills examines comparison mirror beings pork, surname race vegas south carry fabrics athletic, basename workshop payment parent identifier feed arguments, milton. '}
                </div>
            </div>

            <Tooltip
                id={"open-external-service-link-" + props.tittle}
            />
        </div>
    )
}
