import React, { Component } from 'react';
import { Button, Card, CardBody, CardGroup, Col, Container, Input, InputGroup, InputGroupAddon, InputGroupText, Row } from 'reactstrap';
import { login } from "../../../actions/agentActions";
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import { Redirect } from 'react-router-dom';
import * as yup from 'yup';
import { Formik, Form, Field, ErrorMessage } from 'formik';
import  { InvalidEmail, MissingEmail, MissingPassword } from '../ValidationMessages';


const mapDispatchToProps = dispatch => {
  return bindActionCreators({ login }, dispatch);
}

const mapStateToProps = store => {
  return { isLoggedIn: store.agents.isLoggedIn}
}

class Login extends Component {
  constructor(props){
    super(props);
  }
  
  redirectToForgotPassword = (event) => {
    event.preventDefault();    
    this.props.history.push("/forgotpassword");
  }

  renderLoginForm = () => {
    return (
      <Formik
        initialValues={{email:'', password:''}}
        validationSchema = {
            yup.object().shape({                
                email: yup.string().email(InvalidEmail).required(MissingEmail),
                password: yup.string().required(MissingPassword),
            })
        }
        onSubmit={(values, {setSubmitting})=>{
          this.props.login(values.email, values.password)
          .then(()=>{
            setSubmitting(false);            
          })
          .catch(()=>{
            setSubmitting(false);
          });                             
        }}
      >
        {({isSubmitting, touched})=> (

          <Form noValidate>
              <h1>Login</h1>
              <p className="text-muted">Sign In to your account</p>
              <InputGroup className="mb-3">
                <InputGroupAddon addonType="prepend">
                    <InputGroupText>
                      @
                    </InputGroupText>
                </InputGroupAddon>                                    
                  <Input tag={Field} type="email" name="email" placeholder="email"/>
                  {touched.email &&
                    <ErrorMessage name="email">
                        {msg => <div style={{color:'red'}}>{msg}</div>}    
                    </ErrorMessage>
                  }                  
              </InputGroup>              
              <InputGroup className="mb-4">
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
              <Row>
                  <Col xs="6">
                      <Button color="success" className="px-4" type="submit" disabled={isSubmitting}>
                          Login
                      </Button>
                  </Col>
                  <Col xs="6" className="text-right">
                    <Button color="link" onClick={this.redirectToForgotPassword} className="px-0">Forgot password?</Button>
                  </Col> 
              </Row>
              
          </Form>
        )}  
      </Formik>
    )
  }

  render() {
    if(this.props.isLoggedIn){
      return <Redirect to='/' />
    }

    return (
      <div className="app flex-row align-items-center">
        <Container>
          <Row className="justify-content-center">
            <Col md="6">
              <CardGroup>
                <Card className="p-4">
                  <CardBody>
                    {this.renderLoginForm()}
                  </CardBody>
                </Card>
              </CardGroup>
            </Col>
          </Row>
        </Container>
      </div>
    );
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(Login);
