import { useState } from 'react'

import { DeploymentInfo } from '@/app/api/velez'
import DeployRow from '@/widgets/service/DeploymentHistory/DeployRow'
import MockDeployRow from '@/widgets/service/DeploymentHistory/MockDeployRow'
import { MOCK_DEPLOYS } from '@/widgets/service/DeploymentHistory/mockDeploys'
import cls from '@/widgets/service/DeploymentHistory/DeploymentHistory.module.css'

const DEFAULT_VISIBLE = 6

interface DeploymentHistoryProps {
    deployments: DeploymentInfo[]
    currentDeploymentId?: string
}

export default function DeploymentHistory({ deployments, currentDeploymentId }: DeploymentHistoryProps) {
    const [showAll, setShowAll] = useState(false)

    const useMock = deployments.length === 0

    function handleToggle() {
        setShowAll(function toggle(prev) { return !prev })
    }

    if (useMock) {
        const visibleMock = showAll ? MOCK_DEPLOYS : MOCK_DEPLOYS.slice(0, DEFAULT_VISIBLE)
        const hasMoreMock = MOCK_DEPLOYS.length > DEFAULT_VISIBLE

        return (
            <div className={cls.DeploymentHistoryContainer}>
                <div className={cls.TableWrapper}>
                    <div className={cls.HeaderRow}>
                        <div className={cls.HeaderCell}>id</div>
                        <div className={cls.HeaderCell}>version</div>
                        <div className={cls.HeaderCell}>deployed at</div>
                        <div className={cls.HeaderCell}>by</div>
                        <div className={cls.HeaderCell}>commit · branch</div>
                        <div className={cls.HeaderCell}>env</div>
                        <div className={cls.HeaderCell}>actions</div>
                    </div>
                    <div className={cls.RowsWrapper}>
                        {visibleMock.map(function renderMockRow(d, idx) {
                            const prevDeployId = idx < visibleMock.length - 1 ? visibleMock[idx + 1].id : undefined
                            return (
                                <MockDeployRow
                                    key={d.id}
                                    deploy={d}
                                    isLast={idx === visibleMock.length - 1}
                                    prevDeployId={prevDeployId}
                                />
                            )
                        })}
                    </div>
                </div>
                {hasMoreMock && (
                    <button className={cls.ToggleBtn} onClick={handleToggle}>
                        {showAll ? 'Show less' : `Show all ${MOCK_DEPLOYS.length} deployments`}
                    </button>
                )}
            </div>
        )
    }

    const visible = showAll ? deployments : deployments.slice(0, DEFAULT_VISIBLE)
    const hasMore = deployments.length > DEFAULT_VISIBLE

    return (
        <div className={cls.DeploymentHistoryContainer}>
            <div className={cls.TableWrapper}>
                <div className={cls.HeaderRow}>
                    <div className={cls.HeaderCell}>id</div>
                    <div className={cls.HeaderCell}>version</div>
                    <div className={cls.HeaderCell}>deployed at</div>
                    <div className={cls.HeaderCell}>by</div>
                    <div className={cls.HeaderCell}>commit · branch</div>
                    <div className={cls.HeaderCell}>env</div>
                    <div className={cls.HeaderCell}>actions</div>
                </div>

                <div className={cls.RowsWrapper}>
                    {visible.map(function renderRow(dep: DeploymentInfo, idx: number) {
                        const isCurrent = !!currentDeploymentId && dep.id === currentDeploymentId
                        const isLast = idx === visible.length - 1
                        const prevDeployId = idx > 0 ? visible[idx - 1].id : undefined

                        return (
                            <DeployRow
                                key={dep.id ?? idx}
                                deployment={dep}
                                isCurrent={isCurrent}
                                isLast={isLast}
                                prevDeployId={prevDeployId}
                            />
                        )
                    })}
                </div>
            </div>

            {hasMore && (
                <button className={cls.ToggleBtn} onClick={handleToggle}>
                    {showAll ? 'Show less' : `Show all ${deployments.length} deployments`}
                </button>
            )}
        </div>
    )
}
