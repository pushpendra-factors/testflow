import React, { Component } from 'react';
import Select from 'react-select';

import { 
  fetchProjectEventProperties,
  fetchProjectUserProperties,
} from '../../actions/projectsActions';

const PROPERTY_TYPE_OPTS = [
  { value: 'user', label: 'User Property' },
  { value: 'event', label: 'Event Property' }
];

class GroupBy extends Component {
  constructor(props) {
    super(props);
    this.state = {
      nameOpts: []
    };
  }

  addToNameOptsState(props) {
    let opts = [];

    this.setState((ps) => {
      let state = { ...ps };
      state.nameOpts = [ ...ps.nameOpts ];
      // each type.
      for(let type in props) {
        // each value for type.
        for(let pti in props[type]) {
          let v = props[type][pti];
          // checks if opt already exist on the state.
          // before adding.
          let exist = false;
          for(let ei=0; ei<state.nameOpts.length; ei++) {
            if(state.nameOpts[ei].value == v) {
              exist = true;
              break;
            } 
          }
          if(!exist)
            state.nameOpts.push({value: v, label: v, type: type});
        }
      }
      return state;
    })    
  }

  fetchPropertyKeys = () => {
    this.setState({ nameOpts: [] }); // reset opts

    if (this.props.groupByState.type == 'event') {
      // // Todo(Dinesh): Temp. Add an action for fetching all event properties.
      // fetchProjectEventProperties(this.props.projectId, 'View Project', false) 
      //   .then((r) => this.addToNameOptsState(r.data))
      //   .catch(r => console.error("Failed fetching event properties on group by.", r));
      
      let eventNames = this.props.getSelectedEventNames();
      let fetches = [];
      for(let i=0; i < eventNames.length; i++) {
        fetches.push(fetchProjectEventProperties(this.props.projectId, eventNames[i], false));
      }

      Promise.all(fetches)
        .then((r) => { 
          // add response from each as opts for selector.
          for(let i=0; i<r.length; i++) this.addToNameOptsState(r[i].data);
        })
        .catch((r) => console.error("Failed fetching event properties on group by.", r))
    }

    if (this.props.groupByState.type == 'user') {
      fetchProjectUserProperties(this.props.projectId, false)
      .then((r) => this.addToNameOptsState(r.data))
      .catch(r => console.error("Failed fetching user property keys.", r));
    }
  }

  render() {
    return (
      <div style={{ width: '400px', marginTop: '15px', marginLeft: '20px' }}>
        <div style={{display: 'inline-block', width: '185px'}}>
          <Select
            onChange={this.props.onTypeChange}
            options={[...PROPERTY_TYPE_OPTS]}
            placeholder='Property Type'
          />
        </div>
        <div style={{display: 'inline-block', width: '185px', marginLeft: '10px'}}>
          <Select
            onChange={this.props.onNameChange}
            onFocus={this.fetchPropertyKeys}
            options={this.state.nameOpts}
            placeholder='Property Key'
          />
        </div>
      </div>
    );
  }
}

export default GroupBy;