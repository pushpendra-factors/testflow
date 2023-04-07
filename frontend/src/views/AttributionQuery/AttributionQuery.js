import React, { Component } from "react";
import Select from "react-select";
import {
  Button,
  ButtonDropdown,
  ButtonToolbar,
  Col,
  DropdownItem,
  DropdownMenu,
  DropdownToggle,
  Form,
  Input,
  Modal,
  ModalBody,
  ModalFooter,
  ModalHeader,
  Row,
} from "reactstrap";
import { connect } from "react-redux";
import { bindActionCreators } from "redux";
import "react-date-range/dist/styles.css";
import "react-date-range/dist/theme/default.css";

import {
  fetchProjectEvents,
  runAttributionQuery,
} from "../../actions/projectsActions";
import {
  DASHBOARD_TYPE_WEB_ANALYTICS,
  DEFAULT_DATE_RANGE,
  DEFINED_DATE_RANGES,
  getEventsWithProperties,
  getEventWithProperties,
  getQueryPeriod,
  jsonToCSV,
  PRESENTATION_TABLE,
  PROPERTY_LOGICAL_OP_OPTS,
  PROPERTY_TYPE_OPTS,
  QUERY_CLASS_ATTRIBUTION,
  readableDateRange,
  sameDay,
} from "../Query/common";
import ClosableDateRangePicker from "../../common/ClosableDatePicker";
import { getReadableKeyFromSnakeKey, removeElementByIndex } from "../../util";
import TableChart from "../Query/TableChart";
import Loading from "../../loading";
import mt from "moment-timezone";
import moment from "moment";
import { createDashboardUnit } from "../../actions/dashboardActions";
import ConversionEvent from "../Query/ConversionEvent";
import Event from "../Query/Event";

const DEFAULT_LOOKBACK_DAYS = 14;
const SOURCE = "Source";
const CAMPAIGN = "Campaign";
const ATTRIBUTION_KEYS = [
  { label: SOURCE, value: SOURCE },
  { label: CAMPAIGN, value: CAMPAIGN },
];

const FIRST_TOUCH = "First_Touch";
const LAST_TOUCH = "Last_Touch";
const FIRST_TOUCH_NON_DIRECT = "First_Touch_ND";
const LAST_TOUCH_NON_DIRECT = "Last_Touch_ND";
const LINEAR_TOUCH = "Linear";
const LAST_CAMPAIGN_TOUCH = "Last Campaign Touch";

const ATTRIBUTION_METHODOLOGY = [
  { value: FIRST_TOUCH, label: "First Touch" },
  { value: LAST_TOUCH, label: "Last Touch" },
  { value: FIRST_TOUCH_NON_DIRECT, label: "First Touch Non-Direct" },
  { value: LAST_TOUCH_NON_DIRECT, label: "Last Touch Non-Direct" },
  { value: LINEAR_TOUCH, label: "Linear Touch" },
  { value: LAST_CAMPAIGN_TOUCH, label: "Last Campaign Touch" },
];

const IMPRESSIONS = "Impressions";
const CLICKS = "Clicks";
const SPEND = "Spend";

const CAMPAIGN_METRICS = [IMPRESSIONS, CLICKS, SPEND];

const NONE_OPT = { label: "None", value: "none" };
const MAX_LOOKBACK_DAYS = 30;
const LABEL_STYLE = { marginRight: "10px", fontWeight: "600", color: "#777" };

const mapStateToProps = (store) => {
  return {
    currentProjectId: store.projects.currentProjectId,
    dashboards: store.dashboards.dashboards,
    eventNames: store.projects.currentProjectEventNames,
  };
};

const mapDispatchToProps = (dispatch) => {
  return bindActionCreators(
    {
      fetchProjectEvents,
      createDashboardUnit,
    },
    dispatch
  );
};

