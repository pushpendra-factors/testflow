import { connect } from 'react-redux';
import React, { Component } from 'react';
import { bindActionCreators } from 'redux';
import {
  Row,
  Col,
  Card,
  CardBody,
  CardHeader,
  Input,
  Button,
  Modal,
  ModalBody,
  ModalHeader,
} from 'reactstrap';
import { Formik, Form, Field, ErrorMessage } from 'formik';
import * as yup from 'yup';
import Avatar from 'react-avatar';

import { updateAgentInfo, updateAgentPassword } from "../../actions/agentActions";
import { MissingPassword, PasswordMinEightChars, PasswordsDoNotMatch } from "../Pages/ValidationMessages";

const mapStateToProps = store => {
  return {
    agent: store.agents.agent
  }
}

const mapDispatchToProps = dispatch => {
  return bindActionCreators({ updateAgentInfo, updateAgentPassword }, dispatch);
}

class Profile extends Component {
  constructor(props) {
    super(props);

    this.state = {
      editPersonalInformation: false,

      showUpdatePasswordModal: false,
      updatePasswordModalMessage: null,
    }
  }

  renderUpdatePasswordForm = () => {
    if (!this.props.agent) return null;

    return (
      <Formik        
        initialValues={{
            currentPassword: '',
            newPassword: '',
            ReNewPassword: '',
          }
        } 
        validationSchema = {
          yup.object().shape({
            currentPassword: yup.string().required(MissingPassword),
            newPassword: yup.string().required().min(8, PasswordMinEightChars),
            reNewPassword: yup.string().required().oneOf([yup.ref('newPassword')], PasswordsDoNotMatch)
          })
        }
        onSubmit={(values, {setSubmitting}) => {
          let params = {
            current_password: values.currentPassword,
            new_password: values.newPassword,
          };
          this.props.updateAgentPassword(params)
            .then(() => {
                setSubmitting(false);
                this.toggleUpdatePasswordModal();
            })
            .catch((msg) => {
                setSubmitting(false);
                this.setState({ error: msg });
            });
        }}
      >
        {({isSubmitting, touched}) => (
          <Form noValidate>
            <span className='fapp-label'> Current Password* </span>
            <Input className='fapp-input' style={{marginBottom: '20px'}} tag={Field} type="password" name="currentPassword" placeholder="Current Password"/>
            {
              touched.currentPassword &&
              <ErrorMessage name="currentPassword">
                  {msg => <span className='fapp-error-span light' style={{marginTop: '-15px'}}>{msg}</span>}
              </ErrorMessage>
            }
            <span className='fapp-label'> New Password </span>
            <Input className='fapp-input' style={{marginBottom: '20px'}} tag={Field} type="password" name="newPassword" placeholder="New Password"/>
            {
              touched.newPassword &&
              <ErrorMessage name="newPassword">
                  {msg => <span className='fapp-error-span light' style={{marginTop: '-15px'}}>{msg}</span>}
              </ErrorMessage>
            }
            <span className='fapp-label'> Confirm New Password </span>
            <Input className='fapp-input' style={{marginBottom: '20px'}} tag={Field} type="password" name="reNewPassword" placeholder="Confirm New Password"/>
            {
              touched.reNewPassword &&
              <ErrorMessage name="reNewPassword">
                  {msg => <span className='fapp-error-span light' style={{marginTop: '-15px'}}>{msg}</span>}
              </ErrorMessage>
            }
            
            <div style={{ marginTop: '35px', textAlign: 'center', marginBottom: '25px' }}>
              <Button color='primary' outline type='submit' disabled={isSubmitting} style={{ fontSize: '15px', marginRight: '10px', padding: '8px 25px' }}>Update Password</Button>
            </div>
          </Form>
        )}
      </Formik>
    );
  }

  getAgentName = () => {
    return (this.props.agent 
      && this.props.agent.first_name) ? this.props.agent.first_name : '';
  }

