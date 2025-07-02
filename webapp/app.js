const { useState, useEffect, useRef } = React;

// Utility functions
function formatBytes(bytes) {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

function formatNumber(num) {
  return new Intl.NumberFormat().format(num);
}

// Chart component for CPU usage
function CPUChart({ data, history }) {
  const chartRef = useRef(null);
  const chartInstance = useRef(null);

  useEffect(() => {
    if (!data || !chartRef.current) return;

    const ctx = chartRef.current.getContext('2d');

    if (chartInstance.current) {
      chartInstance.current.destroy();
    }

    chartInstance.current = new Chart(ctx, {
      type: 'line',
      data: {
        labels: history.map((_, i) => `${i * 5}s`),
        datasets: [{
          label: 'CPU Usage %',
          data: history,
          borderColor: '#667eea',
          backgroundColor: 'rgba(102, 126, 234, 0.1)',
          borderWidth: 2,
          fill: true,
          tension: 0.4
        }]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        scales: {
          y: {
            beginAtZero: true,
            max: 100,
            ticks: {
              callback: function (value) {
                return value + '%';
              }
            }
          }
        },
        plugins: {
          legend: {
            display: false
          }
        }
      }
    });

    return () => {
      if (chartInstance.current) {
        chartInstance.current.destroy();
      }
    };
  }, [data, history]);

  return React.createElement('canvas', { ref: chartRef });
}

// Chart component for Memory usage
function MemoryChart({ data, history }) {
  const chartRef = useRef(null);
  const chartInstance = useRef(null);

  useEffect(() => {
    if (!data || !chartRef.current) return;

    const ctx = chartRef.current.getContext('2d');

    if (chartInstance.current) {
      chartInstance.current.destroy();
    }

    chartInstance.current = new Chart(ctx, {
      type: 'doughnut',
      data: {
        labels: ['Used', 'Free'],
        datasets: [{
          data: [data.used, data.free],
          backgroundColor: ['#764ba2', '#e9ecef'],
          borderWidth: 0
        }]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: {
            position: 'bottom'
          }
        }
      }
    });

    return () => {
      if (chartInstance.current) {
        chartInstance.current.destroy();
      }
    };
  }, [data]);

  return React.createElement('canvas', { ref: chartRef });
}

// Network chart component
function NetworkChart({ data, history }) {
  const chartRef = useRef(null);
  const chartInstance = useRef(null);

  useEffect(() => {
    if (!data || !chartRef.current || history.length === 0) return;

    const ctx = chartRef.current.getContext('2d');

    if (chartInstance.current) {
      chartInstance.current.destroy();
    }

    chartInstance.current = new Chart(ctx, {
      type: 'line',
      data: {
        labels: history.map((_, i) => `${i * 5}s`),
        datasets: [{
          label: 'Bytes Sent',
          data: history.map(h => h.bytes_sent),
          borderColor: '#28a745',
          backgroundColor: 'rgba(40, 167, 69, 0.1)',
          borderWidth: 2,
          fill: false
        }, {
          label: 'Bytes Received',
          data: history.map(h => h.bytes_recv),
          borderColor: '#dc3545',
          backgroundColor: 'rgba(220, 53, 69, 0.1)',
          borderWidth: 2,
          fill: false
        }]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        scales: {
          y: {
            beginAtZero: true,
            ticks: {
              callback: function (value) {
                return formatBytes(value);
              }
            }
          }
        },
        plugins: {
          legend: {
            display: true,
            position: 'bottom'
          }
        }
      }
    });

    return () => {
      if (chartInstance.current) {
        chartInstance.current.destroy();
      }
    };
  }, [data, history]);

  return React.createElement('canvas', { ref: chartRef });
}

