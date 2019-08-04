import React, { Component } from 'react';
import { Table } from 'reactstrap';

import Funnel from './Funnel';

class FunnelChart extends Component {
  constructor(props) {
    super(props);
  }

  getDisplayName(name, count) {
    return name+ " (" + count + ")" ;
  }

  buildFunnelsFromResultRows(rows, stepNames, stepsIndexes, conversionIndexes) {
    let funnels = [];
    for (let ri = 0; ri < rows.length; ri++) {
      let funnelData = [];

      for (let i=0; i<stepsIndexes.length; i++) {
        let data = null;
        if (i == 0) data = [rows[ri][stepsIndexes[0]], 0];
        else {
          if (rows[ri][stepsIndexes[i-1]] == 0) {
            data = [0, 1]
          } else {
            data = [rows[0][stepsIndexes[i]], [rows[ri][stepsIndexes[i-1]] - rows[ri][stepsIndexes[i]]]]
          }
        }
        let stepName = this.getDisplayName(stepNames[i], rows[ri][stepsIndexes[i]])

        let comp = {};
        comp.conversion_percent = rows[ri][conversionIndexes[i]];
        comp.data = data;
        comp.event = stepName;

        funnelData.push(comp);
      }

      let totalConversionIndex = conversionIndexes[conversionIndexes.length - 1];
      funnels.push(<Funnel data={{ funnels: funnelData, totalConversion: rows[ri][totalConversionIndex] }} />);
    }

    return funnels;
  }

  render() {
    let result = this.props.queryResult;
    // get funnel step names from result meta.
    let stepNames = result.meta.query.ewp.map((e) => (e.na));

    let stepsIndexes = [];
    let conversionIndexes = [];
    let groupIndexes = [];
    let groupHeaders = [];
    
    for (let i=0; i<result.headers.length; i++) {
      if (result.headers[i].indexOf('step_') == 0)
        stepsIndexes.push(i);
      else if (result.headers[i].indexOf('conversion_') == 0) {
        conversionIndexes.push(i);
      }
      else {
        groupIndexes.push(i);
        groupHeaders.push(result.headers[i]);
      }
    }

    let rows = result.rows;
    let funnels = this.buildFunnelsFromResultRows(rows, stepNames, stepsIndexes, conversionIndexes);

    let showGroupsTable = groupIndexes.length > 0;
    let groupRows = [];
    if (showGroupsTable) {
      let conversionsHeader = "conversions";
      groupHeaders.push(conversionsHeader);
      
      // excluding main funnel which is index 0;
      for(let i=1; i<result.rows.length; i++) {
        let row = [];
        // adds group values to row.
        for (let r=0; r<groupIndexes.length; r++) {
          row.push(result.rows[i][groupIndexes[r]]);
        }
        row.push(funnels[i]);
        groupRows.push(row);
      }
    }

    let tableHeaders = [];
    for (let hi=0; hi<groupHeaders.length; hi++) {
      tableHeaders.push(<th>{groupHeaders[hi]}</th>)
    }
    
    let tableRows = [];
    for (let ri=0; ri<groupRows.length; ri++) {
      let tableCols = [];
      let rowLength = groupRows[ri].length
      for (let ci=0; ci<rowLength; ci++) {
        let style = null;
        if (ci == rowLength-1) style = { padding: '30px' }; // conversion col.
        else style = { paddingTop: '30px' }; // group cols.

        tableCols.push(<td style={style}>{groupRows[ri][ci]}</td>);
      }
      tableRows.push(<tr>{tableCols}</tr>)
    }

    
    let present = [];
    // main funnel.
    present.push(<div style={{ marginTop: !this.props.noMargin ? '30px': null }}>{ funnels[0] }</div>);
    // group based funnels.
    if (showGroupsTable) {
      present.push(
        <div style={{ marginTop: '55px' }}>
          <Table className="fapp-table">
            <thead>{ tableHeaders }</thead>
            <tbody>{ tableRows }</tbody>
          </Table>
        </div>
      );
    }
    
    return present;
  }
}

export default FunnelChart;