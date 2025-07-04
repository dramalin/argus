import { useState, useEffect } from 'react';
import { apiClient } from './api';
import type { SystemMetrics } from './types/api';
import ChartWidget from './components/ChartWidget';

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

  // Prepare chart data
  const prepareChartData = (metrics: SystemMetrics) => {
    // CPU data for gauge-like bar chart
    const cpuData = {
      labels: ['CPU Usage'],
      datasets: [
        {
          label: 'CPU Usage (%)',
          data: [metrics.cpu.usage_percent],
          backgroundColor: 'rgba(124, 58, 237, 0.6)',
          borderColor: 'rgba(124, 58, 237, 1)',
          borderWidth: 1,
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
            'rgba(239, 68, 68, 0.7)',
            'rgba(16, 185, 129, 0.7)',
          ],
          borderColor: [
            'rgba(239, 68, 68, 1)',
            'rgba(16, 185, 129, 1)',
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
          backgroundColor: 'rgba(59, 130, 246, 0.6)',
          borderColor: 'rgba(59, 130, 246, 1)',
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
            'rgba(245, 158, 11, 0.6)',
            'rgba(16, 185, 129, 0.6)',
          ],
          borderColor: [
            'rgba(245, 158, 11, 1)',
            'rgba(16, 185, 129, 1)',
          ],
          borderWidth: 1,
        },
      ],
    };

    return { cpuData, memoryData, cpuLoadData, networkData };
  };

  // Common chart options
  const chartOptions = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: {
        position: 'bottom' as const,
      },
      title: {
        display: true,
        color: '#4b5563',
        font: {
          size: 16,
        }
      }
    }
  };

  // Prepare charts data if metrics are available
  const charts = metrics ? prepareChartData(metrics) : null;

  return (
    <div className="dashboard">
      <h2>System Dashboard</h2>
      
      {metrics && (
        <>
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

          {/* Charts Section */}
          <div className="charts-container">
            {charts && (
              <>
                <div className="chart-card">
                  <h3>CPU Usage</h3>
                  <div style={{ height: '250px' }}>
                    <ChartWidget 
                      type="bar" 
                      data={charts.cpuData}
                      options={{
                        ...chartOptions,
                        plugins: {
                          ...chartOptions.plugins,
                          title: {
                            ...chartOptions.plugins.title,
                            text: 'Current CPU Usage (%)'
                          }
                        },
                        scales: {
                          y: {
                            beginAtZero: true,
                            max: 100
                          }
                        }
                      }}
                    />
                  </div>
                </div>

                <div className="chart-card">
                  <h3>Memory Distribution</h3>
                  <div style={{ height: '250px' }}>
                    <ChartWidget 
                      type="doughnut" 
                      data={charts.memoryData}
                      options={{
                        ...chartOptions,
                        plugins: {
                          ...chartOptions.plugins,
                          title: {
                            ...chartOptions.plugins.title,
                            text: 'Memory Usage'
                          }
                        }
                      }}
                    />
                  </div>
                </div>

                <div className="chart-card">
                  <h3>System Load</h3>
                  <div style={{ height: '250px' }}>
                    <ChartWidget 
                      type="bar" 
                      data={charts.cpuLoadData}
                      options={{
                        ...chartOptions,
                        plugins: {
                          ...chartOptions.plugins,
                          title: {
                            ...chartOptions.plugins.title,
                            text: 'System Load Average'
                          }
                        }
                      }}
                    />
                  </div>
                </div>

                <div className="chart-card">
                  <h3>Network Traffic</h3>
                  <div style={{ height: '250px' }}>
                    <ChartWidget 
                      type="pie" 
                      data={charts.networkData}
                      options={{
                        ...chartOptions,
                        plugins: {
                          ...chartOptions.plugins,
                          title: {
                            ...chartOptions.plugins.title,
                            text: 'Network Traffic (MB)'
                          }
                        }
                      }}
                    />
                  </div>
                </div>
              </>
            )}
          </div>
        </>
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
        <p>
          Last updated: {metrics?.timestamp ? new Date(metrics.timestamp).toLocaleString() : 'Never'}
          {loading && <span className="loading-indicator"> (Refreshing...)</span>}
        </p>
      </div>
    </div>
  );
};

export default Dashboard;