import React, { Component } from 'react';
import { connect } from 'react-redux';
import moment from 'moment';
import { bindActionCreators } from 'redux';
import {
  Col,
  Row,
  Card,
  CardHeader,
  CardBody,
  Button,
} from 'reactstrap';

import { fetchProjectReportsList } from '../../actions/reportActions';
import { readableTimstamp } from '../../util';
import Loading from '../../loading';
import NoContent from '../../common/NoContent';

const INIT_LIST_SIZE = 5;

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
      loading: true,
      listByDashboard: null,
      showMore: null,
    }
  }
  
  componentWillMount() {
      this.props.fetchProjectReportsList(this.props.currentProjectId)
        .then(() => { 
          let reportsByDashboard = this.getReportsByDashboard();

          this.setState({ 
            loading: false, 
            listByDashboard: this.getInitListByName(reportsByDashboard),
            showMore: this.getInitShowMore(reportsByDashboard),
          });
      });
  }

  getInitListByName(reportsByDashboard) {
    let names = Object.keys(reportsByDashboard);

    let initReports = {};
    for (let i=0; i<names.length; i++) {
      initReports[names[i]] = reportsByDashboard[names[i]].slice(0, 5);
    }

    return initReports;
  }

  getInitShowMore(reportsByDashboard) {
    let names = Object.keys(reportsByDashboard);

    let showMore = {};
    for (let i=0; i<names.length; i++) {
      showMore[names[i]] = true;
    }

    return showMore;
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

  getReportsByDashboard() {
    let reportsByDashboard = {};

    if (!this.props.reports) return reportsByDashboard;
    
    let reports = this.props.reports;
    for (let i=0; i<reports.length; i++) {
      if (!reportsByDashboard[reports[i].dashboard_id])
        reportsByDashboard[reports[i].dashboard_id] = [];

      let period = reports[i].type == "m" ? moment.unix(reports[i].start_time).utc().format('MMMM, YYYY') : 
        (readableTimstamp(reports[i].start_time) + " - "  + readableTimstamp(reports[i].end_time));
        reportsByDashboard[reports[i].dashboard_id].push({
        id: reports[i].id,
        typeName: this.getTypeName(reports[i].type),
        period: period,
        dashboardName: reports[i].dashboard_name,
      })
    }

    return reportsByDashboard;
  }

  loadListByDashboard = (dashboardId) => {
    let reports = this.getReportsByDashboard();
    let list = reports[dashboardId].slice(INIT_LIST_SIZE); 

    this.setState((prevState) => {
      let _state = prevState;
      _state.listByDashboard[dashboardId] = [...prevState.listByDashboard[dashboardId], ...list];
      _state.showMore[dashboardId] = false;
      return _state;
    });
  }

  renderList(dashboardId) {
    let list = [];
    
    let reports = this.state.listByDashboard[dashboardId];
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

  renderListByDashboard() {
    let dashboards = [];

    let reports = this.state.listByDashboard;
    let dashboardIds = Object.keys(reports);

    for (let i=0; i<dashboardIds.length; i++) {
      let dashboardId = dashboardIds[i];

      dashboards.push(
        <Card className='fapp-card secondary-list'>
          <CardHeader style={{ marginBottom: '5px' }}>
            <strong> { "Report - " + reports[dashboardId][0].dashboardName } </strong>
          </CardHeader>
          <CardBody>
            <Row style={{ marginBottom: '10px' }} >
              <Col md={2} className='fapp-label light'>Type</Col>
              <Col md={3} className='fapp-label light'>Period</Col>
            </Row>
            { this.renderList(dashboardId) }

            <Button style={{ marginTop: '5px' }} size='sm' color='primary' hidden={!this.state.showMore[dashboardId]}
              outline onClick={() => this.loadListByDashboard(dashboardId)}>
              Show more
            </Button>
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

    let reportsByName = this.getReportsByDashboard();
    return (
      <div className='fapp-content' style={{ marginLeft: '2rem', marginRight: '2rem', paddingTop: '30px' }}>
        { this.renderListByDashboard(reportsByName) }
      </div>
    );
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(ReportsList);