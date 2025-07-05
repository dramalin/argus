import React, { useCallback, useMemo } from 'react';
import {
  Box,
  CircularProgress,
  Alert,
  TextField,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Button,
  TableContainer,
  Table,
  TableHead,
  TableRow,
  TableCell,
  Paper,
  Pagination,
  Stack,
  Typography,
  Chip
} from '@mui/material';
import { FixedSizeList as List } from 'react-window';
import AutoSizer from 'react-virtualized-auto-sizer';
import AccessTimeIcon from '@mui/icons-material/AccessTime';
import type { ProcessInfo, ProcessQueryParams } from '../types/process';
import useDebounce from '../hooks/useDebounce';
import useThrottledCallback from '../hooks/useThrottledCallback';

interface ProcessTableProps {
  processes: ProcessInfo[];
  processParams: ProcessQueryParams;
  processTotal: number;
  processLoading: boolean;
  processError: string | null;
  lastUpdated: string | null;
  onParamChange: (key: keyof ProcessQueryParams, value: any) => void;
  onResetFilters: () => void;
}

/**
 * Row component for virtualized table
 */
type RowProps = {
  index: number;
  style: React.CSSProperties;
  data: {
    processes: ProcessInfo[];
    isOdd: (index: number) => boolean;
  };
};

const Row = React.memo(({ index, style, data }: RowProps) => {
  const { processes, isOdd } = data;
  const process = processes[index];
  
  if (!process) return null;
  
  return (
    <TableRow 
      component="div" 
      style={{ ...style, display: 'flex' }}
      sx={{ 
        '&:nth-of-type(odd)': { 
          bgcolor: theme => isOdd(index) ? 'action.hover' : 'inherit' 
        } 
      }}
    >
      <TableCell component="div" style={{ flex: '0 0 80px', display: 'flex', alignItems: 'center' }}>
        {process.pid}
      </TableCell>
      <TableCell component="div" style={{ flex: 1, display: 'flex', alignItems: 'center' }}>
        {process.name}
      </TableCell>
      <TableCell component="div" style={{ flex: '0 0 100px', display: 'flex', alignItems: 'center', justifyContent: 'flex-end' }}>
        {process.cpu_percent.toFixed(1)}
      </TableCell>
      <TableCell component="div" style={{ flex: '0 0 100px', display: 'flex', alignItems: 'center', justifyContent: 'flex-end' }}>
        {process.mem_percent.toFixed(1)}
      </TableCell>
    </TableRow>
  );
});

Row.displayName = 'ProcessTableRow';

/**
 * ProcessTable component
 * Displays a table of processes with virtualization for better performance
 */
