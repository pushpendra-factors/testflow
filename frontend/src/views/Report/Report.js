import React, { Component } from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import {
    Col,
    Row,
    Card,
    CardHeader,
    CardTitle,
    CardBody,
} from 'reactstrap';

import {readableTimstamp} from '../../util';
import { fetchReport } from '../../actions/reportActions';
import LineChart from '../Query/LineChart';
import BarChart from '../Query/BarChart';
import { PRESENTATION_LINE, PRESENTATION_CARD, PRESENTATION_BAR } from '../Query/common';
import Loading from '../../loading';

const mapStateToProps = store => {
  return {
    currentProjectId: store.projects.currentProjectId,
    report: store.reports.report
  };
}

const mapDispatchToProps = dispatch => {
  return bindActionCreators({ 
      fetchReport
  }, dispatch);
}

const mergeLineQR = function(intervalBeforeQR, intervalQR) {
  let intervalBeforeRows = intervalBeforeQR.rows;
  let mergedQR = {};
  mergedQR.headers = intervalQR.headers;
  mergedQR.rows = intervalBeforeRows.concat(intervalQR.rows);
  return mergedQR;
}

const BarUnit = (props) => {
  let interval =  props.interval;
  let intervalBefore = props.intervalBeforeThat ;

  return (
    <Card className='fapp-report-card'>
      <CardHeader>
        <strong>{props.name}</strong>
      </CardHeader>
      <CardBody>
        <div style={{ height: '400px' }}>
          <BarChart
            queryResultLabel={readableTimstamp(intervalBefore.st) + " - " + readableTimstamp(intervalBefore.et)}
            queryResult={intervalBefore.qr}
            compareWithQueryResultLabel={readableTimstamp(interval.st) + " - " + readableTimstamp(interval.et)}
            compareWithQueryResult={interval.qr}
          />
        </div>
      </CardBody>
    </Card>
  );
}

const LineUnit = (props) => {
    let interval =  props.interval;
    let intervalBefore = props.intervalBeforeThat ;

    let intervalQR = interval.qr;
    let intervalBeforeQR = intervalBefore.qr;

    let mergedQR = mergeLineQR(intervalBeforeQR, intervalQR)

    return (
      <Card className='fapp-report-card'>
        <CardHeader>
          <strong>{props.name}</strong>
        </CardHeader>
        <CardBody>
          <div style={{ height: '450px' }}>
            <LineChart queryResult={mergedQR} verticalLine={true}/>
          </div>
        </CardBody>
      </Card>
    );
}

const CardUnit = (props) => {
  let intervalVal =  props.interval.qr.rows[0][0];
  let intervalBeforeVal = props.intervalBeforeThat.qr.rows[0][0];
  
  let calculatePercentage = !(intervalVal == 0 || intervalBeforeVal == 0);
  let percentChange = 0 ;
  let effect = "";
  if (calculatePercentage) {
    // Todo: Move this as part array of insights from backend.
    percentChange = ((intervalVal-intervalBeforeVal) / intervalBeforeVal) * 100;
    effect = percentChange > 0 ? "Increase in" : "decreased in";
    percentChange = percentChange > 0 ? percentChange : -1 * percentChange;
  }
    
    return (
      <Card className='fapp-report-card'>
        <CardHeader>
          <strong>{props.name}</strong>
        </CardHeader>
        <CardTitle>
          { calculatePercentage ? percentChange.toFixed(2) + "% " + effect + " " + props.name : null }
        </CardTitle>
        <CardBody>
          <div style={{ textAlign: 'center', marginBottom: '30px' }}>
            <div style={{ border: '1px solid #AAA', padding: '20px 30px', display: 'inline-block', textAlign: 'center', marginRight: '60px' }}>
              <div className='fapp-label' style={{ marginBottom: '15px' }} >
                <span> 
                  { readableTimstamp(props.intervalBeforeThat.st) + " - " + readableTimstamp(props.intervalBeforeThat.et) } 
                </span>
                <div style={{ fontSize: '12px', color: '#999' }}>Week before last</div>
              </div>
              <div style={{ fontSize: '40px', marginBottom: '12px' }}>
                <span> { intervalBeforeVal } </span>
              </div>
            </div>

            <div style={{ border: '1px solid #AAA', padding: '20px 30px', display: 'inline-block', textAlign: 'center' }}>
              <div className='fapp-label' style={{ marginBottom: '15px' }} >
                <span> 
                  { readableTimstamp(props.interval.st) + " - " + readableTimstamp(props.interval.et) } 
                </span>
                <div style={{ fontSize: '12px', color: '#999' }}>Last Week</div>
              </div>
              <div style={{ fontSize: '40px', marginBottom: '12px' }}>
                <span> { intervalVal } </span>
              </div>
            </div>
          </div>
        </CardBody>
      </Card>
    )
}

class Report extends Component {
  constructor(props) {
    super(props);
  }

  componentWillMount() {
    // TODO: Make url restful on frontend also
    // projects/:project_id/reports/:report_id
    this.props.fetchReport(this.props.currentProjectId, this.getReportIDToDisplay());
  }

  getReportIDToDisplay() {
    let {id} = this.props.match.params;
    return parseInt(id);
  }
  
  renderReportUnits(report) {
    let reportUnits = [];
    let units = report.contents.dashboardunitid_to_dashboardunitreport;

    for(let id in units){
        if (!units.hasOwnProperty(id)) continue;
        let unit = units[id];
        
        if (unit.p === PRESENTATION_CARD) {
            reportUnits.push(<CardUnit name={unit.t} intervalBeforeThat={unit.r[0]} interval={unit.r[1]} />);
        } else if (unit.p === PRESENTATION_LINE){
            reportUnits.push(<LineUnit name={unit.t} intervalBeforeThat={unit.r[0]} interval={unit.r[1]} />);
        } else if(unit.p === PRESENTATION_BAR){
            reportUnits.push(<BarUnit name={unit.t} intervalBeforeThat={unit.r[0]} interval={unit.r[1]} />);
        }
    }

    return reportUnits;
  }

  renderReport(report) {
    return (
      <div>
        <div style={{ textAlign: 'center' }}>
          <h4 style={{ marginBottom: '0.2rem', color: '#555' }}> { 'Weekly Report - ' + report.dashboard_name } </h4>
          <span className='fapp-text light small'> { readableTimstamp(report.start_time) + " - " + readableTimstamp(report.end_time) } </span>
        </div>
        { this.renderReportUnits(report) }
      </div>
    );
  }

  render() {
    if (!this.props.report) return <Loading />;

    return (
      <div className='fapp-content' style={{ marginLeft: '5rem', marginRight: '5rem', paddingTop: '50px' }}>
        { this.renderReport(this.props.report) }
      </div>
    );
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(Report);