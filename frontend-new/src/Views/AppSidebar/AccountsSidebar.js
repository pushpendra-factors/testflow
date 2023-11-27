import React, { Fragment, useMemo, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import cx from 'classnames';
import { Button } from 'antd';
import noop from 'lodash/noop';
import { generateSegmentsList } from 'Components/Profile/AccountProfiles/accountProfiles.helpers';
import { SVG, Text } from 'Components/factorsComponents';
import {
  setAccountPayloadAction,
  setActiveSegmentAction,
  setNewSegmentModeAction
} from 'Reducers/accountProfilesView/actions';
import { selectAccountPayload } from 'Reducers/accountProfilesView/selectors';
import { selectSegments } from 'Reducers/timelines/selectors';
import styles from './index.module.scss';
import SidebarMenuItem from './SidebarMenuItem';
import SidebarSearch from './SidebarSearch';
import ControlledComponent from 'Components/ControlledComponent/ControlledComponent';
import { defaultSegmentIconsMapping } from './appSidebar.constants';
import { useHistory } from 'react-router-dom';
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

const SegmentIcon = (name) => {
  return defaultSegmentIconsMapping[name]
    ? defaultSegmentIconsMapping[name]
    : 'pieChart';
};

const SegmentItem = ({ segment }) => {
  const dispatch = useDispatch();
  const history = useHistory();
  const activeAccountPayload = useSelector((state) =>
    selectAccountPayload(state)
  );

  const { newSegmentMode, activeSegment } = useSelector(
    (state) => state.accountProfilesView
  );

  const changeActiveSegment = () => {
    const opts = { ...activeAccountPayload };
    opts.segment_id = segment[1];
    opts.source = segment[2].type;
    opts.filters = [];
    delete opts.search_filter;
    history.replace({ pathname: '/accounts/segments/' + segment[1] });
    dispatch(setActiveSegmentAction(segment[2]));
    dispatch(setAccountPayloadAction(opts));
  };

  const setActiveSegment = () => {
    if (activeSegment?.id !== segment[1]) {
      changeActiveSegment();
    }
  };

  const isActive = activeSegment?.id === segment[1] && newSegmentMode === false;
  const iconColor = getSegmentColorCode(segment[0]);

  return (
    <SidebarMenuItem
      text={segment[0]}
      isActive={isActive}
      onClick={setActiveSegment}
      icon={SegmentIcon(segment[0])}
      iconColor={iconColor}
    />
  );
};

const AccountsSidebar = () => {
  const [searchText, setSearchText] = useState('');
  const dispatch = useDispatch();
  const segments = useSelector((state) => selectSegments(state));
  const activeAccountPayload = useSelector((state) =>
    selectAccountPayload(state)
  );

  const { newSegmentMode } = useSelector((state) => state.accountProfilesView);

  const segmentsList = useMemo(() => {
    return generateSegmentsList({
      accountPayload: activeAccountPayload,
      segments
    });
  }, [activeAccountPayload, segments]);

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

export default AccountsSidebar;