const ProcessTable: React.FC<ProcessTableProps> = ({
  processes,
  processParams,
  processTotal,
  processLoading,
  processError,
  lastUpdated,
  onParamChange,
  onResetFilters
}) => {
  // For pagination
  const page = Math.floor((processParams.offset || 0) / (processParams.limit || 10)) + 1;
  const pageSize = processParams.limit || 10;
  const totalPages = Math.ceil(processTotal / pageSize);

  // Debounced filter values
  const debouncedNameFilter = useDebounce(processParams.name_contains || '', 300);
  const debouncedMinCpu = useDebounce(processParams.min_cpu, 300);
  const debouncedMinMemory = useDebounce(processParams.min_memory, 300);

  // Memoized handlers
  const handlePageChange = useCallback((newPage: number) => {
    onParamChange('offset', (newPage - 1) * pageSize);
  }, [onParamChange, pageSize]);

  const handleNameChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    onParamChange('name_contains', e.target.value);
  }, [onParamChange]);

  const handleMinCpuChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    onParamChange('min_cpu', e.target.value ? Number(e.target.value) : undefined);
  }, [onParamChange]);

  const handleMinMemoryChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    onParamChange('min_memory', e.target.value ? Number(e.target.value) : undefined);
  }, [onParamChange]);

  const handleSortByChange = useCallback((e: React.ChangeEvent<{ value: unknown }>) => {
    onParamChange('sort_by', e.target.value);
  }, [onParamChange]);

  const handleSortOrderChange = useCallback((e: React.ChangeEvent<{ value: unknown }>) => {
    onParamChange('sort_order', e.target.value as 'asc' | 'desc');
  }, [onParamChange]);

  const throttledResetFilters = useThrottledCallback(onResetFilters, 300);

  // Helper function to determine if a row is odd (for striping)
  const isOdd = useCallback((index: number) => index % 2 === 1, []);

  // Memoize row data to prevent unnecessary re-renders
  const rowData = useMemo(() => ({
    processes,
    isOdd,
  }), [processes, isOdd]);

  // Loading state
  if (processLoading && processes.length === 0) {
    return (
      <Box sx={{ p: 3, display: 'flex', justifyContent: 'center' }}>
        <CircularProgress />
      </Box>
    );
  }

  // Error state
  if (processError) {
    return (
      <Alert severity="error" sx={{ m: 2 }}>
        {processError}
      </Alert>
    );
  }

  return (
    <>
      {/* Filters */}
      <Box sx={{ p: 2, display: 'flex', gap: 2, flexWrap: 'wrap' }}>
        <TextField
          label="Filter by name"
          size="small"
          value={processParams.name_contains || ''}
          onChange={handleNameChange}
        />
        <TextField
          label="Min CPU %"
          type="number"
          size="small"
          value={processParams.min_cpu || ''}
          onChange={handleMinCpuChange}
        />
        <TextField
          label="Min Memory %"
          type="number"
          size="small"
          value={processParams.min_memory || ''}
          onChange={handleMinMemoryChange}
        />
        <FormControl size="small" sx={{ minWidth: 120 }}>
          <InputLabel>Sort by</InputLabel>
          <Select
            value={processParams.sort_by}
            label="Sort by"
            onChange={handleSortByChange}
          >
            <MenuItem value="cpu">CPU Usage</MenuItem>
            <MenuItem value="memory">Memory Usage</MenuItem>
            <MenuItem value="name">Name</MenuItem>
            <MenuItem value="pid">PID</MenuItem>
          </Select>
        </FormControl>
        <FormControl size="small" sx={{ minWidth: 120 }}>
          <InputLabel>Order</InputLabel>
          <Select
            value={processParams.sort_order}
            label="Order"
            onChange={handleSortOrderChange}
          >
            <MenuItem value="asc">Ascending</MenuItem>
            <MenuItem value="desc">Descending</MenuItem>
          </Select>
        </FormControl>
        <Button 
          variant="outlined" 
          onClick={throttledResetFilters}
          size="small"
        >
          Reset Filters
        </Button>
      </Box>

      {/* Virtualized Table */}
      <Paper sx={{ mx: 2, mb: 2, height: 560, overflow: 'hidden' }}>
        <Box sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
          {/* Table Header */}
          <TableContainer component="div" sx={{ overflow: 'hidden' }}>
            <Table component="div" sx={{ display: 'block' }} size="small" aria-label="process list">
              <TableHead component="div" sx={{ display: 'block' }}>
                <TableRow component="div" style={{ display: 'flex' }}>
                  <TableCell component="div" style={{ flex: '0 0 80px' }}>PID</TableCell>
                  <TableCell component="div" style={{ flex: 1 }}>Name</TableCell>
                  <TableCell component="div" style={{ flex: '0 0 100px', textAlign: 'right' }}>CPU %</TableCell>
                  <TableCell component="div" style={{ flex: '0 0 120px', textAlign: 'right' }}>Memory %</TableCell>
                </TableRow>
              </TableHead>
            </Table>
          </TableContainer>

          {/* Virtualized Rows */}
          <Box sx={{ flex: 1, overflow: 'hidden' }}>
            <AutoSizer>
              {({ height, width }) => (
                <List
                  height={height}
                  width={width}
                  itemCount={processes.length}
                  itemSize={48} // Adjust based on your row height
                  itemData={rowData}
                >
                  {Row}
                </List>
              )}
            </AutoSizer>
          </Box>
        </Box>
      </Paper>

      {/* Pagination */}
      <Box sx={{ p: 2, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Stack direction="row" spacing={1} alignItems="center">
          <Typography variant="body2" color="text.secondary">
            {`${processTotal} total processes`}
          </Typography>
          {lastUpdated && (
            <Chip
              size="small"
              icon={<AccessTimeIcon />}
              label={`Updated: ${new Date(lastUpdated).toLocaleTimeString()}`}
              variant="outlined"
            />
          )}
        </Stack>
        <Pagination
          count={totalPages}
          page={page}
          onChange={(_, newPage) => handlePageChange(newPage)}
          color="primary"
          size="small"
        />
      </Box>
    </>
  );
};

export default React.memo(ProcessTable); 