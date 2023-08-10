import { Button, Modal, Slider } from 'antd';
import { Text } from 'Components/factorsComponents';
import { getGroupList } from 'Components/Profile/AccountProfiles/accountProfiles.helpers';
import EventsBlock from 'Components/Profile/MyComponents/EventsBlock';
import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { useSelector } from 'react-redux';

function EngagementModal({ visible, onOk, onCancel, event, editMode }) {
  const [listEvents, setListEvents] = useState([]);
  const [isEventsDDVisible, setEventsDDVisible] = useState(false);
  const groupOpts = useSelector((state) => state.groups.data);

  useEffect(() => {
    if (!event || event === undefined || Object.keys(event).length === 0)
      setListEvents([]);
    else setListEvents([event]);
  }, [event]);

  const groupsList = useMemo(() => {
    return getGroupList(groupOpts);
  }, [groupOpts]);

  const onCancelState = () => {
    onCancel();
  };

  const onSaveState = () => {
    const payload = listEvents[0];
    onOk(payload, editMode);
  };

  const closeEvent = () => setEventsDDVisible(false);

  const queryChange = useCallback(
    (newEvent, index, changeType = 'add') => {
      const queryupdated = [...listEvents];
      if (queryupdated[index]) {
        if (changeType === 'add' || changeType === 'filters_updated') {
          queryupdated[index] = newEvent;
        } else if (changeType === 'delete') {
          queryupdated.splice(index, 1);
        }
      } else {
        queryupdated.push(newEvent);
      }
      setListEvents(queryupdated);
    },
    [listEvents]
  );

  const eventsList = () => {
    const blockList = [];
    if (listEvents.length === 0 && isEventsDDVisible) {
      blockList.push(
        <div key={blockList.length}>
          <EventsBlock
            isEngagementConfig
            availableGroups={groupsList}
            index={1}
            queries={listEvents}
            eventChange={queryChange}
            closeEvent={closeEvent}
            groupAnalysis={'events'}
            dropdownPlacement='bottom'
            propertiesScope={['event', 'user']}
          />
        </div>
      );
    } else if (listEvents.length === 1) {
      blockList.push(
        <div key={0}>
          <EventsBlock
            isEngagementConfig
            disableEventEdit={editMode}
            availableGroups={groupsList}
            index={1}
            event={listEvents[0]}
            closeEvent={closeEvent}
            queries={listEvents}
            eventChange={queryChange}
            groupAnalysis={'events'}
            dropdownPlacement='bottom'
            propertiesScope={['event', 'user']}
          />
        </div>
      );
    }

    return blockList.length ? (
      <div className='segment-query_block'>
        <div className='content'>{blockList}</div>
      </div>
    ) : null;
  };

  const marks = {
    0: '0',
    10: '10',
    20: '20',
    30: '30',
    40: '40',
    50: '50',
    60: '60',
    70: '70',
    80: '80',
    90: '90',
    100: '100'
  };

  const setSliderValue = (value) => {
    const event = listEvents[0];
    event.weight = value;
    setListEvents([event]);
  };

  return (
    <Modal
      title={null}
      width={750}
      visible={visible}
      footer={null}
      className={'fa-modal--regular p-6'}
      closable={false}
      centered
    >
      <div className='p-6'>
        <div className='pb-4'>
          <Text extraClass='m-0' type='title' level={4} weight='bold'>
            {editMode ? 'Edit' : 'Add New'} Score
          </Text>
          <Text extraClass='m-0' type='title' level={7} color='grey'>
            {editMode
              ? 'Edit filters/rules for this event or assign new weights'
              : 'Find and select an event, then assign it weights'}
          </Text>
        </div>
        <div className='pb-4'>
          {eventsList()}
          {!isEventsDDVisible && listEvents.length === 0 && (
            <div className='relative'>
              <Button
                className='btn-total-round'
                type='link'
                onClick={() => setEventsDDVisible(true)}
              >
                Select Event
              </Button>
            </div>
          )}
        </div>
        <div className='pb-4'>
          <Text
            extraClass='m-0'
            type='title'
            level={7}
            weight='bold'
            color='grey'
          >
            Assign weights
          </Text>
          <Slider
            disabled={!listEvents.length}
            onChange={setSliderValue}
            value={listEvents?.[0]?.weight}
            marks={marks}
            min={0}
            max={100}
          />
        </div>
        <div className='flex flex-row-reverse justify-between'>
          <div>
            <Button className='mr-1' type='default' onClick={onCancelState}>
              Cancel
            </Button>
            <Button
              disabled={!listEvents.length}
              className='ml-1'
              type='primary'
              onClick={onSaveState}
            >
              Save
            </Button>
          </div>
          {/* <Button
            type='text'
            onClick={''}
            icon={<SVG size={16} name='trash' color={'grey'} />}
          >
            Delete Rule
          </Button> */}
        </div>
      </div>
    </Modal>
  );
}
export default EngagementModal;
