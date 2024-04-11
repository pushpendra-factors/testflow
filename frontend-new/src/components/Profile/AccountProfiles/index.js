import React, {
  useState,
  useEffect,
  useMemo,
  useCallback,
  useRef
} from 'react';
import isEqual from 'lodash/isEqual';
import cx from 'classnames';
import {
  Table,
  Button,
  Spin,
  Popover,
  Tabs,
  notification,
  Input,
  Form,
  Tooltip
} from 'antd';
import { connect, useDispatch, useSelector } from 'react-redux';
import { bindActionCreators } from 'redux';
import { useHistory, useLocation, useParams } from 'react-router-dom';
import SearchCheckList from 'Components/SearchCheckList';
import { formatUserPropertiesToCheckList } from 'Reducers/timelines/utils';
import {
  selectAccountPayload,
  selectActiveTab
} from 'Reducers/accountProfilesView/selectors';
import {
  setAccountPayloadAction,
  setFiltersDirtyAction,
  setNewSegmentModeAction
} from 'Reducers/accountProfilesView/actions';
import useFeatureLock from 'hooks/useFeatureLock';
import { FEATURES } from 'Constants/plans.constants';
import RangeNudge from 'Components/GenericComponents/RangeNudge';
import uniq from 'lodash/uniq';
import { showUpgradeNudge } from 'Views/Settings/ProjectSettings/Pricing/utils';
import ControlledComponent from 'Components/ControlledComponent/ControlledComponent';
import { selectGroupsList } from 'Reducers/groups/selectors';
import { fetchProfileAccounts, fetchSegmentById } from 'Reducers/timelines';
import { downloadCSV } from 'Utils/csv';
import { PathUrls } from 'Routes/pathUrls';
import { GROUP_NAME_DOMAINS } from 'Components/GlobalFilter/FilterWrapper/utils';
import { defaultSegmentIconsMapping } from 'Views/AppSidebar/appSidebar.constants';
import { isOnboarded } from 'Utils/global';
import _, { cloneDeep } from 'lodash';
import { getSegmentColorCode } from 'Views/AppSidebar/appSidebar.helpers';
import ResizableTitle from 'Components/Resizable';
import logger from 'Utils/logger';
import useAutoFocus from 'hooks/useAutoFocus';
import { invalidBreakdownPropertiesList } from 'Constants/general.constants';
import MomentTz from 'Components/MomentTz';
import styles from './index.module.scss';
import {
  getGroups,
  getGroupProperties
} from '../../../reducers/coreQuery/middleware';
import DownloadCSVModal from './DownloadCSVModal';
import UpdateSegmentModal from './UpdateSegmentModal';
import {
  INITIAL_FILTERS_STATE,
  moreActionsMode
} from './accountProfiles.constants';
import RenameSegmentModal from './RenameSegmentModal';
import DeleteSegmentModal from './DeleteSegmentModal';
import SaveSegmentModal from './SaveSegmentModal';
import UpgradeModal from '../UpgradeModal';
import { defaultSegmentsList, getColumns } from './accountProfiles.helpers';
import ProfilesWrapper from '../ProfilesWrapper';
import NoDataWithMessage from '../MyComponents/NoDataWithMessage';
import {
  fetchProjectSettings,
  udpateProjectSettings
} from '../../../reducers/global';
import {
  getProfileAccounts,
  createNewSegment,
  getSavedSegments,
  updateSegmentForId,
  deleteSegment,
  getTop100Events
} from '../../../reducers/timelines/middleware';
import {
  formatReqPayload,
  getFiltersRequestPayload,
  getSelectedFiltersFromQuery
} from '../utils';
import PropertyFilter from './PropertyFilter';
import { Text, SVG } from '../../factorsComponents';
import AccountsTabs from './AccountsTabs';
import AccountsInsights from './AccountsInsights/AccountsInsights';
import AccountDrawer from './AccountDrawer';

