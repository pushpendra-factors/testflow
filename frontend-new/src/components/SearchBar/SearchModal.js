import React, { useState, useRef, useCallback } from 'react';
import { Modal, Input, Button } from 'antd';
import styles from './index.module.scss';
import { SVG, Text } from '../factorsComponents';
import { useSelector } from 'react-redux';
import { useHistory } from 'react-router-dom';

function SearchModal({ visible, handleClose, handleQueryClick }) {
  const queriesState = useSelector(state => state.queries);
  const history = useHistory();

  const inputRef = useRef(null);
  const [focused, setFocused] = useState(false);
  const [searchValue, setSearchValue] = useState('');

  const handleFocus = () => {
    setFocused(true);
  };

  const handleChange = useCallback((e) => {
    setSearchValue(e.target.value);
  }, []);

  const data = queriesState.data
    .filter(q => parseInt(q.type) === 2)
    .filter(q => q.title.toLowerCase().includes(searchValue.toLowerCase()));

  return (
        <Modal
            centered={true}
            visible={visible}
            width={700}
            title={null}
            className={`fa-modal--regular ${styles.modal} fa-modal--slideInDown`}
            okText={'Save'}
            confirmLoading={false}
            closable={false}
            onCancel={handleClose}
            footer={false}
            transitionName=""
            maskTransitionName=""
        >
            <div className="search-bar">
                <Input
                    value={searchValue}
                    onChange={handleChange}
                    ref={inputRef}
                    onFocus={handleFocus}
                    className={`${styles.inputBox} ${focused ? styles.focused : ''}`}
                    placeholder="Lookup factors.ai"
                    prefix={(<SVG name="search" size={24} color="black" />)}
                />

                {data.length ? (
                    <div className="search-list pb-4">
                        <div className={`p-4 ${styles.searchHeadings}`}>Saved Queries</div>
                        <div>
                            {data.map(d => {
                              let svgName = 'funnels_cq';
                              const requestQuery = d.query;
                              if (requestQuery.query_group) {
                                svgName = 'events_dashboard_cq';
                              }
                              return (
                                    <div onClick={() => handleQueryClick(d)} className={`flex justify-between items-center px-4 py-3 cursor-pointer ${styles.queryItem}`} key={d.id}>
                                        <div className="flex items-center">
                                            <div className="mr-2"><SVG name={svgName} size={24} /></div>
                                            <Text extraClass={styles.hoverTextColor} type={'paragraph'} weight={'thin'}>{d.title}</Text>
                                        </div>
                                        <div className={styles.queryType}>
                                            <Button
                                                style={{ padding: '0px 7px' }}
                                                size="small"
                                            >
                                                {svgName === 'events_dashboard_cq' ? 'Event' : 'Funnel'} Query
                                        </Button>
                                        </div>
                                    </div>
                              );
                            })}
                        </div>
                    </div>
                ) : null}

                {!data.length ? (
                    <div className="search-list pb-2">
                        <div className={'p-4'}><span className="font-bold">No Matches.</span> <span style={{ color: '#0E2647' }}>What kind of analysis are you looking for?</span></div>
                        <div className="flex px-4 py-2">
                            <div className="w-1/2 pr-1">
                                <div onClick={() => history.push('/analyse')} className={`flex flex-col cursor-pointer py-6 px-4 justify-center rounded ${styles.boxStyles}`}>
                                    <div className="flex justify-center items-center">
                                        <SVG name={'corequery_colored'} />
                                    </div>
                                    <div className={'flex justify-center font-bold text-xl leading-6 mt-2'}>
                                        Run a Core Query
                                    </div>
                                    <div className={styles.explanatoryText}>Get to the bottom of User Behaviors, Funnels and Marketing Campaigns.</div>
                                </div>
                            </div>
                            <div className="w-1/2 pl-1">
                                <div onClick={() => history.push('/explain')} className={`flex flex-col cursor-pointer py-6 px-4 justify-center rounded ${styles.boxStyles}`}>
                                    <div className="flex justify-center items-center">
                                        <SVG name={'factors_colored'} />
                                    </div>
                                    <div className={'flex justify-center font-bold text-xl leading-6 mt-2'}>
                                        Find Key Factors
                                    </div>
                                    <div className={styles.explanatoryText}>Discover factors unknown to you that might be affecting users or events. </div>
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
