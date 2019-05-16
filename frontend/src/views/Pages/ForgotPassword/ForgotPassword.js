import React, {Component} from 'react';
import { Formik, Form, Field, ErrorMessage } from 'formik';
import { Alert, Container, Input, InputGroup, InputGroupAddon, InputGroupText, Button, Row, Col, CardGroup, CardBody, Card} from 'reactstrap';
import { bindActionCreators } from 'redux';
import { connect } from 'react-redux';
import * as yup from 'yup';

import { forgotPassword } from "../../../actions/agentActions";
import  { InvalidEmail, MissingEmail } from '../ValidationMessages';
import HalfScreen from '../HalfScreen';
import SubmissionError from '../SubmissionError';


const mapDispatchToProps = dispatch => {
    return bindActionCreators({forgotPassword}, dispatch)
}

class ForgotPassword extends Component {
    constructor(props){
        super(props);
        this.state = {
            forgotPasswordPerformed: false,
            agentEmail: "",
            error: null,
        }
    }

    renderForgotPasswordForm = () => {
        if(!this.state.forgotPasswordPerformed){
            return (
                <Formik
                    initialValues={{ email: '' }}
                    validationSchema = {yup.object().shape({
                        email: yup.string().email(InvalidEmail).required(MissingEmail)
                    })}
                    onSubmit={(values, {setSubmitting}) => {                    
                        this.props.forgotPassword(values.email)
                        .then(() => {
                            setSubmitting(false);
                            this.setState({forgotPasswordPerformed: true, agentEmail: values.email });
                        })
                        .catch((msg) => {
                            setSubmitting(false);
                            this.setState({ error: msg });
                        });                             
                    }}
                >
                    {({isSubmitting, touched})=> (
                        <Form noValidate>
                            <h3 style={{textAlign: 'center', color: '#484848'}}>Forgot Password</h3>
                            <div style={{marginBottom: '15px', textAlign: 'center', color: '#1f3a93', fontWeight: '500'}}>
                                <span>We'll mail you a link to create a new password</span>
                            </div>
                            <SubmissionError message={this.state.error} marginTop='-15px' />
                            <span class='fapp-label'>Email</span>
                            <Input className='fapp-page-input fapp-big-font' tag={Field} type="email" name="email" placeholder="Your Email"/>
                            {touched.email &&
                                <ErrorMessage name="email">
                                    {msg => <span style={{color:'#d64541', fontWeight: '700', textAlign: 'center', display: 'block', marginTop: '-8px'}}>{msg}</span>}    
                                </ErrorMessage>
                            }
                            <div style={{textAlign: 'center'}}>
                                <Button color='success' type='submit' disabled={isSubmitting} className='fapp-cta-button' style={{marginTop: '15px'}}>SEND RESET LINK</Button>
                            </div>
                        </Form>
                    )}
                </Formik>
            );
        } else {
            return (
                <div>
                     <h3 style={{textAlign: 'center', color: '#484848'}}>Forgot Password</h3>
                     <div style={{marginTop: '50px', marginBottom: '50px', textAlign: 'center', color: '#049372', fontWeight: '500', fontSize: '18px'}}>
                         <span>An email has been sent to {this.state.agentEmail}. Please follow the link in the email to reset your password.</span>
                     </div>
                </div>
            );
        }
    }

    render(){
        return <HalfScreen renderForm={this.renderForgotPasswordForm} marginTop='10rem' />;
    }
}

export default connect(null, mapDispatchToProps)(ForgotPassword);