class AttributionQuery extends Component {
  constructor(props) {
    super(props);

    this.state = {
      duringDateRange: [DEFAULT_DATE_RANGE],
      isPresentationLoading: false,
      present: false,
      result: null,
      resultError: null,
      resultMetricsBreakdown: null,
      resultMeta: null,

      conversionEvent: { name: "", properties: [] },
      conversionEventC: { name: "", properties: [] },
      linkedEvents: [],
      lookbackDays: DEFAULT_LOOKBACK_DAYS,
      attributionMethodology: NONE_OPT,
      attributionMethodologyC: "",
      attributionKey: NONE_OPT,
      attributionKeyF: [],

      showDashboardsList: false,
      showAddToDashboardModal: false,
      addToDashboardMessage: null,
      inputDashboardUnitTitle: null,
      selectedDashboardId: null,
      eventNamesLoaded: false,
      eventNamesLoadError: null,

      timeZone: null,
    };
  }

  componentWillMount() {
    this.props
      .fetchProjectEvents(this.props.currentProjectId)
      .then(() => {
        this.setState({
          eventNamesLoaded: true,
          timeZone: this.getCurrentTimeZone(),
        });
      })
      .catch((r) => {
        this.setState({
          eventNamesLoaded: true,
          eventNamesLoadError: r.paylaod,
        });
      });
  }

  getDisplayMetricsBreakdown(metricsBreakdown) {
    if (!metricsBreakdown) return;

    let result = { ...metricsBreakdown };
    for (let i = 0; i < result.headers.length; i++)
      result.headers[i] = getReadableKeyFromSnakeKey(result.headers[i]);

    return result;
  }

  getCurrentTimeZone() {
    return mt.tz.guess();
  }

  isLoaded() {
    return this.state.eventNamesLoaded;
  }

  validateQuery() {
    if (
      this.state.conversionEvent.name == null ||
      this.state.conversionEvent.name === ""
    ) {
      this.props.showError("No conversion event provided.");
      return false;
    }

    for (let i = 0; i < this.state.linkedEvents.length; i++) {
      if (
        this.state.linkedEvents[i].name === "" ||
        this.state.linkedEvents[i].name == null
      ) {
        this.props.showError("Invalid linked funnel event provided.");
        return false;
      }
    }

    if (
      this.state.attributionKey.value !== SOURCE &&
      this.state.attributionKey.value !== CAMPAIGN
    ) {
      this.props.showError("No attribution key provided.");
      return false;
    }

    if (
      this.state.attributionMethodology.value !== FIRST_TOUCH &&
      this.state.attributionMethodology.value !== LAST_TOUCH &&
      this.state.attributionMethodology.value !== FIRST_TOUCH_NON_DIRECT &&
      this.state.attributionMethodology.value !== LAST_TOUCH_NON_DIRECT &&
      this.state.attributionMethodology.value !== LINEAR_TOUCH
    ) {
      this.props.showError("No attribution methodology provided.");
      return false;
    }

    return true;
  }

  getQuery = () => {
    let query = {};
    query.cm = CAMPAIGN_METRICS;
    query.ce = getEventWithProperties(this.state.conversionEvent);
    query.lfe = getEventsWithProperties(this.state.linkedEvents);
    query.attribution_key = this.state.attributionKey.value;
    query.attribution_methodology = this.state.attributionMethodology.value;
    query.lbw = this.state.lookbackDays;
    let period = getQueryPeriod(this.state.duringDateRange[0]);
    query.from = period.from;
    query.to = period.to;

    // additional inputs for support
    query.ce_c = getEventWithProperties(this.state.conversionEventC);
    query.attribution_key_f = this.state.attributionKeyF;
    query.attribution_methodology_c = this.state.attributionMethodologyC;

    return query;
  };

  runQuery = () => {
    let valid = this.validateQuery();
    if (!valid) return;
    // Enable add to dashboard here.
    this.setState({ present: true });

    this.props.resetError();
    this.setState({ isPresentationLoading: true });
    let query = this.getQuery();
    runAttributionQuery(this.props.currentProjectId, query)
      .then((r) => {
        this.setState({
          result: r.data,
          resultMeta: r.data.meta,
          isResultLoading: false,
          isPresentationLoading: false,
          resultMetricsBreakdown: this.getDisplayMetricsBreakdown(r.data),
        });
      })
      .catch((err) => {
        console.log("error occurred while running query: ", err);
      });
  };

