import React from 'react';
import { Card, CardContent, Typography, Divider, Skeleton } from '@mui/material';

interface MetricCardProps {
  title: string;
  value?: string | number;
  unit?: string;
  loading: boolean;
  details?: Array<{ label: string; value: string | number }>;
  titleId?: string;
}

// Not used for now
// export const MetricCardSkeleton: React.FC = () => (
//   <Card elevation={2}>
//     <CardContent>
//       <Skeleton variant="text" width="60%" height={30} />
//       <Skeleton variant="text" width="40%" height={40} sx={{ my: 1 }} />
//       <Divider sx={{ my: 1 }} />
//       <Skeleton variant="text" width="80%" />
//       <Skeleton variant="text" width="70%" />
//       <Skeleton variant="text" width="75%" />
//     </CardContent>
//   </Card>
// );

const SystemMetricsCard: React.FC<MetricCardProps> = ({
  title,
  value,
  unit = '',
  loading,
  details = [],
  titleId,
}) => {
  // Not used for now
  // if (loading) {
  //   return <MetricCardSkeleton />;
  // }

  return (
    <Card
      elevation={2}
      component="section"
      aria-labelledby={titleId || `${title.toLowerCase().replace(/\s+/g, '-')}-title`}
    >
      <CardContent>
        <Typography 
          variant="h6" 
          component="h3" 
          gutterBottom 
          id={titleId || `${title.toLowerCase().replace(/\s+/g, '-')}-title`}
        >
          {title}
        </Typography>
        {value !== undefined && (
          <Typography
            variant="h4"
            color="primary"
            gutterBottom
            aria-label={`${title} ${value}${unit}`}
          >
            {value}
            {unit}
          </Typography>
        )}
        <Divider sx={{ my: 1 }} />
        {details.map((detail, index) => (
          <Typography key={index} variant="body2" color="text.secondary">
            {detail.label}: {detail.value}
          </Typography>
        ))}
      </CardContent>
    </Card>
  );
};

export default React.memo(SystemMetricsCard); 