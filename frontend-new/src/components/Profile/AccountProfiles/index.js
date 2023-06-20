import React, { useState, useEffect, useMemo, useCallback } from 'react';
import {
  Table,
  Button,
  Modal,
  Spin,
  Popover,
  Tabs,
  notification,
  // Divider,
  Input
} from 'antd';
import { connect, useDispatch, useSelector } from 'react-redux';
import { bindActionCreators } from 'redux';
import { Text, SVG } from '../../factorsComponents';
import MomentTz from '../../MomentTz';
import PropertyFilter from '../MyComponents/PropertyFilter';
import { getGroupProperties } from '../../../reducers/coreQuery/middleware';
import FaSelect from '../../FaSelect';
import {
  DEFAULT_TIMELINE_CONFIG,
  // displayFilterOpts,
  EngagementTag,
  // formatEventsFromSegment,
  formatFiltersForPayload,
  // formatPayloadForFilters,
  getHost,
  getPropType,
  propValueFormat,
  sortNumericalColumn,
  sortStringColumn
} from '../utils';
import {
  getProfileAccounts,
  createNewSegment,
  getSavedSegments,
  updateSegmentForId
} from '../../../reducers/timelines/middleware';
import {
  fetchProjectSettings,
  udpateProjectSettings
} from '../../../reducers/global';
import SearchCheckList from 'Components/SearchCheckList';
import { formatUserPropertiesToCheckList } from 'Reducers/timelines/utils';
import { PropTextFormat } from 'Utils/dataFormatter';
// import GroupSelect2 from 'Components/QueryComposer/GroupSelect2';
import SegmentModal from '../UserProfiles/SegmentModal';
// import EventsBlock from '../MyComponents/EventsBlock';
import { useHistory, useLocation } from 'react-router-dom';
import {
  fetchGroupPropertyValues,
  fetchGroups
} from 'Reducers/coreQuery/services';
import NoDataWithMessage from '../MyComponents/NoDataWithMessage';
import ProfilesWrapper from '../ProfilesWrapper';
// import { generateSegmentsList, getGroupList } from './accountProfiles.helpers';
import { getGroupList } from './accountProfiles.helpers';
import {
  selectAccountPayload,
  selectActiveSegment,
  selectSegmentModalState
} from 'Reducers/accountProfilesView/selectors';
import {
  setAccountPayloadAction,
  setActiveSegmentAction,
  updateAccountPayloadAction,
  setSegmentModalStateAction
} from 'Reducers/accountProfilesView/actions';
import _ from 'lodash';

