import React, { Component } from 'react';
import {
  Card,
  CardBody,
  CardHeader,
  CardTitle,
} from 'reactstrap';

import BarChart from '../Query/BarChart';

class BarChartCard extends Component {
  render() {
    var chartData = this.props.chartData;
    var chart = <BarChart bars={chartData} />;

    const explanations = chartData.explanations.map((explainText) =>
      <CardTitle className="fapp-chart-card-title">{explainText}</CardTitle>
    );

    return (
      <Card>
      <CardHeader className="fapp-chart-card-header">
        {chartData.header}
      </CardHeader>
      {explanations}
      <CardBody className="fapp-chart-card-body">
      <div className="chart-wrapper">
      {chart}
      </div>
      </CardBody>
      </Card>
    );
  }
}

export default BarChartCard;
