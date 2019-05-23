import React, { Component } from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import { Col, Card, CardHeader, CardBody, Input, Button } from 'reactstrap';

import { runQuery } from '../../actions/projectsActions';
import { deleteDashboardUnit, updateDashboardUnit } from '../../actions/dashboardActions';
import Loading from '../../loading';
import BarChart from '../Query/BarChart';
import LineChart from '../Query/LineChart';
import TableChart from '../Query/TableChart';
import { PRESENTATION_BAR, PRESENTATION_LINE, 
  PRESENTATION_TABLE, PRESENTATION_CARD, HEADER_COUNT, HEADER_DATE } from '../Query/common';
import { slideUnixTimeWindowToCurrentTime } from '../../util';

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
    deleteDashboardUnit,
    updateDashboardUnit,
  }, dispatch);
}

class DashboardUnit extends Component {
  constructor(props) {
    super(props);

    this.state = {
      loading: false,
      presentation: null,

      title: null,
      editTitle: false
    }
  }

  showLineChartLegend(result) {
    let isMultiGroupBy = result.headers.length > 3;
    
    let uniqueGroups = [];
    let countIndex = result.headers.indexOf(HEADER_COUNT);
    let dateIndex = result.headers.indexOf(HEADER_DATE);
    for(let r=0; r<result.rows.length; r++) {
      for (let c=0; c<result.rows[r].length; c++) {
        if (c != countIndex 
          && c != dateIndex 
          && uniqueGroups.indexOf(result.rows[r][c]) == -1)
            uniqueGroups.push(result.rows[r][c]);
      }
    }

    if (uniqueGroups.length < LINE_LEGEND_DISPLAY_LIMIT && isMultiGroupBy) return false;
    if (uniqueGroups.length > LINE_LEGEND_DISPLAY_LIMIT) return false;

    return true;
  }

  getUnitBackground() {
    let cardIndex = this.props.cardIndex;
    let poolLength = CARD_BACKGROUNDS.length;
    return CARD_BACKGROUNDS[cardIndex % poolLength];
  }

  setPresentation(result) {
    let presentation = null;
    if (this.props.data.presentation === PRESENTATION_BAR) {
      presentation = <BarChart queryResult={result} legend={false} />
    }

    if (this.props.data.presentation === PRESENTATION_LINE) {
      presentation = <LineChart legend={this.showLineChartLegend(result)} queryResult={result} />
    }

    if (this.props.data.presentation === PRESENTATION_TABLE) {
      presentation = <TableChart queryResult={result} />
    }

    if (this.props.data.presentation == PRESENTATION_CARD) {
      presentation = <TableChart noHeader card queryResult={result} />
    }

    this.setState({ presentation: presentation });
  }

  execQuery() {
    this.setState({ loading: true });
    
    let query = this.props.data.query;
    if (query && query.ovp) {
      let newPeriod = slideUnixTimeWindowToCurrentTime(query.fr, query.to);
      query.fr = newPeriod.from;
      query.to = newPeriod.to;
    }

    runQuery(this.props.currentProjectId, query)
      .then((r) => {
        this.setState({ loading: false });
        this.setPresentation(r.data);
      })
      .catch(console.error);
  }

  componentWillMount() {
    this.execQuery();
  }

  present() {
    if (this.state.loading)
      return <Loading paddingTop='12%' />;
    
    return this.state.presentation;
  }

  getCardBodyStyleByProps() {
    let style = { padding: '1.5rem 1.5rem', height: '300px' };

    if (this.props.data.presentation === PRESENTATION_TABLE) {
      let changes = { padding: '0', 'overflowX': 'scroll' };
      style = { ...style, ...changes };
    }

    if (this.props.data.presentation === PRESENTATION_CARD) {
      style.height = '130px';
      style.padding = '0';
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
      borderRadius: '4px',
      marginRight: '6px'
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

  getTitleStyle() {
    if (!this.props.editDashboard) return null;

    return { 
      maxWidth: this.isCard() ? '180px' : null, display: 'inline-block' 
    }
  }

  // Todo: Avoid execQuery on position change by
  // moving the query result to ParentComponent (dashboard).
  componentDidUpdate(prevProps) {
    if (prevProps.data.id != this.props.data.id) {
      this.execQuery();
    }
  }

  render() {
    return (
      <Card className='fapp-dunit' style={this.getCardStyleByProps()}>
        <CardHeader style={this.getCardHeaderStyleByProps()}>
          <div style={{ textAlign: 'right', marginTop: '-10px', marginRight: '-18px', height: '18px' }}>
            <strong onClick={this.delete} style={{ fontSize: '15px', cursor: 'pointer', padding: '0 10px', color: this.isCard() ? '#FFF' : '#AAA' }} hidden={!this.props.editDashboard}>x</strong>
          </div>

          <div style={{ marginTop: '-5px' }} hidden={!this.showTitle()}>
            <div className='fapp-overflow-dot' style={this.getTitleStyle()}> <strong>{ this.getTitle() }</strong> </div>
            <button style={this.getInlineButtonStyle()} onClick={this.editTitle} hidden={!this.props.editDashboard}><i className='icon-pencil'></i></button>
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
          { this.present() }
        </CardBody>
      </Card>
    );
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(DashboardUnit);