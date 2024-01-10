import React, { useCallback, useEffect, useMemo, useState } from 'react';
import cx from 'classnames';
import get from 'lodash/get';
import cloneDeep from 'lodash/cloneDeep';
import {
  Table,
  Button,
  Spin,
  notification,
  Popover,
  Tabs,
  Avatar,
  Input
} from 'antd';
import { connect, useDispatch, useSelector } from 'react-redux';
import { bindActionCreators } from 'redux';
import { Text, SVG } from '../../factorsComponents';
import FaSelect from '../../FaSelect';
import { getUserPropertiesV2 } from '../../../reducers/coreQuery/middleware';
import PropertyFilter from '../AccountProfiles/PropertyFilter';
import MomentTz from '../../MomentTz';
import NoDataWithMessage from '../MyComponents/NoDataWithMessage';
import {
  fetchProjectSettingsV1,
  fetchProjectSettings,
  fetchMarketoIntegration,
  fetchBingAdsIntegration,
  udpateProjectSettings
} from '../../../reducers/global';
import {
  ALPHANUMSTR,
  DEFAULT_TIMELINE_CONFIG,
  formatFiltersForPayload,
  getPropType,
  iconColors,
  propValueFormat,
  sortStringColumn,
  sortNumericalColumn,
  formatReqPayload,
  getFiltersRequestPayload,
  getSelectedFiltersFromQuery
} from '../utils';
import {
  getProfileUsers,
  getProfileUserDetails,
  createNewSegment,
  getSavedSegments,
  updateSegmentForId,
  deleteSegment
} from '../../../reducers/timelines/middleware';
import _, { isEqual } from 'lodash';
import SearchCheckList from 'Components/SearchCheckList';
import { formatUserPropertiesToCheckList } from 'Reducers/timelines/utils';
import {
  PropTextFormat,
  convertGroupedPropertiesToUngrouped
} from 'Utils/dataFormatter';
import { fetchUserPropertyValues } from 'Reducers/coreQuery/services';
import ProfilesWrapper from '../ProfilesWrapper';
import { getUserOptions } from './userProfiles.helpers';
import {
  selectActiveSegment,
  selectTimelinePayload
} from 'Reducers/userProfilesView/selectors';
import {
  setTimelinePayloadAction,
  setActiveSegmentAction,
  setFiltersDirtyAction,
  setNewSegmentModeAction
} from 'Reducers/userProfilesView/actions';
import { useHistory, useLocation } from 'react-router-dom';
import UpgradeModal from '../UpgradeModal';
import RangeNudge from 'Components/GenericComponents/RangeNudge';
import { showUpgradeNudge } from 'Views/Settings/ProjectSettings/Pricing/utils';
import CommonBeforeIntegrationPage from 'Components/GenericComponents/CommonBeforeIntegrationPage';
import ControlledComponent from 'Components/ControlledComponent/ControlledComponent';
import {
  INITIAL_USER_PROFILES_FILTERS_STATE,
  moreActionsMode
} from '../AccountProfiles/accountProfiles.constants';
import { ProfilesSidebarIconsMapping } from 'Views/AppSidebar/appSidebar.constants';
import { checkFiltersEquality } from '../AccountProfiles/accountProfiles.helpers';
import SaveSegmentModal from '../AccountProfiles/SaveSegmentModal';
import DeleteSegmentModal from '../AccountProfiles/DeleteSegmentModal';
import RenameSegmentModal from '../AccountProfiles/RenameSegmentModal';
import UpdateSegmentModal from '../AccountProfiles/UpdateSegmentModal';
import { isOnboarded } from 'Utils/global';
import { PathUrls } from 'Routes/pathUrls';
import styles from './index.module.scss';
import { getSegmentColorCode } from 'Views/AppSidebar/appSidebar.helpers';
import truncateURL from 'Utils/truncateURL';

const userOptions = getUserOptions();

