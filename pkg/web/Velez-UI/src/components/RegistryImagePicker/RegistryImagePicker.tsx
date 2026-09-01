import {useEffect, useState} from "react";
import {Dropdown, DropdownOption, parseGrpcError} from "@vervstack/chures";

import cls from "@/components/RegistryImagePicker/RegistryImagePicker.module.css";

import {ListRegistriesQuery} from "@/processes/queries/control_plane.ts";
import {ListImages} from "@/processes/api/velez.ts";
import useSettings from "@/app/settings/state.ts";
import {useToaster} from "@/app/hooks/toaster/Toaster.ts";

interface Props {
    label?: string;
    value?: string;
    onChange: (v: string) => void;
}

export default function RegistryImagePicker({label, value, onChange}: Props) {
    const registriesQuery = ListRegistriesQuery();
    const settings = useSettings();
    const toaster = useToaster();

    const registries = registriesQuery.data?.registries || [];
    const [registryId, setRegistryId] = useState<string | undefined>(undefined);

    useEffect(function selectDefaultRegistry() {
        if (registryId !== undefined || registries.length === 0) {
            return;
        }
        const defaultRegistry = registries.find((r) => r.isDefault) || registries[0];
        setRegistryId(defaultRegistry.id);
    }, [registries, registryId]);

    function handleError(err: unknown) {
        toaster.catchGrpc(parseGrpcError(err));
    }

    function handleRegistryChange(ids: string[]) {
        setRegistryId(ids[0]);
    }

    function handleImageChange(ids: string[]) {
        onChange(ids[0] || "");
    }

    function handleSearch(query: string): Promise<DropdownOption[]> {
        return ListImages(query, settings.initReq(), registryId).then((res) =>
            (res.images || []).map((img) => ({
                id: img.name || "",
                name: img.latestTag ? `${img.name}:${img.latestTag}` : (img.name || ""),
            }))
        );
    }

    const registryOptions: DropdownOption[] = registries.map((r) => ({id: r.id || "", name: r.name || ""}));

    return (
        <div className={cls.RegistryImagePickerContainer}>
            <Dropdown
                label="Registry"
                options={registryOptions}
                value={registryId ? [registryId] : []}
                onChange={handleRegistryChange}
                isLoading={registriesQuery.isLoading}
                onError={handleError}
                portal
            />
            <Dropdown
                label={label || "Image"}
                value={value ? [value] : []}
                onChange={handleImageChange}
                onSearch={handleSearch}
                onError={handleError}
                portal
            />
        </div>
    );
}
