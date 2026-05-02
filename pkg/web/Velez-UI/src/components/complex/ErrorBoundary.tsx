import {Component, ReactNode} from "react";

interface Props {
    children: ReactNode;
}

interface State {
    error: Error | null;
}

export default class ErrorBoundary extends Component<Props, State> {
    constructor(props: Props) {
        super(props);
        this.state = {error: null};
    }

    static getDerivedStateFromError(error: Error): State {
        return {error};
    }

    render() {
        if (this.state.error) {
            return (
                <div style={{padding: "2rem", display: "flex", flexDirection: "column", gap: "1rem"}}>
                    <div style={{fontSize: "1.1rem", fontWeight: 600}}>Something went wrong</div>
                    <pre style={{fontFamily: "monospace", fontSize: "0.8rem", opacity: 0.7, whiteSpace: "pre-wrap"}}>
                        {this.state.error.message}
                    </pre>
                    <button
                        style={{width: "fit-content", padding: "0.4rem 1rem", cursor: "pointer"}}
                        onClick={() => this.setState({error: null})}
                    >
                        Retry
                    </button>
                </div>
            );
        }
        return this.props.children;
    }
}
