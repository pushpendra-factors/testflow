import React, { Component } from 'react';
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

const queryBuilderStyles = {
  multiValue: () => ({
    background: 'none',
    fontSize: '1.3em',
  }),
  multiValueRemove: () => ({
    display: 'none',
  }),
}
const numericalValueType = "numericalValue";

@connect((store) => {
  return {
    currentProject: store.projects.currentProject,
    currentProjectEventNames : store.projects.currentProjectEventNames,
  };
})

class QueryBuilderCard extends Component {
      constructor(props) {
        super(props);

        var queryStates, allowNumberCreateState;
        [queryStates, allowNumberCreateState] = this.props.getQueryStates(
          this.props.currentProjectEventNames)
        this.state = {
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
        [queryStates, allowNumberCreateState] = this.props.getQueryStates(projectEventNames)
        this.setState({
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
              console.log(event);
              console.log(this.state.values);
              event.preventDefault();
              this.props.onKeyDown(this.state.values);
            }
          }
        };

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
          return (
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
            styles={queryBuilderStyles}
            options={this.state.data}
            value={this.state.values}
            />
            </div>
            </FormGroup>
            </Col>
            </Row>
            <Row>
            <Col md={{ size: 'auto', offset: 5 }}>
            <Button block outline color='primary' onClick={()=>{this.props.onKeyDown(this.state.values)}}>Factor</Button>
            </Col>
            <Col md={{ size: 'auto'}} style={{display: 'none'}}>
            <Button block outline color='primary'>Paths!</Button>
            </Col>
            </Row>
            </CardBody>
            </Card>
          )
        }
      }

      export default QueryBuilderCard;
