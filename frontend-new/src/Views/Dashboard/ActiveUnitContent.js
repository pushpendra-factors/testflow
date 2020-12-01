import React, { useCallback } from 'react';
import { getStateQueryFromRequestQuery } from '../CoreQuery/utils';
import ResultTab from '../EventsAnalytics/ResultTab.js';
import ResultantChart from '../CoreQuery/FunnelsResultPage/ResultantChart';
import { Text, SVG } from '../../components/factorsComponents';
import { Button, Divider, Spin } from 'antd';
import styles from './index.module.scss';
import FiltersInfo from '../CoreQuery/FiltersInfo';
import { useHistory } from 'react-router-dom';

function ActiveUnitContent({ unit, resultState, setwidgetModal, durationObj, handleDurationChange }) {

	const history = useHistory();

	let equivalentQuery;
	if (unit.query.query.query_group) {
		equivalentQuery = getStateQueryFromRequestQuery(unit.query.query.query_group[0]);
	} else {
		equivalentQuery = getStateQueryFromRequestQuery(unit.query.query);
	}

	const breakdown = [...equivalentQuery.breakdown.event, ...equivalentQuery.breakdown.global];
	const events = [...equivalentQuery.events];
	const queryType = equivalentQuery.queryType;

	const eventsMapper = {};
	const reverseEventsMapper = {};

	events.forEach((q, index) => {
		eventsMapper[`${q.label}`] = `event${index + 1}`;
		reverseEventsMapper[`event${index + 1}`] = q.label;
	});

	let content = null;

	if (queryType === 'event') {
		content = (
			<ResultTab
				queries={events.map(elem => elem.label)}
				eventsMapper={eventsMapper}
				reverseEventsMapper={reverseEventsMapper}
				breakdown={breakdown}
				queryType={queryType}
				isWidgetModal={true}
				page="totalEvents"
				durationObj={durationObj}
				handleDurationChange={handleDurationChange}
				resultState={[resultState]}
				index={0}
			/>
		);
	}

	if (queryType === 'funnel') {

		let subcontent = null;

		if (resultState.loading) {
			subcontent = (
				<div className="flex justify-center items-center w-full h-64">
					<Spin size="large" />
				</div>
			)
		}

		if (resultState.error) {
			subcontent = (
				<div className="flex justify-center items-center w-full h-64">
					Something went wrong!
				</div>
			)
		}

		if (resultState.data) {
			subcontent = (
				<ResultantChart
					isWidgetModal={true}
					queries={events.map(elem => elem.label)}
					breakdown={breakdown}
					eventsMapper={eventsMapper}
					reverseEventsMapper={reverseEventsMapper}
					durationObj={durationObj}
					handleDurationChange={handleDurationChange}
					resultState={resultState}
				/>
			)
		}

		content = (
			<>
				<FiltersInfo
					durationObj={durationObj}
					handleDurationChange={handleDurationChange}
					breakdown={breakdown}
				/>
				{subcontent}
			</>

		);
	}

	const handleEditQuery = useCallback(() => {
		history.push({
			pathname: '/core-analytics',
			state: { query: unit.query, global_search: true }
		});
	}, [history, unit])

	return (
		<div className="p-4">
			<div className="flex flex-col">
				<div className="flex justify-between items-center">
					<Text extraClass="m-0" type={'title'} level={3} weight={'bold'}>{unit.title}</Text>
					<div className="flex items-center">
						<Button onClick={handleEditQuery} style={{ display: 'flex' }} className="flex items-center mr-2" size="small">Edit Query</Button>
						<Button style={{ display: 'flex' }} className='flex items-center' size={'small'} type="text" onClick={setwidgetModal.bind(this, false)}>
							<SVG size={24} name="times"></SVG>
						</Button>
					</div>
				</div>
				<div className="flex">
					{equivalentQuery.events.map((event, index) => {
						return (
							<div key={index} className="flex items-center mr-1 mt-3">
								<div className={styles.eventCharacter}>{String.fromCharCode(index + 65)}</div>
								<div className={styles.eventName}>{event.label}</div>
							</div>
						)
					})}
				</div>
			</div>
			<Divider />
			{content}
		</div>
	);
}

export default ActiveUnitContent;
