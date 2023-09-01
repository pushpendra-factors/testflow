import React, { useState, useEffect, useMemo, useCallback } from 'react';
import cx from 'classnames';
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
  getFiltersRequestPayload
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
import { useHistory, useLocation } from 'react-router-dom';
import {
  fetchGroupPropertyValues,
  fetchGroups
} from 'Reducers/coreQuery/services';
import NoDataWithMessage from '../MyComponents/NoDataWithMessage';
import ProfilesWrapper from '../ProfilesWrapper';
import { getColumns } from './accountProfiles.helpers';
import {
  selectAccountPayload,
  selectActiveSegment
} from 'Reducers/accountProfilesView/selectors';
import {
  setAccountPayloadAction,
  setActiveSegmentAction,
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
import styles from './index.module.scss';
import MoreActionsDropdown from './MoreActionsDropdown';
import DeleteSegmentModal from './DeleteSegmentModal';
import RenameSegmentModal from './RenameSegmentModal';
import { moreActionsMode } from './accountProfiles.constants';
import { selectGroupsList } from 'Reducers/groups/selectors';

const groupToCompanyPropMap = {
  $hubspot_company: '$hubspot_company_name',
  $salesforce_account: '$salesforce_account_name',
  $6signal: '$6Signal_name',
  $linkedin_company: '$li_localized_name',
  $g2: '$g2_name'
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
  const { newSegmentMode } = useSelector((state) => state.accountProfilesView);
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
  // accounts 2.0
  const [selectedAccount, setSelectedAccount] = useState({
    account: null
  });
  const [filtersExpanded, setFiltersExpanded] = useState(false);
  const [saveSegmentModal, setSaveSegmentModal] = useState(false);
  const [filtersList, setFiltersList] = useState([]);
  const [appliedFilters, setAppliedFilters] = useState({});
  const [listEvents, setListEvents] = useState([]);
  const [moreActionsModalMode, setMoreActionsModalMode] = useState(null); // DELETE | RENAME
  const [eventProp, setEventProp] = useState('any');

  const { isFeatureLocked: isEngagementLocked } = useFeatureLock(
    FEATURES.FEATURE_ENGAGEMENT
  );

  const activeAgent = agentState?.agent_details?.email;

  const setActiveSegment = useCallback(
    (segmentPayload) => {
      dispatch(setActiveSegmentAction(segmentPayload));
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

  useEffect(() => {
    fetchProjectSettings(activeProject?.id);
    fetchGroups(activeProject?.id, true);
    getSavedSegments(activeProject?.id);
  }, [activeProject?.id, fetchGroups, fetchProjectSettings, getSavedSegments]);

  useEffect(() => {
    if (groupsList.length > 0) {
      if (activeSegment.type != null) {
        const selectedGroup = groupsList.find(
          (g) => g[1] === activeSegment.type
        );
        setSelectedAccount((current) => {
          return {
            ...current,
            account: selectedGroup
          };
        });
      } else {
        const selectedGroup = groupsList.find(
          (g) => g[1] === accountPayload.source
        );
        setSelectedAccount((current) => {
          return {
            ...current,
            account: selectedGroup
          };
        });
      }
    }
  }, [groupsList, accountPayload.source, activeSegment.type]);

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

  useEffect(() => {
    setFiltersList([]);
    setAppliedFilters({});
    setFiltersExpanded(false);
  }, [accountPayload]);

  useEffect(() => {
    if (!accountPayload.search_filter) {
      setListSearchItems([]);
    } else {
      const listValues =
        accountPayload?.search_filter?.map((vl) => vl?.va) || [];
      setListSearchItems(uniq(listValues));
      setSearchBarOpen(true);
    }
  }, [accountPayload?.search_filter]);

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

  const fetchGroupProperties = useCallback(
    async (groupId) => {
      if (!groupProperties[groupId]) {
        await getGroupProperties(activeProject.id, groupId);
      }
    },
    [activeProject.id, getGroupProperties, groupProperties]
  );

  useEffect(() => {
    fetchGroupProperties('$domains');
    Object.keys(groupOpts || {}).forEach((group) => {
      fetchGroupProperties(group);
    });
  }, [activeProject.id, fetchGroupProperties, groupOpts]);

  const getAccounts = useCallback(
    (payload) => {
      const shouldCache = location.state?.fromDetails;
      if (payload.source && payload.source !== '' && !shouldCache) {
        const formatPayload = { ...payload };
        formatPayload.filters =
          formatFiltersForPayload(payload?.filters, 'accounts') || [];
        const reqPayload = formatReqPayload(formatPayload, activeSegment);
        getProfileAccounts(activeProject.id, reqPayload, activeAgent);
      }
      if (shouldCache) {
        setCurrentPage(location.state.currentPage);
        const localeState = { ...history.location.state, fromDetails: false };
        history.replace({ state: localeState });
      }
    },
    [
      activeAgent,
      activeProject.id,
      activeSegment,
      getProfileAccounts,
      history,
      location.state?.currentPage,
      location.state?.fromDetails
    ]
  );

  useEffect(() => {
    getAccounts(accountPayload);
  }, [
    accountPayload.source,
    accountPayload.segment_id,
    getAccounts,
    accountPayload
  ]);

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

  const getTableData = (data) => {
    const sortedData = data.sort(
      (a, b) => new Date(b.last_activity) - new Date(a.last_activity)
    );
    return sortedData.map((row) => ({
      ...row,
      ...row?.tableProps
    }));
  };

  const applyFilters = useCallback(() => {
    const opts = {
      source: selectedAccount.account[1],
      filters: filtersList
    };
    setAppliedFilters({ filters: filtersList, eventsList: listEvents, eventProp });
    setFiltersExpanded(false);
    const reqPayload = getFiltersRequestPayload({
      payload: opts,
      queriesList: listEvents,
      eventProp,
      table_props: displayTableProps
    });
    getProfileAccounts(activeProject.id, reqPayload, activeAgent);
    disableNewSegmentMode();
  }, [
    selectedAccount.account,
    filtersList,
    displayTableProps,
    listEvents,
    eventProp,
    getProfileAccounts,
    activeProject.id,
    activeAgent,
    disableNewSegmentMode
  ]);

  // const clearFilters = () => {
  //   const opts = { ...accountPayload };
  //   opts.filters = [];
  //   setAccountPayload(opts);
  //   setActiveSegment(activeSegment);
  //   getAccounts(opts);
  // };

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
      const updatedQuery = {
        ...activeSegment.query,
        table_props:
          checkListAccountProps
            ?.filter(({ enabled }) => enabled)
            ?.map(({ prop_name }) => prop_name)
            ?.filter((entry) => entry !== '' && entry !== undefined) || []
      };

      updateSegmentForId(activeProject.id, accountPayload.segment_id, {
        query: updatedQuery
      })
        .then(() => getSavedSegments(activeProject.id))
        .then(() => setActiveSegment({ ...activeSegment, query: updatedQuery }))
        .finally(() => getAccounts(accountPayload));
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

      udpateProjectSettings(activeProject.id, {
        timelines_config: updatedConfig
      }).then(() => getAccounts(accountPayload));
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
          showDisabledOption={isEngagementLocked}
          handleDisableOptionClick={handleDisableOptionClick}
        />
      </Tabs.TabPane>
    </Tabs>
  );

  const renderPropertyFilter = () => {
    return (
      <PropertyFilter
        profileType='account'
        source={accountPayload.source}
        filters={accountPayload.filters}
        filtersExpanded={filtersExpanded}
        filtersList={filtersList}
        appliedFilters={appliedFilters}
        selectedAccount={selectedAccount}
        listEvents={listEvents}
        availableGroups={Object.keys(groupOpts || {})}
        eventProp={eventProp}
        applyFilters={applyFilters}
        setFiltersExpanded={setFiltersExpanded}
        setSaveSegmentModal={setSaveSegmentModal}
        setFiltersList={setFiltersList}
        setSelectedAccount={setSelectedAccount}
        setAppliedFilters={setAppliedFilters}
        setListEvents={setListEvents}
        setEventProp={setEventProp}
      />
    );
  };

  // const renderClearFilterButton = () => (
  //   <Button
  //     className='dropdown-btn large mr-2'
  //     type='text'
  //     icon={<SVG name='times_circle' size={16} />}
  //     onClick={clearFilters}
  //   >
  //     Clear Filters
  //   </Button>
  // );

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

  const onApplyClick = (val) => {
    const parsedValues = val.map((vl) => JSON.parse(vl)[0]);
    const searchFilter = [];
    const lookIn =
      accountPayload.source === 'All'
        ? Object.entries(groupToCompanyPropMap)?.filter((item) =>
            groupsList?.map((item) => item?.[1])?.includes(item?.[0])
          )
        : [
            [
              accountPayload.source,
              groupToCompanyPropMap[accountPayload.source]
            ]
          ];
    lookIn.forEach(([group, prop]) => {
      searchFilter.push({
        props: ['', prop, 'categorical', 'group'],
        operator: 'contains',
        values: parsedValues
      });
    });

    const updatedPayload = {
      ...accountPayload,
      search_filter: formatFiltersForPayload(searchFilter)
    };
    const search_filters = updatedPayload.search_filter.map((filter, index) => {
      const isAnd = index === 0 ? filter.lop === 'AND' : filter.lop === 'OR';
      return isAnd ? filter : { ...filter, lop: 'OR' };
    });
    updatedPayload.search_filter = search_filters;

    setListSearchItems(parsedValues);
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

  const renderSaveSegmentButton = () => {
    return (
      <ControlledComponent
        controller={filtersExpanded === false && appliedFilters.length > 0}
      >
        <Button
          onClick={() => setSaveSegmentModal(true)}
          type='default'
          className='flex items-center col-gap-1'
        >
          <SVG color={'#1890ff'} size={16} name='pieChart' />
          <Text type='title' extraClass='mb-0' color={'brand-color-6'}>
            Save segment
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

  const handleTableChange = (pageParams) => {
    setCurrentPage(pageParams.current);
  };

  const tableColumns = useMemo(() => {
    return getColumns({
      accounts,
      isEngagementLocked,
      displayTableProps,
      groupPropNames,
      listProperties
    });
  }, [
    accounts,
    displayTableProps,
    groupPropNames,
    isEngagementLocked,
    listProperties
  ]);

  const renderTable = () => (
    <div>
      <Table
        onRow={(account) => ({
          onClick: () => {
            history.push(
              `/profiles/accounts/${btoa(account.identity)}?group=${
                activeSegment?.type ? activeSegment.type : accountPayload.source
              }&view=birdview`,
              {
                accountPayload: accountPayload,
                activeSegment: activeSegment,
                fromDetails: true,
                currentPage: currentPage
              }
            );
          }
        })}
        className='fa-table--userlist'
        dataSource={getTableData(accounts.data)}
        columns={tableColumns}
        rowClassName='cursor-pointer'
        pagination={{
          position: ['bottom', 'left'],
          defaultPageSize: '25',
          current: currentPage
        }}
        onChange={handleTableChange}
        scroll={{
          x: displayTableProps?.length * 300
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
    [activeProject.id, createNewSegment, getSavedSegments]
  );

  const handleCreateSegment = useCallback(
    (newSegmentName) => {
      const opts = { source: selectedAccount.account[1], filters: filtersList };
      const formatPayload = { ...opts };
      formatPayload.filters =
        formatFiltersForPayload(opts?.filters, 'accounts') || [];
      const reqPayload = formatReqPayload(formatPayload, activeSegment);
      reqPayload.name = newSegmentName;
      reqPayload.type = selectedAccount.account[1];
      handleSaveSegment(reqPayload);
      disableNewSegmentMode();
    },
    [
      activeSegment,
      filtersList,
      selectedAccount,
      handleSaveSegment,
      disableNewSegmentMode
    ]
  );

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
          <div
            className={cx(
              'flex items-center bg-red-200 rounded justify-center h-10 w-10',
              styles['title-icon-container']
            )}
          >
            <SVG name='buildings' size={24} color='#FF4D4F' />
          </div>
          <Text type='title' level={3} weight='bold' extraClass='mb-0'>
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
        {renderPropertyFilter()}
        <div className='inline-flex gap--6'>
          {/* {accountPayload?.filters?.length ? renderClearFilterButton() : null} */}
          {renderSaveSegmentButton()}
          {renderSearchSection()}
          {renderTablePropsSelect()}
        </div>
      </div>
      <ControlledComponent controller={accounts.isLoading === true}>
        <Spin size='large' className='fa-page-loader' />
      </ControlledComponent>
      <ControlledComponent
        controller={
          accounts.isLoading === false &&
          accounts.data.length > 0 &&
          newSegmentMode === false
        }
      >
        <>{renderTable()}</>
      </ControlledComponent>
      <ControlledComponent
        controller={
          accounts.isLoading === false &&
          accounts.data.length === 0 &&
          newSegmentMode === false
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
