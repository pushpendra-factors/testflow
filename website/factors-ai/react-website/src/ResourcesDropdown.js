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

  render() {
    return (
      <Dropdown isOpen={this.state.resourcesDropdownOpen} toggle={this.toggleResources}>
        <DropdownToggle caret>
          Resources
        </DropdownToggle>
        <DropdownMenu>
        <DropdownItem><Link to="/blog">Blog</Link></DropdownItem>
        <DropdownItem><Link to="/integrations/segment">Integration - Segment</Link></DropdownItem>
        </DropdownMenu>
      </Dropdown>
    );
  }
}

export default ResourcesDropdownOpen;