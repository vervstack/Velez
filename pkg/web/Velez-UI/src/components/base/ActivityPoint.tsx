import {Tooltip} from "react-tooltip";
import cn from "classnames";

import cls from "@/components/base/ActivityPoint.module.css";

interface ActivityPointProps {
    isInactive?: boolean;
}

export default function ActivityPoint({isInactive}: ActivityPointProps) {
    
    function getTooltipContent(): string {
        if (isInactive !== undefined) {
            return isInactive ? "Inactive" : "Active";
        }
        return "Unknown";
    }

    return (
        <div>
            <div
                data-tooltip-id={"activity-marker"}
                data-tooltip-content={getTooltipContent()}
                data-tooltip-place="left"

                className={cn(cls.activityPoint, {
                    [cls.inactive]: isInactive !== undefined && isInactive,
                    [cls.active]: isInactive !== undefined && !isInactive,
                    [cls.unknown]: isInactive === undefined,
                })
                }
            >
            </div>
            <Tooltip
                id={"activity-marker"}
            />
        </div>

    );
}
