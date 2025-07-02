// Alert configuration UI components for Argus System Monitor
const { useState, useEffect, useRef } = React;

// Utility function to format time
function formatTime(timestamp) {
	if (!timestamp) return 'N/A';
	const date = new Date(timestamp);
	return date.toLocaleString();
}

// Alert status badge component
function AlertStatusBadge({ state }) {
	const getStatusColor = () => {
		switch (state) {
			case 'active': return 'bg-danger';
			case 'pending': return 'bg-warning';
			case 'resolved': return 'bg-success';
			default: return 'bg-secondary';
		}
	};

	return React.createElement('span', {
		className: `status-badge ${getStatusColor()}`
	}, state);
}

// Alert severity badge component
function AlertSeverityBadge({ severity }) {
	const getSeverityColor = () => {
		switch (severity) {
			case 'critical': return 'bg-danger';
			case 'warning': return 'bg-warning';
			case 'info': return 'bg-info';
			default: return 'bg-secondary';
		}
	};

	return React.createElement('span', {
		className: `severity-badge ${getSeverityColor()}`
	}, severity);
}

// Alert list component
function AlertList({ onEditAlert, onDeleteAlert, onTestAlert }) {
	const [alerts, setAlerts] = useState([]);
	const [alertStatuses, setAlertStatuses] = useState({});
	const [isLoading, setIsLoading] = useState(true);
	const [error, setError] = useState(null);

	// Fetch alerts and their statuses
	useEffect(() => {
		const fetchAlerts = async () => {
			try {
				setIsLoading(true);
				const [alertsRes, statusesRes] = await Promise.all([
					fetch('/api/alerts'),
					fetch('/api/alerts/status')
				]);

				if (!alertsRes.ok || !statusesRes.ok) {
					throw new Error('Failed to fetch alerts data');
				}

				const [alertsData, statusesData] = await Promise.all([
					alertsRes.json(),
					statusesRes.json()
				]);

				setAlerts(alertsData);

				// Convert array of statuses to object keyed by alert ID
				const statusMap = {};
				statusesData.forEach(status => {
					statusMap[status.alert_id] = status;
				});
				setAlertStatuses(statusMap);

				setError(null);
			} catch (err) {
				console.error('Error fetching alerts:', err);
				setError(err.message);
			} finally {
				setIsLoading(false);
			}
		};

		fetchAlerts();
		const interval = setInterval(fetchAlerts, 30000); // Refresh every 30 seconds

		return () => clearInterval(interval);
	}, []);

	if (isLoading) {
		return React.createElement('div', { className: 'loading-spinner' }, 'Loading alerts...');
	}

	if (error) {
		return React.createElement('div', { className: 'error' }, 'Error: ', error);
	}

	if (alerts.length === 0) {
		return React.createElement('div', { className: 'no-alerts' },
			'No alerts configured. Click "Add Alert" to create your first alert.'
		);
	}

	return React.createElement('div', { className: 'alert-list' },
		alerts.map(alert => {
			const status = alertStatuses[alert.id] || { state: 'unknown' };

			return React.createElement('div', {
				key: alert.id,
				className: `alert-item ${status.state === 'active' ? 'alert-active' : ''}`
			},
				React.createElement('div', { className: 'alert-header' },
					React.createElement('h4', { className: 'alert-name' }, alert.name),
					React.createElement('div', { className: 'alert-badges' },
						React.createElement(AlertSeverityBadge, { severity: alert.severity }),
						React.createElement(AlertStatusBadge, { state: status.state })
					)
				),
				React.createElement('div', { className: 'alert-description' }, alert.description),
				React.createElement('div', { className: 'alert-details' },
					React.createElement('div', { className: 'alert-metric' },
						`${alert.threshold.metric_type} ${alert.threshold.metric_name} ${alert.threshold.operator} ${alert.threshold.value}`
					),
					status.current_value !== undefined && React.createElement('div', { className: 'alert-current-value' },
						`Current value: ${status.current_value.toFixed(2)}`
					),
					status.triggered_at && React.createElement('div', { className: 'alert-triggered' },
						`Triggered: ${formatTime(status.triggered_at)}`
					)
				),
				React.createElement('div', { className: 'alert-actions' },
					React.createElement('button', {
						className: 'btn btn-sm btn-primary',
						onClick: () => onEditAlert(alert)
					}, 'Edit'),
					React.createElement('button', {
						className: 'btn btn-sm btn-danger',
						onClick: () => onDeleteAlert(alert.id)
					}, 'Delete'),
					React.createElement('button', {
						className: 'btn btn-sm btn-secondary',
						onClick: () => onTestAlert(alert.id)
					}, 'Test')
				)
			);
		})
	);
}

