import React, { Component } from 'react';
import CreatableSelect from 'react-select/lib/Creatable';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import Toggle from 'react-toggle'
import {
    Row,
    Col,
    Card,
    CardBody,
    CardHeader,
    Input,
    Button,
    InputGroup,
    InputGroupAddon,
    InputGroupText,
} from 'reactstrap';

import { getSDKAssetURL } from "../../util";
import { 
  fetchProjectSettings, 
  udpateProjectSettings,
  fetchFilters,
  createFilter,
  updateFilter,
  deleteFilter
} from '../../actions/projectsActions';
import FilterRecord from './FilterRecord';
import Loading from '../../loading';


const INT_SEGMENT_URI="/integrations/segment";

const FILTER_BUTTON_STATES = {
  success: "green",
  failure: "red",
  nochange: "#23282c"
}

const mapStateToProps = store => {
  return {
    projects: store.projects.projects,
    currentProjectId: store.projects.currentProjectId,
    currentProjectSettings: store.projects.currentProjectSettings,
    filters: store.projects.filters,
  }
}

const mapDispatchToProps = dispatch => {
  return bindActionCreators({ 
    fetchProjectSettings, 
    udpateProjectSettings,
    fetchFilters,
    createFilter,
    updateFilter,
    deleteFilter
  }, dispatch)
}

class Settings extends Component {
  constructor(props) {
    super(props);

    this.state = {
      createProject:{
        projectName : ""
      },
      sdkSettings: {
        loaded: false,
        error: null,
      },
      filterSettings: {
        loaded: false,
        error: null,
        formDomain: null,
        formDomainError: "",
        formExpr: null,
        formExprError: "",
        formName: "",
        formNameError: "",
        formSubmitSuccess: null,
        updatesByIndex: {}
      }
    }
  }

  shouldComponentUpdate(nextProps, nextState) {
    // decide to render or not.
    return true
  }

  componentDidUpdate(prevProps, prevState) {
    // set state based on action or prevProps.
    // use conditions, like this.props.prop1 == prevProps.prop1;
  }

  setSettingsState(prevState, update) {
      return { sdkSettings: { ...prevState.sdkSettings, ...update } };
  }

  setFilterSettingsState(prevState, update) {
    return { filterSettings: { ...prevState.filterSettings, ...update } };
  }

  componentWillMount() {
    if(!this.props.currentProjectId){
      return
    }

    this.props.fetchProjectSettings(this.props.currentProjectId)
      .then((response) => {
        this.setState(prevState => this.setSettingsState(prevState, { loaded: true }))
      })
      .catch((response) => {
        this.setState(prevState => this.setSettingsState(prevState, { loaded: true, error: response.payload }))
      });

    this.props.fetchFilters(this.props.currentProjectId)
      .then((r) => {
        this.setState(prevState => {
          let _state = { ...prevState };
          _state.filterSettings.loaded = true;
          _state.filters = Array.from(this.props.filters);
          return _state;
        })
      })
      .catch((err) => {
        this.setState(prevState => this.setFilterSettingsState(prevState, { loaded: true, error: err }))
      });
  }

  isAutoTrackEnabled() {
    return this.props.currentProjectSettings && 
      this.props.currentProjectSettings.auto_track;
  }

  isIntSegmentEnabled() {
    return this.props.currentProjectSettings && 
      this.props.currentProjectSettings.int_segment;
  }

  toggleAutoTrack = () =>  {
    this.props.udpateProjectSettings(this.props.currentProjectId, 
      { 'auto_track': !this.isAutoTrackEnabled() });
  }

  toggleIntSegment = () =>  {
    this.props.udpateProjectSettings(this.props.currentProjectId, 
      { 'int_segment': !this.isIntSegmentEnabled() });
  }

  getToken() {
    return this.props.projects[this.props.currentProjectId].token;
  }

  getPrivateToken() {
    return this.props.projects[this.props.currentProjectId].private_token;
  }

  getSDKScript() {
    let token = this.getToken();
    let assetURL = getSDKAssetURL();
    return '(function(c){var s=document.createElement("script");s.type="text/javascript";if(s.readyState){s.onreadystatechange=function(){if(s.readyState=="loaded"||s.readyState=="complete"){s.onreadystatechange=null;c()}}}else{s.onload=function(){c()}}s.src="'+assetURL+'";d=!!document.body?document.body:document.head;d.appendChild(s)})(function(){factors.init("'+token+'")})';
  }

  getSegmentWebhookURL() {
    return BUILD_CONFIG.backend_host+INT_SEGMENT_URI;
  }

  isDomainValid(d) {
    return d != undefined && d != "" && d.split(".").length >= 2;
  }

