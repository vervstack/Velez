import styles from './card.module.css'
import {Tooltip} from "react-tooltip";
import {Service} from "@/model/services/Services";

export default function ServiceCard({title, icon, webLink, description}: Service) {
    return (
        <div className={styles.CardContainer}>
            <div className={styles.CardTop}>
                <div className={styles.ServiceIcon}>{icon}</div>

                <div className={styles.Tittle}>{title}</div>

                {
                    webLink ? <div
                        className={styles.ExternalLink}
                        data-tooltip-id={"open-external-service-link-" + title}
                        data-tooltip-content="Open in new window"
                        data-tooltip-place="left"
                        onClick={() => window.open(webLink, '_blank')}
                    >
                        <span
                            className="material-symbols-outlined"
                            children={"open_in_new"}/>
                    </div> : null
                }
            </div>
            <div className={styles.CardBottom}>
                <div className={styles.Content}>
                    {description}
                </div>
            </div>

            <Tooltip
                id={"open-external-service-link-" + title}
            />
        </div>
    )
}
