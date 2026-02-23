import cls from "@/components/Loader.module.css";

interface LoaderWithErrorProps {
    err: string | undefined
}

export default function LoaderWithError({err}: LoaderWithErrorProps) {
    return (
        <div className={cls.LoaderContainer}>
            {
                err ?
                    <div>{err}</div>
                    :
                    <div className={cls.Spinner}/>
            }

        </div>
    )
}
