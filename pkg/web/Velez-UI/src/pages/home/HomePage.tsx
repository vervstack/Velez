import styles from './home.module.css';

import ControlServicesWidget from "@/widgets/services/ControlPlane";
import SmerdsWidget from "@/widgets/smerds/SmerdsWidget.tsx";

export default function HomePage() {
    return (
        <div className={styles.Home}>
            <div className={styles.Smerds}><SmerdsWidget/></div>
            <div className={styles.ControlServices}><ControlServicesWidget/></div>
        </div>
    )
}
