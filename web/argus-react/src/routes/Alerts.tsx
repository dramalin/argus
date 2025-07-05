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
  Alert as MuiAlert
} from '@mui/material';
import RefreshIcon from '@mui/icons-material/Refresh';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import NotificationsActiveIcon from '@mui/icons-material/NotificationsActive';
import NotificationsOffIcon from '@mui/icons-material/NotificationsOff';
import AddIcon from '@mui/icons-material/Add';
import { apiClient } from '../api';
import type { AlertInfo } from '../types/api';
import LoadingErrorHandler from '../components/LoadingErrorHandler';

/**
 * Alerts page component
 * Displays system alerts and allows management
 */
const Alerts: React.FC = () => {
  const [alerts, setAlerts] = useState<AlertInfo[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [lastUpdated, setLastUpdated] = useState<string | null>(null);
  const [createDialogOpen, setCreateDialogOpen] = useState<boolean>(false);

  // Function to fetch alerts from the API
  const fetchAlerts = useCallback(async () => {
    setLoading(true);
    setError(null);
    
    try {
      const response = await apiClient.getAlerts();
      
      if (response.success && response.data) {
        setAlerts(response.data);
        setLastUpdated(new Date().toISOString());
      } else {
        setError(response.error || 'Failed to fetch alerts');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An unknown error occurred');
    } finally {
      setLoading(false);
    }
  }, []);

  // Fetch alerts on component mount
  useEffect(() => {
    fetchAlerts();
  }, [fetchAlerts]);

  // Handle opening the create alert dialog
  const handleCreateAlert = () => {
    setCreateDialogOpen(true);
    // TODO: Implement create alert dialog component
    console.log('Create alert dialog should open');
  };

  // Format date for display
  const formatDate = (dateString?: string) => {
    if (!dateString) return 'Never';
    return new Date(dateString).toLocaleString();
  };

  // Render alert conditions as readable text
  const renderConditions = (conditions: Record<string, unknown>) => {
    try {
      // Convert conditions object to readable text
      const entries = Object.entries(conditions);
      if (entries.length === 0) return 'No conditions defined';
      
      return entries.map(([key, value]) => (
        <Typography key={key} variant="body2" component="div" sx={{ mb: 0.5 }}>
          <strong>{key}:</strong> {JSON.stringify(value)}
        </Typography>
      ));
    } catch (err) {
      return 'Invalid conditions format';
    }
  };

  return (
    <Box sx={{ p: 3 }}>
      <Paper sx={{ p: 3, mb: 4 }}>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
          <Typography variant="h4" gutterBottom>
            System Alerts
          </Typography>
          <Stack direction="row" spacing={2}>
            <Button 
              variant="contained" 
              color="primary"
              startIcon={<AddIcon />}
              onClick={handleCreateAlert}
            >
              Create Alert
            </Button>
            <Button 
              variant="outlined" 
              startIcon={<RefreshIcon />} 
              onClick={fetchAlerts}
              disabled={loading}
            >
              Refresh
            </Button>
          </Stack>
        </Box>
        <Typography variant="body1" paragraph>
          View and manage system alerts. Alerts can be configured to notify you when specific conditions are met.
        </Typography>
        {lastUpdated && (
          <Typography variant="caption" color="text.secondary">
            Last updated: {formatDate(lastUpdated)}
          </Typography>
        )}
      </Paper>

      <LoadingErrorHandler loading={loading} error={error} loadingMessage="Loading alerts...">
        {alerts.length === 0 ? (
          <MuiAlert severity="info" sx={{ mb: 2 }}>
            No alerts configured. Create an alert to get notified about system events.
          </MuiAlert>
        ) : (
          <Grid container spacing={3}>
            {alerts.map((alert) => (
              <Grid item xs={12} md={6} key={alert.id}>
                <Card>
                  <CardContent>
                    <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', mb: 2 }}>
                      <Box>
                        <Typography variant="h6" gutterBottom>
                          {alert.name}
                        </Typography>
                        <Stack direction="row" spacing={1} sx={{ mb: 1 }}>
                          <Chip 
                            label={alert.type} 
                            size="small" 
                            color="primary" 
                          />
                          <Chip 
                            label={alert.enabled ? 'Enabled' : 'Disabled'} 
                            size="small" 
                            color={alert.enabled ? 'success' : 'default'} 
                          />
                          {alert.triggered_at && (
                            <Chip 
                              label="Triggered" 
                              size="small" 
                              color="error" 
                              icon={<NotificationsActiveIcon />}
                            />
                          )}
                        </Stack>
                      </Box>
                      <Stack direction="row" spacing={1}>
                        <Tooltip title={alert.enabled ? 'Disable alert' : 'Enable alert'}>
                          <IconButton size="small" color={alert.enabled ? 'default' : 'primary'}>
                            {alert.enabled ? <NotificationsOffIcon /> : <NotificationsActiveIcon />}
                          </IconButton>
                        </Tooltip>
                        <Tooltip title="Edit alert">
                          <IconButton size="small" color="primary">
                            <EditIcon />
                          </IconButton>
                        </Tooltip>
                        <Tooltip title="Delete alert">
                          <IconButton size="small" color="error">
                            <DeleteIcon />
                          </IconButton>
                        </Tooltip>
                      </Stack>
                    </Box>
                    
                    <Divider sx={{ mb: 2 }} />
                    
                    <Typography variant="subtitle2" gutterBottom>
                      Conditions:
                    </Typography>
                    <Box sx={{ mb: 2, pl: 1 }}>
                      {renderConditions(alert.conditions)}
                    </Box>
                    
                    <Typography variant="subtitle2" gutterBottom>
                      Actions:
                    </Typography>
                    <Box sx={{ mb: 2, pl: 1 }}>
                      {Object.keys(alert.actions).length > 0 ? (
                        Object.entries(alert.actions).map(([key, value]) => (
                          <Typography key={key} variant="body2" component="div" sx={{ mb: 0.5 }}>
                            <strong>{key}:</strong> {JSON.stringify(value)}
                          </Typography>
                        ))
                      ) : (
                        <Typography variant="body2">No actions defined</Typography>
                      )}
                    </Box>
                    
                    <Box sx={{ display: 'flex', justifyContent: 'space-between', mt: 2 }}>
                      <Typography variant="caption" color="text.secondary">
                        Created: {formatDate(alert.created_at)}
                      </Typography>
                      {alert.triggered_at && (
                        <Typography variant="caption" color="error">
                          Last triggered: {formatDate(alert.triggered_at)}
                        </Typography>
                      )}
                    </Box>
                  </CardContent>
                </Card>
              </Grid>
            ))}
          </Grid>
        )}
      </LoadingErrorHandler>

      {/* TODO: Create Alert Dialog Component would be rendered here */}
    </Box>
  );
};

export default Alerts; 