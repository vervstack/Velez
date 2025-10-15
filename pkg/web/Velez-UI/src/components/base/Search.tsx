import cls from "@/components/base/Search.module.css";

import Input from "@/components/base/Input.tsx";

import SearchSvg from "@/assets/icons/Search.svg";

export interface SearchParam {
    name: string;
    values: string | number | boolean;
}

interface SearchProps {
    label?: string;
    onChange: (v: string) => void;
    value?: string
    searchParams?: SearchParam[]
}

export default function Search({label, value, onChange}: SearchProps) {
    return (
        <div className={cls.SearchContainer}>
            <Input
                label={label}
                onChange={onChange}
                inputValue={value}
                style={{
                    borderless: true
                }}
            />
            <div
                className={cls.SearchImage}
            >
                <img
                    src={SearchSvg} alt={'?'}/>
            </div>
        </div>
    )
}
