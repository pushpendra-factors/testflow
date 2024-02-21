import React, { Fragment, useMemo, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import cx from 'classnames';
import { Button } from 'antd';
import noop from 'lodash/noop';
import { SVG, Text } from 'Components/factorsComponents';
import {
  setAccountPayloadAction,
  setNewSegmentModeAction
} from 'Reducers/accountProfilesView/actions';
import { selectAccountPayload } from 'Reducers/accountProfilesView/selectors';
import { selectSegments } from 'Reducers/timelines/selectors';
import ControlledComponent from 'Components/ControlledComponent/ControlledComponent';
import { useHistory } from 'react-router-dom';
import { reorderDefaultDomainSegmentsToTop } from 'Components/Profile/AccountProfiles/accountProfiles.helpers';
import { GROUP_NAME_DOMAINS } from 'Components/GlobalFilter/FilterWrapper/utils';
import styles from './index.module.scss';
import SidebarMenuItem from './SidebarMenuItem';
import SidebarSearch from './SidebarSearch';
import { defaultSegmentIconsMapping } from './appSidebar.constants';
import { getSegmentColorCode } from './appSidebar.helpers';
import { PathUrls } from 'Routes/pathUrls';

function NewSegmentItem() {
  return <SidebarMenuItem text='Untitled Segment 1' isActive onClick={noop} />;
}

const SegmentIcon = (name) => defaultSegmentIconsMapping[name] || 'pieChart';

function SegmentItem({ segment }) {
  const history = useHistory();
  const dispatch = useDispatch();
  const activeAccountPayload = useSelector(selectAccountPayload);
  const activeSegment = activeAccountPayload?.segment;
  const { newSegmentMode } = useSelector((state) => state.accountProfilesView);

  const changeActiveSegment = () => {
    dispatch(setNewSegmentModeAction(false));
    dispatch(setAccountPayloadAction({ source: GROUP_NAME_DOMAINS, segment }));
    history.replace({ pathname: `/accounts/segments/${segment.id}` });
  };

  const setActiveSegment = () => {
    if (activeSegment?.id !== segment?.id) {
      changeActiveSegment();
    }
  };

  const isActive = activeSegment?.id === segment?.id && !newSegmentMode;
  const iconColor = getSegmentColorCode(segment?.name);

  return (
    <SidebarMenuItem
      text={segment?.name}
      isActive={isActive}
      onClick={setActiveSegment}
      icon={SegmentIcon(segment?.name)}
      iconColor={iconColor}
    />
  );
}

function AccountsSidebar() {
  const history = useHistory();
  const [searchText, setSearchText] = useState('');
  const dispatch = useDispatch();
  const segments = useSelector(selectSegments);
  const { newSegmentMode } = useSelector((state) => state.accountProfilesView);

  const segmentsList = useMemo(
    () => reorderDefaultDomainSegmentsToTop(segments[GROUP_NAME_DOMAINS]) || [],
    [segments]
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
              placeholder='Search segment'
            />
            <ControlledComponent controller={newSegmentMode}>
              <NewSegmentItem />
            </ControlledComponent>
            <Fragment key='domains'>
              {segmentsList
                ?.filter((segment) =>
                  segment?.name
                    ?.toLowerCase()
                    .includes(searchText.toLowerCase())
                )
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
            'flex col-gap-2 items-center w-full',
            styles.sidebar_action_button
          )}
          onClick={() => {
            history.replace(PathUrls.ProfileAccounts);
            dispatch(setNewSegmentModeAction(true));
            dispatch(setAccountPayloadAction({}));
          }}
        >
          <SVG
            name='plus'
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

export default AccountsSidebar;
