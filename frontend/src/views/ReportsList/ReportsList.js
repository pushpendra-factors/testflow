import React, { Component } from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import { fetchProjectReportsList } from '../../actions/reportActions';
import {readableTimstamp} from '../../util'
import {
    Col,
    Row,
    Card,
    CardHeader,
    CardBody
} from 'reactstrap';

const mapStateToProps = store => {
    return {
      currentProjectId: store.projects.currentProjectId,
      reports: store.reports.reports_list
    };
}
  
  const mapDispatchToProps = dispatch => {
    return bindActionCreators({ 
        fetchProjectReportsList
    }, dispatch);
  }

const ReportRecord = (props) => {
    return (
        <Row style={{ marginBottom: '10px' }} onClick={props.onClick} >
            <Col md={{size:2}}>{props.name}</Col><Col md={{size:3}}>{props.start_time + " - " + props.end_time}</Col>
        </Row>
    )
}

class ReportsList extends Component {
    constructor(props) {
      super(props);
    }
    
    componentWillMount() {
        this.props.fetchProjectReportsList(this.props.currentProjectId);
    }

    renderReportsList(){
        let reportRecords = [];
        let reports = this.props.reports;
        if(!reports || reports.length == 0) {
            return reportRecords;
        }
        reports.map((report) => {
            reportRecords.push(<ReportRecord 
                key = {report.id}
                name = {report.dashboard_name}
                start_time = {readableTimstamp(report.start_time)}
                end_time = {readableTimstamp(report.end_time)}
                onClick = {()=>{ this.props.history.push("/reports/"+report.id) }}
            />)
        });
        return reportRecords;
    }

    render() {  
      return (
        <div>
            <Card className='fapp-card' style={{ marginBottom: '10px' }}>
                <CardHeader style={{ marginBottom: '5px' }}>
                    <strong>Reports</strong>
                </CardHeader>
                <CardBody className='fapp-medium-font'>
                    {this.renderReportsList()}
                </CardBody>
            </Card>
        </div>
      );
    }
}

export default connect(mapStateToProps, mapDispatchToProps)(ReportsList);