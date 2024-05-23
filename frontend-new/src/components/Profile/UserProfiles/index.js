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
  Input,
  Form,
  Tooltip
} from 'antd';
import { connect, useDispatch, useSelector } from 'react-redux';
import { bindActionCreators } from 'redux';
import _, { isEqual } from 'lodash';
import SearchCheckList from 'Components/SearchCheckList';
import { formatUserPropertiesToCheckList } from 'Reducers/timelines/utils';
import {
  PropTextFormat,
  convertGroupedPropertiesToUngrouped
} from 'Utils/dataFormatter';
import {
  setTimelinePayloadAction,
  setFiltersDirtyAction,
  setNewSegmentModeAction
} from 'Reducers/userProfilesView/actions';
import { useHistory, useLocation } from 'react-router-dom';
import RangeNudge from 'Components/GenericComponents/RangeNudge';
import { showUpgradeNudge } from 'Views/Settings/ProjectSettings/Pricing/utils';
import CommonBeforeIntegrationPage from 'Components/GenericComponents/CommonBeforeIntegrationPage';
import ControlledComponent from 'Components/ControlledComponent/ControlledComponent';
import { ProfilesSidebarIconsMapping } from 'Views/AppSidebar/appSidebar.constants';
import { isOnboarded } from 'Utils/global';
import { getSegmentColorCode } from 'Views/AppSidebar/appSidebar.helpers';
import truncateURL from 'Utils/truncateURL';
import { ACCOUNTS_TABLE_COLUMN_TYPES, COLUMN_TYPE_PROPS } from 'Utils/table';
import ResizableTitle from 'Components/Resizable';
import logger from 'Utils/logger';
import useAutoFocus from 'hooks/useAutoFocus';
import {
  updateTableProperties,
  updateTablePropertiesForSegment
} from 'Reducers/timelines';
import { Text, SVG } from '../../factorsComponents';
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
  formatFiltersForPayload,
  getPropType,
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
import ProfilesWrapper from '../ProfilesWrapper';
import { getUserOptions } from './userProfiles.helpers';
import UpgradeModal from '../UpgradeModal';
import {
  INITIAL_USER_PROFILES_FILTERS_STATE,
  moreActionsMode
} from '../AccountProfiles/accountProfiles.constants';
import { checkFiltersEquality } from '../AccountProfiles/accountProfiles.helpers';
import SaveSegmentModal from '../AccountProfiles/SaveSegmentModal';
import DeleteSegmentModal from '../AccountProfiles/DeleteSegmentModal';
import RenameSegmentModal from '../AccountProfiles/RenameSegmentModal';
import UpdateSegmentModal from '../AccountProfiles/UpdateSegmentModal';
import styles from './index.module.scss';
import {
  ALPHANUMSTR,
  PROFILE_TYPE_USER,
  headerClassStr,
  iconColors
} from '../constants';

const userOptions = getUserOptions();

