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
      listByName: null,
      showMore: null,
    }
  }
  
  componentWillMount() {
      this.props.fetchProjectReportsList(this.props.currentProjectId)
        .then(() => { 
          let reportsByName = this.getReportsByName();

          this.setState({ 
            loading: false, 
            listByName: this.getInitListByName(reportsByName),
            showMore: this.getInitShowMore(reportsByName),
          });
      });
  }

  getInitListByName(reportsByName) {
    let names = Object.keys(reportsByName);

    let initReports = {};
    for (let i=0; i<names.length; i++) {
      initReports[names[i]] = reportsByName[names[i]].slice(0, 5);
    }

    return initReports;
  }

  getInitShowMore(reportsByName) {
    let names = Object.keys(reportsByName);

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

  getReportsByName() {
    let reportsByName = {};

    if (!this.props.reports) return reportsByName;
    
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

    return reportsByName;
  }

  loadListByName = (name) => {
    let reports = this.getReportsByName();
    let list = reports[name].slice(INIT_LIST_SIZE); 

    this.setState((prevState) => {
      let _state = prevState;
      _state.listByName[name] = [...prevState.listByName[name], ...list];
      _state.showMore[name] = false;
      return _state;
    });
  }

  renderListByName(name) {
    let list = [];
    
    let reports = this.state.listByName[name];
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

    let reports = this.state.listByName;
    let names = Object.keys(reports);

    for (let i=0; i<names.length; i++) {
      let name = names[i];

      dashboards.push(
        <Card className='fapp-card secondary-list'>
          <CardHeader style={{ marginBottom: '5px' }}>
            <strong> { "Report - " + name } </strong>
          </CardHeader>
          <CardBody>
            <Row style={{ marginBottom: '10px' }} >
              <Col md={2} className='fapp-label light'>Type</Col>
              <Col md={3} className='fapp-label light'>Period</Col>
            </Row>
            { this.renderListByName(name) }

            <Button style={{ marginTop: '5px' }} size='sm' color='primary' hidden={!this.state.showMore[name]}
              outline onClick={() => this.loadListByName(name)}>
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

    let reportsByName = this.getReportsByName();
    return (
      <div className='fapp-content' style={{ marginLeft: '2rem', marginRight: '2rem', paddingTop: '30px' }}>
        { this.renderListByDashboard(reportsByName) }
      </div>
    );
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(ReportsList);