  renderPersonalInformation = () => {
    if (!this.props.agent) {
      return null;
    }

    if (!this.state.editPersonalInformation) {
      return (
        <Row style={{ marginBottom: '20px' }}>
          <Col md={2}>
            <Avatar name={this.getAgentName()}  maxInitials={1} round={true} color='#3a539b' textSizeRatio={2} size='80' style={{ fontWeight: '700', marginTop: '5px' }} />
          </Col>
          <Col style={{ marginLeft: '20px', marginTop: '15px' }}>
            <div style={{ marginBottom: '8px' }}>
              <span  className='fapp-label light'>First Name: </span>
              <span>{ this.props.agent.first_name }</span>
            </div>
            <div style={{ marginBottom: '8px' }}>
              <span  className='fapp-label light'>Last Name: </span>
              <span>{ this.props.agent.last_name }</span>
            </div>
          </Col>
        </Row>
      )
    }

    return (
      <Formik        
          initialValues={{
              FirstName: this.props.agent.first_name,
              lastName: this.props.agent.last_name,
            }
          }
          validationSchema = {
            yup.object().shape({
              FirstName: yup.string().required(),
              lastName: yup.string(),
            })
          }
          onSubmit={(values, {setSubmitting}) => {
            let params = {
              first_name: values.FirstName,
              last_name: values.lastName,
            };
            this.props.updateAgentInfo(params)
              .then(() => {
                  setSubmitting(false);
                  this.toggleEditPersonalInformation();
              })
              .catch((msg) => {
                  setSubmitting(false);
                  this.setState({ error: msg });
              });
          }}
        >
          {({isSubmitting, touched}) => (
            <Form noValidate>
              <span className='fapp-label'> First Name* </span>
              <Input className='fapp-input' style={{marginBottom: '20px'}} tag={Field} type="text" name="FirstName" placeholder="First Name"/>
              {
                touched.FirstName &&
                <ErrorMessage name="FirstName">
                    {msg => <span className='fapp-error-span light' style={{marginTop: '-15px'}}>{msg}</span>}
                </ErrorMessage>
              }
              <span className='fapp-label'> Last Name </span>
              <Input className='fapp-input' style={{marginBottom: '20px'}} tag={Field} type="text" name="lastName" placeholder="last Name"/>
              {
                touched.lastName &&
                <ErrorMessage name="lastName">
                    {msg => <span className='fapp-error-span light' style={{marginTop: '-15px'}}>{msg}</span>}
                </ErrorMessage>
              }
              
              <div style={{textAlign: 'center'}}>
                <Button color='primary' outline type='submit' disabled={isSubmitting} style={{marginTop: '15px', padding: '8px 25px', fontSize: '15px', marginRight: '10px'}}>Update</Button>
                <Button color='danger' outline style={{marginTop: '15px', padding: '8px 25px', fontSize: '15px'}} onClick={this.toggleEditPersonalInformation}>Cancel</Button>
              </div>
            </Form>
          )}
      </Formik>
    );
  }

  toggleEditPersonalInformation = () => {
    this.setState({ editPersonalInformation: !this.state.editPersonalInformation });
  }

  toggleUpdatePasswordModal = () => {
    this.setState({ showUpdatePasswordModal: !this.state.showUpdatePasswordModal });
  }
  
  render() {
    return (
      <div className='animated fadeIn fapp-content fapp-content-margin'>
        <Row>
          <Col xs='5' md='5' style={{ paddingRight: '40px' }}>
            <Card className="fapp-card">
              <CardHeader style={{ marginBottom: '5px' }}>
                <strong>Personal Information</strong>
                <span className={ this.state.editPersonalInformation ? 'fapp-label light' : 'fapp-button'} style={{ float: 'right', cursor: 'pointer'}} onClick={this.toggleEditPersonalInformation}>{ this.state.editPersonalInformation ? 'x' : 'edit' }</span>
              </CardHeader>
              <CardBody>
                { this.renderPersonalInformation() }
                <Button style={{  marginTop: '20px', padding: '8px 16px' }} onClick={this.toggleUpdatePasswordModal} outline color='primary'>Change Password</Button>
              </CardBody>
            </Card>
          </Col>
        </Row>

        <Modal isOpen={this.state.showUpdatePasswordModal} toggle={this.toggleUpdatePasswordModal} style={{ marginTop: '10rem' }}>
          <ModalHeader toggle={this.toggleUpdatePasswordModal}>Change Password</ModalHeader>
          <ModalBody style={{ padding: '15px 35px' }}>
            <div style={{ textAlign: 'center', marginBottom: '15px' }}>
              <span style={{ display: 'inline-block' }} className='fapp-error' hidden={this.state.updatePasswordModalMessage == null}>{ this.state.updatePasswordModalMessage }</span>
            </div>
            <Form >
              { this.renderUpdatePasswordForm() }
            </Form>
          </ModalBody>
        </Modal>
      </div>
    );
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(Profile);