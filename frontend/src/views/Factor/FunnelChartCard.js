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
  maxWidth: '60%',
  maxHeight: '60%',
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
  buildFunnelUI(funnelData) {
    var graphCols = [];
    var eventCols = [];
    for (var i = 0; i < funnelData.length; i++) {
      var donutGraphData = {
        labels: [
        ],
        datasets: [
          {
            data: funnelData[i].data,
            backgroundColor: [
              '#36A2EB',
              '#C8CED3'
            ],
            hoverBackgroundColor: [
            ],
          }],
        };

        graphCols.push(
          <Col xs={{ size: '2' }} key={i}>
            <Doughnut data={donutGraphData} options={chartOptions}/>
          </Col>);
        if (i < funnelData.length - 1) {
          graphCols.push(
            <Col xs={{ size: '1' }} key={i}>
              <div style={arrowStyle}><FunnelArrow color={arrowColor} uid={1} /></div>
            </Col>);
        }
        if (i == 0) {
          eventCols.push(
            <Col xs={{ size: '2' }} key={i}>
            <div style={eventTextStyle}>{funnelData[i].event}</div>
            </Col>
          );
        } else {
          eventCols.push(
            <Col xs={{ size: '2', offset: '1'}} key={i}>
            <div style={eventTextStyle}>{funnelData[i].event}</div>
            </Col>
          );
        }
    }
    return [graphCols, eventCols];
  }
  render() {
    var chartData = this.props.chartData;
    var baseFunnelGraphCols, baseFunnelEventCols;
    [baseFunnelGraphCols, baseFunnelEventCols] = this.buildFunnelUI(chartData.baseFunnelData);
    var funnelGraphCols, funnelEventCols;
    [funnelGraphCols, funnelEventCols] = this.buildFunnelUI(chartData.funnelData);

    return (
      <Card>
      <CardHeader>
      {chartData.header}
      </CardHeader>
      <CardBody>
      <Row noGutters={true}>
      {
        baseFunnelGraphCols
      }
      </Row>

      <Row style={funnelLabelRowStyle} noGutters={true}>
      {
        baseFunnelEventCols
      }
      </Row>

      <Row noGutters={true}>
      {
        funnelGraphCols
      }
      </Row>

      <Row style={funnelLabelRowStyle} noGutters={true}>
      {
        funnelEventCols
      }
      </Row>
      </CardBody>
      </Card>
    );
  }
}

export default FunnelChartCard;
