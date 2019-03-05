import React, {Component} from 'react';
import { Formik, Form, Field, ErrorMessage } from 'formik';
import { Container, Input, InputGroup, InputGroupAddon, InputGroupText, Button, Row, Col, CardGroup, CardBody, Card} from 'reactstrap';
import { bindActionCreators } from 'redux';
import { connect } from 'react-redux';
import {setPassword} from "../../../actions/agentActions";
import * as yup from 'yup';
import  {  MissingPassword, PasswordsDoNotMatch, PasswordMinEightChars } from '../ValidationMessages';

const mapDispatchToProps = dispatch => {
    return bindActionCreators({setPassword}, dispatch)
}

class SetPassword extends Component {
    render(){
        return (
            <div className="app flex-row align-items-center">
                <Container>            
                <Row className="justify-content-center">
                <Col md="6">
                <CardGroup>
                <Card md="6">
                <CardBody>
                <Formik
                    initialValues={{password:'', confirmPassword:''}}
                    validationSchema = {
                        yup.object().shape({
                            password: yup.string().required(MissingPassword).min(8, PasswordMinEightChars),
                            confirmPassword: yup.string().required().oneOf([yup.ref('password'),null], PasswordsDoNotMatch)
                        })
                    }
                    onSubmit={(values, {setSubmitting})=>{                            
                        const hash = window.location.hash;
                        let paramToken = "token=";
                        let token = hash.substring(hash.indexOf(paramToken)+paramToken.length);
                        this.props.setPassword(values.password, token)
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
                            <h1>Reset Password</h1>
                            <p>After updating, you can login to Factors using this password.</p>
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
                            <InputGroup className="mb-4">
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
                                    <Button color="primary" className="px-4" type="submit" disabled={isSubmitting}>
                                        Update
                                    </Button>
                                </Col>
                            </Row>
                            
                        </Form>
                    )}
                    
                </Formik>
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

export default connect(null, mapDispatchToProps)(SetPassword);