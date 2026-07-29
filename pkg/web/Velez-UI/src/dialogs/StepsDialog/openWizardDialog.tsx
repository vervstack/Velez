import {useDialog} from '@/app/hooks/dialog/Dialog.tsx';
import StepsDialog, {StepsDialogProps} from '@/dialogs/StepsDialog/StepsDialog.tsx';

export function OpenWizardDialog<TContext extends object>(props: StepsDialogProps<TContext>) {
    useDialog.getState().OpenDialog(<StepsDialog<TContext> {...props} />);
}
