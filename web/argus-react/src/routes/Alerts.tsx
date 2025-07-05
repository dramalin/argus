import React from 'react';
import { Box, Typography, Paper } from '@mui/material';

/**
 * Alerts page component
 * Placeholder for future alerts functionality
 */
const Alerts: React.FC = () => {
  return (
    <Box sx={{ p: 3 }}>
      <Paper sx={{ p: 3 }}>
        <Typography variant="h4" gutterBottom>
          Alerts
        </Typography>
        <Typography variant="body1">
          This is a placeholder for the alerts page. It is lazy-loaded for better performance.
        </Typography>
      </Paper>
    </Box>
  );
};

export default Alerts; 