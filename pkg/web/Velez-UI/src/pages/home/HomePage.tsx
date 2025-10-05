import styles from './home.module.css';

import ControlServicesWidget from "@/widgets/services/ControlPlane";

export default function HomePage() {
    return (
        <div className={styles.Home}>
            <div className={styles.ControlServices}>
                <ControlServicesWidget/>
            </div>
        </div>
    )
}
