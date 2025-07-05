import React from 'react';
import { Box, Typography, Paper, Button } from '@mui/material';

/**
 * NotFound page component
 * Displayed when a route is not found
 */
const NotFound: React.FC = () => {
  return (
    <Box sx={{ p: 3, display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '50vh' }}>
      <Paper sx={{ p: 4, maxWidth: '500px', textAlign: 'center' }}>
        <Typography variant="h2" gutterBottom>
          404
        </Typography>
        <Typography variant="h4" gutterBottom>
          Page Not Found
        </Typography>
        <Typography variant="body1" sx={{ mb: 3 }}>
          The page you are looking for does not exist or has been moved.
        </Typography>
        <Button
          variant="contained"
          color="primary"
          onClick={() => window.location.href = '/'}
        >
          Go to Dashboard
        </Button>
      </Paper>
    </Box>
  );
};

export default NotFound; 