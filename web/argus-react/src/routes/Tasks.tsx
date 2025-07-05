import React, { useState, useEffect, useCallback } from 'react';
import { 
  Box, 
  Typography, 
  Paper, 
  Table, 
  TableBody, 
  TableCell, 
  TableContainer, 
  TableHead, 
  TableRow,
  Chip,
  Button,
  Card,
  CardContent,
  IconButton,
  Tooltip,
  Stack,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogContentText,
  DialogActions,
  Snackbar,
  Alert,
  TextField,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  FormControlLabel,
  Switch,
  Grid
} from '@mui/material';
import RefreshIcon from '@mui/icons-material/Refresh';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import InfoIcon from '@mui/icons-material/Info';
import AddIcon from '@mui/icons-material/Add';
import { apiClient } from '../api';
import type { TaskInfo, TaskStatus, TaskExecution } from '../types/api';
import LoadingErrorHandler from '../components/LoadingErrorHandler';

// Task type options for the form
const TASK_TYPES = [
  { value: 'log_rotation', label: 'Log Rotation' },
  { value: 'metrics_aggregation', label: 'Metrics Aggregation' },
  { value: 'health_check', label: 'Health Check' },
  { value: 'system_cleanup', label: 'System Cleanup' }
];

/**
 * Tasks page component
 * Displays system tasks and allows management
 */
