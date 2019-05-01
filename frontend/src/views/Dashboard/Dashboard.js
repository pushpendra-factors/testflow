import React, { Component } from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import { Row, Col, Card, CardHeader, CardBody } from 'reactstrap';
import Select from 'react-select';

import { fetchDashboards } from '../../actions/dashboardActions';

// To be removed.
import { Bar } from 'react-chartjs-2';
import { createSelectOpts, getSelectedOpt, makeSelectOpt } from '../../util';

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
  };
}

const mapDispatchToProps = dispatch => {
  return bindActionCreators({ 
    fetchDashboards
  }, dispatch);
}

class Dashboard extends Component {
  constructor(props) {
      super(props);

      this.state = {
        selectedDashboard: null
      }
  }

  componentWillMount() {
    this.props.fetchDashboards(this.props.currentProjectId);
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
    this.setState({selectedDashboard: option});
  }

  getSelectedDashboard() {
    if (this.state.selectedDashboard != null) 
      return this.state.selectedDashboard;

    if (this.props.dashboards && 
      this.props.dashboards.length > 0) {
      return makeSelectOpt(this.props.dashboards[0].id, 
        this.props.dashboards[0].name);
    }

    return null;
  }

  render() {
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

        <Row class="fapp-select">
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
      </div>
    );
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(Dashboard);