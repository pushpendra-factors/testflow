import React, { Component } from 'react';
import { Line } from 'react-chartjs-2';
import {
  Card,
  CardBody,
  CardHeader,
} from 'reactstrap';
import { CustomTooltips } from '@coreui/coreui-plugin-chartjs-custom-tooltips';

const chartOptions = {
  tooltips: {
    enabled: false,
    custom: CustomTooltips
  },
  maintainAspectRatio: false
};

class LineChartCard extends Component {
  render() {
    var chartData = this.props.chartData;
    var line = {
      labels: chartData.labels,
      datasets: chartData.datasets,
    };
    line.datasets[0].fill = false;
    line.datasets[0].lineTension = 0.1;
    line.datasets[0].backgroundColor = 'rgba(75,192,192,0.4)';
    line.datasets[0].borderColor = 'rgba(75,192,192,1)';
    line.datasets[0].borderCapStyle = 'butt';
    line.datasets[0].borderDash = [];
    line.datasets[0].borderDashOffset = 0.0;
    line.datasets[0].borderJoinStyle = 'miter';
    line.datasets[0].pointBorderColor = 'rgba(75,192,192,1)';
    line.datasets[0].pointBackgroundColor = '#fff';
    line.datasets[0].pointBorderWidth = 1;
    line.datasets[0].pointHoverRadius = 5;
    line.datasets[0].pointHoverBackgroundColor = 'rgba(75,192,192,1)';
    line.datasets[0].pointHoverBorderColor = 'rgba(220,220,220,1)';
    line.datasets[0].pointHoverBorderWidth = 2;
    line.datasets[0].pointRadius = 1;
    line.datasets[0].pointHitRadius = 10;
    var chart = <Line data={line} options={chartOptions} />

    return (
      <Card>
      <CardHeader>
      {chartData.header}
      </CardHeader>
      <CardBody>
      <div className="chart-wrapper">
      {chart}
      </div>
      </CardBody>
      </Card>
    );
  }
}

export default LineChartCard;
