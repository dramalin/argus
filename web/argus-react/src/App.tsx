import Layout from './components/Layout';
import Dashboard from './Dashboard';
import { ThemeProvider, CssBaseline } from '@mui/material';
import theme from './theme/theme';
import './App.css';

function App() {
  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Layout>
        <Dashboard />
      </Layout>
    </ThemeProvider>
  );
}

export default App;
