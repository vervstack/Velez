import cls from "@/components/Loader.module.css";
import {useEffect, useState} from "react";
import {useToaster} from "@/app/hooks/toaster/Toaster.ts";

interface LoaderWithErrorProps {
    children: React.JSX.Element[];

    load?: Promise<void>;
}

export default function LoaderWrapper({load, children}: LoaderWithErrorProps) {
    const [isLoading, setIsLoading] = useState(false);
    const [err, setErr] = useState<Error | undefined>();
    const toaster = useToaster()

    useEffect(() => {
        if (err) {
            toaster.catchGrpc(err)
        }
    }, [err]);

    useEffect(() => {
        setErr(undefined);

        if (!load) {
            setIsLoading(false);
            return
        }

        setIsLoading(true);
        load
            .catch(setErr)
            .finally(() => setIsLoading(false));
    }, [load]);

    return (
        <div className={cls.LoaderContainer}>
            {isLoading && <div className={cls.Spinner}/>}
            {err && <div className={cls.Text}>{err.message}</div>}
            {children && children}
        </div>
    )
}
