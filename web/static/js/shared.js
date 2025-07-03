// Shared React hooks and utilities for Argus System Monitor
// This file must be loaded before any other React components

// Debug log to verify shared hooks are loaded
console.log("Initializing shared React hooks...");

// Export React hooks directly to window (global scope)
// This avoids redeclaration issues across multiple files
window.useState = React.useState;
window.useEffect = React.useEffect;
window.useRef = React.useRef;
window.Component = React.Component;

console.log("Shared React hooks initialized successfully");

// Utility functions that can be shared across components
window.Utils = {
	// Format timestamp to locale string
	formatTime: function (timestamp) {
		if (!timestamp) return 'N/A';
		const date = new Date(timestamp);
		return date.toLocaleString();
	},

	// Format bytes to human-readable format
	formatBytes: function (bytes) {
		if (bytes === 0) return '0 B';
		const k = 1024;
		const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
		const i = Math.floor(Math.log(bytes) / Math.log(k));
		return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
	},

	// Format number with thousands separator
	formatNumber: function (num) {
		return new Intl.NumberFormat().format(num);
	}
};

console.log("Shared utilities initialized");

// Create a component registry to avoid circular dependencies
window.ComponentRegistry = {
	_components: {},
	register: function (name, component) {
		console.log(`Registering component: ${name}`);
		this._components[name] = component;

		// Notify any waiting components
		if (this.onComponentRegistered && typeof this.onComponentRegistered === 'function') {
			setTimeout(() => {
				this.onComponentRegistered(name, component);
			}, 0); // Use setTimeout to ensure this runs after current execution
		}
	},
	get: function (name) {
		return this._components[name];
	},
	onComponentRegistered: null,
	// Debug method to list all registered components
	listRegistered: function () {
		console.log("Registered components:", Object.keys(this._components));
		return Object.keys(this._components);
	}
};