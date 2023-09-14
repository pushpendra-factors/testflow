import React, { Fragment, useMemo, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import cx from 'classnames';
import { Button } from 'antd';
import noop from 'lodash/noop';
import {
  generateSegmentsList,
  getGroupList
} from 'Components/Profile/AccountProfiles/accountProfiles.helpers';
import { SVG, Text } from 'Components/factorsComponents';
import {
  setAccountPayloadAction,
  setActiveSegmentAction,
  setExitConfirmationModalAction,
  setNewSegmentModeAction
} from 'Reducers/accountProfilesView/actions';
import { selectAccountPayload } from 'Reducers/accountProfilesView/selectors';
import { selectGroupOptions } from 'Reducers/groups/selectors';
import { selectSegments } from 'Reducers/timelines/selectors';
import styles from './index.module.scss';
import SidebarMenuItem from './SidebarMenuItem';
import SidebarSearch from './SidebarSearch';
import ControlledComponent from 'Components/ControlledComponent/ControlledComponent';
import { AccountsSidebarIconsMapping } from './appSidebar.constants';

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
  const activeAccountPayload = useSelector((state) =>
    selectAccountPayload(state)
  );
  const { newSegmentMode, filtersDirty: areFiltersDirty } = useSelector(
    (state) => state.accountProfilesView
  );

  const changeAccountPayload = () => {
    dispatch(
      setAccountPayloadAction({
        source: group[1],
        filters: [],
        segment_id: ''
      })
    );
    dispatch(setActiveSegmentAction({}));
  };

  const setAccountPayload = () => {
    if (activeAccountPayload.source !== group[1]) {
      if (areFiltersDirty === false) {
        changeAccountPayload();
      } else {
        dispatch(setExitConfirmationModalAction(true, changeAccountPayload));
      }
    }
  };

  const isActive =
    activeAccountPayload.source === group[1] &&
    !activeAccountPayload.segment_id &&
    newSegmentMode === false;

  return (
    <SidebarMenuItem
      text={group[0]}
      isActive={isActive}
      onClick={setAccountPayload}
      icon={AccountsSidebarIconsMapping[group[1]]}
    />
  );
};

const SegmentItem = ({ segment }) => {
  const dispatch = useDispatch();
  const activeAccountPayload = useSelector((state) =>
    selectAccountPayload(state)
  );
  const { newSegmentMode, filtersDirty: areFiltersDirty } = useSelector(
    (state) => state.accountProfilesView
  );

  const changeActiveSegment = () => {
    const opts = { ...activeAccountPayload };
    opts.segment_id = segment[1];
    opts.source = segment[2].type;
    opts.filters = [];
    delete opts.search_filter;
    dispatch(setActiveSegmentAction(segment[2]));
    dispatch(setAccountPayloadAction(opts));
  };

  const setActiveSegment = () => {
    if (activeAccountPayload.segment_id !== segment[1]) {
      if (areFiltersDirty === false) {
        changeActiveSegment();
      } else {
        dispatch(setExitConfirmationModalAction(true, changeActiveSegment));
      }
    }
  };

  const isActive =
    activeAccountPayload.segment_id === segment[1] && newSegmentMode === false;

  return (
    <SidebarMenuItem
      text={segment[0]}
      isActive={isActive}
      onClick={setActiveSegment}
      icon='pieChart'
      iconColor={'#595959'}
    />
  );
};

const AccountsSidebar = () => {
  const [searchText, setSearchText] = useState('');
  const dispatch = useDispatch();
  const groupOptions = useSelector((state) => selectGroupOptions(state));
  const segments = useSelector((state) => selectSegments(state));
  const activeAccountPayload = useSelector((state) =>
    selectAccountPayload(state)
  );
  const { newSegmentMode } = useSelector((state) => state.accountProfilesView);

  const groupsList = useMemo(() => {
    return getGroupList(groupOptions);
  }, [groupOptions]);

  const segmentsList = useMemo(() => {
    return generateSegmentsList({
      accountPayload: activeAccountPayload,
      segments
    });
  }, [activeAccountPayload, segments]);

  const filteredGroupsList = groupsList
    .slice(1)
    .filter((value) =>
      value[0].toLowerCase().includes(searchText.toLowerCase())
    );

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
            {filteredGroupsList.map((group) => {
              return <GroupItem key={group[0]} group={group} />;
            })}
            <ControlledComponent controller={newSegmentMode === true}>
              <NewSegmentItem />
            </ControlledComponent>
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
          type='secondary'
          onClick={() => {
            dispatch(setNewSegmentModeAction(true));
          }}
        >
          <SVG name={'plus'} size={16} color='#1890FF' />
          <Text level={7} type='title' color='brand-color-6' extraClass='mb-0'>
            New Segment
          </Text>
        </Button>
      </div>
    </div>
  );
};

export default AccountsSidebar;
