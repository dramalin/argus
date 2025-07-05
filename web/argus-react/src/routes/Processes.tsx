import React from 'react';
import { Box, Typography, Paper, Card, CardContent } from '@mui/material';
import useProcesses from '../hooks/useProcesses';
import ProcessTable from '../components/ProcessTable';
import LoadingErrorHandler from '../components/LoadingErrorHandler';

/**
 * Processes page component
 * Displays system processes information using ProcessTable component
 * Uses lazy loading for better performance
 */
const Processes: React.FC = () => {
  // Use the processes hook from context
  const { 
    processes, 
    total: processTotal, 
    lastUpdated, 
    loading: processLoading, 
    error: processError,
    params: processParams,
    handleParamChange,
    getResetFilters
  } = useProcesses();

  const handleResetFilters = () => {
    getResetFilters();
  };

  return (
    <Box sx={{ p: 3 }}>
      <Paper sx={{ p: 3, mb: 4 }}>
        <Typography variant="h4" gutterBottom>
          System Processes
        </Typography>
        <Typography variant="body1" paragraph>
          View and monitor all running processes on the system. Use the filters to find specific processes or sort by resource usage.
        </Typography>
      </Paper>

      <LoadingErrorHandler loading={processLoading && !processes.length} error={processError}>
        {/* Process Monitor */}
        <Card>
          <CardContent sx={{ pb: 3 }}>
            <Typography variant="h6" gutterBottom sx={{ mb: 2 }}>
              Process Monitor
            </Typography>
            <ProcessTable
              processes={processes}
              processParams={processParams}
              processTotal={processTotal}
              processLoading={processLoading}
              processError={processError}
              lastUpdated={lastUpdated}
              onParamChange={handleParamChange}
              onResetFilters={handleResetFilters}
            />
          </CardContent>
        </Card>
      </LoadingErrorHandler>
    </Box>
  );
};

export default Processes; 