function AccountProfiles({
  activeProject,
  groupOpts,
  accounts,
  segments,
  createNewSegment,
  getSavedSegments,
  fetchProjectSettings,
  fetchGroups,
  udpateProjectSettings,
  currentProjectSettings,
  getProfileAccounts,
  getGroupProperties,
  updateSegmentForId
}) {
  const dispatch = useDispatch();
  const history = useHistory();
  const location = useLocation();

  const { groupPropNames } = useSelector((state) => state.coreQuery);
  const groupProperties = useSelector(
    (state) => state.coreQuery.groupProperties
  );
  const accountPayload = useSelector((state) => {
    if (location.state?.accountPayload && location.state?.fromDetails) {
      return location.state.accountPayload;
    } else {
    return selectAccountPayload(state);
    }
  });

  const activeSegment = useSelector((state) => {
    if (location.state?.activeSegment && location.state?.fromDetails) {
      return location.state.activeSegment;
    } else {
    return selectActiveSegment(state);
    }
  });

  const [currentPage, setCurrentPage] = useState(1);

  const showSegmentModal = useSelector((state) =>
    selectSegmentModalState(state)
  );
  const [searchBarOpen, setSearchBarOpen] = useState(false);
  const [searchDDOpen, setSearchDDOpen] = useState(false);
  const [listSearchItems, setListSearchItems] = useState([]);
  // const [isModalVisible, setIsModalVisible] = useState(false);
  // const [isGroupDDVisible, setGroupDDVisible] = useState(false);
  // const [isSegmentDDVisible, setSegmentDDVisible] = useState(false);
  const [listProperties, setListProperties] = useState([]);
  const [showPopOver, setShowPopOver] = useState(false);
  const [checkListAccountProps, setCheckListAccountProps] = useState([]);
  const [tlConfig, setTLConfig] = useState(DEFAULT_TIMELINE_CONFIG);
  const [companyValueOpts, setCompanyValueOpts] = useState({ All: {} });

  const agentState = useSelector((state) => state.agent);
  const activeAgent = agentState?.agent_details?.email;

  const setShowSegmentModal = useCallback(
    (value) => {
      dispatch(setSegmentModalStateAction(value));
    },
    [dispatch]
  );

  useEffect(() => {
    fetchProjectSettings(activeProject.id);
    fetchGroups(activeProject?.id, true);
  }, []);

  const groupsList = useMemo(() => {
    return getGroupList(groupOpts);
  }, [groupOpts]);

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
    return tableProps || [];
  }, [currentProjectSettings, accountPayload, activeSegment]);

  useEffect(() => {
    if (!accountPayload.search_filter) {
      setListSearchItems([]);
    } else {
      const listValues =
        accountPayload?.search_filter?.users?.map((vl) => vl?.va) || [];
      setListSearchItems(_.uniq(listValues));
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

  const setActiveSegment = useCallback(
    (segmentPayload) => {
      dispatch(setActiveSegmentAction(segmentPayload));
    },
    [dispatch]
  );


  useEffect(() => {
    if (!accountPayload.source) {
      const source = groupsList?.[0]?.[1] || '';
      updateAccountPayload({ source });
    }
  }, [groupsList]);

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
  }, [currentProjectSettings]);

  useEffect(() => {
    fetchProjectSettings(activeProject.id);
    getSavedSegments(activeProject.id);
  }, [activeProject.id, fetchProjectSettings, getSavedSegments]);

  useEffect(() => {
    Object.keys(groupOpts || {}).forEach((group) =>
      getGroupProperties(activeProject.id, group)
    );
  }, [activeProject.id, getGroupProperties, groupOpts]);

  const getAccounts = 
    (payload) => {
      const shouldCache = location.state?.fromDetails
      if (payload.source && payload.source !== '' && !shouldCache) {
        const formatPayload = { ...payload };
        formatPayload.filters =
          formatFiltersForPayload(payload?.filters, false) || {};
        getProfileAccounts(activeProject.id, formatPayload, activeAgent);
        
      } 
      if(shouldCache) {
        setCurrentPage(location.state.currentPage);
        const localeState = {...history.location.state, fromDetails: false}
        history.replace({state: localeState});
      }
   }

  useEffect(() => {
    getAccounts(accountPayload);
  }, [accountPayload.source, accountPayload.segment_id]);

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
  }, [groupProperties, accountPayload?.source]);

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

  const headerClassStr =
    'fai-text fai-text__color--grey-2 fai-text__size--h7 fai-text__weight--bold';

  const getTablePropColumn = (prop) => {
    const propDisplayName = groupPropNames[prop]
      ? groupPropNames[prop]
      : PropTextFormat(prop);
    const propType = getPropType(listProperties, prop);
    return {
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
      width: 280,
      sorter: (a, b) =>
        propType === 'numerical'
          ? sortNumericalColumn(a[prop], b[prop])
          : sortStringColumn(a[prop], b[prop]),
      render: (value) => (
        <Text type='title' level={7} extraClass='m-0' truncate>
          {value ? propValueFormat(prop, value, propType) : '-'}
        </Text>
      )
    };
  };

  const getColumns = () => {
    const columns = [
      {
        // Company Name Column
        title: <div className={headerClassStr}>Company Name</div>,
        dataIndex: 'account',
        key: 'account',
        width: 300,
        fixed: 'left',
        ellipsis: true,
        sorter: (a, b) => sortStringColumn(a.account.name, b.account.name),
        render: (item) =>
          (
            <div className='flex items-center'>
              <img
                src={`https://logo.uplead.com/${getHost(item.host)}`}
                onError={(e) => {
                  if (
                    e.target.src !==
                    'https://s3.amazonaws.com/www.factors.ai/assets/img/buildings.svg'
                  ) {
                    e.target.src =
                      'https://s3.amazonaws.com/www.factors.ai/assets/img/buildings.svg';
                  }
                }}
                alt=''
                width='20'
                height='20'
              />
              <span className='ml-2'>{item.name}</span>
            </div>
          ) || '-'
      }
    ];
    // Engagement Column
    const engagementExists = accounts.data?.find(
      (item) =>
        item.engagement &&
        (item.engagement !== undefined || item.engagement !== '')
    );
    if (engagementExists) {
      columns.push({
        title: <div className={headerClassStr}>Engagement</div>,
        width: 150,
        dataIndex: 'engagement',
        key: 'engagement',
        fixed: 'left',
        sorter: (a, b) => sortNumericalColumn(a.score, b.score),
        render: (status) =>
          status ? (
            <div
              className='engagement-tag'
              style={{ '--bg-color': EngagementTag[status]?.bgColor }}
            >
              <img
                src={`../../../assets/icons/${EngagementTag[status]?.icon}.svg`}
                alt=''
              />
              <Text type='title' level={7} extraClass='m-0'>
                {status}
              </Text>
            </div>
          ) : (
            '-'
          )
      });
    }
    // Table Prop Columns
    displayTableProps?.forEach((prop) => {
      columns.push(getTablePropColumn(prop));
    });
    // Last Activity Column
    columns.push({
      title: <div className={headerClassStr}>Last Activity</div>,
      dataIndex: 'lastActivity',
      key: 'lastActivity',
      width: 200,
      align: 'right',
      sorter: (a, b) => sortStringColumn(a.lastActivity, b.lastActivity),
      render: (item) => MomentTz(item).fromNow()
    });
    return columns;
  };

  const getTableData = (data) => {
    const sortedData = data.sort(
      (a, b) => new Date(b.last_activity) - new Date(a.last_activity)
    );
    return sortedData.map((row) => ({
      ...row,
      ...row?.tableProps
    }));
  };

  // const showModal = () => {
  //   setIsModalVisible(true);
  // };

  // const handleCancel = () => {
  //   setIsModalVisible(false);
  // };

  // const onChange = (val) => {
  //   if (val !== accountPayload.source) {
  //     const opts = { ...accountPayload };
  //     opts.source = val;
  //     opts.filters = [];
  //     opts.segment_id = '';
  //     setAccountPayload(opts);
  //   }
  //   setGroupDDVisible(false);
  // };

  const setFilters = (filters) => {
    const opts = { ...accountPayload };
    opts.filters = filters;
    setAccountPayload(opts);
    setActiveSegment(activeSegment);
    getAccounts(opts);
  };

  const clearFilters = () => {
    const opts = { ...accountPayload };
    opts.filters = [];
    setAccountPayload(opts);
    setActiveSegment(activeSegment);
  };

  // const selectGroup = () => (
  //   <div className='absolute top-0'>
  //     {isGroupDDVisible ? (
  //       <FaSelect
  //         options={groupsList}
  //         onClickOutside={() => setGroupDDVisible(false)}
  //         optionClick={(val) => onChange(val[1])}
  //       />
  //     ) : null}
  //   </div>
  // );

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
        table_props: checkListAccountProps
          .filter(({ enabled }) => enabled)
          .map(({ prop_name }) => prop_name)
      };

      updateSegmentForId(activeProject.id, accountPayload.segment_id, {
        query: updatedQuery
      })
        .then(() => getSavedSegments(activeProject.id))
        .finally(() =>
          setActiveSegment({ ...activeSegment, query: updatedQuery })
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

      udpateProjectSettings(activeProject.id, {
        timelines_config: updatedConfig
      });
    }
    setShowPopOver(false);
    getAccounts(accountPayload);
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
        />
      </Tabs.TabPane>
    </Tabs>
  );

  // const onOptionClick = (_, data) => {
  //   const opts = { ...accountPayload };
  //   opts.segment_id = data[1];
  //   opts.source = data[2].type;
  //   setActiveSegment(data[2], opts);
  //   setSegmentDDVisible(false);
  // };

  const handleSaveSegment = async (segmentPayload) => {
    try {
      const response = await createNewSegment(activeProject.id, segmentPayload);
      if (response.type === 'SEGMENT_CREATION_FULFILLED') {
        notification.success({
          message: 'Success!',
          description: response?.payload?.message,
          duration: 3
        });
        setShowSegmentModal(false);
        // setSegmentDDVisible(false);
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
  };

  // const clearSegment = () => {
  //   const opts = { ...accountPayload };
  //   opts.segment_id = '';
  //   setActiveSegment({}, opts);
  //   setSegmentDDVisible(false);
  // };

  // const renderAdditionalActionsInSegment = () => (
  //   <div className='mb-2'>
  //     <Divider className='divider-margin' />
  //     <div className='flex items-center flex-col'>
  //       {accountPayload.segment_id && (
  //         <Button
  //           size='large'
  //           type='text'
  //           className='w-full mb-2'
  //           onClick={clearSegment}
  //           icon={<SVG name='remove' />}
  //         >
  //           Clear Segment
  //         </Button>
  //       )}
  //       <Button
  //         type='link'
  //         size='large'
  //         className='w-full'
  //         icon={<SVG name='plus' color='purple' />}
  //         onClick={() => setShowSegmentModal(true)}
  //       >
  //         Add New Segment
  //       </Button>
  //     </div>
  //   </div>
  // );

  // const selectSegment = () => (
  //   <div className='absolute top-8'>
  //     {isSegmentDDVisible ? (
  //       <GroupSelect2
  //         groupedProperties={generateSegmentsList({ accountPayload, segments })}
  //         placeholder='Search Segments'
  //         optionClick={onOptionClick}
  //         onClickOutside={() => setSegmentDDVisible(false)}
  //         additionalActions={renderAdditionalActionsInSegment()}
  //       />
  //     ) : null}
  //   </div>
  // );

  // const eventsList = (listEvents) => {
  //   const blockList = listEvents.map((event, index) => (
  //     <div key={index} className='m-0 mr-2 mb-2'>
  //       <EventsBlock
  //         availableGroups={groupsList}
  //         index={index + 1}
  //         event={event}
  //         queries={listEvents}
  //         viewMode
  //       />
  //     </div>
  //   ));

  //   if (!blockList.length) {
  //     return null;
  //   }
  //   return (
  //     <div className='segment-query_block'>
  //       <h2
  //         className={`title ${
  //           activeSegment?.query?.gup?.length ? '' : 'width-unset'
  //         }`}
  //       >
  //         Performed Events
  //       </h2>
  //       <div className='content'>{blockList}</div>
  //     </div>
  //   );
  // };

  // const filtersList = (filters) => {
  //   return (
  //     <div className='segment-query_block'>
  //       <h2
  //         className={`title ${
  //           activeSegment?.query?.ewp?.length ? '' : 'width-unset'
  //         }`}
  //       >
  //         Properties
  //       </h2>
  //       <div className='content'>
  //         <PropertyFilter
  //           filtersLimit={10}
  //           profileType='account'
  //           source={accountPayload.source}
  //           filters={filters}
  //           availableGroups={Object.keys(groupOpts || {})}
  //           viewMode
  //         />
  //       </div>
  //     </div>
  //   );
  // };

  // const segmentInfo = () => {
  //   if (!activeSegment.query) {
  //     return null;
  //   }

  //   return (
  //     <div className='p-3'>
  //       {activeSegment.query.ewp && activeSegment.query.ewp.length
  //         ? eventsList(formatEventsFromSegment(activeSegment.query.ewp))
  //         : null}
  //       {activeSegment.query.gup && activeSegment.query.gup.length
  //         ? filtersList(formatPayloadForFilters(activeSegment.query.gup))
  //         : null}
  //       {activeSegment.query.ewp && activeSegment.query.ewp.length ? (
  //         <h2 className='whitespace-no-wrap italic line-height-8 m-0 mr-2'>
  //           {`*Shows ${
  //             displayFilterOpts[activeSegment.type]
  //           } from last 28 days.`}
  //         </h2>
  //       ) : null}
  //     </div>
  //   );
  // };

  // const renderGroupSelectDD = () => (
  //   <div className='relative mr-2'>
  //     <Button
  //       className='dropdown-btn'
  //       type='text'
  //       icon={<SVG name='user_friends' size={16} />}
  //       onClick={() => setGroupDDVisible(!isGroupDDVisible)}
  //     >
  //       {
  //         groupsList?.find(
  //           ([_, groupName]) => groupName === accountPayload?.source
  //         )?.[0]
  //       }
  //       <SVG name='caretDown' size={16} />
  //     </Button>
  //     {selectGroup()}
  //   </div>
  // );

  // const renderSegmentSelect = () => (
  //   <div className='relative mr-2'>
  //     <Popover
  //       overlayClassName='fa-custom-popover'
  //       placement='bottomLeft'
  //       trigger={activeSegment.query ? 'hover' : ''}
  //       content={segmentInfo}
  //       mouseEnterDelay={0.5}
  //     >
  //       <Button
  //         className='dropdown-btn'
  //         type='text'
  //         onClick={() => setSegmentDDVisible(!isSegmentDDVisible)}
  //       >
  //         {Object.keys(activeSegment).length
  //           ? activeSegment.name
  //           : 'Select Segment'}
  //         <SVG name='caretDown' size={16} />
  //       </Button>
  //     </Popover>
  //     {selectSegment()}
  //   </div>
  // );

  const renderPropertyFilter = () => (
    <div key={0} className='max-w-3xl'>
      <PropertyFilter
        profileType='account'
        source={accountPayload.source}
        filters={accountPayload.filters}
        setFilters={setFilters}
        availableGroups={Object.keys(groupOpts || {})}
      />
    </div>
  );

  const renderClearFilterButton = () => (
    <Button
      className='dropdown-btn large mr-2'
      type='text'
      icon={<SVG name='times_circle' size={16} />}
      onClick={clearFilters}
    >
      Clear Filters
    </Button>
  );

  const groupToCompanyPropMap = {
    $hubspot_company: '$hubspot_company_name',
    $salesforce_account: '$salesforce_account_name',
    $6signal: '$6Signal_name',
    $linkedin_company: '$li_localized_name',
    $g2: '$g2_name'
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
        props: [prop, 'categorical', 'group'],
        operator: 'equals',
        values: parsedValues
      });
    });

    const updatedPayload = {
      ...accountPayload,
      search_filter: formatFiltersForPayload(searchFilter, true)
    };
    const search_filter_map = {};
    search_filter_map['users'] = updatedPayload.search_filter.map(
      (filter, index) => {
        const isAnd = index === 0 ? filter.lop === 'AND' : filter.lop === 'OR';
        return isAnd ? filter : { ...filter, lop: 'OR' };
      }
    );
    updatedPayload.search_filter = search_filter_map;

    setListSearchItems(parsedValues);
    setAccountPayload(updatedPayload);
    setActiveSegment(activeSegment);
    getAccounts(updatedPayload);
  };

  const onSearchClose = () => {
    setSearchBarOpen(false);
    setSearchDDOpen(false);
    if (Object.keys(accountPayload?.search_filter || {}).length !== 0) {
      const updatedPayload = { ...accountPayload };
      updatedPayload.search_filter = {};
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

  const renderSearchSection = () => (
    <div className='relative mr-2'>
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
      <Button
        size='large'
        icon={<SVG name='activity_filter' />}
        className='relative'
      >
        Configure
      </Button>
    </Popover>
  );

  const handleTableChange = (pageParams) => {
    setCurrentPage(pageParams.current);
  }

  const renderActions = () => (
    <div className='flex justify-between items-start my-4'>
      <div className='flex justify-between'>
        {/* {renderGroupSelectDD()}
        {renderSegmentSelect()} */}
        {renderPropertyFilter()}
      </div>
      <div className='flex items-center justify-between'>
        {accountPayload?.filters?.length ? renderClearFilterButton() : null}
        {renderSearchSection()}
        {renderTablePropsSelect()}
      </div>
    </div>
  );

  const renderTable = () => (
    <div>
      <Table
        onRow={(account) => ({
          onClick: () => {
            history.push(
              `/profiles/accounts/${btoa(account.identity)}?group=${
                activeSegment?.type ? activeSegment.type : accountPayload.source
              }&view=birdview`,
              { accountPayload: accountPayload, activeSegment: activeSegment, currentPage: currentPage }
            );
          }
        })}
        className='fa-table--userlist'
        dataSource={getTableData(accounts.data)}
        columns={getColumns()}
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

  return (
    <ProfilesWrapper>
      <Text type='title' level={3} weight='bold' extraClass='mb-0'>
        Account Profiles
      </Text>
      {renderActions()}
      {accounts.isLoading ? (
        <Spin size='large' className='fa-page-loader' />
      ) : accounts.data.length ? (
        renderTable()
      ) : (
        <NoDataWithMessage message={'No Accounts Found'} />
      )}
      <SegmentModal
        profileType='account'
        activeProject={activeProject}
        type={accountPayload.source}
        typeOptions={groupsList.filter((group) => group[1] !== 'All')}
        visible={showSegmentModal}
        segment={{}}
        onSave={handleSaveSegment}
        onCancel={() => setShowSegmentModal(false)}
        caller={'account_profiles'}
        tableProps={
          currentProjectSettings.timelines_config?.account_config?.table_props
        }
      />
    </ProfilesWrapper>
  );
}
const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  groupOpts: state.groups.data,
  accounts: state.timelines.accounts,
  segments: state.timelines.segments,
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
      updateSegmentForId
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(AccountProfiles);
