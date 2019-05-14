import React, { Component } from 'react';
import { Formik, Form, Field, ErrorMessage } from 'formik';
import { Button, Card, CardBody, Col, Container, Alert, Input, InputGroup, InputGroupAddon, InputGroupText, Row } from 'reactstrap';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import * as yup from 'yup';

import { signup } from "../../../actions/agentActions";
import  { InvalidEmail, MissingEmail } from '../ValidationMessages';
import HalfScreen from '../HalfScreen';
import SubmissionError from '../SubmissionError';
import factorsai from '../../../common/factorsaiObj';

const mapDispatchToProps = dispatch => {
  return bindActionCreators({ signup }, dispatch);
}

class Signup extends Component {
  
  constructor(props){
    super(props);
    this.state = {
      signupPerformed: false,
      agentEmail:'',
      error: null,
    }
  }

  redirectToLogin = (event) => {
    event.preventDefault();    
    this.props.history.push("/login");
  }
  
  renderSignupForm = () => {
    if (this.state.signupPerformed){
      return (
        <div>
          <h3 style={{textAlign: 'center', color: '#484848'}}>Sign up to factors.ai</h3>
          <div style={{marginTop: '50px', marginBottom: '50px', textAlign: 'center', color: '#049372', fontWeight: '500', fontSize: '18px'}}>
              <span style={{display: 'block', color: '#1f3a93', fontWeight: '500', marginBottom: '12px'}}>Thanks for signing up!</span>
              <span style={{display: 'block'}}>An activation email has been sent to {this.state.agentEmail}. Please follow the link in the email to activate your account.</span>
          </div>
        </div>
      );
    }

    return (
      <Formik
        initialValues={{email:''}}
        validationSchema = {
            yup.object().shape({
                email: yup.string().email(InvalidEmail).required(MissingEmail)
            })
        }
        onSubmit={(values, {setSubmitting}) => {
            let eventProperties = { email: values.email };          
            this.props.signup(values.email)
            .then(() => {
                setSubmitting(false);
                this.setState({signupPerformed: true, agentEmail: values.email });
                factorsai.track('signup', eventProperties);    
            })
            .catch((msg) => {
                setSubmitting(false);
                this.setState({ error: msg });
                factorsai.track('signup_failed', eventProperties);
            });                             
        }}
      >
        {({isSubmitting, touched}) => (
          <Form noValidate>
              <h3 style={{textAlign: 'center', marginBottom: '30px', color: '#484848'}}>Sign up to factors.ai</h3>
              <SubmissionError message={this.state.error} />
              <span class='fapp-label'>Email</span>
              <Input className='fapp-page-input fapp-big-font' style={{marginBottom: '20px'}} tag={Field} type="email" name="email" placeholder="Your Work Email"/>
              {
                touched.email &&
                <ErrorMessage name="email">
                    {msg => <span style={{color:'#d64541', fontWeight: '700',textAlign: 'center', display: 'block', marginTop: '-15px'}}>{msg}</span>}  
                </ErrorMessage>
              }
              <div style={{textAlign: 'center'}}>
                <Button color='success' type='submit' disabled={isSubmitting} className='fapp-cta-button' style={{marginTop: '15px'}}>Create Account</Button>
              </div>
              <Button color='link' onClick={this.redirectToLogin} style={{float: 'right', fontWeight: '300'}} className="px-0"> I have an account already. Sign in now. </Button>
          </Form>
        )}  
      </Formik>

    )
  }

  render() {
    return (
      <HalfScreen renderForm={this.renderSignupForm} marginTop='10rem' />
    );
  }
}

export default connect(null, mapDispatchToProps)(Signup);
