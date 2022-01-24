/* eslint-disable */
import React, { useState, useRef, useCallback } from 'react';
import { Modal, Input, Tag } from 'antd';
import styles from './index.module.scss';
import { SVG, Text } from '../factorsComponents';
import { useSelector } from 'react-redux';
import { useHistory } from 'react-router-dom';
import { getQueryType } from '../../utils/dataFormatter';
import { QUERY_TYPE_WEB, QUERY_TYPE_TEXT } from '../../utils/constants';

function SearchModal({ visible, handleClose, handleQueryClick }) {
  const queriesState = useSelector((state) => state.queries);
  const history = useHistory();

  const inputEl = useRef(null);
  //   const [focused, setFocused] = useState(false);
  const [searchValue, setSearchValue] = useState('');

  const handleChange = useCallback((e) => {
    setSearchValue(e.target.value);
  }, []);

  const data = queriesState.data
    .filter((q) => !(q.query && q.query.cl === QUERY_TYPE_WEB))
    .filter((q) => q.title.toLowerCase().includes(searchValue.toLowerCase()));

  //   useEffect(() => {
  //       console.log('inputEl',inputEl.current.focus)
  //   }, []);

  return (
    <Modal
      centered={true}
      visible={visible}
      width={700}
      title={null}
      className={`fa-modal--regular fa-modal--slideInDown ${styles.modal} fa-global-search--modal`}
      okText={'Save'}
      confirmLoading={false}
      closable={false}
      onCancel={() => {
        setSearchValue('');
        handleClose();
      }}
      footer={false}
      transitionName=''
      maskTransitionName=''
      mask={false}
    >
      <div data-tour = 'step-3' className='search-bar'>
        <div className='flex justify-center px-4'>
          <Input
            value={searchValue}
            onChange={handleChange}
            ref={inputEl}
            // onFocus={handleFocus}
            autoFocus
            // className={`${styles.inputBox} ${focused ? styles.focused : ''}`}
            className={`fa-global-search--input fa-global-search--input-fw py-4 mt-4`}
            placeholder='Search Reports'
            prefix={<SVG name='search' size={16} color={'grey'} />}
          />
        </div>

        {data.length ? (
          <div className='search-list pb-4 fa-global-search--contents'>
            <div className={`p-4 ${styles.searchHeadings}`}>Saved Reports</div>
            <div className='fa-global-search--contents'>
              {data.map((d) => {
                const queryType = getQueryType(d.query);
                const queryTypeName = {
                  events: 'events_cq',
                  funnel: 'funnels_cq',
                  channel_v1: 'campaigns_cq',
                  attribution: 'attributions_cq',
                  profiles: 'profiles_cq',
                  kpi: 'KPI_cq',
                };
                let svgName = '';
                Object.entries(queryTypeName).forEach(([k, v]) => {
                  if (queryType === k) {
                    svgName = v;
                  }
                });

                return (
                  <div
                    onClick={() => handleQueryClick(d)}
                    className={`flex justify-between items-center px-4 py-3 cursor-pointer ${styles.queryItem}`}
                    key={d.id}
                  >
                    <div className='flex items-center'>
                      <div className='mr-2'>
                        <SVG name={svgName} size={20} />
                      </div>
                      <Text
                        type={'title'}
                        level={7}
                        extraClass={`m-0 ${styles.hoverTextColor}`}
                      >
                        {d.title}
                      </Text>
                    </div>
                    <div className={styles.queryType}>
                      <Tag style={{ borderRadius: '4px' }}>
                        {QUERY_TYPE_TEXT[queryType]}
                      </Tag>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        ) : null}

        {!data.length ? (
          <div className='search-list pb-2'>
            <div className={'p-4 flex '}>
              <Text
                type={'title'}
                level={7}
                weight={'bold'}
                align={'center'}
                extraClass={'m-0 ml-1'}
              >
                No Matches.
              </Text>
              <Text
                type={'title'}
                color={'grey'}
                level={7}
                align={'center'}
                extraClass={'m-0 ml-1'}
              >
                What kind of analysis are you looking for?
              </Text>
            </div>
            <div className='flex px-4 py-2'>
              <div className='w-1/2 pr-1'>
                <div
                  onClick={() => history.push('/analyse')}
                  className={`flex flex-col cursor-pointer py-4 px-4 justify-center rounded ${styles.boxStyles}`}
                >
                  <div className='flex justify-center items-center'>
                    <SVG size={40} name={'corequery_colored'} />
                  </div>
                  <Text
                    weight={'bold'}
                    type={'title'}
                    align={'center'}
                    level={5}
                    extraClass={'m-0'}
                  >
                    Run a Core Query
                  </Text>
                  <Text
                    type={'title'}
                    color={'grey'}
                    lineHeight={'medium'}
                    level={7}
                    align={'center'}
                    extraClass={'m-0 mt-2'}
                  >
                    Get to the bottom of User Behaviors, Funnels and Marketing
                    Campaigns.
                  </Text>
                </div>
              </div>
              <div className='w-1/2 pl-1'>
                <div
                  onClick={() => history.push('/explain')}
                  className={`flex flex-col cursor-pointer py-4 px-4 justify-center rounded ${styles.boxStyles}`}
                >
                  <div className='flex justify-center items-center'>
                    <SVG size={40} name={'factors_colored'} />
                  </div>
                  <Text
                    weight={'bold'}
                    type={'title'}
                    align={'center'}
                    level={5}
                    extraClass={'m-0'}
                  >
                    Find Key Factors
                  </Text>
                  <Text
                    type={'title'}
                    color={'grey'}
                    level={7}
                    lineHeight={'medium'}
                    align={'center'}
                    extraClass={'m-0 mt-2'}
                  >
                    Discover factors unknown to you that might be affecting
                    users or events.
                  </Text>
                </div>
              </div>
            </div>
          </div>
        ) : null}
      </div>
    </Modal>
  );
}

export default SearchModal;
