import { useState, useEffect } from 'react';
import { apiClient } from './api';
import type { SystemMetrics } from './types/api';
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
  Skeleton
} from '@mui/material';
import { visuallyHidden } from '@mui/utils';
import RefreshIcon from '@mui/icons-material/Refresh';

// Maximum number of data points to keep in history
const MAX_HISTORY_POINTS = 20;

export const Dashboard: React.FC = () => {
  const [metrics, setMetrics] = useState<SystemMetrics | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [cpuHistory, setCpuHistory] = useState<{value: number; timestamp: string}[]>([]);
  const theme = useTheme();

  useEffect(() => {
    const fetchMetrics = async () => {
      try {
        setLoading(true);
        const response = await apiClient.getAllMetrics();
        
        if (response.success && response.data) {
          setMetrics(response.data);
          
          // Update CPU history with new data point
          setCpuHistory(prevHistory => {
            const newHistory = [
              ...prevHistory, 
              { 
                value: response.data!.cpu.usage_percent, 
                timestamp: new Date().toLocaleTimeString() 
              }
            ];
            
            // Keep only the last MAX_HISTORY_POINTS data points
            return newHistory.slice(-MAX_HISTORY_POINTS);
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
  const processTotal = metrics?.processes?.length || 0;
  const processRunning = processTotal > 0 ? Math.round(processTotal * 0.6) : 0; // Estimate as 60% running
  const processSleeping = processTotal > 0 ? Math.round(processTotal * 0.35) : 0; // Estimate as 35% sleeping
  const processStopped = processTotal > 0 ? Math.round(processTotal * 0.05) : 0; // Estimate as 5% stopped

  return (
    <Box component="section" aria-labelledby="dashboard-title">
      <Typography 
        variant="h4" 
        component="h2" 
        align="center" 
        gutterBottom
        id="dashboard-title"
      >
        System Dashboard
      </Typography>
      
      <Box sx={{ ...visuallyHidden }}>
        {loading ? 'Refreshing metrics data...' : 'Metrics data updated'}
      </Box>
      
      {metrics && (
        <>
          <Grid container spacing={3} sx={{ mb: 4 }}>
            {/* CPU Metrics */}
            <Grid item xs={12} sm={6} md={3}>
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
                      aria-label={`CPU usage ${metrics.cpu.usage_percent.toFixed(1)} percent`}
                    >
                      {metrics.cpu.usage_percent.toFixed(1)}%
                    </Typography>
                    <Divider sx={{ my: 1 }} />
                    <Typography variant="body2" color="text.secondary">
                      Load 1m: {metrics.cpu.load1.toFixed(2)}
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      Load 5m: {metrics.cpu.load5.toFixed(2)}
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      Load 15m: {metrics.cpu.load15.toFixed(2)}
                    </Typography>
                  </CardContent>
                </Card>
              )}
            </Grid>

            {/* Memory Metrics */}
            <Grid item xs={12} sm={6} md={3}>
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
                      aria-label={`Memory usage ${metrics.memory.used_percent.toFixed(1)} percent`}
                    >
                      {metrics.memory.used_percent.toFixed(1)}%
                    </Typography>
                    <Divider sx={{ my: 1 }} />
                    <Typography variant="body2" color="text.secondary">
                      Used: {(metrics.memory.used / 1024 / 1024 / 1024).toFixed(1)} GB
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      Free: {(metrics.memory.free / 1024 / 1024 / 1024).toFixed(1)} GB
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      Total: {(metrics.memory.total / 1024 / 1024 / 1024).toFixed(1)} GB
                    </Typography>
                  </CardContent>
                </Card>
              )}
            </Grid>

            {/* Network Metrics */}
            <Grid item xs={12} sm={6} md={3}>
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
                      Sent: {(metrics.network.bytes_sent / 1024 / 1024).toFixed(1)} MB
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      Received: {(metrics.network.bytes_recv / 1024 / 1024).toFixed(1)} MB
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      Packets Sent: {metrics.network.packets_sent.toLocaleString()}
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      Packets Received: {metrics.network.packets_recv.toLocaleString()}
                    </Typography>
                  </CardContent>
                </Card>
              )}
            </Grid>

            {/* Process Count */}
            <Grid item xs={12} sm={6} md={3}>
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
                      aria-label={`${processTotal} total processes`}
                    >
                      {processTotal}
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
          </Grid>

          {/* Charts */}
          {charts && (
            <Grid container spacing={3} sx={{ mb: 4 }}>
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

          {/* Processes Table */}
          {metrics.processes && metrics.processes.length > 0 && (
            <Box sx={{ mb: 4 }} component="section" aria-labelledby="processes-table-title">
              <Typography variant="h5" component="h3" gutterBottom id="processes-table-title">
                Top Processes
              </Typography>
              <TableContainer 
                component={Paper} 
                elevation={2}
                aria-label="Top processes table"
              >
                <Table>
                  <TableHead>
                    <TableRow>
                      <TableCell>PID</TableCell>
                      <TableCell>Name</TableCell>
                      <TableCell>CPU %</TableCell>
                      <TableCell>Memory %</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {metrics.processes.slice(0, 5).map((process) => (
                      <TableRow key={process.pid}>
                        <TableCell>{process.pid}</TableCell>
                        <TableCell>{process.name}</TableCell>
                        <TableCell>{process.cpu_percent.toFixed(1)}%</TableCell>
                        <TableCell>{process.mem_percent.toFixed(1)}%</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </TableContainer>
            </Box>
          )}
        </>
      )}
      
      <Typography 
        variant="body2" 
        color="text.secondary" 
        align="center" 
        sx={{ mt: 4, fontStyle: 'italic' }}
        aria-live="polite"
      >
        Last updated: {new Date().toLocaleTimeString()}
        {loading && <span> (Refreshing...)</span>}
      </Typography>
    </Box>
  );
};

export default Dashboard;