function UserProfiles({
  createNewSegment,
  getSavedSegments,
  getProfileUsers,
  getUserPropertiesV2,
  fetchProjectSettingsV1,
  fetchProjectSettings,
  fetchMarketoIntegration,
  fetchBingAdsIntegration,
  updateSegmentForId,
  deleteSegment
}) {
  const dispatch = useDispatch();
  const history = useHistory();
  const location = useLocation();

  const [listSearchItems, setListSearchItems] = useState([]);
  const [searchBarOpen, setSearchBarOpen] = useState(false);
  const [loading, setLoading] = useState(true);
  const [checkListUserProps, setCheckListUserProps] = useState([]);
  const [showPopOver, setShowPopOver] = useState(false);
  const [isUpgradeModalVisible, setIsUpgradeModalVisible] = useState(false);
  const [errMsg, setErrMsg] = useState('');

  const [currentPage, setCurrentPage] = useState(1);
  const [currentPageSize, setCurrentPageSize] = useState(25);
  const [defaultSorterInfo, setDefaultSorterInfo] = useState({});

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
  const [moreActionsModalMode, setMoreActionsModalMode] = useState(null);
  const [peopleRow, setPeopleRow] = useState(null);

  const { contacts } = useSelector((state) => state.timelines);

  const {
    bingAds,
    marketo,
    active_project: activeProject,
    currentProjectSettings,
    currentProjectSettings: integration,
    projectSettingsV1: integrationV1,
    projectDomainsList
  } = useSelector((state) => state.global);
  const { dashboards } = useSelector((state) => state.dashboard);
  const { userPropertiesV2, userPropNames } = useSelector(
    (state) => state.coreQuery
  );

  const {
    timelinePayload,
    newSegmentMode,
    filtersDirty: areFiltersDirty
  } = useSelector((state) => state.userProfilesView);

  const { sixSignalInfo } = useSelector((state) => state.featureConfig);

  const setFiltersDirty = useCallback(
    (value) => {
      dispatch(setFiltersDirtyAction(value));
    },
    [dispatch]
  );

  const setTimelinePayload = useCallback((payload) => {
    dispatch(setTimelinePayloadAction(payload));
  }, []);

  const displayTableProps = useMemo(() => {
    const tableProps = timelinePayload?.segment?.id
      ? timelinePayload?.segment?.query?.table_props
      : currentProjectSettings?.timelines_config?.user_config?.table_props;
    return (
      tableProps?.filter((entry) => entry !== '' && entry !== undefined) || []
    );
  }, [currentProjectSettings, timelinePayload]);

  const { tableProperties, tableColumns } = useMemo(() => {
    const columns = [
      {
        title: <div className={headerClassStr}>Identity</div>,
        width: COLUMN_TYPE_PROPS.string.min,
        dataIndex: 'identity',
        key: 'identity',
        fixed: 'left',
        ellipsis: true,
        sorter: (a, b) => sortStringColumn(a.identity.id, b.identity.id),
        render: (identity) => (
          <div className='flex items-center' id={identity.id}>
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

    const tableProps = timelinePayload?.segment?.id
      ? timelinePayload?.segment?.query?.table_props
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
              extraClass='m-0 truncate capitalize'
            >
              {propDisplayName}
            </Text>
          ),
          dataIndex: prop,
          key: prop,
          width:
            COLUMN_TYPE_PROPS[
              ACCOUNTS_TABLE_COLUMN_TYPES[prop]?.Type || 'string'
            ]?.min || 264,
          showSorterTooltip: null,
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
      width: COLUMN_TYPE_PROPS.actions.min,
      align: 'left',
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
        }
      });
    }
    return { tableProperties: tableProps, tableColumns: columns };
  }, [
    contacts?.data,
    currentProjectSettings,
    timelinePayload,
    defaultSorterInfo,
    projectDomainsList
  ]);

  const disableDiscardButton = useMemo(
    () => isEqual(selectedFilters, appliedFilters),
    [selectedFilters, appliedFilters]
  );

  const { saveButtonDisabled } = useMemo(
    () =>
      checkFiltersEquality({
        appliedFilters,
        selectedFilters,
        newSegmentMode,
        isActiveSegment: Boolean(timelinePayload.segment.id),
        areFiltersDirty
      }),
    [
      timelinePayload.segment,
      appliedFilters,
      areFiltersDirty,
      newSegmentMode,
      selectedFilters
    ]
  );

  const showRangeNudge = useMemo(
    () =>
      showUpgradeNudge(
        sixSignalInfo?.usage || 0,
        sixSignalInfo?.limit || 0,
        currentProjectSettings
      ),
    [sixSignalInfo?.usage, sixSignalInfo?.limit, currentProjectSettings]
  );

  const titleIcon = useMemo(() => {
    if (Boolean(timelinePayload.segment.id) === true) {
      return 'pieChart';
    }
    return ProfilesSidebarIconsMapping[timelinePayload.source] != null
      ? ProfilesSidebarIconsMapping[timelinePayload.source]
      : 'userGroup';
  }, [timelinePayload]);

  const titleIconColor = useMemo(
    () => getSegmentColorCode(timelinePayload?.segment?.name ?? ''),
    [timelinePayload?.segment]
  );

  const pageTitle = useMemo(() => {
    if (newSegmentMode === true) {
      return 'Untitled Segment 1';
    }
    if (Boolean(timelinePayload.segment.id) === false) {
      const { source } = timelinePayload;
      const title = get(
        userOptions.find((elem) => elem[1] === source),
        0,
        'All People'
      );
      return title;
    }
    return timelinePayload?.segment?.name;
  }, [timelinePayload, userOptions, newSegmentMode]);

  const restoreFiltersDefaultState = useCallback(
    (
      isClearFilter = false,
      selectedAccount = INITIAL_USER_PROFILES_FILTERS_STATE.account
    ) => {
      const initialFiltersStateWithSelectedAccount = {
        ...INITIAL_USER_PROFILES_FILTERS_STATE,
        account: selectedAccount
      };
      setSelectedFilters(initialFiltersStateWithSelectedAccount);
      setAppliedFilters(initialFiltersStateWithSelectedAccount);
      if (!isClearFilter) setFiltersExpanded(false);
      setFiltersDirty(false);
    },
    [setFiltersDirty]
  );

  const handleClearFilters = useCallback(() => {
    restoreFiltersDefaultState(true);
  }, []);

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
          await getSavedSegments(activeProject.id);
          setTimelinePayload({
            source: 'All',
            segment: response.payload.segment
          });
          setSaveSegmentModal(false);
          setUpdateSegmentModal(false);
          setFiltersDirty(false);
        }
        dispatch(setNewSegmentModeAction(false));
      } catch (err) {
        notification.error({
          message: 'Error',
          description:
            err?.data?.error || 'Segment Creation Failed. Invalid Parameters.',
          duration: 3
        });
      }
    },
    [activeProject.id]
  );

  const handleCreateSegment = useCallback(
    (newSegmentName) => {
      const reqPayload = getFiltersRequestPayload({
        selectedFilters,
        tableProps: displayTableProps,
        caller: 'user_profiles'
      });
      reqPayload.name = newSegmentName;
      reqPayload.type = 'All';
      handleSaveSegment(reqPayload);
    },
    [selectedFilters]
  );

  const handleDeleteActiveSegment = useCallback(() => {
    deleteSegment({
      projectId: activeProject.id,
      segmentId: timelinePayload?.segment?.id
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
            segment: {}
          })
        );
      });
  }, [timelinePayload.segment, activeProject.id, deleteSegment]);

  const handleRenameSegment = useCallback(
    async (name) => {
      if (!timelinePayload.segment) return;

      try {
        const segmentId = timelinePayload.segment.id;

        await updateSegmentForId(activeProject.id, segmentId, { name });
        getSavedSegments(activeProject.id);

        setMoreActionsModalMode(null);
        notification.success({
          message: 'Segment renamed successfully',
          duration: 5
        });

        const updatedPayload = {
          ...timelinePayload,
          segment: { ...timelinePayload.segment, name }
        };
        setTimelinePayload(updatedPayload);
      } catch (error) {
        logger.error(error);
      }
    },
    [activeProject.id, timelinePayload.segment]
  );

  const handleUpdateSegmentDefinition = useCallback(() => {
    const reqPayload = getFiltersRequestPayload({
      selectedFilters,
      tableProps: displayTableProps,
      caller: 'user_profiles'
    });
    updateSegmentForId(
      activeProject.id,
      timelinePayload.segment.id,
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
    timelinePayload.segment,
    getSavedSegments,
    setFiltersDirty
  ]);
  useEffect(() => {
    dispatch(setNewSegmentModeAction(false));
  }, []);
  useEffect(() => {
    if (!timelinePayload.search_filter) {
      setListSearchItems([]);
    } else {
      const listValues = timelinePayload?.search_filter || [];
      setListSearchItems(_.uniq(listValues));
    }
  }, [timelinePayload?.search_filter]);

  useEffect(() => {
    setTimeout(() => {
      setLoading(false);
    }, 1000);

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
    const tableProps = timelinePayload?.segment?.id
      ? timelinePayload?.segment?.query?.table_props
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
  }, [currentProjectSettings, userPropertiesV2, timelinePayload]);

  useEffect(() => {
    getSavedSegments(activeProject.id);
  }, [activeProject.id]);

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
        const reqPayload = formatReqPayload(formatPayload);
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
      activeProject.id,
      history
    ]
  );

  useEffect(() => {
    getUsers(timelinePayload);
  }, [timelinePayload]);

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
    } else if (newSegmentMode === false) {
      // its already opened segment / All People / other sources
      if (timelinePayload?.segment?.query != null) {
        const filters = getSelectedFiltersFromQuery({
          query: timelinePayload?.segment?.query,
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
  }, [timelinePayload, newSegmentMode]);

  const handlePropChange = (option) => {
    if (
      option.enabled ||
      checkListUserProps.filter((item) => item.enabled === true).length < 12
    ) {
      setCheckListUserProps((prev) => {
        const checkListProps = [...prev];
        const optIndex = checkListProps.findIndex(
          (obj) => obj.prop_name === option.prop_name
        );
        checkListProps[optIndex].enabled = !checkListProps[optIndex].enabled;
        checkListProps.sort((a, b) => (b?.enabled || 0) - (a?.enabled || 0));
        return checkListProps;
      });
    } else {
      notification.error({
        message: 'Error',
        description: 'Maximum of 12 Table Properties Selection Allowed.',
        duration: 2
      });
    }
  };

  const applyTableProps = () => {
    const newTableProps =
      checkListUserProps
        ?.filter((item) => item.enabled === true)
        ?.map((item) => item?.prop_name)
        ?.filter(
          (entry) => entry !== '' && entry !== undefined && entry !== null
        ) || [];
    if (timelinePayload?.segment?.id?.length) {
      updateTablePropertiesForSegment(
        activeProject.id,
        timelinePayload.segment.id,
        newTableProps
      )
        .then(() => getSavedSegments(activeProject.id))
        .finally(() => getUsers(timelinePayload));
    } else {
      updateTableProperties(
        activeProject.id,
        PROFILE_TYPE_USER,
        newTableProps
      ).then(() => getUsers(timelinePayload));
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

  const resetSelectedFilters = useCallback(() => {
    setSelectedFilters(appliedFilters);
  }, [appliedFilters]);

  const applyFilters = useCallback(() => {
    setAppliedFilters(selectedFilters);
    setFiltersExpanded(false);
    setFiltersDirty(true);

    const reqPayload = getFiltersRequestPayload({
      selectedFilters,
      tableProps: displayTableProps,
      caller: 'user_profiles'
    });
    reqPayload.search_filter =
      (listSearchItems && listSearchItems.length > 0 && listSearchItems) || [];

    getProfileUsers(activeProject.id, reqPayload);
  }, [
    selectedFilters,
    displayTableProps,
    getProfileUsers,
    activeProject.id,
    setFiltersDirty
  ]);

  const handleSaveSegmentClick = useCallback(() => {
    if (newSegmentMode === true) {
      setSaveSegmentModal(true);
      return;
    }
    if (Boolean(timelinePayload.segment.id) === true) {
      setUpdateSegmentModal(true);
    } else {
      setSaveSegmentModal(true);
    }
  }, [timelinePayload.segment, newSegmentMode]);

  const renderPropertyFilter = () => (
    <PropertyFilter
      profileType='user'
      filtersExpanded={filtersExpanded}
      setFiltersExpanded={setFiltersExpanded}
      selectedFilters={selectedFilters}
      setSelectedFilters={setSelectedFilters}
      resetSelectedFilters={resetSelectedFilters}
      appliedFilters={appliedFilters}
      applyFilters={applyFilters}
      areFiltersDirty={areFiltersDirty}
      disableDiscardButton={disableDiscardButton}
      isActiveSegment={Boolean(timelinePayload?.segment?.id)}
      setSaveSegmentModal={handleSaveSegmentClick}
      onClearFilters={handleClearFilters}
    />
  );

  const renderSaveSegmentButton = () => (
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
        className='flex items-center gap-x-1'
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

  const handleUsersSearch = (values) => {
    if (
      (listSearchItems.length >= 1 && listSearchItems[0] === values?.users) ||
      (listSearchItems.length === 0 && !values?.users)
    ) {
      return;
    }
    if (values?.users) {
      values = [JSON.stringify([values.users])];
    } else {
      values = [];
    }
    const updatedPayload = {
      ...timelinePayload,
      search_filter: values.map((value) => JSON.parse(value)[0])
    };
    setListSearchItems(updatedPayload.search_filter);
    setTimelinePayload(updatedPayload);
    getUsers(updatedPayload);
  };

  const onSearchClose = () => {
    setSearchBarOpen(false);
    handleUsersSearch({ users: '' });
  };

  const onSearchOpen = () => {
    setSearchBarOpen(true);
  };

  function SearchBar() {
    const searchBarRef = useAutoFocus();
    return (
      <div className='flex items-center justify-between'>
        <Form
          name='basic'
          labelCol={{ span: 8 }}
          wrapperCol={{ span: 16 }}
          onFinish={handleUsersSearch}
          autoComplete='off'
        >
          <Form.Item name='users'>
            <Input
              ref={searchBarRef}
              size='large'
              defaultValue={listSearchItems ? listSearchItems.join(', ') : null}
              placeholder='Search Users'
              style={{ width: '240px', 'border-radius': '5px' }}
              prefix={<SVG name='search' size={24} color='#8c8c8c' />}
            />
          </Form.Item>
        </Form>
        <Button type='text' className='search-btn' onClick={onSearchClose}>
          <SVG name='close' size={24} color='#8c8c8c' />
        </Button>
      </div>
    );
  }

  const renderSearchSection = () => (
    <div className='relative'>
      <ControlledComponent controller={searchBarOpen}>
        <SearchBar />
      </ControlledComponent>
      <ControlledComponent controller={!searchBarOpen}>
        <Tooltip title='Search'>
          <Button type='text' className='search-btn' onClick={onSearchOpen}>
            <SVG name='search' size={24} color='#8c8c8c' />
          </Button>
        </Tooltip>
      </ControlledComponent>
    </div>
  );

  const renderTablePropsSelect = () => (
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
      <Tooltip title='Edit columns'>
        <Button className='search-btn' type='text'>
          <SVG size={24} color='#8c8c8c' name='tableColumns' />
        </Button>
      </Tooltip>
    </Popover>
  );

  const moreActionsContent = () => (
    <div className='flex flex-col'>
      <div className='flex flex-col'>
        <div
          role='button'
          tabIndex={-1}
          onClick={() => {
            setShowSegmentActions(false);
            setMoreActionsModalMode(moreActionsMode.RENAME);
          }}
          className='flex cursor-pointer hover:bg-gray-100 gap-x-4 items-center py-2 px-4'
        >
          <SVG size={20} name='edit_query' color='#8c8c8c' />
          <Text type='title' color='character-primary' extraClass='mb-0'>
            Rename Segment
          </Text>
        </div>
        <div
          role='button'
          tabIndex={-2}
          onClick={() => {
            setShowSegmentActions(false);
            setMoreActionsModalMode(moreActionsMode.DELETE);
          }}
          className='flex cursor-pointer hover:bg-gray-100 gap-x-4 items-center py-2 px-4'
        >
          <SVG size={20} name='trash' color='#8c8c8c' />
          <Text type='title' color='character-primary' extraClass='mb-0'>
            Delete Segment
          </Text>
        </div>
      </div>
    </div>
  );

  const renderMoreActions = () => (
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
      <Button className='search-btn' type='text'>
        <SVG size={24} color='#8c8c8c' name='more' />
      </Button>
    </Popover>
  );

  const handleTableChange = (pageParams, somedata, sorter) => {
    setCurrentPage(pageParams.current);
    setCurrentPageSize(pageParams.pageSize);
    setDefaultSorterInfo({ key: sorter.columnKey, order: sorter.order });
  };

  useEffect(() => {
    // This is the name of Account which was opened recently
    const from = location.state?.peoplesTableRow;
    const tableElement = peopleRow?.querySelector('.ant-table-body');

    if (tableElement && from && document.getElementById(from)) {
      const element = document.getElementById(from);
      const y =
        element.getBoundingClientRect().y -
        tableElement.getBoundingClientRect().y -
        15;

      tableElement.scrollTo({ top: y, behavior: 'smooth' });

      location.state.peoplesTableRow = '';
    }
  }, [contacts, tableColumns, location.state, peopleRow]);

  const renderTable = () => {
    const mergeColumns = tableColumns.map((col) => ({
      ...col,
      onHeaderCell: (column) => ({
        width: column.width
      })
    }));
    return (
      <div id='resizing-table-container-div'>
        <Table
          ref={(e) => {
            if (e) setPeopleRow(e);
          }}
          size='large'
          components={{
            header: {
              cell: ResizableTitle
            }
          }}
          onRow={(user) => ({
            onClick: () => {
              history.push(
                `/profiles/people/${btoa(user.identity.id)}?is_anonymous=${
                  user.identity.isAnonymous
                }`,
                {
                  timelinePayload,
                  fromDetails: true,
                  currentPage,
                  currentPageSize,
                  activeSorter: defaultSorterInfo,
                  appliedFilters: areFiltersDirty ? appliedFilters : null,
                  peoplesTableRow: user.identity.id
                }
              );
            }
          })}
          className='fa-table--profileslist'
          dataSource={getTableData(contacts.data)}
          columns={mergeColumns}
          rowClassName='cursor-pointer'
          pagination={{
            position: ['bottom', 'left'],
            defaultPageSize: '25',
            current: currentPage,
            pageSize: currentPageSize
          }}
          onChange={handleTableChange}
          scroll={{
            x: (tableProperties?.length || 0) * 250
          }}
        />
        <div className='flex flex-row-reverse mt-4' />
      </div>
    );
  };

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
          <div className='flex gap-x-2  items-center'>
            <div className='flex items-center rounded justify-center h-10 w-10'>
              <SVG name={titleIcon} size={32} color={titleIconColor} />
            </div>
            <Text
              type='title'
              level={3}
              weight='bold'
              extraClass='mb-0'
              id='fa-at-text--page-title'
            >
              {pageTitle}
            </Text>
          </div>
        </div>

        <div className='flex justify-between items-center my-4'>
          <div className='flex items-center gap-x-2 w-full'>
            {renderPropertyFilter()}
            {renderSaveSegmentButton()}
          </div>
          <div className='inline-flex gap-x-2'>
            <ControlledComponent
              controller={filtersExpanded === false && newSegmentMode === false}
            >
              {renderSearchSection()}
              {renderTablePropsSelect()}
              <ControlledComponent
                controller={Boolean(timelinePayload.segment.id)}
              >
                {renderMoreActions()}
              </ControlledComponent>
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
          <NoDataWithMessage message='No Profiles Found' />
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
          segmentName={timelinePayload?.segment?.name}
          visible={moreActionsModalMode === moreActionsMode.DELETE}
          onCancel={() => setMoreActionsModalMode(null)}
          onOk={handleDeleteActiveSegment}
        />
        <RenameSegmentModal
          segmentName={timelinePayload?.segment?.name}
          visible={moreActionsModalMode === moreActionsMode.RENAME}
          onCancel={() => setMoreActionsModalMode(null)}
          handleSubmit={handleRenameSegment}
        />

        <UpdateSegmentModal
          segmentName={timelinePayload?.segment?.name}
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

  return isOnboarded(integration) ? (
    <CommonBeforeIntegrationPage />
  ) : (
    <NoDataWithMessage message='Onboarding Not Completed' />
  );
}

const mapStateToProps = (state) => ({
  contacts: state.timelines.contacts,
  segments: state.timelines.segments
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
