import VelezSvgIcon from '@/assets/icons/velez/VelezIcon.svg';
import styles from '@/assets/icons/icon.module.css';

export default function VelezIcon() {
    return (
        <img
            className={styles.Icon}
            src={VelezSvgIcon}
            alt={'Icon'}/>
    )
}
