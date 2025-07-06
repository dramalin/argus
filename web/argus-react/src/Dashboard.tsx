import { useEffect, lazy, Suspense } from 'react';
import { Grid, Card, CardContent, Typography } from '@mui/material';
import LoadingErrorHandler from './components/LoadingErrorHandler';
import LoadingFallback from './components/LoadingFallback';
import useMetrics from './hooks/useMetrics';
import useProcesses from './hooks/useProcesses';

// Lazy-loaded components
const SystemOverview = lazy(() => import('./components/SystemOverview'));
const MetricsCharts = lazy(() => import('./components/MetricsCharts'));
const ProcessTable = lazy(() => import('./components/ProcessTable'));

/**
 * Dashboard component
 * Main dashboard view that displays system metrics and process information
 * Uses lazy loading for heavy components
 */
export const Dashboard: React.FC = () => {
  // Use the metrics hook from context
  const { 
    metrics, 
    cpuHistory, 
    loading: metricsLoading, 
    error: metricsError 
  } = useMetrics();

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

  // Announce loading state for screen readers
  useEffect(() => {
    if (metricsLoading) {
      document.title = 'Loading metrics... - Argus Monitor';
    } else {
      document.title = 'System Dashboard - Argus Monitor';
    }
  }, [metricsLoading]);

  const handleResetFilters = () => {
    getResetFilters();
  };

  return (
    <LoadingErrorHandler loading={metricsLoading && !metrics} error={metricsError}>
      {metrics && (
        <>
          <Suspense fallback={<LoadingFallback message="Loading system overview..." />}>
            <SystemOverview metrics={metrics} loading={metricsLoading} processTotal={processTotal} />
          </Suspense>
          
          <Suspense fallback={<LoadingFallback message="Loading metrics charts..." />}>
            <MetricsCharts metrics={metrics} cpuHistory={cpuHistory} />
          </Suspense>
          
          {/* Process Monitor */}
          <Grid item xs={12} sx={{ px: 3, pb: 4, mt: 2 }}>
            <Card>
              <CardContent sx={{ pb: 3 }}>
                <Typography variant="h6" gutterBottom sx={{ mb: 2 }}>
                  Process Monitor
                </Typography>
                <Suspense fallback={<LoadingFallback message="Loading process table..." height="500px" />}>
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
                </Suspense>
              </CardContent>
            </Card>
          </Grid>
        </>
      )}
    </LoadingErrorHandler>
  );
};

export default Dashboard;