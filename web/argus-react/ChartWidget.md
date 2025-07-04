# ChartWidget Component

The `ChartWidget` component is a reusable React component for rendering charts using [react-chartjs-2](https://react-chartjs-2.js.org/) and [Chart.js](https://www.chartjs.org/).

## Usage

```
import ChartWidget from './components/ChartWidget';

const data = {
  labels: ['CPU Usage'],
  datasets: [
    {
      label: 'CPU Usage (%)',
      data: [42],
      backgroundColor: 'rgba(75,192,192,0.4)',
      borderColor: 'rgba(75,192,192,1)',
      borderWidth: 1,
    },
  ],
};

<ChartWidget type="bar" data={data} options={{ responsive: true }} />
```

## Props

| Name    | Type         | Description                                  |
| ------- | ------------ | -------------------------------------------- |
| type    | ChartType    | 'line' \| 'bar' \| 'pie' \| 'doughnut'      |
| data    | ChartData    | Chart.js data object                         |
| options | ChartOptions | (optional) Chart.js options object           |
| height  | number       | (optional) Chart height                      |
| width   | number       | (optional) Chart width                       |

## Examples

### Line Chart
```
<ChartWidget type="line" data={lineData} options={lineOptions} />
```

### Pie Chart
```
<ChartWidget type="pie" data={pieData} />
```

See the [react-chartjs-2 documentation](https://react-chartjs-2.js.org/) for more details on data and options formats.