  handleFilterFormDomainChange = (domain) => {
    // Fix to return null instead of [] on clear.
    if (!domain || domain.value == undefined) domain=null;

    // Reset and evaluate.
    this.resetFilterFormErrors(); // reset domain error.
    if (!this.isDomainValid(domain.value)) {
      this.setState(prevState => this.setFilterSettingsState(prevState, {
        formDomainError: "Invalid domain", 
        formDomain: null,
        formSubmitSuccess: false,
      }));
    }

    this.setState(prevState => this.setFilterSettingsState(prevState, {formDomain: domain}));
  }

  getValidatedExpr(expr) {
    expr = expr.trim();
    if(expr == "") return "/";
    // add / as prefix if not.
    if(expr.indexOf("/") != 0) return "/"+expr;
    return expr
  }

  handleFilterFormExprChange = (expr) => {
    if (!expr || expr.value == undefined) expr=null;

    // can be done only for new records.
    let vexpr = this.getValidatedExpr(expr.value);
    expr.value = vexpr;
    expr.label = vexpr;

    this.setState(prevState => this.setFilterSettingsState(prevState, {formExpr: expr}));
  }

  handleFilterFormNameChange = (e) => {
    let name = e.target.value.trim();
    if(name == "") console.error("Event name cannot be empty");
    this.setState(prevState => this.setFilterSettingsState(prevState, {formName: name}));
  }

  
  
  setStateFilterEventName = (i, e) => {
    let name = e.target.value.trim();
    if(name == "") console.error("Event name cannot be empty");

    this.setState(prevState => {
      let updatesByIndex = {...prevState.filterSettings.updatesByIndex}
      if(updatesByIndex[i] == undefined) updatesByIndex[i] = {};
      updatesByIndex[i].name = name
      return this.setFilterSettingsState(prevState, {updatesByIndex: updatesByIndex});
    });
  }

  parseFilterExprURL(expr) {
    let parser = document.createElement('a');
    parser.href = "http://"+expr;
    return { host: parser.host, path: parser.pathname };
  }

  makeFilterExpr(host, path) {
    return host+path;
  }

  makeSelectOption(optStr) {
    return {
      value: optStr,
      label: optStr
    }
  }

  isFilterFormNameValid() {
    return this.state.filterSettings.formName && 
      this.state.filterSettings.formName != "";
  }

  isFilterFormExprValid() {
    return this.state.filterSettings.formExpr && 
      this.state.filterSettings.formExpr != null
  }
  
  isFilterFormDomainValid() {
    return this.state.filterSettings.formDomain &&
     this.state.filterSettings.formDomain != null
  }


  isFilterFormValid() {
    if (this.isFilterFormDomainValid() && this.isFilterFormExprValid() &&
      this.isFilterFormNameValid()) 
      return true;
    
    this.setState((prevState) => {
      let update = {};
      if(!this.isFilterFormExprValid()) 
        update["formExprError"] = "Invalid expression";
      if(!this.isFilterFormDomainValid()) 
        update["formDomainError"] = "Invalid domain";
      if(!this.isFilterFormNameValid()) 
        update["formNameError"] = "Invalid name";
      
      if(Object.values(update).length > 0) 
        update["formSubmitSuccess"] = false;

      return this.setFilterSettingsState(prevState, update);
    });

    return false;
  }

  makeFilterRequestPayload() {  
    return {
      name: this.state.filterSettings.formName,
      expr: this.makeFilterExpr(
          this.state.filterSettings.formDomain.value, 
          this.state.filterSettings.formExpr.value
        )
    }
  }

  resetFilterFormErrors() {
    this.setState(prevState => this.setFilterSettingsState(prevState, 
      {formDomainError: "", formExprError: "", formNameError: "", formSubmitSuccess: null}));
  }

  createFilter = () => {
    if (!this.isFilterFormValid()) return;
    let payload = this.makeFilterRequestPayload();

    // clear existing errors states.
    this.resetFilterFormErrors(); 

    this.props.createFilter(this.props.currentProjectId, payload)
      .then((r) => {
        this.setState((prevState) => this.setFilterSettingsState(prevState, {formSubmitSuccess: true}))
      })
      .catch((e) => {
        // Todo(Dinesh): Error message should be from backend.
        this.setState((prevState) => this.setFilterSettingsState(prevState, 
          {formExprError: "Invalid or duplicate expression.", formExpr: null, formSubmitSuccess: false}))
      });
  }

  formSelectCreateLabel = (inputValue) => {
    return inputValue;
  };

