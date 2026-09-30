import {Dropdown, DropdownOption} from "@vervstack/chures"

import cls from "@/dialogs/CreateServiceDialog/components/RunnerFields/RunnerFields.module.css"
import {RunnerProvider, RunnerScope} from "@/app/api/velez"
import Input from "@/components/base/Input.tsx"
import type {RunnerForm} from "@/dialogs/CreateServiceDialog/processes/buildRegisterContainerRequest.ts"
import {parseRunnerProvider, parseRunnerScope} from "@/dialogs/CreateServiceDialog/processes/runnerDefaults.ts"

const PROVIDER_OPTIONS: DropdownOption[] = [
    {id: RunnerProvider.GITHUB, name: "GitHub"},
    {id: RunnerProvider.GITLAB, name: "GitLab"},
]

const SCOPE_OPTIONS: DropdownOption[] = [
    {id: RunnerScope.REPO, name: "Repository"},
    {id: RunnerScope.ORG, name: "Organization"},
]

interface Props {
    form: RunnerForm
    isRegistrationTokenFound: boolean
    onChange(form: RunnerForm): void
}

export default function RunnerFields({form, isRegistrationTokenFound, onChange}: Props) {
    const isGitlab = form.provider === RunnerProvider.GITLAB

    function handleProviderChange(ids: string[]) {
        onChange({...form, provider: parseRunnerProvider(ids[0])})
    }

    function handleScopeChange(ids: string[]) {
        onChange({...form, scope: parseRunnerScope(ids[0]) ?? RunnerScope.REPO})
    }

    function handleTargetChange(target: string) {
        onChange({...form, target})
    }

    function handleBaseUrlChange(baseUrl: string) {
        onChange({...form, baseUrl})
    }

    function handleLabelsChange(labels: string) {
        onChange({...form, labels})
    }

    function handleDockerImageChange(dockerImage: string) {
        onChange({...form, dockerImage})
    }

    function handleConcurrentChange(concurrent: string) {
        onChange({...form, concurrent})
    }

    function handleAccessTokenChange(accessToken: string) {
        onChange({...form, accessToken})
    }

    function handleRegistrationTokenChange(registrationToken: string) {
        onChange({...form, registrationToken})
    }

    return (
        <div className={cls.RunnerFieldsContainer}>
            <Dropdown
                label="Provider"
                placeholder="Select provider"
                options={PROVIDER_OPTIONS}
                value={form.provider ? [form.provider] : []}
                onChange={handleProviderChange}
                portal
            />
            <Dropdown
                label="Scope"
                placeholder="Select scope"
                options={SCOPE_OPTIONS}
                value={[form.scope]}
                onChange={handleScopeChange}
                portal
            />
            <Input label="Target" inputValue={form.target} onChange={handleTargetChange}/>
            {isGitlab && <Input label="GitLab Base URL" inputValue={form.baseUrl} onChange={handleBaseUrlChange}/>}
            <Input label="Labels (comma-separated)" inputValue={form.labels} onChange={handleLabelsChange}/>
            {isGitlab && (
                <Input label="Docker Image (optional)" inputValue={form.dockerImage} onChange={handleDockerImageChange}/>
            )}
            {isGitlab && (
                <Input label="Concurrent jobs (optional)" inputValue={form.concurrent} onChange={handleConcurrentChange}/>
            )}
            <Input
                label="Access token (optional, only needed to deregister or mint tokens)"
                inputValue={form.accessToken}
                onChange={handleAccessTokenChange}
            />
            <Input
                label="Registration token (optional)"
                inputValue={form.registrationToken}
                onChange={handleRegistrationTokenChange}
            />
            {isRegistrationTokenFound && (
                <span className={cls.Hint}>
                    A registration token was found in the container; leave the field empty to use it.
                </span>
            )}
        </div>
    )
}
