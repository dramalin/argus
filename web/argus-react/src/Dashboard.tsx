import { useState, useEffect } from 'react';
import { apiClient } from './api';
import type { SystemMetrics } from './types/api';

export const Dashboard: React.FC = () => {
  const [metrics, setMetrics] = useState<SystemMetrics | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchMetrics = async () => {
      try {
        setLoading(true);
        const response = await apiClient.getAllMetrics();
        
        if (response.success && response.data) {
          setMetrics(response.data);
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

  if (loading && !metrics) {
    return (
      <div className="dashboard">
        <div className="loading">Loading system metrics...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="dashboard">
        <div className="error">
          <h3>Error loading metrics</h3>
          <p>{error}</p>
          <button onClick={() => window.location.reload()}>Retry</button>
        </div>
      </div>
    );
  }

  return (
    <div className="dashboard">
      <h2>System Dashboard</h2>
      
      {metrics && (
        <div className="metrics-grid">
          {/* CPU Metrics */}
          <div className="metric-card">
            <h3>CPU Usage</h3>
            <div className="metric-value">
              {metrics.cpu.usage_percent.toFixed(1)}%
            </div>
            <div className="metric-details">
              <div>Load 1m: {metrics.cpu.load1.toFixed(2)}</div>
              <div>Load 5m: {metrics.cpu.load5.toFixed(2)}</div>
              <div>Load 15m: {metrics.cpu.load15.toFixed(2)}</div>
            </div>
          </div>

          {/* Memory Metrics */}
          <div className="metric-card">
            <h3>Memory Usage</h3>
            <div className="metric-value">
              {metrics.memory.used_percent.toFixed(1)}%
            </div>
            <div className="metric-details">
              <div>Used: {(metrics.memory.used / 1024 / 1024 / 1024).toFixed(1)} GB</div>
              <div>Free: {(metrics.memory.free / 1024 / 1024 / 1024).toFixed(1)} GB</div>
              <div>Total: {(metrics.memory.total / 1024 / 1024 / 1024).toFixed(1)} GB</div>
            </div>
          </div>

          {/* Network Metrics */}
          <div className="metric-card">
            <h3>Network Traffic</h3>
            <div className="metric-details">
              <div>Sent: {(metrics.network.bytes_sent / 1024 / 1024).toFixed(1)} MB</div>
              <div>Received: {(metrics.network.bytes_recv / 1024 / 1024).toFixed(1)} MB</div>
              <div>Packets Sent: {metrics.network.packets_sent.toLocaleString()}</div>
              <div>Packets Received: {metrics.network.packets_recv.toLocaleString()}</div>
            </div>
          </div>

          {/* Process Count */}
          <div className="metric-card">
            <h3>Processes</h3>
            <div className="metric-value">
              {metrics.processes.length}
            </div>
            <div className="metric-details">
              <div>Total running processes</div>
            </div>
          </div>
        </div>
      )}

      {/* Top Processes */}
      {metrics && metrics.processes.length > 0 && (
        <div className="processes-section">
          <h3>Top Processes by CPU</h3>
          <div className="processes-table">
            <div className="table-header">
              <div>PID</div>
              <div>Name</div>
              <div>CPU %</div>
              <div>Memory %</div>
            </div>
            {metrics.processes
              .sort((a, b) => b.cpu_percent - a.cpu_percent)
              .slice(0, 10)
              .map((process) => (
                <div key={process.pid} className="table-row">
                  <div>{process.pid}</div>
                  <div>{process.name}</div>
                  <div>{process.cpu_percent.toFixed(1)}%</div>
                  <div>{process.mem_percent.toFixed(1)}%</div>
                </div>
              ))}
          </div>
        </div>
      )}

      <div className="dashboard-footer">
        <p>Last updated: {metrics?.timestamp ? new Date(metrics.timestamp).toLocaleString() : 'Never'}</p>
      </div>
    </div>
  );
};

export default Dashboard; 