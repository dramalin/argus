import React from 'react';
import { Box, Typography, Paper } from '@mui/material';

/**
 * Settings page component
 * Placeholder for future settings functionality
 */
const Settings: React.FC = () => {
  return (
    <Box sx={{ p: 3 }}>
      <Paper sx={{ p: 3 }}>
        <Typography variant="h4" gutterBottom>
          Settings
        </Typography>
        <Typography variant="body1">
          This is a placeholder for the settings page. It is lazy-loaded for better performance.
        </Typography>
      </Paper>
    </Box>
  );
};

export default Settings; 