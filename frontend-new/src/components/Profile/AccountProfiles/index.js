import React, { useState, useEffect, useMemo, useCallback } from 'react';
import isEqual from 'lodash/isEqual';
import cx from 'classnames';
import { Table, Button, Spin, Popover, Tabs, notification, Input } from 'antd';
import { connect, useDispatch, useSelector } from 'react-redux';
import { bindActionCreators } from 'redux';
import { useHistory, useLocation, useParams } from 'react-router-dom';
import { Text, SVG } from '../../factorsComponents';
import PropertyFilter from './PropertyFilter';
import { getGroupProperties } from '../../../reducers/coreQuery/middleware';
import FaSelect from '../../FaSelect';
import {
  DEFAULT_TIMELINE_CONFIG,
  formatFiltersForPayload,
  formatReqPayload,
  getFiltersRequestPayload,
  getSelectedFiltersFromQuery,
  IsDomainGroup
} from '../utils';
import {
  getProfileAccounts,
  createNewSegment,
  getSavedSegments,
  updateSegmentForId,
  deleteSegment
} from '../../../reducers/timelines/middleware';
import {
  fetchProjectSettings,
  udpateProjectSettings
} from '../../../reducers/global';
import SearchCheckList from 'Components/SearchCheckList';
import { formatUserPropertiesToCheckList } from 'Reducers/timelines/utils';
import { fetchGroupPropertyValues } from 'Reducers/coreQuery/services';
import NoDataWithMessage from '../MyComponents/NoDataWithMessage';
import ProfilesWrapper from '../ProfilesWrapper';
import { checkFiltersEquality, getColumns } from './accountProfiles.helpers';
import {
  selectAccountPayload,
  selectActiveSegment
} from 'Reducers/accountProfilesView/selectors';
import {
  setAccountPayloadAction,
  setActiveSegmentAction,
  setFiltersDirtyAction,
  setNewSegmentModeAction,
  updateAccountPayloadAction
} from 'Reducers/accountProfilesView/actions';
import useFeatureLock from 'hooks/useFeatureLock';
import { FEATURES } from 'Constants/plans.constants';
import UpgradeModal from '../UpgradeModal';
import RangeNudge from 'Components/GenericComponents/RangeNudge';
import get from 'lodash/get';
import uniq from 'lodash/uniq';
import { showUpgradeNudge } from 'Views/Settings/ProjectSettings/Pricing/utils';
import ControlledComponent from 'Components/ControlledComponent/ControlledComponent';
import SaveSegmentModal from './SaveSegmentModal';
import DeleteSegmentModal from './DeleteSegmentModal';
import RenameSegmentModal from './RenameSegmentModal';
import {
  INITIAL_FILTERS_STATE,
  moreActionsMode
} from './accountProfiles.constants';
import { selectGroupsList } from 'Reducers/groups/selectors';
import UpdateSegmentModal from './UpdateSegmentModal';
import DownloadCSVModal from './DownloadCSVModal';
import { fetchProfileAccounts } from 'Reducers/timelines';
import { downloadCSV } from 'Utils/csv';
import { formatCount } from 'Utils/dataFormatter';
import { PathUrls } from 'Routes/pathUrls';
import { getGroups } from '../../../reducers/coreQuery/middleware';
import { GROUP_NAME_DOMAINS } from 'Components/GlobalFilter/FilterWrapper/utils';
import styles from './index.module.scss';
import { defaultSegmentIconsMapping } from 'Views/AppSidebar/appSidebar.constants';
import { isOnboarded } from 'Utils/global';
import { cloneDeep } from 'lodash';
import { getSegmentColorCode } from 'Views/AppSidebar/appSidebar.helpers';

import { COLUMN_TYPE_PROPS } from 'Utils/table';
import ResizableTitle from 'Components/Resizable';

const groupToDomainMap = {
  $hubspot_company: '$hubspot_company_domain',
  $salesforce_account: '$salesforce_account_website',
  $6signal: '$6Signal_domain',
  $linkedin_company: '$li_domain',
  $g2: '$g2_domain'
};

const filterPropsMap = {
  $hubspot_company: 'hubspot',
  $salesforce_account: 'salesforce',
  $6signal: '6Signal',
  $linkedin_company: '$li_',
  $g2: '$g2',
  $domains: ''
};

