import React, { Component } from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import {readableTimstamp} from '../../util';
import { fetchReport } from '../../actions/reportActions';
import {
    Col,
    Row,
    Card,
    CardHeader,
    CardBody
} from 'reactstrap';
import LineChart from '../Query/LineChart';
import BarChart from '../Query/BarChart';
import {  PRESENTATION_LINE, PRESENTATION_CARD, PRESENTATION_BAR } from '../Query/common';
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

function mergeLineQR(intervalBeforeQR, intervalQR){
    
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
        <Card className='fapp-card' style={{ marginBottom: '10px' }}>
            <CardHeader style={{ marginBottom: '5px' }}>
                <strong>{props.name}</strong>
            </CardHeader>
            <CardBody className='fapp-medium-font'>
                <Row>
                    <BarChart
                        queryResultLabel={readableTimstamp(intervalBefore.st) + " - " + readableTimstamp(intervalBefore.et)}
                        queryResult={intervalBefore.qr}
                        compareWithQueryResultLabel={readableTimstamp(interval.st) + " - " + readableTimstamp(interval.et)}
                        compareWithQueryResult={interval.qr}
                    />
                </Row>
            </CardBody>
        </Card>
    )
}

const LineUnit = (props) => {
    let interval =  props.interval;
    let intervalBefore = props.intervalBeforeThat ;

    let intervalQR = interval.qr;
    let intervalBeforeQR = intervalBefore.qr;

    let mergedQR = mergeLineQR(intervalBeforeQR, intervalQR)

    return (
        <Card className='fapp-card' style={{ marginBottom: '10px' }}>
            <CardHeader style={{ marginBottom: '5px' }}>
                <strong>{props.name}</strong>
            </CardHeader>
            <CardBody className='fapp-medium-font'>
                <Row>
                    <LineChart queryResult={mergedQR} verticalLine={true}/>
                </Row>
            </CardBody>
        </Card>
    )
}

const CardUnit = (props) => {

    let intervalVal =  props.interval.qr.rows[0][0];
    let intervalBeforeVal = props.intervalBeforeThat.qr.rows[0][0];
    
    let calculatePercentage = !(intervalVal == 0 || intervalBeforeVal == 0);
    let percentChange = 0 ;
    let effect = "";
    if (calculatePercentage) {
        percentChange = ((intervalVal-intervalBeforeVal)/intervalBeforeVal)* 100;
        effect = percentChange > 0 ? "Increase in" : "decreased in";
        percentChange = percentChange > 0 ? percentChange : -1 * percentChange;
    }
    

    return (
        <Card className='fapp-card' style={{ marginBottom: '10px' }}>
            <CardHeader style={{ marginBottom: '5px' }}>
                <strong>{props.name}</strong>
            </CardHeader>
            <CardBody className='fapp-medium-font'>
                <Row>
                    <Col md={{size:3}}>{readableTimstamp(props.intervalBeforeThat.st) + " - " + readableTimstamp(props.intervalBeforeThat.et)}</Col>
                    <Col md={{size:3}}>{readableTimstamp(props.interval.st) + " - " + readableTimstamp(props.interval.et)}</Col>
                </Row>
                <Row>
                    <Col md={{size:3}}>{intervalBeforeVal}</Col>
                    <Col md={{size:3}}>{intervalVal}</Col>
                </Row>
                {
                    calculatePercentage && 
                    <Row>
                        <Col> {percentChange + "% " + effect + " " + props.name}</Col>
                    </Row>
                }
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

    getReportIDToDisplay(){
        let {id} = this.props.match.params;
        return parseInt(id);
    }
    
    renderReportUnits(report){
        let reportUnits = [];
        let units = report.contents.dashboardunitid_to_dashboardunitreport;
        for(let id in units){
            if(!units.hasOwnProperty(id)){
                continue
            }
            let unit = units[id];
            
            if (unit.p === PRESENTATION_CARD){
                reportUnits.push(<CardUnit name={unit.t} intervalBeforeThat={unit.r[0]} interval={unit.r[1]} />);
            }else if (unit.p === PRESENTATION_LINE){
                reportUnits.push(<LineUnit name={unit.t} intervalBeforeThat={unit.r[0]} interval={unit.r[1]} />);
            }else if(unit.p === PRESENTATION_BAR){
                reportUnits.push(<BarUnit name={unit.t} intervalBeforeThat={unit.r[0]} interval={unit.r[1]} />);
            }
        }
        return reportUnits;
    }

    renderReport(report){
        return (
            <Card className='fapp-card' style={{ marginBottom: '10px', marginTop: '10px' }}>
                <CardHeader style={{ marginBottom: '5px' }}>
                    <strong>{report.dashboard_name + " " + readableTimstamp(report.start_time) + " - " + readableTimstamp(report.end_time)}</strong>
                </CardHeader>
                <CardBody className='fapp-medium-font'>
                    {this.renderReportUnits(report)}
                </CardBody>
            </Card>
        );
    }

    render() {

        if(!this.props.report){
            return (<Loading />);
        }

        return (
            <div>
                {this.renderReport(this.props.report)}
            </div>
        );
    }
}

export default connect(mapStateToProps, mapDispatchToProps)(Report);