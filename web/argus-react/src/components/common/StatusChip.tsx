import React from 'react';
import { Chip, type ChipProps } from '@mui/material';

export interface StatusConfig {
  label: string;
  color: 'default' | 'primary' | 'secondary' | 'error' | 'info' | 'success' | 'warning';
}

export interface StatusChipProps extends Omit<ChipProps, 'color' | 'label'> {
  status: string;
  statusMap: Record<string, StatusConfig>;
  defaultStatus?: StatusConfig;
}

const StatusChip: React.FC<StatusChipProps> = ({ 
  status, 
  statusMap, 
  defaultStatus = { label: 'Unknown', color: 'default' },
  ...rest 
}) => {
  const config = statusMap[status] || defaultStatus;

  return <Chip label={config.label} color={config.color} size="small" {...rest} />;
};

export default StatusChip; 