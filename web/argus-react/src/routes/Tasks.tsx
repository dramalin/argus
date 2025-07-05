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
  DialogActions,
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
import { PageHeader, ConfirmDialog } from '../components/common';
import { useNotification, useDataFetching, useDialogState, useDateFormatter } from '../hooks';

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
  // Use the data fetching hook for tasks
  const { data, loading, error, lastUpdated, refetch } = useDataFetching<TaskInfo[]>(
    'tasks',
    apiClient.getTasks,
    { cacheTTL: 30000 }
  );
  
  // Ensure tasks is never null
  const tasks = data || [];
  
  // Use the notification hook for managing notifications
  const { showNotification } = useNotification();
  
  // Use the date formatter hook for consistent date formatting
  const { formatDate } = useDateFormatter();
  
  // Use the dialog state hook for managing multiple dialogs
  const { 
    openDialog, 
    closeDialog, 
    isDialogOpen 
  } = useDialogState<'create' | 'delete' | 'executions'>();
  
  // State for task operations
  const [selectedTaskId, setSelectedTaskId] = useState<string | null>(null);
  const [taskExecutions, setTaskExecutions] = useState<TaskExecution[]>([]);
  const [executionsLoading, setExecutionsLoading] = useState<boolean>(false);
  const [executionsError, setExecutionsError] = useState<string | null>(null);
  
  // State for create task dialog
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
        showNotification('Task created successfully', 'success');
        await refetch(); // Refresh the task list
        closeDialog('create');
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
      showNotification(
        err instanceof Error ? err.message : 'Failed to create task',
        'error'
      );
    } finally {
      setCreateTaskLoading(false);
    }
  }, [refetch, closeDialog, showNotification]);

  // Function to delete a task
  const deleteTask = useCallback(async (id: string) => {
    try {
      const response = await apiClient.deleteTask(id);
      
      if (response.success) {
        showNotification('Task deleted successfully', 'success');
        await refetch(); // Refresh the task list
      } else {
        throw new Error(response.error || 'Failed to delete task');
      }
    } catch (err) {
      showNotification(
        err instanceof Error ? err.message : 'Failed to delete task',
        'error'
      );
    } finally {
      closeDialog('delete');
      setSelectedTaskId(null);
    }
  }, [refetch, closeDialog, showNotification]);

  // Function to run a task immediately
  const runTask = useCallback(async (id: string) => {
    try {
      const response = await apiClient.runTask(id);
      
      if (response.success && response.data) {
        showNotification('Task started successfully', 'success');
        await refetch(); // Refresh to show updated status
      } else {
        // Check for specific error messages
        if (response.error && response.error.includes('not implemented')) {
          showNotification(
            'This task type is not fully implemented in the backend yet. The API endpoint exists but the runner is not implemented.',
            'warning'
          );
        } else {
          throw new Error(response.error || 'Failed to run task');
        }
      }
    } catch (err) {
      showNotification(
        err instanceof Error ? err.message : 'Failed to run task',
        'error'
      );
    }
  }, [refetch, showNotification]);

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
    openDialog('executions');
  }, [getTaskExecutions, openDialog]);

  // Handle opening create task dialog
  const handleOpenCreateDialog = useCallback(() => {
    openDialog('create');
  }, [openDialog]);

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

  // Define page header actions
  const headerActions = (
    <>
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
        onClick={() => refetch()}
        disabled={loading}
        sx={{ ml: 2 }}
      >
        Refresh
      </Button>
    </>
  );

  return (
    <Box sx={{ p: 3 }}>
      {/* Use the PageHeader component */}
      <PageHeader
        title="System Tasks"
        description="View and manage scheduled system tasks. Tasks can be enabled, disabled, or manually triggered."
        lastUpdated={lastUpdated}
        actions={headerActions}
        loading={loading}
      />

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
                                  openDialog('delete');
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
        open={isDialogOpen('create')} 
        onClose={() => !createTaskLoading && closeDialog('create')}
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
              onClick={() => closeDialog('create')} 
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

      {/* Delete Confirmation Dialog - Use ConfirmDialog component */}
      <ConfirmDialog
        open={isDialogOpen('delete')}
        onClose={() => closeDialog('delete')}
        onConfirm={() => selectedTaskId && deleteTask(selectedTaskId)}
        title="Delete Task"
        message="Are you sure you want to delete this task? This action cannot be undone."
        confirmText="Delete"
        cancelText="Cancel"
        severity="error"
      />

      {/* Task Executions Dialog */}
      <Dialog
        open={isDialogOpen('executions')}
        onClose={() => closeDialog('executions')}
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
          <Button onClick={() => closeDialog('executions')}>Close</Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default Tasks; 