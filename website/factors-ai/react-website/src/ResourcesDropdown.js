import React from 'react';
import { Link } from 'react-router-dom';
import { Dropdown, DropdownToggle, DropdownMenu, DropdownItem } from 'reactstrap';

class ResourcesDropdownOpen extends React.Component {
  constructor(props) {
    super(props);

    this.toggleResources = this.toggleResources.bind(this);
    this.state = {
      resourcesDropdownOpen: false
    };
  }

  toggleResources() {
    this.setState(prevState => ({
      resourcesDropdownOpen: !prevState.resourcesDropdownOpen
    }));
  }

  redirectTo(path) {
    window.location.href = path;
  }

  render() {
    return (
      <Dropdown style={{paddingTop: '3px'}} isOpen={this.state.resourcesDropdownOpen} toggle={this.toggleResources}>
        <DropdownToggle>
          Resources
        </DropdownToggle>
        <DropdownMenu style={{ fontSize: "14px", border: "none", boxShadow: "rgba(0, 0, 0, 0.2) 0px 2px 4px 0px, rgba(0, 0, 0, 0) 0px 2px 10px 0px", zIndex: "100"  }}>
          <DropdownItem onClick={() => this.redirectTo('/blog')}>
            Blog
          </DropdownItem>
          <DropdownItem onClick={() => this.redirectTo('/integrations/segment')}>
            Integration - Segment
          </DropdownItem>
        </DropdownMenu>
      </Dropdown>
    );
  }
}

export default ResourcesDropdownOpen;