  getReadableAttributionMetricValue(key, value, meta) {
    if (value === null || value === undefined) return 0;
    if (typeof value != "number") return value;

    let rValue = value;
    let isFloat = value % 1 > 0;
    if (isFloat) rValue = value >= 1 ? value.toFixed(1) : value.toFixed(2);
    // no decimal points for value >= 1 and 2 decimal points < 1.
    if (meta && meta.currency && key.toLowerCase().indexOf("spend") > -1) {
      rValue = rValue + " " + meta.currency;
      return rValue;
    }
    if (isFloat) {
      return Number(rValue);
    }
    return rValue;
  }

  renderAttributionResultAsTable() {
    if (
      !this.state.resultMetricsBreakdown ||
      !this.state.resultMetricsBreakdown.headers ||
      !this.state.resultMetricsBreakdown.rows
    )
      return;

    let resultMetricsBreakdown = { ...this.state.resultMetricsBreakdown };
    for (let ri = 0; ri < resultMetricsBreakdown.rows.length; ri++) {
      for (let ci = 0; ci < resultMetricsBreakdown.rows[ri].length; ci++) {
        let key = resultMetricsBreakdown.headers[ci];
        let value = resultMetricsBreakdown.rows[ri][ci];
        if (typeof resultMetricsBreakdown.rows[ri][ci] == "object") {
          // For each funnel event, rMB.rows[][] is array object of size 1
          value = resultMetricsBreakdown.rows[ri][ci][0];
        }
        resultMetricsBreakdown.rows[ri][ci] =
          this.getReadableAttributionMetricValue(
            key,
            value,
            this.state.resultMeta
          );
      }
    }
    return (
      <Col md={12} style={{ marginTop: "50px" }}>
        <Row>
          <Col md={12}>
            <TableChart
              sort
              bigWidthUptoCols={1}
              queryResult={resultMetricsBreakdown}
            />
          </Col>
        </Row>
      </Col>
    );
  }

  handleDuringDateRangeSelect = (range) => {
    range.selected.label = null; // set null on custom range.
    if (
      sameDay(range.selected.endDate, new Date()) &&
      !sameDay(range.selected.startDate, new Date())
    ) {
      return;
    }
    this.setState({ duringDateRange: [range.selected] });
  };

  closeDatePicker = () => {
    this.setState({ showDatePicker: false });
  };

  toggleDatePickerDisplay = () => {
    this.setState((state) => ({ showDatePicker: !state.showDatePicker }));
  };

  handleMethodologyChange = (option) => {
    this.setState({ attributionMethodology: option });
  };

  handleAttributionKeyChange = (option) => {
    this.setState({ attributionKey: option });
  };

  handleLookbackWindowChange = (event) => {
    let lookbackDays;
    lookbackDays = event.value;
    this.setState({
      lookbackDays: lookbackDays,
    });
  };

  // Dashboard methods
  toggleDashboardsList = () => {
    this.setState((state) => ({
      showDashboardsList: !state.showDashboardsList,
    }));
  };

  toggleAddToDashboardModal = () => {
    this.setState((state) => ({
      showAddToDashboardModal: !state.showAddToDashboardModal,
      addToDashboardMessage: null,
    }));
  };

  selectDashboardToAdd = (event) => {
    let dashboardId = event.currentTarget.getAttribute("value");
    this.setState({ selectedDashboardId: dashboardId });
    this.toggleAddToDashboardModal();
  };

  renderDashboardDropdownOptions() {
    let dashboardsDropdown = [];
    for (let i = 0; i < this.props.dashboards.length; i++) {
      let dashboard = this.props.dashboards[i];
      if (dashboard && dashboard.name !== DASHBOARD_TYPE_WEB_ANALYTICS) {
        dashboardsDropdown.push(
          <DropdownItem
            onClick={this.selectDashboardToAdd}
            value={dashboard.id}
          >
            {dashboard.name}
          </DropdownItem>
        );
      }
    }
    return dashboardsDropdown;
  }

  setDashboardUnitTitle = (e) => {
    this.setState({ addToDashboardMessage: null });

    let title = e.target.value.trim();
    if (title === "") console.error("chart title cannot be empty");
    this.setState({ inputDashboardUnitTitle: title });
  };

