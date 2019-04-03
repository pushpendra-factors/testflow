import React, { Component } from 'react';
import { Bar } from 'react-chartjs-2';

import { CustomTooltips } from '@coreui/coreui-plugin-chartjs-custom-tooltips';

const barBackgroundColors = ['rgba(75,192,192,0.4)', 'rgba(255,99,132,0.2)'];
const barBorderColors = ['rgba(75,192,192,1)', 'rgba(255,99,132,1)'];
const barHoverBackgroundColors = ['rgba(75,192,192,1)', 'rgba(255,99,132,0.4)'];
const barHoverBorderColors = ['rgba(220,220,220,1)', 'rgba(255,99,132,0.4)'];

class BarChart extends Component {
  constructor(props) {
    super(props);
  }

  render() {
    var chartData = this.props.bars;
    let displayLegend = this.props.legend == false ? false : true;

    var chartOptions = {
      legend: {
        display: displayLegend
      },
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

    return <Bar data={bar} options={chartOptions} />
  }

}

export default BarChart;