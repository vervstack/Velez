import React from 'react';
import cls from '@/components/complex/PluginManageDialog/PluginManageDialog.module.css';
import IconButton from '@/components/base/IconButton';

interface PluginManageDialogProps {
    isOpen: boolean;
    pluginName: string;
    onClose: () => void;
}

export default function PluginManageDialog({ isOpen, pluginName, onClose }: PluginManageDialogProps) {
    if (!isOpen) return null;

    function handleOverlayClick() {
        onClose();
    }

    function handleContainerClick(e: React.MouseEvent) {
        e.stopPropagation();
    }

    return (
        <div className={cls.ModalOverlay} onClick={handleOverlayClick}>
            <div className={cls.ModalContainer} onClick={handleContainerClick}>
                <div className={cls.ModalHeader}>
                    <h2 className={cls.ModalTitle}>Manage {pluginName}</h2>
                    <IconButton label="×" onClick={onClose} title="Close" />
                </div>
                <div className={cls.ModalContent}>
                </div>
            </div>
        </div>
    );
}
