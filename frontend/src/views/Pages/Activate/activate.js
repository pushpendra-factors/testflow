import React, { Component } from 'react';
import { Button, Card, CardBody, Col, Container, InputGroup, Input, InputGroupAddon, InputGroupText, Row } from 'reactstrap';
import { activate } from "../../../actions/agentActions";
import { connect } from 'react-redux';
import { Formik, Form, Field, ErrorMessage } from 'formik';
import { bindActionCreators } from 'redux';
import * as yup from 'yup';
import  { MissingFirstname, PasswordsDoNotMatch, MissingPassword, PasswordMinEightChars} from '../ValidationMessages';
const mapDispatchToProps = dispatch => {
  return bindActionCreators({ activate }, dispatch);
}

class Activate extends Component {
  
  renderActivateForm = () => {
    return (
      <Formik
        initialValues={{firstName:'', lastName:'', password:'', confirmPassword:''}}
        validationSchema = {
            yup.object().shape({
                firstName: yup.string().required(MissingFirstname),
                lastName: yup.string(),
                password: yup.string().required(MissingPassword).min(8, PasswordMinEightChars),
                confirmPassword: yup.string().required().oneOf([yup.ref('password'),null], PasswordsDoNotMatch)
            })
        }
        onSubmit={(values, {setSubmitting})=>{
          const hash = window.location.hash;
          var paramToken = "token=";
          const token = hash.substring(hash.indexOf(paramToken)+paramToken.length);
        
          this.props.activate(values.firstName, values.lastName, values.password, token)
          .then(()=>{
            setSubmitting(false);
            this.props.history.push("/login");
          })
          .catch(()=>{
            setSubmitting(false);
          });                             
        }}
      >
        {({isSubmitting, touched})=> (

          <Form noValidate>
              <h1>Activate</h1>    
              <p className="text-muted">Please enter following details to activate your account.</p>
              <InputGroup className="mb-4">
                <InputGroupAddon addonType="prepend">
                    <InputGroupText>
                      Firstname  
                    </InputGroupText>
                </InputGroupAddon>                  
                  <Input tag={Field} type="text" name="firstName" placeholder="Firstname"/>
                  {touched.firstName &&
                    <ErrorMessage name="firstName">
                        {msg => <div style={{color:'red'}}>{msg}</div>}    
                    </ErrorMessage>
                  }                  
              </InputGroup>
              <InputGroup className="mb-3">
                <InputGroupAddon addonType="prepend">
                    <InputGroupText>
                      Lastname  
                    </InputGroupText>
                </InputGroupAddon>                  
                  <Input tag={Field} type="text" name="lastName" placeholder="Lastname"/>
                  {touched.lastName &&
                    <ErrorMessage name="lastName">
                        {msg => <div style={{color:'red'}}>{msg}</div>}    
                    </ErrorMessage>
                  }                  
              </InputGroup>
              <InputGroup className="mb-3">
                <InputGroupAddon addonType="prepend">
                    <InputGroupText>
                    <i className="icon-lock"></i>
                    </InputGroupText>
                </InputGroupAddon>
                <Input tag={Field} type="password" name="password" placeholder="Password"/>
                {touched.password &&
                    <ErrorMessage name="password">
                        {msg => <div style={{color:'red'}}>{msg}</div>}    
                    </ErrorMessage>
                }
              </InputGroup>
              <InputGroup className="mb-3">
                <InputGroupAddon addonType="prepend">
                    <InputGroupText>
                    <i className="icon-lock"></i>
                    </InputGroupText>
                </InputGroupAddon>                                
                <Input tag={Field} type="password" name="confirmPassword" placeholder="Renter Password"/>
                    {   touched.confirmPassword &&
                        <ErrorMessage name="confirmPassword">
                        {msg => <div style={{color:'red'}}>{msg}</div>}    
                        </ErrorMessage>
                    }
              </InputGroup>
              <Row>
                  <Col xs="6">
                      <Button color="success" className="px-4" type="submit" disabled={isSubmitting}>
                          Activate
                      </Button>
                  </Col>
              </Row>
              
          </Form>
        )}  
      </Formik>
    )
  }

  render() {
    return (
      <div className="app flex-row align-items-center">
        <Container>
          <Row className="justify-content-center">
            <Col md="6">
              <Card className="mx-4">
                <CardBody className="p-4">
                  {this.renderActivateForm()}                  
                </CardBody>
              </Card>
            </Col>
          </Row>
        </Container>
      </div>
    );
  }
}

export default connect(null, mapDispatchToProps)(Activate);
