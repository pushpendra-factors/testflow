import { Button, Form, Input, Modal, Slider } from 'antd';
import { Text } from 'Components/factorsComponents';
import { getGroupList } from 'Components/Profile/AccountProfiles/accountProfiles.helpers';
import EventsBlock from 'Components/Profile/MyComponents/EventsBlock';
import React, {
  useCallback,
  useEffect,
  useMemo,
  useState,
  useRef
} from 'react';
import { useSelector } from 'react-redux';
import styles from './index.module.scss';
import { PlusOutlined } from '@ant-design/icons';

function EngagementModal({ visible, onOk, onCancel, event, editMode }) {
  const sliderRef = useRef();
  // This ModalVisible state helps to get animation and
  // Unmount the Modal AntD, which completely destroys, its prev states.
  // This method is used because {event} variable is changing its behaviour
  const [modalVisible, setModalVisible] = useState(visible);
  const [listEvents, setListEvents] = useState([]);
  // Below 2 states are to track the changes in the rulename, and slider value
  const [ruleName, setRuleName] = useState('');
  const [sliderState, setSliderState] = useState(0);

  const [isEventsDDVisible, setEventsDDVisible] = useState(false);
  const availableGroups = useSelector((state) => state.coreQuery.groups);
  useEffect(() => {
    if (visible === false) {
      let handle = setTimeout(() => {
        setModalVisible(false);
        clearTimeout(handle); // Removing the handler after use
      }, 200);
      // This was added to remove the Filter pills on cancel of EngagementModal
      // This ModalVisible is used to show & hide Modal Antd with some delay, which doesn't breaks the animation
    } else {
      setModalVisible(true);
    }
  }, [visible]);
  useEffect(() => {
    if (!event || event === undefined || Object.keys(event).length === 0)
      setListEvents([]);
    else setListEvents([event]);
  }, [event]);
  useEffect(() => {
    setRuleName(event.fname);
    setSliderState(event.weight);
  }, [event.fname, event.weight]);

  const groupsList = useMemo(() => {
    return getGroupList(availableGroups?.all_groups);
  }, [availableGroups]);

  const onCancelState = () => {
    onCancel();
    // Below 2 are to reset the values to its actual values
    // if it is not changed.
    setSliderState(event.weight);
    setRuleName(event.fname);
    // setListEvents([event]); // this is not working, because values of [event] is getting changed even if we are not changing it.
  };

  const onSaveState = () => {
    const payload = { ...listEvents[0], fname: ruleName, weight: sliderState };
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
            groupAnalysis='events'
            dropdownPlacement='bottom'
            propertiesScope={['event', 'user', 'group']}
            isSpecialEvent={true}
          />
        </div>
      );
    } else if (listEvents.length === 1) {
      blockList.push(
        <div key={0}>
          <EventsBlock
            isEngagementConfig
            availableGroups={groupsList}
            index={1}
            event={listEvents[0]}
            closeEvent={closeEvent}
            queries={listEvents}
            eventChange={queryChange}
            groupAnalysis='events'
            dropdownPlacement='bottom'
            propertiesScope={['event', 'user', 'group']}
            viewMode={editMode}
            isSpecialEvent={true}
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
  const handleRuleNameChange = useCallback((e) => {
    setRuleName(e.target.value);
  }, []);
  const handleSliderChange = (value) => {
    setSliderState(value);
  };
  return modalVisible ? (
    <Modal
      title={null}
      width={750}
      visible={visible}
      footer={null}
      className={'fa-modal--regular p-6'}
      closable={false}
      centered
    >
      <div className='p-2'>
        <div className='pb-4'>
          <Text extraClass='m-0' type='title' level={4} weight='bold'>
            {editMode ? 'Edit' : 'Add New'} signal
          </Text>
          <Text extraClass='m-0' type='title' level={7} color='grey'>
            {editMode
              ? 'Define the event conditions for the signal and assign it an appropriate weight.'
              : 'Define the event conditions for the signal and assign it an appropriate weight.'}
          </Text>
        </div>
        <div>
          <Text extraClass='m-0 font-normal' type='title' level={6}>
            Signal Name
          </Text>

          <Input
            onChange={handleRuleNameChange}
            value={ruleName}
            className={styles['signal_name_input']}
            type='text'
            placeholder='Eg: Pricing page visit'
            maxLength={20}
            showCount
          />
        </div>
        <div className={`${styles['eventslist']}`}>
          {eventsList()}
          {!isEventsDDVisible && listEvents.length === 0 && (
            <div className='relative'>
              <Button
                icon={<PlusOutlined />}
                type='dashed'
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
            Assign weight
          </Text>
          <div id='tooltip-container'></div>
          <Slider
            disabled={!listEvents.length}
            onChange={handleSliderChange}
            value={sliderState}
            marks={marks}
            min={0}
            step={5}
            max={100}
            getTooltipPopupContainer={() =>
              document.querySelector('#tooltip-container')
            }
          />
        </div>
        <div className='flex flex-row-reverse justify-between'>
          <div className='inline-flex'>
            <Button
              className='mr-1 dropdown-btn'
              type='text'
              onClick={onCancelState}
            >
              Cancel
            </Button>
            <Button
              disabled={!listEvents.length || !ruleName?.length > 0}
              className='ml-1'
              type='primary'
              onClick={onSaveState}
            >
              {!editMode ? 'Add Signal' : 'Save changes'}
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
  ) : (
    ''
  );
}
export default EngagementModal;
