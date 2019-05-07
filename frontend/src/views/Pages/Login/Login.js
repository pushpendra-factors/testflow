import React, { Component } from 'react';
import { 
  Button,
  Input, 
} from 'reactstrap';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import { Redirect } from 'react-router-dom';
import * as yup from 'yup';
import { Formik, Form, Field, ErrorMessage } from 'formik';

import { login } from "../../../actions/agentActions";
import  { InvalidEmail, MissingEmail, MissingPassword } from '../ValidationMessages';
import HalfScreen from '../HalfScreen';
import SubmissionError from '../SubmissionError';
import factorsai from '../../../common/factorsaiObj';

const mapDispatchToProps = dispatch => {
  return bindActionCreators({ login }, dispatch);
}

const mapStateToProps = store => {
  return { isLoggedIn: store.agents.isLoggedIn}
}

class Login extends Component {
  constructor(props){
    super(props);

    this.state = {
      error: null
    }
  }
  
  redirectToForgotPassword = (event) => {
    event.preventDefault();    
    this.props.history.push("/forgotpassword");
  }

  redirectToSignup = (event) => {
    event.preventDefault();    
    this.props.history.push("/signup");
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
        onSubmit={(values, {setSubmitting}) => {
          let eventProperties = { email: values.email };
          this.props.login(values.email, values.password)
          .then(() => {
            setSubmitting(false);
            factorsai.track('login', eventProperties);    
          })
          .catch((msg) => {
            setSubmitting(false);
            this.setState({ error: msg });
            factorsai.track('login_failed', eventProperties);
          });                      
        }}
      >
        {({isSubmitting, touched})=> (
          <Form noValidate>
              <h3 style={{textAlign: 'center', marginBottom: '30px', color: '#484848'}}>Log in to factors.ai</h3>
              <SubmissionError message={this.state.error} />
              <span class='fapp-label'>Email</span>
              <Input className='fapp-input fapp-big-font' style={{marginBottom: '20px'}} tag={Field} type="email" name="email" placeholder="Your Email"/>
              {
                touched.email &&
                <ErrorMessage name="email">
                    {msg => <span style={{color:'#d64541', fontWeight: '700', textAlign: 'center', display: 'block', marginTop: '-15px'}}>{msg}</span>}    
                </ErrorMessage>
              } 
              <span class='fapp-label'>Password</span>
              <Input className='fapp-input fapp-big-font' style={{marginBottom: '20px'}} tag={Field} type="password" name="password" placeholder="Your Password"/>
              {
                touched.password &&
                  <ErrorMessage name="password">
                      {msg => <span style={{color:'#d64541', fontWeight: '700', display: 'block', textAlign: 'center', display: 'block', marginTop: '-15px'}}>{msg}</span>}    
                  </ErrorMessage>
              }
              <div style={{textAlign: 'center'}}>
                <Button color='success' type='submit' disabled={isSubmitting} className='fapp-cta-button' style={{marginTop: '15px'}}>Log in</Button>
              </div>
              <Button color='link'  onClick={this.redirectToSignup} style={{fontWeight: '300'}} className="px-0"> I don't have an account </Button>
              <Button color='link' onClick={this.redirectToForgotPassword} style={{float: 'right', fontWeight: '300'}} className="px-0"> Forgot password? </Button>
          </Form>
        )}  
      </Formik>
    )
  }

  render() {
    if(this.props.isLoggedIn){
      return <Redirect to='/' />
    }
    
    return <HalfScreen renderForm={this.renderLoginForm} marginTop='8rem' />;
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(Login);
