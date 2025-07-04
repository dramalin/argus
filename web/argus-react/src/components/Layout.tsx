import { useState } from 'react';

interface LayoutProps {
  children?: React.ReactNode;
}

export const Layout: React.FC<LayoutProps> = ({ children }) => {
  const [activeTab, setActiveTab] = useState('dashboard');

  return (
    <div className="layout">
      <header className="layout-header">
        <div className="header-container">
          <div className="header-logo">
            <h1>Argus System Monitor</h1>
            <p>Real-time Linux system monitoring</p>
          </div>
          <nav className="header-nav">
            <a 
              href="#dashboard" 
              className={`nav-link ${activeTab === 'dashboard' ? 'active' : ''}`}
              onClick={() => setActiveTab('dashboard')}
            >
              Dashboard
            </a>
            <a 
              href="#tasks" 
              className={`nav-link ${activeTab === 'tasks' ? 'active' : ''}`}
              onClick={() => setActiveTab('tasks')}
            >
              Tasks
            </a>
            <a 
              href="#alerts" 
              className={`nav-link ${activeTab === 'alerts' ? 'active' : ''}`}
              onClick={() => setActiveTab('alerts')}
            >
              Alerts
            </a>
            <a 
              href="#processes" 
              className={`nav-link ${activeTab === 'processes' ? 'active' : ''}`}
              onClick={() => setActiveTab('processes')}
            >
              Processes
            </a>
          </nav>
        </div>
      </header>
      
      <main className="layout-main">
        <div className="main-container">
          {children}
        </div>
      </main>
      
      <footer className="layout-footer">
        <div className="footer-container">
          <p>&copy; {new Date().getFullYear()} Argus System Monitor</p>
        </div>
      </footer>
    </div>
  );
};

export default Layout; 