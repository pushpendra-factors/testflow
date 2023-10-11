import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { Table, Button, Spin, Popover, Tabs, notification, Input } from 'antd';
import { connect, useDispatch, useSelector } from 'react-redux';
import { bindActionCreators } from 'redux';
import { Text, SVG } from '../../factorsComponents';
import PropertyFilter from './PropertyFilter';
import { getGroupProperties } from '../../../reducers/coreQuery/middleware';
import FaSelect from '../../FaSelect';
import {
  DEFAULT_TIMELINE_CONFIG,
  formatFiltersForPayload,
  formatReqPayload,
  getFiltersRequestPayload,
  getSelectedFiltersFromQuery
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
import { useHistory, useLocation, useParams } from 'react-router-dom';
import {
  fetchGroupPropertyValues,
  fetchGroups
} from 'Reducers/coreQuery/services';
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
import MoreActionsDropdown from './MoreActionsDropdown';
import DeleteSegmentModal from './DeleteSegmentModal';
import RenameSegmentModal from './RenameSegmentModal';
import {
  INITIAL_FILTERS_STATE,
  moreActionsMode
} from './accountProfiles.constants';
import { selectGroupsList } from 'Reducers/groups/selectors';
import UpdateSegmentModal from './UpdateSegmentModal';
import { AccountsSidebarIconsMapping } from 'Views/AppSidebar/appSidebar.constants';
import DownloadCSVModal from './DownloadCSVModal';
import { fetchProfileAccounts } from 'Reducers/timelines';
import { selectSegments } from 'Reducers/timelines/selectors';
import { downloadCSV } from 'Utils/csv';
import { formatCount } from 'Utils/dataFormatter';
import { PathUrls } from 'Routes/pathUrls';

const groupToCompanyPropMap = {
  $hubspot_company: '$hubspot_company_name',
  $salesforce_account: '$salesforce_account_name',
  $6signal: '$6Signal_name',
  $linkedin_company: '$li_localized_name',
  $g2: '$g2_name'
};

const groupToDomainMap = {
  $hubspot_company: '$hubspot_company_domain',
  $salesforce_account: '$salesforce_account_website',
  $6signal: '$6Signal_domain',
  $linkedin_company: '$li_domain',
  $g2: '$g2_domain'
};

function AccountProfiles({
  activeProject,
  groupOpts,
  accounts,
  createNewSegment,
  getSavedSegments,
  fetchProjectSettings,
  fetchGroups,
  udpateProjectSettings,
  currentProjectSettings,
  getProfileAccounts,
  getGroupProperties,
  updateSegmentForId,
  deleteSegment
}) {
  const dispatch = useDispatch();
  const history = useHistory();
  const location = useLocation();
  const {segment_id} = useParams();

  const {segments} = useSelector((state) => selectSegments(state));

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

  const { isFeatureLocked: isScoringLocked } = useFeatureLock(
    FEATURES.FEATURE_ACCOUNT_SCORING
  );

  const activeAgent = agentState?.agent_details?.email;

  useEffect(()=> {
    if(segment_id && segments?.length) {
      if(segment_id !== activeSegment.id) {
        const selectedSegment = segments.find((seg) => seg.id === segment_id);
        setActiveSegment(selectedSegment);
      }
    }

  }, [segment_id, segments])

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
    }).then((response) => {
      setMoreActionsModalMode(null);
      notification.success({
        message: 'Segment deleted successfully',
        duration: 5
      });
    });
  }, [accountPayload.segment_id, activeProject.id, deleteSegment]);

  const displayTableProps = useMemo(() => {
    const filterPropsMap = {
      $hubspot_company: 'hubspot',
      $salesforce_account: 'salesforce',
      $6signal: '6Signal',
      $linkedin_company: '$li_',
      $g2: '$g2',
      All: ''
    };
    const source = filterPropsMap[accountPayload?.source];
    const tableProps = accountPayload.segment_id
      ? activeSegment?.query?.table_props?.filter((item) =>
          item.includes(source)
        )
      : currentProjectSettings?.timelines_config?.account_config?.table_props?.filter(
          (item) => item.includes(source)
        );
    return (
      tableProps?.filter((entry) => entry !== '' && entry !== undefined) || []
    );
  }, [currentProjectSettings, accountPayload, activeSegment]);

  const handleRenameSegment = useCallback(
    (name) => {
      updateSegmentForId(activeProject.id, accountPayload.segment_id, {
        name
      }).then((respnse) => {
        getSavedSegments(activeProject.id);
        setActiveSegment({ ...activeSegment, name });
        setMoreActionsModalMode(null);
        notification.success({
          message: 'Segment renamed successfully',
          duration: 5
        });
      });
    },
    [
      updateSegmentForId,
      activeProject.id,
      accountPayload.segment_id,
      getSavedSegments,
      setActiveSegment,
      activeSegment
    ]
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
    ).then((respnse) => {
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
        getProfileAccounts(activeProject.id, reqPayload, activeAgent);
      }
      if (shouldCache) {
        setCurrentPage(location.state.currentPage);
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
      getProfileAccounts,
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
    setAppliedFilters(selectedFilters);
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
      const filteredProps =
        accountPayload.source !== 'All'
          ? tlConfig.account_config.table_props.filter(
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

  const restoreFiltersDefaultState = useCallback(
    (selectedAccount = INITIAL_FILTERS_STATE.account) => {
      const initialFiltersStateWithSelectedAccount = {
        ...INITIAL_FILTERS_STATE,
        account: selectedAccount
      };
      setSelectedFilters(initialFiltersStateWithSelectedAccount);
      setAppliedFilters(initialFiltersStateWithSelectedAccount);
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
    return Object.keys(groupOpts || {});
  }, [groupOpts]);

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
      />
    );
  };

  useEffect(() => {
    const fetchData = async () => {
      const newCompanyValues = { All: {} };
      for (const [group, prop] of Object.entries(groupToCompanyPropMap)) {
        if (groupOpts[group]) {
          try {
            const res = await fetchGroupPropertyValues(
              activeProject.id,
              group,
              prop
            );
            newCompanyValues[group] = { ...res.data };
          } catch (err) {
            console.log(err);
          }
        }
      }
      for (const [group, prop] of Object.entries(groupToDomainMap)) {
        if (groupOpts[group]) {
          try {
            const res = await fetchGroupPropertyValues(
              activeProject.id,
              group,
              prop
            );
            newCompanyValues['All'] = {
              ...newCompanyValues['All'],
              ...res.data
            };
          } catch (err) {
            console.log(err);
          }
        }
      }
      setCompanyValueOpts(newCompanyValues);
    };
    fetchData();
  }, [activeProject.id, groupOpts]);

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
                placeholder={'Search Accounts'}
                style={{
                  width: '240px',
                  'border-radius': '5px'
                }}
                prefix={<SVG name='search' size={16} color={'grey'} />}
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
        {searchCompanies()}
      </div>
    </ControlledComponent>
  );

  const renderTablePropsSelect = () => {
    return (
      <ControlledComponent
        controller={filtersExpanded === false && newSegmentMode === false}
      >
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
          <Button
            size='large'
            icon={<SVG name='activity_filter' />}
            className='relative'
          >
            Edit Columns
          </Button>
        </Popover>
      </ControlledComponent>
    );
  };

  const handleTableChange = (pageParams, _, sorter) => {
    setCurrentPage(pageParams.current);
    setDefaultSorterInfo({ key: sorter.columnKey, order: sorter.order });
  };

  const tableColumns = useMemo(() => {
    return getColumns({
      accounts,
      source: accountPayload?.source,
      isScoringLocked,
      displayTableProps,
      groupPropNames,
      listProperties,
      defaultSorterInfo
    });
  }, [
    accounts,
    accountPayload?.source,
    displayTableProps,
    groupPropNames,
    isScoringLocked,
    listProperties,
    defaultSorterInfo
  ]);

  const renderTable = useCallback(() => {
    return (
      <div>
        <Table
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
                  activeSorter: defaultSorterInfo
                }
              );
            }
          })}
          className='fa-table--userlist'
          dataSource={tableData}
          columns={tableColumns}
          rowClassName='cursor-pointer'
          pagination={{
            position: ['bottom', 'left'],
            defaultPageSize: '25',
            current: currentPage
          }}
          onChange={handleTableChange}
          scroll={{
            x: displayTableProps?.length * 300,
            y:"calc(100vh - 320px)"
          }}
          footer={() => (
            <div className='text-right'>
              <a
                className='font-size--small'
                href='https://www.uplead.com'
                target='_blank'
                rel='noopener noreferrer'
              >
                Logos provided by UpLead
              </a>
            </div>
          )}
        />
      </div>
    );
  }, [tableData, tableColumns]);

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
        console.log(
          'jklogs ~ file: index.js:908 ~ resultAccounts.data:',
          resultAccounts.data
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
    if (Boolean(accountPayload.segment_id) === false) {
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
    const fetchData = async () => {
      const newCompanyValues = { All: {} };
      for (const [group, prop] of Object.entries(groupToCompanyPropMap)) {
        if (groupOpts[group]) {
          try {
            const res = await fetchGroupPropertyValues(
              activeProject.id,
              group,
              prop
            );
            newCompanyValues[group] = { ...res.data };
            newCompanyValues['All'] = {
              ...newCompanyValues['All'],
              ...res.data
            };
          } catch (err) {
            console.log(err);
          }
        }
      }
      setCompanyValueOpts(newCompanyValues);
    };
    fetchData();
  }, [activeProject.id, groupOpts]);

  useEffect(() => {
    if (newSegmentMode === false) {
      getAccounts(accountPayload);
    }
  }, [newSegmentMode, accountPayload]);

  useEffect(() => {
    let listProps = [];
    if (accountPayload?.source === 'All') {
      listProps = Object.keys(groupOpts || {}).reduce((acc, property) => {
        return groupProperties[property]
          ? acc.concat(groupProperties[property])
          : acc;
      }, []);
    } else {
      listProps = groupProperties?.[accountPayload?.source] || [];
    }
    setListProperties(listProps);
  }, [groupProperties, accountPayload?.source, groupOpts]);

  useEffect(() => {
    const tableProps = accountPayload?.segment_id
      ? activeSegment?.query?.table_props
      : currentProjectSettings.timelines_config?.account_config?.table_props;
    const accountPropsWithEnableKey = formatUserPropertiesToCheckList(
      listProperties,
      tableProps
    );
    setCheckListAccountProps(accountPropsWithEnableKey);
  }, [currentProjectSettings, listProperties, activeSegment, accountPayload]);

  useEffect(() => {
    fetchGroupProperties('$domains');
    Object.keys(groupOpts || {}).forEach((group) => {
      fetchGroupProperties(group);
    });
  }, [activeProject.id, fetchGroupProperties, groupOpts]);

  useEffect(() => {
    if (!accountPayload.source) {
      const source = groupsList?.[0]?.[1] || '';
      updateAccountPayload({ source });
    }
  }, [
    accountPayload.source,
    activeProject.id,
    groupsList,
    updateAccountPayload
  ]);

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
    fetchProjectSettings(activeProject?.id);
    fetchGroups(activeProject?.id, true);
    getSavedSegments(activeProject?.id);
  }, [activeProject?.id, fetchGroups, fetchProjectSettings, getSavedSegments]);

  useEffect(() => {
    if (newSegmentMode === true) {
      restoreFiltersDefaultState();
    }
  }, [newSegmentMode, restoreFiltersDefaultState]);

  useEffect(() => {
    if (newSegmentMode === false) {
      if (
        Boolean(accountPayload.segment_id) === true &&
        activeSegment.query != null
      ) {
        const filters = getSelectedFiltersFromQuery({
          query: activeSegment.query,
          groupsList
        });
        setAppliedFilters(filters);
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
  }, [
    accountPayload,
    activeSegment.query,
    groupsList,
    newSegmentMode,
    restoreFiltersDefaultState,
    setFiltersDirty
  ]);

  const titleIcon = useMemo(() => {
    if (Boolean(accountPayload.segment_id) === true) {
      return 'pieChart';
    }
    return AccountsSidebarIconsMapping[accountPayload.source] != null
      ? AccountsSidebarIconsMapping[accountPayload.source]
      : 'buildings';
  }, [accountPayload]);

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
            <SVG name={titleIcon} size={32} color='#FF4D4F' />
          </div>
          <Text type='title' level={3} weight='bold' extraClass='mb-0' id={'fa-at-text--page-title'}>
            {pageTitle}
          </Text>
        </div>
        <ControlledComponent controller={Boolean(accountPayload.segment_id)}>
          <MoreActionsDropdown
            onRename={() => setMoreActionsModalMode(moreActionsMode.RENAME)}
            onDelete={() => setMoreActionsModalMode(moreActionsMode.DELETE)}
          />
        </ControlledComponent>
      </div>

      <div className='flex justify-between items-center my-4'>
        <div className='flex items-center col-gap-2 w-full'>
          {renderPropertyFilter()}
          {renderSaveSegmentButton()}
        </div>
        <div className='inline-flex gap--6'>
          {/* {accountPayload?.filters?.length ? renderClearFilterButton() : null} */}
          {renderSearchSection()}
          {renderTablePropsSelect()}
          <ControlledComponent
            controller={filtersExpanded === false && newSegmentMode === false}
          >
            <div role='button' onClick={() => setShowDownloadCSVModal(true)}>
              <SVG name='download' />
            </div>
          </ControlledComponent>
        </div>
      </div>
      <ControlledComponent controller={accounts.isLoading === true}>
        <Spin size='large' className='fa-page-loader' />
      </ControlledComponent>
      <ControlledComponent
        controller={
          accounts.isLoading === false &&
          accounts.data.length > 0 &&
          (newSegmentMode === false || areFiltersDirty === true)
        }
      >
        <>{renderTable()}</>
      </ControlledComponent>
      <ControlledComponent
        controller={
          accounts.isLoading === false &&
          accounts.data.length === 0 &&
          (newSegmentMode === false || areFiltersDirty === true)
        }
      >
        <NoDataWithMessage message={'No Accounts Found'} />
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
        segmentName={activeSegment.name}
        visible={moreActionsModalMode === moreActionsMode.DELETE}
        onCancel={() => setMoreActionsModalMode(null)}
        onOk={handleDeleteActiveSegment}
      />
      <RenameSegmentModal
        segmentName={activeSegment.name}
        visible={moreActionsModalMode === moreActionsMode.RENAME}
        onCancel={() => setMoreActionsModalMode(null)}
        handleSubmit={handleRenameSegment}
      />

      <UpdateSegmentModal
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
  groupOpts: state.groups.data,
  accounts: state.timelines.accounts,
  currentProjectSettings: state.global.currentProjectSettings
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      fetchGroups,
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
