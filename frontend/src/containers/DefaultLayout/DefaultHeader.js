import React, { Component } from 'react';
import { Badge, DropdownItem, DropdownMenu, DropdownToggle, Nav, NavItem, NavLink, TextMuted } from 'reactstrap';
import PropTypes from 'prop-types';

import { AppHeaderDropdown, AppSidebarToggler } from '@coreui/react';
import factorslogo from '../../assets/img/brand/factors-logo.svg'
import factorsicon from '../../assets/img/brand/factors-icon.svg'

import {
  AppSidebarForm,
} from '@coreui/react';
import Select from 'react-select';
// sidebar nav config

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

class DefaultHeader extends Component {
  handleChange = (selectedProject) => {
    this.props.fetchProjectDependencies(selectedProject.value);
  }
  
  render() {
    // eslint-disable-next-line
    const { children, ...attributes } = this.props;

    return (
      <React.Fragment>
        <AppSidebarToggler className="d-lg-none" display="md" mobile />
        <AppSidebarToggler className="d-md-down-none fapp-navbar-toggler" display="lg" />
        <AppSidebarForm className="fapp-sidebar-form">
          <Select
            options={this.props.selectableProjects}
            value={this.props.selectedProject}
            onChange={this.handleChange}
            styles={projectSelectStyles}
            placeholder={"Select Project ..."}
            blurInputOnSelect={true}
          />
        </AppSidebarForm>
        <Nav className="ml-auto fapp-header-right" navbar>
          <NavItem className="d-md-down-none">
            <AppHeaderDropdown direction="down">
              <DropdownToggle nav>
                <i className="icon-bell fapp-bell"></i>
                {/* <Badge pill color="danger">5</Badge> */}
              </DropdownToggle>
              <DropdownMenu right style={{ right: 'auto' }}>
                <DropdownItem disabled><span class="text-muted">No messages here.</span></DropdownItem>
              </DropdownMenu>
            </AppHeaderDropdown>
          </NavItem>
          <AppHeaderDropdown direction="down">
            <DropdownToggle nav>
              <img src={'assets/img/avatars/default.png'} className="img-avatar" alt="admin@bootstrapmaster.com" />
            </DropdownToggle>
            <DropdownMenu right style={{ right: 'auto' }}>
              <DropdownItem><i className="fa fa-user"></i> Profile</DropdownItem>
              <DropdownItem><i className="fa fa-wrench"></i> Settings</DropdownItem>
              <DropdownItem><i className="fa fa-lock"></i> Logout</DropdownItem>
            </DropdownMenu>
          </AppHeaderDropdown>
        </Nav>
      </React.Fragment>
    );
  }
}

DefaultHeader.propTypes = propTypes;
DefaultHeader.defaultProps = defaultProps;

export default DefaultHeader;
