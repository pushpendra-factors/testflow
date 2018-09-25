import React, { Component } from 'react';
import { connect } from 'react-redux'
import {
  Card,
  CardHeader,
  Col,
  Row,
} from 'reactstrap';
import CreatableSelect from 'react-select/lib/Creatable';
import { fetchFactors } from "../../actions/factorsActions"
import BarChartCard from './BarChartCard.js';
import LineChartCard from './LineChartCard.js';
import FunnelChartCard from './FunnelChartCard.js';
import QueryBuilderCard from './QueryBuilderCard';

const eventNameType = "eventName";
const eventPropertyStartType = "eventPropertyStart";
const userPropertyStartType = "userPropertyStart";
const toType = "to";
const eventPropertyNameType = "eventPropertyName";
const userPropertyNameType = "userPropertyName";
const numericalValueType = "numericalValue";
const stringValueType = "stringValue";
const operatorEquals = "equals";
const operatorGreaterThan = "greaterThan";
const operatorLesserThan = "lesserThan";

const cardColumnSetting = {
  size: '10',
  offset: '1'
};

const chartCardRowStyle = {
  marginTop: '50px',
  marginBottom: '2px',
  marginRight: '2px',
  marginLeft: '2px',
};

@connect((store) => {
  return {
    currentProject: store.projects.currentProject,
    factors: store.factors.factors,
  };
})

class Factor extends Component {
  getQueryStates(projectEventNames) {
    const queryStates = {
      0: {
        'labels': (projectEventNames ? (
          Array.from(projectEventNames,
            eventName => ({'label': eventName, 'value': eventName,
            currentState: 0, nextState: 1, type: eventNameType }))):
            []),
            'allowNumberCreate': false,
          },
          1: {
            'labels': [
              { label: 'Escape and Enter to close and search', isDisabled: true},
              { label: 'to', value: 1, currentState: 1, nextState: 0, onlyOnce: true, type: toType },
              { label: 'with event property', value: 2, currentState: 1, nextState: 2, type: eventPropertyStartType },
              { label: 'with user property', value: 3, currentState: 1, nextState: 3, type: userPropertyStartType },
            ],
            'allowNumberCreate': false,
          },
          2: {
            'labels': [
              { label: 'occurrence count equals', value: 1, currentState: 2, nextState: 4,
              type: eventPropertyNameType, property: 'occurrence', operator: operatorEquals },
              { label: 'occurrence count greater than', value: 2, currentState: 2, nextState: 4,
              type: eventPropertyNameType, property: 'occurrence', operator: operatorGreaterThan },
              { label: 'occurrence count lesser than', value: 3, currentState: 2, nextState: 4,
              type: eventPropertyNameType, property: 'occurrence', operator: operatorLesserThan },
            ],
            'allowNumberCreate': false,
          },
          3: {
            'labels': [
              { label: 'country equals', value: 1, currentState: 3, nextState: 5,
              type: userPropertyNameType, property: 'country', operator: operatorEquals },
              { label: 'age equals', value: 2, currentState: 3, nextState: 4,
              type: userPropertyNameType, property: 'age', operator: operatorEquals },
              { label: 'age greater than', value: 3, currentState: 3, nextState: 4,
              type: userPropertyNameType, property: 'age', operator: operatorGreaterThan },
              { label: 'age lesser than', value: 4, currentState: 3, nextState: 4,
              type: userPropertyNameType, property: 'age', operator: operatorLesserThan },
            ],
            'allowNumberCreate': false,
          },
          4: {
            'labels': [
              {label: 'Enter number', 'value': 0, isDisabled: true, 'type': numericalValueType},
            ],
            'allowNumberCreate': true,
            'nextState': 1,
          },
          5: {
            'labels': [
              { label: 'India', value: 1, currentState: 5, nextState: 1, 'type': stringValueType},
              { label: 'United States', value: 2, currentState: 5, nextState: 1, 'type': stringValueType},
              { label: 'France', value: 3, currentState: 5, nextState: 1, 'type': stringValueType}
            ],
          },
        };
        const allowNumberCreateState = 4;
        return [queryStates, allowNumberCreateState];
      }

      constructor(props) {
        super(props);
        this.toggle = this.toggle.bind(this);
        this.toggleFade = this.toggleFade.bind(this);
        this.factor = this.factor.bind(this);

        this.state = {
          collapse: true,
          fadeIn: true,
          timeout: 300,
        }
      }

      toggle() {
        this.setState({ collapse: !this.state.collapse });
      }

      toggleFade() {
        this.setState((prevState) => { return { fadeIn: !prevState }});
      }

