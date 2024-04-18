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
  setNewSegmentModeAction,
  setTimelinePayloadAction
} from 'Reducers/userProfilesView/actions';
import ControlledComponent from 'Components/ControlledComponent/ControlledComponent';
import styles from './index.module.scss';
import SidebarMenuItem from './SidebarMenuItem';
import SidebarSearch from './SidebarSearch';
import { ProfilesSidebarIconsMapping } from './appSidebar.constants';
import { getSegmentColorCode } from './appSidebar.helpers';

function NewSegmentItem() {
  return <SidebarMenuItem text='Untitled Segment 1' isActive onClick={noop} />;
}

function GroupItem({ group }) {
  const dispatch = useDispatch();
  const timelinePayload = useSelector((state) => selectTimelinePayload(state));
  const { newSegmentMode } = useSelector((state) => state.userProfilesView);

  const setTimelinePayload = () => {
    if (timelinePayload.source !== group[1] || newSegmentMode === true) {
      dispatch(
        setTimelinePayloadAction({
          source: group[1],
          segment: {}
        })
      );
    }
  };

  const isActive =
    timelinePayload.source === group[1] &&
    !timelinePayload.segment.id &&
    newSegmentMode === false;

  return (
    <SidebarMenuItem
      text={group[0]}
      isActive={isActive}
      onClick={setTimelinePayload}
      icon={ProfilesSidebarIconsMapping[group[1]]}
    />
  );
}

function SegmentItem({ segment }) {
  const dispatch = useDispatch();
  const timelinePayload = useSelector((state) => selectTimelinePayload(state));
  const activeSegment = timelinePayload?.segment;
  const { newSegmentMode } = useSelector((state) => state.userProfilesView);

  const changeActiveSegment = () => {
    const opts = { ...timelinePayload };
    opts.source = segment?.type;
    opts.segment = segment;
    delete opts.search_filter;
    dispatch(setTimelinePayloadAction(opts));
  };

  const setActiveSegment = () => {
    if (activeSegment?.id !== segment?.id) {
      changeActiveSegment();
    }
  };

  const isActive =
    activeSegment?.id === segment?.id && newSegmentMode === false;
  const iconColor = getSegmentColorCode(segment?.name);

  return (
    <SidebarMenuItem
      text={segment?.name}
      isActive={isActive}
      onClick={setActiveSegment}
      icon='pieChart'
      iconColor={iconColor}
    />
  );
}

function ProfilesSidebar() {
  const [searchText, setSearchText] = useState('');
  const dispatch = useDispatch();
  const userOptions = getUserOptionsForDropdown();
  const { newSegmentMode } = useSelector((state) => state.userProfilesView);

  const userSegmentsList = useSelector((state) => selectSegmentsList(state));

  return (
    <div className='flex flex-col gap-y-5'>
      <div
        className={cx(
          'flex flex-col gap-y-6 overflow-auto',
          styles['accounts-list-container']
        )}
      >
        <div className='flex flex-col gap-y-3 px-4'>
          <Text
            type='title'
            level={8}
            extraClass='mb-0 px-2'
            color='character-secondary'
          >
            Segments
          </Text>
          <div className='flex flex-col gap-y-1'>
            <SidebarSearch
              searchText={searchText}
              setSearchText={setSearchText}
              placeholder='Search segment'
            />
            <ControlledComponent controller={newSegmentMode === true}>
              <NewSegmentItem />
            </ControlledComponent>
            {userOptions.slice(1).map((option) => (
              <GroupItem key={option[0]} group={option} />
            ))}
            <Fragment key='users'>
              {userSegmentsList
                ?.filter((segment) =>
                  segment?.name
                    ?.toLowerCase()
                    .includes(searchText.toLowerCase())
                )
                ?.sort((a, b) => a.name.localeCompare(b.name))
                ?.map((value) => (
                  <SegmentItem key={value.id} segment={value} />
                ))}
            </Fragment>
          </div>
        </div>
      </div>
      <div className='px-4'>
        <Button
          className={cx(
            'flex gap-x-2 items-center w-full',
            styles.sidebar_action_button
          )}
          onClick={() => {
            dispatch(setNewSegmentModeAction(true));
          }}
        >
          <SVG
            name={'plus'}
            size={16}
            extraClass={styles.sidebar_action_button__content}
            isFill={false}
          />
          <Text
            level={6}
            type='title'
            extraClass={cx('m-0', styles.sidebar_action_button__content)}
          >
            New Segment
          </Text>
        </Button>
      </div>
    </div>
  );
}

export default ProfilesSidebar;
