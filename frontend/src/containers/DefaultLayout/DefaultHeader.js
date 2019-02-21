import React, { Component } from 'react';
import { DropdownItem, DropdownMenu, DropdownToggle, Input, Button, Nav} from 'reactstrap';
import PropTypes from 'prop-types';
import { AppHeaderDropdown, AppSidebarToggler, AppNavbarBrand } from '@coreui/react';
import { AppSidebarForm } from '@coreui/react';
import Select from 'react-select';
import { bindActionCreators } from 'redux';
import { connect } from 'react-redux';

import factorslogo from '../../assets/img/brand/factors-logo.png';
import factorsicon from '../../assets/img/brand/factors-icon.png';
import { changeProject, createProject } from '../../actions/projectsActions';
import { signout } from '../../actions/agentActions';

const propTypes = {
  children: PropTypes.node,
};

const defaultProps = {};

const projectSelectStyles = {
  option: (base, state) => ({
    ...base,
    color: '#5c6873',
    background: '#fff',
  }),
  singleValue: base => ({
    ...base,
    background: '#fff',
    color: '#5c6873',
  }),
  valueContainer: base => ({
    ...base,
    background: '#fff',
    color: '#5c6873',
  }),
  container: base => ({
    ...base,
    background: '#fff',
    border: 'none',
  }),
  indicatorSeparator: () => ({
    display: 'none',
  }),
}


const mapDispatchToProps = dispatch => {
  return bindActionCreators({ changeProject, createProject, signout }, dispatch);
}

class DefaultHeader extends Component {
  constructor(props){
    super(props);
    this.state = {
      createProject:{
        showForm: false,
        projectName : ""
      }
    }
  }

  handleChange = (selectedProject) => {
    let projectId = selectedProject.value;
    this.props.changeProject(projectId);
    this.props.refresh();
  }

  handleProjectNameFormChange = (e) => {
    let name = e.target.value.trim();
    if(name == "") console.error("project name cannot be empty");
    this.setState({ createProject: { projectName: name } });
  }

  toggleCreateProjectForm = () => {
    this.setState(prevState => {
      let _state = { ...prevState };
      _state.createProject.showForm = !prevState.createProject.showForm;
      return _state;
    })
  }
  
  handleCreateProject = (e) => {
    let projectName = this.state.createProject.projectName;
    if(projectName == "") {
      console.error("project name cannot be empty");
      return
    }

    this.props.createProject(projectName);
  }

  handleLogout = () => {
    this.props.signout();
  }

  render() {
    // eslint-disable-next-line
    const { children, ...attributes } = this.props;

    let dropDown = "";
    if(!!this.props.selectableProjects ){
      dropDown = <Select options={this.props.selectableProjects} value={this.props.selectedProject} onChange={this.handleChange} styles={projectSelectStyles} placeholder={"Select Project ..."} blurInputOnSelect={true}/>;
    }
    return (
      <React.Fragment>
        <AppSidebarToggler className="d-lg-none" display="md" mobile />
        <AppNavbarBrand
          full={{ src: factorslogo, alt: 'factors.ai' }}
          minimized={{ src: factorsicon, alt: 'factors.ai' }}
        />
        <AppSidebarToggler className="d-md-down-none fapp-navbar-toggler" display="lg" />
        <AppSidebarForm className="fapp-select fapp-header-dropdown" style={{width: '80%'}}>
          <div style={{display: 'inline-block', width: '15%', marginRight: '25px'}}> {dropDown} </div>         
          <Input type="text" placeholder="Project Name" onChange={this.handleProjectNameFormChange} style={{width: 'auto', display: 'inline-block', marginRight: '25px'}} />
          <Button color='primary' onClick={this.handleCreateProject}>Create</Button>
        </AppSidebarForm>
        <Nav className="ml-auto fapp-header-right" navbar>          
          <AppHeaderDropdown direction="down">
            <DropdownToggle nav>
              <img src={'assets/img/avatars/default.png'} className="img-avatar" alt="admin@bootstrapmaster.com" />
            </DropdownToggle>
            <DropdownMenu right style={{ right: 'auto' }}>
              <DropdownItem><i className="fa fa-user"></i> Profile</DropdownItem>
              <DropdownItem><i className="fa fa-wrench"></i> Settings</DropdownItem>
              <DropdownItem onClick={this.handleLogout}><i className="fa fa-lock"></i> Logout</DropdownItem>
            </DropdownMenu>
          </AppHeaderDropdown>
        </Nav>
      </React.Fragment>
    );
  }
}

DefaultHeader.propTypes = propTypes;
DefaultHeader.defaultProps = defaultProps;

export default connect(null, mapDispatchToProps)(DefaultHeader);