import React, { Component } from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import { Card, CardHeader, CardBody, Modal, ModalBody } from 'reactstrap';
import { Redirect } from 'react-router-dom';
import moment from 'moment';

import { runQuery, viewQuery } from '../../actions/projectsActions';
import { deleteDashboardUnit, updateDashboardUnit } from '../../actions/dashboardActions';
import Loading from '../../loading';
import BarChart from '../Query/BarChart';
import LineChart from '../Query/LineChart';
import TableChart from '../Query/TableChart';
import { PRESENTATION_BAR, PRESENTATION_LINE, 
  PRESENTATION_TABLE, PRESENTATION_CARD, 
  PRESENTATION_FUNNEL, PROPERTY_VALUE_TYPE_DATE_TIME } from '../Query/common';
import { slideUnixTimeWindowToCurrentTime } from '../../util';
import FunnelChart from '../Query/FunnelChart';

const LINE_LEGEND_DISPLAY_LIMIT = 10;
const CARD_FONT_COLOR = '#FFF';
const CARD_BACKGROUNDS = ['#63c2de', '#eb9532', '#20a8d8', '#4dbd74', '#f86c6b' ]

const mapStateToProps = store => {
  return {
    currentProjectId: store.projects.currentProjectId,
  };
}

const mapDispatchToProps = dispatch => {
  return bindActionCreators({ 
    runQuery,
    viewQuery,
    deleteDashboardUnit,
    updateDashboardUnit,
  }, dispatch);
}

class DashboardUnit extends Component {
  constructor(props) {
    super(props);

    this.state = {
      loading: false,
      presentationProps: null,

      title: null,
      editTitle: false,

      fullScreen: false,
      redirectToViewQuery: false,
    }
  }

  getUnitBackground() {
    let cardIndex = this.props.cardIndex;
    let poolLength = CARD_BACKGROUNDS.length;
    return CARD_BACKGROUNDS[cardIndex % poolLength];
  }

  setPresentationProps(result) {
    let props = null;

    if (this.props.data.presentation === PRESENTATION_BAR) {
      props = { queryResult: result, legend: false } 
    }

    if (this.props.data.presentation === PRESENTATION_LINE) {
      props = { hideLegend: true, queryResult: result }
    }

    if (this.props.data.presentation === PRESENTATION_TABLE) {
      props = { queryResult: result }
    }

    if (this.props.data.presentation == PRESENTATION_CARD) {
      props = { noHeader: true,  card: true, queryResult: result }
    }

    if (this.props.data.presentation == PRESENTATION_FUNNEL) {
      props = { queryResult: result, small: true }
    }

    this.setState({ presentationProps: props });
  }

  setQueryPeriod(query, dateRange) {
    let from =  moment(dateRange[0].startDate).unix();
    let to = moment(dateRange[0].endDate).unix();

    query.fr = from; // in utc.
    query.to = to; // in utc.
  }

  execQuery() {
    this.setState({ loading: true });
    
    let query = this.props.data.query;
    this.setQueryPeriod(query, this.props.dateRange);

    // override datetime property value.
    for(let ei=0; ei<query.ewp.length; ei++) {
      let ewp = query.ewp[ei];

      for(let pi=0; pi < ewp.pr.length; pi++) {
        if (ewp.pr[pi].ty == PROPERTY_VALUE_TYPE_DATE_TIME) {
          let propertyValue = JSON.parse(ewp.pr[pi].va);
          if (propertyValue.ovp) {
            let newPeriod = slideUnixTimeWindowToCurrentTime(propertyValue.fr, propertyValue.to);
            propertyValue.fr = newPeriod.from;
            propertyValue.to = newPeriod.to;
            ewp.pr[pi].va = JSON.stringify(propertyValue);
          }
        }
      }
    }

    runQuery(this.props.currentProjectId, query)
      .then((r) => {
        this.setState({ loading: false });
        this.setPresentationProps(r.data);
      })
      .catch(console.error);
  }

  componentWillMount() {
    this.execQuery();
  }

