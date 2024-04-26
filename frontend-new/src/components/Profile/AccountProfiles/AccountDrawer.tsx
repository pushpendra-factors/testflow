import React, { useEffect, useMemo, useState } from 'react';
import { Drawer, Button, Tooltip } from 'antd';
import { SVG, Text } from 'Components/factorsComponents';
import { AccountDrawerProps } from 'Components/Profile/types';
import { ConnectedProps, connect, useSelector } from 'react-redux';
import { setActivePageviewEvent } from 'Reducers/timelines/middleware';
import {
  getEventPropertiesV2,
  getUserPropertiesV2
} from 'Reducers/coreQuery/middleware';
import { bindActionCreators } from 'redux';
import AccountTimelineTableView from './AccountTimelineTableView';
import { placeholderIcon } from '../constants';
import { getHost } from '../utils';

function AccountDrawer({
  domain,
  visible,
  onClose,
  onClickMore,
  onClickOpenNewtab,
  setActivePageviewEvent,
  getEventPropertiesV2
}: ComponentProps): JSX.Element {
  const [requestedEvents, setRequestedEvents] = useState<{
    [key: string]: boolean;
  }>({});
  const [eventPropertiesType, setEventPropertiesType] = useState<{
    [key: string]: string;
  }>({});
  const [userPropertiesType, setUserPropertiesType] = useState<{
    [key: string]: string;
  }>({});
  const [eventDrawerVisible, setEventDrawerVisible] = useState(false);
  const [scrollPercent, setScrollPercent] = useState<number>(0);

  const { accountPreview } = useSelector((state: any) => state.timelines);
  const { eventPropertiesV2, userPropertiesV2 } = useSelector(
    (state: any) => state.coreQuery
  );
  const { active_project: activeProject } = useSelector(
    (state: any) => state.global
  );

  useEffect(() => {
    setScrollPercent(0);
  }, [domain]);

  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape' && eventDrawerVisible) {
        setEventDrawerVisible(false);
      } else if (event.key === 'Escape' && visible) {
        onClose();
      }
    };

    document.addEventListener('keydown', handleKeyDown);

    return () => {
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [visible, eventDrawerVisible]);

  const uniqueEventNames: string[] = useMemo(() => {
    const accountEvents = accountPreview?.[domain]?.events || [];
    const eventsArray = accountEvents
      .filter((event: any) => event.display_name !== 'Page View')
      .map((event: any) => event.name);

    const pageViewEvent = accountEvents.find(
      (event: any) => event.display_name === 'Page View'
    );

    if (pageViewEvent) {
      eventsArray.push(pageViewEvent.name);
      setActivePageviewEvent(pageViewEvent.name);
    }

    return Array.from(new Set(eventsArray));
  }, [accountPreview, domain]);

  const fetchEventPropertiesWithType = async () => {
    const promises = uniqueEventNames.map(async (eventName: string) => {
      if (!requestedEvents[eventName]) {
        setRequestedEvents((prevRequestedEvents) => ({
          ...prevRequestedEvents,
          [eventName]: true
        }));
        if (!eventPropertiesV2[eventName]) {
          await getEventPropertiesV2(activeProject?.id, eventName);
        }
      }
    });

    await Promise.allSettled(promises);
    const typeMap: { [key: string]: string } = {};
    Object.values(eventPropertiesV2).forEach((propertyGroup) => {
      Object.values(propertyGroup || {}).forEach((arr) => {
        arr.forEach(([, propName, category]: any) => {
          typeMap[propName] = category;
        });
      });
    });
    setEventPropertiesType(typeMap);
  };

  useEffect(() => {
    fetchEventPropertiesWithType();
  }, [uniqueEventNames, requestedEvents, activeProject?.id, eventPropertiesV2]);

  useEffect(() => {
    if (!userPropertiesV2) {
      getUserPropertiesV2(activeProject?.id);
    } else {
      const typeMap: { [key: string]: string } = {};
      Object.values(userPropertiesV2).forEach((arr: any) => {
        arr.forEach(([, propName, category]: any) => {
          typeMap[propName] = category;
        });
      });
      setUserPropertiesType(typeMap);
    }
  }, [userPropertiesV2, activeProject?.id]);

  const showMoreButton = useMemo(
    () => accountPreview?.[domain]?.events?.length > 99 && scrollPercent > 99,
    [accountPreview, domain, scrollPercent]
  );

  return (
    <Drawer
      title={
        <div className='flex justify-between items-center'>
          <div className='inline-flex gap--8'>
            <img
              src={`https://logo.clearbit.com/${getHost(domain)}`}
              onError={(e) => {
                if (e.target.src !== placeholderIcon) {
                  e.target.src = placeholderIcon;
                }
              }}
              alt=''
              width='32'
              height='32'
              loading='lazy'
            />
            <Text type='title' level={4} weight='bold' extraClass='m-0'>
              {domain}
            </Text>
          </div>
          <div className='inline-flex gap--8'>
            <Button onClick={onClickMore}>
              <div className='inline-flex gap--4'>
                <SVG name='expand' />
                Open
              </div>
            </Button>
            <Tooltip title='Open in new tab' placement='bottom'>
              <Button onClick={onClickOpenNewtab} className='flex items-center'>
                <SVG name='ArrowUpRightSquare' size={16} />
              </Button>
            </Tooltip>
            <Tooltip title='Close[Esc]' placement='bottom'>
              <Button onClick={onClose} className='flex items-center'>
                <SVG name='times' size={16} />
              </Button>
            </Tooltip>
          </div>
        </div>
      }
      placement='right'
      closable={false}
      mask={false}
      visible={visible}
      className='fa-account-drawer--right'
      onClose={onClose}
      width='large'
    >
      <div>
        <AccountTimelineTableView
          timelineEvents={accountPreview?.[domain]?.events || []}
          loading={accountPreview?.[domain]?.loading}
          eventPropsType={eventPropertiesType}
          userPropsType={userPropertiesType}
          eventDrawerVisible={eventDrawerVisible}
          setEventDrawerVisible={setEventDrawerVisible}
          hasScrollAction
          setScrollPercent={setScrollPercent}
          isPreview
        />
        {showMoreButton && (
          <div className='see-more-section'>
            <Button onClick={onClickMore}>
              <div className='inline-flex gap--4'>
                <SVG name='expand' />
                Open to see more
              </div>
            </Button>
          </div>
        )}
      </div>
    </Drawer>
  );
}

const mapDispatchToProps = (dispatch: any) =>
  bindActionCreators(
    {
      setActivePageviewEvent,
      getEventPropertiesV2
    },
    dispatch
  );

const connector = connect(null, mapDispatchToProps);
type ReduxProps = ConnectedProps<typeof connector>;
type ComponentProps = ReduxProps & AccountDrawerProps;

export default connector(AccountDrawer);
