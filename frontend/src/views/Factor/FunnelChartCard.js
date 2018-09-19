import React, { Component } from 'react';
import { Doughnut } from 'react-chartjs-2';
import {
  Card,
  CardBody,
  CardHeader,
  Col,
  Container,
  Row,
} from 'reactstrap';
import FunnelArrow from './FunnelArrow.js';

const funnelLabelRowStyle = {
  marginBottom: '20px',
};
const eventTextStyle = {
  fontSize: '0.76562rem',
  color: '#73818f',
  textAlign: 'center',
  margin: 'auto'
};
const arrowStyle = {
  maxWidth: '80%',
  maxHeight: '80%',
  position: 'relative',
  top: '8%'
}
const chartOptions = {
  layout: {
    padding: {
      left: -50,
      right: -50,
      top: -50,
      bottom: 10
    },
  },
};
const arrowColor = "#73818f"

class FunnelChartCard extends Component {
  render() {
    var chartData = this.props.chartData;
    return (
      <Card>
      <CardHeader>
      Funnel
      </CardHeader>
      <CardBody>
      <Row noGutters={true}>
      <Col xs={{ size: '2' }}>
      <Doughnut data={chartData} options={chartOptions}/>
      </Col>
      <Col xs={{ size: '1' }}>
      <div style={arrowStyle}><FunnelArrow color={arrowColor} uid={1} /></div>
      </Col>
      <Col xs={{ size: '2' }}>
      <Doughnut data={chartData} options={chartOptions} />
      </Col>
      </Row>
      <Row style={funnelLabelRowStyle} noGutters={true}>
      <Col xs={{ size: '2' }}>
      <div style={eventTextStyle}>{'PublicMessageSent (20350)'}</div>
      </Col>
      <Col xs={{ size: '2', offset: '1' }}>
      <div style={eventTextStyle}>{'PublicMessageSent (20350)'}</div>
      </Col>
      </Row>

      <Row noGutters={true}>
      <Col xs={{ size: '2' }}>
      <Doughnut data={chartData} options={chartOptions} />
      </Col>
      <Col xs={{ size: '1' }}>
      <div style={arrowStyle}><FunnelArrow color={arrowColor} uid={2} /></div>
      </Col>
      <Col xs={{ size: '2' }}>
      <Doughnut data={chartData} options={chartOptions}/>
      </Col>
      <Col xs={{ size: '1' }}>
      <div style={arrowStyle}><FunnelArrow color={arrowColor} uid={3} /></div>
      </Col>
      <Col xs={{ size: '2' }}>
      <Doughnut data={chartData} options={chartOptions} />
      </Col>
      </Row>
      <Row style={funnelLabelRowStyle} noGutters={true}>
      <Col xs={{ size: '2' }}>
      <div style={eventTextStyle}>{'PublicMessageSent (20350)'}</div>
      </Col>
      <Col xs={{ size: '2', offset: '1' }}>
      <div style={eventTextStyle}>{'PublicMessageSent (20350)'}</div>
      </Col>
      <Col xs={{ size: '2', offset: '1' }}>
      <div style={eventTextStyle}>{'PublicMessageSent (20350)'}</div>
      </Col>
      </Row>

      </CardBody>
      </Card>
    );
  }
}

export default FunnelChartCard;