  renderAddToDashboardModal() {
    return (
      <Modal
        isOpen={this.state.showAddToDashboardModal}
        toggle={this.toggleAddToDashboardModal}
        style={{ marginTop: "10rem" }}
      >
        <ModalHeader toggle={this.toggleAddToDashboardModal}>
          Confirm add to Dashboard
        </ModalHeader>

        <ModalBody style={{ padding: "25px 35px" }}>
          <div style={{ textAlign: "center", marginBottom: "15px" }}>
            <span
              style={{ display: "inline-block" }}
              className="fapp-error"
              hidden={this.state.addToDashboardMessage == null}
            >
              {this.state.addToDashboardMessage}
            </span>
          </div>
          <Form>
            <span className="fapp-label">Title</span>
            <Input
              className="fapp-input"
              type="text"
              placeholder="Your Title"
              onChange={this.setDashboardUnitTitle}
            />
          </Form>
        </ModalBody>

        <ModalFooter
          style={{
            borderTop: "none",
            paddingBottom: "30px",
            paddingRight: "35px",
          }}
        >
          <Button outline color="success" onClick={this.addToDashboard}>
            Add
          </Button>
          <Button
            outline
            color="danger"
            onClick={this.toggleAddToDashboardModal}
          >
            Cancel
          </Button>
        </ModalFooter>
      </Modal>
    );
  }

  addToDashboard = () => {
    if (
      this.state.inputDashboardUnitTitle === null ||
      this.state.inputDashboardUnitTitle === ""
    ) {
      return;
    }
    let queryUnit = {};
    queryUnit.cl = QUERY_CLASS_ATTRIBUTION;
    queryUnit.query = this.getQuery();

    let metricBreakdownQueryUnit = { ...queryUnit };
    metricBreakdownQueryUnit.meta = { metrics_breakdown: true };

    let payload = {
      presentation: PRESENTATION_TABLE,
      query: metricBreakdownQueryUnit,
      title: this.state.inputDashboardUnitTitle,
    };

    this.props
      .createDashboardUnit(
        this.props.currentProjectId,
        this.state.selectedDashboardId,
        payload
      )
      .catch(() =>
        console.error(
          "Failed adding to attribution metrics breakdown to dashboard."
        )
      );

    this.setState({ inputDashboardUnitTitle: null });
    // close modal.
    this.toggleAddToDashboardModal();
  };

  renderDownloadButton = () => {
    return (
      <button
        className="btn btn-primary ml-1"
        style={{ fontWeight: 500, marginLeft: "150px" }}
        onClick={() => jsonToCSV(this.state.result, "", "factors_attribution")}
      >
        Download
      </button>
    );
  };

  // Linked funnel Event Methods
  addEvent = () => {
    this.setState((prevState) => {
      let state = { ...prevState };
      state.linkedEvents = [...prevState.linkedEvents];
      // init with default state for each event row.
      state.linkedEvents.push(this.getDefaultEventState());
      return state;
    });
  };

  onEventStateChange(option, index) {
    this.setState((prevState) => {
      let state = { ...prevState };
      state.linkedEvents = [...prevState.linkedEvents];
      state.linkedEvents[index] = { name: option.value, properties: [] };
      return state;
    });
  }

  addProperty(eventIndex) {
    this.setState((prevState) => {
      let state = { ...prevState };
      state.linkedEvents = [...prevState.linkedEvents];
      // init with default state for each property row by event index.
      state.linkedEvents[eventIndex].properties.push(
        this.getDefaultPropertyState()
      );
      return state;
    });
  }

  setPropertyAttr = (eventIndex, propertyIndex, attr, value) => {
    this.setState((prevState) => {
      let state = { ...prevState };
      state.linkedEvents[eventIndex].properties = [
        ...prevState.linkedEvents[eventIndex].properties,
      ];
      state.linkedEvents[eventIndex]["properties"][propertyIndex][attr] = value;
      return state;
    });
  };

