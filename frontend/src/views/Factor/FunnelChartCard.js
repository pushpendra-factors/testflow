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

const funnelRowStyle = {
  marginBottom: '20px',
};
const eventTextStyle = {
  fontSize: '0.76562rem',
  color: '#73818f',
  textAlign: 'center',
  margin: 'auto'
};
const arrowStyle = {
  maxWidth: '100%',
  maxHeight: '100%',
  position: 'relative',
  top: '15%'
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

class FunnelChartCard extends Component {
  render() {
    var chartData = this.props.chartData;
    return (
      <Card>
      <CardHeader>
      Funnel
      </CardHeader>
      <CardBody>
      <Row style={funnelRowStyle} noGutters={true}>
      <Col xs={{ size: '4' }}>
      <Doughnut data={chartData} options={chartOptions}/>
      <div style={eventTextStyle}>{'PublicMessageSent (20350)'}</div>
      </Col>
      <Col xs={{ size: '1' }}>
      <div style={arrowStyle}><FunnelArrow color={"#4dbd74"} uid={1} /></div>
      </Col>
      <Col xs={{ size: '4' }}>
      <Doughnut data={chartData} options={chartOptions} />
      <div style={eventTextStyle}>{'PublicMessage'}</div>
      </Col>
      </Row>

      <Row style={funnelRowStyle} noGutters={true}>
      <Col xs={{ size: '3' }}>
      <Doughnut data={chartData} options={chartOptions} />
      <div style={eventTextStyle}>{'PublicMessage'}</div>
      </Col>
      <Col xs={{ size: '1' }}>
      <div style={arrowStyle}><FunnelArrow color={"#c8ced3"} uid={2} /></div>
      </Col>
      <Col xs={{ size: '3' }}>
      <Doughnut data={chartData} options={chartOptions}/>
      <div style={eventTextStyle}>{'PublicMessage'}</div>
      </Col>
      <Col xs={{ size: '1' }}>
      <div style={arrowStyle}><FunnelArrow color={"#f86c6b"} uid={3} /></div>
      </Col>
      <Col xs={{ size: '3' }}>
      <Doughnut data={chartData} options={chartOptions} />
      <div style={eventTextStyle}>{'PublicMessage'}</div>
      </Col>
      </Row>

      </CardBody>
      </Card>
    );
  }
}

export default FunnelChartCard;
