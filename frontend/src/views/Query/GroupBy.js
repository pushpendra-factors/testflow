import React, { Component } from 'react';
import Select from 'react-select';
import CreatableSelect from 'react-select/lib/Creatable';
import { Button } from 'reactstrap';

import { 
  fetchProjectEventProperties,
  fetchProjectUserProperties,
} from '../../actions/projectsActions';
import { getSelectedOpt, QUERY_TYPE_ANALYTICS, makeSelectOpt, makeSelectOpts, removeIndexIfExistsFromOptName } from '../../util';
import { PROPERTY_TYPE_OPTS, PROPERTY_TYPE_EVENT, PROPERTY_TYPE_USER, LABEL_STYLE, NUMERICAL_GROUP_BY_METHODS } from './common';
import {TYPE_NUMERICAL} from './Property'

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
      state.nameOpts.sort((a, b)=>a.label > b.label ? 1:-1 );
      return state;
    })    
  }

  fetchPropertyKeys = () => {
    this.setState({ nameOpts: [], isNameOptsLoading: true }); // reset opts

    if (this.props.groupByState.type == PROPERTY_TYPE_EVENT) {
      
      let fetches = [];
      if (this.showEventNameSelector() && this.props.groupByState.eventName != '') {
        // fetch properties of selected group by event name.
        let eventName = this.props.groupByState.eventName
        if (this.props.shouldAddIndexPrefix()) {
          eventName = removeIndexIfExistsFromOptName(eventName).name
        }
        fetches.push(fetchProjectEventProperties(this.props.projectId, eventName, "", false));
      } else {
        // fetch event properties of all selected event names.
        let eventNames = this.props.getSelectedEventNames();
        for(let i=0; i < eventNames.length; i++) {
          fetches.push(fetchProjectEventProperties(this.props.projectId, eventNames[i], "", false));
        }
      }
      
      Promise.all(fetches)
        .then((r) => { 
          // add response from each as opts for selector.
          for(let i=0; i<r.length; i++) this.addToNameOptsState(r[i].data);
          this.setState({ isNameOptsLoading: false });
        })
        .catch((r) => console.error("Failed fetching event properties on group by.", r))
    }

    if (this.props.groupByState.type == PROPERTY_TYPE_USER) {
      fetchProjectUserProperties(this.props.projectId, QUERY_TYPE_ANALYTICS, "", false)
      .then((r) => { 
        this.addToNameOptsState(r.data);
        this.setState({ isNameOptsLoading: false });
      })
      .catch(r => console.error("Failed fetching user property keys.", r));
    }
  }

  // show event name selector when group by event property
  // and show is true by other state of query.
  showEventNameSelector() {
    return this.props.isEventNameRequired() && this.props.groupByState.type == PROPERTY_TYPE_EVENT;
  }

  showGroupByMethod() {
    return this.props.groupByState.ptype == TYPE_NUMERICAL;
  }

  render() {
    return (
      <div style={{ width: '900px', marginBottom: '15px' }}>
        <div style={{display: 'inline-block', width: '150px', marginRight: '10px'}} className='fapp-select light'>
          <Select
            onChange={this.props.onTypeChange}
            options={this.props.getOpts()}
            placeholder='Property Type'
            value={getSelectedOpt(this.props.groupByState.type, PROPERTY_TYPE_OPTS)}
          />
        </div>
        <span style={LABEL_STYLE} hidden={!this.props.shouldAddIndexPrefix()}> at </span>
        <div style={{display: 'inline-block', width: '275px'}} className='fapp-select light' hidden={!this.props.shouldAddIndexPrefix()}>
          <Select
            onChange={this.props.onEventNameChange}
            options={makeSelectOpts(this.props.getSelectedEventNames(), this.props.shouldAddIndexPrefix(), this.props.groupByState.type == PROPERTY_TYPE_USER)}
            placeholder='Select Event'
            value={getSelectedOpt(this.props.groupByState.eventName)}
          />
        </div>
        <div style={{display: 'inline-block', width: '195px', marginLeft: '10px'}} className='fapp-select light'>
          <CreatableSelect
            onChange={this.props.onNameChange}
            onFocus={this.fetchPropertyKeys}
            options={this.state.nameOpts}
            placeholder={this.props.groupByState.type == PROPERTY_TYPE_EVENT ? 'Event Property': 'User Property'}
            value={getSelectedOpt(this.props.groupByState.name)}
            formatCreateLabel={(value) => (value)}
            isLoading={this.state.isNameOptsLoading}
          />
        </div>
        <div style={{display: 'inline-block', width: '150px', marginLeft: '10px'}} className='fapp-select light' 
            hidden={!this.showGroupByMethod()}>
          <Select
            onChange={this.props.onNumericalGroupByChange}
            options={this.props.getNumericalGroupByMethods()}
            placeholder='Method'
            value={getSelectedOpt(this.props.groupByState.method, NUMERICAL_GROUP_BY_METHODS)}
          />
        </div>
        <button className='fapp-close-button' onClick={this.props.remove}>x</button>
      </div>
    );
  }
}

export default GroupBy;
