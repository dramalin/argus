import { StrictMode, lazy, Suspense } from 'react'
import { createRoot } from 'react-dom/client'
import { prefetchComponents } from './utils/lazyImport'
import LoadingFallback from './components/LoadingFallback'
import './index.css'

// Lazy load the App component
const App = lazy(() => import('./App'))

// Dynamically import Chart.js to reduce initial bundle size
const registerChartComponents = async () => {
  const { 
    Chart, 
    CategoryScale, 
    LinearScale, 
    PointElement, 
    LineElement, 
    BarElement, 
    ArcElement, 
    Title, 
    Tooltip, 
    Legend 
  } = await import('chart.js')

  // Register Chart.js components
  Chart.register(
    CategoryScale,
    LinearScale,
    PointElement,
    LineElement,
    BarElement,
    ArcElement,
    Title,
    Tooltip,
    Legend
  )
}

// Register Chart.js components asynchronously
registerChartComponents().catch(console.error)

// Prefetch important components for better UX
prefetchComponents([
  () => import('./components/SystemOverview'),
  () => import('./components/MetricsCharts'),
  () => import('./components/ProcessTable'),
])

// Render the app
createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <Suspense fallback={<LoadingFallback message="Loading application..." contained={false} />}>
      <App />
    </Suspense>
  </StrictMode>,
)
