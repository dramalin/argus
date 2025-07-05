import React from 'react';
import { Box, Typography, Paper } from '@mui/material';

/**
 * Tasks page component
 * Placeholder for future tasks functionality
 */
const Tasks: React.FC = () => {
  return (
    <Box sx={{ p: 3 }}>
      <Paper sx={{ p: 3 }}>
        <Typography variant="h4" gutterBottom>
          Tasks
        </Typography>
        <Typography variant="body1">
          This is a placeholder for the tasks page. It is lazy-loaded for better performance.
        </Typography>
      </Paper>
    </Box>
  );
};

export default Tasks; 