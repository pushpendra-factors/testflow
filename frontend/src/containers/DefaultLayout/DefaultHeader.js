import React, { Component } from 'react';
import { Badge, DropdownItem, DropdownMenu, DropdownToggle, Nav, NavItem, NavLink, TextMuted } from 'reactstrap';
import PropTypes from 'prop-types';
import { AppHeaderDropdown, AppSidebarToggler, AppNavbarBrand } from '@coreui/react';
import { AppSidebarForm } from '@coreui/react';
import Select from 'react-select';
import { bindActionCreators } from 'redux';
import { connect } from 'react-redux';

import factorslogo from '../../assets/img/brand/factors-logo.png';
import factorsicon from '../../assets/img/brand/factors-icon.png';
import { changeProject } from '../../actions/projectsActions';

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
  return bindActionCreators({ changeProject }, dispatch);
}

class DefaultHeader extends Component {
  handleChange = (selectedProject) => {
    let projectId = selectedProject.value;
    this.props.changeProject(projectId);
    this.props.refresh();
  }
  
  render() {
    // eslint-disable-next-line
    const { children, ...attributes } = this.props;

    return (
      <React.Fragment>
        <AppSidebarToggler className="d-lg-none" display="md" mobile />
        <AppNavbarBrand
          full={{ src: factorslogo, alt: 'factors.ai' }}
          minimized={{ src: factorsicon, alt: 'factors.ai' }}
        />
        <AppSidebarToggler className="d-md-down-none fapp-navbar-toggler" display="lg" />
        <AppSidebarForm className="fapp-select fapp-header-dropdown">
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
          <AppHeaderDropdown direction="down">
            <DropdownToggle nav>
              <i className="icon-bell fapp-bell"></i>
              {/* <Badge pill color="danger">5</Badge> */}
            </DropdownToggle>
            <DropdownMenu right style={{ right: 'auto' }}>
              <DropdownItem disabled><span class="text-muted">No messages here.</span></DropdownItem>
            </DropdownMenu>
          </AppHeaderDropdown>
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

export default connect(null, mapDispatchToProps)(DefaultHeader);