import React from 'react';
import { 
  Box, 
  Typography, 
  Paper, 
  Stack,
  useTheme,
  useMediaQuery
} from '@mui/material';

/**
 * Props for the PageHeader component
 */
export interface PageHeaderProps {
  /** Title of the page */
  title: string;
  /** Optional description text */
  description?: string;
  /** Optional timestamp of when the data was last updated */
  lastUpdated?: string | null;
  /** Optional action buttons to display in the header */
  actions?: React.ReactNode;
  /** Whether the page is currently loading data */
  loading?: boolean;
}

/**
 * A reusable page header component that displays a title, description,
 * last updated timestamp, and action buttons.
 */
export const PageHeader: React.FC<PageHeaderProps> = ({
  title,
  description,
  lastUpdated,
  actions,
  loading = false
}) => {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  
  return (
    <Paper sx={{ p: 3, mb: 4 }}>
      <Box 
        sx={{ 
          display: 'flex', 
          flexDirection: isMobile ? 'column' : 'row',
          justifyContent: 'space-between', 
          alignItems: isMobile ? 'flex-start' : 'center',
          mb: description ? 2 : 0
        }}
      >
        <Typography 
          variant="h4" 
          component="h1" 
          gutterBottom={isMobile}
        >
          {title}
        </Typography>
        
        {actions && (
          <Stack 
            direction="row" 
            spacing={2}
            sx={{ 
              mt: isMobile ? 2 : 0,
              width: isMobile ? '100%' : 'auto',
              justifyContent: isMobile ? 'flex-start' : 'flex-end'
            }}
          >
            {actions}
          </Stack>
        )}
      </Box>
      
      {description && (
        <Typography variant="body1" paragraph>
          {description}
        </Typography>
      )}
      
      {lastUpdated && (
        <Typography variant="caption" color="text.secondary">
          Last updated: {lastUpdated}
        </Typography>
      )}
    </Paper>
  );
};

export default PageHeader; 