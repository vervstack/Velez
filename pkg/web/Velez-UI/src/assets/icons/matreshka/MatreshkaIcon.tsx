import MatreshkaIconPng from './MatreshkaIcon.png';
import styles from '@/assets/icons/icon.module.css';

export default function MatreshkaIcon() {
    return (
        <img
            src={MatreshkaIconPng}
            className={styles.Icon}
            alt={'Matreshka'}
        />
    )
}