  onPropertyEntityChange = (eventIndex, propertyIndex, value) => {
    this.setPropertyAttr(eventIndex, propertyIndex, "entity", value);
    this.setPropertyAttr(eventIndex, propertyIndex, "name", "");
    this.setPropertyAttr(eventIndex, propertyIndex, "value", "");
    this.setPropertyAttr(eventIndex, propertyIndex, "valueType", "");
  };

  onPropertyLogicalOpChange = (eventIndex, propertyIndex, value) => {
    this.setPropertyAttr(eventIndex, propertyIndex, "logicalOp", value);
  };

  onPropertyNameChange = (eventIndex, propertyIndex, value) => {
    this.setPropertyAttr(eventIndex, propertyIndex, "name", value);
    this.setPropertyAttr(eventIndex, propertyIndex, "value", "");
  };

  onPropertyOpChange = (eventIndex, propertyIndex, value) => {
    this.setPropertyAttr(eventIndex, propertyIndex, "op", value);
    this.setPropertyAttr(eventIndex, propertyIndex, "value", "");
  };

  onPropertyValueChange = (eventIndex, propertyIndex, value, type) => {
    this.setPropertyAttr(eventIndex, propertyIndex, "value", value);
    this.setPropertyAttr(eventIndex, propertyIndex, "valueType", type);
  };

  getEventNames = () => {
    return this.state.linkedEvents.map((e) => {
      return e.name;
    });
  };

  remove = (arrayKey, index) => {
    this.setState((pState) => {
      let state = { ...pState };
      state[arrayKey] = removeElementByIndex(state[arrayKey], index);
      return state;
    });
  };

  removeEventProperty = (eventIndex, propertyIndex) => {
    this.setState((pState) => {
      let state = { ...pState };
      state["linkedEvents"][eventIndex]["properties"] = removeElementByIndex(
        state["linkedEvents"][eventIndex]["properties"],
        propertyIndex
      );
      return state;
    });
  };

  renderLinkedEventsWithProperties() {
    let linkedEvents = [];
    for (let i = 0; i < this.state.linkedEvents.length; i++) {
      linkedEvents.push(
        <Event
          index={i}
          key={"linkedEvents_" + i}
          projectId={this.props.currentProjectId}
          nameOpts={this.props.eventNames}
          eventState={this.state.linkedEvents[i]}
          remove={() => this.remove("linkedEvents", i)}
          removeProperty={(propertyIndex) =>
            this.removeEventProperty(i, propertyIndex)
          }
          // event handlers.
          onNameChange={(value) => this.onEventStateChange(value, i)}
          // property handlers.
          onAddProperty={() => this.addProperty(i)}
          onPropertyEntityChange={this.onPropertyEntityChange}
          onPropertyLogicalOpChange={this.onPropertyLogicalOpChange}
          onPropertyNameChange={this.onPropertyNameChange}
          onPropertyOpChange={this.onPropertyOpChange}
          onPropertyValueChange={this.onPropertyValueChange}
        />
      );
    }

    let addEventButton = (
      <Row style={{ marginBottom: "15px" }}>
        <Col xs="12" md="12">
          <Button
            outline
            color="primary"
            onClick={this.addEvent}
            style={{ marginTop: "3px" }}
          >
            + LinkedEvent
          </Button>
        </Col>
      </Row>
    );

    return [linkedEvents, addEventButton];
  }

  getDefaultEventState() {
    return { name: "", properties: [] };
  }

  getDefaultPropertyState() {
    let entities = Object.keys(PROPERTY_TYPE_OPTS);
    let logicalOps = Object.keys(PROPERTY_LOGICAL_OP_OPTS);
    return {
      entity: entities[0],
      name: "",
      op: "equals",
      value: "",
      valueType: "",
      logicalOp: logicalOps[0],
    };
  }

  // Conversion Event Methods
  removeCEEventProperty = (propertyIndex) => {
    this.setState((pState) => {
      let state = { ...pState };
      state["conversionEvent"]["properties"] = removeElementByIndex(
        state["conversionEvent"]["properties"],
        propertyIndex
      );
      return state;
    });
  };

  onConversionEventStateChange(option) {
    this.setState((prevState) => {
      let state = { ...prevState };
      state.conversionEvent.name = option.value;
      return state;
    });
  }

