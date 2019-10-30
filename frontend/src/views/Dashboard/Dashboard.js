import React, { Component } from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import { Redirect } from 'react-router-dom';
import { Row, Col, Button, Modal, ModalHeader, 
  ModalBody, ModalFooter, Form, Input } from 'reactstrap';
import Select from 'react-select';
import arrayMove from 'array-move';
import { SortableContainer, SortableElement } from 'react-sortable-hoc';
import 'react-date-range/dist/styles.css';
import 'react-date-range/dist/theme/default.css';
import moment from 'moment';

import DashboardUnit from './DashboardUnit';
import { fetchDashboards, createDashboard, updateDashboard,
  fetchDashboardUnits } from '../../actions/dashboardActions';
import { createSelectOpts, makeSelectOpt } from '../../util';
import NoContent from '../../common/NoContent';
import ClosableDateRangePicker from '../../common/ClosableDatePicker';
import Loading from '../../loading';
import { PRESENTATION_CARD, DEFAULT_DATE_RANGE, 
  DEFINED_DATE_RANGES } from '../Query/common';

const TYPE_OPTS = [
  { label: "Only me", value: "pr" },
  { label: "All agents", value: "pv" }
]

const UNIT_TYPE_CARD = "card";
const UNIT_TYPE_CHART = "chart";

const SortableUnit = SortableElement(({ value, card }) => {
  let size = card ? 3 : 6;
  return <Col md={size} style={{ display:'inline-block', padding: '0 15px' }}> { value } </Col>
});

const SortableUnitList = SortableContainer(({ children }) => {
  return <Row> { children } </Row>;
});

const mapStateToProps = store => {
  return {
    currentProjectId: store.projects.currentProjectId,
    currentAgent: store.agents.agent,
    dashboards: store.dashboards.dashboards,
    dashboardUnits: store.dashboards.units,
  };
}

