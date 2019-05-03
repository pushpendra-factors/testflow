import React, { Component } from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import { Row } from 'reactstrap';
import Select from 'react-select';
import DashboardUnit from './DashboardUnit';

import { fetchDashboards, fetchDashboardUnits } from '../../actions/dashboardActions';
import { createSelectOpts, makeSelectOpt } from '../../util';
import Loading from '../../loading';

const mapStateToProps = store => {
  return {
    currentProjectId: store.projects.currentProjectId,
    dashboards: store.dashboards.dashboards,
    dashboardUnits: store.dashboards.units,
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
    
    let units = [];
    let pDashUnits = this.props.dashboardUnits;
    for (let i=0; i < pDashUnits.length; i++) 
      units.push(<DashboardUnit data={pDashUnits[i]} />);

    return <Row class="fapp-select"> { units } </Row>
  }

  isLoading() {
    return !this.state.loaded;
  }

  render() {
    if (this.isLoading()) return <Loading paddingTop='20%'/>;

    return (
      <div className='fapp-content' style={{marginLeft: '1rem', marginRight: '1rem'}}>
        <div class="fapp-select" style={{width: '300px', marginBottom: '25px'}}>
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