const Tasks: React.FC = () => {
  const [tasks, setTasks] = useState<TaskInfo[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [lastUpdated, setLastUpdated] = useState<string | null>(null);
  
  // State for task operations
  const [selectedTaskId, setSelectedTaskId] = useState<string | null>(null);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState<boolean>(false);
  const [executionsDialogOpen, setExecutionsDialogOpen] = useState<boolean>(false);
  const [taskExecutions, setTaskExecutions] = useState<TaskExecution[]>([]);
  const [executionsLoading, setExecutionsLoading] = useState<boolean>(false);
  const [executionsError, setExecutionsError] = useState<string | null>(null);
  
  // State for create task dialog
  const [createDialogOpen, setCreateDialogOpen] = useState<boolean>(false);
  const [newTask, setNewTask] = useState<Partial<TaskInfo> & { type: string }>({
    name: '',
    type: 'health_check',
    enabled: true,
    schedule: {
      cron_expression: '0 * * * *', // Default: Run hourly
      one_time: false,
      next_run_time: new Date().toISOString()
    }
  });
  const [createTaskLoading, setCreateTaskLoading] = useState<boolean>(false);
  
  // Snackbar notification state
  const [notification, setNotification] = useState<{
    open: boolean;
    message: string;
    severity: 'success' | 'error' | 'info' | 'warning';
  }>({
    open: false,
    message: '',
    severity: 'info'
  });

  // Function to fetch tasks from the API
  const fetchTasks = useCallback(async () => {
    setLoading(true);
    setError(null);
    
    try {
      const response = await apiClient.getTasks();
      
      if (response.success && response.data) {
        setTasks(response.data);
        setLastUpdated(new Date().toISOString());
      } else {
        setError(response.error || 'Failed to fetch tasks');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An unknown error occurred');
    } finally {
      setLoading(false);
    }
  }, []);

  // Function to get a specific task
  const getTask = useCallback(async (id: string) => {
    try {
      const response = await apiClient.getTask(id);
      
      if (response.success && response.data) {
        return response.data;
      } else {
        throw new Error(response.error || 'Failed to fetch task');
      }
    } catch (err) {
      throw new Error(err instanceof Error ? err.message : 'An unknown error occurred');
    }
  }, []);

  // Function to create a new task
  const createTask = useCallback(async (task: Partial<TaskInfo>) => {
    setCreateTaskLoading(true);
    
    try {
      const response = await apiClient.createTask(task);
      
      if (response.success && response.data) {
        setNotification({
          open: true,
          message: 'Task created successfully',
          severity: 'success'
        });
        fetchTasks(); // Refresh the task list
        setCreateDialogOpen(false);
        // Reset the form
        setNewTask({
          name: '',
          type: 'health_check',
          enabled: true,
          schedule: {
            cron_expression: '0 * * * *',
            one_time: false,
            next_run_time: new Date().toISOString()
          }
        });
      } else {
        throw new Error(response.error || 'Failed to create task');
      }
    } catch (err) {
      setNotification({
        open: true,
        message: err instanceof Error ? err.message : 'Failed to create task',
        severity: 'error'
      });
    } finally {
      setCreateTaskLoading(false);
    }
  }, [fetchTasks]);

  // Function to delete a task
  const deleteTask = useCallback(async (id: string) => {
    try {
      const response = await apiClient.deleteTask(id);
      
      if (response.success) {
        setNotification({
          open: true,
          message: 'Task deleted successfully',
          severity: 'success'
        });
        fetchTasks(); // Refresh the task list
      } else {
        throw new Error(response.error || 'Failed to delete task');
      }
    } catch (err) {
      setNotification({
        open: true,
        message: err instanceof Error ? err.message : 'Failed to delete task',
        severity: 'error'
      });
    } finally {
      setDeleteDialogOpen(false);
      setSelectedTaskId(null);
    }
  }, [fetchTasks]);

  // Function to run a task immediately
  const runTask = useCallback(async (id: string) => {
    try {
      const response = await apiClient.runTask(id);
      
      if (response.success && response.data) {
        setNotification({
          open: true,
          message: 'Task started successfully',
          severity: 'success'
        });
        fetchTasks(); // Refresh to show updated status
      } else {
        // Check for specific error messages
        if (response.error && response.error.includes('not implemented')) {
          setNotification({
            open: true,
            message: 'This task type is not fully implemented in the backend yet. The API endpoint exists but the runner is not implemented.',
            severity: 'warning'
          });
        } else {
          throw new Error(response.error || 'Failed to run task');
        }
      }
    } catch (err) {
      setNotification({
        open: true,
        message: err instanceof Error ? err.message : 'Failed to run task',
        severity: 'error'
      });
    }
  }, [fetchTasks]);

  // Function to get task executions
  const getTaskExecutions = useCallback(async (id: string) => {
    setExecutionsLoading(true);
    setExecutionsError(null);
    
    try {
      const response = await apiClient.getTaskExecutions(id);
      
      if (response.success && response.data) {
        setTaskExecutions(response.data);
        
        // If there are no executions, show a helpful message
        if (response.data.length === 0) {
          setExecutionsError("No execution history found. The task may not have been run yet or the runner might not be fully implemented.");
        }
      } else {
        throw new Error(response.error || 'Failed to fetch task executions');
      }
    } catch (err) {
      setExecutionsError(err instanceof Error ? err.message : 'An unknown error occurred');
      setTaskExecutions([]);
    } finally {
      setExecutionsLoading(false);
    }
  }, []);

  // Handle opening executions dialog
  const handleViewExecutions = useCallback((id: string) => {
    setSelectedTaskId(id);
    getTaskExecutions(id);
    setExecutionsDialogOpen(true);
  }, [getTaskExecutions]);

  // Handle opening create task dialog
  const handleOpenCreateDialog = useCallback(() => {
    setCreateDialogOpen(true);
  }, []);

  // Handle form field changes
  const handleNewTaskChange = useCallback((field: string, value: string | boolean) => {
    if (field === 'cron_expression') {
      // Handle cron_expression field which is now inside schedule
      setNewTask(prev => ({
        ...prev,
        schedule: {
          ...prev.schedule!,
          cron_expression: value as string
        }
      }));
    } else if (field === 'one_time') {
      // Handle one_time field which is now inside schedule
      setNewTask(prev => ({
        ...prev,
        schedule: {
          ...prev.schedule!,
          one_time: value as boolean
        }
      }));
    } else {
      // Handle other fields directly on the task object
      setNewTask(prev => ({
        ...prev,
        [field]: value
      }));
    }
  }, []);

  // Handle form submission
  const handleCreateTask = useCallback((e: React.FormEvent) => {
    e.preventDefault();
    createTask(newTask);
  }, [createTask, newTask]);

  // Fetch tasks on component mount
  useEffect(() => {
    fetchTasks();
  }, [fetchTasks]);

  // Function to render status chip with appropriate color
  const renderStatusChip = (status: TaskStatus) => {
    let color: 'default' | 'primary' | 'secondary' | 'error' | 'info' | 'success' | 'warning' = 'default';
    
    switch (status) {
      case 'pending':
        color = 'info';
        break;
      case 'running':
        color = 'primary';
        break;
      case 'completed':
        color = 'success';
        break;
      case 'failed':
        color = 'error';
        break;
    }
    
    return <Chip label={status} color={color} size="small" />;
  };

  // Format date for display
  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  // Handle close notification
  const handleCloseNotification = () => {
    setNotification({
      ...notification,
      open: false
    });
  };

  return (
    <Box sx={{ p: 3 }}>
      <Paper sx={{ p: 3, mb: 4 }}>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
          <Typography variant="h4" gutterBottom>
            System Tasks
          </Typography>
          <Stack direction="row" spacing={2}>
            <Button 
              variant="contained" 
              color="success"
              startIcon={<AddIcon />}
              onClick={handleOpenCreateDialog}
            >
              Create Task
            </Button>
            <Button 
              variant="contained" 
              startIcon={<RefreshIcon />} 
              onClick={fetchTasks}
              disabled={loading}
            >
              Refresh
            </Button>
          </Stack>
        </Box>
        <Typography variant="body1" paragraph>
          View and manage scheduled system tasks. Tasks can be enabled, disabled, or manually triggered.
        </Typography>
        {lastUpdated && (
          <Typography variant="caption" color="text.secondary">
            Last updated: {formatDate(lastUpdated)}
          </Typography>
        )}
      </Paper>

      <LoadingErrorHandler loading={loading} error={error} loadingMessage="Loading tasks...">
        <Card>
          <CardContent>
            <TableContainer>
              <Table aria-label="tasks table">
                <TableHead>
                  <TableRow>
                    <TableCell>Name</TableCell>
                    <TableCell>Type</TableCell>
                    <TableCell>Status</TableCell>
                    <TableCell>Created</TableCell>
                    <TableCell>Enabled</TableCell>
                    <TableCell align="right">Actions</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {tasks.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={6} align="center">
                        <Box sx={{ py: 3, display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 2 }}>
                          <Typography variant="body1">
                            No tasks found
                          </Typography>
                          <Button 
                            variant="contained" 
                            color="primary"
                            onClick={handleOpenCreateDialog}
                          >
                            Create Your First Task
                          </Button>
                        </Box>
                      </TableCell>
                    </TableRow>
                  ) : (
                    tasks.map((task) => (
                      <TableRow key={task.id}>
                        <TableCell>{task.name}</TableCell>
                        <TableCell>{task.type}</TableCell>
                        <TableCell>{renderStatusChip(task.status)}</TableCell>
                        <TableCell>{formatDate(task.created_at)}</TableCell>
                        <TableCell>
                          <Chip 
                            label={task.enabled ? 'Enabled' : 'Disabled'} 
                            color={task.enabled ? 'success' : 'default'} 
                            size="small" 
                          />
                        </TableCell>
                        <TableCell align="right">
                          <Stack direction="row" spacing={1} justifyContent="flex-end">
                            <Tooltip title="Run task">
                              <IconButton 
                                size="small" 
                                color="primary"
                                onClick={() => runTask(task.id)}
                              >
                                <PlayArrowIcon fontSize="small" />
                              </IconButton>
                            </Tooltip>
                            <Tooltip title="View executions">
                              <IconButton 
                                size="small" 
                                color="info"
                                onClick={() => handleViewExecutions(task.id)}
                              >
                                <InfoIcon fontSize="small" />
                              </IconButton>
                            </Tooltip>
                            <Tooltip title="Edit task">
                              <IconButton 
                                size="small" 
                                color="primary"
                                onClick={() => {
                                  // This would be implemented in a future feature with a form dialog
                                  console.log('Edit task:', task.id);
                                }}
                              >
                                <EditIcon fontSize="small" />
                              </IconButton>
                            </Tooltip>
                            <Tooltip title="Delete task">
                              <IconButton 
                                size="small" 
                                color="error"
                                onClick={() => {
                                  setSelectedTaskId(task.id);
                                  setDeleteDialogOpen(true);
                                }}
                              >
                                <DeleteIcon fontSize="small" />
                              </IconButton>
                            </Tooltip>
                          </Stack>
                        </TableCell>
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            </TableContainer>
          </CardContent>
        </Card>
      </LoadingErrorHandler>

      {/* Create Task Dialog */}
      <Dialog 
        open={createDialogOpen} 
        onClose={() => !createTaskLoading && setCreateDialogOpen(false)}
        maxWidth="md"
        fullWidth
      >
        <form onSubmit={handleCreateTask}>
          <DialogTitle>Create New Task</DialogTitle>
          <DialogContent>
            <Grid container spacing={3} sx={{ mt: 1 }}>
              <Grid item xs={12}>
                <TextField
                  required
                  fullWidth
                  label="Task Name"
                  value={newTask.name}
                  onChange={(e) => handleNewTaskChange('name', e.target.value)}
                  disabled={createTaskLoading}
                />
              </Grid>
              <Grid item xs={12} sm={6}>
                <FormControl fullWidth required>
                  <InputLabel>Task Type</InputLabel>
                  <Select<string>
                    value={newTask.type}
                    label="Task Type"
                    onChange={(e) => handleNewTaskChange('type', e.target.value)}
                    disabled={createTaskLoading}
                  >
                    {TASK_TYPES.map(option => (
                      <MenuItem key={option.value} value={option.value}>
                        {option.label}
                      </MenuItem>
                    ))}
                  </Select>
                </FormControl>
              </Grid>
              <Grid item xs={12} sm={6}>
                <TextField
                  fullWidth
                  label="Cron Expression"
                  value={newTask.schedule?.cron_expression || ''}
                  onChange={(e) => handleNewTaskChange('cron_expression', e.target.value)}
                  disabled={createTaskLoading}
                  helperText="e.g., '0 * * * *' for hourly execution"
                />
              </Grid>
              <Grid item xs={12} sm={6}>
                <FormControlLabel
                  control={
                    <Switch
                      checked={!!newTask.schedule?.one_time}
                      onChange={(e) => handleNewTaskChange('one_time', e.target.checked)}
                      disabled={createTaskLoading}
                    />
                  }
                  label="One-time task"
                />
              </Grid>
              <Grid item xs={12}>
                <FormControlLabel
                  control={
                    <Switch
                      checked={!!newTask.enabled}
                      onChange={(e) => handleNewTaskChange('enabled', e.target.checked)}
                      disabled={createTaskLoading}
                    />
                  }
                  label="Enabled"
                />
              </Grid>
            </Grid>
          </DialogContent>
          <DialogActions>
            <Button 
              onClick={() => setCreateDialogOpen(false)} 
              disabled={createTaskLoading}
            >
              Cancel
            </Button>
            <Button 
              type="submit" 
              variant="contained" 
              color="primary"
              disabled={createTaskLoading || !newTask.name || !newTask.type}
            >
              {createTaskLoading ? 'Creating...' : 'Create Task'}
            </Button>
          </DialogActions>
        </form>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <Dialog
        open={deleteDialogOpen}
        onClose={() => setDeleteDialogOpen(false)}
        aria-labelledby="delete-dialog-title"
        aria-describedby="delete-dialog-description"
      >
        <DialogTitle id="delete-dialog-title">Delete Task</DialogTitle>
        <DialogContent>
          <DialogContentText id="delete-dialog-description">
            Are you sure you want to delete this task? This action cannot be undone.
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteDialogOpen(false)}>Cancel</Button>
          <Button 
            onClick={() => selectedTaskId && deleteTask(selectedTaskId)} 
            color="error" 
            autoFocus
          >
            Delete
          </Button>
        </DialogActions>
      </Dialog>

      {/* Task Executions Dialog */}
      <Dialog
        open={executionsDialogOpen}
        onClose={() => setExecutionsDialogOpen(false)}
        aria-labelledby="executions-dialog-title"
        maxWidth="md"
        fullWidth
      >
        <DialogTitle id="executions-dialog-title">Task Execution History</DialogTitle>
        <DialogContent>
          <LoadingErrorHandler loading={executionsLoading} error={executionsError} loadingMessage="Loading executions...">
            {taskExecutions.length === 0 ? (
              <Typography variant="body1" sx={{ py: 2 }}>
                No execution history found for this task.
              </Typography>
            ) : (
              <TableContainer>
                <Table aria-label="executions table">
                  <TableHead>
                    <TableRow>
                      <TableCell>Execution ID</TableCell>
                      <TableCell>Status</TableCell>
                      <TableCell>Start Time</TableCell>
                      <TableCell>End Time</TableCell>
                      <TableCell>Output/Error</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {taskExecutions.map((execution) => (
                      <TableRow key={execution.id}>
                        <TableCell>{execution.id}</TableCell>
                        <TableCell>{renderStatusChip(execution.status)}</TableCell>
                        <TableCell>{formatDate(execution.start_time)}</TableCell>
                        <TableCell>{execution.end_time ? formatDate(execution.end_time) : 'N/A'}</TableCell>
                        <TableCell>
                          {execution.error ? (
                            <Typography color="error" variant="body2">
                              {execution.error}
                            </Typography>
                          ) : execution.output ? (
                            <Typography variant="body2">
                              {execution.output.length > 50 
                                ? `${execution.output.substring(0, 50)}...` 
                                : execution.output}
                            </Typography>
                          ) : 'No output'}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </TableContainer>
            )}
          </LoadingErrorHandler>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setExecutionsDialogOpen(false)}>Close</Button>
        </DialogActions>
      </Dialog>

      {/* Notification Snackbar */}
      <Snackbar 
        open={notification.open} 
        autoHideDuration={6000} 
        onClose={handleCloseNotification}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
      >
        <Alert 
          onClose={handleCloseNotification} 
          severity={notification.severity} 
          sx={{ width: '100%' }}
        >
          {notification.message}
        </Alert>
      </Snackbar>
    </Box>
  );
};

export default Tasks; 