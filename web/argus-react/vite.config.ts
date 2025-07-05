import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import { visualizer } from 'rollup-plugin-visualizer'
import { compression } from 'vite-plugin-compression2'

/**
 * Vite configuration
 * @see https://vitejs.dev/config/
 */
export default defineConfig(({ mode }) => {
  // Load environment variables based on mode
  const env = loadEnv(mode, process.cwd(), '');
  
  const isProduction = mode === 'production';
  
  return {
    base: '/',
    plugins: [
      react(),
      // Generate bundle visualization in production
      isProduction && visualizer({
        filename: 'dist/stats.html',
        gzipSize: true,
        brotliSize: true,
        open: false
      }),
      // Compress assets in production
      isProduction && compression({
        algorithm: 'brotliCompress',
        exclude: [/\.(br)$/, /\.(gz)$/],
        threshold: 10240, // only compress files > 10kb
      }),
      // Gzip compression as fallback
      isProduction && compression({
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
          target: env.VITE_API_BASE_URL?.replace('http', 'ws') || 'ws://localhost:8080',
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
  }
})
