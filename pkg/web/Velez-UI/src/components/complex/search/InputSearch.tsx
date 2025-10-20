import cls from "@/components/complex/search/InputSearch.module.css";
import {useEffect, useRef, useState} from "react";

import Input, {InputProps} from "@/components/base/Input.tsx";

import {SuggestElem} from "@/model/common/suggest.ts";
import cn from "classnames";

interface InputSearchProps extends InputProps {
    suggests?: SuggestElem[];
    onSuggestDismiss?: () => void;
}

export default function InputSearch(props: InputSearchProps) {
    const [isSuggestOpen, setIsSuggestOpen] = useState(false);

    const dropdownRef = useRef(null);

    useEffect(() => {
        function handleClickOutside(event: MouseEvent) {

            if (dropdownRef.current &&
                // @ts-ignore
                !dropdownRef.current.contains(event.target)) {
                setIsSuggestOpen(false);
            }

            if (props.onSuggestDismiss) props.onSuggestDismiss()
        }

        // Bind the listener
        document.addEventListener("mousedown", handleClickOutside);

        return () => {
            // Clean up the listener on unmount
            document.removeEventListener("mousedown", handleClickOutside);
        };
    }, [dropdownRef]);

    return (
        <div className={cls.InputSearchContainer}>
            <div onClick={() => setIsSuggestOpen(true)}>
                <Input {...props}/>
            </div>
            <div
                ref={dropdownRef}
                className={cn(cls.SuggestionList, {
                    [cls.Open]: isSuggestOpen,
                })}>
                {props.suggests ?
                    props.suggests.map((s) =>
                        <div
                            key={s.name}
                            className={cls.Item}>{s.name}</div>
                    )
                    :
                    'Nothing found'}
            </div>
        </div>
    )
}
