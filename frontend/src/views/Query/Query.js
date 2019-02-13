import React, { Component } from 'react';
import {
  Button,
  Card,
  CardBody,
  CardHeader,
  Col,
  FormGroup,
  Row,
} from 'reactstrap';
import CreatableSelect from 'react-select/lib/Creatable';
import Loading from '../../loading';

const customStyles = {
  multiValue: () => ({
    background: 'none',
    fontSize: '1.3em',
  }),
  multiValueRemove: () => ({
    display: 'none',
  }),
}

const queryData = {
  0: {
    "labels": [
      { label: 'ListenedSong', value: 1, currentState: 0, nextState: 1},
      { label: 'RecordedSong', value: 2, currentState: 0, nextState: 1 },
      { label: 'SharedSong', value: 3, currentState: 0, nextState: 1 },
      { label: 'PurchasedPremium', value: 4, currentState: 0, nextState: 1 },
      { label: 'CreatedPlaylist', value: 5, currentState: 0, nextState: 1 },
    ],
    "allowNumberCreate": false,
  },
  1: {
    "labels": [
      {label: "Escape and Enter to close and search", isDisabled: true},
      { label: 'with property', value: 1, currentState: 1, nextState: 2 },
      { label: 'followed by event', value: 2, currentState: 1, nextState: 0}
    ],
    "allowNumberCreate": false,
  },
  2: {
    "labels": [
      { label: 'occurrence count equals', value: 1, currentState: 2, nextState: 3},
      { label: 'occurrence count greater than', value: 2, currentState: 2, nextState: 3},
      { label: 'occurrence count lesser than ', value: 3, currentState: 2, nextState: 3},
      { label: 'country equals', value: 4, currentState: 2, nextState: 4},
      { label: 'age equals', value: 5, currentState: 2, nextState: 3},
      { label: 'age greater than', value: 6, currentState: 2, nextState: 3},
      { label: 'age lesser than', value: 7, currentState: 2, nextState: 3},
    ],
    "allowNumberCreate": false,
  },
  3: {
    "labels": [
      {label: 'Enter number', "value": 0, isDisabled: true},
    ],
    "allowNumberCreate": true,
    "nextState": 1,
  },
  4: {
    "labels": [
      { label: 'India', value: 1, currentState: 4, nextState: 1},
      { label: 'United States', value: 2, currentState: 4, nextState: 1 },
      { label: 'France', value: 3, currentState: 4, nextState: 1}
    ],
    "allowNumberCreate": false,
  },
};

const allowNumberCreateState = 3;

class Query extends Component {
  constructor(props) {
    super(props);

    this.toggle = this.toggle.bind(this);
    this.toggleFade = this.toggleFade.bind(this);

    this.state = {
      collapse: true,
      fadeIn: true,
      timeout: 300,
      menuIsOpen: false,
      data: queryData[0]["labels"],
      allowNewOption: queryData[0]["allowNumberCreate"] ? this.allowNumberOption: this.disallowNewOption,
      noOptionsMessage: queryData[0]["allowNumberCreate"] ? this.enterNumberMessage: this.noOptionsMessage,
    }
  }

  toggle() {
    this.setState({ collapse: !this.state.collapse });
  }

  toggleFade() {
    this.setState((prevState) => { return { fadeIn: !prevState }});
  }

  handleChange = (newValue, actionMeta) => {
    var nextState = 0;
    var numEnteredValues = newValue.length
    if (!!newValue && numEnteredValues > 0) {
      var currentEnteredOption = newValue[numEnteredValues - 1];
      if (!!currentEnteredOption["__isNew__"] && actionMeta.action === "create-option") {
        currentEnteredOption["value"] = 0;
        currentEnteredOption["currentState"] = allowNumberCreateState;
        currentEnteredOption["nextState"] = queryData[allowNumberCreateState]["nextState"];
      }
      nextState = currentEnteredOption["nextState"];
    }
    console.group('Value Changed');
    console.log(newValue);
    console.log(`action: ${actionMeta.action}`);
    console.groupEnd();
    // Create new options with updated values to unique number so that
    // repeatitions of the same options can be allowed.
    // Not expecting more than 1000 options.
    var nextData = [];
    queryData[nextState]["labels"].forEach(function(element) {
      let newElement = Object.assign({}, element);
      newElement["value"] = (1000 * numEnteredValues) + element["value"];
      nextData.push(newElement);
    });
    this.setState({
      data: nextData,
      allowNewOption: queryData[nextState]["allowNumberCreate"] ? this.allowNumberOption: this.disallowNewOption,
      noOptionsMessage: queryData[nextState]["allowNumberCreate"] ? this.enterNumberMessage: this.noOptionsMessage,
    });
  };

  handleKeyDown = (event) => {
    console.log(event)
    switch (event.key) {
      case 'Enter':
        console.log("Enter");
        if (!this.state.menuIsOpen) {
          console.log("Query");
          event.preventDefault();
        }
    }
  };

  disallowNewOption = (inputOption, valueType, optionsType) => {
     return false;
  };

  noOptionsMessage = (inputValue) => {
     return "No Options";
  };

  allowNumberOption = (inputOption, valueType, optionsType) => {
     return !!inputOption && !isNaN(inputOption);
  };

  enterNumberMessage = (inputValue) => {
    return "Enter a valid number";
  };

  formatCreateLabel = (inputValue) => {
    return inputValue;
  };

  isLoaded() {
    return true;
  }

  render() {
    if(!this.isLoaded()) return <Loading />;

    /* 
    return (
      <div className="animated fadeIn">
      <Row>
      <Col xs="12" md="12">
      <Card className="fapp-search-card fapp-select">
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
        styles={customStyles}
        options={this.state.data}
        placeholder="Enter query."
      />
      </div>
      </FormGroup>
      </Col>
      </Row>
      <Row>
      <Col md={{ size: 'auto', offset: 5 }}>
      <Button block color="primary">Query</Button>
      </Col>
      <Col md={{ size: 'auto'}} style={{display: 'none'}}>
      <Button block color="primary">Paths!</Button>
      </Col>
      </Row>
      </CardBody>
      </Card>
      </Col>
      </Row>
      </div>
    );
    */

    return <div style={{marginTop: "10%"}}><p style={{color: "grey", fontSize: "35px", textAlign: "center", color: "#c0c0c0", fontWeight: "500", letterSpacing: "2px"}}>Coming Soon! We are on it.</p></div>
  }
}

export default Query;
