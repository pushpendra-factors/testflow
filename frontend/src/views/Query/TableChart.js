import React, { Component } from 'react';
import { Table } from 'reactstrap';
import { isSingleCountResult } from '../../util';

class TableChart extends Component {
  constructor(props) {
    super(props);
    this.state = {};
  }

  tableHeader() {
    if (this.props.noHeader) return null; 

    let result = this.props.queryResult;
    let cardable = this.props.card && isSingleCountResult(result);
    let headers = result.headers.map((h, i) => {
      let style = cardable ? { border: 'none', fontSize: '40px', padding: '0' } : null;
      return <th style={style} key={'header_'+i}>{ h }</th> 
    });

    return (
      <thead>
        <tr> { headers } </tr>
      </thead>
    );
  }

  render() {
    let result = this.props.queryResult;
    let rows = [];

    let rowKeys = Object.keys(result.rows);
    let cardable = this.props.card && isSingleCountResult(result);

    // card.
    if (cardable) {
      return (
        <Table className='animated fadeIn' style={{fontSize: '45px', textAlign: 'center', border: 'none', marginTop: '8%' }} >
          { this.tableHeader() }
          <tbody> { result.rows[rowKeys[0]][0] } </tbody>
        </Table>
      )
    }

    for(let i=0; i<rowKeys.length; i++) {
      let cols = result.rows[i.toString()];
      if (cols != undefined) {
        let tds = cols.map((c) => { return <td> { c } </td> });
        rows.push(<tr> { tds } </tr>);
      }
    }

    return (
      <Table className='fapp-table animated fadeIn' >
        { this.tableHeader() }
        <tbody>
          { rows }
        </tbody>
      </Table>
    );
  } 
}

export default TableChart;