  present(props, showLegend=false) {
    if (this.state.loading) {
      return <Loading paddingTop={ this.isCard() ? '6%':'12%' } />;
    }

    if (!props) return null;

    if (this.props.data.presentation === PRESENTATION_BAR) {
      return <BarChart {...props} />;
    }

    if (this.props.data.presentation === PRESENTATION_LINE) {
      let lineProps = { ...props, hideLegend: !showLegend }
      return <LineChart {...lineProps} />;
    }

    if (this.props.data.presentation === PRESENTATION_TABLE) {
      return <TableChart {...props} />;
    }

    if (this.props.data.presentation == PRESENTATION_CARD) {
      return <TableChart {...props} />;
    }

    if (this.props.data.presentation == PRESENTATION_FUNNEL) {
      return <FunnelChart {...props} />;
    }
    
    return null;
  }

  getCardBodyStyleByProps() {
    let style = { padding: '1.5rem 1.5rem', paddingTop: '0.6rem', height: '320px' };

    if (this.props.data.presentation === PRESENTATION_TABLE) {
      let changes = { padding: '0', 'overflowX': 'scroll' };
      style = { ...style, ...changes };
    }

    if (this.props.data.presentation === PRESENTATION_CARD) {
      style.height = '130px';
      style.padding = '0';
      style.paddingTop = '0';
      style.background = this.getUnitBackground();
      style.color = CARD_FONT_COLOR;
    }

    return style;
  }

  getInlineButtonStyle() {
    return { 
      background: 'none', 
      border: 'none',
      padding: '0 4px', 
      fontSize: '17px', 
      color: this.isCard() ? '#FFF' : '#444'
    }
  }

  getCardHeaderStyleByProps() {
    if (this.props.data.presentation !== PRESENTATION_CARD) return null;
    let style = {};
    style.textAlign = 'center';
    style.background = this.getUnitBackground();
    style.color = CARD_FONT_COLOR;
    return style;
  }
  
  getCardStyleByProps() {
    let style = { marginBottom: '30px' };
    if (this.props.editDashboard) style.cursor = 'all-scroll';
    if (this.props.data.presentation === PRESENTATION_CARD) {
      style.border = 'none';
    }

    return style;
  }

  delete = () => {
    let unit = this.props.data;
    this.props.deleteDashboardUnit(unit.project_id, unit.dashboard_id, unit.id);
  }

  isCard() {
    return this.props.data.presentation === PRESENTATION_CARD;
  }

  onTitleChange = (e) => {
    this.setState({ title: e.target.value });
  }

  getTitleInputStyle() {
    let style = {
      width: '70%',
      background: 'transparent',
      fontWeight: '500',
      fontSize: '13px',
      borderRadius: '4px',
      marginRight: '6px',
    }

    let isCard = this.isCard();
    style.color = isCard ? '#fff' : '#444';
    style.border = isCard ? '1px solid #fff' : '1px solid #DDD'; 
    style.padding = isCard ? '0 7px' : '3px 7px';

    return style;
  }

  editTitle = () => {
    this.setState({ editTitle: true });
  }

  isTitleChanged() {
    return this.state.title != null && this.state.title.trim() != "" &&
      this.state.title != this.props.data.title;
  }

  closeEditTitle = () => {
    let state = { editTitle : false };
    // reset state.
    if (this.isTitleChanged()) state.title = this.props.data.title;
  
    this.setState(state);
  }

  showTitleEditor() {
    return this.state.editTitle && this.props.editDashboard
  }

  showTitle() {
    return (!this.props.editDashboard || !this.state.editTitle);
  }

  getTitle() {
    return this.state.title == null ? this.props.data.title : this.state.title;
  }
  
  handleUpdateTitleFailure() {
    this.setState({ title: this.props.data.title });
    // Todo: show title update failure on UI.
    console.error("Failed to update title.");
  }

  saveEditedTitle = () => {
    let unit = this.props.data;

    if (!this.isTitleChanged()) {
      this.setState({ editTitle: false, title: unit.title });
      return;
    }
    
    
    this.props.updateDashboardUnit(unit.project_id, unit.dashboard_id, 
      unit.id, {title: this.state.title})
      .then((r) => {
        if (r.error) this.handleUpdateTitleFailure();
      })
      .catch(this.handleUpdateTitleFailure);
    // close editor.
    this.setState({ editTitle: false });
  }

