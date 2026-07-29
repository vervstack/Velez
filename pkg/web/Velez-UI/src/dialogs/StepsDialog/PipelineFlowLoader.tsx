import cls from '@/dialogs/StepsDialog/PipelineFlowLoader.module.css';

export default function PipelineFlowLoader() {
    return (
        <div className={cls.PipelineFlowLoaderContainer}>
            <svg className={cls.Svg} viewBox="0 0 240 200" aria-hidden="true">
                <path
                    className={cls.ConnectorS1Db}
                    d="M120 140 L120 121"
                    pathLength={100}
                    data-testid="pipeline-connector-s1-db"
                />
                <path
                    className={cls.ConnectorDbS2}
                    d="M108 76 L40 52"
                    pathLength={100}
                    data-testid="pipeline-connector-db-s2"
                />
                <path
                    className={cls.ConnectorDbS3}
                    d="M132 76 L200 52"
                    pathLength={100}
                    data-testid="pipeline-connector-db-s3"
                />

                <g className={cls.ServerBottom} data-testid="pipeline-server-bottom">
                    <rect x="105" y="140" width="30" height="42" rx="4"/>
                    <line x1="112" y1="152" x2="122" y2="152"/>
                    <line x1="112" y1="164" x2="128" y2="164"/>
                </g>

                <g data-testid="pipeline-database">
                    <rect className={cls.DbBody} x="102" y="82" width="36" height="32"/>
                    <ellipse className={cls.DbBottom} cx="120" cy="114" rx="18" ry="7" pathLength={100}/>
                    <line className={cls.DbSides} x1="102" y1="82" x2="102" y2="114" pathLength={100}/>
                    <line className={cls.DbSides} x1="138" y1="82" x2="138" y2="114" pathLength={100}/>
                    <ellipse className={cls.DbMiddle} cx="120" cy="98" rx="18" ry="7" pathLength={100}/>
                    <ellipse className={cls.DbTop} cx="120" cy="82" rx="18" ry="7" pathLength={100}/>
                </g>

                <g className={cls.ServerTopLeft} data-testid="pipeline-server-top-left">
                    <rect x="25" y="10" width="30" height="42" rx="4"/>
                    <line x1="32" y1="22" x2="42" y2="22"/>
                    <line x1="32" y1="34" x2="48" y2="34"/>
                </g>

                <g className={cls.ServerTopRight} data-testid="pipeline-server-top-right">
                    <rect x="185" y="10" width="30" height="42" rx="4"/>
                    <line x1="192" y1="22" x2="202" y2="22"/>
                    <line x1="192" y1="34" x2="208" y2="34"/>
                </g>
            </svg>
        </div>
    );
}
