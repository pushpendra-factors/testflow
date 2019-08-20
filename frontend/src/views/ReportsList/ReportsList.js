import React, { Component } from 'react';
import { connect } from 'react-redux';
import moment from 'moment';
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
import NoContent from '../../common/NoContent';

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

  getReadableType(typ) {
    if (typ == 'w') return 'Weekly';
    else if (typ == 'm') return 'Monthly';
    return "";
  }

  getTypeName(type) {
    if (type == "w") return "Weekly Report";
    if (type == "m") return "Monthly Report";
    return "";
  }

  getReportsByName() {
    let reportsByName = {};
    
    let reports = this.props.reports;
    for (let i=0; i<reports.length; i++) {
      if (!reportsByName[reports[i].dashboard_name])
        reportsByName[reports[i].dashboard_name] = [];

      let period = reports[i].type == "m" ? moment.unix(reports[i].start_time).utc().format('MMMM, YYYY') : 
        (readableTimstamp(reports[i].start_time) + " - "  + readableTimstamp(reports[i].end_time));
      reportsByName[reports[i].dashboard_name].push({
        id: reports[i].id,
        typeName: this.getTypeName(reports[i].type),
        period: period,
      })
    }

    return reportsByName
  }

  renderList(reports) {
    let list = [];
    
    for(let i=0; i<reports.length; i++) {
      list.push(
        <Row style={{ marginBottom: '5px' }} >
          <Col md={2} className="fapp-clickable" style={{ cursor: "pointer" }} onClick={() => { this.props.history.push("/reports/"+reports[i].id) }}>
            { reports[i].typeName }
          </Col>
          <Col md={3}>{ reports[i].period  }</Col>
        </Row>
      );
    }

    return list;
  }

  renderListByDashboard(reports) {
    let dashboards = [];

    let names = Object.keys(reports);
    for (let i=0; i<names.length; i++) {
      let name = names[i];

      dashboards.push(
        <Card className='fapp-card secondary-list'>
          <CardHeader style={{ marginBottom: '5px' }}>
            <strong> { "Report - " + name + " (" + reports[name].length + ")" } </strong>
          </CardHeader>
          <CardBody>
            <Row style={{ marginBottom: '10px' }} >
              <Col md={2} className='fapp-label light'>Type</Col>
              <Col md={3} className='fapp-label light'>Period</Col>
            </Row>
            { this.renderList(reports[names[i]]) }
          </CardBody>
        </Card>
      );
    }

    return dashboards;
  }

  render() {
    if (this.state.loading) return <Loading />;

    if (this.props.reports && this.props.reports.length == 0)
      return <NoContent paddingTop='18%' center msg='No Reports' />;

    let reportsByName = this.getReportsByName();
    return (
      <div className='fapp-content' style={{ marginLeft: '2rem', marginRight: '2rem', paddingTop: '30px' }}>
        { this.renderListByDashboard(reportsByName) }
      </div>
    );
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(ReportsList);