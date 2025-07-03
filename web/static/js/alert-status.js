// Alert status and history display components for Argus System Monitor
// React hooks are now globally available from shared.js
// No need to redeclare them here

// Use shared utility function for formatting time
const formatTime = window.Utils ? window.Utils.formatTime : function (timestamp) {
	if (!timestamp) return 'N/A';
	const date = new Date(timestamp);
	return date.toLocaleString();
};

// Utility function to format duration
function formatDuration(milliseconds) {
	if (!milliseconds) return 'N/A';

	const seconds = Math.floor(milliseconds / 1000);
	if (seconds < 60) return `${seconds} sec`;

	const minutes = Math.floor(seconds / 60);
	if (minutes < 60) return `${minutes} min ${seconds % 60} sec`;

	const hours = Math.floor(minutes / 60);
	if (hours < 24) return `${hours} hr ${minutes % 60} min`;

	const days = Math.floor(hours / 24);
	return `${days} days ${hours % 24} hr`;
}

// Alert status dashboard component
function AlertStatusDashboard() {
	const [alertStatuses, setAlertStatuses] = useState([]);
	const [alertConfigs, setAlertConfigs] = useState({});
	const [isLoading, setIsLoading] = useState(true);
	const [error, setError] = useState(null);
	const [activeAlerts, setActiveAlerts] = useState(0);

	// Fetch alert statuses and configurations
	useEffect(() => {
		const fetchData = async () => {
			try {
				setIsLoading(true);
				const [statusesRes, alertsRes] = await Promise.all([
					fetch('/api/alerts/status'),
					fetch('/api/alerts')
				]);

				if (!statusesRes.ok || !alertsRes.ok) {
					throw new Error('Failed to fetch alert data');
				}

				const [statusesData, alertsData] = await Promise.all([
					statusesRes.json(),
					alertsRes.json()
				]);

				// Convert alerts array to object keyed by ID for easier lookup
				const alertsMap = {};
				alertsData.forEach(alert => {
					alertsMap[alert.id] = alert;
				});

				setAlertConfigs(alertsMap);
				setAlertStatuses(statusesData);

				// Count active alerts
				const activeCount = statusesData.filter(status => status.state === 'active').length;
				setActiveAlerts(activeCount);

				setError(null);
			} catch (err) {
				console.error('Error fetching alert data:', err);
				setError(err.message);
			} finally {
				setIsLoading(false);
			}
		};

		fetchData();
		const interval = setInterval(fetchData, 10000); // Refresh every 10 seconds

		return () => clearInterval(interval);
	}, []);

	if (isLoading) {
		return React.createElement('div', { className: 'loading-spinner' }, 'Loading alert status...');
	}

	if (error) {
		return React.createElement('div', { className: 'error' }, 'Error: ', error);
	}

	return React.createElement('div', { className: 'alert-status-dashboard' },
		// Summary panel
		React.createElement('div', { className: 'alert-summary-panel' },
			React.createElement('div', { className: 'alert-summary-item' },
				React.createElement('div', { className: 'alert-summary-value' }, alertStatuses.length),
				React.createElement('div', { className: 'alert-summary-label' }, 'Total Alerts')
			),
			React.createElement('div', { className: `alert-summary-item ${activeAlerts > 0 ? 'active' : ''}` },
				React.createElement('div', { className: 'alert-summary-value' }, activeAlerts),
				React.createElement('div', { className: 'alert-summary-label' }, 'Active Alerts')
			),
			React.createElement('div', { className: 'alert-summary-item' },
				React.createElement('div', { className: 'alert-summary-value' },
					alertStatuses.filter(status => status.state === 'resolved').length
				),
				React.createElement('div', { className: 'alert-summary-label' }, 'Resolved Alerts')
			),
			React.createElement('div', { className: 'alert-summary-item' },
				React.createElement('div', { className: 'alert-summary-value' },
					alertStatuses.filter(status => status.state === 'pending').length
				),
				React.createElement('div', { className: 'alert-summary-label' }, 'Pending Alerts')
			)
		),

		// Active alerts panel
		activeAlerts > 0 && React.createElement('div', { className: 'active-alerts-panel' },
			React.createElement('h4', { className: 'panel-title' }, 'Active Alerts'),
			React.createElement('div', { className: 'active-alerts-list' },
				alertStatuses
					.filter(status => status.state === 'active')
					.map(status => {
						const alertConfig = alertConfigs[status.alert_id];
						if (!alertConfig) return null;

						return React.createElement('div', {
							key: status.alert_id,
							className: `active-alert-item severity-${alertConfig.severity}`
						},
							React.createElement('div', { className: 'active-alert-header' },
								React.createElement('div', { className: 'active-alert-name' }, alertConfig.name),
								React.createElement('div', { className: 'active-alert-severity' },
									React.createElement('span', {
										className: `severity-badge bg-${alertConfig.severity === 'critical' ? 'danger' :
											alertConfig.severity === 'warning' ? 'warning' : 'info'}`
									}, alertConfig.severity)
								)
							),
							React.createElement('div', { className: 'active-alert-message' }, status.message),
							React.createElement('div', { className: 'active-alert-details' },
								React.createElement('span', { className: 'active-alert-time' },
									`Triggered: ${formatTime(status.triggered_at)}`
								),
								React.createElement('span', { className: 'active-alert-duration' },
									`Duration: ${formatDuration(Date.now() - new Date(status.triggered_at).getTime())}`
								)
							)
						);
					})
			)
		),

		// Status overview panel
		React.createElement('div', { className: 'status-overview-panel' },
			React.createElement('h4', { className: 'panel-title' }, 'Alert Status Overview'),
			alertStatuses.length === 0 ?
				React.createElement('div', { className: 'no-alerts' }, 'No alerts configured') :
				React.createElement('table', { className: 'status-table' },
					React.createElement('thead', null,
						React.createElement('tr', null,
							React.createElement('th', null, 'Alert Name'),
							React.createElement('th', null, 'Status'),
							React.createElement('th', null, 'Current Value'),
							React.createElement('th', null, 'Threshold'),
							React.createElement('th', null, 'Last Updated')
						)
					),
					React.createElement('tbody', null,
						alertStatuses.map(status => {
							const alertConfig = alertConfigs[status.alert_id];
							if (!alertConfig) return null;

							return React.createElement('tr', {
								key: status.alert_id,
								className: status.state === 'active' ? 'status-active' : ''
							},
								React.createElement('td', null, alertConfig.name),
								React.createElement('td', null,
									React.createElement('span', {
										className: `status-badge bg-${status.state === 'active' ? 'danger' :
											status.state === 'pending' ? 'warning' : 'success'}`
									}, status.state)
								),
								React.createElement('td', null,
									status.current_value !== undefined ? status.current_value.toFixed(2) : 'N/A'
								),
								React.createElement('td', null,
									`${alertConfig.threshold.operator} ${alertConfig.threshold.value}`
								),
								React.createElement('td', null,
									status.triggered_at ? formatTime(status.triggered_at) :
										status.resolved_at ? formatTime(status.resolved_at) : 'N/A'
								)
							);
						})
					)
				)
		)
	);
}

