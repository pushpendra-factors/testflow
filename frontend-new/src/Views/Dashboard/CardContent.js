import React from 'react';
import { Spin } from 'antd';
import { getStateQueryFromRequestQuery, presentationObj } from '../CoreQuery/utils';
import EventsAnalytics from './EventsAnalytics';
import Funnels from './Funnels';

function CardContent({ unit, resultState }) {

	let content = null;

	if (resultState.loading) {
		content = (
			<div className="flex justify-center items-center w-full h-64">
				<Spin size="small" />
			</div>
		);
	}

	if (resultState.error) {
		content = (
			<div className="flex justify-center items-center w-full h-64">
				Something went wrong!
			</div>
		);
	}

	if (resultState.data) {
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

		let dashboardPresentation = 'pl';

		try {
			dashboardPresentation = unit.settings.chart;
		} catch (err) {
			console.log(err);
		}

		if (queryType === 'funnel') {
			content = (
				<Funnels
					breakdown={breakdown}
					events={events.map(elem => elem.label)}
					resultState={resultState}
					chartType={presentationObj[dashboardPresentation]}
					title={unit.id}
					eventsMapper={eventsMapper}
					reverseEventsMapper={reverseEventsMapper}
				/>
			);
		}

		if (queryType === 'event') {
			content = (
				<EventsAnalytics
					breakdown={breakdown}
					events={events.map(elem => elem.label)}
					resultState={resultState}
					chartType={presentationObj[dashboardPresentation]}
					title={unit.id}
					eventsMapper={eventsMapper}
					reverseEventsMapper={reverseEventsMapper}
				/>
			);
		}
	}

	return (
		<>
			{content}
		</>
	);
}

export default CardContent;
