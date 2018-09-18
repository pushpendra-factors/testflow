import React, { Component } from 'react';
import { Doughnut } from 'react-chartjs-2';
import {
  Card,
  CardBody,
  CardHeader,
  Col,
  Row,
} from 'reactstrap';

const funnelRowStyle = {
  marginBottom: '20px',
}

class FunnelChartCard extends Component {
  render() {
    var chartData = this.props.chartData;
    return (
      <Card>
      <CardHeader>
      Funnel
      </CardHeader>
      <CardBody>
      <Row style={funnelRowStyle}>
      <Col sm={{ size: 4 }}>
      <Col>
      <Doughnut data={chartData} />
      <span className='progress-group-text'>{'PublicMessage'}</span>
      </Col>
      </Col>
      <Col sm={{ size: 4 }}>
      <Doughnut data={chartData} />
      </Col>
      </Row>

      <Row>
      <Col sm={{ size: 4 }}>
      <Doughnut data={chartData} />
      </Col>
      <Col sm={{ size: 4 }}>
      <Doughnut data={chartData} />
      </Col>
      <Col sm={{ size: 4 }}>
      <Doughnut data={chartData} />
      </Col>
      </Row>
      </CardBody>
      </Card>
    );
  }
}

export default FunnelChartCard;