  setFilterUpdateSuccess(index, status) {
    this.setState((prevState) => {
      let updatesByIndex = {...prevState.filterSettings.updatesByIndex}
      if(updatesByIndex[index] == undefined) updatesByIndex[i] = {};
      updatesByIndex[index].success = status;
      return this.setFilterSettingsState(prevState, {updatesByIndex: updatesByIndex});
    });
  }

  
  updateFilterEventName = (index) => {
    if (this.state.filterSettings.updatesByIndex[index] == undefined ||
      this.state.filterSettings.updatesByIndex[index].name == undefined ||
      this.state.filterSettings.updatesByIndex[index].name.length == 0 ){
      this.setFilterUpdateSuccess(index, false);
      return;
    }

    let updated_name = this.state.filterSettings.updatesByIndex[index].name;
    if(this.props.filters[index].name != updated_name){
      let _filter = this.props.filters[index];
      this.props.updateFilter(_filter.project_id, _filter.id, { name: updated_name }, index)
        .then(() => this.setFilterUpdateSuccess(index, true))
        .catch(() => this.setFilterUpdateSuccess(index, false));
    }
  }

  getFilterDomainOptions() {
    let domains = [];
    for(let i in this.props.filters) {
      let purl = this.parseFilterExprURL(this.props.filters[i].expr);
      if(domains.indexOf(purl.host) == -1){
        domains.push(purl.host);
      }
    }

    return domains.map(this.makeSelectOption);
  }

  getFilterExprOptions() {
    let exprs = [];
    for(let i in this.props.filters) {
      let purl = this.parseFilterExprURL(this.props.filters[i].expr);
      if(exprs.indexOf(purl.path) == -1){
        exprs.push(purl.path);
      }
    }

    return exprs.map(this.makeSelectOption);
  }

  resetFilterForm = () => {
    this.setState(prevState => this.setFilterSettingsState(prevState, 
      {formDomain: null, formExpr: null, formName: "", formSubmitSuccess: null}));
    
    // clearing error as no values exist.
    this.resetFilterFormErrors();
  }

  getErrorDisplayState(errorMessage) {
    if(errorMessage && errorMessage.trim().length > 0) {
      return "block";
    }
    return "none";
  }

  deleteFilter = (index) => {
    // if local state and redux state name doesn't match.
    if(this.props.filters[index]) {
      let _filter = this.props.filters[index];
      this.props.deleteFilter(_filter.project_id, _filter.id, index)
        .catch(() => {
          console.log("delete failed.");
        });
    }
  }

  getFilterUpdateButtonColor = (index) => {
    let updates = this.state.filterSettings.updatesByIndex[index];
    if(updates != undefined && updates.success != undefined)
      return updates.success ? FILTER_BUTTON_STATES.success : FILTER_BUTTON_STATES.failure;
    
    return FILTER_BUTTON_STATES.nochange;
  }

  getFormCreateButtonColor = () => {
    if (this.state.filterSettings.formSubmitSuccess != null)
      return this.state.filterSettings.formSubmitSuccess ? FILTER_BUTTON_STATES.success : FILTER_BUTTON_STATES.failure;
    return FILTER_BUTTON_STATES.nochange;
  }

  getFilterEventName = (index) => {
    let updates = this.state.filterSettings.updatesByIndex[index]
    if(updates != undefined && updates.name != undefined) return updates.name;
    return this.props.filters[index].name;
  }

  isLoaded() {
    if(!this.props.currentProjectId){
      return true;
    }
    return this.state.sdkSettings.loaded &&
      this.state.filterSettings.loaded;
  }

