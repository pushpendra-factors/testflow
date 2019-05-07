import React, { Component } from 'react'
import { Line } from 'react-chartjs-2';
import moment from 'moment';

import { getColor, getChartScaleWithSpace } from '../../util';
import { HEADER_COUNT, HEADER_DATE } from './common';

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

  getLinesByGroupsIfExist(rows, countIndex, dateIndex) { 
    let lines = {}
    let keySep = " / ";
    let maxScale = 0;

    for(let i=0; i<Object.keys(rows).length; i++) {
      let row = rows[i.toString()];
      if (row == undefined) continue;

      // All group properties joined together 
      // with a seperator is a key.
      let key = "";
      for(let c=0; c < row.length; c++) {
        if(c != countIndex && c != dateIndex) {
          let prop = row[c];
          if (key === "") {
            key = prop;
            continue;
          }
          key = key + keySep + prop;
        }
      }
      
      // init.
      if (!(key in lines)) {
        lines[key] = { counts: [], timestamps: [] }
      }
      
      lines[key].counts.push(row[countIndex]);

      let isToday = moment(row[dateIndex]).isSame(moment(), 'year');
      let formatStr = isToday ? 'MMM DD' : 'MMM DD, YYYY';
      lines[key].timestamps.push(moment(row[dateIndex]).format(formatStr));
      
      if (maxScale < row[countIndex]) maxScale = row[countIndex];
    }
    
    return { lines: lines, maxScale: maxScale };
  }

  render() {
    let result = this.props.queryResult;
    let displayLegend = this.props.legend === false ? false : true;

    let countIndex = result.headers.indexOf(HEADER_COUNT);
    if (countIndex == -1) { 
        throw new Error('No counts to plot as lines.');
    }

    let dateIndex = result.headers.indexOf(HEADER_DATE);
    if (dateIndex == -1) { 
        throw new Error('No dates to plot as lines.');
    }

    let lines = [];
    let groups = this.getLinesByGroupsIfExist(result.rows, countIndex, dateIndex);
    for(let key in groups.lines) {
      let line = { title: key, xAxisLabels: groups.lines[key].timestamps, yAxisLabels: groups.lines[key].counts };
      lines.push(line);
    }

    let options = {
      maintainAspectRatio: false,
      responsive: true,
      legend: {
        display: displayLegend
      },
      scales: {
        yAxes: [{
          display: true,
          ticks: {
            beginAtZero: true,
            max: getChartScaleWithSpace(groups.maxScale) 
          }
        }]
      }
    };

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