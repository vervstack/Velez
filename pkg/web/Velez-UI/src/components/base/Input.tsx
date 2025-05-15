import {ChangeEventHandler} from "react";

export default function Input(
    {
        onChange,
    }: {
        onChange: ChangeEventHandler<HTMLInputElement> | undefined
    }) {
    return (
        <div>
            <input
                onChange={onChange}
            ></input>
        </div>
    )
}