// Process table component
function ProcessTable({ processes }) {
  const [sortBy, setSortBy] = useState('cpu_percent');
  const [sortOrder, setSortOrder] = useState('desc');

  const sortedProcesses = [...processes].sort((a, b) => {
    const aVal = a[sortBy];
    const bVal = b[sortBy];
    const modifier = sortOrder === 'asc' ? 1 : -1;

    if (typeof aVal === 'string') {
      return aVal.localeCompare(bVal) * modifier;
    }
    return (aVal - bVal) * modifier;
  });

  const handleSort = (column) => {
    if (sortBy === column) {
      setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc');
    } else {
      setSortBy(column);
      setSortOrder('desc');
    }
  };

  return React.createElement('div', { className: 'process-table' },
    React.createElement('h3', { className: 'card-title' }, 'Top Processes'),
    React.createElement('table', null,
      React.createElement('thead', null,
        React.createElement('tr', null,
          React.createElement('th', { onClick: () => handleSort('pid') }, 'PID'),
          React.createElement('th', { onClick: () => handleSort('name') }, 'Name'),
          React.createElement('th', { onClick: () => handleSort('cpu_percent') }, 'CPU %'),
          React.createElement('th', { onClick: () => handleSort('mem_percent') }, 'Memory %')
        )
      ),
      React.createElement('tbody', null,
        sortedProcesses.slice(0, 10).map(process =>
          React.createElement('tr', { key: process.pid },
            React.createElement('td', null, process.pid),
            React.createElement('td', null, process.name),
            React.createElement('td', null, process.cpu_percent.toFixed(1) + '%'),
            React.createElement('td', null, process.mem_percent.toFixed(1) + '%')
          )
        )
      )
    )
  );
}

