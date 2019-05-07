import React, { Component } from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import { Col, Card, CardHeader, CardBody } from 'reactstrap';

import { runQuery } from '../../actions/projectsActions';
import { deleteDashboardUnit } from '../../actions/dashboardActions';
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
    deleteDashboardUnit
  }, dispatch);
}

class DashboardUnit extends Component {
  constructor(props) {
    super(props);

    this.state = {
      loading: false,
      presentation: null,
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

  componentWillMount() {
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

  getCardHeaderStyleByProps() {
    if (this.props.data.presentation !== PRESENTATION_CARD) return null;
    let style = {};
    style.textAlign = 'center';
    style.background = this.getUnitBackground();
    style.color = CARD_FONT_COLOR;
    return style;
  }

  getColSizeByProps() {
    return this.props.card ? 3 : 6;
  }

  getCardStyleByProps() {
    let style = { marginBottom: '30px' };
    if (this.props.data.presentation === PRESENTATION_CARD) {
      style.border = 'none';
    }

    return style;
  }

  delete = () => {
    let unit = this.props.data;
    this.props.deleteDashboardUnit(unit.project_id, unit.dashboard_id, unit.id);
  }

  render() {
    let data = this.props.data;
    let isCard = this.props.data.presentation === PRESENTATION_CARD;

    return (
      <Col md={{ size: this.getColSizeByProps() }}  style={{padding: '0 15px'}}>
        <Card className='fapp-dunit' style={this.getCardStyleByProps()}>
          <CardHeader style={this.getCardHeaderStyleByProps()}>
            <div onClick={this.delete} style={{ textAlign: 'right', marginTop: '-10px', marginRight: '-18px', height: '18px', cursor: 'pointer' }}>
              <strong style={{ fontSize: '15px', padding: '0 10px', color: isCard ? '#FFF' : '#AAA' }} hidden={!this.props.showClose}>x</strong>
            </div>
            <div style={{ marginTop: isCard ? '-10px' : '-5px' }}><strong>{ data.title }</strong></div>
          </CardHeader>
          <CardBody style={this.getCardBodyStyleByProps()}>
            { this.present() }
          </CardBody>
        </Card>
      </Col>
    );
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(DashboardUnit);