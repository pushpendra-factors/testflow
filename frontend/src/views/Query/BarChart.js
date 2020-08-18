import React, { Component } from 'react';
import { Bar } from 'react-chartjs-2';
import { CustomTooltips } from '../../common/custom-tooltips';

import { getChartScaleWithSpace, isSingleCountResult } from '../../util';
import { HEADER_COUNT, getYAxesStr } from './common';

const barBackgroundColors = ['rgba(75,192,192,0.4)', 'rgba(255,99,132,0.2)'];
const barBorderColors = ['rgba(75,192,192,1)', 'rgba(255,99,132,1)'];
const barHoverBackgroundColors = ['rgba(75,192,192,1)', 'rgba(255,99,132,0.4)'];
const barHoverBorderColors = ['rgba(220,220,220,1)', 'rgba(255,99,132,0.4)'];

class BarChart extends Component {
  constructor(props) {
    super(props);
  }

  sortByLabel(labels, counts) {
    let cLabels = [ ...labels ];
    let cCounts = [];
    let labelCountLookup = {}
    for (let i in cLabels) {
      labelCountLookup[cLabels[i]] = counts[i]
    }

    cLabels.sort();
    for (let i in cLabels) {
      cCounts.push(labelCountLookup[cLabels[i]]);
    }

    return { labels: cLabels, counts: cCounts };
  }

  getBarsAndScaleFromResult(result) {
    let bars = {};

    let countIndex = result.headers.indexOf(HEADER_COUNT);
    // Need a count and a group col for bar.
    if (countIndex == -1) { 
      console.error("Invalid query result for bar chart.");
      return null;
    }
    
    let maxScale = 0;
    let data = [], labels = [];
    if (result.headers.length == 2) {
      // Other col apart from count is group col.
      let groupIndex = countIndex == 0 ? 1 : 0;
      for(let i=0; i<Object.keys(result.rows).length; i++) {
        let cols = result.rows[i.toString()];
        if (cols != undefined && cols[countIndex] != undefined) {
          data.push(cols[countIndex]);
          labels.push(cols[groupIndex]);
          if (maxScale < cols[countIndex]) maxScale = cols[countIndex];
        }
      }
      bars.x_label = result.headers[groupIndex];
    } else if (result.headers.length == 1) {
      let col = result.rows["0"];
      data.push(col[countIndex]);
      if (maxScale < col[countIndex]) maxScale = col[countIndex];
      bars.x_label = "";
    } else {
      console.error("Invalid no.of result columns for vertical bar.");
      return null;
    }

    bars.datasets = [{ data: data  }];
    if (isSingleCountResult(result)) {
      // use event name as xAxisLabel when no groups available.
      bars.labels = [result.meta.query.ewp[0].na];
    } else {
      bars.labels = labels;
    }
    
    bars.y_label = "";

    return { bars: bars, maxScale: maxScale };
  }

  getChartOptions(displayLegend, maxScale){
    return {
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
            display: true,
            labelString: getYAxesStr(this.props.queryResult.meta.query.ty)
          },
          display: true,
          ticks: {
            beginAtZero: true,
            max: maxScale
          }
        }],
      },
    };
  }

  getValuesByLabel(bar, datasetIndex=0) {
    let map = {};
    for(let i=0; i<bar.labels.length; i++)
      map[bar.labels[i]] = bar.datasets[datasetIndex].data[i];

    return map;
  }

  createDataByLookup(labels, valueByLabel) {
    let data = [];
    for (let i=0; i<labels.length; i++) {
      let label = labels[i];
      if (valueByLabel[label]) data.push(valueByLabel[label]);
      else data.push(0);
    }

    return data;
  }

  mergeBars(bar1, bar2) {
    let bar1Lookup = this.getValuesByLabel(bar1);
    let bar2Lookup = this.getValuesByLabel(bar2);

    // merge labels. dedupe.
    let labels = bar1.labels;
    for(let i=0; i<bar2.labels.length; i++) {
      if (labels.indexOf(bar2.labels[i]) == -1)
        labels.push(bar2.labels[i]);
    }

    // using bar1 to preserve style attributes 
    // which is common for both.
    bar1.labels = labels;
    bar1.datasets[0].data = this.createDataByLookup(labels, bar1Lookup);
    bar1.datasets.push({ data: this.createDataByLookup(labels, bar2Lookup) });

    return bar1;
  }

  render() {
    var barsAndScale = this.getBarsAndScaleFromResult(this.props.queryResult);
    if (!barsAndScale) return null;

    let displayLegend = this.props.legend == false ? false : true;
    var chartData = barsAndScale.bars;

    var bar = {
      labels: chartData.labels,
      datasets: chartData.datasets,
    };

    let maxScale = getChartScaleWithSpace(barsAndScale.maxScale);

    if (this.props.compareWithQueryResult ){
      var barsAndScaleComp = this.getBarsAndScaleFromResult(this.props.compareWithQueryResult);
      var barComp = barsAndScaleComp.bars;
      // overrides with merged values.
      bar = this.mergeBars(bar, barComp);

      let maxScale2 = getChartScaleWithSpace(barsAndScaleComp.maxScale);
      if (maxScale < maxScale2) maxScale = maxScale2;

      bar.datasets[0].label = this.props.queryResultLabel;
      bar.datasets[1].label = this.props.compareWithQueryResultLabel;
    }

    var chartOptions = this.getChartOptions(displayLegend, maxScale);

    // Styling.
    for (var i = 0; i < bar.datasets.length; i++) {
      bar.datasets[i].backgroundColor = barBackgroundColors[i % barBackgroundColors.length];
      bar.datasets[i].borderColor = barBorderColors[i % barBorderColors.length];
      bar.datasets[i].borderWidth = 1;
      bar.datasets[i].hoverBackgroundColor = barHoverBackgroundColors[i % barHoverBackgroundColors.length];
      bar.datasets[i].hoverBorderColor = barHoverBorderColors[i % barHoverBorderColors.length]; 
    }

    if (chartData.x_label != '') {
      chartOptions.scales.xAxes[0].scaleLabel.display = true;
      chartOptions.scales.xAxes[0].scaleLabel.labelString = "Property: " + chartData.x_label;
    }
    
    return <Bar data={bar} options={chartOptions} /> 
  }

}

export default BarChart;