  render() {
    if (!this.isLoaded()) return <Loading />;
    
    // Render empty view
    if(!this.props.currentProjectId){
      return (
        <div></div>
      )
    }

    let segmentWebhookURL = this.getSegmentWebhookURL();
    let segmentURLInputLength = segmentWebhookURL.length.toString();
    let segmentPrivateToken = this.getPrivateToken();
    let segmentTokenInputLength = (segmentPrivateToken.length+3).toString();

    return (
        <div className='animated fadeIn'>
          <div>
            <Row>
              <Col xs='12' md='12'>              
                <Card className="fapp-card"> 
                  <CardHeader className="fapp-card-header">
                    <strong>Web SDK Code</strong>
                  </CardHeader>
                  <CardBody>
                    <Row>
                      <Col md={{ size: '10' }}>                                            
                          <span id="code-snippet" className="sdk-code">{this.getSDKScript()}</span>
                      </Col>
                    </Row>
                  </CardBody>
                </Card>
                <Card className="fapp-card">
                  <CardHeader className="fapp-card-header">
                    <strong>Web SDK Settings</strong>
                  </CardHeader>
                  <CardBody>
                    <Toggle
                      checked={this.isAutoTrackEnabled()}
                      icons={false}
                      onChange={this.toggleAutoTrack} />
                    <div style={{display: "inline-block", position: "absolute"}}>
                      <span className="fapp-toggle-label">Auto-track</span>
                    </div>
                  </CardBody>
                </Card>
                <Card className="fapp-card">
                  <CardHeader className="fapp-card-header">
                    <strong>Integrations</strong>
                  </CardHeader>
                  <CardBody>
                    <Row>
                      <Col md="2" style={{paddingTop: "10px"}}>
                        <Toggle
                        checked={this.isIntSegmentEnabled()}
                        icons={false}
                        onChange={this.toggleIntSegment} />
                        <div style={{display: "inline-block", position: "absolute"}}><span className="fapp-toggle-label">Segment</span></div>
                      </Col>
                      <Col>
                        <div style={{padding: "5px 20px 5px 0", display: this.isIntSegmentEnabled()? "inline-block": "none"}}>
                          <InputGroup>
                            <InputGroupAddon addonType="prepend">
                              <InputGroupText style={{fontWeight: 600}}>Webhook URL</InputGroupText>
                            </InputGroupAddon>
                            <Input defaultValue={segmentWebhookURL} size={segmentURLInputLength} />
                          </InputGroup>
                        </div>
                        <div style={{padding: "5px 0", display: this.isIntSegmentEnabled()? "inline-block": "none"}}>
                          <InputGroup>
                            <InputGroupAddon addonType="prepend">
                              <InputGroupText style={{fontWeight: 600}}>Authorization Header</InputGroupText>
                            </InputGroupAddon>
                            <Input defaultValue={segmentPrivateToken} size={segmentTokenInputLength} />
                          </InputGroup>
                        </div>
                      </Col>
                    </Row>
                  </CardBody>
                </Card>
                <Card className="fapp-card">
                  <CardHeader className="fapp-card-header">
                    <strong>Web event name rules</strong>
                  </CardHeader>
                  <CardBody>
                    <span className="fapp-card-subtitle">Add filter</span>
                    <Row style={{padding: "10px 0"}}>
                      <Col md={{size: 4}}>
                        <div style={{height: "20px"}}>
                          <span className="fapp-error" style={{display: this.getErrorDisplayState(this.state.filterSettings.formDomainError)}}>{this.state.filterSettings.formDomainError}</span>
                        </div>
                        <CreatableSelect
                          value={this.state.filterSettings.formDomain}
                          onChange={this.handleFilterFormDomainChange}
                          options={this.getFilterDomainOptions()}
                          placeholder="Domain"
                          formatCreateLabel={this.formSelectCreateLabel}
                          ref={this.refFilterDomainSelect}
                        />
                      </Col>
                      <Col md={{size: 4}}>
                        <div style={{height: "20px"}}>
                          <span className="fapp-error" style={{display: this.getErrorDisplayState(this.state.filterSettings.formExprError)}}>{this.state.filterSettings.formExprError}</span>
                        </div>
                        <CreatableSelect
                          value={this.state.filterSettings.formExpr}
                          onChange={this.handleFilterFormExprChange}
                          options={this.getFilterExprOptions()}
                          placeholder="Filter Expression"
                          formatCreateLabel={this.formSelectCreateLabel}
                          ref={this.refFilterExprSelect}
                        />
                      </Col>
                      <Col md={{size: 3}}>
                        <div style={{height: "20px"}}>
                          <span className="fapp-error" style={{display: this.getErrorDisplayState(this.state.filterSettings.formNameError)}}>{this.state.filterSettings.formNameError}</span>
                        </div>
                        <Input type="text" placeholder="Event Name" onChange={this.handleFilterFormNameChange} value={this.state.filterSettings.formName} />
                      </Col>
                      <Col>
                        <Button className="fapp-inline-button" style={{marginTop: "20px", color: this.getFormCreateButtonColor()}} onClick={this.createFilter}>
                          <i className="icon-check"></i>
                        </Button>
                        <Button className="fapp-inline-button" style={{marginTop: "20px"}} onClick={this.resetFilterForm}>
                          <i className="icon-close"></i>
                        </Button>
                      </Col>
                    </Row>
                    <span className="fapp-card-subtitle" style={{display: this.props.filters.length > 0 ? "inline-block" : "none", marginTop: "15px", marginBottom: "20px"}}>Existing filters</span>
                    { 
                      // existing filters list.
                      this.props.filters.map((v, i) => {
                        let exprURL = this.parseFilterExprURL(v.expr);
                        return <FilterRecord name={this.getFilterEventName(i)} domain={exprURL.host} expr={exprURL.path} key={"filter_"+v.id} handleEventNameChange={(e) => this.setStateFilterEventName(i, e)} handleUpdate={() => this.updateFilterEventName(i)} handleDelete={() => this.deleteFilter(i)} getUpdateButtonColor={() => this.getFilterUpdateButtonColor(i)}/> 
                      })
                    }
                  </CardBody>
                </Card>
              </Col>
            </Row>
          </div>      
        </div>
    );
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(Settings);