// Main App component
function App() {
  const [cpu, setCpu] = useState(null);
  const [memory, setMemory] = useState(null);
  const [network, setNetwork] = useState(null);
  const [processes, setProcesses] = useState([]);
  const [error, setError] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isOnline, setIsOnline] = useState(true);
  const [activeTab, setActiveTab] = useState('dashboard'); // 'dashboard' or 'alerts'

  // History for charts (keep last 20 data points)
  const [cpuHistory, setCpuHistory] = useState([]);
  const [networkHistory, setNetworkHistory] = useState([]);

  useEffect(() => {
    fetchData();
    const id = setInterval(fetchData, 5000);
    return () => clearInterval(id);
  }, []);

  async function fetchData() {
    try {
      const [cpuRes, memRes, netRes, procRes] = await Promise.all([
        fetch('/api/cpu'),
        fetch('/api/memory'),
        fetch('/api/network'),
        fetch('/api/process')
      ]);

      if (!cpuRes.ok || !memRes.ok || !netRes.ok || !procRes.ok) {
        throw new Error('Failed to fetch data from server');
      }

      const [cpuData, memData, netData, procData] = await Promise.all([
        cpuRes.json(),
        memRes.json(),
        netRes.json(),
        procRes.json()
      ]);

      setCpu(cpuData);
      setMemory(memData);
      setNetwork(netData);
      setProcesses(procData);

      // Update CPU history
      setCpuHistory(prev => {
        const newHistory = [...prev, cpuData.usage_percent];
        return newHistory.slice(-20); // Keep last 20 points
      });

      // Update network history
      setNetworkHistory(prev => {
        const newHistory = [...prev, netData];
        return newHistory.slice(-20); // Keep last 20 points
      });

      setError(null);
      setIsOnline(true);
      setIsLoading(false);
    } catch (err) {
      console.error('Error fetching data:', err);
      setError(err.message);
      setIsOnline(false);
      setIsLoading(false);
    }
  }

  if (isLoading) {
    return React.createElement('div', { className: 'container' },
      React.createElement('div', { className: 'loading' },
        'Loading system monitoring data...'
      )
    );
  }

  // Navigation tabs
  const renderNavigation = () => {
    return React.createElement('div', { className: 'navigation-tabs' },
      React.createElement('div', {
        className: `nav-tab ${activeTab === 'dashboard' ? 'active' : ''}`,
        onClick: () => setActiveTab('dashboard')
      }, 'System Dashboard'),
      React.createElement('div', {
        className: `nav-tab ${activeTab === 'alerts' ? 'active' : ''}`,
        onClick: () => setActiveTab('alerts')
      }, 'Alert Management')
    );
  };

  // Main content based on active tab
  const renderContent = () => {
    if (activeTab === 'alerts') {
      return React.createElement(window.AlertManagement);
    }

    // Default dashboard content
    return React.createElement(React.Fragment, null,
      // Dashboard cards
      React.createElement('div', { className: 'dashboard' },
        // CPU Card
        React.createElement('div', { className: 'card' },
          React.createElement('h3', { className: 'card-title' }, 'CPU Usage'),
          cpu && React.createElement('div', null,
            React.createElement('div', { className: 'metric' },
              React.createElement('span', { className: 'metric-label' }, 'Current Usage:'),
              React.createElement('span', { className: 'metric-value' }, cpu.usage_percent.toFixed(1) + '%')
            ),
            React.createElement('div', { className: 'metric' },
              React.createElement('span', { className: 'metric-label' }, 'Load Average (1m):'),
              React.createElement('span', { className: 'metric-value' }, cpu.load1.toFixed(2))
            ),
            React.createElement('div', { className: 'metric' },
              React.createElement('span', { className: 'metric-label' }, 'Load Average (5m):'),
              React.createElement('span', { className: 'metric-value' }, cpu.load5.toFixed(2))
            ),
            React.createElement('div', { className: 'metric' },
              React.createElement('span', { className: 'metric-label' }, 'Load Average (15m):'),
              React.createElement('span', { className: 'metric-value' }, cpu.load15.toFixed(2))
            ),
            React.createElement('div', { className: 'chart-container' },
              React.createElement(CPUChart, { data: cpu, history: cpuHistory })
            )
          )
        ),

        // Memory Card
        React.createElement('div', { className: 'card' },
          React.createElement('h3', { className: 'card-title' }, 'Memory Usage'),
          memory && React.createElement('div', null,
            React.createElement('div', { className: 'metric' },
              React.createElement('span', { className: 'metric-label' }, 'Total:'),
              React.createElement('span', { className: 'metric-value' }, formatBytes(memory.total))
            ),
            React.createElement('div', { className: 'metric' },
              React.createElement('span', { className: 'metric-label' }, 'Used:'),
              React.createElement('span', { className: 'metric-value' }, formatBytes(memory.used))
            ),
            React.createElement('div', { className: 'metric' },
              React.createElement('span', { className: 'metric-label' }, 'Free:'),
              React.createElement('span', { className: 'metric-value' }, formatBytes(memory.free))
            ),
            React.createElement('div', { className: 'metric' },
              React.createElement('span', { className: 'metric-label' }, 'Usage:'),
              React.createElement('span', { className: 'metric-value' }, memory.used_percent.toFixed(1) + '%')
            ),
            React.createElement('div', { className: 'chart-container' },
              React.createElement(MemoryChart, { data: memory })
            )
          )
        ),

        // Network Card
        React.createElement('div', { className: 'card' },
          React.createElement('h3', { className: 'card-title' }, 'Network Statistics'),
          network && React.createElement('div', null,
            React.createElement('div', { className: 'metric' },
              React.createElement('span', { className: 'metric-label' }, 'Bytes Sent:'),
              React.createElement('span', { className: 'metric-value' }, formatBytes(network.bytes_sent))
            ),
            React.createElement('div', { className: 'metric' },
              React.createElement('span', { className: 'metric-label' }, 'Bytes Received:'),
              React.createElement('span', { className: 'metric-value' }, formatBytes(network.bytes_recv))
            ),
            React.createElement('div', { className: 'metric' },
              React.createElement('span', { className: 'metric-label' }, 'Packets Sent:'),
              React.createElement('span', { className: 'metric-value' }, formatNumber(network.packets_sent))
            ),
            React.createElement('div', { className: 'metric' },
              React.createElement('span', { className: 'metric-label' }, 'Packets Received:'),
              React.createElement('span', { className: 'metric-value' }, formatNumber(network.packets_recv))
            ),
            React.createElement('div', { className: 'chart-container' },
              React.createElement(NetworkChart, { data: network, history: networkHistory })
            )
          )
        )
      ),

      // Process Table
      processes.length > 0 && React.createElement(ProcessTable, { processes })
    );
  };

  return React.createElement('div', { className: 'container' },
    // Header
    React.createElement('div', { className: 'header' },
      React.createElement('h1', null,
        React.createElement('span', {
          className: `status-indicator ${isOnline ? 'status-online' : 'status-offline'}`
        }),
        'Argus System Monitor'
      ),
      React.createElement('p', null, 'Real-time Linux system performance monitoring')
    ),

    // Error message
    error && React.createElement('div', { className: 'error' },
      'Error: ', error
    ),

    // Navigation
    renderNavigation(),

    // Main content
    renderContent()
  );
}

ReactDOM.createRoot(document.getElementById('root')).render(React.createElement(App));
