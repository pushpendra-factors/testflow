import React, {Component} from 'react';
import { Formik, Form, Field, ErrorMessage } from 'formik';
import { Alert, Container, Input, InputGroup, InputGroupAddon, InputGroupText, Button, Row, Col, CardGroup, CardBody, Card} from 'reactstrap';
import { bindActionCreators } from 'redux';
import { connect } from 'react-redux';
import {forgotPassword} from "../../../actions/agentActions";
import * as yup from 'yup';
import  { InvalidEmail, MissingEmail} from '../ValidationMessages';

const mapDispatchToProps = dispatch => {
    return bindActionCreators({forgotPassword}, dispatch)
}

class ForgotPassword extends Component {
    constructor(props){
        super(props);
        this.state = {
            forgotPasswordPerformed: false,
            agentEmail:''
        }
    }
    renderForgotPasswordForm = () => {
        if(this.state.forgotPasswordPerformed){
            return
        }
        return (
            <Formik
                initialValues={{email:''}}
                validationSchema = {yup.object().shape({
                    email: yup.string().email(InvalidEmail).required(MissingEmail)
                })}
                onSubmit={(values, {setSubmitting})=>{                    
                    this.props.forgotPassword(values.email)
                    .then(()=>{
                        setSubmitting(false);
                        this.setState({forgotPasswordPerformed: true, agentEmail: values.email });
                    })
                    .catch(()=>{
                        setSubmitting(false);
                    });                             
                }}
            >
                {({isSubmitting, touched})=> (
                    <Form noValidate>
                        <h1>Forgot Password</h1>
                        <p>We'll send you a link to create a new password.</p>
                        <InputGroup className="mb-3">
                            <InputGroupAddon addonType="prepend">
                                <InputGroupText>
                                <i className="icon-lock"></i>
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
                                <Button color="primary" className="px-4" type="submit" disabled={isSubmitting}>
                                    Send Reset Link
                                </Button>
                            </Col>
                        </Row>
                    </Form>
                )}
            </Formik>
        )
    }
    
    renderMessage = () => {
        if(!this.state.forgotPasswordPerformed){
            return
        }
        return (
            <Alert color="success">
            <h4 className="alert-heading">Check your inbox</h4>              
              <hr />
              <p>
                An email has been sent to {this.state.agentEmail}.
                Please follow the instructions to reset your password.
              </p>        
            </Alert>
          )
    }

    render(){
        return (
            <div className="app flex-row align-items-center">
                <Container>            
                    <Row className="justify-content-center">
                        <Col md="6">
                            <CardGroup>
                                <Card md="6">
                                    <CardBody>
                                        { this.renderForgotPasswordForm() }
                                        { this.renderMessage() }
                                    </CardBody>
                                </Card>
                            </CardGroup>
                        </Col>
                    </Row>
                </Container>
            </div>

        )    
    }
}

export default connect(null, mapDispatchToProps)(ForgotPassword);