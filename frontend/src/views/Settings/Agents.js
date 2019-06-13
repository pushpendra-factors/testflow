import React, { Component } from 'react';
import { bindActionCreators } from 'redux';
import { connect } from 'react-redux';
import {
    Col,
    Row,
    Card,
    CardHeader,
    CardBody,
    Input,
    Button
} from 'reactstrap';
import { fetchProjectAgents, projectAgentInvite, projectAgentRemove } from "../../actions/projectsActions";
import * as yup from 'yup';
import { Formik, Form, Field, ErrorMessage } from 'formik';
import  { InvalidEmail, MissingEmail } from '../Pages/ValidationMessages';
import SubmissionError from '../Pages/SubmissionError';

const ROLE_AGENT = 1;
const ROLE_ADMIN = 2;

const AgentRecord = (props) => { 
  return (
    <Row>      
      <Col md={{size: props.emailColSize}} style={{ paddingTop: '5px' }}> { props.email } </Col>
      <Col md={{size: 1}} style={{ paddingTop: '5px' }}> { props.role == ROLE_ADMIN ? "Admin" : "User" } </Col>
      <div> { props.email != props.currentAgentEmail && <Button className="fapp-inline-button" onClick={props.handleDelete}><i className="icon-close" style={{ fontSize: '17px', fontWeight: 700, color: '#888' }}></i></Button> }</div>
      { props.isEmailVerified === false ?  <Col md={{ size: 2 }} style={{ paddingTop: '5px' }} className="fapp-label light"> Pending </Col> : null }
    </Row>
  )
}

const mapStateToProps = store => {
  return {		
    currentProjectId: store.projects.currentProjectId,
    projectAgentMappings: store.projects.projectAgents,
    agent: store.agents.agent,
    agents: store.projects.agents
  }
}

const mapDispatchToProps = dispatch => {
  return bindActionCreators({  fetchProjectAgents, projectAgentInvite, projectAgentRemove }, dispatch);
}

class Agents extends Component {
  constructor(props) {
    super(props);
  }

  componentWillMount(){
    this.props.fetchProjectAgents(this.props.currentProjectId);
  }

  createAgentInviteForm(){
    return <Formik
      initialValues={{ email: '' }}
      validationSchema = {yup.object().shape({
          email: yup.string().email(InvalidEmail).required(MissingEmail)
      })}
      onSubmit={(values, {setSubmitting, resetForm, setFieldError}) => {
          this.props.projectAgentInvite(this.props.currentProjectId, values.email)
          .then(() => {
              setSubmitting(false);
              resetForm({email:''});
          })
          .catch((msg) => {
              setSubmitting(false);
              setFieldError('general', msg);
          });
      }}
    >
      {({isSubmitting, touched, errors})=> (
        <Form noValidate>
            <Row>
            <Col md={{size: 3}} >
              <Input className='fapp-input fapp-medium-font' tag={Field} type="email" name="email" placeholder="Your invitee's email"/>
              {
                touched.email &&
                  <ErrorMessage name="email">
                      {msg => <span style={{ color:'#d64541', textAlign: 'center', display: 'block', marginTop: '-6px', fontSize: '14px' }}>{msg}</span>}
                  </ErrorMessage>
              }
              { errors.general && <span style={{ color:'#d64541', textAlign: 'center', display: 'block', marginTop: '-6px', fontSize: '14px' }}>{errors.general}</span>}
            </Col>
            <Col style={{ marginLeft: '-18px', paddingTop: '12px' }}> <Button type='submit' outline color='primary' disabled={isSubmitting} style={{ padding: '8px 15px' }}> Send Invitation </Button> </Col> 
          </Row>
        </Form>
      )}
    </Formik>
  }

  renderInviteAgentsForm() {
    return (
      <Card className='fapp-card' style={{ marginBottom: '10px' }}>
        <CardHeader style={{ marginBottom: '5px' }}>
          <strong>Invite User</strong>
        </CardHeader>
        <CardBody className='fapp-medium-font'>
          { this.createAgentInviteForm() }
        </CardBody>
      </Card>
    );
  }

  removeProjectAgent = (projectId, agentUUID) => {
    this.props.projectAgentRemove(projectId, agentUUID);
  }

  getColSizeByEmailLen() {
    if (!this.props.projectAgentMappings || !this.props.agents) return 2;

    let maxLength = 0;
    for (let i in this.props.projectAgentMappings) {
      let agent = this.props.agents[this.props.projectAgentMappings[i].agent_uuid];
      if (agent && agent.email && agent.email.length > maxLength) 
        maxLength = agent.email.length;
    }

    // use col size 3 for lengthy emails.
    return maxLength > 18 ? 3 : 2;
  }

  renderAgentsList(emailColSize) {
    let agents = []
    let currentAgent = null;

    this.props.projectAgentMappings.map((v, i) => {
      let agent = this.props.agents[v.agent_uuid]
      let record = <AgentRecord
        key={i}
        email={agent.email}
        emailColSize={emailColSize}
        role={v.role}
        isEmailVerified={agent.is_email_verified}
        handleDelete= {() => this.removeProjectAgent(this.props.currentProjectId, agent.uuid)}
        currentAgentEmail = {this.props.agent.email}
      />

      if (agent.email == this.props.agent.email) {
        currentAgent = record;
      } else {
        agents.push(record); 
      }
    })

    // current agent always on the top.
    agents.unshift(currentAgent);

    return agents;
  }

  render() {
    let emailColSize = this.getColSizeByEmailLen();

    return (
        <div className='fapp-content fapp-content-margin'>
        {this.renderInviteAgentsForm()}
        <Card className='fapp-card' style={{ marginTop: '-30px' }}>
          <CardHeader style={{ marginBottom: '5px' }}>
              <strong>Users</strong>
          </CardHeader>
          <CardBody className='fapp-medium-font'>
              <Row style={{ marginBottom: '4px' }}>
                <Col md={emailColSize} className='fapp-label light'>Email</Col>
                <Col md={1} className='fapp-label light'>Role</Col>
              </Row>
              { this.renderAgentsList(emailColSize) }
          </CardBody>
        </Card>
      </div>
    );
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(Agents);