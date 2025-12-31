import {CreateServiceRequest} from "@vervstack/velez";
import Button from "@/components/base/Button.tsx";
import Input from "@/components/base/Input.tsx";
import {useState} from "react";

interface InitServiceWidgetProps {
    createCallback: (req: CreateServiceRequest) => Promise<void>;
}

export default function InitServiceWidget({createCallback}: InitServiceWidgetProps) {
    const [isActive, setIsActive] = useState<boolean>(true);

    // TODO - when user inputs service name - check that this name is not already exists
    const [serviceName, setServiceName] = useState<string>('');


    function callCreateCallback() {
        setIsActive(false);
        const req: CreateServiceRequest = {
            name: serviceName
        }

        createCallback(req)
            .then(() => {setIsActive(true
            )})
    }

    return (
        <div>
            <Input
                label={'Service Name'}
                inputValue={serviceName} onChange={setServiceName}/>

            <Button
                isDisabled={!isActive}
                onClick={callCreateCallback}
                title={"Create"}/>
        </div>
    )
}
