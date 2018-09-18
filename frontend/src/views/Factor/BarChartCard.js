import React, { Component } from 'react';
import { Bar } from 'react-chartjs-2';
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

class BarChartCard extends Component {
  render() {
    var chartData = this.props.chartData;
    var bar = {
      labels: chartData.labels,
      datasets: chartData.datasets,
    };
    bar.datasets[0].backgroundColor = 'rgba(255,99,132,0.2)';
    bar.datasets[0].borderColor = 'rgba(255,99,132,1)';
    bar.datasets[0].borderWidth = 1;
    bar.datasets[0].hoverBackgroundColor = 'rgba(255,99,132,0.4)';
    bar.datasets[0].hoverBorderColor = 'rgba(255,99,132,1)';
    var chart = <Bar data={bar} options={chartOptions} />

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

export default BarChartCard;
