import React, { useState, useEffect, useCallback } from 'react';
import {
  Box,
  Typography,
  Paper,
  Card,
  CardContent,
  Grid,
  Chip,
  Button,
  Divider,
  Stack,
  IconButton,
  Tooltip,
} from '@mui/material';
import RefreshIcon from '@mui/icons-material/Refresh';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import NotificationsActiveIcon from '@mui/icons-material/NotificationsActive';
import NotificationsOffIcon from '@mui/icons-material/NotificationsOff';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import AddIcon from '@mui/icons-material/Add';
import { apiClient } from '../api';
import type { AlertConfig, AlertStatus } from '../types/api';
import LoadingErrorHandler from '../components/LoadingErrorHandler';
import AlertDialog from '../components/AlertDialog';
import { PageHeader, ConfirmDialog, StatusChip, type StatusConfig } from '../components/common';
import { useNotification, useDateFormatter, useResourceCRUD } from '../hooks';

// Define status and severity maps for StatusChip
const ALERT_STATUS_MAP: Record<string, StatusConfig> = {
  'active': { label: 'Triggered', color: 'error' },
  'pending': { label: 'Pending', color: 'warning' },
  'resolved': { label: 'Normal', color: 'success' },
  'inactive': { label: 'Inactive', color: 'default' },
};

const ALERT_SEVERITY_MAP: Record<string, StatusConfig> = {
  'critical': { label: 'Critical', color: 'error' },
  'warning': { label: 'Warning', color: 'warning' },
  'info': { label: 'Info', color: 'info' },
};

/**
 * Alerts page component
 * Displays system alerts and allows management
 */
