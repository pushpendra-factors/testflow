import React, { useState, useCallback } from 'react';
import { connect } from 'react-redux';
import moment from 'moment';
import FunnelsResultPage from './FunnelsResultPage';
import QueryComposer from '../../components/QueryComposer';
import CoreQueryHome from '../CoreQueryHome';
import { Drawer, Button } from 'antd';
import { SVG, Text } from '../../components/factorsComponents';
import EventsAnalytics from '../EventsAnalytics';
import { runQuery as runQueryService } from '../../reducers/coreQuery/services';
import { initialResultState, calculateFrequencyData, calculateActiveUsersData } from './utils';

function CoreQuery({ activeProject }) {
	const [drawerVisible, setDrawerVisible] = useState(false);
	const [queryType, setQueryType] = useState('event');
	const [activeKey, setActiveKey] = useState('1');
	const [showResult, setShowResult] = useState(false);
	const [queries, setQueries] = useState([]);
	const [appliedQueries, setAppliedQueries] = useState([]);
	const [breakdown] = useState([]);
	const [queryOptions, setQueryOptions] = useState({
		groupBy: [{
			prop_category: '', // user / event
			property: '', // user/eventproperty
			prop_type: '', // categorical  /numberical
			eventValue: '' // event name (funnel only)
		}],
		event_analysis_seq: '',
		session_analytics_seq: {
			start: 1,
			end: 2
		}
	});
	const [resultState, setResultState] = useState(initialResultState);

	const queryChange = (newEvent, index, changeType = 'add') => {
		const queryupdated = [...queries];
		if (queryupdated[index]) {
			if (changeType === 'add') {
				queryupdated[index] = newEvent;
			} else {
				queryupdated.splice(index, 1);
			}
		} else {
			queryupdated.push(newEvent);
		}
		setQueries(queryupdated);
	};

	const getEventsWithProperties = useCallback(() => {
		const ewps = [];
		queries.forEach(ev => {
			ewps.push({
				na: ev.label,
				pr: []
			});
		});
		return ewps;
	}, [queries]);

	const getQuery = useCallback((activeTab) => {
		const query = {};
		query.cl = queryType === 'event' ? 'events' : 'funnel';
		query.ty = parseInt(activeTab) === 1 ? 'unique_users' : 'events_occurrence';
		query.ec = 'each_given_event';

		// Check date range validity

		// let period = getQueryPeriod(this.state.resultDateRange[0], this.state.timeZone)

		const period = {
			from: moment().subtract(6, 'days').startOf('day').utc().unix(),
			to: moment().utc().unix()
		};

		query.fr = period.from;
		query.to = period.to;

		if (activeTab === '2') {
			query.ewp = [
				{
					"na": "$session",
					"pr": []
				}
			];
			query.gbt = '';
		} else {
			query.ewp = getEventsWithProperties();
			query.gbt = 'date';
		}


		query.gbp = [];
		// queryOptions.groupBy.forEach(opt => {
		//   const group = {
		//     pr: opt.property,
		//     en: opt.prop_category,
		//     pty: opt.prop_type
		//   };
		//   query.gbp.push(group);
		// });

		// for(let i=0; i < this.state.groupBys.length; i++) {
		//   let groupBy = this.state.groupBys[i];
		//   let cGroupBy = {};

		//   if (groupBy.name != '' && groupBy.type != '') {
		//     cGroupBy.pr = groupBy.name;
		//     cGroupBy.en = groupBy.type;
		//     cGroupBy.pty = groupBy.ptype

		//     // add group by event name.
		//     if (this.isEventNameRequiredForGroupBy() && groupBy.eventName != '') {
		//       let nameWithIndex = removeIndexIfExistsFromOptName(groupBy.eventName);
		//       cGroupBy.ena = nameWithIndex.name
		//       // let eni = getIndexIfExistsFromOptName(groupBy.eventName)
		//       if (!isNaN(nameWithIndex.index)) {
		//         cGroupBy.eni = nameWithIndex.index  // 1 valued index to distinguish in backend from default 0.
		//       }
		//     }
		//     query.gbp.push(cGroupBy)
		//   }
		// }

		// query.gbt = (presentation == PRESENTATION_LINE) ?
		//   getGroupByTimestampType(query.fr, query.to) : '';
		query.tz = 'Asia/Kolkata';

		// query.sse = sessionStartEvent.value
		// query.see = sessionEndEvent.value

		return query;
	}, [getEventsWithProperties, queryType]);

	const closeDrawer = () => {
		setDrawerVisible(false);
	};

	const setExtraOptions = (options) => {
		setQueryOptions(options);
	};

	const updateResultState = useCallback((activeTab, newState) => {
		const idx = parseInt(activeTab);
		setResultState(currState => {
			return currState.map((elem, index) => {
				if (index === idx) {
					return newState;
				}
				return elem;
			})
		})
	}, []);

	const callRunQueryApiService = useCallback(async (activeProjectId, activeTab) => {
		try {
			const query = getQuery(activeTab);
			const res = await runQueryService(activeProjectId, [query]);
			if (res.status === 200) {
				if(activeTab !== '2') {
					updateResultState(activeTab, { loading: false, error: false, data: res.data });
				}
				return res.data;
			} else {
				updateResultState(activeTab, { loading: false, error: true, data: null });
				return null;
			}
		} catch (err) {
			updateResultState(activeTab, { loading: false, error: true, data: null });
			return null;
		}
	}, [updateResultState, getQuery])

	const runQuery = useCallback(async (activeTab, refresh = false) => {
		setActiveKey(activeTab);

		if (!refresh) {
			if (resultState[parseInt(activeTab)].data) {
				return false;
			}

			if (activeTab === '2') {
				let activeUsersData = null;
				updateResultState(activeTab, { loading: true, error: false, data: null });

				if (resultState[1].data) {
					const res = await callRunQueryApiService(activeProject.id, '2');
					if (res) {
						activeUsersData = calculateActiveUsersData(resultState[1].data, res);
					}
				} else {
					const userData = await callRunQueryApiService(activeProject.id, '1');
					const sessionData = await callRunQueryApiService(activeProject.id, '2');
					if (userData && sessionData) {
						activeUsersData = calculateActiveUsersData(userData, sessionData);
					}
				}

				updateResultState(activeTab, { loading: false, error: false, data: activeUsersData });
				return false;
			}

			if (activeTab === '3') {
				let frequencyData = null;
				if (resultState[1].data) {
					frequencyData = calculateFrequencyData(resultState[0].data, resultState[1].data);
				} else {
					updateResultState(activeTab, { loading: true, error: false, data: null });
					const res = await callRunQueryApiService(activeProject.id, '1');
					if (res) {
						frequencyData = calculateFrequencyData(resultState[0].data, res);
					}
				}
				updateResultState(activeTab, { loading: false, error: false, data: frequencyData });
				return false;
			}

		} else {
			const obj = { loading: false, error: false, data: null };
			updateResultState('1', obj);
			updateResultState('2', obj);
			updateResultState('3', obj);
			setAppliedQueries(queries.map(elem => elem.label))
		}

		closeDrawer();
		setShowResult(true);
		updateResultState(activeTab, { loading: true, error: false, data: null });
		callRunQueryApiService(activeProject.id, activeTab)
	}, [activeProject, resultState, queries, updateResultState, callRunQueryApiService]);

	const title = () => {
		return (
			<div className={'flex justify-between items-center'}>
				<div className={'flex'}>
					<SVG name={queryType === 'funnel' ? 'funnels_cq' : 'events_cq'} size="24px"></SVG>
					<Text type={'title'} level={4} weight={'bold'} extraClass={'ml-2 m-0'}>{queryType === 'funnel' ? 'Find event funnel for' : 'Analyse Events'}</Text>
				</div>
				<div className={'flex justify-end items-center'}>
					<Button type="text"><SVG name="play"></SVG>Help</Button>
					<Button type="text" onClick={() => closeDrawer()}><SVG name="times"></SVG></Button>
				</div>
			</div>
		);
	};

	const eventsMapper = {};
	const reverseEventsMapper = {};

	appliedQueries.forEach((q, index) => {
		eventsMapper[`${q}`] = `event${index + 1}`;
		reverseEventsMapper[`event${index + 1}`] = q;
	});

	let result = (
		<EventsAnalytics
			queries={appliedQueries}
			eventsMapper={eventsMapper}
			reverseEventsMapper={reverseEventsMapper}
			breakdown={breakdown}
			resultState={resultState}
			setDrawerVisible={setDrawerVisible}
			runQuery={runQuery}
			activeKey={activeKey}
		/>
	);

	if (queryType === 'funnel') {
		result = (
			<FunnelsResultPage
				setDrawerVisible={setDrawerVisible}
				queries={appliedQueries}
				eventsMapper={eventsMapper}
				reverseEventsMapper={reverseEventsMapper}
			/>
		);
	}

	return (
		<>
			<Drawer
				title={title()}
				placement="left"
				closable={false}
				visible={drawerVisible}
				onClose={closeDrawer}
				getContainer={false}
				width={'600px'}
				className={'fa-drawer'}
			>

				<QueryComposer
					queries={queries}
					runQuery={runQuery}
					eventChange={queryChange}
					queryType={queryType}
					queryOptions={queryOptions}
					setQueryOptions={setExtraOptions}
				/>
			</Drawer>

			{showResult ? (
				<>
					{result}
				</>

			) : (
					<CoreQueryHome setQueryType={setQueryType} setDrawerVisible={setDrawerVisible} />
				)}

		</>
	);
}

const mapStateToProps = (state) => ({
	activeProject: state.global.active_project
});

export default connect(mapStateToProps)(CoreQuery);