function UserProfiles({
  activeProject,
  contacts,
  createNewSegment,
  getSavedSegments,
  getProfileUsers,
  getUserPropertiesV2,
  fetchProjectSettingsV1,
  fetchProjectSettings,
  fetchMarketoIntegration,
  fetchBingAdsIntegration,
  currentProjectSettings,
  udpateProjectSettings,
  updateSegmentForId,
  deleteSegment
}) {
  const dispatch = useDispatch();
  const history = useHistory();
  const location = useLocation();
  const integration = useSelector(
    (state) => state.global.currentProjectSettings
  );
  const integrationV1 = useSelector((state) => state.global.projectSettingsV1);
  const { bingAds, marketo } = useSelector((state) => state.global);
  const { dashboards } = useSelector((state) => state.dashboard);
  const userPropertiesV2 = useSelector(
    (state) => state.coreQuery.userPropertiesV2
  );
  const { userPropNames } = useSelector((state) => state.coreQuery);
  const timelinePayload = useSelector((state) => selectTimelinePayload(state));
  const activeSegment = useSelector((state) => selectActiveSegment(state));
  // const showSegmentModal = useSelector((state) =>
  //   selectSegmentModalState(state)
  // );
  const { sixSignalInfo } = useSelector((state) => state.featureConfig);

  //// segments 2.0 selectors
  const { newSegmentMode, filtersDirty: areFiltersDirty } = useSelector(
    (state) => state.userProfilesView
  );
  const { projectDomainsList } = useSelector((state) => state.global);
  const [listSearchItems, setListSearchItems] = useState([]);
  const [searchBarOpen, setSearchBarOpen] = useState(false);
  const [searchDDOpen, setSearchDDOpen] = useState(false);
  const [loading, setLoading] = useState(true);
  const [checkListUserProps, setCheckListUserProps] = useState([]);
  const [showPopOver, setShowPopOver] = useState(false);
  const [tlConfig, setTLConfig] = useState(DEFAULT_TIMELINE_CONFIG);
  const [userValueOpts, setUserValueOpts] = useState({});
  const [isUpgradeModalVisible, setIsUpgradeModalVisible] = useState(false);
  const [errMsg, setErrMsg] = useState('');

  const [currentPage, setCurrentPage] = useState(1);
  const [currentPageSize, setCurrentPageSize] = useState(25);
  const [defaultSorterInfo, setDefaultSorterInfo] = useState({});

  // segments 2.0 state
  const [filtersExpanded, setFiltersExpanded] = useState(false);
  const [saveSegmentModal, setSaveSegmentModal] = useState(false);
  const [updateSegmentModal, setUpdateSegmentModal] = useState(false);
  const [selectedFilters, setSelectedFilters] = useState(
    INITIAL_USER_PROFILES_FILTERS_STATE
  );
  const [appliedFilters, setAppliedFilters] = useState(
    INITIAL_USER_PROFILES_FILTERS_STATE
  );
  const [showSegmentActions, setShowSegmentActions] = useState(false);
  const [moreActionsModalMode, setMoreActionsModalMode] = useState(null); // DELETE | RENAME

  const setFiltersDirty = useCallback(
    (value) => {
      dispatch(setFiltersDirtyAction(value));
    },
    [dispatch]
  );

  const setTimelinePayload = useCallback(
    (payload) => {
      dispatch(setTimelinePayloadAction(payload));
    },
    [dispatch]
  );

  const setActiveSegment = useCallback(
    (segmentPayload) => {
      dispatch(setActiveSegmentAction(segmentPayload));
    },
    [dispatch]
  );

  const displayTableProps = useMemo(() => {
    const tableProps = timelinePayload.segment_id
      ? activeSegment?.query?.table_props
      : currentProjectSettings?.timelines_config?.user_config?.table_props;
    return (
      tableProps?.filter((entry) => entry !== '' && entry !== undefined) || []
    );
  }, [currentProjectSettings, timelinePayload, activeSegment]);

  const restoreFiltersDefaultState = useCallback(
    (selectedAccount = INITIAL_USER_PROFILES_FILTERS_STATE.account) => {
      const initialFiltersStateWithSelectedAccount = {
        ...INITIAL_USER_PROFILES_FILTERS_STATE,
        account: selectedAccount
      };
      setSelectedFilters(initialFiltersStateWithSelectedAccount);
      setAppliedFilters(initialFiltersStateWithSelectedAccount);
      setFiltersExpanded(false);
      setFiltersDirty(false);
    },
    [setFiltersDirty]
  );

  const handleClearFilters = useCallback(() => {
    restoreFiltersDefaultState();
    const reqPayload = getFiltersRequestPayload({
      selectedFilters: INITIAL_USER_PROFILES_FILTERS_STATE,
      table_props: displayTableProps
    });
    getProfileUsers(activeProject.id, reqPayload);
  }, [
    activeProject.id,
    displayTableProps,
    getProfileUsers,
    restoreFiltersDefaultState
  ]);

  const handleSaveSegment = useCallback(
    async (segmentPayload) => {
      try {
        const response = await createNewSegment(
          activeProject.id,
          segmentPayload
        );
        if (response.type === 'SEGMENT_CREATION_FULFILLED') {
          notification.success({
            message: 'Success!',
            description: response?.payload?.message,
            duration: 3
          });
          setSaveSegmentModal(false);
          setUpdateSegmentModal(false);
          setFiltersDirty(false);
        }
        await getSavedSegments(activeProject.id);
      } catch (err) {
        notification.error({
          message: 'Error',
          description:
            err?.data?.error || 'Segment Creation Failed. Invalid Parameters.',
          duration: 3
        });
      }
    },
    [activeProject.id, createNewSegment, getSavedSegments, setFiltersDirty]
  );

  const disableNewSegmentMode = useCallback(() => {
    dispatch(setNewSegmentModeAction(false));
  }, [dispatch]);

  const handleCreateSegment = useCallback(
    (newSegmentName) => {
      const reqPayload = getFiltersRequestPayload({
        selectedFilters,
        table_props: displayTableProps,
        caller: 'user_profiles'
      });
      reqPayload.name = newSegmentName;
      reqPayload.type = 'All';
      handleSaveSegment(reqPayload);
      disableNewSegmentMode();
    },
    [
      selectedFilters,
      displayTableProps,
      handleSaveSegment,
      disableNewSegmentMode
    ]
  );

  const handleDeleteActiveSegment = useCallback(() => {
    deleteSegment({
      projectId: activeProject.id,
      segmentId: timelinePayload.segment_id
    })
      .then(() => {
        setMoreActionsModalMode(null);
        notification.success({
          message: 'Segment deleted successfully',
          duration: 5
        });
      })
      .finally(() => {
        dispatch(
          setTimelinePayloadAction({
            source: 'All',
            filters: [],
            segment_id: ''
          })
        );
        dispatch(setActiveSegmentAction({}));
      });
  }, [timelinePayload.segment_id, activeProject.id, deleteSegment]);

  const handleRenameSegment = useCallback(
    (name) => {
      updateSegmentForId(activeProject.id, timelinePayload.segment_id, {
        name
      }).then(() => {
        getSavedSegments(activeProject.id);
        setActiveSegment({ ...activeSegment, name });
        setMoreActionsModalMode(null);
        notification.success({
          message: 'Segment renamed successfully',
          duration: 5
        });
      });
    },
    [activeProject.id, timelinePayload.segment_id, activeSegment]
  );

  const handleUpdateSegmentDefinition = useCallback(() => {
    const reqPayload = getFiltersRequestPayload({
      selectedFilters,
      table_props: displayTableProps,
      caller: 'user_profiles'
    });
    updateSegmentForId(
      activeProject.id,
      timelinePayload.segment_id,
      reqPayload
    ).then(() => {
      getSavedSegments(activeProject.id);
      setUpdateSegmentModal(false);
      setFiltersDirty(false);
      notification.success({
        message: 'Segment updated successfully',
        duration: 5
      });
    });
  }, [
    selectedFilters,
    displayTableProps,
    updateSegmentForId,
    activeProject.id,
    timelinePayload.segment_id,
    getSavedSegments,
    setFiltersDirty
  ]);

  useEffect(() => {
    if (!timelinePayload.search_filter) {
      setListSearchItems([]);
    } else {
      const listValues = timelinePayload?.search_filter || [];
      setListSearchItems(_.uniq(listValues));
      setSearchBarOpen(true);
    }
  }, [timelinePayload?.search_filter]);

  useEffect(() => {
    if (currentProjectSettings?.timelines_config) {
      const timelinesConfig = {};
      timelinesConfig.disabled_events = [
        ...currentProjectSettings?.timelines_config?.disabled_events
      ];
      timelinesConfig.user_config = {
        ...DEFAULT_TIMELINE_CONFIG.user_config,
        ...currentProjectSettings?.timelines_config?.user_config
      };
      timelinesConfig.account_config = {
        ...DEFAULT_TIMELINE_CONFIG.account_config,
        ...currentProjectSettings?.timelines_config?.account_config
      };
      setTLConfig(timelinesConfig);
    }
  }, [currentProjectSettings?.timelines_config]);

  useEffect(() => {
    setTimeout(() => {
      setLoading(false);
    }, 1000);
  }, [activeProject]);

  useEffect(() => {
    fetchProjectSettingsV1(activeProject.id);
    fetchProjectSettings(activeProject.id);
    if (_.isEmpty(dashboards?.data)) {
      fetchBingAdsIntegration(activeProject?.id);
      fetchMarketoIntegration(activeProject?.id);
    }
  }, [activeProject]);

  useEffect(() => {
    getUserPropertiesV2(activeProject.id);
  }, [activeProject?.id]);

  const isIntegrationEnabled =
    integration?.int_segment ||
    integration?.int_adwords_enabled_agent_uuid ||
    integration?.int_linkedin_agent_uuid ||
    integration?.int_facebook_user_id ||
    integration?.int_hubspot ||
    integration?.int_salesforce_enabled_agent_uuid ||
    integration?.int_drift ||
    integration?.int_google_organic_enabled_agent_uuid ||
    integration?.int_clear_bit ||
    integrationV1?.int_completed ||
    bingAds?.accounts ||
    marketo?.status ||
    integrationV1?.int_slack ||
    integration?.lead_squared_config !== null ||
    integration?.int_client_six_signal_key ||
    integration?.int_factors_six_signal_key ||
    integration?.int_rudderstack;

  useEffect(() => {
    const tableProps = timelinePayload?.segment_id
      ? activeSegment?.query?.table_props
      : currentProjectSettings.timelines_config?.user_config?.table_props || [];
    const userPropertiesModified = [];
    if (userPropertiesV2) {
      convertGroupedPropertiesToUngrouped(
        userPropertiesV2,
        userPropertiesModified
      );
    }
    const userPropsWithEnableKey = formatUserPropertiesToCheckList(
      userPropertiesModified,
      tableProps.filter(
        (entry) => entry !== '' && entry !== undefined && entry !== null
      )
    );
    setCheckListUserProps(userPropsWithEnableKey);
  }, [
    currentProjectSettings,
    userPropertiesV2,
    activeSegment,
    timelinePayload
  ]);

  useEffect(() => {
    getSavedSegments(activeProject.id);
  }, [activeProject.id]);

  const headerClassStr =
    'fai-text fai-text__color--grey-2 fai-text__size--h7 fai-text__weight--bold';

  const { tableProperties, tableColumns } = useMemo(() => {
    const columns = [
      {
        title: <div className={headerClassStr}>Identity</div>,
        width: 280,
        dataIndex: 'identity',
        key: 'identity',
        fixed: 'left',
        ellipsis: true,
        sorter: (a, b) => sortStringColumn(a.identity.id, b.identity.id),
        render: (identity) => (
          <div className='flex items-center'>
            {identity.isAnonymous ? (
              <SVG
                name={`TrackedUser${identity.id?.match(/\d/)?.[0] || 0}`}
                size={24}
              />
            ) : (
              <Avatar
                size={24}
                className='userlist-avatar'
                style={{
                  backgroundColor: `${
                    iconColors[
                      ALPHANUMSTR.indexOf(identity.id.charAt(0).toUpperCase()) %
                        8
                    ]
                  }`,
                  fontSize: '16px'
                }}
              >
                {identity.id.charAt(0).toUpperCase()}
              </Avatar>
            )}
            <span className='ml-2 truncate'>
              {identity.isAnonymous ? 'New User' : identity.id}
            </span>
          </div>
        )
      }
    ];

    const tableProps = timelinePayload?.segment_id
      ? activeSegment?.query?.table_props
      : currentProjectSettings?.timelines_config?.user_config?.table_props ||
        [];

    const userPropertiesModified = [];
    if (userPropertiesV2) {
      convertGroupedPropertiesToUngrouped(
        userPropertiesV2,
        userPropertiesModified
      );
    }
    tableProps
      ?.filter((entry) => entry !== '' && entry !== undefined && entry !== null)
      ?.forEach((prop) => {
        const propDisplayName = userPropNames[prop]
          ? userPropNames[prop]
          : prop
            ? PropTextFormat(prop)
            : '';
        const propType = getPropType(userPropertiesModified, prop);
        columns.push({
          title: (
            <Text
              type='title'
              level={7}
              color='grey-2'
              weight='bold'
              extraClass='m-0'
              truncate
              charLimit={25}
            >
              {propDisplayName}
            </Text>
          ),
          dataIndex: prop,
          key: prop,
          width: 260,
          sorter: (a, b) =>
            propType === 'numerical'
              ? sortNumericalColumn(a[prop], b[prop])
              : sortStringColumn(a[prop], b[prop]),
          render: (value) => {
            const formattedValue =
              propValueFormat(prop, value, propType) || '-';
            const urlTruncatedValue = truncateURL(
              formattedValue,
              projectDomainsList
            );
            return (
              <Text
                type='title'
                level={7}
                extraClass='m-0'
                truncate
                toolTipTitle={formattedValue}
              >
                {urlTruncatedValue}
              </Text>
            );
          }
        });
      });

    columns.push({
      title: <div className={headerClassStr}>Last Activity</div>,
      dataIndex: 'lastActivity',
      key: 'lastActivity',
      width: 200,
      align: 'right',
      sorter: {
        compare: (a, b) => sortStringColumn(a.lastActivity, b.lastActivity),
        multiple: 2
      },
      render: (item) => MomentTz(item).fromNow()
    });

    columns.forEach((column) => {
      if (column.key === defaultSorterInfo?.key) {
        column.sortOrder = defaultSorterInfo?.order;
      } else {
        delete column.sortOrder;
      }
    });
    const hasSorter = columns.find((item) =>
      ['ascend', 'descend'].includes(item.sortOrder)
    );
    if (!hasSorter) {
      columns.forEach((column) => {
        if (['engagement', 'lastActivity'].includes(column.key)) {
          column.defaultSortOrder = 'descend';
          return;
        }
      });
    }
    return { tableProperties: tableProps, tableColumns: columns };
  }, [
    contacts?.data,
    currentProjectSettings,
    timelinePayload,
    activeSegment,
    defaultSorterInfo,
    projectDomainsList
  ]);

  const getTableData = (data) => {
    const sortedData = data.sort(
      (a, b) => new Date(b.last_activity) - new Date(a.last_activity)
    );
    return sortedData.map((row) => ({
      ...row,
      ...row?.tableProps
    }));
  };

  const getUsers = useCallback(
    (payload) => {
      const shouldCache =
        location.state?.fromDetails && contacts?.data?.length > 0;
      if (payload.source && payload.source !== '' && !shouldCache) {
        setDefaultSorterInfo({ key: 'lastActivity', order: 'descend' });
        const formatPayload = { ...payload };
        formatPayload.filters = formatFiltersForPayload(payload?.filters) || [];
        const reqPayload = formatReqPayload(formatPayload, activeSegment);
        getProfileUsers(activeProject.id, reqPayload).then((response) => {
          if (response.type === 'FETCH_PROFILE_USERS_FAILED') {
            if (response.error.status === 400) {
              setErrMsg('400 Bad Request');
            } else if (response.error.status === 500) {
              setErrMsg(
                'The server encountered an internal error and could not complete your request'
              );
            }
          }

          if (response.type === 'FETCH_PROFILE_USERS_FULFILLED') {
            if (response.status === 200) {
              if (response.payload.length === 0) {
                setErrMsg('No User Profiles Found');
              }
            }
          }
        });
      }
      if (shouldCache) {
        setCurrentPage(location.state.currentPage);
        setCurrentPageSize(location.state.currentPageSize);
        setDefaultSorterInfo(location.state.activeSorter);
        const localeState = { ...history.location.state, fromDetails: false };
        history.replace({ state: localeState });
      }
    },
    [
      location.state?.fromDetails,
      location.state?.currentPage,
      location.state?.activeSorter,
      activeSegment,
      activeProject.id,
      history
    ]
  );

  useEffect(() => {
    getUsers(timelinePayload);
  }, [timelinePayload.source, timelinePayload.segment_id]);

  useEffect(() => {
    if (newSegmentMode === true) {
      restoreFiltersDefaultState();
    }
  }, [newSegmentMode]);

  useEffect(() => {
    if (
      location.state?.fromDetails === true &&
      location.state?.appliedFilters != null
    ) {
      setAppliedFilters(cloneDeep(location.state?.appliedFilters));
      setSelectedFilters(location.state?.appliedFilters);
      setFiltersExpanded(false);
      setFiltersDirty(true);
    } else {
      if (newSegmentMode === false) {
        if (
          Boolean(activeSegment?.name) === true &&
          activeSegment.query != null
        ) {
          const filters = getSelectedFiltersFromQuery({
            query: activeSegment.query,
            groupsList: [],
            caller: 'user_profiles'
          });
          setAppliedFilters(filters);
          setSelectedFilters(filters);
          setFiltersExpanded(false);
          setFiltersDirty(false);
        } else {
          restoreFiltersDefaultState();
        }
      }
    }
  }, [activeSegment, newSegmentMode]);

  const handlePropChange = (option) => {
    if (
      option.enabled ||
      checkListUserProps.filter((item) => item.enabled === true).length < 8
    ) {
      setCheckListUserProps((prev) => {
        const checkListProps = [...prev];
        const optIndex = checkListProps.findIndex(
          (obj) => obj.prop_name === option.prop_name
        );
        checkListProps[optIndex].enabled = !checkListProps[optIndex].enabled;
        checkListProps.sort((a, b) => {
          return (b?.enabled || 0) - (a?.enabled || 0);
        });
        return checkListProps;
      });
    } else {
      notification.error({
        message: 'Error',
        description: 'Maximum of 8 Table Properties Selection Allowed.',
        duration: 2
      });
    }
  };

  const applyTableProps = () => {
    if (timelinePayload?.segment_id?.length) {
      const updatedQuery = { ...activeSegment.query };
      updatedQuery.table_props =
        checkListUserProps
          ?.filter((item) => item.enabled === true)
          ?.map((item) => item?.prop_name)
          ?.filter(
            (entry) => entry !== '' && entry !== undefined && entry !== null
          ) || [];
      updateSegmentForId(activeProject.id, timelinePayload.segment_id, {
        query: { ...updatedQuery }
      })
        .then(() => getSavedSegments(activeProject.id))
        .then(() => setActiveSegment({ ...activeSegment, query: updatedQuery }))
        .finally(() => getUsers(timelinePayload));
    } else {
      const config = { ...tlConfig };
      config.user_config.table_props = checkListUserProps
        ?.filter((item) => item.enabled === true)
        ?.map((item) => item?.prop_name);
      udpateProjectSettings(activeProject.id, {
        timelines_config: { ...config }
      }).then(() => getUsers(timelinePayload));
    }
    setShowPopOver(false);
  };

  const handleDisableOptionClick = () => {
    setIsUpgradeModalVisible(true);
    setShowPopOver(false);
  };

  const popoverContent = () => (
    <Tabs defaultActiveKey='events' size='small'>
      <Tabs.TabPane
        tab={
          <span className='fa-activity-filter--tabname'>Table Properties</span>
        }
        key='props'
      >
        <SearchCheckList
          placeholder='Search Properties'
          mapArray={checkListUserProps}
          sortable
          updateList={setCheckListUserProps}
          titleKey='display_name'
          checkedKey='enabled'
          onChange={handlePropChange}
          showApply
          onApply={applyTableProps}
          handleDisableOptionClick={handleDisableOptionClick}
        />
      </Tabs.TabPane>
    </Tabs>
  );

  const selectedAccount = useMemo(() => {
    return { account: selectedFilters.account };
  }, [selectedFilters.account]);

  const setFiltersList = useCallback((filters) => {
    setSelectedFilters((curr) => {
      return {
        ...curr,
        filters
      };
    });
  }, []);

  const resetSelectedFilters = useCallback(() => {
    setSelectedFilters(appliedFilters);
  }, [appliedFilters]);

  const setSelectedAccount = useCallback((account) => {
    setSelectedFilters((current) => {
      return {
        ...current,
        account
      };
    });
  }, []);

  const setListEvents = useCallback((eventsList) => {
    setSelectedFilters((curr) => {
      return {
        ...curr,
        eventsList
      };
    });
  }, []);

  const setEventProp = useCallback((eventProp) => {
    setSelectedFilters((curr) => {
      return {
        ...curr,
        eventProp
      };
    });
  }, []);

  const availableGroups = [];

  const disableDiscardButton = useMemo(() => {
    return isEqual(selectedFilters, appliedFilters);
  }, [selectedFilters, appliedFilters]);

  const setSecondaryFiltersList = useCallback((secondaryFilters) => {
    setSelectedFilters((curr) => {
      return {
        ...curr,
        secondaryFilters
      };
    });
  }, []);

  const applyFilters = useCallback(() => {
    setAppliedFilters(selectedFilters);
    setFiltersExpanded(false);
    setFiltersDirty(true);
    const reqPayload = getFiltersRequestPayload({
      selectedFilters,
      table_props: displayTableProps,
      caller: 'user_profiles'
    });
    getProfileUsers(activeProject.id, reqPayload);
  }, [
    selectedFilters,
    displayTableProps,
    getProfileUsers,
    activeProject.id,
    setFiltersDirty
  ]);

  const setEventTimeline = useCallback((eventTimeline) => {
    setSelectedFilters((curr) => {
      return {
        ...curr,
        eventTimeline
      };
    });
  }, []);

  const renderPropertyFilter = () => (
    <PropertyFilter
      profileType='user'
      source={timelinePayload.source}
      filters={timelinePayload.filters}
      secondaryFiltersList={selectedFilters.secondaryFilters}
      filtersExpanded={filtersExpanded}
      filtersList={selectedFilters.filters}
      appliedFilters={appliedFilters}
      selectedAccount={selectedAccount}
      listEvents={selectedFilters.eventsList}
      availableGroups={availableGroups}
      eventProp={selectedFilters.eventProp}
      eventTimeline={selectedFilters.eventTimeline}
      areFiltersDirty={areFiltersDirty}
      disableDiscardButton={disableDiscardButton}
      isActiveSegment={Boolean(timelinePayload.segment_id) === true}
      applyFilters={applyFilters}
      setFiltersExpanded={setFiltersExpanded}
      setSaveSegmentModal={handleSaveSegmentClick}
      setFiltersList={setFiltersList}
      setAppliedFilters={setAppliedFilters}
      setListEvents={setListEvents}
      setEventProp={setEventProp}
      resetSelectedFilters={resetSelectedFilters}
      onClearFilters={handleClearFilters}
      setSelectedAccount={setSelectedAccount}
      setSecondaryFiltersList={setSecondaryFiltersList}
      setEventTimeline={setEventTimeline}
    />
  );

  const { saveButtonDisabled } = useMemo(() => {
    return checkFiltersEquality({
      appliedFilters,
      newSegmentMode,
      filtersList: selectedFilters.filters,
      secondaryFiltersList: selectedFilters.secondaryFilters,
      eventProp: selectedFilters.eventProp,
      eventsList: selectedFilters.eventsList,
      isActiveSegment: Boolean(timelinePayload.segment_id),
      areFiltersDirty
    });
  }, [
    timelinePayload.segment_id,
    appliedFilters,
    areFiltersDirty,
    newSegmentMode,
    selectedFilters.eventProp,
    selectedFilters.eventsList,
    selectedFilters.filters,
    selectedFilters.secondaryFilters
  ]);

  const handleSaveSegmentClick = useCallback(() => {
    if (newSegmentMode === true) {
      setSaveSegmentModal(true);
      return;
    }
    if (Boolean(timelinePayload.segment_id) === true) {
      setUpdateSegmentModal(true);
    } else {
      setSaveSegmentModal(true);
    }
  }, [timelinePayload.segment_id, newSegmentMode]);

  const renderSaveSegmentButton = () => {
    return (
      <ControlledComponent
        controller={
          filtersExpanded === false &&
          saveButtonDisabled === false &&
          newSegmentMode === false
        }
      >
        <Button
          onClick={handleSaveSegmentClick}
          type='default'
          className='flex items-center col-gap-1'
          disabled={saveButtonDisabled}
        >
          <SVG
            color={saveButtonDisabled ? '#BFBFBF' : '#1890ff'}
            size={16}
            name='pieChart'
          />
          <Text
            type='title'
            extraClass='mb-0'
            color={saveButtonDisabled ? 'disabled' : 'brand-color-6'}
          >
            Save as Segment
          </Text>
        </Button>
      </ControlledComponent>
    );
  };

  useEffect(() => {
    fetchUserPropertyValues(activeProject.id, '$user_id')
      .then((res) => {
        setUserValueOpts({ ...res.data });
      })
      .catch((err) => {
        console.log(err);
        setUserValueOpts({});
      });
  }, [activeProject.id]);

  const onApplyClick = (values) => {
    const updatedPayload = {
      ...timelinePayload,
      search_filter: values.map((value) => JSON.parse(value)[0])
    };
    setListSearchItems(updatedPayload.search_filter);
    setTimelinePayload(updatedPayload);
    setActiveSegment(activeSegment);
    getUsers(updatedPayload);
  };

  const searchUsers = () => (
    <div className='absolute top-0'>
      {searchDDOpen ? (
        <FaSelect
          multiSelect
          options={
            userValueOpts
              ? Object.keys(userValueOpts).map((value) => [value])
              : []
          }
          displayNames={userValueOpts}
          applClick={(val) => onApplyClick(val)}
          onClickOutside={() => setSearchDDOpen(false)}
          selectedOpts={listSearchItems}
          allowSearch
          placeholder='Search Users'
          style={{
            top: '-8px',
            right: 0,
            padding: '8px 8px 12px',
            overflowX: 'hidden'
          }}
          posRight
        />
      ) : null}
    </div>
  );

  const onSearchClose = () => {
    setSearchBarOpen(false);
    setSearchDDOpen(false);
    if (timelinePayload?.search_filter?.length) {
      const payload = { ...timelinePayload };
      payload.search_filter = [];
      setListSearchItems([]);
      setTimelinePayload(payload);
      setActiveSegment(activeSegment);
      getUsers(payload);
    }
  };

  const onSearchOpen = () => {
    setSearchBarOpen(true);
    setSearchDDOpen(true);
  };

  const renderSearchSection = () => (
    <ControlledComponent
      controller={filtersExpanded === false && newSegmentMode === false}
    >
      <div className='relative'>
        {searchBarOpen ? (
          <div className={'flex items-center justify-between'}>
            {!searchDDOpen && (
              <Input
                size='large'
                value={listSearchItems ? listSearchItems.join(', ') : null}
                placeholder={'Search Users'}
                style={{ width: '240px', 'border-radius': '5px' }}
                prefix={<SVG name='search' size={20} color={'grey'} />}
                onClick={() => setSearchDDOpen(true)}
              />
            )}
            <Button type='text' className='search-btn' onClick={onSearchClose}>
              <SVG name={'close'} size={20} color={'grey'} />
            </Button>
          </div>
        ) : (
          <Button type='text' className='search-btn' onClick={onSearchOpen}>
            <SVG name={'search'} size={20} color={'grey'} />
          </Button>
        )}
        {searchUsers()}
      </div>
    </ControlledComponent>
  );

  const renderTablePropsSelect = () => {
    return (
      <Popover
        overlayClassName='fa-activity--filter'
        placement='bottomLeft'
        visible={showPopOver}
        onVisibleChange={(visible) => {
          setShowPopOver(visible);
        }}
        onClick={() => {
          setShowPopOver(true);
        }}
        trigger='click'
        content={popoverContent}
      >
        <Button type='text'>
          <SVG size={20} name={'tableColumns'} />
        </Button>
      </Popover>
    );
  };

  const navigateToAccountsEngagement = useCallback(() => {
    history.push(PathUrls.ConfigureEngagements);
  }, []);

  const moreActionsContent = () => {
    const accountEngagement = (
      <div
        role='button'
        onClick={navigateToAccountsEngagement}
        className='flex cursor-pointer col-gap-4 items-center py-2 px-4 hover:bg-gray-100'
      >
        <SVG size={20} name='fireFlameCurved' color='#8c8c8c' />
        <Text type='title' color='character-primary' extraClass='mb-0'>
          Account engagement rules
        </Text>
      </div>
    );

    if (Boolean(timelinePayload.segment_id) === false) {
      return accountEngagement;
    }
    return (
      <div className='flex flex-col'>
        <div className='flex flex-col'>
          <div
            role='button'
            onClick={() => {
              setShowSegmentActions(false);
              setMoreActionsModalMode(moreActionsMode.RENAME);
            }}
            className='flex cursor-pointer hover:bg-gray-100 col-gap-4 items-center py-2 px-4'
          >
            <SVG size={20} name='edit_query' color='#8c8c8c' />
            <Text type='title' color='character-primary' extraClass='mb-0'>
              Rename Segment
            </Text>
          </div>
          <div
            role='button'
            onClick={() => {
              setShowSegmentActions(false);
              setMoreActionsModalMode(moreActionsMode.DELETE);
            }}
            className='flex cursor-pointer hover:bg-gray-100 col-gap-4 border-b items-center py-2 px-4'
          >
            <SVG size={20} name='trash' color='#8c8c8c' />
            <Text type='title' color='character-primary' extraClass='mb-0'>
              Delete Segment
            </Text>
          </div>
        </div>
        {accountEngagement}
      </div>
    );
  };

  const renderMoreActions = () => {
    return (
      <Popover
        placement='bottomLeft'
        visible={showSegmentActions}
        onVisibleChange={(visible) => {
          setShowSegmentActions(visible);
        }}
        onClick={() => {
          setShowSegmentActions(true);
        }}
        trigger='click'
        content={moreActionsContent}
        overlayClassName={cx(
          'fa-activity--filter',
          styles['more-actions-popover']
        )}
      >
        <Button className={styles['more-actions-button']} type='default'>
          <SVG size={20} name={'more'} />
        </Button>
      </Popover>
    );
  };

  const handleTableChange = (pageParams, somedata, sorter) => {
    setCurrentPage(pageParams.current);
    setCurrentPageSize(pageParams.pageSize);
    setDefaultSorterInfo({ key: sorter.columnKey, order: sorter.order });
  };

  const renderTable = () => (
    <div>
      <Table
        size='large'
        onRow={(user) => ({
          onClick: () => {
            history.push(
              `/profiles/people/${btoa(user.identity.id)}?is_anonymous=${
                user.identity.isAnonymous
              }`,
              {
                timelinePayload: timelinePayload,
                activeSegment: activeSegment,
                fromDetails: true,
                currentPage: currentPage,
                currentPageSize: currentPageSize,
                activeSorter: defaultSorterInfo,
                appliedFilters: areFiltersDirty ? appliedFilters : null
              }
            );
          }
        })}
        className='fa-table--userlist'
        dataSource={getTableData(contacts.data)}
        columns={tableColumns}
        rowClassName='cursor-pointer'
        pagination={{
          position: ['bottom', 'left'],
          defaultPageSize: '25',
          current: currentPage,
          pageSize: currentPageSize
        }}
        onChange={handleTableChange}
        scroll={{
          x: tableProperties?.length * 250
        }}
      />
      <div className='flex flex-row-reverse mt-4'></div>
    </div>
  );

  const showRangeNudge = useMemo(() => {
    return showUpgradeNudge(
      sixSignalInfo?.usage || 0,
      sixSignalInfo?.limit || 0,
      currentProjectSettings
    );
  }, [sixSignalInfo?.usage, sixSignalInfo?.limit, currentProjectSettings]);

  const titleIcon = useMemo(() => {
    if (Boolean(timelinePayload.segment_id) === true) {
      return 'pieChart';
    }
    return ProfilesSidebarIconsMapping[timelinePayload.source] != null
      ? ProfilesSidebarIconsMapping[timelinePayload.source]
      : 'userGroup';
  }, [timelinePayload]);

  const titleIconColor = useMemo(() => {
    return getSegmentColorCode(activeSegment?.name ?? '');
  }, [activeSegment?.name]);

  const pageTitle = useMemo(() => {
    if (newSegmentMode === true) {
      return 'Untitled Segment 1';
    }
    if (Boolean(timelinePayload.segment_id) === false) {
      const source = timelinePayload.source;
      const title = get(
        userOptions.find((elem) => elem[1] === source),
        0,
        'All People'
      );
      return title;
    }
    return activeSegment.name;
  }, [timelinePayload, userOptions, activeSegment, newSegmentMode]);

  if (loading) {
    return (
      <div className='flex justify-center items-center w-full h-64'>
        <Spin size='large' />
      </div>
    );
  }

  if (isIntegrationEnabled) {
    return (
      <ProfilesWrapper>
        <ControlledComponent controller={showRangeNudge === true}>
          <div className='mb-4'>
            <RangeNudge
              title='Users Identified'
              amountUsed={sixSignalInfo?.usage || 0}
              totalLimit={sixSignalInfo?.limit || 0}
            />
          </div>
        </ControlledComponent>

        <div className='flex justify-between items-center'>
          <div className='flex col-gap-2  items-center'>
            <div className='flex items-center rounded justify-center h-10 w-10'>
              <SVG name={titleIcon} size={32} color={titleIconColor} />
            </div>
            <Text
              type='title'
              level={3}
              weight='bold'
              extraClass='mb-0'
              id={'fa-at-text--page-title'}
            >
              {pageTitle}
            </Text>
          </div>
        </div>

        <div className='flex justify-between items-center my-4'>
          <div className='flex items-center col-gap-2 w-full'>
            {renderPropertyFilter()}
            {renderSaveSegmentButton()}
          </div>
          <div className='inline-flex col-gap-2'>
            <ControlledComponent
              controller={filtersExpanded === false && newSegmentMode === false}
            >
              {renderSearchSection()}
              {renderTablePropsSelect()}
              {renderMoreActions()}
            </ControlledComponent>
          </div>
        </div>

        <ControlledComponent controller={contacts.isLoading}>
          <Spin size='large' className='fa-page-loader' />
        </ControlledComponent>
        <ControlledComponent
          controller={
            contacts.isLoading === false &&
            contacts.data.length > 0 &&
            (newSegmentMode === false || areFiltersDirty === true)
          }
        >
          <>{renderTable()}</>
        </ControlledComponent>
        <ControlledComponent
          controller={
            contacts.isLoading === false &&
            contacts.data.length === 0 &&
            (newSegmentMode === false || areFiltersDirty === true)
          }
        >
          <NoDataWithMessage message={'No Profiles Found'} />
        </ControlledComponent>

        <UpgradeModal
          visible={isUpgradeModalVisible}
          variant='account'
          onCancel={() => setIsUpgradeModalVisible(false)}
        />
        <SaveSegmentModal
          visible={saveSegmentModal}
          handleCancel={() => setSaveSegmentModal(false)}
          handleSubmit={handleCreateSegment}
          isLoading={false}
        />
        <DeleteSegmentModal
          segmentName={activeSegment?.name}
          visible={moreActionsModalMode === moreActionsMode.DELETE}
          onCancel={() => setMoreActionsModalMode(null)}
          onOk={handleDeleteActiveSegment}
        />
        <RenameSegmentModal
          segmentName={activeSegment?.name}
          visible={moreActionsModalMode === moreActionsMode.RENAME}
          onCancel={() => setMoreActionsModalMode(null)}
          handleSubmit={handleRenameSegment}
        />

        <UpdateSegmentModal
          segmentName={activeSegment?.name}
          visible={updateSegmentModal}
          onCancel={() => setUpdateSegmentModal(false)}
          onCreate={handleCreateSegment}
          onUpdate={handleUpdateSegmentDefinition}
        />
      </ProfilesWrapper>
    );
  }

  if (errMsg !== '' && isIntegrationEnabled) {
    return <NoDataWithMessage message={errMsg} />;
  }

  return isOnboarded(currentProjectSettings) ? (
    <CommonBeforeIntegrationPage />
  ) : (
    <NoDataWithMessage message={'Onboarding Not Completed'} />
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  contacts: state.timelines.contacts,
  segments: state.timelines.segments,
  currentProjectSettings: state.global.currentProjectSettings
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      createNewSegment,
      getProfileUsers,
      getProfileUserDetails,
      getSavedSegments,
      getUserPropertiesV2,
      fetchProjectSettingsV1,
      fetchProjectSettings,
      fetchMarketoIntegration,
      fetchBingAdsIntegration,
      udpateProjectSettings,
      updateSegmentForId,
      deleteSegment
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(UserProfiles);
