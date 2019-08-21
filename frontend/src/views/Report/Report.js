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
import FunnelChart from '../Query/FunnelChart';
import { PRESENTATION_LINE, PRESENTATION_CARD, 
  PRESENTATION_BAR, PRESENTATION_FUNNEL } from '../Query/common';
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
  mergedQR.meta = intervalQR.meta;
  return mergedQR;
}

const renderExplanations = function(explanations) {
  if (!explanations || explanations.length == 0)
    return null;

  return explanations.map((exp) => <div style={{ marginBottom: '0.4rem' }}>{ exp }</div>);
}

const BarUnit = (props) => {
  let interval =  props.interval;
  let intervalBefore = props.intervalBeforeThat ;

  return (
    <Card className='fapp-report-card'>
      <CardHeader>
        <strong>{props.name}</strong>
      </CardHeader>
      <CardTitle> { renderExplanations(props.explanations) } </CardTitle>
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
        <CardTitle> { renderExplanations(props.explanations) } </CardTitle>
        <CardBody>
          <div style={{ height: '400px' }}>
            <LineChart queryResult={mergedQR} verticalLine={true}/>
          </div>
        </CardBody>
      </Card>
    );
}

const CardUnit = (props) => {
  let intervalVal =  props.interval.qr.rows[0][0];
  let intervalBeforeVal = props.intervalBeforeThat.qr.rows[0][0];
    
  return (
    <Card className='fapp-report-card'>
      <CardHeader>
        <strong>{props.name}</strong>
      </CardHeader>
      <CardTitle> { renderExplanations(props.explanations) } </CardTitle>
      <CardBody>
        <div style={{ textAlign: 'center', marginBottom: '30px' }}>
          <div style={{ border: '1px solid #AAA', padding: '20px 30px', display: 'inline-block', textAlign: 'center', marginRight: '60px' }}>
            <div style={{ marginBottom: '15px' }} >
              <span> 
                { readableTimstamp(props.intervalBeforeThat.st) + " - " + readableTimstamp(props.intervalBeforeThat.et) } 
              </span>
            </div>
            <div style={{ fontSize: '40px', marginBottom: '12px' }}>
              <span> { intervalBeforeVal } </span>
            </div>
          </div>

          <div style={{ border: '1px solid #AAA', padding: '20px 30px', display: 'inline-block', textAlign: 'center' }}>
            <div style={{ marginBottom: '15px' }} >
              <span> 
                { readableTimstamp(props.interval.st) + " - " + readableTimstamp(props.interval.et) } 
              </span>
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

const FunnelUnit = (props) => {
  let curResult =  props.interval.qr;
  let prevResult = props.intervalBeforeThat.qr;

  return (
    <Card className='fapp-report-card'>
      <CardHeader>
        <strong>{props.name}</strong>
      </CardHeader>
      <CardTitle> { renderExplanations(props.explanations) } </CardTitle>
      <CardBody style={{ marginBottom: '30px' }}>
        <Row>
          <Col md={6}>
            <FunnelChart queryResult={prevResult} noMargin small /> 
          </Col>
          <Col md={6}>
            <FunnelChart queryResult={curResult} noMargin small />
          </Col>
        </Row>
      </CardBody>
    </Card>
  );
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
    let units = report.units;

    for(let i=0; i<units.length; i++){
        let unit = units[i];

        if (unit.p === PRESENTATION_CARD) {
          reportUnits.push(<CardUnit name={unit.t} intervalBeforeThat={unit.r[0]} interval={unit.r[1]} explanations={unit.e} />);
        } else if (unit.p === PRESENTATION_LINE){
          reportUnits.push(<LineUnit name={unit.t} intervalBeforeThat={unit.r[0]} interval={unit.r[1]} explanations={unit.e} />);
        } else if(unit.p === PRESENTATION_BAR){
          reportUnits.push(<BarUnit name={unit.t} intervalBeforeThat={unit.r[0]} interval={unit.r[1]} explanations={unit.e} />);
        } else if(unit.p === PRESENTATION_FUNNEL) {
          reportUnits.push(<FunnelUnit name={unit.t} intervalBeforeThat={unit.r[0]} interval={unit.r[1]} explanations={unit.e} />);
        }
    }

    return reportUnits;
  }

  getTitle(report) {
    let typ = '';
    if (report.type == 'w') typ = 'Weekly';
    else if (report.type == 'm') typ = 'Monthly';
    
    return typ + ' Report - ' + report.dashboard_name; 
  }

  renderReport(report) {
    return (
      <div className='fapp-gray'>
        <div style={{ textAlign: 'center' }}>
          <h4 style={{ marginBottom: '0.2rem', color: '#555' }}> { this.getTitle(report) } </h4>
          <span className='fapp-text light small'> { readableTimstamp(report.start_time) + " - " + readableTimstamp(report.end_time) } </span>
        </div>
        { this.renderReportUnits(report) }
      </div>
    );
  }

  render() {
    if (!this.props.report) return <Loading />;

    return (
      <div className='fapp-content' style={{ marginLeft: '7rem', marginRight: '7rem', paddingTop: '50px' }}>
        { this.renderReport(this.props.report) }
      </div>
    );
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(Report);