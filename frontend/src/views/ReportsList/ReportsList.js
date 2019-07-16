import React, { Component } from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import {
  Col,
  Row,
  Card,
  CardHeader,
  CardBody
} from 'reactstrap';

import { fetchProjectReportsList } from '../../actions/reportActions';
import { readableTimstamp } from '../../util';
import Loading from '../../loading';

const mapStateToProps = store => {
  return {
    currentProjectId: store.projects.currentProjectId,
    reports: store.reports.reportsList
  };
}
  
  const mapDispatchToProps = dispatch => {
    return bindActionCreators({ 
        fetchProjectReportsList
    }, dispatch);
  }

const ReportRecord = (props) => {
  return (
    <Row style={{ marginBottom: '10px' }}>
        <Col md={2} className='fapp-clickable' onClick={ props.onClick }>{props.name}</Col>
        <Col md={2} > { props.start_time + " - " + props.end_time } </Col>
    </Row>
  )
}

class ReportsList extends Component {
  constructor(props) {
    super(props);

    this.state = {
      loading: true
    }
  }
  
  componentWillMount() {
      this.props.fetchProjectReportsList(this.props.currentProjectId)
        .then(() => { this.setState({ loading: false }) });
  }

  renderReportsList() {
    let reportRecords = [];
    let reports = this.props.reports;

    if (!reports || reports.length == 0) {
        return reportRecords;
    }

    reportRecords = reports.map((report) => (
      <ReportRecord 
        key = {report.id}
        name = {report.dashboard_name}
        start_time = {readableTimstamp(report.start_time)}
        end_time = {readableTimstamp(report.end_time)}
        onClick = {()=>{ this.props.history.push("/reports/"+report.id) }}
      />
    ));
    return reportRecords;
  }

  render() {
    if (this.state.loading) return <Loading />;

    return (
      <div className='fapp-content' style={{ marginLeft: '2rem', marginRight: '2rem', paddingTop: '30px' }}>
        <Card className='fapp-card' style={{ marginBottom: '10px' }}>
          <CardHeader style={{ marginBottom: '5px' }}>
            <strong>Reports</strong>
          </CardHeader>
          <CardBody style={{ fontSize: '0.95em' }}>
            <Row style={{ marginBottom: '10px' }} >
              <Col md={2} className='fapp-label light'>Name</Col>
              <Col md={2} className='fapp-label light'>Period</Col>
            </Row>
            { this.renderReportsList() }
          </CardBody>
        </Card>
      </div>
    );
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(ReportsList);