  addCEProperty() {
    this.setState((prevState) => {
      let state = { ...prevState };
      // init with default state for each propety row by event index.
      state.conversionEvent.properties.push(this.getDefaultPropertyState());
      return state;
    });
  }

  setCEPropertyAttr = (propertyIndex, attr, value) => {
    this.setState((prevState) => {
      let state = { ...prevState };
      state.conversionEvent.properties = [
        ...prevState.conversionEvent.properties,
      ];
      state.conversionEvent["properties"][propertyIndex][attr] = value;
      return state;
    });
  };

  onCEPropertyEntityChange = (propertyIndex, value) => {
    this.setCEPropertyAttr(propertyIndex, "entity", value);
    this.setCEPropertyAttr(propertyIndex, "name", "");
    this.setCEPropertyAttr(propertyIndex, "value", "");
    this.setCEPropertyAttr(propertyIndex, "valueType", "");
  };

  onCEPropertyLogicalOpChange = (propertyIndex, value) => {
    this.setCEPropertyAttr(propertyIndex, "logicalOp", value);
  };

  onCEPropertyNameChange = (propertyIndex, value) => {
    this.setCEPropertyAttr(propertyIndex, "name", value);
    this.setCEPropertyAttr(propertyIndex, "value", "");
  };

  onCEPropertyOpChange = (propertyIndex, value) => {
    this.setCEPropertyAttr(propertyIndex, "op", value);
    this.setCEPropertyAttr(propertyIndex, "value", "");
  };

  onCEPropertyValueChange = (propertyIndex, value, type) => {
    this.setCEPropertyAttr(propertyIndex, "value", value);
    this.setCEPropertyAttr(propertyIndex, "valueType", type);
  };

  renderConversionEventWithProperties() {
    return (
      <ConversionEvent
        index={0}
        key={"events_0"}
        projectId={this.props.currentProjectId}
        nameOpts={this.props.eventNames}
        eventState={this.state.conversionEvent}
        removeProperty={(propertyIndex) =>
          this.removeCEEventProperty(propertyIndex)
        }
        // event handlers.
        onNameChange={(value) => this.onConversionEventStateChange(value)}
        // property handlers.
        onAddProperty={() => this.addCEProperty()}
        onPropertyEntityChange={this.onCEPropertyEntityChange}
        onPropertyLogicalOpChange={this.onCEPropertyLogicalOpChange}
        onPropertyNameChange={this.onCEPropertyNameChange}
        onPropertyOpChange={this.onCEPropertyOpChange}
        onPropertyValueChange={this.onCEPropertyValueChange}
      />
    );
  }

  getAllowedLookbackDays() {
    let allowedDays = [];
    for (let i = 0; i < MAX_LOOKBACK_DAYS + 1; i++) {
      allowedDays.push({ value: i, label: i });
    }
    return allowedDays;
  }