const mapDispatchToProps = dispatch => {
  return bindActionCreators({ 
    fetchDashboards,
    fetchDashboardUnits,
    createDashboard,
    updateDashboard,
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
        createSelectedType: null,
        createName: null,

        showDatePicker: false,
        dateRange: null,
      }
  }

  componentWillMount() {
    this.props.fetchDashboards(this.props.currentProjectId)
      .then(() => {
        if (this.props.dashboards.length == 0) {
          this.setState({ loaded: true })
        }
        
        let selectedDashboard = this.getSelectedDashboard();
        if (selectedDashboard != null) {
          this.props.fetchDashboardUnits(this.props.currentProjectId, selectedDashboard.value)
            .then(() => this.setState({ loaded: true }))
            .catch(console.error);
        }
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
    this.setLastSeenDashboard(option.value);

    this.props.fetchDashboardUnits(this.props.currentProjectId, option.value)
      .then(() => this.setState({ loadingUnits: false }))
      .catch(console.error);
  }

  getLastSeenDashboardKey() {
    //_dashboard_ls:<agent_id>:<project_id>
    return this.props.currentAgent && this.props.currentProjectId ? 
      '_dashboard_ls:'+this.props.currentAgent.uuid+':'+this.props.currentProjectId : '';
  }

  getLastSeenDateRangeForDashboardKey() {
    //_dashboard_daterange_ls:<agent_id>:<project_id>:<dashboard_id>
    let key = this.props.currentAgent && this.props.currentProjectId ? 
    '_dashboard_daterange_ls:'+this.props.currentAgent.uuid+':'+this.props.currentProjectId : '';

    let currentDashboard = this.getCurrentDashboard();
    return key != '' && currentDashboard && currentDashboard.id ?
      key + ':' + currentDashboard.id : '';
  }

  setLastSeenDashboard(dashboardId) {
    let dashboardKey = this.getLastSeenDashboardKey();
    if (dashboardKey == '') return;
    localStorage.setItem(dashboardKey, dashboardId);
  }

  getLastSeenDashboard() {
    let dashboardKey = this.getLastSeenDashboardKey();
    if (dashboardKey == '') return null;
    return localStorage.getItem(dashboardKey);
  }

  setLastSeenDateRangeForDashboard(range) {
    let dateRangeKey = this.getLastSeenDateRangeForDashboardKey();
    if (dateRangeKey == '') return null;
    return localStorage.setItem(dateRangeKey, range);
  }

  getLastSeenDateRangeForDashboard() {
    let dateRangeKey = this.getLastSeenDateRangeForDashboardKey();
    if (dateRangeKey == '') return null;
    return localStorage.getItem(dateRangeKey);
  }

  getSelectedDashboard() {
    if (this.state.selectedDashboard != null) 
      return this.state.selectedDashboard;

    let lsDashboardId = this.getLastSeenDashboard();
    // if user seen dashboard found.
    if (lsDashboardId) {
      let lsDashboardName = this.getDashboardsOptSrc()[lsDashboardId];
      // if dashboard found on existing accessible dashboards list.
      if (lsDashboardName) return makeSelectOpt(lsDashboardId, lsDashboardName)
    }
    

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

  getPositionsMapFromList(order) {
    let positions = {}
    // uses array index as position.
    for (let i=0; i < order.length; i++) 
      positions[order[i]] = i;
    
    return positions;
  }

  handleUnitPositionChange(unitType, oldIndex, newIndex) {
    let positionMap = this.getUnitsPositionByType(unitType);
    let currentPositionById = [];
    
    for(let k in positionMap) {
      currentPositionById[positionMap[k]] = k;
    }

    // moves the id as per position change.
    let newPosition = arrayMove(currentPositionById, oldIndex, newIndex);
    let newPositionMap = this.getPositionsMapFromList(newPosition);
    
    let dashboard = this.getCurrentDashboard();
    let updatablePosition = { ...dashboard.units_position };
    // updates positions only for the changed type.
    updatablePosition[unitType] = newPositionMap;

    let dashboardOption = this.getSelectedDashboard();
    let currentDashboardId = dashboardOption.value;

    // drags without position change should no trigger update.
    if (JSON.stringify(positionMap) != JSON.stringify(newPositionMap))
      this.props.updateDashboard(this.props.currentProjectId, 
        currentDashboardId, { units_position: updatablePosition });
  }

  handleCardUnitPositionChange = ({ oldIndex, newIndex }) => {
    this.handleUnitPositionChange(UNIT_TYPE_CARD, oldIndex, newIndex);
  }

  handleChartUnitPositionChange = ({ oldIndex, newIndex }) => {
    this.handleUnitPositionChange(UNIT_TYPE_CHART, oldIndex, newIndex);
  }

  getCurrentDashboard() {
    let dashboard = this.getSelectedDashboard()
    let dashboardId = dashboard.value;

    for(let i in this.props.dashboards) 
      if (dashboardId == this.props.dashboards[i].id)
        return this.props.dashboards[i];
  }

  getUnitType(presentation) {
    return presentation === PRESENTATION_CARD ? UNIT_TYPE_CARD : UNIT_TYPE_CHART;
  }

  getInitialPositionFromOrderOfUnits(unitType) {
    let positionMap = {}
    
    let position = 0;
    for (let i in this.props.dashboardUnits) {
      let unit = this.props.dashboardUnits[i];
      if (this.getUnitType(unit.presentation) == unitType) {
        positionMap[unit.id] = position;
        position++;
      }
    }
    console.warn("Positioning charts by given order as positions of "+unitType+" is null.");

    return positionMap;
  }

  getUnitsPositionByType(unitType) {
    let dashboard = this.getCurrentDashboard();

    if (!dashboard.units_position || !dashboard.units_position[unitType]) 
      return this.getInitialPositionFromOrderOfUnits(unitType);

    return dashboard['units_position'][unitType];
  }
  
  renderDashboard() {
    if (this.state.loadingUnits) return <Loading paddingTop='10%' />
    if (this.props.dashboardUnits.length == 0) 
      return <NoContent center msg='No charts' />

    let pDashUnits = this.props.dashboardUnits;
    let cardPositions = this.getUnitsPositionByType(UNIT_TYPE_CARD);
    let chartPositions = this.getUnitsPositionByType(UNIT_TYPE_CHART);
    let chartUnits = [], cardUnits = [];

    // Arranges units by position from dashboard.
    let cardIndex = 1;
    for (let i=0; i < pDashUnits.length; i++) {
      let pUnit = pDashUnits[i];
      if (pUnit.presentation && pUnit.presentation === PRESENTATION_CARD) {
        cardUnits[cardPositions[pUnit.id]] = {
          unit: <DashboardUnit dateRange={this.getCurrentDateRange()} editDashboard={this.state.editDashboard} cardIndex={cardIndex} data={pUnit} position={cardPositions[pUnit.id]} />,
          position: cardPositions[pUnit.id],
        };
        cardIndex++;
      } else {
        chartUnits[chartPositions[pUnit.id]] = {
          unit: <DashboardUnit dateRange={this.getCurrentDateRange()} editDashboard={this.state.editDashboard} data={pUnit} position={cardPositions[pUnit.id]} />,
          position: chartPositions[pUnit.id],
        };
      }
    }

    return (
      <div>
        <SortableUnitList distance={10} axis='xy' onSortEnd={this.handleCardUnitPositionChange}>
          { cardUnits.map((value) => (<SortableUnit disabled={!this.state.editDashboard} key={`card-${value.position}`} index={value.position} value={value.unit} card />)) }
        </SortableUnitList>
        <SortableUnitList distance={10} axis='xy' onSortEnd={this.handleChartUnitPositionChange}>
          { chartUnits.map((value) => (<SortableUnit disabled={!this.state.editDashboard} key={`chart-${value.position}`} index={value.position} value={value.unit} />)) }
        </SortableUnitList>
      </div>
    )
  }

  toggleEditDashboard = () => {
    this.setState({ editDashboard: !this.state.editDashboard });
  }

  isLoading() {
    return !this.state.loaded;
  }

  renderEditButton() {
    if (!this.isEditable()) return null;
    let text = this.state.editDashboard ? 'Done Editing' : 'Edit';
    let color = this.state.editDashboard ? 'success' : 'danger' 
    return <Button style={{ marginLeft: '10px', height: 'auto', marginBottom: '4px' }} 
      onClick={this.toggleEditDashboard} outline={!this.state.editDashboard} color={color}> { text } </Button>
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
    
    let selectedType = this.getSelectedCreateType();
    this.props.createDashboard(this.props.currentProjectId, { name: this.state.createName, type: selectedType.value })
      .then((r) => {
        if (!r.ok) this.showCreateFailure();
        else this.toggleCreateModal();
      })
      .catch(this.showCreateFailure);
  }

  onCreateTypeChange = (option) => {
    this.setState({ createSelectedType: option });
  }

  getSelectedCreateType() {
    if (this.state.createSelectedType != null) 
      return this.state.createSelectedType;

    return TYPE_OPTS[0];
  }

  handleDateRangeSelect = (range) => {
    range.selected.label = null; // set null on custom range.
    this.setState({ dateRange: [range.selected] });
    this.setLastSeenDateRangeForDashboard(JSON.stringify(range.selected));
  }

  closeDatePicker = () => {
    this.setState({ showDatePicker: false }); 
  }

  toggleDatePickerDisplay = () => {
    this.setState({ showDatePicker: !this.state.showDatePicker });
  }

  readableDateRange(range) {
    // Use label for default date range.
    if(range.startDate ==  DEFAULT_DATE_RANGE.startDate 
      && range.endDate == DEFAULT_DATE_RANGE.endDate)
      return DEFAULT_DATE_RANGE.label;

    return moment(range.startDate).format('MMM DD, YYYY') + " - " +
      moment(range.endDate).format('MMM DD, YYYY');
  }

  getCurrentDateRange() {
    if (this.state.dateRange) return this.state.dateRange;

    let lsDateRangeStr = this.getLastSeenDateRangeForDashboard()
    return lsDateRangeStr ? [JSON.parse(lsDateRangeStr)] : [DEFAULT_DATE_RANGE]
  }

  render() {
    if (this.isLoading()) return <Loading paddingTop='20%'/>;

    return (
      <div className='fapp-content' style={{marginLeft: '1rem', marginRight: '1rem', paddingTop: '30px' }}>
        <div style={{ marginBottom: '32px', width: '100%', textAlign: 'center'}}>
          <div className="fapp-select light center" style={{ width: '260px', display: 'inline-block' }}>
            <Select
              onChange={this.onSelectDashboard}
              options={createSelectOpts(this.getDashboardsOptSrc())}
              placeholder='Select a dashboard'
              value={this.getSelectedDashboard()}
            />
          </div>
          <Button onClick={this.toggleCreateModal} style={{ marginLeft: '10px', height: 'auto', marginBottom: '4px' }} outline color='primary'> Create </Button>
          { this.renderEditButton() }

          <button style={{ border: '1px solid #bbb', color: '#444', right: '45px', position: 'absolute', marginTop: '5px', padding: '7px 15px', borderRadius: '5px', outline: 'none' }} 
            onClick={this.toggleDatePickerDisplay}>
            <i className="fa fa-calendar" style={{marginRight: '10px'}}></i>
            { this.readableDateRange(this.getCurrentDateRange()[0]) }
          </button>
          <div className='fapp-date-picker' style={{ display: 'block', marginTop: '10px', right: '45px' }} hidden={!this.state.showDatePicker}>
            <ClosableDateRangePicker
              ranges={this.getCurrentDateRange()}
              onChange={this.handleDateRangeSelect}
              staticRanges={ DEFINED_DATE_RANGES }
              inputRanges={[]}
              minDate={new Date('01 Jan 2000 00:00:00 GMT')} // range starts from given date.
              maxDate={new Date()}
              closeDatePicker={this.closeDatePicker}
            />
          </div>
        </div>

        { this.renderDashboard() }

        <Modal isOpen={this.state.showCreateModal} toggle={this.toggleCreateModal} style={{marginTop: '10rem'}}>
          <ModalHeader toggle={this.toggleCreateModal}>New dashboard</ModalHeader>
          <ModalBody style={{padding: '15px 35px'}}>
            <div style={{textAlign: 'center', marginBottom: '15px'}}>
              <span style={{display: 'inline-block'}} className='fapp-error' hidden={this.state.createModalMessage == null}>{ this.state.createModalMessage }</span>
            </div>
            <Form >
              <span className='fapp-label'>Name</span>         
              <Input className='fapp-input' type="text" placeholder="Your dashboard name" onChange={this.setCreateDashboardName} />
              <span className='fapp-label' style={{ marginTop: '18px', marginBottom: '10px', display: 'block' }}>Visiblity</span> 
              <div className='fapp-select'>
                <Select
                  onChange={this.onCreateTypeChange}
                  options={TYPE_OPTS}
                  placeholder='Select visiblity'
                  value={this.getSelectedCreateType()}
                />
              </div>        
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