import styles from './home.module.css';

import {useState} from "react";
import ControlPageWidget from "@/widgets/services/ControlPlane";

export default function HomePage() {
    const [openedWidget, setOpenedWidget]
        = useState<React.JSX.Element>(<ControlPageWidget/>)

    return (
        <div className={styles.Home}>
            <div className={styles.OpenedWidget}>
                {openedWidget}
            </div>
        </div>
    )
}
