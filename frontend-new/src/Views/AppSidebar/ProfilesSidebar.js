import React, { Fragment, useState } from 'react';
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
  setSegmentModalStateAction,
  setTimelinePayloadAction
} from 'Reducers/userProfilesView/actions';
import styles from './index.module.scss';
import SidebarMenuItem from './SidebarMenuItem';
import SidebarSearch from './SidebarSearch';

const GroupItem = ({ group }) => {
  const dispatch = useDispatch();
  const timelinePayload = useSelector((state) => selectTimelinePayload(state));

  const setTimelinePayload = () => {
    dispatch(
      setTimelinePayloadAction({
        source: group[1],
        filters: [],
        segment_id: ''
      })
    );
  };

  const isActive =
    timelinePayload.source === group[1] && !timelinePayload.segment_id;

  return (
    <SidebarMenuItem
      text={group[0]}
      isActive={isActive}
      onClick={setTimelinePayload}
    />
  );
};

const SegmentItem = ({ segment }) => {
  const dispatch = useDispatch();
  const timelinePayload = useSelector((state) => selectTimelinePayload(state));

  const setActiveSegment = () => {
    const opts = { ...timelinePayload };
    opts.source = segment[2].type;
    opts.segment_id = segment[1];
    dispatch(
      setActiveSegmentAction({
        segmentPayload: segment[2],
        timelinePayload: opts
      })
    );
  };

  const isActive = timelinePayload.segment_id === segment[1];

  return (
    <SidebarMenuItem
      text={segment[0]}
      isActive={isActive}
      onClick={setActiveSegment}
    />
  );
};

const ProfilesSidebar = () => {
  const [searchText, setSearchText] = useState('');
  const dispatch = useDispatch();
  const userOptions = getUserOptionsForDropdown();

  const segmentsList = useSelector((state) => selectSegmentsList(state));

  return (
    <div className='flex flex-col row-gap-5'>
      <div
        className={cx(
          'flex flex-col row-gap-6 overflow-auto',
          styles['accounts-list-container']
        )}
      >
        <div className='flex flex-col row-gap-1 px-4 pb-6 border-b'>
          <Text
            type='title'
            level={8}
            extraClass='mb-0 px-2'
            color='character-secondary'
          >
            Default
          </Text>
          {userOptions.map((option) => {
            return <GroupItem key={option[0]} group={option} />;
          })}
        </div>
        <div className='flex flex-col row-gap-3 px-4'>
          <Text
            type='title'
            level={8}
            extraClass='mb-0 px-2'
            color='character-secondary'
          >
            Custom
          </Text>
          <div className='flex flex-col row-gap-1'>
            <SidebarSearch
              searchText={searchText}
              setSearchText={setSearchText}
              placeholder={'Search segment'}
            />
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
      <Button
        className={cx(
          'flex col-gap-2 items-center',
          styles['sidebar-action-button']
        )}
        type='secondary'
        onClick={() => {
          dispatch(setSegmentModalStateAction(true));
        }}
      >
        <SVG name={'plus'} size={16} color='#1890FF' />
        <Text level={7} type='title' color='brand-color-6' extraClass='mb-0'>
          New Segment
        </Text>
      </Button>
    </div>
  );
};

export default ProfilesSidebar;