// Alert history view component
function AlertHistoryView() {
	const [alertHistory, setAlertHistory] = useState([]);
	const [alertConfigs, setAlertConfigs] = useState({});
	const [isLoading, setIsLoading] = useState(true);
	const [error, setError] = useState(null);

	// Filtering and sorting state
	const [filters, setFilters] = useState({
		severity: 'all',
		status: 'all',
		timeRange: '24h'
	});
	const [sortBy, setSortBy] = useState('time');
	const [sortOrder, setSortOrder] = useState('desc');

	// Fetch alert history and configurations
	useEffect(() => {
		const fetchData = async () => {
			try {
				setIsLoading(true);

				// In a real implementation, we would fetch from a history endpoint
				// For this demo, we'll simulate by fetching current statuses and alerts
				const [statusesRes, alertsRes, notificationsRes] = await Promise.all([
					fetch('/api/alerts/status'),
					fetch('/api/alerts'),
					fetch('/api/alerts/notifications')
				]);

				if (!statusesRes.ok || !alertsRes.ok || !notificationsRes.ok) {
					throw new Error('Failed to fetch alert data');
				}

				const [statusesData, alertsData, notificationsData] = await Promise.all([
					statusesRes.json(),
					alertsRes.json(),
					notificationsRes.json()
				]);

				// Convert alerts array to object keyed by ID for easier lookup
				const alertsMap = {};
				alertsData.forEach(alert => {
					alertsMap[alert.id] = alert;
				});

				setAlertConfigs(alertsMap);

				// Create synthetic history from notifications and statuses
				const historyEntries = [];

				// Add entries from notifications
				notificationsData.forEach(notification => {
					// Find the associated alert if possible
					const alertId = notification.alert_id;
					const alertConfig = alertsMap[alertId];

					historyEntries.push({
						id: notification.id,
						alertId: alertId,
						alertName: alertConfig ? alertConfig.name : 'Unknown Alert',
						severity: notification.severity || (alertConfig ? alertConfig.severity : 'info'),
						status: notification.read ? 'acknowledged' : 'active',
						message: notification.message,
						timestamp: notification.timestamp,
						type: 'notification'
					});
				});

				// Add entries from statuses
				statusesData.forEach(status => {
					const alertConfig = alertsMap[status.alert_id];
					if (!alertConfig) return;

					// Add trigger entry if available
					if (status.triggered_at) {
						historyEntries.push({
							id: `${status.alert_id}-triggered`,
							alertId: status.alert_id,
							alertName: alertConfig.name,
							severity: alertConfig.severity,
							status: 'triggered',
							message: status.message || `Alert triggered: ${alertConfig.threshold.metric_type} ${alertConfig.threshold.metric_name} ${alertConfig.threshold.operator} ${alertConfig.threshold.value}`,
							timestamp: status.triggered_at,
							value: status.current_value,
							type: 'status'
						});
					}

					// Add resolve entry if available
					if (status.resolved_at) {
						historyEntries.push({
							id: `${status.alert_id}-resolved`,
							alertId: status.alert_id,
							alertName: alertConfig.name,
							severity: alertConfig.severity,
							status: 'resolved',
							message: `Alert resolved: ${alertConfig.threshold.metric_type} ${alertConfig.threshold.metric_name} returned to normal`,
							timestamp: status.resolved_at,
							value: status.current_value,
							type: 'status'
						});
					}
				});

				setAlertHistory(historyEntries);
				setError(null);
			} catch (err) {
				console.error('Error fetching alert history:', err);
				setError(err.message);
			} finally {
				setIsLoading(false);
			}
		};

		fetchData();
		const interval = setInterval(fetchData, 30000); // Refresh every 30 seconds

		return () => clearInterval(interval);
	}, []);

	// Apply filters and sorting to history data
	const getFilteredAndSortedHistory = () => {
		// Apply filters
		let filtered = [...alertHistory];

		if (filters.severity !== 'all') {
			filtered = filtered.filter(entry => entry.severity === filters.severity);
		}

		if (filters.status !== 'all') {
			filtered = filtered.filter(entry => entry.status === filters.status);
		}

		if (filters.timeRange !== 'all') {
			const now = Date.now();
			let cutoff;

			switch (filters.timeRange) {
				case '1h':
					cutoff = now - 60 * 60 * 1000;
					break;
				case '6h':
					cutoff = now - 6 * 60 * 60 * 1000;
					break;
				case '24h':
					cutoff = now - 24 * 60 * 60 * 1000;
					break;
				case '7d':
					cutoff = now - 7 * 24 * 60 * 60 * 1000;
					break;
				default:
					cutoff = 0;
			}

			filtered = filtered.filter(entry => new Date(entry.timestamp).getTime() > cutoff);
		}

		// Apply sorting
		filtered.sort((a, b) => {
			let aVal, bVal;

			switch (sortBy) {
				case 'time':
					aVal = new Date(a.timestamp).getTime();
					bVal = new Date(b.timestamp).getTime();
					break;
				case 'severity':
					// Map severity to numeric value for sorting
					const severityMap = { critical: 3, warning: 2, info: 1 };
					aVal = severityMap[a.severity] || 0;
					bVal = severityMap[b.severity] || 0;
					break;
				case 'status':
					aVal = a.status;
					bVal = b.status;
					break;
				case 'name':
					aVal = a.alertName;
					bVal = b.alertName;
					break;
				default:
					aVal = new Date(a.timestamp).getTime();
					bVal = new Date(b.timestamp).getTime();
			}

			// Apply sort order
			const modifier = sortOrder === 'asc' ? 1 : -1;

			if (typeof aVal === 'string') {
				return aVal.localeCompare(bVal) * modifier;
			}

			return (aVal - bVal) * modifier;
		});

		return filtered;
	};

	const filteredHistory = getFilteredAndSortedHistory();

	// Handle filter changes
	const handleFilterChange = (filterName, value) => {
		setFilters({
			...filters,
			[filterName]: value
		});
	};

	// Handle sort changes
	const handleSort = (column) => {
		if (sortBy === column) {
			setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc');
		} else {
			setSortBy(column);
			setSortOrder('desc');
		}
	};

	if (isLoading) {
		return React.createElement('div', { className: 'loading-spinner' }, 'Loading alert history...');
	}

	if (error) {
		return React.createElement('div', { className: 'error' }, 'Error: ', error);
	}

	return React.createElement('div', { className: 'alert-history-view' },
		// Filters section
		React.createElement('div', { className: 'history-filters' },
			React.createElement('div', { className: 'filter-group' },
				React.createElement('label', null, 'Severity:'),
				React.createElement('select', {
					value: filters.severity,
					onChange: (e) => handleFilterChange('severity', e.target.value),
					className: 'filter-select'
				},
					React.createElement('option', { value: 'all' }, 'All Severities'),
					React.createElement('option', { value: 'critical' }, 'Critical'),
					React.createElement('option', { value: 'warning' }, 'Warning'),
					React.createElement('option', { value: 'info' }, 'Info')
				)
			),
			React.createElement('div', { className: 'filter-group' },
				React.createElement('label', null, 'Status:'),
				React.createElement('select', {
					value: filters.status,
					onChange: (e) => handleFilterChange('status', e.target.value),
					className: 'filter-select'
				},
					React.createElement('option', { value: 'all' }, 'All Statuses'),
					React.createElement('option', { value: 'active' }, 'Active'),
					React.createElement('option', { value: 'triggered' }, 'Triggered'),
					React.createElement('option', { value: 'resolved' }, 'Resolved'),
					React.createElement('option', { value: 'acknowledged' }, 'Acknowledged')
				)
			),
			React.createElement('div', { className: 'filter-group' },
				React.createElement('label', null, 'Time Range:'),
				React.createElement('select', {
					value: filters.timeRange,
					onChange: (e) => handleFilterChange('timeRange', e.target.value),
					className: 'filter-select'
				},
					React.createElement('option', { value: '1h' }, 'Last Hour'),
					React.createElement('option', { value: '6h' }, 'Last 6 Hours'),
					React.createElement('option', { value: '24h' }, 'Last 24 Hours'),
					React.createElement('option', { value: '7d' }, 'Last 7 Days'),
					React.createElement('option', { value: 'all' }, 'All Time')
				)
			)
		),

		// History table
		filteredHistory.length === 0 ?
			React.createElement('div', { className: 'no-history' }, 'No alert history found for the selected filters') :
			React.createElement('table', { className: 'history-table' },
				React.createElement('thead', null,
					React.createElement('tr', null,
						React.createElement('th', {
							onClick: () => handleSort('time'),
							className: sortBy === 'time' ? `sort-${sortOrder}` : ''
						}, 'Time'),
						React.createElement('th', {
							onClick: () => handleSort('name'),
							className: sortBy === 'name' ? `sort-${sortOrder}` : ''
						}, 'Alert Name'),
						React.createElement('th', {
							onClick: () => handleSort('severity'),
							className: sortBy === 'severity' ? `sort-${sortOrder}` : ''
						}, 'Severity'),
						React.createElement('th', {
							onClick: () => handleSort('status'),
							className: sortBy === 'status' ? `sort-${sortOrder}` : ''
						}, 'Status'),
						React.createElement('th', null, 'Message')
					)
				),
				React.createElement('tbody', null,
					filteredHistory.map(entry =>
						React.createElement('tr', {
							key: entry.id,
							className: `history-row ${entry.status === 'active' || entry.status === 'triggered' ? 'history-active' : ''}`
						},
							React.createElement('td', { className: 'history-time' }, formatTime(entry.timestamp)),
							React.createElement('td', null, entry.alertName),
							React.createElement('td', null,
								React.createElement('span', {
									className: `severity-badge bg-${entry.severity === 'critical' ? 'danger' :
										entry.severity === 'warning' ? 'warning' : 'info'}`
								}, entry.severity)
							),
							React.createElement('td', null,
								React.createElement('span', {
									className: `status-badge bg-${entry.status === 'active' || entry.status === 'triggered' ? 'danger' :
										entry.status === 'acknowledged' ? 'warning' : 'success'}`
								}, entry.status)
							),
							React.createElement('td', { className: 'history-message' }, entry.message)
						)
					)
				)
			)
	);
}

// Export components using the component registry - do this immediately
(function registerComponents() {
	console.log("Registering Alert Status components...");

	// Register with component registry first
	if (window.ComponentRegistry) {
		window.ComponentRegistry.register('AlertStatusDashboard', AlertStatusDashboard);
		window.ComponentRegistry.register('AlertHistoryView', AlertHistoryView);
	} else {
		console.error("ComponentRegistry not found! Waiting 100ms to retry...");
		setTimeout(registerComponents, 100);
		return;
	}

	// Also export to window for backward compatibility
	window.AlertStatusDashboard = AlertStatusDashboard;
	window.AlertHistoryView = AlertHistoryView;

	// Debug log to verify component exports
	console.log("Alert status components exported and registered successfully");
})();