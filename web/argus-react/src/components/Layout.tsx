import React, { useState, useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { 
  AppBar, 
  Toolbar, 
  Typography, 
  Container, 
  Box, 
  Tabs, 
  Tab, 
  useTheme,
  CssBaseline,
  useMediaQuery,
  IconButton,
  Drawer,
  List,
  ListItem,
  ListItemIcon,
  ListItemText,
  Divider
} from '@mui/material';
import DashboardIcon from '@mui/icons-material/Dashboard';
import TaskIcon from '@mui/icons-material/Task';
import NotificationsIcon from '@mui/icons-material/Notifications';
import TerminalIcon from '@mui/icons-material/Terminal';
import MenuIcon from '@mui/icons-material/Menu';

interface LayoutProps {
  children?: React.ReactNode;
}

// A11y props for tabs
function a11yProps(index: number) {
  return {
    id: `tab-${index}`,
    'aria-controls': `tabpanel-${index}`,
  };
}

export const Layout: React.FC<LayoutProps> = ({ children }) => {
  const navigate = useNavigate();
  const location = useLocation();
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));
  const [mobileOpen, setMobileOpen] = useState(false);
  
  // Menu items with their corresponding routes
  const menuItems = [
    { text: 'Dashboard', icon: <DashboardIcon />, path: '/', index: 0 },
    { text: 'Tasks', icon: <TaskIcon />, path: '/tasks', index: 1 },
    { text: 'Alerts', icon: <NotificationsIcon />, path: '/alerts', index: 2 },
    { text: 'Processes', icon: <TerminalIcon />, path: '/processes', index: 3 },
  ];
  
  // Determine the active tab based on current location
  const getActiveTabIndex = () => {
    const currentPath = location.pathname;
    const activeItem = menuItems.find(item => {
      if (item.path === '/' && currentPath === '/') {
        return true;
      }
      return currentPath.startsWith(item.path) && item.path !== '/';
    });
    return activeItem ? activeItem.index : 0;
  };
  
  const activeTabIndex = getActiveTabIndex();

  const handleChange = (_event: React.SyntheticEvent, newValue: number) => {
    const selectedItem = menuItems.find(item => item.index === newValue);
    if (selectedItem) {
      navigate(selectedItem.path);
    }
  };

  const handleDrawerToggle = () => {
    setMobileOpen(!mobileOpen);
  };

  const handleNavigation = (path: string) => {
    navigate(path);
    if (mobileOpen) {
      setMobileOpen(false);
    }
  };

  const drawer = (
    <Box onClick={handleDrawerToggle} sx={{ textAlign: 'center' }}>
      <Typography variant="h6" sx={{ my: 2 }}>
        Argus Monitor
      </Typography>
      <Divider />
      <List>
        {menuItems.map((item) => (
          <ListItem 
            button 
            key={item.text} 
            onClick={() => handleNavigation(item.path)}
            selected={activeTabIndex === item.index}
            sx={{
              '&.Mui-selected': {
                backgroundColor: theme.palette.primary.main,
                color: 'white',
                '& .MuiListItemIcon-root': {
                  color: 'white',
                }
              }
            }}
          >
            <ListItemIcon>{item.icon}</ListItemIcon>
            <ListItemText primary={item.text} />
          </ListItem>
        ))}
      </List>
    </Box>
  );

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', minHeight: '100vh' }}>
      <CssBaseline />
      <AppBar 
        position="static" 
        color="primary" 
        elevation={2}
        component="nav"
        sx={{ mb: 2 }}
      >
        <Toolbar>
          {isMobile && (
            <IconButton
              color="inherit"
              aria-label="open drawer"
              edge="start"
              onClick={handleDrawerToggle}
              sx={{ mr: 2 }}
            >
              <MenuIcon />
            </IconButton>
          )}
          <Box sx={{ flexGrow: 1 }}>
            <Typography variant="h5" component="h1">
              Argus System Monitor
            </Typography>
            <Typography variant="subtitle2" sx={{ opacity: 0.8 }}>
              Real-time Linux system monitoring
            </Typography>
          </Box>
          {!isMobile && (
            <Tabs 
              value={activeTabIndex} 
              onChange={handleChange} 
              textColor="inherit"
              indicatorColor="secondary"
              aria-label="navigation tabs"
              sx={{ 
                '& .MuiTab-root': { 
                  minWidth: 'unset',
                  px: 2,
                  color: 'rgba(255, 255, 255, 0.7)',
                  '&.Mui-selected': {
                    color: '#fff',
                  }
                }
              }}
            >
              {menuItems.map((item, index) => (
                <Tab 
                  key={item.text}
                  icon={item.icon} 
                  iconPosition="start" 
                  label={item.text} 
                  {...a11yProps(index)}
                />
              ))}
            </Tabs>
          )}
        </Toolbar>
      </AppBar>
      
      {/* Mobile drawer */}
      <Box component="nav">
        <Drawer
          variant="temporary"
          open={mobileOpen}
          onClose={handleDrawerToggle}
          ModalProps={{
            keepMounted: true, // Better mobile performance
          }}
          sx={{
            display: { xs: 'block', md: 'none' },
            '& .MuiDrawer-paper': { boxSizing: 'border-box', width: 240 },
          }}
        >
          {drawer}
        </Drawer>
      </Box>
      
      <Container 
        component="main" 
        sx={{ 
          flexGrow: 1, 
          py: 2,
          px: { xs: 2, sm: 3 },
          display: 'flex',
          flexDirection: 'column'
        }}
      >
        {children}
      </Container>
      
      <Box 
        component="footer" 
        sx={{ 
          py: 3, 
          bgcolor: theme.palette.primary.main,
          color: 'rgba(255, 255, 255, 0.7)',
          mt: 'auto'
        }}
      >
        <Container maxWidth="lg">
          <Typography variant="body2" align="center">
            &copy; {new Date().getFullYear()} Argus System Monitor
          </Typography>
        </Container>
      </Box>
    </Box>
  );
};

export default Layout; 