      factor = (queryElements) => {
        console.log('Factor ' + JSON.stringify(this.state.values));

        var query = {
          userProperties: [],
          eventsWithProperties: [],
        }

        var nextExpectedTypes = [eventNameType];

        queryElements.forEach(function(queryElement) {
          if (nextExpectedTypes.length > 0 &&
            nextExpectedTypes.indexOf(queryElement.type) < 0) {
              console.error("Invalid Query: " + JSON.stringify(query));
              return;
            }

            switch(queryElement.type) {
              case eventNameType:
              // Create a new event and add it to query.
              var newEvent = {}
              newEvent["name"] = queryElement.label;
              newEvent["properties"] = [];
              query.eventsWithProperties.push(newEvent);
              nextExpectedTypes = [];
              break;
              case eventPropertyStartType:
              // Create a new event property condition.
              var newEventProperty = {}
              numEvents = query.eventsWithProperties.length;
              query.eventsWithProperties[numEvents - 1].properties.push(newEventProperty);
              nextExpectedTypes = [eventPropertyNameType];
              break;
              case userPropertyStartType:
              // Create a new event property condition type.
              var newUserProperty = {}
              query.userProperties.push(newUserProperty);
              nextExpectedTypes = [userPropertyNameType];
              break;
              case toType:
              nextExpectedTypes = [eventNameType];
              break;
              case eventPropertyNameType:
              var numEvents = query.eventsWithProperties.length;
              var currentEvent = query.eventsWithProperties[numEvents - 1];
              var numProperties = currentEvent.properties.length;
              var currentProperty = currentEvent.properties[numProperties - 1];
              currentProperty['property'] = queryElement.property;
              currentProperty['operator'] = queryElement.operator;
              nextExpectedTypes = [numericalValueType, stringValueType];
              break;
              case userPropertyNameType:
              var numProperties = query.userProperties.length;
              var currentProperty = query.userProperties[numProperties - 1];
              currentProperty['property'] = queryElement.property;
              currentProperty['operator'] = queryElement.operator;
              nextExpectedTypes = [numericalValueType, stringValueType];
              break;
              case numericalValueType:
              var numUserProperties = query.userProperties.length;
              var currentUserProperty = query.userProperties[numUserProperties - 1];
              if (!currentUserProperty || currentUserProperty.hasOwnProperty('value')) {
                var numEvents = query.eventsWithProperties.length;
                var currentEvent = query.eventsWithProperties[numEvents - 1];
                var numEventProperties = currentEvent.properties.length;
                var currentEventProperty = currentEvent.properties[numEventProperties - 1];
                currentEventProperty['value'] = parseFloat(queryElement.label);
              } else {
                currentUserProperty['value'] = parseFloat(queryElement.label);
              }
              nextExpectedTypes = [];
              break;
              case stringValueType:
              var numUserProperties = query.userProperties.length;
              var currentUserProperty = query.userProperties[numUserProperties - 1];
              if (!currentUserProperty || currentUserProperty.hasOwnProperty('value')) {
                var numEvents = query.eventsWithProperties.length;
                var currentEvent = query.eventsWithProperties[numEvents - 1];
                var numEventProperties = currentEvent.properties.length;
                var currentEventProperty = currentEvent.properties[numEventProperties - 1];
                currentEventProperty['value'] = queryElement.label;
              } else {
                currentUserProperty['value'] = queryElement.label;
              }
              nextExpectedTypes = [];
              break;
            }
          });

          if (nextExpectedTypes.length > 0) {
            console.error('Invalid Query: ' + JSON.stringify(query));
            return;
          }
          console.log('Fire Query: ' + JSON.stringify(query));
          this.props.dispatch(fetchFactors(this.props.currentProject.value,
            {query: query}));
          }

          render() {
            var charts = [];
            let resultElements;
            if (!!this.props.factors.charts) {
              for (var i = 0; i < this.props.factors.charts.length; i++) {
                // note: we add a key prop here to allow react to uniquely identify each
                // element in this array. see: https://reactjs.org/docs/lists-and-keys.html
                var chartData = this.props.factors.charts[i];
                if (chartData.type === 'line') {
                  charts.push(<Row style={chartCardRowStyle} key={i}><Col sm={cardColumnSetting}><LineChartCard chartData={chartData}/></Col></Row>)
                } else if (chartData.type === 'bar') {
                  charts.push(<Row style={chartCardRowStyle} key={i}><Col sm={cardColumnSetting}><BarChartCard chartData={chartData} key={i} /></Col></Row>)
                } else if (chartData.type === 'funnel') {
                  charts.push(<Row style={chartCardRowStyle} key={i}><Col sm={cardColumnSetting}><FunnelChartCard chartData={chartData}/></Col></Row>);
                }
              }
                resultElements = <Card>{charts}</Card>;
              }


              return (
                <div className='animated fadeIn'>

                <div>
                <Row>
                <Col xs='12' md='12'>
                <QueryBuilderCard getQueryStates={this.getQueryStates} onKeyDown={this.factor} />
                </Col>
                </Row>
                </div>

                {resultElements}

                </div>
              );
            }
          }

          export default Factor;