// Alert form component for creating/editing alerts
function AlertForm({ alert = null, onSave, onCancel }) {
	const defaultAlert = {
		name: '',
		description: '',
		enabled: true,
		severity: 'warning',
		threshold: {
			metric_type: 'cpu',
			metric_name: 'usage_percent',
			operator: '>',
			value: 80,
			sustained_for: 3
		},
		notifications: [
			{
				type: 'in-app',
				enabled: true
			}
		]
	};

	const [formData, setFormData] = useState(alert || defaultAlert);
	const [validationError, setValidationError] = useState(null);

	// Handle form field changes
	const handleChange = (e) => {
		const { name, value, type, checked } = e.target;

		if (name.startsWith('threshold.')) {
			const field = name.split('.')[1];
			setFormData({
				...formData,
				threshold: {
					...formData.threshold,
					[field]: type === 'number' ? parseFloat(value) : value
				}
			});
		} else {
			setFormData({
				...formData,
				[name]: type === 'checkbox' ? checked : value
			});
		}
	};

	// Handle notification changes
	const handleNotificationChange = (index, field, value) => {
		const updatedNotifications = [...formData.notifications];

		if (field === 'type') {
			// Reset settings when changing notification type
			updatedNotifications[index] = {
				type: value,
				enabled: updatedNotifications[index].enabled,
				settings: value === 'email' ? { recipient: '' } : {}
			};
		} else if (field === 'enabled') {
			updatedNotifications[index].enabled = value;
		} else if (field.startsWith('settings.')) {
			const settingName = field.split('.')[1];
			updatedNotifications[index].settings = {
				...updatedNotifications[index].settings,
				[settingName]: value
			};
		}

		setFormData({
			...formData,
			notifications: updatedNotifications
		});
	};

	// Add a new notification channel
	const addNotification = () => {
		setFormData({
			...formData,
			notifications: [
				...formData.notifications,
				{
					type: 'in-app',
					enabled: true
				}
			]
		});
	};

	// Remove a notification channel
	const removeNotification = (index) => {
		const updatedNotifications = [...formData.notifications];
		updatedNotifications.splice(index, 1);

		setFormData({
			...formData,
			notifications: updatedNotifications
		});
	};

	// Handle form submission
	const handleSubmit = (e) => {
		e.preventDefault();
		setValidationError(null);

		// Basic validation
		if (!formData.name.trim()) {
			setValidationError('Alert name is required');
			return;
		}

		if (formData.notifications.length === 0) {
			setValidationError('At least one notification channel is required');
			return;
		}

		onSave(formData);
	};

	// Metric options based on selected metric type
	const getMetricNameOptions = () => {
		switch (formData.threshold.metric_type) {
			case 'cpu':
				return [
					{ value: 'usage_percent', label: 'CPU Usage %' },
					{ value: 'load1', label: 'Load Average (1m)' },
					{ value: 'load5', label: 'Load Average (5m)' },
					{ value: 'load15', label: 'Load Average (15m)' }
				];
			case 'memory':
				return [
					{ value: 'used_percent', label: 'Memory Usage %' },
					{ value: 'used', label: 'Memory Used' },
					{ value: 'free', label: 'Memory Free' }
				];
			case 'network':
				return [
					{ value: 'bytes_sent', label: 'Bytes Sent' },
					{ value: 'bytes_recv', label: 'Bytes Received' },
					{ value: 'packets_sent', label: 'Packets Sent' },
					{ value: 'packets_recv', label: 'Packets Received' }
				];
			default:
				return [];
		}
	};

	return React.createElement('div', { className: 'alert-form-container' },
		React.createElement('h3', null, alert ? 'Edit Alert' : 'Create New Alert'),

		validationError && React.createElement('div', { className: 'error' }, validationError),

		React.createElement('form', { onSubmit: handleSubmit, className: 'alert-form' },
			// Basic info section
			React.createElement('div', { className: 'form-section' },
				React.createElement('h4', null, 'Basic Information'),

				React.createElement('div', { className: 'form-group' },
					React.createElement('label', null, 'Name'),
					React.createElement('input', {
						type: 'text',
						name: 'name',
						value: formData.name,
						onChange: handleChange,
						className: 'form-control',
						placeholder: 'Alert name'
					})
				),

				React.createElement('div', { className: 'form-group' },
					React.createElement('label', null, 'Description'),
					React.createElement('textarea', {
						name: 'description',
						value: formData.description,
						onChange: handleChange,
						className: 'form-control',
						placeholder: 'Optional description'
					})
				),

				React.createElement('div', { className: 'form-group' },
					React.createElement('label', null, 'Severity'),
					React.createElement('select', {
						name: 'severity',
						value: formData.severity,
						onChange: handleChange,
						className: 'form-control'
					},
						React.createElement('option', { value: 'info' }, 'Info'),
						React.createElement('option', { value: 'warning' }, 'Warning'),
						React.createElement('option', { value: 'critical' }, 'Critical')
					)
				),

				React.createElement('div', { className: 'form-check' },
					React.createElement('input', {
						type: 'checkbox',
						name: 'enabled',
						checked: formData.enabled,
						onChange: handleChange,
						className: 'form-check-input',
						id: 'alertEnabled'
					}),
					React.createElement('label', { className: 'form-check-label', htmlFor: 'alertEnabled' }, 'Enabled')
				)
			),

			// Threshold section
			React.createElement('div', { className: 'form-section' },
				React.createElement('h4', null, 'Threshold Configuration'),

				React.createElement('div', { className: 'form-group' },
					React.createElement('label', null, 'Metric Type'),
					React.createElement('select', {
						name: 'threshold.metric_type',
						value: formData.threshold.metric_type,
						onChange: handleChange,
						className: 'form-control'
					},
						React.createElement('option', { value: 'cpu' }, 'CPU'),
						React.createElement('option', { value: 'memory' }, 'Memory'),
						React.createElement('option', { value: 'network' }, 'Network')
					)
				),

				React.createElement('div', { className: 'form-group' },
					React.createElement('label', null, 'Metric Name'),
					React.createElement('select', {
						name: 'threshold.metric_name',
						value: formData.threshold.metric_name,
						onChange: handleChange,
						className: 'form-control'
					},
						getMetricNameOptions().map(option =>
							React.createElement('option', { key: option.value, value: option.value }, option.label)
						)
					)
				),

				React.createElement('div', { className: 'form-row' },
					React.createElement('div', { className: 'form-group col' },
						React.createElement('label', null, 'Operator'),
						React.createElement('select', {
							name: 'threshold.operator',
							value: formData.threshold.operator,
							onChange: handleChange,
							className: 'form-control'
						},
							React.createElement('option', { value: '>' }, '>'),
							React.createElement('option', { value: '>=' }, '>='),
							React.createElement('option', { value: '<' }, '<'),
							React.createElement('option', { value: '<=' }, '<='),
							React.createElement('option', { value: '==' }, '=='),
							React.createElement('option', { value: '!=' }, '!=')
						)
					),

					React.createElement('div', { className: 'form-group col' },
						React.createElement('label', null, 'Value'),
						React.createElement('input', {
							type: 'number',
							name: 'threshold.value',
							value: formData.threshold.value,
							onChange: handleChange,
							className: 'form-control',
							step: '0.01'
						})
					)
				),

				React.createElement('div', { className: 'form-group' },
					React.createElement('label', null, 'Sustained For (checks)'),
					React.createElement('input', {
						type: 'number',
						name: 'threshold.sustained_for',
						value: formData.threshold.sustained_for,
						onChange: handleChange,
						className: 'form-control',
						min: '1'
					}),
					React.createElement('small', { className: 'form-text text-muted' },
						'Number of consecutive checks the condition must persist before triggering'
					)
				)
			),

			// Notifications section
			React.createElement('div', { className: 'form-section' },
				React.createElement('h4', null, 'Notification Channels'),

				formData.notifications.map((notification, index) =>
					React.createElement('div', { key: index, className: 'notification-item' },
						React.createElement('div', { className: 'form-row' },
							React.createElement('div', { className: 'form-group col' },
								React.createElement('label', null, 'Type'),
								React.createElement('select', {
									value: notification.type,
									onChange: (e) => handleNotificationChange(index, 'type', e.target.value),
									className: 'form-control'
								},
									React.createElement('option', { value: 'in-app' }, 'In-App'),
									React.createElement('option', { value: 'email' }, 'Email')
								)
							),

							React.createElement('div', { className: 'form-check col-auto align-self-end mb-2' },
								React.createElement('input', {
									type: 'checkbox',
									checked: notification.enabled,
									onChange: (e) => handleNotificationChange(index, 'enabled', e.target.checked),
									className: 'form-check-input',
									id: `notification-${index}-enabled`
								}),
								React.createElement('label', {
									className: 'form-check-label',
									htmlFor: `notification-${index}-enabled`
								}, 'Enabled')
							),

							React.createElement('div', { className: 'col-auto align-self-end mb-2' },
								formData.notifications.length > 1 && React.createElement('button', {
									type: 'button',
									className: 'btn btn-sm btn-danger',
									onClick: () => removeNotification(index)
								}, 'Remove')
							)
						),

						// Email-specific settings
						notification.type === 'email' && React.createElement('div', { className: 'form-group' },
							React.createElement('label', null, 'Recipient Email'),
							React.createElement('input', {
								type: 'email',
								value: notification.settings?.recipient || '',
								onChange: (e) => handleNotificationChange(index, 'settings.recipient', e.target.value),
								className: 'form-control',
								placeholder: 'email@example.com'
							})
						)
					)
				),

				React.createElement('button', {
					type: 'button',
					className: 'btn btn-sm btn-secondary',
					onClick: addNotification
				}, 'Add Notification Channel')
			),

			// Form actions
			React.createElement('div', { className: 'form-actions' },
				React.createElement('button', {
					type: 'button',
					className: 'btn btn-secondary',
					onClick: onCancel
				}, 'Cancel'),
				React.createElement('button', {
					type: 'submit',
					className: 'btn btn-primary'
				}, 'Save Alert')
			)
		)
	);
}

