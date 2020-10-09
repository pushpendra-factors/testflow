import React, { useState, useEffect } from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import {
  Button, Collapse, Select, Popover
} from 'antd';
import { SVG, Text } from '../factorsComponents';
import styles from './index.module.scss';
import QueryBlock from './QueryBlock';
import SeqSelector from './AnalysisSeqSelector';
import GroupBlock from './GroupBlock';

import { fetchEventNames } from '../../reducers/coreQuery/middleware';
import { fetchEventProperties, fetchUserProperties } from '../../reducers/coreQuery/services';

const { Option } = Select;

const { Panel } = Collapse;

function QueryComposer({
  queries, runQuery, eventChange, queryType,
  fetchEventNames, activeProject,
  queryOptions,
  setQueryOptions
}) {
  const [analyticsSeqOpen, setAnalyticsSeqVisible] = useState(false);

  const [filterOptions, setFilterOptions] = useState([
    {
      label: 'User Properties',
      icon: 'fav',
      values: []
    }, {
      label: 'Event Properties',
      icon: 'virtual',
      values: []
    }
  ]);

  // const [queryOptions, setQueryOptions] = useState({
  //   groupBy: {
  //     prop_category: '', // user / event
  //     property: '', // user/eventproperty
  //     prop_type: '', // categorical  /numberical
  //     eventValue: '' // event name (funnel only)
  //   },
  //   event_analysis_seq: '',
  //   session_analytics_seq: {
  //     start: 1,
  //     end: 2
  //   }
  // });

  useEffect(() => {
    if (activeProject && activeProject.id) {
      fetchEventNames(activeProject.id);
    }
  }, [activeProject, fetchEventNames]);

  useEffect(() => {
    const eventPropertyFetches = [];

    fetchUserProperties(activeProject.id, queryType).then(res => {
      convertUserPropertyData(res);
    });

    queries.forEach(ev => {
      eventPropertyFetches.push(fetchEventProperties(activeProject.id, ev.label));
    });

    Promise.all(eventPropertyFetches).then((val) => {
      convertEventPropertyData(val);
    });
  }, [queries]);

  const convertUserPropertyData = (res) => {
    const filterOpts = [...filterOptions];
    filterOpts[0].values = [];
    if (res.status === 200) {
      const data = res.data;
      Object.keys(data).forEach(key => {
        data[key].forEach(val => {
          if (!filterOpts[0].values.find(v => v.name === val)) {
            filterOpts[0].values.push([val, key]);
          }
        });
      });
    }
    setFilterOptions(filterOpts);
  };

  const convertEventPropertyData = (val) => {
    const filterOpts = [...filterOptions];
    filterOpts[1].values = [];
    val.forEach(res => {
      if (res.status === 200) {
        const data = res.data;
        Object.keys(data).forEach(key => {
          data[key].forEach(val => {
            if (!filterOpts[1].values.find(v => v.name === val)) {
              filterOpts[1].values.push([val, key]);
            }
          });
        });
      }
    });
    setFilterOptions(filterOpts);
  };

  const queryList = () => {
    const blockList = [];

    queries.forEach((event, index) => {
      blockList.push(
				<div key={index} className={styles.composer_body__query_block}>
					<QueryBlock index={index + 1} queryType={queryType} event={event} queries={queries} eventChange={eventChange}></QueryBlock>
				</div>
      );
    });

    if (queries.length < 6) {
      blockList.push(
				<div key={'init'} className={styles.composer_body__query_block}>
					<QueryBlock queryType={queryType} index={queries.length + 1} queries={queries} eventChange={eventChange}></QueryBlock>
				</div>
      );
    }

    return blockList;
  };

  const setGroupByState = (value, index, action = 'add') => {
    const options = Object.assign({}, queryOptions);
    if (action === 'add') {
      options.groupBy[index] = value;
      if (options.groupBy.length - 1 === index) {
        options.groupBy.push({
          prop_category: '', // user / event
          property: '', // user/eventproperty
          prop_type: '', // categorical  /numberical
          eventValue: '' // event name (funnel only)
        });
      }
    }

    setQueryOptions(options);
  };

  const groupByBlock = () => {
    if (queryType === 'event' && queries.length < 1) { return null; }
    if (queryType === 'funnel' && queries.length < 2) { return null; }

    return (
			<div key={0} className={'fa--query_block bordered '}>
				<GroupBlock
					filterOptions={filterOptions}
					setGroupByState={setGroupByState}
					queryType={queryType}
					groupByState={queryOptions.groupBy}
					events={queries}>

				</GroupBlock>
			</div>
    );
  };

  const setEventSequence = (value) => {
    const options = Object.assign({}, queryOptions);
    options.event_analysis_seq = value;
    setQueryOptions(options);
  };

  const setAnalysisSequence = (seq) => {
    const options = Object.assign({}, queryOptions);
    options.session_analytics_seq = seq;
    setQueryOptions(options);
  };

  const moreOptionsBlock = () => {
    if (queries.length >= 2) {
      return (
				<div className={' fa--query_block bordered '}>
					<Collapse bordered={false} expandIcon={() => { }} expandIconPosition={'right'}>
						<Panel header={<div className={'flex justify-between items-center'}>
							<Text type={'title'} level={6} weight={'bold'} extraClass={'m-0 mb-2 inline'}>More options</Text>
							<SVG name="plus" />
						</div>
						}>
							<div className={'flex justify-start items-center'}>
								<span className={'mr-2'}>
									<SVG name="sortdown" size={16} color={'purple'}></SVG>
								</span>
								<Text type={'title'} level={7} extraClass={'m-0 mr-2 inline'}>Analyse events in the</Text>
								<div>
									<Select
										style={{ width: 170 }}
										value="same_sequence" onChange={setEventSequence}
										className={'no-ant-border'}
									>
										<Option value="same_sequence"> Same Sequence</Option>
										<Option value="exact_sequence"> Exact Sequence</Option>
									</Select>
								</div>
							</div>

							<div className={'flex flex-col justify-start items-start mt-4'}>
								<div className={'flex justify-start items-center'}>
									<span className={'mr-2'}>
										<SVG name="sortdown" size={16} color={'purple'}></SVG>
									</span>
									<Text type={'title'} level={7} extraClass={'m-0 mr-2 inline'}>In Session Analytics</Text>
								</div>

								<div className={'flex justify-start items-center mt-2'}>
									<div className={styles.composer_body__session_analytics__options}>
										<Popover
											className="fa-event-popover"
											content={
												<SeqSelector
													seq={queryOptions.session_analytics_seq}
													queryCount={queries.length}
													setAnalysisSequence={setAnalysisSequence}
												/>
											}
											trigger="click"
											visible={analyticsSeqOpen}
											onVisibleChange={(visible) => setAnalyticsSeqVisible(visible)}
										>
											<Button Button type="link" className={'ml-4'} size={'small'}>
												Between &nbsp;
                                            {queryOptions.session_analytics_seq.start}
                                            &nbsp;
                                                to
                                                &nbsp;
                                            {queryOptions.session_analytics_seq.end}
											</Button>
										</Popover>
										<Text type={'paragraph'} mini weight={'thin'} extraClass={'m-0 ml-2 inline'}>happened in the same session</Text>

									</div>
								</div>
							</div>
						</Panel>
					</Collapse>
				</div>
      );
    }
  };

  const footer = () => {
    if (queryType === 'event' && queries.length < 1) { return null; }
    if (queryType === 'funnel' && queries.length < 2) { return null; } else {
      return (
				<div className={styles.composer_footer}>
					<Button><SVG name={'calendar'} extraClass={'mr-1'} />Last Week </Button>
					<Button type="primary" onClick={() => runQuery('0', true)}>Run Query</Button>
				</div>
      );
    }
  };

  return (
		<div className={styles.composer_body}>
			{queryList()}
			{groupByBlock()}
			{queryType === 'funnel' ? moreOptionsBlock() : null}
			{footer()}
		</div>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project
});

const mapDispatchToProps = dispatch => bindActionCreators({
  fetchEventNames
}, dispatch);

export default connect(mapStateToProps, mapDispatchToProps)(QueryComposer);
