import React, { Component } from 'react';
import { Bar, Line } from 'react-chartjs-2';
import { connect } from 'react-redux'
import {
  Button,
  Card,
  CardBody,
  CardColumns,
  CardHeader,
  Col,
  FormGroup,
  Row,
} from 'reactstrap';
import CreatableSelect from 'react-select/lib/Creatable';
import { fetchFactors } from "../../actions/factorsActions"
import { CustomTooltips } from '@coreui/coreui-plugin-chartjs-custom-tooltips';

const customSelectStyles = {
  multiValue: () => ({
    background: 'none',
    fontSize: '1.3em',
  }),
  multiValueRemove: () => ({
    display: 'none',
  }),
}

const chartOptions = {
  tooltips: {
    enabled: false,
    custom: CustomTooltips
  },
  maintainAspectRatio: false
};

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

@connect((store) => {
  return {
    currentProject: store.projects.currentProject,
    currentProjectEventNames : store.projects.currentProjectEventNames,
    factors: store.factors.factors,
  };
})

class Factor extends Component {
  setupQueryStates(projectEventNames) {
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

        var queryStates, allowNumberCreateState;
        [queryStates, allowNumberCreateState] = this.setupQueryStates(this.props.currentProjectEventNames)
        this.state = {
          collapse: true,
          fadeIn: true,
          timeout: 300,
          menuIsOpen: false,
          queryStates: queryStates,
          allowNumberCreateState: allowNumberCreateState,
          data: queryStates[0]['labels'],
          allowNewOption: queryStates[0]['allowNumberCreate'] ? this.allowNumberOption: this.disallowNewOption,
          noOptionsMessage: queryStates[0]['allowNumberCreate'] ? this.enterNumberMessage: this.noOptionsMessage,
          values: null,
        }
      }

      resetProject(projectEventNames) {
        var queryStates, allowNumberCreateState;
        [queryStates, allowNumberCreateState] = this.setupQueryStates(projectEventNames)
        this.setState({
          collapse: true,
          fadeIn: true,
          timeout: 300,
          menuIsOpen: false,
          queryStates: queryStates,
          allowNumberCreateState: allowNumberCreateState,
          data: queryStates[0]['labels'],
          allowNewOption: queryStates[0]['allowNumberCreate'] ? this.allowNumberOption: this.disallowNewOption,
          noOptionsMessage: queryStates[0]['allowNumberCreate'] ? this.enterNumberMessage: this.noOptionsMessage,
          values: null,
        })
      }

      shouldComponentUpdate(nextProps, nextState) {
        if(this.props.currentProject.value != nextProps.currentProject.value) {
          this.resetProject(nextProps.currentProjectEventNames);
        }
        return true
      }

      toggle() {
        this.setState({ collapse: !this.state.collapse });
      }

      toggleFade() {
        this.setState((prevState) => { return { fadeIn: !prevState }});
      }

      handleChange = (newValues: any, actionMeta: any) => {
        var nextState = 0;
        var numEnteredValues = newValues.length
        if (!!newValues && numEnteredValues > 0) {
          var currentEnteredOption = newValues[numEnteredValues - 1];
          if (!!currentEnteredOption['__isNew__'] && actionMeta.action === 'create-option') {
            currentEnteredOption['value'] = 0;
            currentEnteredOption['type'] = numericalValueType;
            currentEnteredOption['currentState'] = this.state.allowNumberCreateState;
            currentEnteredOption['nextState'] = this.state.queryStates[this.state.allowNumberCreateState]['nextState'];
          }
          nextState = currentEnteredOption['nextState'];
        }
        console.group('Value Changed');
        console.log(newValues);
        console.log(`action: ${actionMeta.action}`);
        console.groupEnd();
        // Create new options with updated values to unique number so that
        // repeatitions of the same options can be allowed.
        // Not expecting more than 1000 options.
        var nextData = [];
        this.state.queryStates[nextState]['labels'].forEach(function(element) {
          var shouldSkip = false;
          if (element.onlyOnce) {
            // Do not add options that can occur only once and has already occurred.
            var found = false
            newValues.forEach(function(entry) {
              if (entry.currentState === element.currentState &&
                  entry.label === element.label) {
                shouldSkip = true;
              }
            });
          }
          if (!shouldSkip) {
            let newElement = Object.assign({}, element);
            newElement['value'] = (1000 * numEnteredValues) + element['value'];
            nextData.push(newElement);
          }
        });
        this.setState({
          data: nextData,
          allowNewOption: this.state.queryStates[nextState]['allowNumberCreate'] ? this.allowNumberOption: this.disallowNewOption,
          noOptionsMessage: this.state.queryStates[nextState]['allowNumberCreate'] ? this.enterNumberMessage: this.noOptionsMessage,
          values: newValues,
        });
      };

