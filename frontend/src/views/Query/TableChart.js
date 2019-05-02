import React from 'react';
import { Table } from 'reactstrap'; 

const TableChart = (props) => {
  let result = props.queryResult;
  let headers = result.headers.map((h, i) => { return <th key={'header_'+i}>{ h }</th> });
  let rows = [];

  for(let i=0; i<Object.keys(result.rows).length; i++) {
    let cols = result.rows[i.toString()];
    if (cols != undefined) {
      let tds = cols.map((c) => { return <td> { c } </td> });
      rows.push(<tr>{tds}</tr>);
    }
  }

  return (
    <Table className='fapp-table animated fadeIn'> 
      <thead>
        <tr> { headers } </tr>
      </thead>
      <tbody>
        { rows }
      </tbody>
    </Table>
  );    
}

export default TableChart;