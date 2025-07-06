import React, { useState, useEffect } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  TextField,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  FormControlLabel,
  Switch,
  Grid,
  Typography,
  Divider,
  Box,
  FormHelperText,
  IconButton,
  InputAdornment
} from '@mui/material';
import CloseIcon from '@mui/icons-material/Close';
import AddIcon from '@mui/icons-material/Add';
import DeleteIcon from '@mui/icons-material/Delete';
import type { 
  AlertConfig, 
  MetricType, 
  ComparisonOperator, 
  AlertSeverity,
  NotificationType,
  ThresholdConfig,
  NotificationConfig
} from '../types/api';

// Default values for a new alert
const DEFAULT_ALERT: Partial<AlertConfig> = {
  name: '',
  description: '',
  enabled: true,
  severity: 'warning',
  threshold: {
    metric_type: 'cpu',
    metric_name: 'usage_percent',
    operator: '>',
    value: 80
  },
  notifications: [
    {
      type: 'in-app',
      enabled: true,
      settings: {}
    }
  ]
};

// Metric name options based on metric type
const METRIC_NAME_OPTIONS: Record<MetricType, { value: string, label: string }[]> = {
  'cpu': [
    { value: 'usage_percent', label: 'CPU Usage (%)' },
    { value: 'load1', label: 'Load Average (1m)' },
    { value: 'load5', label: 'Load Average (5m)' },
    { value: 'load15', label: 'Load Average (15m)' }
  ],
  'memory': [
    { value: 'used_percent', label: 'Memory Usage (%)' },
    { value: 'used', label: 'Used Memory (bytes)' },
    { value: 'free', label: 'Free Memory (bytes)' }
  ],
  'load': [
    { value: 'load1', label: 'Load Average (1m)' },
    { value: 'load5', label: 'Load Average (5m)' },
    { value: 'load15', label: 'Load Average (15m)' }
  ],
  'network': [
    { value: 'bytes_sent', label: 'Bytes Sent' },
    { value: 'bytes_recv', label: 'Bytes Received' },
    { value: 'packets_sent', label: 'Packets Sent' },
    { value: 'packets_recv', label: 'Packets Received' }
  ],
  'disk': [
    { value: 'usage_percent', label: 'Disk Usage (%)' },
    { value: 'read_bytes', label: 'Read Bytes' },
    { value: 'write_bytes', label: 'Write Bytes' }
  ],
  'process': [
    { value: 'cpu_percent', label: 'Process CPU Usage (%)' },
    { value: 'memory_percent', label: 'Process Memory Usage (%)' }
  ]
};

interface AlertDialogProps {
  open: boolean;
  onClose: () => void;
  onSave: (alert: Partial<AlertConfig>) => Promise<void>;
  alert?: Partial<AlertConfig> | undefined;
  isEditing?: boolean;
  loading?: boolean;
}