  getEditTitleStyle() {
    if (!this.props.editDashboard) return null;

    return { 
      maxWidth: this.isCard() ? '180px' : null, 
      display: 'inline-block'
    }
  }

  // Todo: Avoid execQuery on position change by
  // moving the query result to ParentComponent (dashboard).
  componentDidUpdate(prevProps) {
    if (prevProps.data.id != this.props.data.id || 
      JSON.stringify(prevProps.dateRange) != JSON.stringify(this.props.dateRange)) {
      this.execQuery();
    }
  }

  addQueryToViewStore = () => {
    if (this.props.data && this.props.data.query) {
      this.props.viewQuery(this.props.data.query);
      this.setState({ redirectToViewQuery: true })
    }
  }

  toggleFullScreen = () => {
    this.setState({ fullScreen: !this.state.fullScreen });
  } 

  render() {
    if (this.state.redirectToViewQuery) 
      return <Redirect to='/core?view=true' />;

    return (
      <Card className='fapp-dunit' style={this.getCardStyleByProps()}>
        <CardHeader style={this.getCardHeaderStyleByProps()}>
          <div style={{ textAlign: 'right', marginTop: '-10px', marginRight: '-18px', height: '18px' }}>
            <strong onClick={this.delete} style={{ fontSize: '14px', cursor: 'pointer', padding: '0 10px', color: this.isCard() ? '#FFF' : '#AAA' }} hidden={!this.props.editDashboard}>x</strong>
          </div>

          <div style={{ textAlign: 'right', marginTop: '-15px', marginRight: '-22px', height: '18px' }} hidden={this.isCard()}>
            <strong onClick={this.toggleFullScreen} style={{ fontSize: '13px', cursor: 'pointer', padding: '0 10px', color: '#888' }} hidden={this.props.editDashboard} >
              <i className='fa fa-expand'></i>
            </strong>
          </div>

          <div style={{ textAlign: 'right', marginTop: this.isCard() ? '-17px' : '-18px', height: '18px', marginRight: this.isCard() ? '-22px' : null }}>
            <strong onClick={this.addQueryToViewStore} style={{ fontSize: '13px', cursor: 'pointer', padding: '0 10px', color: this.isCard() ? '#FFF' : '#444' }} hidden={this.props.editDashboard} >
              <i className='cui-graph'></i>
            </strong>
          </div>

          <div hidden={!this.showTitle()}>
            <div className='fapp-overflow-dot' style={this.getEditTitleStyle()}> 
              <strong style={{ fontWeight: !this.isCard() ? '500' : null, fontSize: '0.85rem' }} >{ this.getTitle() }</strong> 
            </div>
            <button style={{...this.getInlineButtonStyle(), fontSize: '14px'}} onClick={this.editTitle} hidden={!this.props.editDashboard}><i className='icon-pencil'></i></button>
          </div>

          <div hidden={!this.showTitleEditor()}>
            <input className='no-outline' style={this.getTitleInputStyle()} value={this.getTitle()} onChange={this.onTitleChange} />
            <button style={this.getInlineButtonStyle()} onClick={this.saveEditedTitle}>
              <i className='icon-check'></i>
            </button>
            <button style={this.getInlineButtonStyle()} onClick={this.closeEditTitle}>
              <i className='icon-close'></i>
            </button>
          </div>
        </CardHeader>
        <CardBody style={this.getCardBodyStyleByProps()}>
          { this.present(this.state.presentationProps) }
        </CardBody>

        <Modal isOpen={this.state.fullScreen} toggle={this.toggleFullScreen} style={{ marginTop: "2.5rem", minWidth: "80rem"  }}> 
          <ModalBody>
            <div>
              <span onClick={this.toggleFullScreen} style={{ position: 'absolute', right: '25px', fontSize: '18px', fontWeight: '600', color: '#888', cursor: 'pointer' }}>x</span>
            </div>
            <div style={{ height: "40rem", padding: "40px", overflow: "scroll" }}>
              { this.present(this.state.presentationProps, true) }
            </div>
          </ModalBody>
        </Modal>
      </Card>
    );
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(DashboardUnit);