const Alerts: React.FC = () => {
  // Use the useResourceCRUD hook for alert management
  const {
    items: alerts,
    loading,
    error,
    lastUpdated,
    refetch,
    actionLoading,
    selectedItem: selectedAlert,
    setSelectedItem: setSelectedAlert,
    isDialogOpen,
    openDialog,
    closeDialog,
    handleCreate,
    handleUpdate,
    handleDelete,
  } = useResourceCRUD<AlertConfig, Partial<AlertConfig>, Partial<AlertConfig>>({
    resourceName: 'alert',
    fetchFn: apiClient.getAlerts,
    createFn: apiClient.createAlert,
    updateFn: apiClient.updateAlert,
    deleteFn: apiClient.deleteAlert,
    cacheTTL: 30000,
  });

  const [alertStatuses, setAlertStatuses] = useState<Record<string, AlertStatus>>({});
  const [testActionLoading, setTestActionLoading] = useState<boolean>(false);

  // Use the notification hook for managing notifications
  const { showNotification } = useNotification();

  // Use the date formatter hook for consistent date formatting
  const { formatDate } = useDateFormatter();

  // Function to fetch alert statuses
  const fetchAlertStatuses = useCallback(async () => {
    try {
      const response = await apiClient.getAllAlertStatus();

      if (response.success && response.data) {
        setAlertStatuses(response.data);
      } else {
        console.error('Failed to fetch alert statuses:', response.error);
      }
    } catch (err) {
      console.error('Failed to fetch alert statuses:', err);
    }
  }, []);

  // Fetch alerts and statuses on component mount
  useEffect(() => {
    fetchAlertStatuses();

    // Set up interval to refresh alert statuses
    const intervalId = setInterval(fetchAlertStatuses, 30000); // Every 30 seconds

    return () => {
      clearInterval(intervalId);
    };
  }, [fetchAlertStatuses]);

  // Handle opening the create alert dialog
  const openCreateAlertDialog = () => {
    setSelectedAlert(null);
    openDialog('create');
  };

  // Handle opening the edit alert dialog
  const openEditAlertDialog = (alert: AlertConfig) => {
    setSelectedAlert(alert);
    openDialog('edit');
  };

  // Handle opening the delete alert dialog
  const openDeleteAlertDialog = (alert: AlertConfig) => {
    setSelectedAlert(alert);
    openDialog('delete');
  };

  // Handle opening the test alert dialog
  const openTestAlertDialog = (alert: AlertConfig) => {
    setSelectedAlert(alert);
    openDialog('test');
  };

  // Handle toggling alert enabled status
  const handleToggleEnabled = async (alert: AlertConfig) => {
    try {
      const updatedAlert = { ...alert, enabled: !alert.enabled };
      await handleUpdate(alert.id, updatedAlert);
      // After update, refetch alert statuses as well
      fetchAlertStatuses();
    } catch (err) {
      // Notification is handled by useResourceCRUD's handleUpdate
      console.error('Toggle enabled failed:', err);
    }
  };

  // Handle saving a new alert
  const handleSaveAlert = async (alert: Partial<AlertConfig>) => {
    await handleCreate(alert);
    // After create, refetch alert statuses
    fetchAlertStatuses();
  };

  // Handle updating an existing alert
  const handleUpdateAlert = async (alert: Partial<AlertConfig>) => {
    if (!selectedAlert) return;
    await handleUpdate(selectedAlert.id, alert);
  };

  // Handle deleting an alert
  const handleConfirmDelete = async () => {
    if (!selectedAlert) return;
    await handleDelete(selectedAlert.id);
    setSelectedAlert(null);
  };

  // Handle testing an alert
  const handleConfirmTest = async () => {
    if (!selectedAlert) return;

    openDialog('test'); // Ensure dialog is open for the loading state

    // This action is specific to Alerts and not part of generic CRUD
    setTestActionLoading(true);
    try {
      const response = await apiClient.testAlert(selectedAlert.id);

      if (response.success) {
        closeDialog('test');
        showNotification('Test alert sent successfully', 'success');
      } else {
        throw new Error(response.error || 'Failed to test alert');
      }
    } catch (err) {
      showNotification(
        err instanceof Error ? err.message : 'An unknown error occurred',
        'error'
      );
    } finally {
      setTestActionLoading(false);
      setSelectedAlert(null);
    }
  };

  // Render threshold information
  const renderThreshold = (alert: AlertConfig) => {
    const { threshold } = alert;

    // Get the label for the metric name
    const metricTypeOptions: Record<string, string> = {
      'cpu': 'CPU',
      'memory': 'Memory',
      'load': 'System Load',
      'network': 'Network',
      'disk': 'Disk',
      'process': 'Process'
    };

    const metricType = threshold.metric_type;
    const metricName = threshold.metric_name;

    // Find the display label for the metric name
    let metricNameLabel = metricName;
    const metricNameOptions: Record<string, string> = {
      'usage_percent': 'Usage (%)',
      'load1': 'Load (1m)',
      'load5': 'Load (5m)',
      'load15': 'Load (15m)',
      'used_percent': 'Used (%)',
      'used': 'Used (bytes)',
      'free': 'Free (bytes)',
      'bytes_sent': 'Bytes Sent',
      'bytes_recv': 'Bytes Received',
      'packets_sent': 'Packets Sent',
      'packets_recv': 'Packets Received',
      'read_bytes': 'Read (bytes)',
      'write_bytes': 'Write (bytes)',
      'cpu_percent': 'CPU (%)',
      'memory_percent': 'Memory (%)'
    };

    if (metricName in metricNameOptions) {
      metricNameLabel = metricNameOptions[metricName as keyof typeof metricNameOptions];
    }

    return (
      <Typography variant="body2">
        <strong>{metricTypeOptions[metricType as keyof typeof metricTypeOptions]} {metricNameLabel}</strong>{' '}
        {threshold.operator} {threshold.value}
        {threshold.duration ? ` for ${threshold.duration / 1000}s` : ''}
      </Typography>
    );
  };

  // Render notification information
  const renderNotifications = (alert: AlertConfig) => {
    if (!alert.notifications || alert.notifications.length === 0) {
      return <Typography variant="body2">No notifications configured</Typography>;
    }

    return (
      <Stack direction="row" spacing={1}>
        {alert.notifications.map((notification, index) => {
          if (!notification.enabled) return null;

          return (
            <Chip
              key={index}
              label={notification.type === 'in-app' ? 'In-App' : 'Email'}
              size="small"
              color={notification.type === 'email' ? 'primary' : 'default'}
            />
          );
        })}
      </Stack>
    );
  };

  // Get status timestamp based on alert state
  const getStatusTimestamp = (alertId: string) => {
    const status = alertStatuses[alertId];
    if (!status) return null;

    if (status.state === 'active' && status.triggered_at) {
      return `Triggered at ${formatDate(status.triggered_at)}`;
    } else if (status.state === 'resolved' && status.resolved_at) {
      return `Resolved at ${formatDate(status.resolved_at)}`;
    }

    return null;
  };

  // Define page header actions
  const headerActions = (
    <>
      <Button
        variant="contained"
        color="success"
        startIcon={<AddIcon />}
        onClick={openCreateAlertDialog}
      >
        Create Alert
      </Button>
      <Button
        variant="contained"
        startIcon={<RefreshIcon />}
        onClick={() => { refetch(); fetchAlertStatuses(); }}
        disabled={loading}
        sx={{ ml: 2 }}
      >
        Refresh
      </Button>
    </>
  );

  const formattedLastUpdated = lastUpdated ? formatDate(lastUpdated.toISOString()) : null;

  return (
    <Box sx={{ p: 3 }}>
      {/* Use the PageHeader component */}
      <PageHeader
        title="Alerts"
        description="View and manage alerts. Alerts can be enabled, disabled, or manually triggered."
        lastUpdated={formattedLastUpdated}
        actions={headerActions}
        loading={loading}
      />

      <LoadingErrorHandler loading={loading} error={error} loadingMessage="Loading alerts...">
        <Box>
          {alerts.length === 0 ? (
            <Paper sx={{ p: 3, textAlign: 'center' }}>
              <Typography variant="h6">No alerts configured</Typography>
              <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
                Create your first alert to start monitoring your system
              </Typography>
            </Paper>
          ) : (
            <Grid container spacing={3}>
              {alerts.map((alert) => (
                <Grid item xs={12} md={6} lg={4} key={alert.id}>
                  <Card>
                    <CardContent>
                      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                        <Box>
                          <Typography variant="h6" component="div" sx={{ mb: 0.5 }}>
                            {alert.name}
                          </Typography>
                          <StatusChip 
                            status={alertStatuses[alert.id]?.state || 'inactive'}
                            statusMap={ALERT_STATUS_MAP}
                            sx={{ mr: 1 }} // Add margin right to StatusChip
                          />
                          <StatusChip
                            status={alert.severity}
                            statusMap={ALERT_SEVERITY_MAP}
                          />
                        </Box>
                        <Tooltip title={alert.enabled ? 'Disable' : 'Enable'}>
                          <IconButton
                            onClick={() => handleToggleEnabled(alert)}
                            disabled={actionLoading}
                            color={alert.enabled ? 'primary' : 'default'}
                            size="small"
                          >
                            {alert.enabled ? <NotificationsActiveIcon /> : <NotificationsOffIcon />}
                          </IconButton>
                        </Tooltip>
                      </Box>

                      {alert.description && (
                        <Typography variant="body2" color="text.secondary" sx={{ mt: 1, mb: 2 }}>
                          {alert.description}
                        </Typography>
                      )}

                      <Divider sx={{ my: 1.5 }} />

                      <Typography variant="subtitle2" sx={{ mt: 1 }}>
                        Threshold:
                      </Typography>
                      {renderThreshold(alert)}

                      <Typography variant="subtitle2" sx={{ mt: 1.5 }}>
                        Notifications:
                      </Typography>
                      {renderNotifications(alert)}

                      {getStatusTimestamp(alert.id) && (
                        <Typography variant="body2" color="text.secondary" sx={{ mt: 1.5, fontSize: '0.8rem' }}>
                          {getStatusTimestamp(alert.id)}
                        </Typography>
                      )}

                      <Box sx={{ display: 'flex', justifyContent: 'flex-end', mt: 2 }}>
                        <Tooltip title="Test">
                          <IconButton
                            onClick={() => openTestAlertDialog(alert)}
                            disabled={actionLoading || testActionLoading || !alert.enabled}
                            size="small"
                            sx={{ mr: 1 }}
                          >
                            <PlayArrowIcon />
                          </IconButton>
                        </Tooltip>
                        <Tooltip title="Edit">
                          <IconButton
                            onClick={() => openEditAlertDialog(alert)}
                            disabled={actionLoading}
                            size="small"
                            sx={{ mr: 1 }}
                          >
                            <EditIcon />
                          </IconButton>
                        </Tooltip>
                        <Tooltip title="Delete">
                          <IconButton
                            onClick={() => openDeleteAlertDialog(alert)}
                            disabled={actionLoading}
                            color="error"
                            size="small"
                          >
                            <DeleteIcon />
                          </IconButton>
                        </Tooltip>
                      </Box>
                    </CardContent>
                  </Card>
                </Grid>
              ))}
            </Grid>
          )}
        </Box>
      </LoadingErrorHandler>

      {/* Create Alert Dialog */}
      <AlertDialog
        open={isDialogOpen('create')}
        onClose={() => closeDialog('create')}
        onSave={handleSaveAlert}
        loading={actionLoading}
      />

      {/* Edit Alert Dialog */}
      <AlertDialog
        open={isDialogOpen('edit')}
        onClose={() => closeDialog('edit')}
        onSave={handleUpdateAlert}
        alert={selectedAlert ? { ...selectedAlert } : undefined}
        isEditing={true}
        loading={actionLoading}
      />

      {/* Use the ConfirmDialog component for Delete Confirmation */}
      <ConfirmDialog
        open={isDialogOpen('delete')}
        onClose={() => closeDialog('delete')}
        onConfirm={handleConfirmDelete}
        title="Delete Alert"
        message={`Are you sure you want to delete the alert "${selectedAlert?.name}"? This action cannot be undone.`}
        confirmText="Delete"
        cancelText="Cancel"
        loading={actionLoading}
        severity="error"
      />

      {/* Use the ConfirmDialog component for Test Alert */}
      <ConfirmDialog
        open={isDialogOpen('test')}
        onClose={() => closeDialog('test')}
        onConfirm={handleConfirmTest}
        title="Test Alert"
        message={`This will trigger a test notification for the alert "${selectedAlert?.name}". Do you want to continue?`}
        confirmText="Test Alert"
        cancelText="Cancel"
        loading={testActionLoading}
        severity="info"
      />
    </Box>
  );
};

export default Alerts; 