import React, { Component } from 'react';
import {
  Card,
  CardBody,
  CardHeader,
  CardTitle,
} from 'reactstrap';
import { Bar } from 'react-chartjs-2';
import { CustomTooltips } from '../../common/custom-tooltips';	

const barBackgroundColors = ['rgba(75,192,192,0.4)', 'rgba(255,99,132,0.2)'];
const barBorderColors = ['rgba(75,192,192,1)', 'rgba(255,99,132,1)'];	
const barHoverBackgroundColors = ['rgba(75,192,192,1)', 'rgba(255,99,132,0.4)'];	
const barHoverBorderColors = ['rgba(220,220,220,1)', 'rgba(255,99,132,0.4)'];


class BarChartCard extends Component {
  render() {
    var chartData = this.props.chartData;
    var chartOptions = {
      tooltips: {
        enabled: false,	
        custom: CustomTooltips	
      },	
      maintainAspectRatio: false,	
      scales: {	
        xAxes: [{	
          scaleLabel: {	
            display: false,	
          }	
        }],	
        yAxes: [{	
          scaleLabel: {	
            display: false,	
          }	
        }],	
      },	
    };	
    var bar = {	
      labels: chartData.labels,	
      datasets: chartData.datasets,	
    };	
    // Styling.	
    for (var i = 0; i < bar.datasets.length; i++) {	
      bar.datasets[i].backgroundColor = barBackgroundColors[i % barBackgroundColors.length];	
      bar.datasets[i].borderColor = barBorderColors[i % barBorderColors.length];	
      bar.datasets[i].borderWidth = 1;	
      bar.datasets[i].hoverBackgroundColor = barHoverBackgroundColors[i % barHoverBackgroundColors.length];	
      bar.datasets[i].hoverBorderColor = barHoverBorderColors[i % barHoverBorderColors.length];	
    }	

    if (chartData.x_label != "") {	
      chartOptions.scales.xAxes[0].scaleLabel.display = true;	
      chartOptions.scales.xAxes[0].scaleLabel.labelString = chartData.x_label;	
    }	
    if (chartData.y_label != "") {	
      chartOptions.scales.yAxes[0].scaleLabel.display = true	
      chartOptions.scales.yAxes[0].scaleLabel.labelString = chartData.y_label	
    }	
    var chart = <Bar data={bar} options={chartOptions} />

    const explanations = chartData.explanations.map((explainText) =>
      <CardTitle>{explainText}</CardTitle>
    );

    return (
      <Card className="fapp-chart-card">
        <CardHeader>
          {chartData.header}
        </CardHeader>
        {explanations}
        <CardBody>
          <div className="chart-wrapper" style={{ minHeight: '450px' }}>
            {chart}
          </div>
        </CardBody>
      </Card>
    );
  }
}

export default BarChartCard;
