import { useState, useEffect, useRef } from 'react';
import { apiClient } from './api';
import type { SystemMetrics } from './types/api';
import type { ProcessInfo, ProcessQueryParams } from './types/process';
import ChartWidget from './components/ChartWidget';
import { 
  Grid, 
  Card, 
  CardContent, 
  Typography, 
  CircularProgress, 
  Alert, 
  Button, 
  Box,
  Divider,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  useTheme,
  Skeleton,
  TextField,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Pagination,
  Stack,
  Chip
} from '@mui/material';
import RefreshIcon from '@mui/icons-material/Refresh';
import AccessTimeIcon from '@mui/icons-material/AccessTime';

// Maximum number of data points to keep in history
const MAX_HISTORY_POINTS = 20;

export const Dashboard: React.FC = () => {
  const [metrics, setMetrics] = useState<SystemMetrics | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [cpuHistory, setCpuHistory] = useState<{value: number; timestamp: string}[]>([]);
  const theme = useTheme();

  // Process table state
  const [processParams, setProcessParams] = useState<ProcessQueryParams>({
    limit: 10,
    offset: 0,
    sort_by: 'cpu',
    sort_order: 'desc',
    name_contains: '',
    min_cpu: undefined,
    min_memory: undefined,
  });
  const [processes, setProcesses] = useState<ProcessInfo[]>([]);
  const [processLoading, setProcessLoading] = useState(false);
  const [processError, setProcessError] = useState<string | null>(null);
  const [processTotal, setProcessTotal] = useState(0);

  const [lastUpdated, setLastUpdated] = useState<string | null>(null);

  // For pagination
  const page = Math.floor((processParams.offset || 0) / (processParams.limit || 10)) + 1;
  const pageSize = processParams.limit || 10;
  const totalPages = Math.ceil(processTotal / pageSize);

  const isInitialLoad = useRef(true);

  // Check if query params have changed
  const queryParamsRef = useRef<string | null>(null);

  // Check if metrics have changed
  const metricsRef = useRef<string | null>(null);

  // Function to fetch processes
  const fetchProcesses = async (showLoading = false) => {
    if (showLoading) setProcessLoading(true);
    setProcessError(null);
    
    try {
      const resp = await apiClient.getProcesses(processParams);
      
      if (resp.success && resp.data) {
        const { processes, total_count, updated_at } = resp.data;
        setProcesses(processes);
        setProcessTotal(total_count);
        setLastUpdated(updated_at);
      } else {
        setProcessError(resp.error || 'Failed to fetch processes');
        setProcesses([]);
        setProcessTotal(0);
        setLastUpdated(null);
      }
    } catch (err) {
      setProcessError(err instanceof Error ? err.message : 'Unknown error');
      setProcesses([]);
      setProcessTotal(0);
      setLastUpdated(null);
    } finally {
      if (showLoading) setProcessLoading(false);
    }
  };

  // Fetch metrics and poll every 5 seconds
  useEffect(() => {
    const fetchMetrics = async () => {
      try {
        const response = await apiClient.getAllMetrics();
        if (response.success && response.data) {
          setMetrics(response.data);
          // Update CPU history
          setCpuHistory(prev => {
            const newPoint = {
              timestamp: new Date().toLocaleTimeString(),
              value: response.data!.cpu.usage_percent
            };
            return [...prev.slice(-MAX_HISTORY_POINTS), newPoint];
          });
          setError(null);
        } else {
          setError(response.error || 'Failed to fetch metrics');
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Unknown error');
      } finally {
        setLoading(false);
      }
    };

    fetchMetrics();
    
    // Set up polling every 5 seconds
    const interval = setInterval(fetchMetrics, 5000);
    
    return () => clearInterval(interval);
  }, []);

  // Fetch processes with params and poll every 5 seconds
  useEffect(() => {
    let ignore = false;

    // Initial fetch with loading spinner
    fetchProcesses(true);
    isInitialLoad.current = false;

    // Polling without spinner
    const interval = setInterval(() => {
      if (!ignore) fetchProcesses(false);
    }, 5000);

    return () => {
      ignore = true;
      clearInterval(interval);
    };
  }, [processParams]);

  // Check if query params have changed
  useEffect(() => {
    const queryParamsString = JSON.stringify(processParams);
    if (queryParamsString !== queryParamsRef.current) {
      fetchProcesses();
      queryParamsRef.current = queryParamsString;
    }
  }, [processParams]);

  // Check if metrics have changed
  useEffect(() => {
    const metricsString = JSON.stringify(metrics);
    if (metricsString !== metricsRef.current) {
      metricsRef.current = metricsString;
    }
  }, [metrics]);

  // Announce loading state for screen readers
  useEffect(() => {
    if (loading) {
      // This would be better with a live region, but for simplicity we'll use document.title
      document.title = 'Loading metrics... - Argus Monitor';
    } else {
      document.title = 'System Dashboard - Argus Monitor';
    }
  }, [loading]);

  if (loading && !metrics) {
    return (
      <Box 
        sx={{ 
          display: 'flex', 
          justifyContent: 'center', 
          alignItems: 'center', 
          minHeight: '50vh',
          flexDirection: 'column'
        }}
        role="status"
        aria-live="polite"
        aria-busy="true"
      >
        <CircularProgress color="primary" aria-hidden="true" />
        <Typography variant="h6" sx={{ ml: 2, mt: 2 }}>
          Loading system metrics...
        </Typography>
      </Box>
    );
  }

  if (error) {
    return (
      <Box 
        sx={{ maxWidth: 600, mx: 'auto', mt: 4 }}
        role="alert"
        aria-live="assertive"
      >
        <Alert 
          severity="error" 
          action={
            <Button 
              color="inherit" 
              size="small" 
              startIcon={<RefreshIcon />}
              onClick={() => window.location.reload()}
              aria-label="Retry loading metrics"
            >
              Retry
            </Button>
          }
        >
          <Typography variant="h6">Error loading metrics</Typography>
          <Typography>{error}</Typography>
        </Alert>
      </Box>
    );
  }

  // Prepare chart data
  const prepareChartData = (metrics: SystemMetrics) => {
    // CPU usage history line chart
    const cpuHistoryData = {
      labels: cpuHistory.map(point => point.timestamp),
      datasets: [
        {
          label: 'CPU Usage (%)',
          data: cpuHistory.map(point => point.value),
          borderColor: 'rgba(106, 123, 162, 1)',        // Morandi blue
          backgroundColor: 'rgba(106, 123, 162, 0.2)',  // Morandi blue with transparency
          borderWidth: 2,
          tension: 0.4,                                 // Adds curve to the line
          fill: true,                                   // Fill area under the line
          pointRadius: 2,                               // Small points
          pointHoverRadius: 5,                          // Larger points on hover
        },
      ],
    };

    // Memory data for doughnut chart
    const memoryData = {
      labels: ['Used', 'Free'],
      datasets: [
        {
          data: [metrics.memory.used, metrics.memory.free],
          backgroundColor: [
            'rgba(185, 122, 122, 0.7)',  // Morandi rose (error color)
            'rgba(122, 162, 158, 0.7)',  // Morandi teal (success color)
          ],
          borderColor: [
            'rgba(185, 122, 122, 1)',
            'rgba(122, 162, 158, 1)',
          ],
          borderWidth: 1,
        },
      ],
    };

    // CPU load history (simulated with current values)
    const cpuLoadData = {
      labels: ['1 min', '5 min', '15 min'],
      datasets: [
        {
          label: 'System Load',
          data: [metrics.cpu.load1, metrics.cpu.load5, metrics.cpu.load15],
          backgroundColor: 'rgba(122, 158, 185, 0.6)',  // Morandi info blue
          borderColor: 'rgba(122, 158, 185, 1)',        // Morandi info blue
          borderWidth: 1,
        },
      ],
    };

    // Network traffic
    const networkData = {
      labels: ['Sent', 'Received'],
      datasets: [
        {
          label: 'Network Traffic (MB)',
          data: [
            metrics.network.bytes_sent / 1024 / 1024,
            metrics.network.bytes_recv / 1024 / 1024
          ],
          backgroundColor: [
            'rgba(185, 169, 122, 0.6)',  // Morandi warning gold
            'rgba(122, 162, 158, 0.6)',  // Morandi teal (success color)
          ],
          borderColor: [
            'rgba(185, 169, 122, 1)',
            'rgba(122, 162, 158, 1)',
          ],
          borderWidth: 1,
        },
      ],
    };

    return { cpuHistoryData, memoryData, cpuLoadData, networkData };
  };

  // Common chart options
  const chartOptions = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: {
        position: 'bottom' as const,
        labels: {
          color: theme.palette.text.primary, // Ensure good contrast
          font: {
            size: 14, // Larger font for better readability
          },
          padding: 15, // Larger touch targets
        }
      },
      title: {
        display: false, // We'll use MUI Typography instead
      },
      tooltip: {
        enabled: true,
        mode: 'nearest' as const, // Type assertion to make it compatible
        intersect: false, // More accessible tooltips
        callbacks: {
          // Format numbers for better readability
          label: function(context: any) {
            let label = context.dataset.label || '';
            if (label) {
              label += ': ';
            }
            if (context.parsed !== undefined) {
              label += typeof context.parsed === 'object' 
                ? (context.parsed.y !== undefined ? context.parsed.y.toFixed(1) : context.parsed)
                : context.parsed.toFixed(1);
            }
            return label;
          }
        }
      }
    }
  };

  // Line chart specific options
  const lineChartOptions = {
    ...chartOptions,
    scales: {
      x: {
        grid: {
          color: theme.palette.divider,
        },
        ticks: {
          color: theme.palette.text.secondary,
          maxRotation: 0,
          autoSkipPadding: 15,
          callback: function(value: any, index: number, values: any[]) {
            // Show fewer x-axis labels for better readability
            return index % Math.ceil(values.length / 6) === 0 ? value : '';
          }
        },
      },
      y: {
        beginAtZero: true,
        max: 100, // CPU percentage is 0-100
        grid: {
          color: theme.palette.divider,
        },
        ticks: {
          color: theme.palette.text.secondary,
          callback: function(value: any) {
            return value + '%';
          }
        },
      }
    },
    interaction: {
      mode: 'index' as const,
      intersect: false,
    },
    elements: {
      line: {
        tension: 0.4 // Smoother curve
      },
      point: {
        radius: 2,
        hoverRadius: 5,
      }
    },
  };

  // Prepare charts data if metrics are available
  const charts = metrics ? prepareChartData(metrics) : null;

  // Loading skeleton for metrics cards
  const MetricCardSkeleton = () => (
    <Card elevation={2}>
      <CardContent>
        <Skeleton variant="text" width="60%" height={30} />
        <Skeleton variant="text" width="40%" height={40} sx={{ my: 1 }} />
        <Divider sx={{ my: 1 }} />
        <Skeleton variant="text" width="80%" />
        <Skeleton variant="text" width="70%" />
        <Skeleton variant="text" width="75%" />
      </CardContent>
    </Card>
  );

  // Get process counts
  const processTotalSummary = metrics?.processes?.length || 0;
  const processRunning = processTotalSummary > 0 ? Math.round(processTotalSummary * 0.6) : 0; // Estimate as 60% running
  const processSleeping = processTotalSummary > 0 ? Math.round(processTotalSummary * 0.35) : 0; // Estimate as 35% sleeping
  const processStopped = processTotalSummary > 0 ? Math.round(processTotalSummary * 0.05) : 0; // Estimate as 5% stopped

  // Handlers for controls
  const handleParamChange = (key: keyof ProcessQueryParams, value: any) => {
    setProcessParams(prev => ({
      ...prev,
      [key]: value,
      // Reset offset when changing filters
      offset: key !== 'offset' ? 0 : value
    }));
  };

  const handlePageChange = (newPage: number) => {
    handleParamChange('offset', (newPage - 1) * pageSize);
  };

  const handleResetFilters = () => {
    setProcessParams({
      limit: 10,
      offset: 0,
      sort_by: 'cpu',
      sort_order: 'desc',
      name_contains: '',
      min_cpu: undefined,
      min_memory: undefined,
    });
  };

  // Process table section
  const renderProcessTable = () => {
    if (processLoading && processes.length === 0) {
      return (
        <Box sx={{ p: 3 }}>
          <CircularProgress />
        </Box>
      );
    }

    if (processError) {
      return (
        <Alert severity="error" sx={{ m: 2 }}>
          {processError}
        </Alert>
      );
    }

    return (
      <>
        <Box sx={{ p: 2, display: 'flex', gap: 2, flexWrap: 'wrap' }}>
          <TextField
            label="Filter by name"
            size="small"
            value={processParams.name_contains || ''}
            onChange={(e) => handleParamChange('name_contains', e.target.value)}
          />
          <TextField
            label="Min CPU %"
            type="number"
            size="small"
            value={processParams.min_cpu || ''}
            onChange={(e) => handleParamChange('min_cpu', e.target.value ? Number(e.target.value) : undefined)}
          />
          <TextField
            label="Min Memory %"
            type="number"
            size="small"
            value={processParams.min_memory || ''}
            onChange={(e) => handleParamChange('min_memory', e.target.value ? Number(e.target.value) : undefined)}
          />
          <FormControl size="small" sx={{ minWidth: 120 }}>
            <InputLabel>Sort by</InputLabel>
            <Select
              value={processParams.sort_by}
              label="Sort by"
              onChange={(e) => handleParamChange('sort_by', e.target.value)}
            >
              <MenuItem value="cpu">CPU Usage</MenuItem>
              <MenuItem value="memory">Memory Usage</MenuItem>
              <MenuItem value="name">Name</MenuItem>
              <MenuItem value="pid">PID</MenuItem>
            </Select>
          </FormControl>
          <FormControl size="small" sx={{ minWidth: 120 }}>
            <InputLabel>Order</InputLabel>
            <Select
              value={processParams.sort_order}
              label="Order"
              onChange={(e) => handleParamChange('sort_order', e.target.value as 'asc' | 'desc')}
            >
              <MenuItem value="asc">Ascending</MenuItem>
              <MenuItem value="desc">Descending</MenuItem>
            </Select>
          </FormControl>
          <Button 
            variant="outlined" 
            onClick={handleResetFilters}
            size="small"
          >
            Reset Filters
          </Button>
        </Box>

        <TableContainer component={Paper} sx={{ mx: 2, mb: 2 }}>
          <Table size="small" aria-label="process list">
            <TableHead>
              <TableRow>
                <TableCell>PID</TableCell>
                <TableCell>Name</TableCell>
                <TableCell align="right">CPU %</TableCell>
                <TableCell align="right">Memory %</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {processes.map((process) => (
                <TableRow key={process.pid}>
                  <TableCell>{process.pid}</TableCell>
                  <TableCell>{process.name}</TableCell>
                  <TableCell align="right">{process.cpu_percent.toFixed(1)}</TableCell>
                  <TableCell align="right">{process.mem_percent.toFixed(1)}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>

        <Box sx={{ p: 2, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <Stack direction="row" spacing={1} alignItems="center">
            <Typography variant="body2" color="text.secondary">
              {`${processTotal} total processes`}
            </Typography>
            {lastUpdated && (
              <Chip
                size="small"
                icon={<AccessTimeIcon />}
                label={`Updated: ${new Date(lastUpdated).toLocaleTimeString()}`}
                variant="outlined"
              />
            )}
          </Stack>
          <Pagination
            count={totalPages}
            page={page}
            onChange={(_, newPage) => handlePageChange(newPage)}
            color="primary"
            size="small"
          />
        </Box>
      </>
    );
  };

  return (
    <Grid container spacing={3}>
      {/* CPU Metrics */}
      <Grid item xs={12} sm={6} md={3} sx={{ pl: 3, pt: 3 }}>
        {loading && !metrics ? <MetricCardSkeleton /> : (
          <Card 
            elevation={2}
            component="section"
            aria-labelledby="cpu-title"
          >
            <CardContent>
              <Typography variant="h6" component="h3" gutterBottom id="cpu-title">
                CPU Usage
              </Typography>
              <Typography 
                variant="h4" 
                color="primary" 
                gutterBottom
                aria-label={`CPU usage ${metrics!.cpu.usage_percent.toFixed(1)} percent`}
              >
                {metrics!.cpu.usage_percent.toFixed(1)}%
              </Typography>
              <Divider sx={{ my: 1 }} />
              <Typography variant="body2" color="text.secondary">
                Load 1m: {metrics!.cpu.load1.toFixed(2)}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Load 5m: {metrics!.cpu.load5.toFixed(2)}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Load 15m: {metrics!.cpu.load15.toFixed(2)}
              </Typography>
            </CardContent>
          </Card>
        )}
      </Grid>

      {/* Memory Metrics */}
      <Grid item xs={12} sm={6} md={3} sx={{ pt: 3 }}>
        {loading && !metrics ? <MetricCardSkeleton /> : (
          <Card 
            elevation={2}
            component="section"
            aria-labelledby="memory-title"
          >
            <CardContent>
              <Typography variant="h6" component="h3" gutterBottom id="memory-title">
                Memory Usage
              </Typography>
              <Typography 
                variant="h4" 
                color="primary" 
                gutterBottom
                aria-label={`Memory usage ${metrics!.memory.used_percent.toFixed(1)} percent`}
              >
                {metrics!.memory.used_percent.toFixed(1)}%
              </Typography>
              <Divider sx={{ my: 1 }} />
              <Typography variant="body2" color="text.secondary">
                Used: {(metrics!.memory.used / 1024 / 1024 / 1024).toFixed(1)} GB
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Free: {(metrics!.memory.free / 1024 / 1024 / 1024).toFixed(1)} GB
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Total: {(metrics!.memory.total / 1024 / 1024 / 1024).toFixed(1)} GB
              </Typography>
            </CardContent>
          </Card>
        )}
      </Grid>

      {/* Network Metrics */}
      <Grid item xs={12} sm={6} md={3} sx={{ pt: 3 }}>
        {loading && !metrics ? <MetricCardSkeleton /> : (
          <Card 
            elevation={2}
            component="section"
            aria-labelledby="network-title"
          >
            <CardContent>
              <Typography variant="h6" component="h3" gutterBottom id="network-title">
                Network Traffic
              </Typography>
              <Divider sx={{ my: 1 }} />
              <Typography variant="body2" color="text.secondary">
                Sent: {(metrics!.network.bytes_sent / 1024 / 1024).toFixed(1)} MB
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Received: {(metrics!.network.bytes_recv / 1024 / 1024).toFixed(1)} MB
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Packets Sent: {metrics!.network.packets_sent.toLocaleString()}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Packets Received: {metrics!.network.packets_recv.toLocaleString()}
              </Typography>
            </CardContent>
          </Card>
        )}
      </Grid>

      {/* Process Count */}
      <Grid item xs={12} sm={6} md={3} sx={{ pr: 3, pt: 3 }}>
        {loading && !metrics ? <MetricCardSkeleton /> : (
          <Card 
            elevation={2}
            component="section"
            aria-labelledby="processes-title"
          >
            <CardContent>
              <Typography variant="h6" component="h3" gutterBottom id="processes-title">
                Processes
              </Typography>
              <Typography 
                variant="h4" 
                color="primary" 
                gutterBottom
                aria-label={`${processTotalSummary} total processes`}
              >
                {processTotalSummary}
              </Typography>
              <Divider sx={{ my: 1 }} />
              <Typography variant="body2" color="text.secondary">
                Running: {processRunning}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Sleeping: {processSleeping}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Stopped: {processStopped}
              </Typography>
            </CardContent>
          </Card>
        )}
      </Grid>

      {/* Charts */}
      {charts && (
        <Grid container spacing={3} sx={{ px: 3, mt: 0 }}>
          {/* CPU Usage History Chart - Now in 2-column layout */}
          <Grid item xs={12} md={6}>
            <ChartWidget 
              type="line" 
              data={charts.cpuHistoryData} 
              options={lineChartOptions} 
              title="CPU Usage History"
              description="Real-time CPU usage percentage over time"
              height={300}
              id="cpu-history-chart"
            />
          </Grid>
          
          <Grid item xs={12} md={6}>
            <ChartWidget 
              type="bar" 
              data={charts.cpuLoadData} 
              options={chartOptions} 
              title="System Load"
              description="Average system load over 1, 5, and 15 minute periods"
              height={300}
              id="system-load-chart"
            />
          </Grid>
          <Grid item xs={12} md={6}>
            <ChartWidget 
              type="doughnut" 
              data={charts.memoryData} 
              options={chartOptions} 
              title="Memory Distribution"
              description="Distribution of used and free memory"
              height={300}
              id="memory-chart"
            />
          </Grid>
          <Grid item xs={12} md={6}>
            <ChartWidget 
              type="bar" 
              data={charts.networkData} 
              options={chartOptions} 
              title="Network Traffic"
              description="Network traffic sent and received in megabytes"
              height={300}
              id="network-chart"
            />
          </Grid>
        </Grid>
      )}

      {/* Process Monitor */}
      <Grid item xs={12} sx={{ px: 3, pb: 3 }}>
        <Card>
          <CardContent>
            <Typography variant="h6" gutterBottom>
              Process Monitor
            </Typography>
            {renderProcessTable()}
          </CardContent>
        </Card>
      </Grid>
    </Grid>
  );
};

export default Dashboard;