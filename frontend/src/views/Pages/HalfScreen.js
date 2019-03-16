import React from 'react';
import LeftSvg from '../../assets/img/brand/factors-left.svg';
import { 
  Container,
  Row,
  Col,
  Card,
  CardBody 
} from 'reactstrap';

const HalfScreen = (props) => {
  return (
    <Container fluid>
      <Row style={{backgroundColor: '#F7F8FD'}}>
        <Col md='6' style={{height: '100vh', padding: '0'}}>
          <img src={LeftSvg} height='100%'/>
        </Col>
        <Col md='6'>
          <Card style={{marginTop: '23%', width: '65%', marginLeft: '15%'}} className="p-4 fapp-block-shadow">
            <CardBody>
              { props.renderForm() }
            </CardBody>
          </Card>
        </Col>
      </Row>
    </Container>
  );
}

export default HalfScreen;