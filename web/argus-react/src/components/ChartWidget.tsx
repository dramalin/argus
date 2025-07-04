import React from 'react';
import { Line, Bar, Pie, Doughnut } from 'react-chartjs-2';
import type { ChartData, ChartOptions } from 'chart.js';

export type ChartType = 'line' | 'bar' | 'pie' | 'doughnut';

interface ChartWidgetProps {
  type: ChartType;
  data: ChartData<'line'> | ChartData<'bar'> | ChartData<'pie'> | ChartData<'doughnut'>;
  options?: ChartOptions<'line'> | ChartOptions<'bar'> | ChartOptions<'pie'> | ChartOptions<'doughnut'>;
  height?: number;
  width?: number;
}

const ChartWidget: React.FC<ChartWidgetProps> = ({ type, data, options, height, width }) => {
  switch (type) {
    case 'line':
      return <Line data={data as ChartData<'line'>} options={options as ChartOptions<'line'>} height={height} width={width} />;
    case 'bar':
      return <Bar data={data as ChartData<'bar'>} options={options as ChartOptions<'bar'>} height={height} width={width} />;
    case 'pie':
      return <Pie data={data as ChartData<'pie'>} options={options as ChartOptions<'pie'>} height={height} width={width} />;
    case 'doughnut':
      return <Doughnut data={data as ChartData<'doughnut'>} options={options as ChartOptions<'doughnut'>} height={height} width={width} />;
    default:
      return <div>Unsupported chart type</div>;
  }
};

export default ChartWidget;
