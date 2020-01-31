import React, { Component } from 'react';
import { Formik, Form, Field, ErrorMessage, useFormik } from 'formik';
import { Button, Card, CardBody, Col, Container, Alert, Input, InputGroup, InputGroupAddon, InputGroupText, Row } from 'reactstrap';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import * as yup from 'yup';
import queryString from 'query-string';

import { signup } from "../../../actions/agentActions";
import  { InvalidEmail, MissingEmail, MissingPhoneNo } from '../ValidationMessages';
import HalfScreen from '../HalfScreen';
import SubmissionError from '../SubmissionError';
import factorsai from '../../../common/factorsaiObj';
import { isProduction } from '../../../util';
import PhoneInput from 'react-phone-input-2'
import 'react-phone-input-2/lib/style.css'
import './Signup.css';

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
        initialValues={{email:'', phone: ''}}
        validationSchema = {
            yup.object().shape({
                email: yup.string().email(InvalidEmail).required(MissingEmail),
                phone: yup.string().test('phone number', 'Invalid phone number', val => (val && val.match(/(\d+)/)[0].length >= 6)).required(MissingPhoneNo)
            })
        }
        onSubmit={(values, {setSubmitting}) => {
          // track create account conversion.
          if (isProduction()) gtag_report_conversion();
          let parsed = queryString.parse(this.props.location.search);
            let planCode = parsed.plan;
            let eventProperties = { email: values.email, phone: values.phone, plan_code: planCode };
            this.props.signup(values.email, values.phone, planCode)
            .then(() => {
                  setSubmitting(false);
                  this.setState({signupPerformed: true, agentEmail: values.email,});
                  factorsai.track('signup', eventProperties);
            })
            .catch((msg) => {
                setSubmitting(false);
                this.setState({ error: msg });
                factorsai.track('signup_failed', eventProperties);
            });                       
        }}
      >
        {({isSubmitting, touched, setFieldValue, values}) => (
          <Form noValidate>
              <h3 style={{textAlign: 'center', marginBottom: '30px', color: '#484848'}}>Sign up to factors.ai</h3>
              <SubmissionError message={this.state.error} />
              <span className='fapp-label'>Email</span>
              <Input className='fapp-page-input fapp-big-font' style={{marginBottom: '20px'}} tag={Field} type="email" name="email" placeholder="Your Work Email"/>
              {
                touched.email &&
                <ErrorMessage name="email">
                    {msg => <span style={{color:'#d64541', fontWeight: '700',textAlign: 'center', display: 'block', marginTop: '-15px'}}>{msg}</span>}  
                </ErrorMessage>
              }
              <span className='fapp-label'>Phone</span>
              <PhoneInput
                placeholder="Enter phone number"
                onChange={(e)=> setFieldValue('phone', e)}
                tag={Field}
                name="phone"
                value={values.phone}
                autoFormat={false}
                enableLongNumbers={true}
                enableSearch={true}
                disableSearchIcon={true}
                country={'us'}
                style={{marginBottom: '20px', marginTop: '10px',}}
              />
              {
                touched.phone &&
                <ErrorMessage name="phone">
                    {msg => <span style={{color:'#d64541', fontWeight: '700',textAlign: 'center', display: 'block', marginTop: '-15px', marginBottom: '10px'}}>{msg}</span>}  
                </ErrorMessage>
              }
              <div style={{textAlign: 'center'}}>
                <Button color='success' type='submit' disabled={isSubmitting} className='fapp-cta-button' style={{marginTop: '8px'}}>CREATE ACCOUNT</Button>
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
