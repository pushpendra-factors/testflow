import React, { useState, useEffect } from 'react';
import { Modal, Button } from 'antd';
import PropertyFilter from './PropertyFilter';
import { formatFiltersForPayload } from '../utils';
import { SVG, Text } from 'Components/factorsComponents';
import InputFieldWithLabel from '../MyComponents/InputFieldWithLabel/index';
import {
  ProfileMapper,
  profileOptions,
  ReverseProfileMapper
} from 'Utils/constants';
import FaSelect from 'Components/FaSelect';

function SegmentModal({
  type,
  editMode = false,
  visible,
  segment = {},
  onSave,
  onCancel
}) {
  const [isUserDDVisible, setUserDDVisible] = useState(false);
  const [segmentPayload, setSegmentPayload] = useState({
    name: '',
    description: '',
    query: { ewp: [], gp: [] },
    type: type
  });
  const [filterProps, setFilterProps] = useState([]);

  const handleNameInput = (e) => {
    const payload = { ...segmentPayload };
    payload.name = e.target.value;
    setSegmentPayload(payload);
  };

  const handleDescInput = (e) => {
    const payload = { ...segmentPayload };
    payload.description = e.target.value;
    setSegmentPayload(payload);
  };

  const setSegmentType = (val) => {
    if ((ProfileMapper[val[0]] || val[0]) !== segmentPayload.type) {
      const opts = { ...segmentPayload };
      opts.type = ProfileMapper[val[0]] || val[0];
      setSegmentPayload(opts);
    }
    setUserDDVisible(false);
  };

  const setFilters = (filters) => {
    setFilterProps(filters);
  };

  useEffect(() => {
    const opts = { ...segmentPayload };
    const filters = formatFiltersForPayload(filterProps);
    opts.query = { gp: filters };
    setSegmentPayload(opts);
  }, [filterProps]);

  const selectUsers = () => (
    <div className='absolute top-0'>
      {isUserDDVisible ? (
        <FaSelect
          options={[['All'], ...profileOptions.users]}
          onClickOutside={() => setUserDDVisible(false)}
          optionClick={(val) => setSegmentType(val)}
        />
      ) : null}
    </div>
  );

  const renderModalHeader = () => (
    <Text extraClass='m-0 p-4' type={'title'} level={5} weight={'bold'}>
      {editMode ? 'Edit Segment' : 'New Segment'}
    </Text>
  );

  const renderNameSection = () => (
    <InputFieldWithLabel
      extraClass='px-4 pb-4'
      inputClass='fa-input'
      title='Name'
      placeholder='Segment Name'
      value={segmentPayload.name}
      onChange={handleNameInput}
    />
  );

  const renderDescSection = () => (
    <InputFieldWithLabel
      isTextArea
      extraClass='px-4 pb-4'
      inputClass='fa-input'
      title='Description'
      placeholder='Description'
      value={segmentPayload.description}
      onChange={handleDescInput}
    />
  );

  const renderQuerySection = () => (
    <div className='p-4'>
      <div className='flex items-center mb-2'>
        <Text
          type={'title'}
          level={6}
          weight={'medium'}
          extraClass={`m-0 mr-3`}
        >
          Analyse
        </Text>
        <div className='relative mr-2'>
          <Button
            type='text'
            icon={<SVG name='user_friends' size={16} />}
            onClick={() => setUserDDVisible(!isUserDDVisible)}
          >
            {ReverseProfileMapper[segmentPayload.type]?.users || 'All'}
            <SVG name='caretDown' size={16} />
          </Button>
          {selectUsers()}
        </div>
      </div>
      <div className='px-8 py-6 border-with-radius--small'>
        <div className='flex items-start'>
          <Text
            type='title'
            level={7}
            weight='medium'
            extraClass='m-0 mr-3 whitespace-no-wrap'
            lineHeight='large'
          >
            With Properties
          </Text>
          <PropertyFilter
            profileType='user'
            source={segmentPayload.type}
            filters={filterProps}
            setFilters={setFilters}
          />
        </div>
      </div>
    </div>
  );

  const renderModalFooter = () => (
    <div className={`p-6 flex flex-row-reverse justify-between`}>
      <div>
        <Button className='mr-1' type='default' onClick={() => onCancel()}>
          Cancel
        </Button>
        <Button
          className='ml-1'
          type='primary'
          onClick={() => onSave(segmentPayload)}
        >
          {editMode ? 'Save Changes' : 'Save Segments'}
        </Button>
      </div>
      {/* {editMode ? (
    <Button
      type='text'
      onClick={resetInputField}
      icon={<SVG size={16} name='trash' color={'grey'} />}
    >
      Delete Segment
    </Button>
  ) : null} */}
    </div>
  );

  return (
    <Modal
      title={null}
      width={850}
      visible={visible}
      footer={null}
      className={'fa-modal--regular p-6'}
      closable={false}
    >
      {renderModalHeader()}
      {renderNameSection()}
      {renderDescSection()}
      {renderQuerySection()}
      {renderModalFooter()}
    </Modal>
  );
}

export default SegmentModal;
