import React, { Component } from 'react';
import { Table } from 'reactstrap';
import { isSingleCountResult } from '../../util';
import {BootstrapTable, TableHeaderColumn} from 'react-bootstrap-table';
import 'react-bootstrap-table/dist/react-bootstrap-table.min.css';

class TableChart extends Component {
  constructor(props) {
    super(props);
    this.state = {};
  }

  tableHeader() {
    if (this.props.noHeader) return null; 

    let result = this.props.queryResult;
    let cardable = this.props.card && isSingleCountResult(result);
    let sortable = this.props.sort
    if (this.props.compareWithQueryResult) {
      // For reports. Tables with comparision.
      var headerCols = [];
      for (var i = 0; i < result.headers.length - 1; i++) {
        headerCols.push(
          <th>{result.headers[i]}</th>
        )
      }
      headerCols.push(
        <th>{result.headers[result.headers.length - 1] + ' for ' + this.props.queryResultLabel }</th>
      )
      headerCols.push(
        <th>{result.headers[result.headers.length - 1] + ' for ' + this.props.compareWithQueryResultLabel }</th>
      )
      return (
        <thead>
          {headerCols}
        </thead>
      );
    } else if(sortable){
      let thStyle = {overflowWrap: 'break-word', whiteSpace:'normal', paddingBottom:"0px", paddingTop:"0px", border: "none"}
      let tdStyle = {overflow:"break-word", whiteSpace:"normal"}
      let headers = result.headers.map((h, i) => {
      let width = (this.props.bigWidthUptoCols && i < this.props.bigWidthUptoCols) ? "280px" : "120px";
      return  (
      <TableHeaderColumn dataSort={i!=0} width={width} tdStyle={tdStyle} thStyle={thStyle} dataAlign="center" isKey={i==0} dataField={"header_"+i}  >
        <div>
          {h}
        </div>
      </TableHeaderColumn>
      )});
      return headers;
    }
    else {
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
  }
  getData(){
    let rows=[];
    let result = this.props.queryResult;
    for (let i =0; i< result.rows.length; i++){
      let cols= result.rows[i].reduce((p, c, i)=>{
        return {...p, ["header_"+i]:c}
      },{});
      rows.push(cols);
    }
    return rows
  }

  getCountStyleByProps() {
    if (!this.props.bordered) return null;
    
    let style = {};
    style.padding = '60px 100px';
    style.border = '1px solid #AAA';
    style.borderRadius = '5px';

    return style;
  }

  render() {
    let result = this.props.queryResult;
    let rows = [];

    let rowKeys = Object.keys(result.rows);
    let cardable = this.props.card && isSingleCountResult(result);

    let sortable = this.props.sort;
    // card.
    if (cardable) {
      return (
        <Table className='animated fadeIn' style={{fontSize: '40px', textAlign: 'center', border: 'none', marginTop: '10px' }} >
          { this.tableHeader() }
          <tbody> <span style={this.getCountStyleByProps()}> { result.rows[rowKeys[0]][0] } </span> </tbody>
        </Table>
      )
    }
    if (sortable){
      return (
      <BootstrapTable bodyStyle={{paddingBottom:"4px"}} containerStyle={{paddingBottom:"-2px"}} bordered={false} trStyle={{overflowWrap: 'break-word'}} containerClass='fapp-table animated fadeIn' data={this.getData()} options={{sortIndicator:true}} version="4">
        {this.tableHeader()}
      </BootstrapTable>
      )
    }

    for(let i=0; i<rowKeys.length; i++) {
      let cols = result.rows[i.toString()];
      if (cols != undefined) {
        let tds = cols.map((c, i) => {
          // Remove max width to allow larger col size for upto given initial no.of cols.
          let maxWidth = (this.props.bigWidthUptoCols && i < this.props.bigWidthUptoCols) ? null : '40px';
          return <td style={{ maxWidth: maxWidth, overflowWrap: 'break-word' }}> { c } </td> 
        });

        if (this.props.compareWithQueryResult) {
          let compareValue = 0;
          if (this.props.compareWithQueryResult.rows.length > 0) {
            let compareCols = this.props.compareWithQueryResult.rows[i.toString()];
            if (compareCols) compareValue = compareCols[compareCols.length - 1];
          }

          // Remove max width to allow larger col size for upto given initial no.of cols.
          let maxWidth = (this.props.bigWidthUptoCols && i < this.props.bigWidthUptoCols) ? null : '40px';
          tds.push(
            <td style={{ maxWidth: maxWidth, overflowWrap: 'break-word' }}> { compareValue } </td>
          );
        }
        
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