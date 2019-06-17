import React, { Component } from 'react';
import Select from 'react-select';
import CreatableSelect from 'react-select/lib/Creatable';
import { Button } from 'reactstrap';

import { 
  fetchProjectEventProperties,
  fetchProjectUserProperties,
} from '../../actions/projectsActions';
import { getSelectedOpt, QUERY_TYPE_ANALYTICS } from '../../util';
import { PROPERTY_TYPE_OPTS } from './common';

export const USER_PROPERTY_JOIN_TIME = '$joinTime'

const EXCLUDE_PROPS = [ USER_PROPERTY_JOIN_TIME ];

class GroupBy extends Component {
  constructor(props) {
    super(props);
    this.state = {
      nameOpts: [],
      isNameOptsLoading: false,
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
          if(!exist && EXCLUDE_PROPS.indexOf(v) == -1)
            state.nameOpts.push({value: v, label: v, type: type});
        }
      }
      return state;
    })    
  }

  fetchPropertyKeys = () => {
    this.setState({ nameOpts: [], isNameOptsLoading: true }); // reset opts

    if (this.props.groupByState.type == 'event') {      
      let eventNames = this.props.getSelectedEventNames();
      let fetches = [];
      for(let i=0; i < eventNames.length; i++) {
        fetches.push(fetchProjectEventProperties(this.props.projectId, eventNames[i], "", false));
      }

      Promise.all(fetches)
        .then((r) => { 
          // add response from each as opts for selector.
          for(let i=0; i<r.length; i++) this.addToNameOptsState(r[i].data);
          this.setState({ isNameOptsLoading: false });
        })
        .catch((r) => console.error("Failed fetching event properties on group by.", r))
    }

    if (this.props.groupByState.type == 'user') {
      fetchProjectUserProperties(this.props.projectId, QUERY_TYPE_ANALYTICS, "", false)
      .then((r) => { 
        this.addToNameOptsState(r.data);
        this.setState({ isNameOptsLoading: false });
      })
      .catch(r => console.error("Failed fetching user property keys.", r));
    }
  }

  render() {
    return (
      <div style={{ width: '450px', marginBottom: '15px' }}>
        <div style={{display: 'inline-block', width: '150px'}} className='fapp-select light'>
          <Select
            onChange={this.props.onTypeChange}
            options={this.props.getOpts()}
            placeholder='Property Type'
            value={getSelectedOpt(this.props.groupByState.type, PROPERTY_TYPE_OPTS)}
          />
        </div>
        <div style={{display: 'inline-block', width: '195px', marginLeft: '10px'}} className='fapp-select light'>
          <CreatableSelect
            onChange={this.props.onNameChange}
            onFocus={this.fetchPropertyKeys}
            options={this.state.nameOpts}
            placeholder='Select Property'
            value={getSelectedOpt(this.props.groupByState.name)}
            formatCreateLabel={(value) => (value)}
            isLoading={this.state.isNameOptsLoading}
          />
        </div>
        <button className='fapp-close-button' onClick={this.props.remove}>x</button>
      </div>
    );
  }
}

export default GroupBy;