import React, { Fragment, useState } from 'react';
import noop from 'lodash/noop';
import cx from 'classnames';
import { useDispatch, useSelector } from 'react-redux';
import { Button } from 'antd';
import { getUserOptionsForDropdown } from 'Components/Profile/UserProfiles/userProfiles.helpers';
import { SVG, Text } from 'Components/factorsComponents';
import {
  selectSegmentsList,
  selectTimelinePayload
} from 'Reducers/userProfilesView/selectors';
import {
  setActiveSegmentAction,
  setNewSegmentModeAction,
  setTimelinePayloadAction
} from 'Reducers/userProfilesView/actions';
import styles from './index.module.scss';
import SidebarMenuItem from './SidebarMenuItem';
import SidebarSearch from './SidebarSearch';
import { ProfilesSidebarIconsMapping } from './appSidebar.constants';
import ControlledComponent from 'Components/ControlledComponent/ControlledComponent';
import { getSegmentColorCode } from './appSidebar.helpers';

const NewSegmentItem = () => {
  return (
    <SidebarMenuItem
      text={'Untitled Segment 1'}
      isActive={true}
      onClick={noop}
    />
  );
};

const GroupItem = ({ group }) => {
  const dispatch = useDispatch();
  const timelinePayload = useSelector((state) => selectTimelinePayload(state));
  const { newSegmentMode } = useSelector((state) => state.userProfilesView);

  const setTimelinePayload = () => {
    if (timelinePayload.source !== group[1] || newSegmentMode === true) {
      dispatch(
        setTimelinePayloadAction({
          source: group[1],
          filters: [],
          segment_id: ''
        })
      );
      dispatch(setActiveSegmentAction({}));
    }
  };

  const isActive =
    timelinePayload.source === group[1] &&
    !timelinePayload.segment_id &&
    newSegmentMode === false;

  return (
    <SidebarMenuItem
      text={group[0]}
      isActive={isActive}
      onClick={setTimelinePayload}
      icon={ProfilesSidebarIconsMapping[group[1]]}
    />
  );
};

const SegmentItem = ({ segment }) => {
  const dispatch = useDispatch();
  const timelinePayload = useSelector((state) => selectTimelinePayload(state));
  const { newSegmentMode } = useSelector((state) => state.userProfilesView);

  const setActiveSegment = () => {
    if (timelinePayload.segment_id !== segment[1] || newSegmentMode === true) {
      const opts = { ...timelinePayload };
      opts.source = segment[2].type;
      opts.segment_id = segment[1];
      opts.filters = [];
      delete opts.search_filter;
      dispatch(setActiveSegmentAction(segment[2]));
      dispatch(setTimelinePayloadAction(opts));
    }
  };

  const isActive =
    timelinePayload.segment_id === segment[1] && newSegmentMode === false;
  const iconColor = getSegmentColorCode(segment[0]);

  return (
    <SidebarMenuItem
      text={segment[0]}
      isActive={isActive}
      onClick={setActiveSegment}
      icon={'pieChart'}
      iconColor={iconColor}
    />
  );
};

const ProfilesSidebar = () => {
  const [searchText, setSearchText] = useState('');
  const dispatch = useDispatch();
  const userOptions = getUserOptionsForDropdown();
  const { newSegmentMode } = useSelector((state) => state.userProfilesView);

  const segmentsList = useSelector((state) => selectSegmentsList(state));

  return (
    <div className='flex flex-col row-gap-5'>
      <div
        className={cx(
          'flex flex-col row-gap-6 overflow-auto',
          styles['accounts-list-container']
        )}
      >
        <div className='flex flex-col row-gap-3 px-4'>
          <Text
            type='title'
            level={8}
            extraClass='mb-0 px-2'
            color='character-secondary'
          >
            Segments
          </Text>
          <div className='flex flex-col row-gap-1'>
            <SidebarSearch
              searchText={searchText}
              setSearchText={setSearchText}
              placeholder={'Search segment'}
            />
            <ControlledComponent controller={newSegmentMode === true}>
              <NewSegmentItem />
            </ControlledComponent>
            {userOptions.slice(1).map((option) => {
              return <GroupItem key={option[0]} group={option} />;
            })}
            {segmentsList.map((segment) => {
              if (segment.values != null) {
                const filteredSegments = segment.values.filter((value) =>
                  value[0].toLowerCase().includes(searchText.toLowerCase())
                );
                return (
                  <Fragment key={segment.label}>
                    {filteredSegments.map((value) => {
                      return <SegmentItem key={value[1]} segment={value} />;
                    })}
                  </Fragment>
                );
              }
              return null;
            })}
          </div>
        </div>
      </div>
      <div className='px-4'>
        <Button
          className={cx(
            'flex col-gap-2 items-center w-full',
            styles['sidebar-action-button']
          )}
          type='dashed'
          onClick={() => {
            dispatch(setNewSegmentModeAction(true));
          }}
        >
          <SVG name={'plus'} size={16} />
          <Text level={7} type='title' extraClass='mb-0'>
            New Segment
          </Text>
        </Button>
      </div>
    </div>
  );
};

export default ProfilesSidebar;
