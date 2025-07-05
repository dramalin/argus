import React from 'react';
import { Box, CircularProgress, Typography, Paper } from '@mui/material';

/**
 * Props for the LoadingFallback component
 */
interface LoadingFallbackProps {
  /** Message to display while loading */
  message?: string;
  /** Height of the loading container */
  height?: string | number;
  /** Whether to show the loading indicator in a contained box */
  contained?: boolean;
}

/**
 * LoadingFallback component
 * Displays a loading indicator and optional message
 * Used as a fallback for React.lazy() and Suspense
 */
const LoadingFallback: React.FC<LoadingFallbackProps> = ({ 
  message = 'Loading...', 
  height = '200px',
  contained = true
}) => {
  const content = (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        height: height,
        width: '100%',
      }}
    >
      <CircularProgress size={40} thickness={4} />
      {message && (
        <Typography
          variant="body1"
          color="text.secondary"
          sx={{ mt: 2 }}
        >
          {message}
        </Typography>
      )}
    </Box>
  );

  if (contained) {
    return (
      <Paper
        elevation={0}
        sx={{
          p: 3,
          m: 2,
          backgroundColor: 'background.default',
          borderRadius: 2,
        }}
      >
        {content}
      </Paper>
    );
  }

  return content;
};

export default LoadingFallback; 