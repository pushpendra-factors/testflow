import { connect } from 'react-redux';
import React, { Component } from 'react';
import { bindActionCreators } from 'redux';
import { Formik, Form, Field, ErrorMessage } from 'formik';
import * as yup from 'yup';
import {
    Row,
    Col,
    Card,
    Input,
    Button,
    CardBody,
    CardHeader,
} from 'reactstrap';
import Select from 'react-select';

import { fetchAgentBillingAccount, updateBillingAccount } from "../../actions/agentActions";
import  { MissingPincode, MissingPhoneNo, MissingOrgName, MissingBillingAddr } from '../Pages/ValidationMessages';

const mapStateToProps = store => {
  return {
    projects: store.projects.projects,
    agent: store.agents.agent,
    billingAccount: store.agents.billing.billingAccount,
    projects: store.agents.billing.projects,
    accountAgents: store.agents.billing.accountAgents,
    accountPlan: store.agents.billing.plan,
    availablePlans: store.agents.billing.availablePlans,
  }
}

const mapDispatchToProps = dispatch => {
  return bindActionCreators({ fetchAgentBillingAccount, updateBillingAccount  }, dispatch);
}

class AccountSettings extends Component {
  constructor(props) {
    super(props);    

    this.state = {
      editBillingDetails: false
    }
  }

  renderProjects = () => {
    return (
      this.props.projects.map((project, i) => {
        return <div className='fapp-small-font ' key={i}>{project.name}</div>
      })
    )
  }

  renderAgents = () => {
    return (
      <div>
        {
          this.props.accountAgents.map((agent, i) => {
            return <div className='fapp-small-font ' key={i}>{agent.email}</div>
          })
        }
      </div>
    )
  }

  toggleEditBillingDetails = () => {
    this.setState({ editBillingDetails: !this.state.editBillingDetails });
  }

  renderBillingAccount = () => {
    if (!this.props.billingAccount){
      return null;
    }

    if (!this.state.editBillingDetails) {
      return (
        <div>
          <div style={{ marginBottom: '8px' }}>
            <span  className='fapp-label light'>Organization Name: </span>
            <span>{ this.props.billingAccount.organization_name }</span>
          </div>
          <div style={{ marginBottom: '8px' }}>
            <span  className='fapp-label light'>Billing Address: </span>
            <div style={{ marginTop: '5px' }}>{ this.props.billingAccount.billing_address }</div>
          </div>
          <div  style={{ marginBottom: '8px' }}>
            <span  className='fapp-label light'>Pincode: </span>
            <span>{ this.props.billingAccount.pincode }</span>
          </div>
          <div style={{ marginBottom: '8px' }}>
            <span  className='fapp-label light'>Phone Number: </span>
            <span>{ this.props.billingAccount.phone_no }</span>
          </div>
          <div  style={{ marginBottom: '8px' }}>
            <span  className='fapp-label light'>Plan: </span>
            <span>{ this.props.accountPlan.name }</span>
            <span className='fapp-button' style={{ marginLeft: '6px'}} onClick={this.toggleEditBillingDetails}>change</span>
          </div>
        </div>
      ) 
    }
    
    return (
        <Formik        
          initialValues={{
              OrganizationName:this.props.billingAccount.organization_name,
              Pincode: this.props.billingAccount.Pincode,
              Address: this.props.billingAccount.billing_address,
              PhoneNumber: this.props.billingAccount.phone_no,
              planCode: this.props.accountPlan.code
            }
          }
          validationSchema = {
            yup.object().shape({
              planCode: yup.string().required(),
              OrganizationName: yup.string().when("planCode",{
                is: "startup",
                then: yup.string().required(MissingOrgName)
              }),
              Pincode: yup.string().when("planCode",{
                is: "startup",
                then: yup.string().required(MissingPincode)
              }),
              Address: yup.string().when("planCode",{
                is: "startup",
                then: yup.string().required(MissingBillingAddr)
              }),
              PhoneNumber: yup.string().when("planCode",{
                is: "startup",
                then: yup.string().required(MissingPhoneNo)
              })
            })
          }
          onSubmit={(values, {setSubmitting}) => {
            let params = {
              Pincode: values.Pincode,
              organization_name: values.OrganizationName,
              phone_no: values.PhoneNumber,
              billing_address: values.Address,
              plan_code: values.planCode
            };
            this.props.updateBillingAccount(params)
              .then(() => {
                  setSubmitting(false);
                  this.toggleEditBillingDetails();
              })
              .catch((msg) => {
                  setSubmitting(false);
                  this.setState({ error: msg });
              });
          }}
        >
          {({values, isSubmitting, touched, setFieldTouched, setFieldValue}) => (
            <Form noValidate>
              <span className='fapp-label'> Organization Name* </span>
              <Input className='fapp-input'  tag={Field} type="text" name="OrganizationName" placeholder="Organization name"/>
              {
                touched.OrganizationName &&
                <ErrorMessage name="OrganizationName">
                    {msg => <span className='fapp-error-span light' >{msg}</span>}
                </ErrorMessage>
              }
              <span className='fapp-label'> Phone Number</span>
              <Input className='fapp-input'  tag={Field} type="text" name="PhoneNumber" placeholder="Your phone number"/>
              {
                touched.PhoneNumber &&
                <ErrorMessage name="PhoneNumber">
                    {msg => <span className='fapp-error-span light'>{msg}</span>}    
                </ErrorMessage>
              }
              <span className='fapp-label'> Billing Address*</span>
              <Input className='fapp-input'  tag={Field} type="text" name="Address" placeholder="Your billing address"/>
              {
                touched.Address &&
                <ErrorMessage name="Address">
                    {msg => <span className='fapp-error-span light' >{msg}</span>}    
                </ErrorMessage>
              }
              <span className='fapp-label'> Pincode* </span>
              <Input className='fapp-input'  tag={Field} type="text" name="Pincode" placeholder="Your pincode"/>
              {
                touched.Pincode &&
                <ErrorMessage name="Pincode">
                    {msg => <span className='fapp-error-span light' >{msg}</span>}    
                </ErrorMessage>
              }
              <span className='fapp-label'> Plan* </span>
              <div style={{ marginTop: '10px', marginBottom: '10px' }}>
                <Select className='fapp-select light'
                  placeholder="plan"
                  onBlur={() => setFieldTouched("planCode", true)}
                  onChange={(item) => setFieldValue("planCode", item.value)}
                  tag={Field}
                  options={[{ label:'Free', value: 'free' }, { label:'Startup', value: 'startup' }]}
                  name="planCode"
                  value={{label:this.props.availablePlans[values.planCode], value: values.planCode}}
                />
              </div>
              {
                touched.plan &&
                <ErrorMessage name="plan">
                    {msg => <span className='fapp-error-span light' style={{marginTop: '-15px'}}>{msg}</span>}
                </ErrorMessage>
              }
              <div style={{ textAlign: 'center' }}>
                <Button color='primary' outline type='submit' disabled={isSubmitting} style={{marginTop: '15px', padding: '8px 20px', marginRight: '10px', fontSize: '15px'}}>Update</Button>
                <Button color='danger' outline style={{marginTop: '15px', padding: '8px 20px', fontSize: '15px'}} onClick={this.toggleEditBillingDetails}>Cancel</Button>
              </div>
            </Form>
          )}
        </Formik>
      )           
  }
  
