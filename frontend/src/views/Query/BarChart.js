import React, { Component } from 'react';
import { Bar } from 'react-chartjs-2';
import { CustomTooltips } from '@coreui/coreui-plugin-chartjs-custom-tooltips';

import { getChartScaleWithSpace, isSingleCountResult } from '../../util';
import { HEADER_COUNT } from './common';

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
      throw new Error('Invalid query result for bar chart.');
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
      throw new Error("Invalid no.of result columns for vertical bar.");
    }

    if (isSingleCountResult(result)) {
      bars.datasets = [{ data: data  }];
      bars.labels = labels;
    } else {
      let labelsCounts = this.sortByLabel(labels, data);
      bars.datasets = [{ data: labelsCounts.counts  }];
      bars.labels = labelsCounts.labels;
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
            display: false,
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

  render() {
    var barsAndScale = this.getBarsAndScaleFromResult(this.props.queryResult);
    let displayLegend = this.props.legend == false ? false : true;
    var chartData = barsAndScale.bars;

    var bar = {
      labels: chartData.labels,
      datasets: chartData.datasets,
    };

    let maxScale = getChartScaleWithSpace(barsAndScale.maxScale);

    if (this.props.compareWithQueryResult ){
      var barsAndScaleComp = this.getBarsAndScaleFromResult(this.props.compareWithQueryResult);
      var chartDataComp = barsAndScaleComp.bars;
      bar.labels = bar.labels.concat(chartDataComp.labels);
      bar.datasets = bar.datasets.concat(chartDataComp.datasets);

      let maxScale2 = getChartScaleWithSpace(barsAndScaleComp.maxScale);
      if(maxScale < maxScale2){
        maxScale = maxScale2;
      }
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