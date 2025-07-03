interface LayoutProps {
  children?: React.ReactNode;
}

export const Layout: React.FC<LayoutProps> = ({ children }) => {
  return (
    <div className="layout">
      <header className="layout-header">
        <div className="header-container">
          <div className="header-logo">
            <h1>Argus System Monitor</h1>
          </div>
          <nav className="header-nav">
            <a href="#dashboard" className="nav-link">Dashboard</a>
            <a href="#tasks" className="nav-link">Tasks</a>
            <a href="#alerts" className="nav-link">Alerts</a>
            <a href="#processes" className="nav-link">Processes</a>
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
          <p>&copy; 2024 Argus System Monitor</p>
        </div>
      </footer>
    </div>
  );
};

export default Layout; 