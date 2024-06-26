import React, {
  useState,
  useEffect,
  useMemo,
  useCallback,
  useRef
} from 'react';
import isEqual from 'lodash/isEqual';
import {
  Table,
  Button,
  Popover,
  Tabs,
  notification,
  Input,
  Form,
  Tooltip,
  Spin,
  message
} from 'antd';
import { connect, useDispatch, useSelector } from 'react-redux';
import { bindActionCreators } from 'redux';
import { useHistory, useLocation, useParams } from 'react-router-dom';
import SearchCheckList from 'Components/SearchCheckList';
import { formatUserPropertiesToCheckList } from 'Reducers/timelines/utils';
import {
  setAccountPayloadAction,
  setActiveDomainAction,
  setDrawerVisibleAction,
  setFiltersDirtyAction,
  setNewSegmentModeAction
} from 'Reducers/accountProfilesView/actions';
import useFeatureLock from 'hooks/useFeatureLock';
import { FEATURES } from 'Constants/plans.constants';
import ControlledComponent from 'Components/ControlledComponent/ControlledComponent';
import { selectGroupsList } from 'Reducers/groups/selectors';
import {
  fetchProfileAccounts,
  fetchSegmentById,
  moveSegmentToNewFolder,
  updateSegmentToFolder,
  updateTableProperties,
  updateTablePropertiesForSegment
} from 'Reducers/timelines';
import { downloadCSV } from 'Utils/csv';
import { PathUrls } from 'Routes/pathUrls';
import { GROUP_NAME_DOMAINS } from 'Components/GlobalFilter/FilterWrapper/utils';
import { defaultSegmentIconsMapping } from 'Views/AppSidebar/appSidebar.constants';
import { isOnboarded } from 'Utils/global';
import _, { cloneDeep } from 'lodash';
import { getSegmentColorCode } from 'Views/AppSidebar/appSidebar.helpers';
import ResizableTitle from 'Components/Resizable';
import logger from 'Utils/logger';
import { invalidBreakdownPropertiesList } from 'Constants/general.constants';
import { INITIAL_ACCOUNT_PAYLOAD } from 'Reducers/accountProfilesView';
import usePrevious from 'hooks/usePrevious';
import { getGroups, getGroupProperties } from 'Reducers/coreQuery/middleware';
import { fetchProjectSettings } from 'Reducers/global';
import {
  getProfileAccounts,
  createNewSegment,
  getSavedSegments,
  updateSegmentForId,
  deleteSegment,
  getTop100Events,
  getSegmentFolders
} from 'Reducers/timelines/middleware';
import { FolderItemOptions } from 'Components/FolderStructure/FolderItem';
import UpgradeNudge from 'Components/GenericComponents/UpgradeNudge';
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
import {
  defaultSegmentsList,
  getColumns,
  renderValue,
  checkFiltersEquality
} from './accountProfiles.helpers';
import ProfilesWrapper from '../ProfilesWrapper';
import NoDataWithMessage from '../MyComponents/NoDataWithMessage';
import {
  formatReqPayload,
  getFiltersRequestPayload,
  getPropType,
  getSelectedFiltersFromQuery
} from '../utils';
import PropertyFilter from './PropertyFilter';
import { Text, SVG } from '../../factorsComponents';
import AccountsTabs from './AccountsTabs';
import AccountsInsights from './AccountsInsights/AccountsInsights';
import AccountDrawer from './AccountDrawer';
import InsightsWrapper from './InsightsWrapper';
import { PROFILE_TYPE_ACCOUNT } from '../constants';
import useAutoFocus from 'hooks/useAutoFocus';
import { EditOutlined, EnterOutlined } from '@ant-design/icons';
import EditableTitle from 'Components/GenericComponents/EditableTitle';