// Notification list component
function NotificationList() {
	const [notifications, setNotifications] = useState([]);
	const [isLoading, setIsLoading] = useState(true);
	const [error, setError] = useState(null);

	// Fetch notifications
	const fetchNotifications = async () => {
		try {
			setIsLoading(true);
			const response = await fetch('/api/alerts/notifications');

			if (!response.ok) {
				throw new Error('Failed to fetch notifications');
			}

			const data = await response.json();
			setNotifications(data);
			setError(null);
		} catch (err) {
			console.error('Error fetching notifications:', err);
			setError(err.message);
		} finally {
			setIsLoading(false);
		}
	};

	useEffect(() => {
		fetchNotifications();
		const interval = setInterval(fetchNotifications, 30000); // Refresh every 30 seconds

		return () => clearInterval(interval);
	}, []);

	// Mark notification as read
	const markAsRead = async (id) => {
		try {
			const response = await fetch(`/api/alerts/notifications/${id}/read`, {
				method: 'POST'
			});

			if (!response.ok) {
				throw new Error('Failed to mark notification as read');
			}

			// Update local state
			setNotifications(notifications.map(notification =>
				notification.id === id ? { ...notification, read: true } : notification
			));
		} catch (err) {
			console.error('Error marking notification as read:', err);
		}
	};

	// Mark all notifications as read
	const markAllAsRead = async () => {
		try {
			const response = await fetch('/api/alerts/notifications/read-all', {
				method: 'POST'
			});

			if (!response.ok) {
				throw new Error('Failed to mark all notifications as read');
			}

			// Update local state
			setNotifications(notifications.map(notification => ({ ...notification, read: true })));
		} catch (err) {
			console.error('Error marking all notifications as read:', err);
		}
	};

	// Clear all notifications
	const clearAll = async () => {
		try {
			const response = await fetch('/api/alerts/notifications', {
				method: 'DELETE'
			});

			if (!response.ok) {
				throw new Error('Failed to clear notifications');
			}

			// Update local state
			setNotifications([]);
		} catch (err) {
			console.error('Error clearing notifications:', err);
		}
	};

	if (isLoading) {
		return React.createElement('div', { className: 'loading-spinner' }, 'Loading notifications...');
	}

	if (error) {
		return React.createElement('div', { className: 'error' }, 'Error: ', error);
	}

	if (notifications.length === 0) {
		return React.createElement('div', { className: 'no-notifications' }, 'No notifications');
	}

	return React.createElement('div', { className: 'notifications-container' },
		React.createElement('div', { className: 'notifications-header' },
			React.createElement('h3', null, 'Notifications'),
			React.createElement('div', { className: 'notifications-actions' },
				React.createElement('button', {
					className: 'btn btn-sm btn-secondary',
					onClick: markAllAsRead
				}, 'Mark All Read'),
				React.createElement('button', {
					className: 'btn btn-sm btn-danger',
					onClick: clearAll
				}, 'Clear All')
			)
		),

		React.createElement('div', { className: 'notification-list' },
			notifications.map(notification =>
				React.createElement('div', {
					key: notification.id,
					className: `notification-item ${notification.read ? 'read' : 'unread'}`
				},
					React.createElement('div', { className: 'notification-header' },
						React.createElement('span', {
							className: `notification-severity ${notification.severity}`
						}, notification.severity),
						React.createElement('span', { className: 'notification-time' },
							formatTime(notification.timestamp)
						)
					),
					React.createElement('div', { className: 'notification-title' }, notification.title),
					React.createElement('div', { className: 'notification-message' }, notification.message),
					!notification.read && React.createElement('button', {
						className: 'btn btn-sm btn-outline-secondary',
						onClick: () => markAsRead(notification.id)
					}, 'Mark Read')
				)
			)
		)
	);
}

