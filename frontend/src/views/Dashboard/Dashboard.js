import React, { Component } from 'react';
import { Row, Col, Card, CardHeader, CardBody } from 'reactstrap';
import {Bar} from 'react-chartjs-2';

const data = {
  labels: ['January', 'February', 'March', 'April', 'May', 'June', 'July', 'January', 'February', 'March', 'April', 'May', 'June', 'July'],
  datasets: [
    {
      label: 'My First dataset',
      backgroundColor: 'rgba(255,99,132,0.2)',
      borderColor: 'rgba(255,99,132,1)',
      borderWidth: 1,
      hoverBackgroundColor: 'rgba(255,99,132,0.4)',
      hoverBorderColor: 'rgba(255,99,132,1)',
      data: [65, 59, 80, 81, 56, 55, 40, 65, 59, 80, 81, 56, 55, 40]
    }
  ]
};


class Dashboard extends Component {
  constructor(props) {
      super(props);

      this.state = {}
  }

  render() {
    return (
      <div className='fapp-content' style={{marginLeft: '1rem', marginRight: '1rem'}}>
        <Row class="fapp-select">
          <Col md={{ size: 6 }} style={{padding: '0 15px'}}>
            <Card className='fapp-bordered-card' style={{marginTop: '15px'}}>
              <CardHeader>
                <strong>Chart Title</strong>
              </CardHeader>
              <CardBody style={{padding: '1.5rem 0.5rem'}}>
                <div style={{height: '250px'}}>
                  <Bar
                    data={data}
                    options={{
                      maintainAspectRatio: false
                    }}
                  />
                </div>
              </CardBody>
            </Card>
          </Col>
          <Col md={{ size: 6 }} style={{padding: '0 15px'}}>
            <Card className='fapp-bordered-card' style={{marginTop: '15px'}}>
              <CardHeader>
                <strong>Chart Title</strong>
              </CardHeader>
              <CardBody style={{padding: '1.5rem 0.5rem'}}>
                <div style={{height: '250px'}}>
                  <Bar
                    data={data}
                    options={{
                      maintainAspectRatio: false
                    }}
                  />
                </div>
              </CardBody>
            </Card>
          </Col>
          <Col md={{ size: 6 }} style={{padding: '0 15px'}}>
            <Card className='fapp-bordered-card' style={{marginTop: '15px'}}>
              <CardHeader>
                <strong>Chart Title</strong>
              </CardHeader>
              <CardBody style={{padding: '1.5rem 0.5rem'}}>
                <div style={{height: '250px'}}>
                  <Bar
                    data={data}
                    options={{
                      maintainAspectRatio: false
                    }}
                  />
                </div>
              </CardBody>
            </Card>
          </Col>
        </Row>
      </div>
    );
  }
}

export default Dashboard;