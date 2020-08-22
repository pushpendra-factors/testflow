import React, { Component } from 'react';
import { Table, Button, Input } from 'reactstrap';
import { isSingleCountResult, getReadableValue } from '../../util';
import {BootstrapTable, TableHeaderColumn} from 'react-bootstrap-table';
import 'react-bootstrap-table/dist/react-bootstrap-table.min.css';

class TableChart extends Component {
  constructor(props) {
    super(props);
    this.state = {
      searchValue: "",
      search: false,
      dataRows: []
    };
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
    let result = {...this.props.queryResult};
    if (this.state.search) {
      result.rows = [...this.state.dataRows]
    }
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
  onChange = (e) => {
    this.setState({
      searchValue: e.target.value
    })
  }
  search = () => {
    if(this.state.searchValue !== "") {
      let searchValue = this.state.searchValue.toLowerCase()
      let result = {...this.props.queryResult}
      let rows = result.rows.filter(row=> {
        let newRow = [...row]
        let found= false
        for(let i=0;i<newRow.length;i++) {
          if(row[i].toString().toLowerCase().includes(searchValue, 0)) {
            found =true
            break;
          }
        }
        if( found) return newRow
      })
      this.setState({
        dataRows: rows,
        search: true
      })
    }
  }
  renderSearchButtons = () => {
    return (
    <div className="d-flex align-items-center mb-1">
      <Input
          type="text"
          onChange={this.onChange}
          placeholder="Enter search value"
          value={this.state.searchValue}
          className="mx-1"
          style={{ border: "1px solid #ddd", color: "#444444", width: "30%" }}
        />
        <Button 
          outline color="primary"
          onClick = {this.search}
        >Search
        </Button>
        <Button 
          className="ml-1"
          outline color="success"
          onClick = {()=> this.setState({search: false})}
        >Show Original
        </Button>
    </div>
    )
  }

  render() {
    let result = {...this.props.queryResult};
    if(this.state.search) {
      result.rows = this.state.dataRows
    }
    let rows = [];

    let rowKeys = Object.keys(result.rows);
    let cardable = this.props.card && isSingleCountResult(result);

    let sortable = this.props.sort;
    // card.
    if (cardable) {
      return (
        <Table className='animated fadeIn' style={{fontSize: '40px', textAlign: 'center', border: 'none', marginTop: '10px' }} >
          { this.tableHeader() }
          <tbody> <span style={this.getCountStyleByProps()}> { getReadableValue(result.rows[rowKeys[0]][0]) } </span> </tbody>
        </Table>
      )
    }

    if (sortable){
      return (
        <div>
          {this.renderSearchButtons()}
        <BootstrapTable bodyStyle={{paddingBottom:"4px"}} containerStyle={{paddingBottom:"-2px"}} bordered={false} trStyle={{overflowWrap: 'break-word'}} containerClass='fapp-table animated fadeIn' data={this.getData()} options={{sortIndicator:true}} version="4">
          {this.tableHeader()}
        </BootstrapTable>
      </div>
      )
    }

    for(let i=0; i<rowKeys.length; i++) {
      let cols = result.rows[i.toString()];
      if (cols != undefined) {
        let tds = cols.map((c, i) => {
          // Remove max width to allow larger col size for upto given initial no.of cols.
          let maxWidth = (this.props.bigWidthUptoCols && i < this.props.bigWidthUptoCols) ? null : '40px';
          return <td style={{ maxWidth: maxWidth, overflowWrap: 'break-word' }}> { getReadableValue(c) } </td>
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
      <div>
        {this.renderSearchButtons()}
        <Table className='fapp-table animated fadeIn' >
          { this.tableHeader() }
          <tbody>
            { rows }
          </tbody>
        </Table>
      </div>
    );
  } 
}

export default TableChart;