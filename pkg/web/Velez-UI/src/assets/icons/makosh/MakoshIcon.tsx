import MakoshIconPng from "@/assets/icons/makosh/Makosh.png";
import styles from "@/assets/icons/icon.module.css";

export default function MakoshIcon() {
    return (
        <img
            src={MakoshIconPng}
            className={styles.Icon}
            alt={'Matreshka'}
        />
    )
}