const AlertDialog: React.FC<AlertDialogProps> = ({
  open,
  onClose,
  onSave,
  alert,
  isEditing = false,
  loading = false
}) => {
  // Initialize alert state with default values or provided alert
  const [alertData, setAlertData] = useState<Partial<AlertConfig>>(
    alert || DEFAULT_ALERT
  );
  
  // Form validation state
  const [errors, setErrors] = useState<Record<string, string>>({});

  // Update alert data when the prop changes
  useEffect(() => {
    if (alert) {
      setAlertData(alert);
    } else {
      setAlertData(DEFAULT_ALERT);
    }
  }, [alert, open]);

  // Handle field changes
  const handleChange = (field: string, value: any) => {
    setAlertData(prev => ({
      ...prev,
      [field]: value
    }));
    
    // Clear error for this field if exists
    if (errors[field]) {
      setErrors(prev => {
        const newErrors = { ...prev };
        delete newErrors[field];
        return newErrors;
      });
    }
  };

  // Handle threshold field changes
  const handleThresholdChange = (field: string, value: any) => {
    setAlertData(prev => ({
      ...prev,
      threshold: {
        ...prev.threshold,
        [field]: value
      }
    }));
    
    // Clear error for this field if exists
    if (errors[`threshold.${field}`]) {
      setErrors(prev => {
        const newErrors = { ...prev };
        delete newErrors[`threshold.${field}`];
        return newErrors;
      });
    }
    
    // Update metric name options when metric type changes
    if (field === 'metric_type') {
      const metricType = value as MetricType;
      const options = METRIC_NAME_OPTIONS[metricType];
      if (options && options.length > 0) {
        setAlertData(prev => ({
          ...prev,
          threshold: {
            ...prev.threshold,
            metric_type: metricType,
            metric_name: options[0].value
          }
        }));
      }
    }
  };

  // Handle notification changes
  const handleNotificationChange = (index: number, field: string, value: any) => {
    setAlertData(prev => {
      const notifications = [...(prev.notifications || [])];
      notifications[index] = {
        ...notifications[index],
        [field]: value
      };
      return {
        ...prev,
        notifications
      };
    });
  };

  const handleNotificationSettingsChange = (index: number, key: string, value: string) => {
    setAlertData(prev => {
      const notifications = [...(prev.notifications || [])];
      if (!notifications[index].settings) {
        notifications[index].settings = {};
      }
      notifications[index].settings![key] = value;
      return {
        ...prev,
        notifications
      };
    });
  };

  // Add a new notification
  const addNotification = () => {
    setAlertData(prev => ({
      ...prev,
      notifications: [
        ...(prev.notifications || []),
        {
          type: 'in-app',
          enabled: true,
          settings: {}
        }
      ]
    }));
  };

  // Remove a notification
  const removeNotification = (index: number) => {
    setAlertData(prev => {
      const notifications = [...(prev.notifications || [])];
      notifications.splice(index, 1);
      return {
        ...prev,
        notifications
      };
    });
  };

  // Validate form before submission
  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};
    
    // Required fields
    if (!alertData.name) {
      newErrors.name = 'Alert name is required';
    }
    
    if (!alertData.threshold?.metric_type) {
      newErrors['threshold.metric_type'] = 'Metric type is required';
    }
    
    if (!alertData.threshold?.metric_name) {
      newErrors['threshold.metric_name'] = 'Metric name is required';
    }
    
    if (!alertData.threshold?.operator) {
      newErrors['threshold.operator'] = 'Operator is required';
    }
    
    if (alertData.threshold?.value === undefined || alertData.threshold.value === null) {
      newErrors['threshold.value'] = 'Threshold value is required';
    }
    
    if (alertData.threshold?.metric_type === 'process' && (!alertData.threshold.target || alertData.threshold.target.trim() === '')) {
      newErrors['threshold.target'] = 'Process target (name or PID) is required';
    }
    
    alertData.notifications?.forEach((notification, index) => {
      if (notification.type === 'email' && (!notification.settings?.recipient || String(notification.settings.recipient).trim() === '')) {
        newErrors[`notification.${index}.recipient`] = 'Recipient email is required';
      }
    });
    
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  // Handle form submission
  const handleSubmit = async () => {
    if (!validateForm()) {
      return;
    }
    
    try {
      await onSave(alertData);
    } catch (error) {
      console.error('Error saving alert:', error);
    }
  };

  return (
    <Dialog
      open={open}
      onClose={onClose}
      maxWidth="md"
      fullWidth
      aria-labelledby="alert-dialog-title"
    >
      <DialogTitle id="alert-dialog-title">
        {isEditing ? 'Edit Alert' : 'Create New Alert'}
        <IconButton
          aria-label="close"
          onClick={onClose}
          sx={{
            position: 'absolute',
            right: 8,
            top: 8,
            color: (theme) => theme.palette.grey[500],
          }}
        >
          <CloseIcon />
        </IconButton>
      </DialogTitle>
      <DialogContent dividers>
        <Grid container spacing={3}>
          {/* Basic Information */}
          <Grid item xs={12}>
            <Typography variant="subtitle1" gutterBottom>
              Basic Information
            </Typography>
            <Grid container spacing={2}>
              <Grid item xs={12} md={6}>
                <TextField
                  label="Alert Name"
                  value={alertData.name || ''}
                  onChange={(e) => handleChange('name', e.target.value)}
                  fullWidth
                  required
                  error={!!errors.name}
                  helperText={errors.name}
                  disabled={loading}
                />
              </Grid>
              <Grid item xs={12} md={6}>
                <FormControl fullWidth>
                  <InputLabel id="alert-severity-label">Severity</InputLabel>
                  <Select
                    labelId="alert-severity-label"
                    value={alertData.severity || 'warning'}
                    onChange={(e) => handleChange('severity', e.target.value)}
                    label="Severity"
                    disabled={loading}
                  >
                    <MenuItem value="info">Info</MenuItem>
                    <MenuItem value="warning">Warning</MenuItem>
                    <MenuItem value="critical">Critical</MenuItem>
                  </Select>
                </FormControl>
              </Grid>
              <Grid item xs={12}>
                <TextField
                  label="Description"
                  value={alertData.description || ''}
                  onChange={(e) => handleChange('description', e.target.value)}
                  fullWidth
                  multiline
                  rows={2}
                  disabled={loading}
                />
              </Grid>
              <Grid item xs={12}>
                <FormControlLabel
                  control={
                    <Switch
                      checked={alertData.enabled ?? true}
                      onChange={(e) => handleChange('enabled', e.target.checked)}
                      disabled={loading}
                    />
                  }
                  label="Enable Alert"
                />
              </Grid>
            </Grid>
          </Grid>
          
          {/* Threshold Configuration */}
          <Grid item xs={12}>
            <Divider sx={{ my: 2 }} />
            <Typography variant="subtitle1" gutterBottom>
              Threshold Configuration
            </Typography>
            <Grid container spacing={2}>
              <Grid item xs={12} md={6}>
                <FormControl fullWidth error={!!errors['threshold.metric_type']}>
                  <InputLabel id="metric-type-label">Metric Type</InputLabel>
                  <Select
                    labelId="metric-type-label"
                    value={alertData.threshold?.metric_type || 'cpu'}
                    onChange={(e) => handleThresholdChange('metric_type', e.target.value)}
                    label="Metric Type"
                    disabled={loading}
                  >
                    <MenuItem value="cpu">CPU</MenuItem>
                    <MenuItem value="memory">Memory</MenuItem>
                    <MenuItem value="load">System Load</MenuItem>
                    <MenuItem value="network">Network</MenuItem>
                    <MenuItem value="disk">Disk</MenuItem>
                    <MenuItem value="process">Process</MenuItem>
                  </Select>
                  {errors['threshold.metric_type'] && (
                    <FormHelperText>{errors['threshold.metric_type']}</FormHelperText>
                  )}
                </FormControl>
              </Grid>
              <Grid item xs={12} md={6}>
                <FormControl fullWidth error={!!errors['threshold.metric_name']}>
                  <InputLabel id="metric-name-label">Metric Name</InputLabel>
                  <Select
                    labelId="metric-name-label"
                    value={alertData.threshold?.metric_name || ''}
                    onChange={(e) => handleThresholdChange('metric_name', e.target.value)}
                    label="Metric Name"
                    disabled={loading}
                  >
                    {alertData.threshold?.metric_type &&
                      METRIC_NAME_OPTIONS[alertData.threshold.metric_type as MetricType].map((option) => (
                        <MenuItem key={option.value} value={option.value}>
                          {option.label}
                        </MenuItem>
                      ))}
                  </Select>
                  {errors['threshold.metric_name'] && (
                    <FormHelperText>{errors['threshold.metric_name']}</FormHelperText>
                  )}
                </FormControl>
              </Grid>
              <Grid item xs={12} md={4}>
                <FormControl fullWidth error={!!errors['threshold.operator']}>
                  <InputLabel id="operator-label">Operator</InputLabel>
                  <Select
                    labelId="operator-label"
                    value={alertData.threshold?.operator || '>'}
                    onChange={(e) => handleThresholdChange('operator', e.target.value)}
                    label="Operator"
                    disabled={loading}
                  >
                    <MenuItem value=">">Greater than (&gt;)</MenuItem>
                    <MenuItem value=">=">Greater than or equal (&gt;=)</MenuItem>
                    <MenuItem value="<">Less than (&lt;)</MenuItem>
                    <MenuItem value="<=">Less than or equal (&lt;=)</MenuItem>
                    <MenuItem value="==">Equal to (==)</MenuItem>
                    <MenuItem value="!=">Not equal to (!=)</MenuItem>
                  </Select>
                  {errors['threshold.operator'] && (
                    <FormHelperText>{errors['threshold.operator']}</FormHelperText>
                  )}
                </FormControl>
              </Grid>
              <Grid item xs={12} md={4}>
                <TextField
                  label="Threshold Value"
                  type="number"
                  value={alertData.threshold?.value ?? ''}
                  onChange={(e) => handleThresholdChange('value', parseFloat(e.target.value))}
                  fullWidth
                  required
                  error={!!errors['threshold.value']}
                  helperText={errors['threshold.value']}
                  disabled={loading}
                  InputProps={{
                    inputProps: { min: 0 },
                  }}
                />
              </Grid>
              <Grid item xs={12} md={4}>
                <TextField
                  label="Duration (seconds)"
                  type="number"
                  value={alertData.threshold?.duration ?? ''}
                  onChange={(e) => handleThresholdChange('duration', parseInt(e.target.value, 10))}
                  fullWidth
                  disabled={loading}
                  InputProps={{
                    inputProps: { min: 0 },
                    endAdornment: <InputAdornment position="end">sec</InputAdornment>,
                  }}
                  helperText="Optional: Sustained duration for alert"
                />
              </Grid>
              <Grid item xs={12}>
                <TextField
                  fullWidth
                  label="Process Name or PID"
                  value={alertData.threshold?.target || ''}
                  onChange={(e) => handleThresholdChange('target', e.target.value)}
                  error={!!errors['threshold.target']}
                  helperText={errors['threshold.target']}
                  disabled={loading}
                />
              </Grid>
            </Grid>
          </Grid>
          
          {/* Notification Configuration */}
          <Grid item xs={12}>
            <Divider sx={{ my: 2 }} />
            <Box display="flex" justifyContent="space-between" alignItems="center" mb={2}>
              <Typography variant="subtitle1">
                Notification Configuration
              </Typography>
              <Button
                startIcon={<AddIcon />}
                onClick={addNotification}
                variant="outlined"
                size="small"
                disabled={loading}
              >
                Add Notification
              </Button>
            </Box>
            
            {alertData.notifications?.map((notification, index) => (
              <Box key={index} sx={{ mb: 2, p: 2, border: '1px solid #e0e0e0', borderRadius: 1 }}>
                <Grid container spacing={2} alignItems="center">
                  <Grid item xs={12} md={4}>
                    <FormControl fullWidth>
                      <InputLabel id={`notification-type-label-${index}`}>Type</InputLabel>
                      <Select
                        labelId={`notification-type-label-${index}`}
                        value={notification.type || 'in-app'}
                        onChange={(e) => handleNotificationChange(index, 'type', e.target.value)}
                        label="Type"
                        disabled={loading}
                      >
                        <MenuItem value="in-app">In-App</MenuItem>
                        <MenuItem value="email">Email</MenuItem>
                      </Select>
                    </FormControl>
                  </Grid>
                  <Grid item xs={12} md={4}>
                    <FormControlLabel
                      control={
                        <Switch
                          checked={notification.enabled ?? true}
                          onChange={(e) => handleNotificationChange(index, 'enabled', e.target.checked)}
                          disabled={loading}
                        />
                      }
                      label="Enable"
                    />
                  </Grid>
                  <Grid item xs={12} md={4} sx={{ display: 'flex', justifyContent: 'flex-end' }}>
                    <IconButton
                      onClick={() => removeNotification(index)}
                      disabled={alertData.notifications?.length === 1 || loading}
                      color="error"
                    >
                      <DeleteIcon />
                    </IconButton>
                  </Grid>
                  
                  {/* Email-specific settings */}
                  {notification.type === 'email' && (
                    <Grid item xs={12}>
                      <TextField
                        fullWidth
                        label="Recipient Email"
                        value={notification.settings?.recipient as string || ''}
                        onChange={(e) => handleNotificationSettingsChange(index, 'recipient', e.target.value)}
                        error={!!errors[`notification.${index}.recipient`]}
                        helperText={errors[`notification.${index}.recipient`]}
                        disabled={loading || !notification.enabled}
                      />
                    </Grid>
                  )}
                </Grid>
              </Box>
            ))}
          </Grid>
        </Grid>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose} disabled={loading}>
          Cancel
        </Button>
        <Button 
          onClick={handleSubmit} 
          variant="contained" 
          color="primary"
          disabled={loading}
        >
          {loading ? 'Saving...' : isEditing ? 'Update' : 'Create'}
        </Button>
      </DialogActions>
    </Dialog>
  );
};

export default AlertDialog; 