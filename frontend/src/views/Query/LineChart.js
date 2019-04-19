import React, { Component } from 'react'
import { Line } from 'react-chartjs-2';
import { getColor, getChartScaleWithSpace } from '../../util';

class LineChart extends Component {
  constructor(props) {
    super(props)
  }

  createDataset(label, data=[], color) {
    let dataset = {
      fill: false,
      lineTension: 0.1,
      backgroundColor: color,
      borderColor: color,
      borderCapStyle: 'butt',
      borderWidth: 2,
      borderDash: [],
      borderDashOffset: 0.0,
      pointBorderColor: color,
      pointBackgroundColor: '#fff',
      pointBorderWidth: 3,
      pointHoverRadius: 1,
      pointHoverBackgroundColor: color,
      pointHoverBorderColor: color,
      pointHoverBorderWidth: 3,
      pointRadius: 0,
      pointHitRadius: 5,
    }

    dataset.data = data; // yAxis points. [65, 59, 80, 81, 56, 55, 40]
    dataset.label = label; // Line name on header. Event name.
    return dataset;
  }

  

  render() {
    let options = {
      maintainAspectRatio: false,
      responsive: true,
      scales: {
        yAxes: [{
          display: true,
          ticks: {
            beginAtZero: true,
            max: getChartScaleWithSpace(this.props.maxYScale)
          }
        }]
      }
    };

    let lines = this.props.lines;
    let datasets = [];
    let plotXAxisLabels = [];

    for(let li=0; li < lines.length; li++) {
      let line = lines[li];

      datasets.push(this.createDataset(line.title, line.yAxisLabels, getColor(li)));
      // merge xAxisLabels from multiple lines.
      for(let lxi=0; lxi < line.xAxisLabels.length; lxi++) {
        if (plotXAxisLabels.indexOf(line.xAxisLabels[lxi]) == -1) {
          plotXAxisLabels.push(line.xAxisLabels[lxi]);
        }
      }
    }

    let data = {
      labels: plotXAxisLabels, // ['January', 'February', 'March', 'April', 'May', 'June', 'July']
      datasets: datasets
    }
    
    // Todo: Support multiple lines. Individual line for a group by.
    return <Line data={data} options={options} />;  
  }
}

export default LineChart