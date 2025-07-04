import React from 'react';
import { Line, Bar, Pie, Doughnut } from 'react-chartjs-2';
import type { ChartData, ChartOptions } from 'chart.js';
import { Paper, Box, Typography, useMediaQuery, useTheme } from '@mui/material';

export type ChartType = 'line' | 'bar' | 'pie' | 'doughnut';

interface ChartWidgetProps {
  type: ChartType;
  data: ChartData<'line'> | ChartData<'bar'> | ChartData<'pie'> | ChartData<'doughnut'>;
  options?: ChartOptions<'line'> | ChartOptions<'bar'> | ChartOptions<'pie'> | ChartOptions<'doughnut'>;
  height?: number;
  width?: number;
  title?: string;
  description?: string;
  id?: string;
}

const ChartWidget: React.FC<ChartWidgetProps> = ({ 
  type, 
  data, 
  options, 
  height, 
  width, 
  title,
  description,
  id = `chart-${Math.random().toString(36).substring(2, 9)}` 
}) => {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const isTablet = useMediaQuery(theme.breakpoints.down('md'));
  
  // Adjust height based on screen size
  const responsiveHeight = isMobile ? 200 : (isTablet ? 250 : (height || 300));
  
  const renderChart = () => {
    // Common chart options with accessibility improvements
    const accessibleOptions = {
      ...options,
      plugins: {
        ...(options?.plugins || {}),
        // Add a11y plugin options if they exist
        ...(options?.plugins?.tooltip ? {
          tooltip: {
            ...options.plugins.tooltip,
            enabled: true,
            mode: 'nearest',
            intersect: false, // Makes tooltips more accessible
          }
        } : {}),
        legend: {
          ...(options?.plugins?.legend || {}),
          display: true,
          position: 'bottom' as const,
          labels: {
            ...(options?.plugins?.legend?.labels || {}),
            // Increase font size for better readability
            font: {
              size: isMobile ? 12 : 14,
            },
            // Increase padding for better touch targets
            padding: isMobile ? 15 : 10,
          },
        },
      },
      // Make responsive by default
      responsive: true,
      maintainAspectRatio: false,
    };

    switch (type) {
      case 'line':
        return (
          <Line 
            data={data as ChartData<'line'>} 
            options={accessibleOptions as ChartOptions<'line'>} 
            height={responsiveHeight} 
            width={width}
            aria-label={description || `Line chart${title ? ` for ${title}` : ''}`}
          />
        );
      case 'bar':
        return (
          <Bar 
            data={data as ChartData<'bar'>} 
            options={accessibleOptions as ChartOptions<'bar'>} 
            height={responsiveHeight} 
            width={width}
            aria-label={description || `Bar chart${title ? ` for ${title}` : ''}`}
          />
        );
      case 'pie':
        return (
          <Pie 
            data={data as ChartData<'pie'>} 
            options={accessibleOptions as ChartOptions<'pie'>} 
            height={responsiveHeight} 
            width={width}
            aria-label={description || `Pie chart${title ? ` for ${title}` : ''}`}
          />
        );
      case 'doughnut':
        return (
          <Doughnut 
            data={data as ChartData<'doughnut'>} 
            options={accessibleOptions as ChartOptions<'doughnut'>} 
            height={responsiveHeight} 
            width={width}
            aria-label={description || `Doughnut chart${title ? ` for ${title}` : ''}`}
          />
        );
      default:
        return <Typography color="error">Unsupported chart type</Typography>;
    }
  };

  return (
    <Paper 
      elevation={2} 
      sx={{ 
        p: { xs: 1, sm: 2 }, 
        borderRadius: 2, 
        height: '100%',
        display: 'flex',
        flexDirection: 'column'
      }}
      role="region"
      aria-label={title || "Chart"}
      id={id}
      tabIndex={0} // Make focusable for keyboard navigation
    >
      {title && (
        <Typography 
          variant="h6" 
          component="h3" 
          gutterBottom 
          align="center"
          id={`${id}-title`}
        >
          {title}
        </Typography>
      )}
      {description && (
        <Typography 
          variant="body2" 
          color="text.secondary" 
          sx={{ mb: 2, px: 1 }}
          id={`${id}-description`}
        >
          {description}
        </Typography>
      )}
      <Box 
        sx={{ 
          flexGrow: 1, 
          display: 'flex', 
          alignItems: 'center', 
          justifyContent: 'center',
          minHeight: { xs: '200px', sm: '250px', md: '300px' },
          overflow: 'hidden'
        }}
        aria-labelledby={title ? `${id}-title` : undefined}
        aria-describedby={description ? `${id}-description` : undefined}
      >
        {renderChart()}
      </Box>
    </Paper>
  );
};

export default ChartWidget;
