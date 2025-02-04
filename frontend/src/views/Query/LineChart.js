import React, { Component } from 'react'
import { Line } from 'react-chartjs-2';
import moment from 'moment';
import mt from "moment-timezone"

import { getColor, getChartScaleWithSpace } from '../../util';
import { HEADER_COUNT, HEADER_DATE, getYAxesStr } from './common';


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

  getLinesByGroupsIfExist(rows, timestampType, countIndex, dateTimeIndex, timezone) { 
    if (!timezone || timezone == "" || !mt.tz(timezone)){
      console.error("Invalid timezone ", timezone, " default to UTC.");
      timezone = "UTC";
    } 

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
        if(c != countIndex && c != dateTimeIndex) {
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

      let isThisYear = moment(row[dateTimeIndex]).isSame(moment(), 'year');
      let formatStr = isThisYear ? 'MMM DD' : 'MMM DD, YYYY';
      if (timestampType == 'hour') formatStr = 'MMM DD, HH:mm';

      // moment uses user's current location timezone.
      lines[key].timestamps.push(mt(row[dateTimeIndex]).tz(timezone).format(formatStr));
      
      if (maxScale < row[countIndex]) maxScale = row[countIndex];
    }
    
    return { lines: lines, maxScale: maxScale };
  }

  isResultWithGroupBy(result) {
    return result.headers.length > 2
  }

  render() {
    let result = this.props.queryResult;
    let displayLegend = this.props.hideLegend ? false : this.isResultWithGroupBy(result);

    let countIndex = result.headers.indexOf(HEADER_COUNT);
    if (countIndex == -1) { 
      console.error('No counts to plot as lines.');
      return null;
    }

    let dateIndex = result.headers.indexOf(HEADER_DATE);
    if (dateIndex == -1) { 
      console.error('No dates to plot as lines.');
      return null;
    }
    let lines = [];
    let groups = this.getLinesByGroupsIfExist(result.rows, result.meta.query.gbt, countIndex, dateIndex, result.meta.query.tz);
    for(let key in groups.lines) {
      let line = { title: key, xAxisLabels: groups.lines[key].timestamps, yAxisLabels: groups.lines[key].counts };
      lines.push(line);
    }

    let options = {
      maintainAspectRatio: false,
      responsive: true,
      legend: {
        display: displayLegend,
        labels: {
          filter: function(item, chart) {
            return !!item.text;
          }
        }
      },
      scales: {
        xAxes: [{
          scaleLabel: {
            display: true,
            labelString: 'Date of Occurrence'
          },
        }],
        yAxes: [{
          scaleLabel: {
            display: true,
            labelString: getYAxesStr(result.meta.query.ty)
          },
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

    let mid = plotXAxisLabels.length%2 == 0 ? (plotXAxisLabels.length/2) -1 : (plotXAxisLabels.length-1)/2;

    // TODO: find a better way to do this
    if(!!this.props.verticalLine && mid > 0 ){
      datasets.push({
        fill: false,
        lineTension: 0.1,
        backgroundColor: "rgba(32, 201, 151, 1.0)",
        borderColor: "rgba(128, 128, 128, 1.0)",
        borderCapStyle: 'butt',
        borderWidth: 2,
        borderDash: [8,5],
        borderDashOffset: 0.0,
        pointBorderColor: "rgba(255, 0, 0, 1.0)",
        pointBackgroundColor: '#fff',
        pointBorderWidth: 3,
        pointHoverRadius: 1,
        pointHoverBackgroundColor: "rgba(32, 201, 151, 1.0)",
        pointHoverBorderColor: "rgba(32, 201, 151, 1.0)",
        pointHoverBorderWidth: 3,
        pointRadius: 0,
        pointHitRadius: 5,
        data: [{x:plotXAxisLabels[mid], y:0}, {x:plotXAxisLabels[mid], y:getChartScaleWithSpace(groups.maxScale)}]
      });
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