      handleKeyDown = (event: SyntheticKeyboardEvent<HTMLElement>) => {
        console.log(event)
        switch (event.key) {
          case 'Enter':
          if (!this.state.menuIsOpen) {
            console.log('Factor query');
            console.log(event);
            console.log(this.state.values);
            event.preventDefault();
            this.factor()
          }
        }
      };

      factor = () => {
        console.log('Factor ' + JSON.stringify(this.state.values));

        var query = {
          userProperties: [],
          eventsWithProperties: [],
        }

        var nextExpectedTypes = [eventNameType];

        this.state.values.forEach(function(queryElement) {
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

      disallowNewOption = (inputOption, valueType, optionsType) => {
        return false;
      };

      noOptionsMessage = (inputValue) => {
        return 'No Options';
      };

      allowNumberOption = (inputOption, valueType, optionsType) => {
        return !!inputOption && !isNaN(inputOption);
      };

      enterNumberMessage = (inputValue) => {
        return 'Enter a valid number';
      };

      formatCreateLabel = (inputValue) => {
        return inputValue;
      };

      render() {
        var charts = [];
        if (!!this.props.factors.charts) {
          for (var i = 0; i < this.props.factors.charts.length; i++) {
            // note: we add a key prop here to allow react to uniquely identify each
            // element in this array. see: https://reactjs.org/docs/lists-and-keys.html
            var chartData = this.props.factors.charts[i];
            var chart;
            if (chartData.type === 'line') {
              var line = {
                labels: chartData.labels,
                datasets: chartData.datasets,
              };
              line.datasets[0].fill = false;
              line.datasets[0].lineTension = 0.1;
              line.datasets[0].backgroundColor = 'rgba(75,192,192,0.4)';
              line.datasets[0].borderColor = 'rgba(75,192,192,1)';
              line.datasets[0].borderCapStyle = 'butt';
              line.datasets[0].borderDash = [];
              line.datasets[0].borderDashOffset = 0.0;
              line.datasets[0].borderJoinStyle = 'miter';
              line.datasets[0].pointBorderColor = 'rgba(75,192,192,1)';
              line.datasets[0].pointBackgroundColor = '#fff';
              line.datasets[0].pointBorderWidth = 1;
              line.datasets[0].pointHoverRadius = 5;
              line.datasets[0].pointHoverBackgroundColor = 'rgba(75,192,192,1)';
              line.datasets[0].pointHoverBorderColor = 'rgba(220,220,220,1)';
              line.datasets[0].pointHoverBorderWidth = 2;
              line.datasets[0].pointRadius = 1;
              line.datasets[0].pointHitRadius = 10;
              chart = <Line data={line} options={chartOptions} />
            } else if (chartData.type === 'bar') {
              var bar = {
                labels: chartData.labels,
                datasets: chartData.datasets,
              };
              bar.datasets[0].backgroundColor = 'rgba(255,99,132,0.2)';
              bar.datasets[0].borderColor = 'rgba(255,99,132,1)';
              bar.datasets[0].borderWidth = 1;
              bar.datasets[0].hoverBackgroundColor = 'rgba(255,99,132,0.4)';
              bar.datasets[0].hoverBorderColor = 'rgba(255,99,132,1)';
              chart = <Bar data={bar} options={chartOptions} />
            }
            charts.push(
              <Card key={i}>
              <CardHeader>
              {chartData.header}
              </CardHeader>
              <CardBody>
              <div className="chart-wrapper">
                {chart}
              </div>
              </CardBody>
              </Card>
            )
          }
        }


        return (
          <div className='animated fadeIn'>

          <div>
          <Row>
          <Col xs='12' md='12'>
          <Card>
          <CardHeader>
          <strong>Goal</strong>
          </CardHeader>
          <CardBody>
          <Row>
          <Col md={{ size: '8', offset: 2 }}>
          <FormGroup>
          <div>
          <CreatableSelect
          onChange={this.handleChange}
          onKeyDown={this.handleKeyDown}
          isValidNewOption={this.state.allowNewOption}
          noOptionsMessage={this.state.noOptionsMessage}
          formatCreateLabel={this.formatCreateLabel}
          isMulti={true}
          onMenuOpen={() => this.setState({ menuIsOpen: true })}
          onMenuClose={() => this.setState({ menuIsOpen: false })}
          closeMenuOnSelect={false}
          styles={customSelectStyles}
          options={this.state.data}
          value={this.state.values}
          />
          </div>
          </FormGroup>
          </Col>
          </Row>
          <Row>
          <Col md={{ size: 'auto', offset: 5 }}>
          <Button block outline color='primary' onClick={this.factor}>Factor</Button>
          </Col>
          <Col md={{ size: 'auto'}} style={{display: 'none'}}>
          <Button block outline color='primary'>Paths!</Button>
          </Col>
          </Row>
          </CardBody>
          </Card>
          </Col>
          </Row>
          </div>

          <div>
          <CardColumns className="cols-2">
          {charts}
          </CardColumns>
          </div>

          </div>
        );
      }
    }

    export default Factor;