function AccountProfiles({
  activeProject,
  accounts,
  accountPreview,
  currentProjectSettings,
  createNewSegment,
  getSavedSegments,
  fetchProjectSettings,
  getGroups,
  udpateProjectSettings,
  getProfileAccounts,
  getGroupProperties,
  updateSegmentForId,
  deleteSegment,
  getTop100Events
}) {
  const [currentPage, setCurrentPage] = useState(1);
  const [currentPageSize, setCurrentPageSize] = useState(25);
  const [searchBarOpen, setSearchBarOpen] = useState(false);
  const [listSearchItems, setListSearchItems] = useState([]);
  const [listProperties, setListProperties] = useState([]);
  const [showPopOver, setShowPopOver] = useState(false);
  const [checkListAccountProps, setCheckListAccountProps] = useState([]);
  const [tlConfig, setTLConfig] = useState({});
  const [isUpgradeModalVisible, setIsUpgradeModalVisible] = useState(false);
  const [downloadCSVOptions, setDownloadCSVOptions] = useState([]);
  const [filtersExpanded, setFiltersExpanded] = useState(false);
  const [saveSegmentModal, setSaveSegmentModal] = useState(false);
  const [updateSegmentModal, setUpdateSegmentModal] = useState(false);
  const [selectedFilters, setSelectedFilters] = useState(INITIAL_FILTERS_STATE);
  const [appliedFilters, setAppliedFilters] = useState(INITIAL_FILTERS_STATE);
  const [moreActionsModalMode, setMoreActionsModalMode] = useState(null); // DELETE | RENAME
  const [showDownloadCSVModal, setShowDownloadCSVModal] = useState(false);
  const [csvDataLoading, setCSVDataLoading] = useState(false);
  const [defaultSorterInfo, setDefaultSorterInfo] = useState({});
  const [showSegmentActions, setShowSegmentActions] = useState(false);
  const [errMsg, setErrMsg] = useState('');
  const [newTableColumns, setNewTableColumns] = useState([]);
  const [processedDomains, setProcessedDomains] = useState(new Set());
  const [preview, setPreview] = useState({ drawerVisible: false, domain: {} });

  const onDrawerClose = () =>
    setPreview((prevState) => ({ ...prevState, drawerVisible: false }));

  useEffect(() => {
    if (filtersExpanded) onDrawerClose();
  }, [filtersExpanded]);

  const dispatch = useDispatch();
  const history = useHistory();
  const location = useLocation();
  const { segment_id: segmentID } = useParams();

  const { projectDomainsList } = useSelector((state) => state.global);
  const activeTab = useSelector((state) => selectActiveTab(state));
  const { groups, groupProperties, groupPropNames, eventNames } = useSelector(
    (state) => state.coreQuery
  );

  const { sixSignalInfo } = useSelector((state) => state.featureConfig);
  const { newSegmentMode, filtersDirty: areFiltersDirty } = useSelector(
    (state) => state.accountProfilesView
  );
  const groupsList = useSelector((state) => selectGroupsList(state));

  const accountPayload = useSelector((state) => selectAccountPayload(state));

  const { isFeatureLocked: isScoringLocked } = useFeatureLock(
    FEATURES.FEATURE_ACCOUNT_SCORING
  );

  const searchAccountsInputRef = useAutoFocus(searchBarOpen);

  const setAccountPayload = (payload) => {
    dispatch(setAccountPayloadAction(payload));
    if (payload.segment?.id) {
      dispatch(setNewSegmentModeAction(false));
    }
  };

  const getAccountPayload = async () => {
    if (segmentID && !accountPayload.isUnsaved) {
      const response = await fetchSegmentById(activeProject.id, segmentID);
      if (segmentID === accountPayload?.segment?.id) {
        return { ...accountPayload, segment: response.data };
      }
      return { source: GROUP_NAME_DOMAINS, segment: response.data };
    }
    return accountPayload;
  };

  useEffect(() => {
    dispatch(setNewSegmentModeAction(false));
  }, []);

  useEffect(() => {
    if (activeProject?.id) {
      fetchProjectSettings(activeProject.id);
      getGroups(activeProject.id);
      getSavedSegments(activeProject.id);
    }
  }, [activeProject?.id]);

  const runInit = async () => {
    const payload = await getAccountPayload();
    if (!_.isEqual(payload, accountPayload)) setAccountPayload(payload);
  };

  useEffect(() => {
    runInit();
  }, [segmentID, accountPayload]);

  useEffect(() => {
    const filteredDomainProps = (
      groupProperties[GROUP_NAME_DOMAINS] || []
    ).filter((item) => !invalidBreakdownPropertiesList.includes(item[1]));

    const groupProps = Object.keys(groups?.account_groups || {}).reduce(
      (properties, group) =>
        groupProperties[group]
          ? properties.concat(groupProperties[group])
          : properties,
      []
    );

    setListProperties([...filteredDomainProps, ...groupProps]);
  }, [groupProperties, groups]);

  useEffect(() => {
    const tableProps = accountPayload?.segment?.id
      ? accountPayload?.segment?.query?.table_props
      : currentProjectSettings.timelines_config?.account_config?.table_props ||
        [];
    const accountPropsWithEnableKey = formatUserPropertiesToCheckList(
      listProperties,
      tableProps?.filter(
        (entry) => entry !== '' && entry !== undefined && entry !== null
      )
    );
    const csvPropsWithEnableKey = formatUserPropertiesToCheckList(
      [...listProperties, ['Last Activity', 'last_activity', 'datetime']],
      tableProps?.filter(
        (entry) => entry !== '' && entry !== undefined && entry !== null
      )
    );
    setDownloadCSVOptions(csvPropsWithEnableKey);
    setCheckListAccountProps(accountPropsWithEnableKey);
  }, [currentProjectSettings, listProperties, accountPayload]);

  const getGroupPropsFromAPI = useCallback(
    async (groupId) => {
      if (!groupProperties[groupId]) {
        await getGroupProperties(activeProject.id, groupId);
      }
    },
    [activeProject.id, groupProperties]
  );

  useEffect(() => {
    getGroupPropsFromAPI(GROUP_NAME_DOMAINS);
    Object.keys(groups?.all_groups || {}).forEach((group) => {
      getGroupPropsFromAPI(group);
    });
  }, [activeProject.id, groups]);

  useEffect(() => {
    if (!currentProjectSettings?.timelines_config) return;
    setTLConfig(currentProjectSettings.timelines_config);
  }, [currentProjectSettings?.timelines_config]);

  useEffect(() => {
    if (!accountPayload.search_filter?.length) {
      setListSearchItems([]);
      setSearchBarOpen(false);
    } else {
      const listValues = accountPayload.search_filter || [];
      setListSearchItems(uniq(listValues));
      setSearchBarOpen(true);
    }
  }, [accountPayload?.search_filter]);

  const setFiltersDirty = useCallback(
    (value) => {
      dispatch(setFiltersDirtyAction(value));
    },
    [dispatch]
  );

  const restoreFiltersDefaultState = (
    isClearFilter = false,
    selectedAccount = INITIAL_FILTERS_STATE.account
  ) => {
    const initialFiltersStateWithSelectedAccount = {
      ...INITIAL_FILTERS_STATE,
      account: selectedAccount
    };
    setSelectedFilters(initialFiltersStateWithSelectedAccount);
    setAppliedFilters(cloneDeep(initialFiltersStateWithSelectedAccount));
    if (!isClearFilter) setFiltersExpanded(false);
    setFiltersDirty(false);
  };

  useEffect(() => {
    if (newSegmentMode) {
      restoreFiltersDefaultState();
    }
  }, [newSegmentMode]);

  useEffect(() => {
    if (!newSegmentMode) {
      if (accountPayload?.segment?.query != null) {
        const filters = getSelectedFiltersFromQuery({
          query: accountPayload?.segment?.query,
          groupsList
        });
        setAppliedFilters(cloneDeep(filters));
        setSelectedFilters(filters);
        setFiltersDirty(false);
        if (accountPayload?.segment?.id) setFiltersExpanded(false);
      } else {
        restoreFiltersDefaultState();
      }
    }
  }, [accountPayload, newSegmentMode]);

  const displayTableProps = useMemo(() => {
    const filterNullEntries = (entry) =>
      entry !== '' && entry !== undefined && entry !== null;

    const getFilteredTableProps = (tableProps) =>
      tableProps?.filter(filterNullEntries) || [];

    const segmentTableProps = accountPayload?.segment?.query?.table_props;
    const projectTableProps =
      currentProjectSettings?.timelines_config?.account_config?.table_props;

    const tableProps = accountPayload?.segment?.id
      ? getFilteredTableProps(segmentTableProps)
      : getFilteredTableProps(projectTableProps);

    return tableProps;
  }, [currentProjectSettings, accountPayload?.segment]);

  const handleRenameSegment = useCallback(
    (name) => {
      updateSegmentForId(activeProject.id, accountPayload?.segment?.id, {
        name
      }).then(() => {
        const updatedPayload = { ...accountPayload };
        updatedPayload.segment.name = name;
        setAccountPayload(updatedPayload);
        getSavedSegments(activeProject.id);
        setMoreActionsModalMode(null);
        notification.success({
          message: 'Segment renamed successfully',
          duration: 5
        });
      });
    },
    [activeProject.id, accountPayload?.segment]
  );

  const handleUpdateSegmentDefinition = useCallback(() => {
    const reqPayload = getFiltersRequestPayload({
      selectedFilters,
      tableProps: displayTableProps
    });
    updateSegmentForId(
      activeProject.id,
      accountPayload?.segment?.id,
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
    accountPayload?.segment
  ]);

  const handleDeleteActiveSegment = useCallback(() => {
    const segmentId = accountPayload?.segment?.id;
    if (!segmentId) {
      return;
    }

    deleteSegment({
      projectId: activeProject.id,
      segmentId
    })
      .then(() => {
        setMoreActionsModalMode(null);
        notification.success({
          message: 'Segment deleted successfully',
          duration: 5
        });
      })
      .finally(() => {
        history.replace(PathUrls.ProfileAccounts);
      });
  }, [accountPayload?.segment, activeProject.id]);

  const getAccounts = useCallback(
    async (payload) => {
      try {
        setDefaultSorterInfo({ key: '$engagement_level', order: 'descend' });
        const reqPayload = formatReqPayload(payload);
        const response = await getProfileAccounts(activeProject.id, reqPayload);

        if (response.type === 'FETCH_PROFILE_ACCOUNTS_FAILED') {
          if (response.error.status === 400) {
            setErrMsg('400 Bad Request');
          } else if (response.error.status === 500) {
            setErrMsg(
              'The server encountered an internal error and could not complete your request'
            );
          }
        } else if (response.type === 'FETCH_PROFILE_ACCOUNTS_FULFILLED') {
          if (response.status === 200 && response.payload.length === 0) {
            setErrMsg('No accounts found');
          }
        }
      } catch (err) {
        logger.error(err);
      }
    },
    [accountPayload, activeProject.id]
  );

  useEffect(() => {
    const shouldCache = location?.state?.fromDetails;
    if (shouldCache) {
      if (!location.state.accountPayload) {
        setAccountPayload({ source: GROUP_NAME_DOMAINS });
      } else {
        setCurrentPage(location.state.currentPage);
        setCurrentPageSize(location.state.currentPageSize);
        setDefaultSorterInfo(location.state.activeSorter);
        setAccountPayload(location.state.accountPayload);
        dispatch(setNewSegmentModeAction(false));
      }
      const updatedLocationState = { ...location.state, fromDetails: false };
      history.replace(location.pathname, { ...updatedLocationState });
    } else if (
      (segmentID && segmentID === accountPayload?.segment?.id) ||
      !segmentID ||
      (accountPayload.segment && !accountPayload.segment?.id)
    ) {
      getAccounts(accountPayload);
    }
  }, [segmentID, accountPayload]);

  const tableData = useMemo(() => {
    const activeID = segmentID || 'default';
    const sortedData = accounts.data[activeID]?.sort(
      (a, b) => new Date(b.last_activity) - new Date(a.last_activity)
    );
    return sortedData?.map((row) => ({
      ...row,
      ...row?.table_props
    }));
  }, [accounts, segmentID]);

  const applyFilters = useCallback(() => {
    const updatedFilters = cloneDeep(selectedFilters);
    setAppliedFilters(updatedFilters);
    setFiltersDirty(true);

    const reqPayload = getFiltersRequestPayload({
      selectedFilters,
      tableProps: displayTableProps
    });

    if (newSegmentMode) {
      setAccountPayload({
        source: GROUP_NAME_DOMAINS,
        segment: { ...reqPayload },
        isUnsaved: true
      });
    } else {
      const newPayload = { ...accountPayload };
      if (!newPayload.segment) {
        newPayload.segment = {};
      }
      newPayload.segment.query = reqPayload.query;
      newPayload.isUnsaved = true;
      setAccountPayload(newPayload);
      setFiltersExpanded(false);
    }
  }, [selectedFilters, newSegmentMode]);

  const handlePropChange = (option) => {
    if (
      option.enabled ||
      checkListAccountProps.filter((item) => item.enabled).length < 12
    ) {
      setCheckListAccountProps((prev) => {
        const checkListProps = [...prev];
        const optIndex = checkListProps.findIndex(
          (obj) => obj.prop_name === option.prop_name
        );
        checkListProps[optIndex].enabled = !checkListProps[optIndex].enabled;
        // Sorting to bubble up the selected elements onClick
        checkListProps.sort((a, b) => (b?.enabled || 0) - (a?.enabled || 0));
        return checkListProps;
      });
    } else {
      notification.error({
        message: 'Error',
        description: 'Maximum Table Properties Selection Reached.',
        duration: 2
      });
    }
  };

  const applyTableProps = async () => {
    if (accountPayload?.segment?.id?.length) {
      const newTableProps =
        checkListAccountProps
          ?.filter(({ enabled }) => enabled)
          ?.map(({ prop_name }) => prop_name)
          ?.filter((entry) => entry !== '' && entry !== undefined) || [];

      const updatedQuery = {
        ...accountPayload.segment.query,
        table_props: newTableProps
      };

      await updateSegmentForId(activeProject.id, accountPayload.segment.id, {
        query: updatedQuery
      });
      const updatedPayload = { ...accountPayload };
      updatedPayload.segment.query = updatedQuery;
      setAccountPayload({ ...updatedPayload });
      await getSavedSegments(activeProject.id);
    } else {
      const enabledProps = checkListAccountProps
        .filter(({ enabled }) => enabled)
        .map(({ prop_name }) => prop_name);

      const updatedConfig = {
        ...tlConfig,
        account_config: {
          ...tlConfig.account_config,
          table_props: enabledProps
        }
      };

      await udpateProjectSettings(activeProject.id, {
        timelines_config: updatedConfig
      });
      setAccountPayload({ ...accountPayload });
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
          mapArray={checkListAccountProps}
          sortable
          updateList={setCheckListAccountProps}
          titleKey='display_name'
          checkedKey='enabled'
          onChange={handlePropChange}
          showApply
          onApply={applyTableProps}
          disabledOptions={['Engagement']}
          showDisabledOption={isScoringLocked}
          handleDisableOptionClick={handleDisableOptionClick}
        />
      </Tabs.TabPane>
    </Tabs>
  );

  const navigateToAccountsEngagement = useCallback(() => {
    history.push(PathUrls.ConfigureEngagements);
  }, []);

  const moreActionsContent = () => {
    const accountEngagement = (
      <div
        role='button'
        tabIndex='0'
        onClick={navigateToAccountsEngagement}
        className='flex cursor-pointer gap-x-4 items-center py-2 px-4 hover:bg-gray-100'
      >
        <SVG size={20} name='fireFlameCurved' color='#8c8c8c' />
        <Text type='title' color='character-primary' extraClass='mb-0'>
          Account engagement rules
        </Text>
      </div>
    );

    if (
      !accountPayload?.segment?.id ||
      defaultSegmentsList.includes(accountPayload?.segment?.name)
    ) {
      return accountEngagement;
    }
    return (
      <div className='flex flex-col'>
        <div className='flex flex-col'>
          <div
            role='button'
            tabIndex='0'
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
            tabIndex='0'
            onClick={() => {
              setShowSegmentActions(false);
              setMoreActionsModalMode(moreActionsMode.DELETE);
            }}
            className='flex cursor-pointer hover:bg-gray-100 gap-x-4 border-b items-center py-2 px-4'
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

  const setFiltersList = useCallback((filters) => {
    setSelectedFilters((curr) => ({
      ...curr,
      filters
    }));
  }, []);

  const setSecondaryFiltersList = useCallback((secondaryFilters) => {
    setSelectedFilters((curr) => ({
      ...curr,
      secondaryFilters
    }));
  }, []);

  const setListEvents = useCallback((eventsList) => {
    setSelectedFilters((curr) => ({
      ...curr,
      eventsList
    }));
  }, []);

  const setEventProp = useCallback((eventProp) => {
    setSelectedFilters((curr) => ({
      ...curr,
      eventProp
    }));
  }, []);

  const setEventTimeline = useCallback((eventTimeline) => {
    setSelectedFilters((curr) => ({
      ...curr,
      eventTimeline
    }));
  }, []);

  const handleSaveSegmentClick = useCallback(() => {
    if (newSegmentMode) {
      setSaveSegmentModal(true);
      return;
    }
    if (accountPayload?.segment?.id) {
      setUpdateSegmentModal(true);
    } else {
      setSaveSegmentModal(true);
    }
  }, [accountPayload?.segment?.id, newSegmentMode]);

  const resetSelectedFilters = useCallback(() => {
    setSelectedFilters(appliedFilters);
  }, [appliedFilters]);

  const handleClearFilters = () => {
    restoreFiltersDefaultState(true);
  };
  const saveButtonDisabled = useMemo(
    () => !accountPayload?.isUnsaved,
    [accountPayload?.isUnsaved]
  );

  const disableDiscardButton = useMemo(
    () => isEqual(selectedFilters, appliedFilters),
    [selectedFilters, appliedFilters]
  );

  const renderPropertyFilter = () => (
    <PropertyFilter
      profileType='account'
      source={GROUP_NAME_DOMAINS}
      filtersExpanded={filtersExpanded}
      setFiltersExpanded={setFiltersExpanded}
      filtersList={selectedFilters.filters}
      setFiltersList={setFiltersList}
      secondaryFiltersList={selectedFilters.secondaryFilters}
      setSecondaryFiltersList={setSecondaryFiltersList}
      listEvents={selectedFilters.eventsList}
      setListEvents={setListEvents}
      appliedFilters={appliedFilters}
      selectedAccount={selectedFilters.account}
      eventProp={selectedFilters.eventProp}
      eventTimeline={selectedFilters.eventTimeline}
      areFiltersDirty={areFiltersDirty}
      disableDiscardButton={disableDiscardButton}
      isActiveSegment={Boolean(accountPayload?.segment?.id)}
      applyFilters={applyFilters}
      setSaveSegmentModal={handleSaveSegmentClick}
      setEventProp={setEventProp}
      setEventTimeline={setEventTimeline}
      resetSelectedFilters={resetSelectedFilters}
      onClearFilters={handleClearFilters}
    />
  );

  const renderSaveSegmentButton = () => (
    <ControlledComponent
      controller={!filtersExpanded && !saveButtonDisabled && !newSegmentMode}
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

  const handleAccountSearch = (values) => {
    let valString;
    if (
      (listSearchItems.length >= 1 &&
        listSearchItems[0] === values?.accounts_search) ||
      (listSearchItems.length === 0 && !values?.accounts_search)
    ) {
      return;
    }
    if (values?.accounts_search) {
      valString = [JSON.stringify([values.accounts_search])];
    } else {
      valString = [];
    }
    const updatedPayload = {
      ...accountPayload,
      search_filter: valString.map((vl) => JSON.parse(vl)[0])
    };
    setAccountPayload(updatedPayload);
  };

  const onSearchClose = () => {
    setSearchBarOpen(false);
    handleAccountSearch({ accounts_search: '' });
  };

  const onSearchOpen = () => {
    setSearchBarOpen(true);
  };

  const renderSearchSection = () => (
    <div className='relative'>
      <ControlledComponent controller={searchBarOpen}>
        <div className='flex items-center justify-between'>
          <Form
            name='basic'
            labelCol={{ span: 8 }}
            wrapperCol={{ span: 16 }}
            onFinish={handleAccountSearch}
            autoComplete='off'
          >
            <Form.Item name='accounts_search'>
              <Input
                ref={searchAccountsInputRef}
                size='large'
                value={
                  listSearchItems.length ? listSearchItems.join(', ') : null
                }
                defaultValue={
                  listSearchItems.length ? listSearchItems.join(', ') : null
                }
                placeholder='Search Accounts'
                style={{
                  width: '240px',
                  'border-radius': '5px'
                }}
                prefix={<SVG name='search' size={24} color='#8c8c8c' />}
              />
            </Form.Item>
          </Form>
          <Button type='text' className='search-btn' onClick={onSearchClose}>
            <SVG name='close' size={24} color='#8c8c8c' />
          </Button>
        </div>
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

  const renderDownloadSection = () => (
    <Tooltip title='Download CSV'>
      <Button
        className='search-btn'
        onClick={() => setShowDownloadCSVModal(true)}
        type='text'
      >
        <SVG size={24} name='download' color='#8c8c8c' />
      </Button>
    </Tooltip>
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
        <SVG color='#8c8c8c' size={24} name='more' />
      </Button>
    </Popover>
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

  const handleTableChange = (pageParams, somedata, sorter) => {
    setCurrentPage(pageParams.current);
    setCurrentPageSize(pageParams.pageSize);
    setDefaultSorterInfo({ key: sorter.columnKey, order: sorter.order });
    onDrawerClose();
  };

  const onRowRender = (domainName) => {
    if (!processedDomains.has(domainName)) {
      setProcessedDomains(processedDomains.add(domainName));
      getTop100Events(activeProject.id, domainName);
    }
  };

  const onClickOpen = (domain) => {
    const domID = domain.identity || domain.id;
    const domName = domain.name;
    history.push(`/profiles/accounts/${btoa(domID)}?view=birdview`, {
      accountPayload,
      currentPage,
      currentPageSize,
      activeSorter: defaultSorterInfo,
      appliedFilters: areFiltersDirty ? appliedFilters : null,
      accountsTableRow: domName,
      path: location.pathname
    });
  };

  const onClickOpenNewTab = (domain) => {
    const domID = domain.identity || domain.id;
    window.open(`/profiles/accounts/${btoa(domID)}?view=birdview`);
  };

  useEffect(() => {
    setNewTableColumns(
      getColumns({
        displayTableProps,
        groupPropNames,
        eventNames,
        listProperties,
        defaultSorterInfo,
        projectDomainsList,
        onRowRender,
        onClickOpen,
        onClickOpenNewTab
      })
    );
  }, [
    displayTableProps,
    groupPropNames,
    listProperties,
    defaultSorterInfo,
    projectDomainsList
  ]);

  const tableRef = useRef();

  useEffect(() => {
    // This is the name of Account which was opened recently
    const from = location.state?.accountsTableRow;
    // Finding the tableElement because we have only one .ant-table-body inside tableRef Tree
    // If in future we add table body inside it, need to change it later on
    const tableElement = tableRef.current?.querySelector('.ant-table-body');

    if (tableElement && from && document.getElementById(from)) {
      const element = document.getElementById(from);
      // Y is the relative position that we want to scroll by
      // this is calculated by ORIGINALELEMENTY-TABLEELEMENT - 15 ( because of some padding or margin )
      const y =
        element.getBoundingClientRect().y -
        tableElement.getBoundingClientRect().y -
        15;

      tableElement.scrollTo({ top: y, behavior: 'smooth' });

      location.state.accountsTableRow = '';
    }
  }, [newTableColumns, location.state]);

  const renderTable = useCallback(() => {
    const mergeColumns = newTableColumns.map((col) => ({
      ...col,
      onHeaderCell: (column) => ({
        width: column.width
      })
    }));
    return (
      <div id='resizing-table-container-div'>
        <Table
          ref={tableRef}
          components={{
            header: {
              cell: ResizableTitle
            }
          }}
          onRow={(account) => ({
            onClick: () =>
              setPreview({
                drawerVisible: true,
                domain: { id: account.identity, name: account.domain_name }
              })
          })}
          className='fa-table--profileslist'
          dataSource={tableData}
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
            x: (displayTableProps?.length || 0) * 300,
            y: 'calc(100vh - 340px)'
          }}
        />
      </div>
    );
  }, [tableData, newTableColumns]);

  const showRangeNudge = useMemo(
    () =>
      showUpgradeNudge(
        sixSignalInfo?.usage || 0,
        sixSignalInfo?.limit || 0,
        currentProjectSettings
      ),
    [currentProjectSettings, sixSignalInfo?.limit, sixSignalInfo?.usage]
  );

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
          setAccountPayload({
            source: GROUP_NAME_DOMAINS,
            segment: {},
            isUnsaved: false
          });
          history.replace(PathUrls.ProfileAccounts);
          await getSavedSegments(activeProject.id);
        }
        dispatch(setNewSegmentModeAction(false));
        handleClearFilters();
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
      const reqPayload = {
        ...getFiltersRequestPayload({
          selectedFilters,
          tableProps: displayTableProps
        }),
        name: newSegmentName,
        type: selectedFilters.account[1]
      };

      handleSaveSegment(reqPayload);
    },
    [selectedFilters, displayTableProps]
  );

  const generateCSVData = useCallback(
    (data, selectedOptions) => {
      const csvRows = [];
      const headers = selectedOptions.map(
        (propName) =>
          downloadCSVOptions.find((elem) => elem.prop_name === propName)
            ?.display_name
      );
      headers.unshift('Name');
      csvRows.push(headers.join(','));

      data.forEach((d) => {
        const values = selectedOptions.map((elem) =>
          elem === 'last_activity'
            ? MomentTz(d.last_activity).format('DD MMM YYYY hh:mm A zz')
            : d.table_props[elem] != null
              ? `"${d.table_props[elem]}"`
              : '-'
        );
        values.unshift(d.name);

        csvRows.push(values);
      });

      return csvRows.join('\n');
    },
    [checkListAccountProps]
  );

  const handleDownloadCSV = useCallback(
    async (selectedOptions) => {
      try {
        setCSVDataLoading(true);
        const reqPayload = getFiltersRequestPayload({
          selectedFilters: appliedFilters,
          tableProps: selectedOptions
        });
        const updatedPayload = { ...reqPayload, segment_id: segmentID };
        const resultAccounts = await fetchProfileAccounts(
          activeProject.id,
          updatedPayload,
          true
        );
        const csvData = generateCSVData(resultAccounts.data, selectedOptions);
        downloadCSV(csvData, 'accounts.csv');
        setCSVDataLoading(false);
        setShowDownloadCSVModal(false);
      } catch (err) {
        logger.error(err);
        setCSVDataLoading(false);
        notification.error({
          message: 'Error',
          description: 'CSV download failed',
          duration: 2
        });
      }
    },
    [activeProject.id, appliedFilters, downloadCSVOptions]
  );

  const closeDownloadCSVModal = useCallback(() => {
    setShowDownloadCSVModal(false);
  }, []);

  const pageTitle = useMemo(
    () =>
      newSegmentMode
        ? 'Untitled Segment 1'
        : accountPayload?.segment?.id
          ? accountPayload?.segment?.name
          : 'All Accounts',
    [accountPayload, newSegmentMode]
  );

  const titleIcon = useMemo(() => {
    if (accountPayload?.segment?.id) {
      return defaultSegmentIconsMapping[accountPayload?.segment?.name]
        ? defaultSegmentIconsMapping[accountPayload?.segment?.name]
        : 'pieChart';
    }
    return 'buildings';
  }, [accountPayload?.segment]);

  const titleIconColor = useMemo(
    () => getSegmentColorCode(accountPayload?.segment?.name || ''),
    [accountPayload?.segment?.name]
  );

  return (
    <ProfilesWrapper>
      <ControlledComponent controller={showRangeNudge}>
        <div className='mb-4'>
          <RangeNudge
            title='Accounts Identified'
            amountUsed={sixSignalInfo?.usage || 0}
            totalLimit={sixSignalInfo?.limit || 0}
          />
        </div>
      </ControlledComponent>

      <div className='flex flex-col gap-y-6'>
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
        <ControlledComponent controller={Boolean(accountPayload?.segment?.id)}>
          <AccountsTabs />
        </ControlledComponent>
      </div>

      <ControlledComponent controller={activeTab === 'accounts'}>
        <div className='flex justify-between items-center my-4'>
          <div className='flex items-center gap-x-2 w-full'>
            {renderPropertyFilter()}
            {renderSaveSegmentButton()}
          </div>
          <div className='inline-flex gap-x-2'>
            <ControlledComponent
              controller={!filtersExpanded && !newSegmentMode}
            >
              {renderSearchSection()}
              {renderDownloadSection()}
              {renderTablePropsSelect()}
              {renderMoreActions()}
            </ControlledComponent>
          </div>
        </div>
        <ControlledComponent controller={accounts.isLoading}>
          <Spin size='large' className='fa-page-loader' />
        </ControlledComponent>
        <ControlledComponent
          controller={
            !accounts.isLoading &&
            accounts.data?.[segmentID || 'default']?.length > 0 &&
            (!newSegmentMode || areFiltersDirty)
          }
        >
          <>
            {renderTable()}
            <div className='logo-attrib'>
              <a
                className='font-size--small'
                href='https://clearbit.com'
                target='_blank'
                rel='noopener noreferrer'
              >
                Logos provided by Clearbit
              </a>
            </div>
          </>
        </ControlledComponent>
        <ControlledComponent
          controller={
            !accounts.isLoading &&
            (!accounts.data[segmentID || 'default'] ||
              accounts.data[segmentID || 'default'].length === 0) &&
            (!newSegmentMode || areFiltersDirty)
          }
        >
          <NoDataWithMessage
            message={
              isOnboarded(currentProjectSettings)
                ? !accounts.data[segmentID || 'default'] ||
                  accounts.data[segmentID || 'default'].length === 0
                  ? 'No Accounts found'
                  : errMsg
                : 'Onboarding not completed'
            }
          />
        </ControlledComponent>
      </ControlledComponent>

      <ControlledComponent controller={activeTab === 'insights'}>
        <div className='my-4'>
          <AccountsInsights />
        </div>
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
        segmentName={accountPayload?.segment?.name}
        visible={moreActionsModalMode === moreActionsMode.DELETE}
        onCancel={() => setMoreActionsModalMode(null)}
        onOk={handleDeleteActiveSegment}
      />
      <RenameSegmentModal
        segmentName={accountPayload?.segment?.name}
        visible={moreActionsModalMode === moreActionsMode.RENAME}
        onCancel={() => setMoreActionsModalMode(null)}
        handleSubmit={handleRenameSegment}
      />

      <UpdateSegmentModal
        segmentName={accountPayload?.segment?.name}
        visible={updateSegmentModal}
        onCancel={() => setUpdateSegmentModal(false)}
        onCreate={handleCreateSegment}
        onUpdate={handleUpdateSegmentDefinition}
      />
      <DownloadCSVModal
        visible={showDownloadCSVModal}
        onCancel={closeDownloadCSVModal}
        options={downloadCSVOptions}
        displayTableProps={displayTableProps}
        onSubmit={handleDownloadCSV}
        isLoading={csvDataLoading}
      />
      <AccountDrawer
        domain={preview.domain.name}
        events={accountPreview[preview.domain.name]}
        visible={preview.drawerVisible}
        onClose={onDrawerClose}
        onClickMore={() => onClickOpen(preview.domain)}
        onClickOpenNewtab={() => onClickOpenNewTab(preview.domain)}
      />
    </ProfilesWrapper>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  accounts: state.timelines.accounts,
  segments: state.timelines.segments,
  accountPreview: state.timelines.accountPreview,
  currentProjectSettings: state.global.currentProjectSettings
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      getGroups,
      getProfileAccounts,
      createNewSegment,
      getSavedSegments,
      getGroupProperties,
      fetchProjectSettings,
      udpateProjectSettings,
      updateSegmentForId,
      deleteSegment,
      getTop100Events
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(AccountProfiles);
