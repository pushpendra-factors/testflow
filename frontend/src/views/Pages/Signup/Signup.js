import React, { Component } from 'react';
import { Formik, Form, Field, ErrorMessage } from 'formik';
import { Button, Card, CardBody, Col, Container, Alert, Input, InputGroup, InputGroupAddon, InputGroupText, Row } from 'reactstrap';
import { signup } from "../../../actions/agentActions";
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import * as yup from 'yup';
import  { InvalidEmail, MissingEmail } from '../ValidationMessages';

const mapDispatchToProps = dispatch => {
  return bindActionCreators({ signup }, dispatch);
}

class Signup extends Component {
  
  constructor(props){
    super(props);
    this.state = {
      signupPerformed: false,
      agentEmail:''
    }
  }
  
  renderSignupForm = () => {
    if (this.state.signupPerformed){
      return
    }

    return (
      <Formik
        initialValues={{email:''}}
        validationSchema = {
            yup.object().shape({
                email: yup.string().email(InvalidEmail).required(MissingEmail)
            })
        }
        onSubmit={(values, {setSubmitting})=>{            
            this.props.signup(values.email)
            .then(()=>{
                setSubmitting(false);
                this.setState({signupPerformed: true, agentEmail: values.email });
            })
            .catch(()=>{
                setSubmitting(false);
            });                             
        }}
      >
        {({isSubmitting, touched})=> (
          <Form noValidate>
              <h1>Signup</h1>
              <p className="text-muted">We'll send you a link to create a new Factors Account.</p>
              <InputGroup className="mb-4">
                <InputGroupAddon addonType="prepend">
                    <InputGroupText>
                      @
                    </InputGroupText>
                </InputGroupAddon>                  
                  <Input tag={Field} type="email" name="email" placeholder="Email"/>
                  {touched.email &&
                    <ErrorMessage name="email">
                        {msg => <div style={{color:'red'}}>{msg}</div>}    
                    </ErrorMessage>
                  }                  
              </InputGroup>                            
              <Row>
                  <Col xs="6">
                      <Button color="success" className="px-4" type="submit" disabled={isSubmitting}>
                          Create Account
                      </Button>
                  </Col>
              </Row>
              
          </Form>
        )}  
      </Formik>

    )
  }

  renderMessage = () => {
    if (!this.state.signupPerformed){
      return
    }
    return (
      <Alert color="success">
        <h4 className="alert-heading">Thanks for signing up!</h4>
        <hr />
        <p>
          A verification email has been sent to {this.state.agentEmail}.
          Please follow the instructions to activate your account.
        </p>        
      </Alert>
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
                  { this.renderSignupForm()}
                  {this.renderMessage()}
                </CardBody>
              </Card>
            </Col>
          </Row>
        </Container>
      </div>
    );
  }
}

export default connect(null, mapDispatchToProps)(Signup);
