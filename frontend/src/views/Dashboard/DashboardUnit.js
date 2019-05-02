import React, { Component } from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import { Col, Card, CardHeader, CardBody } from 'reactstrap';

import { runQuery } from '../../actions/projectsActions';
import Loading from '../../loading';
import BarChart from '../Query/BarChart';
import LineChart from '../Query/LineChart';
import TableChart from '../Query/TableChart';
import { PRESENTATION_BAR, PRESENTATION_LINE, PRESENTATION_TABLE } from '../Query/common';

const mapStateToProps = store => {
  return {
    currentProjectId: store.projects.currentProjectId,
  };
}

const mapDispatchToProps = dispatch => {
  return bindActionCreators({ 
    runQuery
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

  setPresentation(result) {
    let presentation = null;
    if (this.props.data.presentation === PRESENTATION_BAR) {
      presentation = <BarChart queryResult={result} legend={false} />
    }

    if (this.props.data.presentation === PRESENTATION_LINE) {
      presentation = <LineChart queryResult={result} />
    }

    if (this.props.data.presentation === PRESENTATION_TABLE) {
      presentation = <TableChart queryResult={result} />
    }

    this.setState({ presentation: presentation });
  }

  componentWillMount() {
    this.setState({ loading: true });
    runQuery(this.props.currentProjectId, this.props.data.query)
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

  cardStyleByPresentation() {
    let style = { padding: '1.5rem 0.5rem', height: '300px' };

    if (this.props.data.presentation === PRESENTATION_TABLE) {
      let changes = { padding: '0', 'overflowX': 'scroll' };
      style = { ...style, ...changes };
    }
       
    return style;
  }

  render() {
    let data = this.props.data;

    return (
      <Col md={{ size: 6 }}  style={{padding: '0 15px'}}>
        <Card className='fapp-bordered-card' style={{marginTop: '15px'}}>
          <CardHeader>
            <strong>{ data.title }</strong>
          </CardHeader>
          <CardBody style={this.cardStyleByPresentation()}>
            { this.present() }
          </CardBody>
        </Card>
      </Col>
    );
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(DashboardUnit);