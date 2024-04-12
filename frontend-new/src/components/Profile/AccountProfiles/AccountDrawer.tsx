import React, { useEffect, useMemo, useState } from 'react';
import { Drawer, Button } from 'antd';
import { SVG, Text } from 'Components/factorsComponents';
import { AccountDrawerProps } from 'Components/Profile/types';
import { ConnectedProps, connect, useDispatch, useSelector } from 'react-redux';
import { setActivePageviewEvent } from 'Reducers/timelines/middleware';
import {
  getEventPropertiesV2,
  getUserPropertiesV2
} from 'Reducers/coreQuery/middleware';
import { bindActionCreators } from 'redux';
import AccountTimelineTableView from './AccountTimelineTableView';

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

  const { accountPreview } = useSelector((state: any) => state.timelines);
  const { eventPropertiesV2, userPropertiesV2 } = useSelector(
    (state: any) => state.coreQuery
  );
  const { active_project: activeProject } = useSelector(
    (state: any) => state.global
  );

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

  return (
    <Drawer
      title={
        <div className='flex justify-between items-center'>
          <Text type='title' level={4} weight='bold' extraClass='m-0'>
            {domain}
          </Text>
          <div className='inline-flex gap--8'>
            <Button onClick={onClickMore}>
              <div className='inline-flex gap--4'>
                <SVG name='expand' />
                Open
              </div>
            </Button>
            <Button onClick={onClickOpenNewtab} className='flex items-center'>
              <SVG name='ArrowUpRightSquare' size={16} />
            </Button>
            <Button onClick={onClose} className='flex items-center'>
              <SVG name='times' size={16} />
            </Button>
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
        />
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