// Main alert management component
function AlertManagement() {
	const [view, setView] = useState('list'); // 'list', 'form', 'notifications', 'status', 'history'
	const [selectedAlert, setSelectedAlert] = useState(null);

	// Switch to alert form view for creating a new alert
	const handleAddAlert = () => {
		setSelectedAlert(null);
		setView('form');
	};

	// Switch to alert form view for editing an existing alert
	const handleEditAlert = (alert) => {
		setSelectedAlert(alert);
		setView('form');
	};

	// Delete an alert
	const handleDeleteAlert = async (id) => {
		if (!confirm('Are you sure you want to delete this alert?')) {
			return;
		}

		try {
			const response = await fetch(`/api/alerts/${id}`, {
				method: 'DELETE'
			});

			if (!response.ok) {
				throw new Error('Failed to delete alert');
			}

			// Refresh the list view
			setView('list');
		} catch (err) {
			console.error('Error deleting alert:', err);
			alert(`Error deleting alert: ${err.message}`);
		}
	};

	// Test an alert
	const handleTestAlert = async (id) => {
		try {
			const response = await fetch(`/api/alerts/test/${id}`, {
				method: 'POST'
			});

			if (!response.ok) {
				throw new Error('Failed to test alert');
			}

			alert('Test alert sent successfully');
		} catch (err) {
			console.error('Error testing alert:', err);
			alert(`Error testing alert: ${err.message}`);
		}
	};

	// Save an alert (create or update)
	const handleSaveAlert = async (formData) => {
		try {
			const isUpdate = !!selectedAlert;
			const url = isUpdate ? `/api/alerts/${selectedAlert.id}` : '/api/alerts';
			const method = isUpdate ? 'PUT' : 'POST';

			const response = await fetch(url, {
				method,
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify(formData)
			});

			if (!response.ok) {
				const errorData = await response.json();
				throw new Error(errorData.error || 'Failed to save alert');
			}

			// Return to list view
			setView('list');
		} catch (err) {
			console.error('Error saving alert:', err);
			alert(`Error saving alert: ${err.message}`);
		}
	};

	// Cancel form and return to list view
	const handleCancelForm = () => {
		setView('list');
	};

	// Navigation tabs for different views
	const renderNavigation = () => {
		return React.createElement('div', { className: 'alert-management-tabs' },
			React.createElement('div', {
				className: `alert-tab ${view === 'status' ? 'active' : ''}`,
				onClick: () => setView('status')
			}, 'Alert Status'),
			React.createElement('div', {
				className: `alert-tab ${view === 'list' ? 'active' : ''}`,
				onClick: () => setView('list')
			}, 'Alert Configurations'),
			React.createElement('div', {
				className: `alert-tab ${view === 'history' ? 'active' : ''}`,
				onClick: () => setView('history')
			}, 'Alert History'),
			React.createElement('div', {
				className: `alert-tab ${view === 'notifications' ? 'active' : ''}`,
				onClick: () => setView('notifications')
			}, 'Notifications')
		);
	};

	// Render the appropriate view
	const renderView = () => {
		switch (view) {
			case 'form':
				return React.createElement(AlertForm, {
					alert: selectedAlert,
					onSave: handleSaveAlert,
					onCancel: handleCancelForm
				});
			case 'notifications':
				return React.createElement(NotificationList);
			case 'status':
				return React.createElement(window.AlertStatusDashboard);
			case 'history':
				return React.createElement(window.AlertHistoryView);
			case 'list':
			default:
				return React.createElement(React.Fragment, null,
					React.createElement('div', { className: 'alert-list-header' },
						React.createElement('h3', null, 'Alert Configurations'),
						React.createElement('div', { className: 'alert-actions' },
							React.createElement('button', {
								className: 'btn btn-primary',
								onClick: handleAddAlert
							}, 'Add Alert')
						)
					),
					React.createElement(AlertList, {
						onEditAlert: handleEditAlert,
						onDeleteAlert: handleDeleteAlert,
						onTestAlert: handleTestAlert
					})
				);
		}
	};

	return React.createElement('div', { className: 'alert-management' },
		renderNavigation(),
		renderView()
	);
}

// Export components for use in the main app
window.AlertManagement = AlertManagement; 