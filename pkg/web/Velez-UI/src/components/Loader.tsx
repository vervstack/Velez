import cls from "@/components/Loader.module.css";

export default function Loader() {
    return (
        <div className={cls.LoaderContainer}>
            <div className={cls.Spinner}></div>
        </div>
    )
}
