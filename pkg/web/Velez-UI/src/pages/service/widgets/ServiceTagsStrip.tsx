import cls from "@/pages/service/widgets/ServiceTagsStrip.module.css";
import {ListSmerdsByServiceIdQuery} from "@/processes/queries/smerds.ts";
import {getSmerdTags} from "@/processes/mappings/smerds.ts";
import TagChip from "@/components/base/chips/TagChip.tsx";

interface Props {
    serviceName: string;
}

export default function ServiceTagsStrip({serviceName}: Props) {
    const smerdsQuery = ListSmerdsByServiceIdQuery(serviceName);
    const currentSmerd = smerdsQuery.data?.smerds?.[0];
    const tags = getSmerdTags(currentSmerd);

    if (tags.length === 0) return null;

    return (
        <div className={cls.ServiceTagsStripContainer}>
            {tags.map(function renderTag(tag) {
                return <TagChip key={tag.key} tagKey={tag.key} value={tag.value}/>;
            })}
        </div>
    );
}
