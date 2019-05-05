import React, { Component } from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import { Row, Button, Modal, ModalHeader, 
  ModalBody, ModalFooter, Form, Input } from 'reactstrap';
import Select from 'react-select';
import DashboardUnit from './DashboardUnit';

import { fetchDashboards, createDashboard, 
  fetchDashboardUnits } from '../../actions/dashboardActions';
import { createSelectOpts, makeSelectOpt } from '../../util';
import NoContent from '../../common/NoContent';
import Loading from '../../loading';
import { PRESENTATION_CARD } from '../Query/common';

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
    createDashboard,
  }, dispatch);
}

class Dashboard extends Component {
  constructor(props) {
      super(props);

      this.state = {
        loaded: false,

        selectedDashboard: null,
        loadingUnits: false,

        editDashboard: false,
        showCreateModal: false,
        createModalMessage: null,
        createName: null,
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
    this.setState({ selectedDashboard: option, loadingUnits: true });
    this.props.fetchDashboardUnits(this.props.currentProjectId, option.value)
      .then(() => this.setState({ loadingUnits: false }))
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

  isEditable() {
    return this.props.dashboardUnits && this.props.dashboardUnits.length > 0;
  }

  renderDashboard() {
    if (this.state.loadingUnits) return <Loading paddingTop='10%' />
    if (this.props.dashboardUnits.length == 0) 
      return <NoContent center msg='No charts' />

    let pDashUnits = this.props.dashboardUnits;
    let largeUnits = [];
    let cardUnits = [];

    let cardIndex = 1;
    for (let i=0; i < pDashUnits.length; i++) {
      let pUnit = pDashUnits[i];
      if (pUnit.presentation && pUnit.presentation === PRESENTATION_CARD) {
        cardUnits.push(<DashboardUnit showClose={this.state.editDashboard} card cardIndex={cardIndex} data={pUnit} />)
        cardIndex++;
      } else {
        largeUnits.push(<DashboardUnit showClose={this.state.editDashboard} data={pUnit} />);
      }
    }
      
    return <div>
      <Row class="fapp-select"> { cardUnits } </Row>
      <Row class="fapp-select"> { largeUnits } </Row>
    </div>
  }

  toggleEditDashboard = () => {
    this.setState({ editDashboard: !this.state.editDashboard });
  }

  isLoading() {
    return !this.state.loaded;
  }

  renderEditButton() {
    if (!this.isEditable()) return null;
    let text = this.state.editDashboard ? 'Save' : 'Edit';
    let color = this.state.editDashboard ? 'success' : 'danger' 
    return <Button style={{ marginLeft: '10px', height: 'auto', marginBottom: '4px' }} onClick={this.toggleEditDashboard} outline={!this.state.editDashboard} color={color}> { text } </Button>
  }

  toggleCreateModal = () => {
    this.setState({ showCreateModal: !this.state.showCreateModal });
  }

  setCreateDashboardName = (e) => {
    this.setState({ createModalMessage: null });

    let name = e.target.value.trim();
    if (name == "") console.error("Dashboard name cannot be empty.");
    this.setState({ createName: name });
  }

  showCreateFailure(msg='Failed to create dashboard') {
    this.setState({ createModalMessage: msg });
  }

  create = () => {
    if (this.state.createName == null || this.state.createName == "" ){
      this.showCreateFailure('Dashboard name cannot be empty');
      return
    }
    this.props.createDashboard(this.props.currentProjectId, { name: this.state.createName })
      .then((r) => {
        if (!r.ok) this.showCreateFailure();
        else this.toggleCreateModal();
      })
      .catch(this.showCreateFailure);
  }

  render() {
    if (this.isLoading()) return <Loading paddingTop='20%'/>;

    return (
      <div className='fapp-content' style={{marginLeft: '1rem', marginRight: '1rem', paddingTop: '30px' }}>
        <div style={{ marginBottom: '45px', width: '100%'}}>
          <div class="fapp-select" style={{ width: '300px', display: 'inline-block' }}>
            <span style={{ fontSize: '11px', color: '#444', fontWeight: '500'}}> Dashboards </span>
            <Select
              onChange={this.onSelectDashboard}
              options={createSelectOpts(this.getDashboardsOptSrc())}
              placeholder='Select a dashboard'
              value={this.getSelectedDashboard()}
            />
          </div>
          <Button onClick={this.toggleCreateModal} style={{ marginLeft: '10px', height: 'auto', marginBottom: '4px' }} outline color='primary'> Create </Button>
          { this.renderEditButton() }
        </div>
        { this.renderDashboard() }

        <Modal isOpen={this.state.showCreateModal} toggle={this.toggleCreateModal} style={{marginTop: '10rem'}}>
          <ModalHeader toggle={this.toggleCreateModal}>Create dashboard</ModalHeader>
          <ModalBody style={{padding: '25px 35px'}}>
            <div style={{textAlign: 'center', marginBottom: '15px'}}>
              <span style={{display: 'inline-block'}} className='fapp-error' hidden={this.state.createModalMessage == null}>{ this.state.createModalMessage }</span>
            </div>
            <Form >
              <span class='fapp-label'>Name</span>         
              <Input className='fapp-input' type="text" placeholder="Your dashboard name" onChange={this.setCreateDashboardName} />
            </Form>
          </ModalBody>
          <ModalFooter style={{borderTop: 'none', paddingBottom: '30px', paddingRight: '35px'}}>
            <Button outline color="success" onClick={this.create}>Create</Button>
            <Button outline color='danger' onClick={this.toggleCreateModal}>Cancel</Button>
          </ModalFooter>
        </Modal>
      </div>
    );
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(Dashboard);