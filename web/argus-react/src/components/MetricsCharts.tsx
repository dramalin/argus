import React, { useMemo, useCallback } from 'react';
import { Grid, useTheme } from '@mui/material';
import ChartWidget from './ChartWidget';
import type { SystemMetrics } from '../types/api';

interface MetricsChartsProps {
  metrics: SystemMetrics;
  cpuHistory: {value: number; timestamp: string}[];
}

/**
 * MetricsCharts component
 * Displays various charts for system metrics
 * Optimized with useMemo and useCallback for better performance
 */
const MetricsCharts: React.FC<MetricsChartsProps> = ({ metrics, cpuHistory }) => {
  const theme = useTheme();

  // Prepare chart data with useMemo to avoid recalculation on every render
  const charts = useMemo(() => {
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
  }, [cpuHistory, metrics.cpu.load1, metrics.cpu.load15, metrics.cpu.load5, metrics.memory.free, metrics.memory.used, metrics.network.bytes_recv, metrics.network.bytes_sent]);

  // Format tooltip labels
  const formatTooltipLabel = useCallback((context: any) => {
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
  }, []);

  // Format x-axis ticks
  const formatXAxisTick = useCallback((value: any, index: number, values: any[]) => {
    // Show fewer x-axis labels for better readability
    return index % Math.ceil(values.length / 6) === 0 ? value : '';
  }, []);

  // Format y-axis ticks
  const formatYAxisTick = useCallback((value: any) => {
    return value + '%';
  }, []);

  // Common chart options
  const chartOptions = useMemo(() => ({
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
          label: formatTooltipLabel
        }
      }
    }
  }), [theme.palette.text.primary, formatTooltipLabel]);

  // Line chart specific options
  const lineChartOptions = useMemo(() => ({
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
          callback: formatXAxisTick
        },
        offset: false, // Fix for right shift issue
        alignToPixels: true, // Improve alignment
      },
      y: {
        beginAtZero: true,
        max: 100, // CPU percentage is 0-100
        grid: {
          color: theme.palette.divider,
        },
        ticks: {
          color: theme.palette.text.secondary,
          callback: formatYAxisTick
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
    layout: {
      padding: {
        left: 0, // Reduce left padding
        right: 0, // Reduce right padding
      }
    },
  }), [chartOptions, theme.palette.divider, theme.palette.text.secondary, formatXAxisTick, formatYAxisTick]);

  return (
    <Grid container spacing={3} sx={{ px: 3, mt: 0 }}>
      {/* CPU Usage History Chart */}
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
      
      {/* System Load Chart */}
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

      {/* Memory Distribution Chart */}
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

      {/* Network Traffic Chart */}
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
  );
};

export default React.memo(MetricsCharts); 