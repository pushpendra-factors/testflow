import React, { Component } from 'react';
import { HorizontalBar } from 'react-chartjs-2';
import { Table } from 'reactstrap';

import { trimQuotes, firstToUpperCase, getChartScaleWithSpace } from '../../util'
import NoContent from '../../common/NoContent';

const barBackgroundColors = ['rgba(75,192,192,0.4)', 'rgba(255,99,132,0.2)'];
const barBorderColors = ['rgba(75,192,192,1)', 'rgba(255,99,132,1)'];
const barHoverBackgroundColors = ['rgba(75,192,192,1)', 'rgba(255,99,132,0.4)'];
const barHoverBorderColors = ['rgba(220,220,220,1)', 'rgba(255,99,132,0.4)'];

class TableBarChart extends Component {
  constructor(props) {
    super(props);
  }

  getBarChart(bars, maxXScale, legend=false) {
    var chartData = bars;
    let displayLegend = legend == false ? false : true;

    var chartOptions = {
      legend: {
        display: displayLegend
      },
      maintainAspectRatio: false,
      scales: {
        xAxes: [{
          ticks: {
            beginAtZero: true,
            max: getChartScaleWithSpace(maxXScale)
          },
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

    return <HorizontalBar data={bar} options={chartOptions} />;
  }

  getMaxXScaleFromGroups(groups) {
    let keys = Object.keys(groups);
    let maxAcrossGroups = 0;
    for(let k=0; k<keys.length; k++) {
      let g = groups[keys[k]];
      for(let di = 0; di < g.datasets.length; di++) {
        if(maxAcrossGroups < g.datasets[di])
          maxAcrossGroups = g.datasets[di];
      }
    }
    
    return maxAcrossGroups;
  }

  render() {
    if (this.props.data.rows && this.props.data.rows.length == 0)
      return <NoContent center msg='No Result' />;

    // Temp fix for chart breakage on query change after render.
    if (this.props.data.headers != undefined &&
        this.props.data.rows != undefined &&
        this.props.data.headers.length != this.props.data.rows[0].length)
          return <NoContent center msg='No Result' />;

    let data = this.props.data;
    let headers = data.headers.map((h, i) => { return <th key={'header_'+i}>{ h }</th> });
    const HEADER_COUNT= "count";

    let countIndex = data.headers.indexOf(HEADER_COUNT);
    if (countIndex == -1 || data.headers.length < 2) { 
      throw new Error('Invalid query result for bar chart.');
    }

    let lastIndex = data.headers.length - 1;
    if (countIndex != lastIndex) {
      throw new Error('Count is not the last index');
    }
    let labelIndex = countIndex - 1;
    
    let heads = headers.slice(0, headers.length - 2);
    heads.push(<th>{ firstToUpperCase(data.headers[countIndex]) + " by Property - " + data.headers[labelIndex] }</th>);
    
    let groups = {};
    for(let i=0; i<Object.keys(data.rows).length; i++) {
      let cols = [...data.rows[i.toString()]];
      if (cols != undefined) {
        let encKey = cols.slice(0, cols.length - 2).join('');
        let popped = cols.splice(labelIndex, 2);
        let g = groups[encKey];
        if (groups[encKey] == undefined) {
          // init.
          groups[encKey] = { 
            row: cols,
            labels: [trimQuotes(popped[0])], 
            datasets: [popped[1]],
          };
        } else {
          // add to key.
          g.row = cols; // combination of row + labels = actual row.
          g.labels.push(trimQuotes(popped[0]));
          g.datasets.push(popped[1]);
        }
      }
    }

    let rows = [];
    let maxXScale = this.getMaxXScaleFromGroups(groups);
    let keys = Object.keys(groups);
    for(let k=0; k<keys.length; k++) {
      let g = groups[keys[k]];
      let tds = g.row.map((r) => { return <td>{trimQuotes(r)}</td> });
      let gtd = <td style={{height: (g.datasets.length * 22.5)}}>{ this.getBarChart({ labels: g.labels, datasets:[{data: g.datasets}] }, maxXScale, false) } </td>;
      tds.push(gtd)
      rows.push(<tr> {tds} </tr>)
    }
    
    return (
      <Table className='fapp-table'> 
        <thead>
          <tr> { heads } </tr>
        </thead>
        <tbody>
          { rows }
        </tbody>
      </Table>
    );
  }
}

export default TableBarChart;