import React, {
  useState,
  useEffect,
  useCallback,
  useRef,
  useMemo
} from 'react';
import { connect, useDispatch, useSelector } from 'react-redux';
import {
  Row,
  Col,
  Menu,
  Dropdown,
  Button,
  Table,
  notification,
  Tabs,
  Badge,
  Switch,
  Modal,
  Space,
  Input,
  Tag,
  Collapse,
  Select,
  Form,
  message
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import {
  QUERY_TYPE_EVENT,
  INITIAL_SESSION_ANALYTICS_SEQ,
  QUERY_OPTIONS_DEFAULT_VALUE,
  TOTAL_EVENTS_CRITERIA,
  TOTAL_USERS_CRITERIA
} from 'Utils/constants';
import {
  DefaultDateRangeFormat,
  formatBreakdownsForQuery,
  formatFiltersForQuery,
  processBreakdownsFromQuery,
  processFiltersFromQuery
} from 'Views/CoreQuery/utils';
import {
  deleteGroupByForEvent,
  setGroupBy,
  delGroupBy,
  getUserPropertiesV2,
  resetGroupBy,
  getGroupProperties,
  getEventPropertiesV2,
  getGroups,
  fetchEventNames
} from 'Reducers/coreQuery/middleware';
import { reorderDefaultDomainSegmentsToTop } from 'Components/Profile/AccountProfiles/accountProfiles.helpers';
import { selectSegments } from 'Reducers/timelines/selectors';
import { GROUP_NAME_DOMAINS } from 'Components/GlobalFilter/FilterWrapper/utils';
import { getSavedSegments } from 'Reducers/timelines/middleware';
import { getSegmentColorCode } from 'Views/AppSidebar/appSidebar.helpers';
import { defaultSegmentIconsMapping } from 'Views/AppSidebar/appSidebar.constants';
import cx from 'classnames';
import ControlledComponent from 'Components/ControlledComponent/ControlledComponent';
import { paragon } from '@useparagon/connect/dist/src/index';
import useParagon from 'hooks/useParagon';
import { get, getHostUrl } from 'Utils/request';
import {
  fetchSavedWorkflows,
  saveWorkflow,
  updateWorkflow
} from 'Reducers/workflows';
import logger from 'Utils/logger';
import WorkflowTrigger from './trigger';
import MapComponent from './MapComponent';
import FactorsHubspotCompany from './Templates/FactorsHubspotCompany';
import FactorsApolloHubspotContacts from './Templates/FactorsApolloHubspotContacts';
import { TemplateIDs } from '../utils';
import FactorsSalesforceCompany from './Templates/FactorsSalesforceCompany';
import FactorsApolloSalesforceContacts from './Templates/FactorsApolloSalesforceContacts';
import WorkflowHubspotThumbnail from '../../../../../assets/images/workflow-hubspot-thumbnail.png';
import WorkflowCAPIThumbnail from '../../../../../assets/images/workflow-capi-thumbnail.png';
import QueryBlock from '../../Alerts/EventBasedAlert/QueryBlock';
import {
  defaultPropertyList,
  alertsGroupPropertyList
} from 'Components/QueryComposer/EventGroupBlock/utils';
import FactorsLinkedInCAPI from './Templates/FactorsLinkedInCAPI';

const host = getHostUrl();

const { Panel } = Collapse;
const SegmentIcon = (name) => defaultSegmentIconsMapping[name] || 'pieChart';

const WorkflowBuilder = ({
  setBuilderMode,
  groups,
  getGroups,
  fetchEventNames,
  activeProject,
  getSavedSegments,
  selectedTemp,
  fetchSavedWorkflows,
  saveWorkflow,
  updateWorkflow,
  alertId,
  editMode,
  setEditMode,
  eventUserPropertiesV2,
  eventPropertiesV2,
  groupProperties,
  userPropertiesV2,
  getGroupProperties
}) => {
  const configureRef = useRef(null);
  const [loading, setLoading] = useState(false);
  const [queries, setQueries] = useState([]);
  const [workflowName, setWorkflowName] = useState('');
  const [isTemplate, setIsTemplate] = useState(false);
  const [queryType, setQueryType] = useState(QUERY_TYPE_EVENT);
  const [activeGrpBtn, setActiveGrpBtn] = useState(QUERY_TYPE_EVENT);
  const [queryOptions, setQueryOptions] = useState({
    ...QUERY_OPTIONS_DEFAULT_VALUE,
    session_analytics_seq: INITIAL_SESSION_ANALYTICS_SEQ,
    date_range: { ...DefaultDateRangeFormat }
  });
  const [form] = Form.useForm();
  // Segment Support
  const [segmentType, setSegmentType] = useState('action_event');
  const [selectedSegment, setSelectedSegment] = useState('');
  const [segmentOptions, setSegmentOptions] = useState([]);
  const segments = useSelector(selectSegments);
  const [filterOptions, setFilterOptions] = useState([]);
  const [propertyMapMandatory, setPropertyMapMandatory] = useState([]);
  const [propertyMapAdditional, setPropertyMapAdditional] = useState([]);
  const [propertyMapAdditional2, setPropertyMapAdditional2] = useState([]);

  const [apolloFormDetails, setApolloFormDetails] = useState({
    ApiKey: '',
    PersonTitles: '',
    PersonSeniorities: '',
    MaxContacts: ''
  });
  const [showConfigureOptions, setShowConfigureOptions] = useState(false);

  // paragon hook and states
  const [state, setState] = useState({
    token: ''
  });
  const { user, error, isLoaded } = useParagon(state.token);

  const fetchToken = async () => {
    get(null, `${host}projects/${activeProject?.id}/paragon/auth`)
      .then((res) => {
        if (!res?.data) {
          logger.error('JWT Token not found');
          return;
        }
        setState((prev) => ({
          ...prev,
          token: res?.data
        }));
      })
      .catch((err) => {
        logger.error(err);
        message.error('Token not found!');
      });
  };
  useEffect(() => {
    // Authenticate();
    fetchToken();
  }, []);

  useEffect(() => {
    if (selectedTemp) {
      const queryData = [];
      const isTemplateWorkflow = !!selectedTemp?.is_workflow;
      if (
        (selectedTemp?.action_performed == 'action_event' ||
          isTemplateWorkflow) &&
        (selectedTemp?.workflow_config?.trigger?.event || selectedTemp?.event)
      ) {
        queryData.push({
          alias: '',
          label: isTemplateWorkflow
            ? selectedTemp?.workflow_config?.trigger?.event
            : selectedTemp?.event,
          filters: processFiltersFromQuery(
            isTemplateWorkflow
              ? selectedTemp?.workflow_config?.trigger?.filter
              : selectedTemp?.filters
          ),
          group: ''
        });
        setQueries(queryData);
      } else if (
        selectedTemp?.action_performed == 'action_segment_entry' ||
        selectedTemp?.action_performed == 'action_segment_exit'
      ) {
        setSegmentType(selectedTemp?.action_performed);
        setSelectedSegment(
          selectedTemp?.event || selectedTemp?.workflow_config?.trigger?.event
        );
        setQueries([]);
        setSegmentType('action_event');
      }
      if (
        selectedTemp?.workflow_config?.trigger?.event_level === '' ||
        selectedTemp?.event_level === '' ||
        selectedTemp?.workflow_config?.trigger?.event_level === 'events' ||
        selectedTemp?.event_level === 'events' ||
        selectedTemp?.workflow_config?.trigger?.event_level === 'account' ||
        selectedTemp?.event_level === 'account'
      ) {
        setActiveGrpBtn('events');
      } else {
        setActiveGrpBtn('users');
      }
      setWorkflowName(isTemplateWorkflow ? '' : selectedTemp?.title);
      setIsTemplate(isTemplateWorkflow);
      setShowConfigureOptions(!isTemplateWorkflow);
      setApolloFormDetails(selectedTemp?.addtional_configuration?.[0]);
    }
    return () => {
      setIsTemplate(false);
    };
  }, [selectedTemp]);

  // fetch segments and on Change functions
  useEffect(() => {
    getSavedSegments(activeProject?.id);
  }, [activeProject?.id]);

  const segmentsList = useMemo(
    () => reorderDefaultDomainSegmentsToTop(segments[GROUP_NAME_DOMAINS]) || [],
    [segments]
  );

  const renderOptions = (segment) => {
    const iconColor = getSegmentColorCode(segment?.name);
    const icon = SegmentIcon(segment?.name);
    return (
      <div className={cx('flex col-gap-1 items-center w-full')}>
        <ControlledComponent controller={icon != null}>
          <SVG name={icon} size={20} color={iconColor} />
        </ControlledComponent>
        {segment?.name}
      </div>
    );
  };

  const getSegmentNameFromId = (Id) => {
    const segmentName = segmentsList.find((segment) => segment?.id === Id);
    if (segmentName) return segmentName?.name;
    return '';
  };

  useEffect(() => {
    const segmentListWithLabels = segmentsList.map((segment) => ({
      value: segment?.id,
      label: renderOptions(segment)
    }));
    setSegmentOptions(segmentListWithLabels);
  }, [segmentsList]);

  useEffect(() => {
    if (!groups || Object.keys(groups).length === 0) {
      getGroups(activeProject?.id);
    }
  }, [activeProject?.id, groups]);

  useEffect(() => {
    fetchEventNames(activeProject?.id, true);
  }, [activeProject]);

  const groupsList = useMemo(() => {
    const listGroups = [];
    Object.entries(groups?.all_groups || {}).forEach(
      ([group_name, display_name]) => {
        listGroups.push([display_name, group_name]);
      }
    );
    return listGroups;
  }, [groups]);

  const getGroupPropsFromAPI = useCallback(
    async (group) => {
      if (!groupProperties[group]) {
        await getGroupProperties(activeProject.id, group);
      }
    },
    [activeProject.id, groupProperties]
  );

  const fetchGroupProperties = async () => {
    // separate call for $domain = All account group.
    getGroupPropsFromAPI(GROUP_NAME_DOMAINS);

    const missingGroups = Object.keys(groups?.all_groups || {}).filter(
      (group) => !groupProperties[group]
    );
    if (missingGroups && missingGroups?.length > 0) {
      await Promise.allSettled(
        missingGroups?.map((group) =>
          getGroupProperties(activeProject?.id, group)
        )
      );
    }
  };

  useEffect(() => {
    fetchGroupProperties();
  }, [activeProject?.id, groups, groupProperties]);

  useEffect(() => {
    let filterOptsObj = {};
    let eventGroup = '';
    let event = queries[0] || '';
    let groupAnalysis = activeGrpBtn;

    if (!groupAnalysis || groupAnalysis === 'users') {
      filterOptsObj = defaultPropertyList(
        eventPropertiesV2,
        eventUserPropertiesV2,
        groupProperties,
        eventGroup,
        groups?.all_groups,
        event
      );
    } else {
      filterOptsObj = alertsGroupPropertyList(
        eventPropertiesV2,
        userPropertiesV2,
        groupProperties,
        eventGroup,
        groups?.all_groups,
        event
      );
    }

    setFilterOptions(Object.values(filterOptsObj));
  }, [
    eventPropertiesV2,
    userPropertiesV2,
    groupProperties,
    eventUserPropertiesV2,
    groups
  ]);

  const queryChange = useCallback(
    (newEvent, index, changeType = 'add', flag = null) => {
      const queryupdated = [...queries];
      if (queryupdated[index]) {
        if (changeType === 'add') {
          if (
            JSON.stringify(queryupdated[index]) !== JSON.stringify(newEvent)
          ) {
            deleteGroupByForEvent(newEvent, index);
            resetGroupBy();
            // setEventPropertyDetails({});
          }
          queryupdated[index] = newEvent;
        } else if (changeType === 'filters_updated') {
          // dont remove group by if filter is changed
          queryupdated[index] = newEvent;
        } else {
          deleteGroupByForEvent(newEvent, index);
          resetGroupBy();
          queryupdated.splice(index, 1);
          // setEventPropertyDetails({});
        }
      } else {
        if (flag) {
          Object.assign(newEvent, { pageViewVal: flag });
        }
        queryupdated.push(newEvent);
      }
      setQueries(queryupdated);
    },
    [queries]
  );

  const queryList = () => {
    const blockList = [];
    queries.forEach((event, index) => {
      blockList.push(
        <div key={index}>
          <QueryBlock
            availableGroups={groupsList}
            index={index + 1}
            queryType={queryType}
            event={event}
            queries={queries}
            eventChange={queryChange}
            groupAnalysis={activeGrpBtn}
          />
        </div>
      );
    });

    if (queries.length < 1) {
      blockList.push(
        <div key='init'>
          <QueryBlock
            availableGroups={groupsList}
            queryType={queryType}
            index={queries.length + 1}
            queries={queries}
            eventChange={queryChange}
            groupBy={queryOptions.groupBy}
            groupAnalysis={activeGrpBtn}
          />
        </div>
      );
    }

    return blockList;
  };

  const onChangeSegmentType = (value) => {
    setSegmentType(value);
  };

  const onChangeSegment = (segment) => {
    setSelectedSegment(segment?.value);
  };

  const dropdownOptions = filterOptions?.map((item) => ({
    label: item.label,
    options: item.values
  }));

  const VerticalDivider = () => (
    <>
      <div className='fa-workflow_section--dot top' />
      <div className='fa-workflow_section--line' />
      <div className='fa-workflow_section--dot bottom' />
    </>
  );

  const isArrayAndObjectsNotEmpty = (arr) => {
    // Check if the array itself is not empty
    if (!Array.isArray(arr) || arr.length === 0) {
      return false;
    }

    // Check if each object in the array is not empty
    for (const obj of arr) {
      if (Object.keys(obj).length === 0 && obj.constructor === Object) {
        return false;
      }
    }

    return true;
  };

  const saveWorkflowFn = (value) => {
    let message_propertiesObj = {};
    let additional_config;
    if (
      selectedTemp?.id == TemplateIDs.FACTORS_HUBSPOT_COMPANY ||
      selectedTemp?.template_id == TemplateIDs.FACTORS_HUBSPOT_COMPANY ||
      selectedTemp?.id == TemplateIDs.FACTORS_SALESFORCE_COMPANY ||
      selectedTemp?.template_id == TemplateIDs.FACTORS_SALESFORCE_COMPANY
    ) {
      message_propertiesObj = {
        mandatory_properties: propertyMapMandatory,
        additional_properties_company: propertyMapAdditional
      };
    }
    if (
      selectedTemp?.id == TemplateIDs.FACTORS_APOLLO_HUBSPOT_CONTACTS ||
      selectedTemp?.template_id ==
        TemplateIDs.FACTORS_APOLLO_HUBSPOT_CONTACTS ||
      selectedTemp?.id == TemplateIDs.FACTORS_APOLLO_SALESFORCE_CONTACTS ||
      selectedTemp?.template_id ==
        TemplateIDs.FACTORS_APOLLO_SALESFORCE_CONTACTS
    ) {
      message_propertiesObj = {
        mandatory_properties: propertyMapMandatory,
        additional_properties_company: propertyMapAdditional,
        additional_properties_contact: propertyMapAdditional2
      };
      additional_config = [apolloFormDetails];
    }

    if (
      selectedTemp?.id == TemplateIDs.FACTORS_LINKEDIN_CAPI ||
      selectedTemp?.template_id == TemplateIDs.FACTORS_LINKEDIN_CAPI
    ) {
      additional_config = propertyMapMandatory;
    }

    const payload = {
      action_performed: segmentType,
      addtional_configuration: additional_config,
      alert_limit: 5,
      breakdown_properties: [],
      cool_down_time: 1800,
      event:
        segmentType == 'action_event' ? queries[0]?.label : selectedSegment,
      event_level: activeGrpBtn === 'events' ? 'account' : 'user',
      filters: formatFiltersForQuery(queries?.[0]?.filters),
      notifications: false,
      repeat_alerts: true,
      template_title:
        selectedTemp?.alert?.title || selectedTemp?.template_title,
      template_description:
        selectedTemp?.alert?.description || selectedTemp?.template_description,
      title: workflowName || '',
      description: workflowName || '',
      template_id: selectedTemp?.id || selectedTemp?.template_id,
      message_properties: message_propertiesObj
    };

    if (isArrayAndObjectsNotEmpty(propertyMapMandatory)) {
      if (!editMode) {
        saveWorkflow(activeProject?.id, payload)
          .then((res) => {
            fetchSavedWorkflows(activeProject?.id);
            setBuilderMode(false);
            setEditMode(false);
            notification.success({
              message: 'Workflow Saved',
              description: 'New workflow is created and saved successfully.'
            });
          })
          .catch((err) => message.error(err?.data?.error));
      } else {
        updateWorkflow(activeProject?.id, alertId, payload)
          .then((res) => {
            fetchSavedWorkflows(activeProject?.id);
            setBuilderMode(false);
            setEditMode(false);
            notification.success({
              message: 'Workflow Updated',
              description: 'Workflow is updated and saved successfully.'
            });
          })
          .catch((err) => message.error(err?.data?.error));
      }
    } else {
      message.error('Add mandatory properties');
    }
  };

  const returnIntegrationComponent = (workflowItem) => {
    if (
      workflowItem?.id == TemplateIDs.FACTORS_HUBSPOT_COMPANY ||
      workflowItem?.template_id == TemplateIDs.FACTORS_HUBSPOT_COMPANY
    ) {
      return (
        <FactorsHubspotCompany
          user={user}
          propertyMapMandatory={propertyMapMandatory}
          setPropertyMapMandatory={setPropertyMapMandatory}
          filterOptions={filterOptions}
          dropdownOptions={dropdownOptions}
          propertyMapAdditional={propertyMapAdditional}
          setPropertyMapAdditional={setPropertyMapAdditional}
          saveWorkflowFn={saveWorkflowFn}
          selectedTemp={selectedTemp}
          isTemplate={isTemplate}
        />
      );
    }
    if (
      workflowItem?.id == TemplateIDs.FACTORS_APOLLO_HUBSPOT_CONTACTS ||
      workflowItem?.template_id == TemplateIDs.FACTORS_APOLLO_HUBSPOT_CONTACTS
    ) {
      return (
        <FactorsApolloHubspotContacts
          user={user}
          propertyMapMandatory={propertyMapMandatory}
          setPropertyMapMandatory={setPropertyMapMandatory}
          filterOptions={filterOptions}
          dropdownOptions={dropdownOptions}
          propertyMapAdditional={propertyMapAdditional}
          setPropertyMapAdditional={setPropertyMapAdditional}
          saveWorkflowFn={saveWorkflowFn}
          selectedTemp={selectedTemp}
          isTemplate={isTemplate}
          setPropertyMapAdditional2={setPropertyMapAdditional2}
          propertyMapAdditional2={propertyMapAdditional2}
          apolloFormDetails={apolloFormDetails}
          setApolloFormDetails={setApolloFormDetails}
        />
      );
    }
    if (
      workflowItem?.id == TemplateIDs.FACTORS_SALESFORCE_COMPANY ||
      workflowItem?.template_id == TemplateIDs.FACTORS_SALESFORCE_COMPANY
    ) {
      return (
        <FactorsSalesforceCompany
          user={user}
          propertyMapMandatory={propertyMapMandatory}
          setPropertyMapMandatory={setPropertyMapMandatory}
          filterOptions={filterOptions}
          dropdownOptions={dropdownOptions}
          propertyMapAdditional={propertyMapAdditional}
          setPropertyMapAdditional={setPropertyMapAdditional}
          saveWorkflowFn={saveWorkflowFn}
          selectedTemp={selectedTemp}
          isTemplate={isTemplate}
        />
      );
    }
    if (
      workflowItem?.id == TemplateIDs.FACTORS_APOLLO_SALESFORCE_CONTACTS ||
      workflowItem?.template_id ==
        TemplateIDs.FACTORS_APOLLO_SALESFORCE_CONTACTS
    ) {
      return (
        <FactorsApolloSalesforceContacts
          user={user}
          propertyMapMandatory={propertyMapMandatory}
          setPropertyMapMandatory={setPropertyMapMandatory}
          filterOptions={filterOptions}
          dropdownOptions={dropdownOptions}
          propertyMapAdditional={propertyMapAdditional}
          setPropertyMapAdditional={setPropertyMapAdditional}
          saveWorkflowFn={saveWorkflowFn}
          selectedTemp={selectedTemp}
          isTemplate={isTemplate}
          setPropertyMapAdditional2={setPropertyMapAdditional2}
          propertyMapAdditional2={propertyMapAdditional2}
          apolloFormDetails={apolloFormDetails}
          setApolloFormDetails={setApolloFormDetails}
        />
      );
    }
    if (
      workflowItem?.template_id == TemplateIDs.FACTORS_LINKEDIN_CAPI ||
      workflowItem?.id == TemplateIDs.FACTORS_LINKEDIN_CAPI
    ) {
      return (
        <FactorsLinkedInCAPI
          user={user}
          propertyMapMandatory={propertyMapMandatory}
          setPropertyMapMandatory={setPropertyMapMandatory}
          saveWorkflowFn={saveWorkflowFn}
          selectedTemp={selectedTemp}
          isTemplate={isTemplate}
        />
      );
    }
    return null;

    return null;
  };

  const handleConfigure = () => {
    setShowConfigureOptions(true);
    setTimeout(() => {
      configureRef.current.scrollIntoView({ behavior: 'smooth' });
    }, 300);
  };

  return (
    <>
      <Row className='border-bottom--thin-2 pt-4 pb-4'>
        <Col span={12}>
          <div className='flex justify-start items-center'>
            <Button
              disabled={loading}
              type='text'
              size='large'
              onClick={() => setBuilderMode(false)}
              icon={<SVG name='arrowLeft' size={24} />}
            />
            <Input
              size='large'
              style={{ width: '300px' }}
              placeholder='Untitled Workflow '
              value={workflowName}
              onChange={(e) => setWorkflowName(e.target.value)}
              className='fa-input ml-4'
            />
          </div>
        </Col>
        <Col span={12}>
          <div className='flex justify-end'>
            {/* <div className={'flex items-center justify-end mr-4'}>
              <Text
                type={'title'}
                level={8}
                color={'grey'}
                extraClass={'m-0 mr-2'}
              >
                Not published
              </Text>
              <Switch
                checkedChildren='On'
                unCheckedChildren='OFF'
                size='large'
                onChange={(checked) => setTeamsEnabled(checked)}
                checked={teamsEnabled}
              />
            </div> */}

            {/* <Button
              size={'large'}
              disabled={loading}
              onClick={() => setBuilderMode(false)}
            >
              Cancel
            </Button> */}
            <Button
              size='large'
              disabled={loading}
              loading={loading}
              className='ml-2'
              type='primary'
              onClick={() => saveWorkflowFn()}
            >
              Save and Publish
            </Button>
          </div>
        </Col>
      </Row>

      <Row className='my-6 background-color--mono-color-1 border-radius--sm'>
        <Col span={24}>
          <div
            className='flex flex-col justify-center items-center'
            style={{ 'min-height': '500px', padding: '3%' }}
          >
            {/* trigger div */}
            <div
              className='relative border--thin-2 w-full border-radius--lg background-color--white flex flex-col fa-workflow_section'
              style={{ 'margin-top': '0%', padding: '3% 3%' }}
            >
              <Tag
                className='absolute top-0 mx-auto'
                style={{ 'margin-top': '-10px' }}
              >
                Trigger
              </Tag>

              <WorkflowTrigger
                onChangeSegmentType={onChangeSegmentType}
                segmentType={segmentType}
                selectedSegment={selectedSegment}
                onChangeSegment={onChangeSegment}
                segmentOptions={segmentOptions}
                queryList={queryList}
                activeGrpBtn={activeGrpBtn}
              />
            </div>

            {/* workflow config */}

            {queries?.length > 0 || selectedSegment !== '' ? (
              <>
                <VerticalDivider />

                <div
                  className='w-full relative border--thin-2 border-radius--lg background-color--white'
                  style={{ 'margin-top': '0%', 'min-height': '250px' }}
                >
                  <div
                    className='flex items-center justify-between'
                    style={{ padding: '3% 3%' }}
                  >
                    <Tag
                      className='absolute top-0 mx-auto'
                      style={{ 'margin-top': '-10px' }}
                    >
                      Action
                    </Tag>
                    <div className='pr-6'>
                      <Text
                        type='title'
                        level={7}
                        color='black'
                        weight='bold'
                        extraClass='m-0'
                      >
                        {isTemplate
                          ? selectedTemp?.title
                          : selectedTemp?.template_title}
                      </Text>
                      <Text
                        type='title'
                        level={7}
                        color='grey'
                        extraClass='mt-2'
                      >
                        {isTemplate
                          ? selectedTemp?.description
                          : selectedTemp?.template_description}
                      </Text>
                      <Button
                        type='primary'
                        extraClass='mt-2'
                        onClick={() => handleConfigure()}
                      >
                        Configure Action
                      </Button>
                    </div>
                    <div className='px-4 flex justify-center'>
                      <img
                        src={
                          activeGrpBtn !== 'users'
                            ? WorkflowHubspotThumbnail
                            : WorkflowCAPIThumbnail
                        }
                        style={{ height: '175px' }}
                      />
                    </div>
                  </div>

                  <div ref={configureRef}>
                    {showConfigureOptions &&
                      returnIntegrationComponent(selectedTemp)}
                  </div>
                </div>
              </>
            ) : (
              <></>
            )}
          </div>
        </Col>
      </Row>
    </>
  );
};

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  groupBy: state.coreQuery.groupBy.event,
  userPropertiesV2: state.coreQuery.userPropertiesV2,
  eventUserPropertiesV2: state.coreQuery.eventUserPropertiesV2,
  eventPropertiesV2: state.coreQuery.eventPropertiesV2,
  groupProperties: state.coreQuery.groupProperties,
  groups: state.coreQuery.groups
});

export default connect(mapStateToProps, {
  getGroups,
  fetchEventNames,
  getSavedSegments,
  fetchSavedWorkflows,
  saveWorkflow,
  updateWorkflow,
  getGroupProperties
})(WorkflowBuilder);