  render() {
    if (!this.isLoaded()) return <Loading />;
    return (
      <div>
        {this.renderConversionEventWithProperties()}
        {this.renderLinkedEventsWithProperties()}

        <Row style={{ marginBottom: "15px", marginTop: "-8px" }}>
          <Col xs="2" md="2" style={{ paddingTop: "5px" }}>
            <span style={LABEL_STYLE}> Attribution Key</span>
          </Col>
          <Col xs="8" md="8">
            <div
              className="fapp-select light"
              style={{ display: "inline-block", width: "250px" }}
            >
              <Select
                options={ATTRIBUTION_KEYS}
                onChange={this.handleAttributionKeyChange}
                placeholder="Select"
              />
            </div>
          </Col>
        </Row>

        <Row style={{ marginBottom: "15px" }}>
          <Col xs="2" md="2" style={{ paddingTop: "5px" }}>
            <span style={LABEL_STYLE}> Attribution Methodology</span>
          </Col>
          <Col xs="8" md="8">
            <div
              className="fapp-select light"
              style={{ display: "inline-block", width: "250px" }}
            >
              <Select
                options={ATTRIBUTION_METHODOLOGY}
                onChange={this.handleMethodologyChange}
                placeholder="Select Event"
              />
            </div>
          </Col>
        </Row>

        <Row style={{ marginBottom: "15px" }}>
          <Col xs="2" md="2" style={{ paddingTop: "5px" }}>
            <span style={LABEL_STYLE}>Lookback Window (in days)</span>
          </Col>
          <Col xs="8" md="8">
            <div
              style={{
                display: "inline-block",
                width: "168px",
                marginRight: "10px",
              }}
              className="fapp-select light"
            >
              <Select
                onChange={this.handleLookbackWindowChange}
                options={this.getAllowedLookbackDays()}
                placeholder={this.state.lookbackDays}
              />
            </div>
          </Col>
        </Row>

        <Row style={{ marginBottom: "15px" }}>
          <Col xs="2" md="2" style={{ paddingTop: "5px" }}>
            <span style={LABEL_STYLE}> Period </span>
          </Col>
          <Col xs="8" md="8">
            <Button
              outline
              style={{
                border: "1px solid #ccc",
                color: "grey",
                marginRight: "10px",
              }}
              onClick={this.toggleDatePickerDisplay}
            >
              <i className="fa fa-calendar" style={{ marginRight: "10px" }}></i>
              {readableDateRange(this.state.duringDateRange[0])}
            </Button>

            <div
              className="fapp-date-picker"
              hidden={!this.state.showDatePicker}
            >
              <ClosableDateRangePicker
                ranges={this.state.duringDateRange}
                onChange={this.handleDuringDateRangeSelect}
                staticRanges={DEFINED_DATE_RANGES}
                inputRanges={[]}
                minDate={new Date("01 Jan 2000 00:00:00 GMT")} // range starts from given date.
                maxDate={moment(new Date())
                  .subtract(1, "days")
                  .endOf("day")
                  .toDate()}
                closeDatePicker={this.closeDatePicker}
              />
              <button
                className="fapp-close-round-button"
                style={{
                  float: "right",
                  marginLeft: "0px",
                  borderLeft: "none",
                }}
                onClick={this.toggleDatePickerDisplay}
              >
                x
              </button>
            </div>
          </Col>
        </Row>

        <div style={{ width: "100%", textAlign: "center", marginTop: "15px" }}>
          <Button
            color="primary"
            style={{ fontSize: "0.9rem", padding: "8px 18px", fontWeight: 500 }}
            onClick={this.runQuery}
          >
            {" "}
            Run Query
          </Button>
        </div>

        <div
          hidden={!this.state.present}
          style={{
            borderTop: "1px solid rgb(221, 221, 221)",
            marginTop: "30px",
            marginLeft: "-60px",
            marginRight: "-60px",
          }}
        ></div>

        {/*Dashboard*/}
        <div
          style={{
            paddingLeft: "30px",
            paddingRight: "30px",
            paddingTop: "10px",
            minHeight: "500px",
          }}
        >
          <Row
            style={{ marginTop: "15px", marginRight: "10px" }}
            hidden={!this.state.present}
          >
            <Col xs="12" md="12">
              <ButtonToolbar className="pull-right">
                <ButtonDropdown
                  isOpen={this.state.showDashboardsList}
                  toggle={this.toggleDashboardsList}
                >
                  <DropdownToggle caret outline color="primary">
                    Add to dashboard
                  </DropdownToggle>
                  <DropdownMenu
                    style={{
                      height: "auto",
                      maxHeight: "210px",
                      overflowX: "scroll",
                    }}
                    right
                  >
                    {this.renderDashboardDropdownOptions()}
                  </DropdownMenu>
                </ButtonDropdown>
                {this.renderDownloadButton()}
              </ButtonToolbar>
            </Col>
          </Row>

          {this.state.isPresentationLoading ? (
            <Loading paddingTop="12%" />
          ) : null}
          <div
            className="animated fadeIn"
            hidden={this.state.isPresentationLoading}
            style={{ marginTop: "50px" }}
          >
            <Row> {this.renderAttributionResultAsTable()} </Row>
          </div>

          {this.renderAddToDashboardModal()}
        </div>
      </div>
    );
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(AttributionQuery);