function AccountProfiles({
  createNewSegment,
  getSavedSegments,
  fetchProjectSettings,
  getGroups,
  getProfileAccounts,
  getGroupProperties,
  updateSegmentForId,
  deleteSegment,
  getTop100Events,
  getSegmentFolders
}) {
  const tableRef = useRef();

  const dispatch = useDispatch();
  const history = useHistory();
  const location = useLocation();
  const { segment_id: segmentID } = useParams();

  // General
  const [isUpgradeModalVisible, setIsUpgradeModalVisible] = useState(false);
  const [errMsg, setErrMsg] = useState('');

  // Ant Table
  const [newTableColumns, setNewTableColumns] = useState([]);
  const [currentPage, setCurrentPage] = useState(1);
  const [currentPageSize, setCurrentPageSize] = useState(25);
  const [defaultSorterInfo, setDefaultSorterInfo] = useState({});

  // Search Bar
  const [searchTerm, setSearchTerm] = useState('');

  // Column Selector
  const [showTableColumnsDD, setShowTableColumnsDD] = useState(false);
  const [tableColumnsList, setTableColumnsList] = useState([]);
  const [selectedTableColumnsList, setSelectedTableColumnsList] = useState([]);

  // Segment
  const [filtersExpanded, setFiltersExpanded] = useState(false);
  const [selectedFilters, setSelectedFilters] = useState(INITIAL_FILTERS_STATE);
  const [appliedFilters, setAppliedFilters] = useState(INITIAL_FILTERS_STATE);
  const [moreActionsModalMode, setMoreActionsModalMode] = useState(null); // DELETE | RENAME
  const [saveSegmentModal, setSaveSegmentModal] = useState(false);
  const [updateSegmentModal, setUpdateSegmentModal] = useState(false);

  // Download
  const [showDownloadCSVModal, setShowDownloadCSVModal] = useState(false);
  const [csvDataLoading, setCSVDataLoading] = useState(false);
  const [downloadCSVOptions, setDownloadCSVOptions] = useState([]);

  // Drawer
  const [processedDomains, setProcessedDomains] = useState(new Set());

  const handleDrawerClose = () => {
    dispatch(setDrawerVisibleAction(false));
  };

  useEffect(() => {
    if (filtersExpanded) handleDrawerClose();
  }, [filtersExpanded]);

  const { accounts, accountSegments, segmentFolders } = useSelector(
    (state) => state.timelines
  );

  const {
    active_project: activeProject,
    currentProjectSettings,
    projectDomainsList
  } = useSelector((state) => state.global);

  const { groups, groupProperties, groupPropNames, eventNames } = useSelector(
    (state) => state.coreQuery
  );

  const {
    activeTab,
    accountPayload,
    newSegmentMode,
    filtersDirty: areFiltersDirty,
    preview
  } = useSelector((state) => state.accountProfilesView);

  const { loading: isFeatureLoading } = useSelector(
    (state) => state.featureConfig
  );

  const { isFeatureLocked: isScoringLocked } = useFeatureLock(
    FEATURES.FEATURE_ACCOUNT_SCORING
  );

  const groupsList = useSelector((state) => selectGroupsList(state));

  const Wrapper = activeTab === 'accounts' ? ProfilesWrapper : InsightsWrapper;

  const previousSegmentId = usePrevious(segmentID);

  const activeID = useMemo(() => segmentID || 'default', [segmentID]);

  const activeSegmentName = useMemo(() => {
    const activeSegment = accountSegments.find(
      (segment) => segment.id === segmentID
    );
    return activeSegment?.name || 'All Accounts';
  }, [accountSegments, segmentID]);

  const pageTitle = useMemo(() => {
    if (newSegmentMode) {
      return 'Untitled Segment 1';
    }

    if (segmentID) {
      return activeSegmentName;
    }

    return 'All Accounts';
  }, [segmentID, activeSegmentName, newSegmentMode]);

  const titleIcon = useMemo(() => {
    if (segmentID) {
      return defaultSegmentIconsMapping[activeSegmentName]
        ? defaultSegmentIconsMapping[activeSegmentName]
        : 'pieChart';
    }
    return 'regularBuilding';
  }, [segmentID, activeSegmentName]);

  const titleIconColor = useMemo(
    () => getSegmentColorCode(activeSegmentName || ''),
    [activeSegmentName]
  );

  const displayTableProps = useMemo(() => {
    const filterNullEntries = (entry) =>
      entry !== '' && entry !== undefined && entry !== null;

    const getFilteredTableProps = (tableProps) =>
      tableProps?.filter(filterNullEntries) || [];

    const segmentTableProps = accountPayload?.segment?.query?.table_props;
    const projectTableProps =
      currentProjectSettings?.timelines_config?.account_config?.table_props;

    const tableProps = segmentID
      ? getFilteredTableProps(segmentTableProps)
      : getFilteredTableProps(projectTableProps);

    return tableProps;
  }, [
    currentProjectSettings,
    accountPayload?.segment,
    accountPayload?.segment?.query?.table_props
  ]);

  const tableData = useMemo(() => {
    const sortedData = accounts?.[activeID]?.profiles?.sort(
      (a, b) => new Date(b.last_activity) - new Date(a.last_activity)
    );
    return sortedData?.map((row) => ({
      ...row,
      ...row?.table_props
    }));
  }, [accounts, segmentID]);

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
        areFiltersDirty,
        isActiveSegment: Boolean(segmentID)
      }),
    [
      appliedFilters,
      selectedFilters,
      newSegmentMode,
      areFiltersDirty,
      segmentID
    ]
  );

  const setAccountPayload = (payload) => {
    dispatch(setAccountPayloadAction(payload));
    if (payload?.segment?.id) {
      dispatch(setNewSegmentModeAction(false));
    }
  };

  useEffect(() => {
    if (location.pathname === PathUrls.ProfileAccounts && !newSegmentMode) {
      setAccountPayload(INITIAL_ACCOUNT_PAYLOAD);
    }
  }, [location.pathname, newSegmentMode]);

  const getAccountPayload = async () => {
    if (newSegmentMode) {
      return {};
    }

    if (accountPayload?.isUnsaved || accountPayload?.search_filter) {
      return accountPayload;
    }

    if (!segmentID) {
      return INITIAL_ACCOUNT_PAYLOAD;
    }

    const response = await fetchSegmentById(activeProject.id, segmentID);
    if (!response.ok) {
      return INITIAL_ACCOUNT_PAYLOAD;
    }

    return segmentID === accountPayload?.segment?.id
      ? { ...accountPayload, segment: response.data }
      : { source: GROUP_NAME_DOMAINS, segment: response.data };
  };

  const runInit = async () => {
    const payload = await getAccountPayload();
    if (!_.isEqual(payload, accountPayload)) {
      setAccountPayload(payload);
    }
  };

  useEffect(() => {
    if (activeProject.id) runInit();
  }, [activeProject?.id, segmentID, accountPayload]);

  const getAccounts = async (payload) => {
    if (!payload || Object.keys(payload).length === 0) {
      return;
    }
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
  };

  useEffect(() => {
    const shouldCache = location?.state?.fromDetails;
    if (shouldCache) {
      if (!location.state.accountPayload) {
        setAccountPayload(INITIAL_ACCOUNT_PAYLOAD);
        getAccounts(accountPayload);
      } else {
        const {
          currentPage: cachedCurrentPage,
          currentPageSize: cachedCurrentPageSize,
          activeSorter,
          accountPayload: payload
        } = location.state;
        setCurrentPage(cachedCurrentPage);
        setCurrentPageSize(cachedCurrentPageSize);
        setDefaultSorterInfo(activeSorter);
        setAccountPayload(payload);
        dispatch(setNewSegmentModeAction(false));
      }
      const updatedLocationState = { ...location.state, fromDetails: false };
      history.replace(location.pathname, updatedLocationState);
    } else if (
      !segmentID ||
      !accountPayload?.segment?.id ||
      segmentID === accountPayload.segment.id
    ) {
      getAccounts(accountPayload);
    }
  }, [accountPayload]);

  useEffect(() => {
    if (activeProject?.id) {
      fetchProjectSettings(activeProject.id);
      getGroups(activeProject.id);
      getSavedSegments(activeProject.id);
    }
  }, [activeProject?.id]);

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

    setTableColumnsList([...filteredDomainProps, ...groupProps]);
  }, [groupProperties, groups]);

  useEffect(() => {
    const tableProps = segmentID
      ? accountPayload?.segment?.query?.table_props
      : currentProjectSettings.timelines_config?.account_config?.table_props ||
        [];
    const accountPropsWithEnableKey = formatUserPropertiesToCheckList(
      tableColumnsList,
      tableProps?.filter(
        (entry) => entry !== '' && entry !== undefined && entry !== null
      )
    );
    const csvPropsWithEnableKey = formatUserPropertiesToCheckList(
      [...tableColumnsList, ['Last Activity', 'last_activity', 'datetime']],
      tableProps?.filter(
        (entry) => entry !== '' && entry !== undefined && entry !== null
      )
    );
    setDownloadCSVOptions(csvPropsWithEnableKey);
    setSelectedTableColumnsList(accountPropsWithEnableKey);
  }, [currentProjectSettings, tableColumnsList, accountPayload]);

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
    if (!accountPayload?.search_filter?.length) {
      setSearchTerm('');
    } else {
      const listValues = accountPayload.search_filter || [];
      setSearchTerm(listValues[0]);
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

        if (previousSegmentId !== segmentID) {
          setFiltersDirty(false);
        }
        if (segmentID) setFiltersExpanded(false);
      } else {
        restoreFiltersDefaultState();
      }
    }
  }, [accountPayload, newSegmentMode]);

  const onRenameSegment = async (name) => {
    if (name === activeSegmentName) {
      return;
    }

    const messageHandler = message.loading('Renaming Segment', 0);
    try {
      await updateSegmentForId(activeProject.id, segmentID, {
        name
      });
      await getSavedSegments(activeProject.id);
      setMoreActionsModalMode(null);
      notification.success({
        message: 'Segment renamed successfully',
        duration: 3
      });
    } catch (error) {
      notification.error({
        message: 'Segment rename failed',
        duration: 3
      });
    } finally {
      messageHandler();
    }
  };

  const handleRenameSegment = async (name) => {
    if (name === activeSegmentName) return;
    const messageHandler = message.loading('Renaming Segment', 0);
    try {
      await updateSegmentForId(activeProject.id, segmentID, { name });
      await getSavedSegments(activeProject.id);
      setMoreActionsModalMode(null);
      notification.success({
        message: 'Segment renamed successfully',
        duration: 3
      });
    } catch (error) {
      notification.error({
        message: 'Segment rename failed',
        duration: 3
      });
    } finally {
      messageHandler();
    }
  };

  const handleUpdateSegmentDefinition = async () => {
    try {
      const reqPayload = getFiltersRequestPayload({
        selectedFilters,
        tableProps: displayTableProps
      });

      await updateSegmentForId(activeProject.id, segmentID, reqPayload);

      await getSavedSegments(activeProject.id);
      setUpdateSegmentModal(false);
      setFiltersDirty(false);
      notification.success({
        message: 'Segment updated successfully',
        duration: 3
      });
    } catch (error) {
      notification.error({
        message: 'Segment update failed',
        duration: 3
      });
    }
  };

  const handleDeleteActiveSegment = () => {
    const messageHandler = message.loading('Deleting Segment', 0);
    deleteSegment({
      projectId: activeProject.id,
      segmentId: accountPayload.segment.id
    })
      .then(() => {
        setMoreActionsModalMode(null);
        notification.success({
          message: 'Segment deleted successfully',
          duration: 5
        });
      })
      .finally(() => {
        messageHandler();
        setAccountPayload(INITIAL_ACCOUNT_PAYLOAD);
        history.replace(PathUrls.ProfileAccounts);
      });
  };

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
      selectedTableColumnsList.filter((item) => item.enabled).length < 12
    ) {
      setSelectedTableColumnsList((prev) => {
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
    const newTableProps =
      selectedTableColumnsList
        ?.filter(({ enabled }) => enabled)
        ?.map(({ prop_name }) => prop_name)
        ?.filter((entry) => entry !== '' && entry !== undefined) || [];

    if (segmentID?.length) {
      const response = await updateTablePropertiesForSegment(
        activeProject.id,
        segmentID,
        newTableProps
      );

      if (!response.ok) return;

      await getSavedSegments(activeProject.id);

      const updatedPayload = { ...accountPayload };
      updatedPayload.segment.query.table_props = newTableProps;
      setAccountPayload(updatedPayload);
    } else {
      await updateTableProperties(
        activeProject.id,
        PROFILE_TYPE_ACCOUNT,
        newTableProps
      );
      await fetchProjectSettings(activeProject.id);
      setAccountPayload({ ...accountPayload });
    }
    setShowTableColumnsDD(false);
  };

  const handleDisableOptionClick = () => {
    setIsUpgradeModalVisible(true);
    setShowTableColumnsDD(false);
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
          mapArray={selectedTableColumnsList}
          sortable
          updateList={setSelectedTableColumnsList}
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
    history.push(
      `${PathUrls.SettingsAccountScoring}?activeTab=engagementScoring`
    );
  }, []);

  const handleSaveSegmentClick = useCallback(() => {
    if (newSegmentMode) {
      setSaveSegmentModal(true);
      return;
    }
    if (segmentID) {
      setUpdateSegmentModal(true);
    } else {
      setSaveSegmentModal(true);
    }
  }, [segmentID, newSegmentMode]);

  const resetSelectedFilters = useCallback(() => {
    setSelectedFilters(appliedFilters);
  }, [appliedFilters]);

  const handleClearFilters = () => {
    restoreFiltersDefaultState(true);
  };

  const renderPropertyFilter = () => (
    <PropertyFilter
      profileType='account'
      filtersExpanded={filtersExpanded}
      setFiltersExpanded={setFiltersExpanded}
      selectedFilters={selectedFilters}
      setSelectedFilters={setSelectedFilters}
      resetSelectedFilters={resetSelectedFilters}
      appliedFilters={appliedFilters}
      applyFilters={applyFilters}
      areFiltersDirty={areFiltersDirty}
      disableDiscardButton={disableDiscardButton}
      isActiveSegment={Boolean(segmentID)}
      setSaveSegmentModal={handleSaveSegmentClick}
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
      (searchTerm !== '' && searchTerm === values?.accounts_search) ||
      (searchTerm === '' && !values?.accounts_search)
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

  const renderSearchSection = () => (
    <div className='relative flex items-center justify-between'>
      <Form
        name='basic'
        labelCol={{ span: 8 }}
        wrapperCol={{ span: 16 }}
        onFinish={handleAccountSearch}
        autoComplete='off'
      >
        <Form.Item name='accounts_search'>
          <Input
            value={searchTerm}
            defaultValue={searchTerm}
            placeholder='Search Account'
            style={{
              width: '224px',
              'border-radius': '5px'
            }}
            prefix={<SVG name='search' size={16} color='#8c8c8c' />}
          />
        </Form.Item>
      </Form>
    </div>
  );

  const renderDownloadSection = () => (
    <Tooltip title='Download CSV'>
      <Button onClick={() => setShowDownloadCSVModal(true)} type='text'>
        <SVG size={16} name='download' color='#8c8c8c' />
      </Button>
    </Tooltip>
  );

  const moveSegmentToFolder = async (folderID, segment_id) => {
    const messageHandler = message.loading('Moving Segment to Folder', 0);
    try {
      await updateSegmentToFolder(
        activeProject.id,
        segment_id,
        {
          folder_id: folderID
        },
        'account'
      );
      await getSavedSegments(activeProject.id);
      message.success('Segment Moved');
    } catch (err) {
      logger.error(err);
      message.error('Segment failed to move');
    } finally {
      messageHandler();
    }
  };
  const handleMoveToNewFolder = async (segment_id, folder_name) => {
    const messageHandler = message.loading(
      `Moving Segment to \`${folder_name}\` Folder`,
      0
    );
    try {
      await moveSegmentToNewFolder(
        activeProject.id,
        segment_id,
        {
          name: folder_name
        },
        'account'
      );
      getSegmentFolders(activeProject.id, 'account');
      await getSavedSegments(activeProject.id);
      message.success('Segment Moved to New Folder');
    } catch (err) {
      logger.error(err);
      message.error('Failed to move segment');
    } finally {
      messageHandler();
    }
  };

  const renderMoreActions = () => (
    <div className='cursor-pointer'>
      <FolderItemOptions
        id={segmentID}
        unit='segment'
        folder_id={accountPayload?.segment?.folder_id}
        folders={[{ id: '', name: 'All Segments' }, ...segmentFolders.accounts]}
        handleEditUnit={() => {
          setMoreActionsModalMode(moreActionsMode.RENAME);
        }}
        handleDeleteUnit={() => {
          setMoreActionsModalMode(moreActionsMode.DELETE);
        }}
        moveToExistingFolder={moveSegmentToFolder}
        handleNewFolder={handleMoveToNewFolder}
        extraOptions={[
          {
            id: 'extra-4',
            title: 'Account Engagement Rules',
            icon: <SVG size={20} name='fireFlameCurved' color='#8c8c8c' />,
            onClick: navigateToAccountsEngagement
          }
        ]}
        hideDefaultOptions={
          !!segmentID === !!'' ||
          defaultSegmentsList.includes(activeSegmentName)
        }
        placement='bottom'
        hideMoveTo={!!segmentID === !!''}
      >
        <Button className='button-shadow'>
          <SVG size={16} color='#8C8C8C' name='more' />
        </Button>
      </FolderItemOptions>
    </div>
  );

  const renderTablePropsSelect = () => (
    <Popover
      overlayClassName='fa-activity--filter'
      placement='bottomLeft'
      visible={showTableColumnsDD}
      onVisibleChange={(visible) => {
        setShowTableColumnsDD(visible);
      }}
      onClick={() => {
        setShowTableColumnsDD(true);
      }}
      trigger='click'
      content={popoverContent}
    >
      <Tooltip title='Edit columns'>
        <Button type='text'>
          <SVG size={16} color='#8c8c8c' name='tableColumns' />
        </Button>
      </Tooltip>
    </Popover>
  );

  const handleTableChange = (pageParams, somedata, sorter) => {
    setCurrentPage(pageParams.current);
    setCurrentPageSize(pageParams.pageSize);
    setDefaultSorterInfo({ key: sorter.columnKey, order: sorter.order });
    handleDrawerClose();
  };

  const onClickOpen = (domain) => {
    const domID = domain.identity || domain.id;
    const domName = domain.name;
    history.push(`/profiles/accounts/${btoa(domID)}?view=timeline`, {
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
    window.open(`/profiles/accounts/${btoa(domID)}?view=timeline`);
  };

  useEffect(() => {
    setNewTableColumns(
      getColumns({
        displayTableProps,
        groupPropNames,
        eventNames,
        listProperties: tableColumnsList,
        defaultSorterInfo,
        projectDomainsList,
        onClickOpen,
        onClickOpenNewTab,
        previewState: preview
      })
    );
  }, [
    displayTableProps,
    groupPropNames,
    tableColumnsList,
    defaultSorterInfo,
    projectDomainsList,
    preview
  ]);

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

  const handleTableRowClick = (account) => {
    dispatch(
      setActiveDomainAction({ id: account.identity, name: account.domain_name })
    );

    if (!processedDomains.has(account.domain_name)) {
      setProcessedDomains(processedDomains.add(account.domain_name));
      getTop100Events(activeProject.id, account.domain_name);
    }
  };

  const getRowClassName = (account) => {
    const isActive =
      preview?.drawerVisible && account?.domain_name === preview?.domain?.name;
    return isActive ? 'active cursor-pointer' : 'cursor-pointer';
  };

  const renderTable = useCallback(() => {
    const getHeaderCellProps = (column) => ({
      width: column.width
    });

    const createMergedColumns = (columns) =>
      columns.map((col) => ({
        ...col,
        onHeaderCell: getHeaderCellProps
      }));

    const handleRowClick = (account) => ({
      onClick: () => handleTableRowClick(account)
    });

    const mergedColumns = createMergedColumns(newTableColumns);

    return (
      <div id='resizing-table-container-div'>
        <Table
          ref={tableRef}
          components={{
            header: {
              cell: ResizableTitle
            }
          }}
          onRow={handleRowClick}
          className='fa-table--profileslist'
          dataSource={tableData}
          columns={mergedColumns}
          rowClassName={getRowClassName}
          pagination={{
            position: ['bottom', 'left'],
            defaultPageSize: 25,
            current: currentPage,
            pageSize: currentPageSize,
            size: 'small'
          }}
          onChange={handleTableChange}
          scroll={{
            x: '100%'
          }}
        />
      </div>
    );
  }, [
    tableData,
    newTableColumns,
    preview,
    displayTableProps,
    currentPage,
    currentPageSize
  ]);

  const handleSaveSegment = async (segmentPayload) => {
    try {
      const response = await createNewSegment(activeProject.id, segmentPayload);
      if (response.type === 'SEGMENT_CREATION_FULFILLED') {
        notification.success({
          message: 'Success!',
          description: response.payload.message,
          duration: 3
        });
        await getSavedSegments(activeProject.id);
        history.replace({
          pathname: `/accounts/segments/${response.payload.segment.id}`
        });
        setSaveSegmentModal(false);
        setUpdateSegmentModal(false);
        setFiltersDirty(false);
      }
      dispatch(setNewSegmentModeAction(false));
      dispatch(
        setAccountPayloadAction({
          source: GROUP_NAME_DOMAINS,
          segment: response?.payload?.segment
        })
      );
    } catch (err) {
      notification.error({
        message: 'Error',
        description:
          err?.data?.error || 'Segment Creation Failed. Invalid Parameters.',
        duration: 3
      });
    }
  };

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
      // Generate CSV headers
      const headers = [
        '"Account Domain"',
        ...selectedOptions.map((propName) => {
          const option = downloadCSVOptions.find(
            (elem) => elem.prop_name === propName
          );
          return `"${option?.display_name}"`;
        })
      ];

      // Sort data
      const sortedData = data.sort((a, b) => {
        if ('score' in a) {
          return b.score - a.score;
        }
        return a.domain_name.localeCompare(b.domain_name);
      });

      // Generate CSV rows
      const csvRows = sortedData.map((d) => {
        const values = selectedOptions.map((elem) => {
          const propType = getPropType(tableColumnsList, elem);
          if (elem === 'last_activity') {
            return d.last_activity?.replace('T', ' ').replace('Z', '');
          }
          return renderValue(
            d.table_props[elem],
            propType,
            elem,
            projectDomainsList,
            true
          );
        });
        return [d.domain_name, ...values];
      });

      // Combine headers and rows into CSV string
      return [headers.join(','), ...csvRows.map((row) => row.join(','))].join(
        '\n'
      );
    },
    [selectedTableColumnsList]
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

  const renderPageTitle = () => (
    <div className='flex gap-x-2 items-center'>
      <div className='flex items-center rounded justify-center'>
        <SVG name={titleIcon} size={24} color={titleIconColor} />
      </div>
      <EditableTitle
        title={pageTitle}
        editable={
          segmentID &&
          !defaultSegmentsList.includes(pageTitle) &&
          !newSegmentMode
        }
        editIcon={<EditOutlined />}
        enterIcon={<EnterOutlined />}
        handleEdit={handleRenameSegment}
      />
    </div>
  );

  const renderHeader = () => (
    <div className='profiles-header'>
      {renderPageTitle()}
      <ControlledComponent controller={Boolean(segmentID)}>
        <div className='flex items-center justify-between gap-x-1'>
          <AccountsTabs />
          {renderMoreActions()}
        </div>
      </ControlledComponent>
    </div>
  );

  const renderRangeNudge = () => <UpgradeNudge showCarousel />;

  const renderProfileActions = () => (
    <div className='flex justify-between items-cente py-4 px-8'>
      <div className='flex items-center gap-x-2 w-full'>
        {renderPropertyFilter()}
        {renderSaveSegmentButton()}
      </div>
      <div className='inline-flex gap-x-2 h-8'>
        <ControlledComponent controller={!filtersExpanded && !newSegmentMode}>
          {renderSearchSection()}
          {renderDownloadSection()}
          {renderTablePropsSelect()}
        </ControlledComponent>
      </div>
    </div>
  );

  const renderLoaderDiv = () => (
    <div className='accounts-loader-div'>
      <Spin size='large' />
    </div>
  );

  const renderLogoAttribution = () => (
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
  );

  const renderNoDataComponent = () => (
    <NoDataWithMessage
      message={
        isOnboarded(currentProjectSettings)
          ? !accounts?.[activeID]?.profiles ||
            accounts?.[activeID]?.profiles?.length === 0
            ? 'No Accounts found'
            : errMsg
          : 'Onboarding not completed'
      }
    />
  );

  const renderPreviewBanner = () => {
    if (!accounts?.[activeID]?.isPreview) {
      return null;
    }

    return (
      <div className='pb-4 px-8'>
        <div className='preview-info'>
          <SVG name='CircleExclamation' color='#FAAD14' />
          <div>
            This is only a preview of your segment. Save changes to the segment
            to begin processing the full results.
          </div>
        </div>
      </div>
    );
  };

  if (isFeatureLoading) return renderLoaderDiv();

  return (
    <Wrapper>
      {renderHeader()}
      {renderRangeNudge()}
      <ControlledComponent controller={activeTab === 'accounts'}>
        {renderProfileActions()}
        <ControlledComponent controller={false}>
          {renderPreviewBanner()}
        </ControlledComponent>
        <ControlledComponent controller={accounts?.[activeID]?.isLoading}>
          {renderLoaderDiv()}
        </ControlledComponent>
        <ControlledComponent
          controller={
            !accounts?.[activeID]?.isLoading &&
            accounts?.[activeID]?.profiles?.length > 0 &&
            (!newSegmentMode || areFiltersDirty)
          }
        >
          {renderTable()}
          {renderLogoAttribution()}
        </ControlledComponent>
        <ControlledComponent
          controller={
            !accounts?.[activeID]?.isLoading &&
            (!accounts?.[activeID]?.profiles ||
              accounts?.[activeID]?.profiles?.length === 0) &&
            (!newSegmentMode || areFiltersDirty)
          }
        >
          {renderNoDataComponent()}
        </ControlledComponent>
      </ControlledComponent>

      <ControlledComponent
        controller={activeTab === 'insights' && segmentID != null}
      >
        <div className='my-4 flex-1 flex flex-col px-10'>
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
        segmentName={activeSegmentName}
        visible={moreActionsModalMode === moreActionsMode.DELETE}
        onCancel={() => setMoreActionsModalMode(null)}
        onOk={handleDeleteActiveSegment}
      />
      <RenameSegmentModal
        segmentName={activeSegmentName}
        visible={moreActionsModalMode === moreActionsMode.RENAME}
        onCancel={() => setMoreActionsModalMode(null)}
        handleSubmit={handleRenameSegment}
      />

      <UpdateSegmentModal
        segmentName={activeSegmentName}
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
        visible={preview.drawerVisible}
        onClose={handleDrawerClose}
        onClickMore={() => onClickOpen(preview.domain)}
        onClickOpenNewtab={() => onClickOpenNewTab(preview.domain)}
      />
    </Wrapper>
  );
}

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      getGroups,
      getProfileAccounts,
      createNewSegment,
      getSavedSegments,
      getGroupProperties,
      fetchProjectSettings,
      updateSegmentForId,
      deleteSegment,
      getTop100Events,
      getSegmentFolders
    },
    dispatch
  );

export default connect(null, mapDispatchToProps)(AccountProfiles);
