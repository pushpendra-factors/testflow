import React, { Component } from 'react';
import { Line } from 'react-chartjs-2';
import {
  Card,
  CardBody,
  CardHeader
} from 'reactstrap';
import { CustomTooltips } from '@coreui/coreui-plugin-chartjs-custom-tooltips';

const chartOptions = {
  tooltips: {
    enabled: false,
    custom: CustomTooltips
  },
  maintainAspectRatio: false
};
const lineBackgroundColors = ['rgba(75,192,192,0.4)', 'rgba(255,99,132,0.2)'];
const lineBorderColors = ['rgba(75,192,192,1)', 'rgba(255,99,132,1)'];
const lineHoverBackgroundColors = ['rgba(75,192,192,1)', 'rgba(255,99,132,0.4)'];
const lineHoverBorderColors = ['rgba(220,220,220,1)', 'rgba(255,99,132,0.4)'];

class LineChartCard extends Component {
  render() {
    var chartData = this.props.chartData;
    var line = {
      labels: chartData.labels,
      datasets: chartData.datasets,
    };
    // Styling.
    for (var i = 0; i < line.datasets.length; i++) {
      line.datasets[i].fill = false;
      line.datasets[i].lineTension = 0.1;
      line.datasets[i].backgroundColor = lineBackgroundColors[i % lineBackgroundColors.length];
      line.datasets[i].borderColor = lineBorderColors[i % lineBorderColors.length];
      line.datasets[i].borderCapStyle = 'butt';
      line.datasets[i].borderDash = [];
      line.datasets[i].borderDashOffset = 0.0;
      line.datasets[i].borderJoinStyle = 'miter';
      line.datasets[i].pointBorderColor = lineBorderColors[i % lineBorderColors.length];
      line.datasets[i].pointBackgroundColor = '#fff';
      line.datasets[i].pointBorderWidth = 1;
      line.datasets[i].pointHoverRadius = 5;
      line.datasets[i].pointHoverBackgroundColor = lineHoverBackgroundColors[i % lineHoverBackgroundColors.length];
      line.datasets[i].pointHoverBorderColor = lineHoverBorderColors[i % lineHoverBorderColors.length];
      line.datasets[i].pointHoverBorderWidth = 2;
      line.datasets[i].pointRadius = 1;
      line.datasets[i].pointHitRadius = 10;
    }

    var chart = <Line data={line} options={chartOptions} />

    return (
      <Card className="fapp-chart-card">
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
