import React, { Component } from 'react';
import UUID from 'node-uuid';
import { Doughnut } from 'react-chartjs-2';
import { Col, Row, Card, CardBody } from 'reactstrap';
import FunnelArrow from '../Factor/FunnelArrow.js';

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

const eventTextStyle = {
  fontSize: '0.76562rem',
  color: '#73818f',
  textAlign: 'center',
  margin: 'auto'
};

const arrowStyle = {
  width: '45px',
  maxHeight: '60%',
  position: 'relative',
  top: '8%'
}

const getEventColSize = function(data) {
  if (data.length == 2) return '4';
  return data.length > 3 ? '2' : '3';
}

const getGraphColClassName = function(data) {
  if (data.length == 2) return 'col-4';
  return data.length <= 3 ? 'col-3' : null;
}

const getMarginLeftForGraphParent = function(data) {
  if (data.length == 3) return '6%';
  return data.length == 2 ? '20%' : null;
}

const SmallFunnel = (props) => {
  var funnelData = props.data.funnels;

  var graphCols = [];
  var eventCols = [];
  for (var i = 0; i < funnelData.length; i++) {
    var nodeColor = '#20a8d8';
    var arrowColor = '#73818f';
    // custom arrow color by type.
    if (funnelData[i].node_type === "positive") {
      nodeColor = '#4dbd74';
      arrowColor = '#4dbd74';
    } else if (funnelData[i].node_type === "negative") {
      nodeColor = '#f86c6b';
      arrowColor = '#f86c6b';
    }

    // Empty labels.
    let labels = funnelData[i].data.map(() => (''));
    
    var donutGraphData = {
      labels: labels,
      datasets: [
        { 
          data: funnelData[i].data,
          backgroundColor: [
            nodeColor,
            '#C8CED3'
          ],
          hoverBackgroundColor: [],
        }],
      };

    var conversionString;
    if (!!funnelData[i].conversion_percent) {
      conversionString = funnelData[i].conversion_percent + "%";
    }

    graphCols.push(
      <div className={getGraphColClassName(funnelData)} style={{ maxWidth: funnelData.length > 3 ? '17%' : null }}  key={i*4}>
        <Doughnut data={donutGraphData} options={chartOptions}/>
      </div>
    );

    if (i < funnelData.length - 1) {
      graphCols.push(
        <div style={arrowStyle}>
          <FunnelArrow color={arrowColor} conversionString={conversionString} uid={UUID.v4()} />
        </div>
      );
    }

    if (i == 0) {
      eventCols.push(
        <Col xs={{ size: getEventColSize(funnelData) }} key={i*4 + 2}>
          <div style={eventTextStyle}>{funnelData[i].event}</div>
        </Col>
      );
    } else {
      eventCols.push(
        <Col xs={{ size: getEventColSize(funnelData) }} style={{ marginLeft: '48px' }} key={i*4 + 3}>
          <div style={eventTextStyle}>{funnelData[i].event}</div>
        </Col>
      );
    }
  }

  let totalConvStr = props.data.totalConversion + '%';
  return (
    <Col md='12'>
      <div style={{ textAlign: 'center', marginBottom: '40px' }}>  
        <span style={{fontWeight: '600', color: '#777', fontSize: '15px'}}> Total Conversion Rate: { totalConvStr } </span> 
      </div>
      <div style={{ marginLeft: getMarginLeftForGraphParent(funnelData) }}>
        <Row noGutters={true}>{graphCols}</Row>
        <Row noGutters={true}>{eventCols} </Row>
      </div>
    </Col>
  );
  
}

export default SmallFunnel;