function AccountProfiles({
  activeProject,
  groups,
  accounts,
  createNewSegment,
  getSavedSegments,
  fetchProjectSettings,
  getGroups,
  udpateProjectSettings,
  currentProjectSettings,
  getProfileAccounts,
  getGroupProperties,
  updateSegmentForId,
  deleteSegment,
  segments
}) {
  const dispatch = useDispatch();
  const history = useHistory();
  const location = useLocation();
  const { segment_id } = useParams();

  const { groupPropNames } = useSelector((state) => state.coreQuery);
  const groupProperties = useSelector(
    (state) => state.coreQuery.groupProperties
  );
  const accountPayload = useSelector((state) => {
    return selectAccountPayload(state);
  });

  const activeSegment = useSelector((state) => {
    return selectActiveSegment(state);
  });
  const { sixSignalInfo } = useSelector((state) => state.featureConfig);
  const { newSegmentMode, filtersDirty: areFiltersDirty } = useSelector(
    (state) => state.accountProfilesView
  );
  const groupsList = useSelector((state) => selectGroupsList(state));
  const agentState = useSelector((state) => state.agent);

  const [currentPage, setCurrentPage] = useState(1);
  const [currentPageSize, setCurrentPageSize] = useState(25);
  const [searchBarOpen, setSearchBarOpen] = useState(false);
  const [searchDDOpen, setSearchDDOpen] = useState(false);
  const [listSearchItems, setListSearchItems] = useState([]);
  const [listProperties, setListProperties] = useState([]);
  const [showPopOver, setShowPopOver] = useState(false);
  const [checkListAccountProps, setCheckListAccountProps] = useState([]);
  const [tlConfig, setTLConfig] = useState(DEFAULT_TIMELINE_CONFIG);
  const [companyValueOpts, setCompanyValueOpts] = useState({ All: {} });
  const [isUpgradeModalVisible, setIsUpgradeModalVisible] = useState(false);

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

  const { isFeatureLocked: isScoringLocked } = useFeatureLock(
    FEATURES.FEATURE_ACCOUNT_SCORING
  );

  const activeAgent = agentState?.agent_details?.email;

  useEffect(() => {
    if (segment_id && segments['$domains']) {
      if (segment_id !== activeSegment.id) {
        const selectedSegment = segments['$domains']?.find(
          (seg) => seg.id === segment_id
        );
        setActiveSegment(selectedSegment);
      }
    }
  }, [segment_id, segments]);

  useEffect(() => {
    let listProps = [];
    if (IsDomainGroup(accountPayload?.source)) {
      listProps = Object.keys(groups?.account_groups || {}).reduce(
        (properties, group) => {
          return groupProperties[group]
            ? properties.concat(groupProperties[group])
            : properties;
        },
        []
      );
    } else {
      listProps = groupProperties?.[accountPayload?.source] || [];
    }
    setListProperties(listProps);
  }, [groupProperties, accountPayload?.source, groups]);

  useEffect(() => {
    const tableProps = accountPayload?.segment_id
      ? activeSegment?.query?.table_props
      : currentProjectSettings.timelines_config?.account_config?.table_props ||
        [];
    const accountPropsWithEnableKey = formatUserPropertiesToCheckList(
      listProperties,
      tableProps?.filter(
        (entry) => entry !== '' && entry !== undefined && entry !== null
      )
    );
    setCheckListAccountProps(accountPropsWithEnableKey);
  }, [currentProjectSettings, listProperties, activeSegment, accountPayload]);

  useEffect(() => {
    fetchGroupProperties(GROUP_NAME_DOMAINS);
    Object.keys(groups?.account_groups || {}).forEach((group) => {
      fetchGroupProperties(group);
    });
  }, [activeProject.id, groups]);

  useEffect(() => {
    if (!accountPayload.source) {
      const source = GROUP_NAME_DOMAINS;
      updateAccountPayload({ source });
    }
  }, [accountPayload.source, activeProject.id]);

  useEffect(() => {
    if (!currentProjectSettings?.timelines_config) return;

    const { disabled_events, user_config, account_config } =
      currentProjectSettings.timelines_config;
    const timelinesConfig = {
      disabled_events: [...disabled_events],
      user_config: { ...DEFAULT_TIMELINE_CONFIG.user_config, ...user_config },
      account_config: {
        ...DEFAULT_TIMELINE_CONFIG.account_config,
        ...account_config
      }
    };
    setTLConfig(timelinesConfig);
  }, [currentProjectSettings?.timelines_config]);

  useEffect(() => {
    if (!accountPayload.search_filter) {
      setListSearchItems([]);
    } else {
      const listValues = accountPayload?.search_filter || [];
      setListSearchItems(uniq(listValues));
      setSearchBarOpen(true);
    }
  }, [accountPayload?.search_filter]);

  useEffect(() => {
    fetchProjectSettings(activeProject?.id).then(() => {
      if (!groups || Object.keys(groups).length === 0) {
        getGroups(activeProject?.id);
      }
      getSavedSegments(activeProject?.id);
    });
  }, [activeProject?.id, groups]);

  useEffect(() => {
    if (newSegmentMode === true) {
      restoreFiltersDefaultState();
    }
  }, [newSegmentMode]);

  useEffect(() => {
    if (newSegmentMode === false) {
      if (Boolean(activeSegment?.id) === true && activeSegment.query != null) {
        const filters = getSelectedFiltersFromQuery({
          query: activeSegment.query,
          groupsList
        });
        setAppliedFilters(cloneDeep(filters));
        setSelectedFilters(filters);
        setFiltersExpanded(false);
        setFiltersDirty(false);
      } else {
        const selectedGroup = groupsList.find(
          (g) => g[1] === accountPayload.source
        );
        restoreFiltersDefaultState(selectedGroup);
      }
    }
  }, [accountPayload, activeSegment, groupsList, newSegmentMode]);

  const setActiveSegment = useCallback(
    (segmentPayload) => {
      // history.replace(PathUrls.ProfileAccountsSegmentsURL + '/' + segmentPayload.id);
      dispatch(setActiveSegmentAction(segmentPayload));
    },
    [dispatch]
  );

  const setFiltersDirty = useCallback(
    (value) => {
      dispatch(setFiltersDirtyAction(value));
    },
    [dispatch]
  );

  const disableNewSegmentMode = useCallback(() => {
    dispatch(setNewSegmentModeAction(false));
  }, [dispatch]);

  const handleDeleteActiveSegment = useCallback(() => {
    deleteSegment({
      projectId: activeProject.id,
      segmentId: accountPayload.segment_id
    })
      .then(() => {
        setMoreActionsModalMode(null);
        notification.success({
          message: 'Segment deleted successfully',
          duration: 5
        });
      })
      .finally(() => history.replace(PathUrls.ProfileAccounts));
  }, [accountPayload.segment_id, activeProject.id, deleteSegment]);

  const displayTableProps = useMemo(() => {
    const filterNullEntries = (entry) =>
      entry !== '' && entry !== undefined && entry !== null;

    const getFilteredTableProps = (tableProps) => {
      return tableProps?.filter(filterNullEntries) || [];
    };

    const segmentTableProps = activeSegment?.query?.table_props;
    const projectTableProps =
      currentProjectSettings?.timelines_config?.account_config?.table_props;

    const tableProps = accountPayload.segment_id
      ? getFilteredTableProps(segmentTableProps)
      : getFilteredTableProps(projectTableProps);

    return tableProps;
  }, [currentProjectSettings, accountPayload, activeSegment]);

  const handleRenameSegment = useCallback(
    (name) => {
      updateSegmentForId(activeProject.id, accountPayload.segment_id, {
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
    [activeProject.id, accountPayload.segment_id, activeSegment]
  );

  const handleUpdateSegmentDefinition = useCallback(() => {
    const reqPayload = getFiltersRequestPayload({
      selectedFilters,
      table_props: displayTableProps
    });
    updateSegmentForId(
      activeProject.id,
      accountPayload.segment_id,
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
    accountPayload.segment_id,
    getSavedSegments,
    setFiltersDirty
  ]);

  const setAccountPayload = useCallback(
    (payload) => {
      dispatch(setAccountPayloadAction(payload));
    },
    [dispatch]
  );

  const updateAccountPayload = useCallback(
    (payload) => {
      dispatch(updateAccountPayloadAction(payload));
    },
    [dispatch]
  );

  const fetchGroupProperties = useCallback(
    async (groupId) => {
      if (!groupProperties[groupId]) {
        await getGroupProperties(activeProject.id, groupId);
      }
    },
    [activeProject.id, getGroupProperties, groupProperties]
  );

  const getAccounts = useCallback(
    (payload) => {
      const shouldCache = location.state?.fromDetails;
      if (payload.source && payload.source !== '' && !shouldCache) {
        setDefaultSorterInfo({ key: 'engagement', order: 'descend' });
        const formatPayload = { ...payload };
        formatPayload.filters =
          formatFiltersForPayload(payload?.filters, 'accounts') || [];
        const reqPayload = formatReqPayload(formatPayload, activeSegment);
        getProfileAccounts(activeProject.id, reqPayload, activeAgent).then(
          (response) => {
            if (response.type === 'FETCH_PROFILE_ACCOUNTS_FAILED') {
              if (response.error.status === 400) {
                setErrMsg('400 Bad Request');
              } else if (response.error.status === 500) {
                setErrMsg(
                  'The server encountered an internal error and could not complete your request'
                );
              }
            }

            if (response.type === 'FETCH_PROFILE_ACCOUNTS_FULFILLED') {
              if (response.status === 200) {
                if (response.payload.length === 0) {
                  setErrMsg('No accounts Found');
                }
              }
            }
          }
        );
      }
      if (shouldCache) {
        setCurrentPage(location.state.currentPage);
        setCurrentPageSize(location.state.currentPageSize);
        setDefaultSorterInfo(location.state.activeSorter);
        const localeState = { ...history.location.state, fromDetails: false };
        history.replace('/' + location.hash, { state: localeState });
      }
    },
    [
      location.state?.fromDetails,
      location.state?.currentPage,
      location.state?.activeSorter,
      activeSegment,
      activeProject.id,
      activeAgent,
      history
    ]
  );

  const tableData = useMemo(() => {
    const sortedData = accounts?.data?.sort(
      (a, b) => new Date(b.last_activity) - new Date(a.last_activity)
    );
    return sortedData?.map((row) => ({
      ...row,
      ...row?.tableProps
    }));
  }, [accounts]);

  const applyFilters = useCallback(() => {
    setAppliedFilters(cloneDeep(selectedFilters));
    setFiltersExpanded(false);
    setFiltersDirty(true);
    const reqPayload = getFiltersRequestPayload({
      selectedFilters,
      table_props: displayTableProps
    });
    getProfileAccounts(activeProject.id, reqPayload, activeAgent);
  }, [
    selectedFilters,
    displayTableProps,
    getProfileAccounts,
    activeProject.id,
    activeAgent,
    setFiltersDirty
  ]);

  const handlePropChange = (option) => {
    if (
      option.enabled ||
      checkListAccountProps.filter((item) => item.enabled === true).length < 8
    ) {
      const checkListProps = [...checkListAccountProps];
      const optIndex = checkListProps.findIndex(
        (obj) => obj.prop_name === option.prop_name
      );
      checkListProps[optIndex].enabled = !checkListProps[optIndex].enabled;
      setCheckListAccountProps(checkListProps);
    } else {
      notification.error({
        message: 'Error',
        description: 'Maximum Table Properties Selection Reached.',
        duration: 2
      });
    }
  };

  const applyTableProps = () => {
    if (accountPayload?.segment_id?.length) {
      const newTableProps =
        checkListAccountProps
          ?.filter(({ enabled }) => enabled)
          ?.map(({ prop_name }) => prop_name)
          ?.filter((entry) => entry !== '' && entry !== undefined) || [];

      const updatedQuery = {
        ...activeSegment.query,
        table_props: newTableProps
      };

      const queryForFetch = getFiltersRequestPayload({
        selectedFilters: appliedFilters,
        table_props: newTableProps
      });

      updateSegmentForId(activeProject.id, accountPayload.segment_id, {
        query: updatedQuery
      })
        .then(() => getSavedSegments(activeProject.id))
        .then(() => setActiveSegment({ ...activeSegment, query: updatedQuery }))
        .finally(() =>
          getProfileAccounts(activeProject.id, queryForFetch, activeAgent)
        );
    } else {
      const filteredProps = !IsDomainGroup(accountPayload.source)
        ? tlConfig.account_config.table_props
            ?.filter(
              (entry) => entry !== '' && entry !== undefined && entry !== null
            )
            .filter(
              (item) =>
                !checkListAccountProps.some(
                  ({ prop_name }) => prop_name === item
                )
            )
        : [];
      const enabledProps = checkListAccountProps
        .filter(({ enabled }) => enabled)
        .map(({ prop_name }) => prop_name);

      const updatedConfig = {
        ...tlConfig,
        account_config: {
          ...tlConfig.account_config,
          table_props: [...filteredProps, ...enabledProps]
        }
      };

      const queryForFetch = getFiltersRequestPayload({
        selectedFilters: appliedFilters,
        table_props: [...filteredProps, ...enabledProps]
      });

      udpateProjectSettings(activeProject.id, {
        timelines_config: updatedConfig
      }).then(() =>
        getProfileAccounts(activeProject.id, queryForFetch, activeAgent)
      );
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
        onClick={navigateToAccountsEngagement}
        className='flex cursor-pointer col-gap-4 items-center py-2 px-4 hover:bg-gray-100'
      >
        <SVG size={20} name='fireFlameCurved' color='#8c8c8c' />
        <Text type='title' color='character-primary' extraClass='mb-0'>
          Account engagement rules
        </Text>
      </div>
    );

    if (Boolean(accountPayload.segment_id) === false) {
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

  const restoreFiltersDefaultState = useCallback(
    (selectedAccount = INITIAL_FILTERS_STATE.account) => {
      const initialFiltersStateWithSelectedAccount = {
        ...INITIAL_FILTERS_STATE,
        account: selectedAccount
      };
      setSelectedFilters(initialFiltersStateWithSelectedAccount);
      setAppliedFilters(cloneDeep(initialFiltersStateWithSelectedAccount));
      setFiltersExpanded(false);
      setFiltersDirty(false);
    },
    [setFiltersDirty]
  );

  const setFiltersList = useCallback((filters) => {
    setSelectedFilters((curr) => {
      return {
        ...curr,
        filters
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

  const handleSaveSegmentClick = useCallback(() => {
    if (newSegmentMode === true) {
      setSaveSegmentModal(true);
      return;
    }
    if (Boolean(accountPayload.segment_id) === true) {
      setUpdateSegmentModal(true);
    } else {
      setSaveSegmentModal(true);
    }
  }, [accountPayload.segment_id, newSegmentMode]);

  const resetSelectedFilters = useCallback(() => {
    setSelectedFilters(appliedFilters);
  }, [appliedFilters]);

  const handleClearFilters = useCallback(() => {
    restoreFiltersDefaultState();
    const reqPayload = getFiltersRequestPayload({
      selectedFilters: INITIAL_FILTERS_STATE,
      table_props: displayTableProps
    });
    getProfileAccounts(activeProject.id, reqPayload, activeAgent);
  }, [
    activeAgent,
    activeProject.id,
    displayTableProps,
    getProfileAccounts,
    restoreFiltersDefaultState
  ]);

  const setSelectedAccount = useCallback((account) => {
    setSelectedFilters((current) => {
      return {
        ...current,
        account
      };
    });
  }, []);

  const selectedAccount = useMemo(() => {
    return { account: selectedFilters.account };
  }, [selectedFilters.account]);

  const availableGroups = useMemo(() => {
    return Object.keys(groups?.account_groups || {});
  }, [groups]);

  const disableDiscardButton = useMemo(() => {
    return isEqual(selectedFilters, appliedFilters);
  }, [selectedFilters, appliedFilters]);

  const renderPropertyFilter = () => {
    return (
      <PropertyFilter
        profileType='account'
        source={accountPayload.source}
        filters={accountPayload.filters}
        filtersExpanded={filtersExpanded}
        filtersList={selectedFilters.filters}
        appliedFilters={appliedFilters}
        selectedAccount={selectedAccount}
        listEvents={selectedFilters.eventsList}
        availableGroups={availableGroups}
        eventProp={selectedFilters.eventProp}
        areFiltersDirty={areFiltersDirty}
        disableDiscardButton={disableDiscardButton}
        isActiveSegment={Boolean(accountPayload.segment_id) === true}
        applyFilters={applyFilters}
        setFiltersExpanded={setFiltersExpanded}
        setSaveSegmentModal={handleSaveSegmentClick}
        setFiltersList={setFiltersList}
        setListEvents={setListEvents}
        setEventProp={setEventProp}
        resetSelectedFilters={resetSelectedFilters}
        onClearFilters={handleClearFilters}
        setSelectedAccount={setSelectedAccount}
      />
    );
  };

  useEffect(() => {
    fetchData();
  }, [activeProject.id, groups]);

  const fetchData = async () => {
    const newCompanyValues = { $domains: {} };
    for (const [group, prop] of Object.entries(groupToDomainMap)) {
      if (groups && groups?.account_groups && groups?.account_groups[group]) {
        try {
          const res = await fetchGroupPropertyValues(
            activeProject.id,
            group,
            prop
          );
          newCompanyValues[GROUP_NAME_DOMAINS] = {
            ...newCompanyValues[GROUP_NAME_DOMAINS],
            ...res.data
          };
        } catch (err) {
          console.log(err);
        }
      }
    }
    setCompanyValueOpts(newCompanyValues);
  };

  const onApplyClick = (values) => {
    const updatedPayload = {
      ...accountPayload,
      search_filter: values.map((vl) => JSON.parse(vl)[0])
    };
    setListSearchItems(updatedPayload.search_filter);
    setAccountPayload(updatedPayload);
    setActiveSegment(activeSegment);
    getAccounts(updatedPayload);
  };

  const onSearchClose = () => {
    setSearchBarOpen(false);
    setSearchDDOpen(false);
    if (accountPayload?.search_filter?.length !== 0) {
      const updatedPayload = { ...accountPayload };
      updatedPayload.search_filter = [];
      setAccountPayload(updatedPayload);
      setListSearchItems([]);
      setActiveSegment(activeSegment);
      getAccounts(updatedPayload);
    }
  };

  const onSearchOpen = () => {
    setSearchBarOpen(true);
    setSearchDDOpen(true);
  };

  const searchCompanies = () => (
    <div className='absolute top-0'>
      {searchDDOpen ? (
        <FaSelect
          placeholder='Search Accounts'
          multiSelect
          options={
            companyValueOpts?.[accountPayload?.source]
              ? Object.keys(companyValueOpts[accountPayload?.source]).map(
                  (value) => [value]
                )
              : []
          }
          displayNames={companyValueOpts?.[accountPayload?.source]}
          applClick={(val) => onApplyClick(val)}
          onClickOutside={() => setSearchDDOpen(false)}
          selectedOpts={listSearchItems}
          style={{
            top: '-8px',
            right: 0,
            padding: '8px 8px 12px',
            overflowX: 'hidden'
          }}
          allowSearch
          posRight
        />
      ) : null}
    </div>
  );

  const { saveButtonDisabled } = useMemo(() => {
    return checkFiltersEquality({
      appliedFilters,
      newSegmentMode,
      filtersList: selectedFilters.filters,
      eventProp: selectedFilters.eventProp,
      eventsList: selectedFilters.eventsList,
      isActiveSegment: Boolean(accountPayload.segment_id),
      areFiltersDirty
    });
  }, [
    accountPayload.segment_id,
    appliedFilters,
    areFiltersDirty,
    newSegmentMode,
    selectedFilters.eventProp,
    selectedFilters.eventsList,
    selectedFilters.filters
  ]);

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

  const renderSearchSection = () => (
    <div className='relative'>
      <ControlledComponent controller={searchBarOpen}>
        <div className={'flex items-center justify-between'}>
          {!searchDDOpen && (
            <Input
              size='large'
              value={listSearchItems ? listSearchItems.join(', ') : null}
              placeholder={'Search Accounts'}
              style={{
                width: '240px',
                'border-radius': '5px'
              }}
              prefix={<SVG name='search' size={16} color={'#8C8C8C'} />}
              onClick={() => setSearchDDOpen(true)}
            />
          )}
          <Button type='text' onClick={onSearchClose}>
            <SVG name={'close'} size={24} color={'grey'} />
          </Button>
        </div>
      </ControlledComponent>
      <ControlledComponent controller={!searchBarOpen}>
        <Button type='text' onClick={onSearchOpen}>
          <SVG name={'search'} size={24} color={'#8c8c8c'} />
        </Button>
      </ControlledComponent>
      {searchCompanies()}
    </div>
  );

  const renderDownloadSection = () => {
    return (
      <Button onClick={() => setShowDownloadCSVModal(true)} type='text'>
        <SVG size={24} name={'download'} color={'#8c8c8c'} />
      </Button>
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
        <Button type='default'>
          <SVG size={24} name={'more'} />
        </Button>
      </Popover>
    );
  };

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
          <SVG size={24} name={'tableColumns'} />
        </Button>
      </Popover>
    );
  };

  const handleTableChange = (pageParams, somedata, sorter) => {
    setCurrentPage(pageParams.current);
    setCurrentPageSize(pageParams.pageSize);
    setDefaultSorterInfo({ key: sorter.columnKey, order: sorter.order });
  };
  const [newTableColumns, setNewTableColumns] = useState([]);
  const [columnsType, setColumnTypes] = useState({});

  useEffect(() => {
    setNewTableColumns(
      getColumns({
        accounts,
        source: accountPayload?.source,
        isScoringLocked,
        displayTableProps,
        groupPropNames,
        listProperties,
        defaultSorterInfo
      })
    );
  }, [
    accounts,
    accountPayload?.source,
    displayTableProps,
    groupPropNames,
    isScoringLocked,
    listProperties,
    defaultSorterInfo
  ]);
  // const tableColumns = useMemo(() => {
  //   return getColumns({
  //     accounts,
  //     source: accountPayload?.source,
  //     isScoringLocked,
  //     displayTableProps,
  //     groupPropNames,
  //     listProperties,
  //     defaultSorterInfo
  //   });
  // }, [
  //   accounts,
  //   accountPayload?.source,
  //   displayTableProps,
  //   groupPropNames,
  //   isScoringLocked,
  //   listProperties,
  //   defaultSorterInfo
  // ]);
  useEffect(() => {
    let from = location.state?.state?.accountsTableRow;
    if (from && document.getElementById(from)) {
      const element = document.getElementById(from);
      const y = element?.getBoundingClientRect().top + window.scrollY - 100;

      window.scrollTo({ top: y, behavior: 'smooth' });

      location.state.state.accountsTableRow = '';
      // document.getElementById(location.hash.split('#')[1])?.scrollIntoView();
      // window.scrollBy(0, -150);
    }
  }, [newTableColumns]);
  const renderTable = useCallback(() => {
    const handleResize =
      (index) =>
      (_, { size }) => {
        const tmpColType = newTableColumns[index]?.type;
        const tmpColWidthRange =
          COLUMN_TYPE_PROPS[tmpColType ? tmpColType : 'string'];
        const newColumns = [...newTableColumns];
        newColumns[index] = {
          ...newColumns[index],
          width: (() => {
            if (size.width < tmpColWidthRange.min) return tmpColWidthRange.min;
            else if (size.width > tmpColWidthRange.max)
              return tmpColWidthRange.max;
            return size.width;
          })()
        };
        setNewTableColumns(newColumns);
      };

    const mergeColumns = newTableColumns.map((col, index) => ({
      ...col,
      onHeaderCell: (column) => ({
        width: column.width,
        onResize: handleResize(index)
      })
    }));

    return (
      <div>
        <Table
          components={{
            header: {
              cell: ResizableTitle
            }
          }}
          onRow={(account) => ({
            onClick: () => {
              history.push(
                `/profiles/accounts/${btoa(account.identity)}?group=${
                  activeSegment?.type
                    ? activeSegment.type
                    : accountPayload.source
                }&view=birdview`,
                {
                  accountPayload: accountPayload,
                  activeSegment: activeSegment,
                  fromDetails: true,
                  currentPage: currentPage,
                  currentPageSize: currentPageSize,
                  activeSorter: defaultSorterInfo,
                  accountsTableRow: account.name
                }
              );
            }
          })}
          className='fa-table--userlist'
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
            x: displayTableProps?.length * 300
          }}
        />
      </div>
    );
  }, [tableData, newTableColumns]);

  const showRangeNudge = useMemo(() => {
    return showUpgradeNudge(
      sixSignalInfo?.usage || 0,
      sixSignalInfo?.limit || 0,
      currentProjectSettings
    );
  }, [currentProjectSettings, sixSignalInfo?.limit, sixSignalInfo?.usage]);

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

  const handleCreateSegment = useCallback(
    (newSegmentName) => {
      const reqPayload = getFiltersRequestPayload({
        selectedFilters,
        table_props: displayTableProps
      });
      reqPayload.name = newSegmentName;
      reqPayload.type = selectedFilters.account[1];
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

  const generateCSVData = useCallback(
    (data, selectedOptions) => {
      const csvRows = [];

      const headers = selectedOptions.map(
        (prop_name) =>
          checkListAccountProps.find((elem) => elem.prop_name === prop_name)
            .display_name
      );
      headers.unshift('Name', 'Engagement category', 'Engagement score');
      csvRows.push(headers.join(','));

      data.forEach((d) => {
        const values = selectedOptions.map((elem) => {
          return d.table_props[elem] != null ? `"${d.table_props[elem]}"` : '-';
        });
        values.unshift(
          d.name,
          d.engagement != null ? d.engagement : '-',
          d.score != null ? formatCount(d.score) : '-'
        );
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
          source: selectedAccount.account[1],
          selectedFilters: appliedFilters,
          table_props: selectedOptions
        });
        const resultAccounts = await fetchProfileAccounts(
          activeProject.id,
          reqPayload,
          activeAgent
        );
        const csvData = generateCSVData(resultAccounts.data, selectedOptions);
        downloadCSV(csvData, 'accounts.csv');
        setCSVDataLoading(false);
        setShowDownloadCSVModal(false);
      } catch (err) {
        console.log(err);
        setCSVDataLoading(false);
        notification.error({
          message: 'Error',
          description: 'CSV download failed',
          duration: 2
        });
      }
    },
    [
      activeAgent,
      activeProject.id,
      appliedFilters,
      generateCSVData,
      selectedAccount.account
    ]
  );

  const closeDownloadCSVModal = useCallback(() => {
    setShowDownloadCSVModal(false);
  }, []);

  const pageTitle = useMemo(() => {
    if (newSegmentMode === true) {
      return 'Untitled Segment 1';
    }
    if (Boolean(activeSegment?.id) === false) {
      const source = accountPayload.source;
      const title = get(
        groupsList.find((elem) => elem[1] === source),
        0,
        'All Accounts'
      );
      return title;
    }
    return activeSegment.name;
  }, [accountPayload, groupsList, activeSegment, newSegmentMode]);

  useEffect(() => {
    if (newSegmentMode === false) {
      getAccounts(accountPayload);
    }
  }, [newSegmentMode, accountPayload, activeSegment]);

  const titleIcon = useMemo(() => {
    if (Boolean(activeSegment?.id) === true) {
      return defaultSegmentIconsMapping[activeSegment?.name]
        ? defaultSegmentIconsMapping[activeSegment?.name]
        : 'pieChart';
    }
    return 'buildings';
  }, [activeSegment]);

  const titleIconColor = useMemo(() => {
    return getSegmentColorCode(activeSegment?.name ?? '');
  }, [activeSegment?.name]);

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
            {renderDownloadSection()}
            {renderTablePropsSelect()}
            {renderMoreActions()}
          </ControlledComponent>
        </div>
      </div>
      <ControlledComponent controller={accounts.isLoading === true}>
        <Spin size='large' className='fa-page-loader' />
      </ControlledComponent>
      <ControlledComponent
        controller={
          accounts.isLoading === false &&
          accounts.data?.length > 0 &&
          (newSegmentMode === false || areFiltersDirty === true)
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
          accounts.isLoading === false &&
          accounts.data.length === 0 &&
          (newSegmentMode === false || areFiltersDirty === true)
        }
      >
        <NoDataWithMessage
          message={
            isOnboarded(currentProjectSettings)
              ? accounts?.data?.length === 0
                ? 'No Accounts found'
                : errMsg
              : 'Onboarding not completed'
          }
        />
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
      <DownloadCSVModal
        visible={showDownloadCSVModal}
        onCancel={closeDownloadCSVModal}
        options={checkListAccountProps}
        displayTableProps={displayTableProps}
        onSubmit={handleDownloadCSV}
        isLoading={csvDataLoading}
      />
    </ProfilesWrapper>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  groups: state.coreQuery.groups,
  accounts: state.timelines.accounts,
  segments: state.timelines.segments,
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
      deleteSegment
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(AccountProfiles);