  render() {
    return (
      <div className='animated fadeIn fapp-content fapp-content-margin'>
        <Row>
          <Col xs='6' md='6' style={{ paddingRight: '40px' }}>
            <Card className="fapp-card">
              <CardHeader>
                <strong>Billing Information</strong>
                <span className={ this.state.editBillingDetails ? 'fapp-label light' : 'fapp-button'} style={{ float: 'right', cursor: 'pointer'}} onClick={this.toggleEditBillingDetails}>{ this.state.editBillingDetails ? 'x' : 'edit' }</span>
              </CardHeader>
              <CardBody>
                  { this.renderBillingAccount() }
              </CardBody>
            </Card>
          </Col>
          <Col xs='6' md='6' style={{ paddingLeft: '40px' }}>
            <Card className="fapp-card">
              <CardHeader>
                <strong>Usage</strong>
              </CardHeader>
              <CardBody>
                <div>
                  <span className='fapp-label light'>Total seats consumed: </span>
                  <span>{ this.props.accountAgents.length} of {this.props.accountPlan.max_no_of_agents }</span>
                </div>
              </CardBody>
            </Card>
          </Col>
        </Row>
        <Row>
          <Col>
            <Card className="fapp-card">
              <CardHeader style={{ marginBottom: '5px' }}>
                <strong>Projects</strong>
              </CardHeader>
              <CardBody>
                <div className='fapp-label light' style={{ marginBottom: '5px' }} >Name</div>
                { !!this.props.billingAccount && this.renderProjects()}
              </CardBody>
            </Card>
          </Col>
        </Row>
        <Row>
          <Col>
            <Card className="fapp-card">
              <CardHeader style={{ marginBottom: '5px' }}>
                <strong>All Project Users</strong>
              </CardHeader>
              <CardBody>
                <div className='fapp-label light' style={{ marginBottom: '5px' }} >Email</div>
                { !!this.props.billingAccount && this.renderAgents()}
              </CardBody>
            </Card>
          </Col>
        </Row>
      </div>
    );
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(AccountSettings);