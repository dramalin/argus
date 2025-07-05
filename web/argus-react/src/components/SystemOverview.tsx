import React, { useMemo } from 'react';
import { Grid } from '@mui/material';
import SystemMetricsCard from './SystemMetricsCard';
import type { SystemMetrics } from '../types/api';

interface SystemOverviewProps {
  metrics: SystemMetrics | null;
  loading: boolean;
}

/**
 * SystemOverview component
 * Displays system metrics cards for CPU, memory, network, and processes
 * Optimized with useMemo for better performance
 */
const SystemOverview: React.FC<SystemOverviewProps> = ({ metrics, loading }) => {
  // Memoize process counts to avoid recalculation on every render
  const processCounts = useMemo(() => {
    const processTotalSummary = metrics?.processes?.length || 0;
    const processRunning = processTotalSummary > 0 ? Math.round(processTotalSummary * 0.6) : 0; // Estimate as 60% running
    const processSleeping = processTotalSummary > 0 ? Math.round(processTotalSummary * 0.35) : 0; // Estimate as 35% sleeping
    const processStopped = processTotalSummary > 0 ? Math.round(processTotalSummary * 0.05) : 0; // Estimate as 5% stopped
    
    return {
      total: processTotalSummary,
      running: processRunning,
      sleeping: processSleeping,
      stopped: processStopped
    };
  }, [metrics?.processes?.length]);

  // Memoize CPU details to avoid recalculation on every render
  const cpuDetails = useMemo(() => (
    metrics ? [
      { label: 'Load 1m', value: metrics.cpu.load1.toFixed(2) },
      { label: 'Load 5m', value: metrics.cpu.load5.toFixed(2) },
      { label: 'Load 15m', value: metrics.cpu.load15.toFixed(2) }
    ] : []
  ), [metrics?.cpu.load1, metrics?.cpu.load5, metrics?.cpu.load15]);

  // Memoize memory details to avoid recalculation on every render
  const memoryDetails = useMemo(() => (
    metrics ? [
      { label: 'Used', value: `${(metrics.memory.used / 1024 / 1024 / 1024).toFixed(1)} GB` },
      { label: 'Free', value: `${(metrics.memory.free / 1024 / 1024 / 1024).toFixed(1)} GB` },
      { label: 'Total', value: `${(metrics.memory.total / 1024 / 1024 / 1024).toFixed(1)} GB` }
    ] : []
  ), [metrics?.memory.used, metrics?.memory.free, metrics?.memory.total]);

  // Memoize network details to avoid recalculation on every render
  const networkDetails = useMemo(() => (
    metrics ? [
      { label: 'Sent', value: `${(metrics.network.bytes_sent / 1024 / 1024).toFixed(1)} MB` },
      { label: 'Received', value: `${(metrics.network.bytes_recv / 1024 / 1024).toFixed(1)} MB` },
      { label: 'Packets Sent', value: metrics.network.packets_sent.toLocaleString() },
      { label: 'Packets Received', value: metrics.network.packets_recv.toLocaleString() }
    ] : []
  ), [
    metrics?.network.bytes_sent,
    metrics?.network.bytes_recv,
    metrics?.network.packets_sent,
    metrics?.network.packets_recv
  ]);

  // Memoize process details to avoid recalculation on every render
  const processDetails = useMemo(() => [
    { label: 'Running', value: processCounts.running },
    { label: 'Sleeping', value: processCounts.sleeping },
    { label: 'Stopped', value: processCounts.stopped }
  ], [processCounts.running, processCounts.sleeping, processCounts.stopped]);

  return (
    <Grid container spacing={3}>
      {/* CPU Metrics */}
      <Grid item xs={12} sm={6} md={3} sx={{ pl: 3, pt: 3 }}>
        <SystemMetricsCard
          title="CPU Usage"
          value={metrics?.cpu.usage_percent.toFixed(1)}
          unit="%"
          loading={loading}
          titleId="cpu-title"
          details={cpuDetails}
        />
      </Grid>

      {/* Memory Metrics */}
      <Grid item xs={12} sm={6} md={3} sx={{ pt: 3 }}>
        <SystemMetricsCard
          title="Memory Usage"
          value={metrics?.memory.used_percent.toFixed(1)}
          unit="%"
          loading={loading}
          titleId="memory-title"
          details={memoryDetails}
        />
      </Grid>

      {/* Network Metrics */}
      <Grid item xs={12} sm={6} md={3} sx={{ pt: 3 }}>
        <SystemMetricsCard
          title="Network Traffic"
          loading={loading}
          titleId="network-title"
          details={networkDetails}
        />
      </Grid>

      {/* FIXME: processCount is not correct, always 50 only, need to fix */}
      {/* Process Count */}
      <Grid item xs={12} sm={6} md={3} sx={{ pr: 3, pt: 3 }}>
        <SystemMetricsCard
          title="Processes"
          value={processCounts.total}
          loading={loading}
          titleId="processes-title"
          details={processDetails}
        />
      </Grid>
    </Grid>
  );
};

export default React.memo(SystemOverview); 