import React, { Component } from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import { Row, Col, Card, CardHeader, CardBody } from 'reactstrap';
import Select from 'react-select';

import { fetchDashboards, fetchDashboardUnits } from '../../actions/dashboardActions';
import { createSelectOpts, makeSelectOpt } from '../../util';
import Loading from '../../loading';

// To be removed.
import { Bar } from 'react-chartjs-2';

const data = {
  labels: ['January', 'February', 'March', 'April', 'May', 'June', 'July', 'January', 'February', 'March', 'April', 'May', 'June', 'July'],
  datasets: [
    {
      label: 'My First dataset',
      backgroundColor: 'rgba(255,99,132,0.2)',
      borderColor: 'rgba(255,99,132,1)',
      borderWidth: 1,
      hoverBackgroundColor: 'rgba(255,99,132,0.4)',
      hoverBorderColor: 'rgba(255,99,132,1)',
      data: [65, 59, 80, 81, 56, 55, 40, 65, 59, 80, 81, 56, 55, 40]
    }
  ]
};

const mapStateToProps = store => {
  return {
    currentProjectId: store.projects.currentProjectId,
    dashboards: store.dashboards.dashboards,
    dashboardUnits: store.dashboards.dashboardUnits,
  };
}

const mapDispatchToProps = dispatch => {
  return bindActionCreators({ 
    fetchDashboards,
    fetchDashboardUnits,
  }, dispatch);
}

class Dashboard extends Component {
  constructor(props) {
      super(props);

      this.state = {
        loaded: false,

        selectedDashboard: null,
        loadingDashboard: false,
      }
  }

  componentWillMount() {
    this.props.fetchDashboards(this.props.currentProjectId)
      .then(() => {
        let selectedDashboard = this.getSelectedDashboard();
        this.props.fetchDashboardUnits(this.props.currentProjectId, selectedDashboard.value)
          .then(() => this.setState({ loaded: true }))
          .catch(console.error);
      })
  }

  getDashboardsOptSrc() {
    let opts = {}
    for(let i in this.props.dashboards) {
      let dashboard = this.props.dashboards[i];
      opts[dashboard.id] = dashboard.name;
    }
    return opts;
  }

  onSelectDashboard = (option) => {
    this.setState({ selectedDashboard: option, loadingDashboard: true });
    this.props.fetchDashboardUnits(this.props.currentProjectId, option.value)
      .then(() => this.setState({ loadingDashboard: false }))
      .catch(console.error);
  }

  getSelectedDashboard() {
    if (this.state.selectedDashboard != null) 
      return this.state.selectedDashboard;

    // inits selector with first dashboard.
    if (this.props.dashboards  
      && this.props.dashboards.length > 0) {
      return makeSelectOpt(this.props.dashboards[0].id, 
        this.props.dashboards[0].name);
    }

    return null;
  }

  renderDashboard() {
    if (this.state.loadingDashboard) return <Loading paddingTop='10%' />

    return <Row class="fapp-select">
      <Col md={{ size: 6 }}  style={{padding: '0 15px'}}>
        <Card className='fapp-bordered-card' style={{marginTop: '15px'}}>
          <CardHeader>
            <strong>Chart Title</strong>
          </CardHeader>
          <CardBody style={{padding: '1.5rem 0.5rem'}}>
            <div style={{height: '250px'}}>
              <Bar
                data={data}
                options={{
                  maintainAspectRatio: false
                }}
              />
            </div>
          </CardBody>
        </Card>
      </Col>
      <Col md={{ size: 6 }} style={{padding: '0 15px'}}>
        <Card className='fapp-bordered-card' style={{marginTop: '15px'}}>
          <CardHeader>
            <strong>Chart Title</strong>
          </CardHeader>
          <CardBody style={{padding: '1.5rem 0.5rem'}}>
            <div style={{height: '250px'}}>
              <Bar
                data={data}
                options={{
                  maintainAspectRatio: false
                }}
              />
            </div>
          </CardBody>
        </Card>
      </Col>
      <Col md={{ size: 6 }} style={{padding: '0 15px'}}>
        <Card className='fapp-bordered-card' style={{marginTop: '15px'}}>
          <CardHeader>
            <strong>Chart Title</strong>
          </CardHeader>
          <CardBody style={{padding: '1.5rem 0.5rem'}}>
            <div style={{height: '250px'}}>
              <Bar
                data={data}
                options={{
                  maintainAspectRatio: false
                }}
              />
            </div>
          </CardBody>
        </Card>
      </Col>
    </Row>
  }

  isLoading() {
    return !this.state.loaded;
  }

  render() {
    if (this.isLoading()) return <Loading paddingTop='20%'/>;

    return (
      <div className='fapp-content' style={{marginLeft: '1rem', marginRight: '1rem'}}>
        <div class="fapp-select" style={{width: '300px', marginBottom: '20px'}}>
          <span style={{ fontSize: '11px', color: '#444', fontWeight: '500'}}> Select Dashboard </span>
          <Select
            onChange={this.onSelectDashboard}
            options={createSelectOpts(this.getDashboardsOptSrc())}
            placeholder='Select a dashboard'
            value={this.getSelectedDashboard()}
          />
        </div>
        { this.renderDashboard() }
      </div>
    );
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(Dashboard);