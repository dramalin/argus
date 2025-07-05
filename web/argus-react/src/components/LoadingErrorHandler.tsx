import React from 'react';
import { Box, CircularProgress, Typography, Alert, Button } from '@mui/material';
import RefreshIcon from '@mui/icons-material/Refresh';

interface LoadingErrorHandlerProps {
  loading: boolean;
  error: string | null;
  children: React.ReactNode;
  loadingMessage?: string;
}

const LoadingErrorHandler: React.FC<LoadingErrorHandlerProps> = ({
  loading,
  error,
  children,
  loadingMessage = 'Loading system metrics...'
}) => {
  if (loading) {
    return (
      <Box 
        sx={{ 
          display: 'flex', 
          justifyContent: 'center', 
          alignItems: 'center', 
          minHeight: '50vh',
          flexDirection: 'column'
        }}
        role="status"
        aria-live="polite"
        aria-busy="true"
      >
        <CircularProgress color="primary" aria-hidden="true" />
        <Typography variant="h6" sx={{ ml: 2, mt: 2 }}>
          {loadingMessage}
        </Typography>
      </Box>
    );
  }

  if (error) {
    return (
      <Box 
        sx={{ maxWidth: 600, mx: 'auto', mt: 4 }}
        role="alert"
        aria-live="assertive"
      >
        <Alert 
          severity="error" 
          action={
            <Button 
              color="inherit" 
              size="small" 
              startIcon={<RefreshIcon />}
              onClick={() => window.location.reload()}
              aria-label="Retry loading metrics"
            >
              Retry
            </Button>
          }
        >
          <Typography variant="h6">Error loading metrics</Typography>
          <Typography>{error}</Typography>
        </Alert>
      </Box>
    );
  }

  return <>{children}</>;
};

export default React.memo(LoadingErrorHandler); 