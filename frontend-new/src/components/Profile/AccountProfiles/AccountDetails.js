import React, { useState, useEffect } from 'react';
import { Button, Dropdown, Menu, Popover, Tabs } from 'antd';
import { bindActionCreators } from 'redux';
import { connect } from 'react-redux';
import { Text, SVG } from '../../factorsComponents';
import AccountTimeline from './AccountTimeline';
import { getHost, granularityOptions } from '../utils';
import {
  udpateProjectSettings,
  fetchProjectSettings
} from '../../../reducers/global';
import { getProfileAccountDetails } from '../../../reducers/timelines/middleware';
import { getActivitiesWithEnableKeyConfig } from '../../../reducers/timelines/utils';
import SearchCheckList from '../../SearchCheckList';
import LeftPanePropBlock from '../LeftPanePropBlock';
import GroupSelect2 from '../../QueryComposer/GroupSelect2';

function AccountDetails({
  accountId,
  onCancel,
  accountDetails,
  activeProject,
  currentProjectSettings,
  fetchProjectSettings,
  udpateProjectSettings,
  getProfileAccountDetails,
  userProperties,
  groupProperties
}) {
  const [granularity, setGranularity] = useState('Daily');
  const [collapseAll, setCollapseAll] = useState(true);
  const [activities, setActivities] = useState([]);
  const [checkListUserProps, setCheckListUserProps] = useState([]);
  const [propSelectOpen, setPropSelectOpen] = useState(false);
  const [tlConfig, setTimelinesConfig] = useState({
    disabled_events: [],
    user_config: {
      props_to_show: []
    },
    account_config: {
      account_props_to_show: [],
      user_prop_to_show: ''
    }
  });

  useEffect(() => {
    if (currentProjectSettings?.timelines_config) {
      setTimelinesConfig({
        ...tlConfig,
        ...currentProjectSettings.timelines_config
      });
    }
  }, [currentProjectSettings]);

  useEffect(() => {
    fetchProjectSettings(activeProject.id);
  }, [activeProject]);

  useEffect(() => {
    const listActivities = getActivitiesWithEnableKeyConfig(
      accountDetails.data?.account_events,
      currentProjectSettings.timelines_config?.disabled_events
    );
    setActivities(listActivities);
  }, [currentProjectSettings, accountDetails]);

  useEffect(() => {
    const userPropsWithEnableKey = userProperties.map((userProp) => {
      const retObj = {
        display_name: userProp[0],
        prop_name: userProp[1],
        type: userProp[2],
        enabled: false
      };
      if (
        userProp[1] ===
        currentProjectSettings.timelines_config?.account_config
          ?.user_prop_to_show
      ) {
        retObj.enabled = true;
      }
      return retObj;
    });
    userPropsWithEnableKey.sort((a, b) => b.enabled - a.enabled);
    setCheckListUserProps(userPropsWithEnableKey);
  }, [currentProjectSettings, userProperties]);

  const handlePropChange = (option) => {
    const timelinesConfig = { ...currentProjectSettings.timelines_config };
    timelinesConfig.account_config.user_prop_to_show = option.prop_name;
    udpateProjectSettings(activeProject.id, {
      timelines_config: { ...timelinesConfig }
    }).then(() =>
      getProfileAccountDetails(
        activeProject.id,
        accountId,
        currentProjectSettings?.timelines_config
      )
    );
  };

  const handleEventsChange = (option) => {
    const timelinesConfig = { ...currentProjectSettings.timelines_config };
    if (option.enabled) {
      timelinesConfig.disabled_events.push(option.display_name);
    } else if (!option.enabled) {
      timelinesConfig.disabled_events.splice(
        timelinesConfig.disabled_events.indexOf(option.display_name),
        1
      );
    }
    udpateProjectSettings(activeProject.id, {
      timelines_config: { ...timelinesConfig }
    });
  };

  const controlsPopover = () => (
    <Tabs defaultActiveKey='events' size='small'>
      <Tabs.TabPane
        tab={<span className='fa-activity-filter--tabname'>Events</span>}
        key='events'
      >
        <SearchCheckList
          placeholder='Search Events'
          mapArray={activities}
          titleKey='display_name'
          checkedKey='enabled'
          onChange={handleEventsChange}
        />
      </Tabs.TabPane>
      <Tabs.TabPane
        tab={<span className='fa-activity-filter--tabname'>Properties</span>}
        key='properties'
      >
        <SearchCheckList
          placeholder='Search Properties'
          mapArray={checkListUserProps}
          titleKey='display_name'
          checkedKey='enabled'
          onChange={handlePropChange}
        />
      </Tabs.TabPane>
    </Tabs>
  );

  const granularityMenu = (
    <Menu>
      {granularityOptions.map((option) => (
        <Menu.Item key={option} onClick={(key) => setGranularity(key.key)}>
          <div className='flex items-center'>
            <span className='mr-3'>{option}</span>
          </div>
        </Menu.Item>
      ))}
    </Menu>
  );

  const handleOptionClick = (group, value) => {
    const timelinesConfig = { ...tlConfig };
    if (!timelinesConfig.user_config.props_to_show.includes(value[1])) {
      timelinesConfig.account_config.account_props_to_show.push(value[1]);
      udpateProjectSettings(activeProject.id, {
        timelines_config: { ...timelinesConfig }
      }).then(() =>
        getProfileAccountDetails(
          activeProject.id,
          accountId,
          currentProjectSettings?.timelines_config
        )
      );
    }
    setPropSelectOpen(false);
  };

  const onDelete = (option) => {
    const timelinesConfig = { ...tlConfig };
    timelinesConfig.account_config.account_props_to_show.splice(
      timelinesConfig.account_config.account_props_to_show.indexOf(option),
      1
    );
    udpateProjectSettings(activeProject.id, {
      timelines_config: { ...timelinesConfig }
    });
  };

  const renderModalHeader = () => (
    <div className='fa-timeline-modal--header'>
      <div className='flex items-center'>
        <Button
          style={{ padding: 0 }}
          type='text'
          icon={<SVG name='brand' size={36} />}
          size='large'
          onClick={() => {
            onCancel();
            setGranularity('Daily');
            setActivities([]);
            setCollapseAll(true);
            setPropSelectOpen(false);
          }}
        />
        <Text type='title' level={4} weight='bold'>
          Account Details
        </Text>
      </div>
      <Button
        size='large'
        type='text'
        onClick={() => {
          onCancel();
          setGranularity('Daily');
          setActivities([]);
          setCollapseAll(true);
          setPropSelectOpen(false);
        }}
        icon={<SVG name='times' />}
      />
    </div>
  );

  const listLeftPaneProps = (props = []) => {
    const propsList = [];
    const showProps =
      currentProjectSettings?.timelines_config?.account_config
        ?.account_props_to_show || [];

    showProps.forEach((prop, index) => {
      const value = props[prop] || '-';
      propsList.push(
        <div key={index}>
          <LeftPanePropBlock
            property={prop}
            value={value}
            onDelete={onDelete}
          />
        </div>
      );
    });
    return propsList;
  };

  const generateAccountProps = () => {
    const groupProps = [
      { label: 'Account Properties', icon: 'users', values: [] }
    ];
    groupProps[0].values = [
      ...(groupProperties.$hubspot_company
        ? groupProperties.$hubspot_company
        : []),
      ...(groupProperties.$salesforce_account
        ? groupProperties.$salesforce_account
        : [])
    ];
    return groupProps;
  };

  const selectProps = () =>
    propSelectOpen && (
      <div className='relative'>
        <GroupSelect2
          groupedProperties={generateAccountProps()}
          placeholder='Select Event'
          optionClick={handleOptionClick}
          onClickOutside={() => setPropSelectOpen(false)}
        />
      </div>
    );

  const renderAddNewProp = () =>
    currentProjectSettings?.timelines_config?.account_config
      ?.account_props_to_show?.length < 5 ? (
      <div>
        <Button
          type='link'
          icon={<SVG name='plus' color='purple' />}
          onClick={() => setPropSelectOpen(!propSelectOpen)}
        >
          Add property
        </Button>
        {selectProps()}
      </div>
    ) : null;

  const renderLeftPane = () => (
    <div className='fa-timeline-content__leftpane'>
      <div className='fa-timeline-content__leftpane__user'>
        <img
          src={`https://logo.clearbit.com/${getHost(
            accountDetails?.data?.host
          )}`}
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
          height={72}
          width={72}
        />
        <Text type='title' level={6} extraClass='m-0 py-2' weight='bold'>
          {accountDetails?.data?.name}
        </Text>
      </div>
      <div className='fa-timeline-content__leftpane__props'>
        {listLeftPaneProps(accountDetails.data.left_pane_props)}
      </div>
      <div className='px-8 pb-8'>{renderAddNewProp()}</div>
    </div>
  );

  const renderTimelineWithActions = () => (
    <div className='fa-timeline-content__activities'>
      <div className='fa-timeline-content__actions'>
        <Text type='title' level={3} weight='bold'>
          Timeline
        </Text>
        <div className='fa-timeline-content__actions__group'>
          <div className='fa-timeline-content__actions__group__collapse'>
            <Button
              className='collapse-btn collapse-btn--left'
              type='text'
              onClick={() => setCollapseAll(false)}
            >
              <SVG name='line_height' size={22} />
            </Button>
            <Button
              className='collapse-btn collapse-btn--right'
              type='text'
              onClick={() => setCollapseAll(true)}
            >
              <SVG name='grip_lines' size={22} />
            </Button>
          </div>

          <Popover
            overlayClassName='fa-activity--filter'
            placement='bottomLeft'
            trigger='hover'
            content={controlsPopover}
          >
            <Button
              size='large'
              className='fa-btn--custom mx-2 relative'
              type='text'
            >
              <SVG name='activity_filter' />
            </Button>
          </Popover>
          <Dropdown overlay={granularityMenu} placement='bottomRight'>
            <Button type='text' className='flex items-center'>
              {granularity}
              <SVG name='caretDown' size={16} extraClass='ml-1' />
            </Button>
          </Dropdown>
        </div>
      </div>
      <AccountTimeline
        timelineEvents={
          activities?.filter((activity) => activity.enabled === true) || []
        }
        timelineUsers={accountDetails.data?.account_users || []}
        collapseAll={collapseAll}
        setCollapseAll={setCollapseAll}
        granularity={granularity}
        loading={accountDetails?.isLoading}
      />
    </div>
  );

  return (
    <div>
      {renderModalHeader()}
      <div className='fa-timeline-content'>
        {renderLeftPane()}
        {renderTimelineWithActions()}
      </div>
    </div>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  currentProjectSettings: state.global.currentProjectSettings,
  accountDetails: state.timelines.accountDetails,
  userProperties: state.coreQuery.userProperties,
  groupProperties: state.coreQuery.groupProperties
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      getProfileAccountDetails,
      fetchProjectSettings,
      udpateProjectSettings
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(AccountDetails);
