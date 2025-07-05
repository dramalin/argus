"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
var vite_1 = require("vite");
var plugin_react_1 = require("@vitejs/plugin-react");
var rollup_plugin_visualizer_1 = require("rollup-plugin-visualizer");
var vite_plugin_compression2_1 = require("vite-plugin-compression2");
/**
 * Vite configuration
 * @see https://vitejs.dev/config/
 */
exports.default = (0, vite_1.defineConfig)(function (_a) {
    var _b;
    var mode = _a.mode;
    // Load environment variables based on mode
    var env = (0, vite_1.loadEnv)(mode, process.cwd(), '');
    var isProduction = mode === 'production';
    return {
        base: '/',
        plugins: [
            (0, plugin_react_1.default)(),
            // Generate bundle visualization in production
            isProduction && (0, rollup_plugin_visualizer_1.visualizer)({
                filename: 'dist/stats.html',
                gzipSize: true,
                brotliSize: true,
                open: false
            }),
            // Compress assets in production
            isProduction && (0, vite_plugin_compression2_1.compression)({
                algorithm: 'brotliCompress',
                exclude: [/\.(br)$/, /\.(gz)$/],
                threshold: 10240, // only compress files > 10kb
            }),
            // Gzip compression as fallback
            isProduction && (0, vite_plugin_compression2_1.compression)({
                algorithm: 'gzip',
                exclude: [/\.(br)$/, /\.(gz)$/],
                threshold: 10240,
            }),
        ],
        server: {
            port: 5173,
            host: true, // Allow external connections
            proxy: {
                // Proxy API calls to the Go backend
                '/api': {
                    target: env.VITE_API_BASE_URL || 'http://localhost:8080',
                    changeOrigin: true,
                    secure: false,
                },
                '/health': {
                    target: env.VITE_API_BASE_URL || 'http://localhost:8080',
                    changeOrigin: true,
                    secure: false,
                },
                '/ws': {
                    target: ((_b = env.VITE_API_BASE_URL) === null || _b === void 0 ? void 0 : _b.replace('http', 'ws')) || 'ws://localhost:8080',
                    ws: true,
                    changeOrigin: true,
                }
            }
        },
        build: {
            outDir: 'dist',
            // Generate source maps in development, but not in production for better performance
            sourcemap: !isProduction,
            // Minify options
            minify: isProduction ? 'terser' : false,
            terserOptions: {
                compress: {
                    drop_console: isProduction, // Remove console.log in production
                    drop_debugger: isProduction, // Remove debugger statements in production
                }
            },
            // Optimize chunk splitting
            rollupOptions: {
                output: {
                    manualChunks: {
                        vendor: ['react', 'react-dom'],
                        mui: ['@mui/material', '@mui/icons-material', '@emotion/react', '@emotion/styled'],
                        charts: ['chart.js', 'react-chartjs-2'],
                    }
                }
            },
            // Optimize CSS
            cssCodeSplit: true,
            // Reduce chunk size warnings threshold
            chunkSizeWarningLimit: 1000,
        },
        define: {
            // Define environment variables
            __APP_VERSION__: JSON.stringify(process.env.npm_package_version),
            // Make environment variables available in the app
            'process.env.NODE_ENV': JSON.stringify(mode),
        },
        resolve: {
            // Set up path aliases
            alias: {
                '@': '/src',
                '@components': '/src/components',
                '@hooks': '/src/hooks',
                '@utils': '/src/utils',
                '@contexts': '/src/contexts',
                '@types': '/src/types',
            }
        },
        // Optimize dependencies pre-bundling
        optimizeDeps: {
            include: ['react', 'react-dom', 'chart.js', 'react-chartjs-2'],
            exclude: []
        },
        // CSS optimization
        css: {
            devSourcemap: true,
            preprocessorOptions: {
            // Add preprocessor options if needed
            }
        },
        // Enable JSON import with type support
        json: {
            stringify: true,
        }
    };
});
