import type {ComponentType} from 'react';

export type StepScreenProps<TContext> = TContext & {
    updateContext(partial: Partial<TContext>): void;
};

export interface Step<TContext> {
    name: string;
    label?: string;
    nextLabel?: string;
    component: ComponentType<StepScreenProps<TContext>>;
    canProceed?(context: TContext): boolean;
}

export interface ConfirmFinishProps<TContext> {
    context: TContext;
    onConfirm(): void;
    onCancel(): void;
}

export interface FinalStepProps<TContext> {
    context: TContext;
    onClose(): void;
}
