import React from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogContentText,
  DialogActions,
  Button,
  CircularProgress
} from '@mui/material';

/**
 * Props for the ConfirmDialog component
 */
export interface ConfirmDialogProps {
  /** Whether the dialog is open */
  open: boolean;
  /** Function to call when the dialog is closed */
  onClose: () => void;
  /** Function to call when the action is confirmed */
  onConfirm: () => void;
  /** Title of the dialog */
  title: string;
  /** Message to display in the dialog */
  message: string;
  /** Text for the confirm button (default: "Confirm") */
  confirmText?: string;
  /** Text for the cancel button (default: "Cancel") */
  cancelText?: string;
  /** Whether the action is currently loading */
  loading?: boolean;
  /** Severity of the action (affects confirm button color) */
  severity?: 'error' | 'warning' | 'info' | 'success';
}

/**
 * A reusable confirmation dialog component for delete operations and other confirmations.
 */
export const ConfirmDialog: React.FC<ConfirmDialogProps> = ({
  open,
  onClose,
  onConfirm,
  title,
  message,
  confirmText = 'Confirm',
  cancelText = 'Cancel',
  loading = false,
  severity = 'error'
}) => {
  // Map severity to button color
  const buttonColor = {
    error: 'error',
    warning: 'warning',
    info: 'primary',
    success: 'success'
  } as const;
  
  return (
    <Dialog
      open={open}
      onClose={() => !loading && onClose()}
      aria-labelledby="confirm-dialog-title"
      aria-describedby="confirm-dialog-description"
    >
      <DialogTitle id="confirm-dialog-title">{title}</DialogTitle>
      <DialogContent>
        <DialogContentText id="confirm-dialog-description">
          {message}
        </DialogContentText>
      </DialogContent>
      <DialogActions>
        <Button 
          onClick={onClose} 
          disabled={loading}
        >
          {cancelText}
        </Button>
        <Button 
          onClick={onConfirm} 
          color={buttonColor[severity]}
          disabled={loading}
          variant={severity === 'error' ? 'contained' : 'text'}
          startIcon={loading ? <CircularProgress size={16} /> : undefined}
        >
          {loading ? `${confirmText}...` : confirmText}
        </Button>
      </DialogActions>
    </Dialog>
  );
};

export default ConfirmDialog; 