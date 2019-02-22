import React, { Component } from 'react';
import { Button, Card, CardBody, Col, Container, Form, Input, InputGroup, InputGroupAddon, InputGroupText, Row } from 'reactstrap';
import { verify } from "../../../actions/agentActions";
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';

const mapDispatchToProps = dispatch => {
  return bindActionCreators({ verify }, dispatch);
}

class Verify extends Component {
  
  constructor(props){
    super(props);
  }
  
  handleSubmit = (event) => {
    
    event.preventDefault();

    const target = event.target;
    const firstName = target.firstName.value;
    const lastName = target.lastName.value;
    const password = target.password.value;
    const confirm_password = target.confirm_password.value;
    
    if(password == "" || confirm_password == "" || password != confirm_password){
      return 
    }

    const hash = window.location.hash;
    var paramToken = "token=";
    const token = hash.substring(hash.indexOf(paramToken)+paramToken.length);
  
    this.props.verify(firstName, lastName, password,token);
  }

  render() {
    return (
      <div className="app flex-row align-items-center">
        <Container>
          <Row className="justify-content-center">
            <Col md="6">
              <Card className="mx-4">
                <CardBody className="p-4">
                  <Form onSubmit={this.handleSubmit}>          
                    <h1>Verify</h1>
                    <p className="text-muted">Enter your details</p>                    
                    <InputGroup className="mb-3">
                      <InputGroupAddon addonType="prepend">
                        <InputGroupText>FirstName</InputGroupText>
                      </InputGroupAddon>
                      <Input type="text" name="firstName" placeholder="FirstName" autoComplete="firstName" required/>
                    </InputGroup>                  
                    <InputGroup className="mb-3">
                      <InputGroupAddon addonType="prepend">
                        <InputGroupText>LastName</InputGroupText>
                      </InputGroupAddon>
                      <Input type="text" name="lastName" placeholder="LastName" autoComplete="lastName" required/>
                    </InputGroup>      
                    <InputGroup className="mb-4">
                        <InputGroupAddon addonType="prepend">
                          <InputGroupText>
                            <i className="icon-lock"></i>
                          </InputGroupText>
                        </InputGroupAddon>
                        <Input type="password" name="password" placeholder="Password" autoComplete="current-password" required/>
                      </InputGroup>
                      <InputGroup className="mb-4">
                        <InputGroupAddon addonType="prepend">
                          <InputGroupText>
                            <i className="icon-lock"></i>
                          </InputGroupText>
                        </InputGroupAddon>
                        <Input type="password" name="confirm_password" placeholder="Confirm Password" autoComplete="current-password" required/>
                      </InputGroup>            
                    <Button color="success" block>Verify</Button>
                  </Form>
                </CardBody>
              </Card>
            </Col>
          </Row>
        </Container>
      </div>
    );
  }
}

export default connect(null, mapDispatchToProps)(Verify);
