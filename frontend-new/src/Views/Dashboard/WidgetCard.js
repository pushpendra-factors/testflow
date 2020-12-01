import React, {
	useRef, useEffect, useCallback, useState
} from 'react';
import { Button } from 'antd';
import { Text } from '../../components/factorsComponents';
import { RightOutlined, LeftOutlined, FullscreenOutlined } from '@ant-design/icons';
import CardContent from './CardContent';
import { useSelector } from 'react-redux';
import { initialState, formatApiData } from '../CoreQuery/utils';
import { cardClassNames } from '../../reducers/dashboard/utils';
import { getDataFromServer } from './utils';

function WidgetCard({
	unit,
	onDrop,
	setwidgetModal,
	durationObj
}) {
	const [resultState, setResultState] = useState(initialState);
	const { active_project } = useSelector(state => state.global);
	const { activeDashboardUnits } = useSelector(state => state.dashboard);

	const getData = useCallback(async (refresh = false) => {
		try {
			setResultState({
				...initialState,
				loading: true
			});

			const res = await getDataFromServer(unit.query, unit.id, unit.dashboard_id, durationObj, refresh, active_project.id);
			let queryType;

			if (unit.query.query.query_group) {
				queryType = 'event';
			} else {
				queryType = 'funnel';
			}

			if (queryType === 'funnel') {
				let resultantData;
				if (res.data.result) {
					// cached data
					resultantData = res.data.result;
				} else {
					// refreshed data
					resultantData = res.data;
				}
				setResultState({
					...initialState,
					data: resultantData
				});
			} else {
				if (res.data.result) {
					// cached data
					setResultState({
						...initialState,
						data: formatApiData(res.data.result.result_group[0], res.data.result.result_group[1])
					});
				} else {
					// refreshed data
					setResultState({
						...initialState,
						data: formatApiData(res.data.result_group[0], res.data.result_group[1])
					});
				}
			}
		} catch (err) {
			console.log(err);
			console.log(err.response);
			setResultState({
				...initialState,
				error: true
			});
		}
	}, [active_project.id, unit.query, unit.id, unit.dashboard_id, durationObj]);

	useEffect(() => {
		getData();
	}, [getData, durationObj]);

	const changeCardSize = useCallback((cardSize) => {
		const unitIndex = activeDashboardUnits.data.findIndex(au => au.id === unit.id);
		const updatedUnit = {
			...unit,
			className: cardClassNames[cardSize],
			cardSize
		};
		const newState = [...activeDashboardUnits.data.slice(0, unitIndex), updatedUnit, ...activeDashboardUnits.data.slice(unitIndex + 1)];
		onDrop(newState);
	}, [unit, activeDashboardUnits.data, onDrop]);

	const { dashboardsLoaded } = useSelector(state => state.dashboard);

	const cardRef = useRef();

	return (
		<div className={`${unit.title.split(' ').join('-')} ${unit.className} py-4 px-2 flex widget-card-top-div`} >
			<div id={`card-${unit.id}`} ref={cardRef} className={'fa-dashboard--widget-card w-full flex'}>
				<div className={'px-8 py-4 flex justify-between items-start w-full'}>
					<div className={'w-full'} >
						<div className="flex items-center justify-between">
							<div className="flex flex-col">
								<Text ellipsis type={'title'} level={5} weight={'bold'} extraClass={'m-0'}>{unit.title}</Text>
								<Text ellipsis type={'paragraph'} mini color={'grey'} extraClass={'m-0'}>{unit.description}</Text>
							</div>
							<div>
								<Button size={'large'} onClick={() => setwidgetModal({ unit, data: resultState.data })} icon={<FullscreenOutlined />} type="text" />
							</div>
						</div>
						<div className="mt-4">
							<CardContent
								unit={unit}
								resultState={resultState}
								dashboardsLoaded={dashboardsLoaded}
							/>
						</div>
					</div>
				</div>
			</div>
			<div id={`resize-${unit.id}`} className={'fa-widget-card--resize-container'}>
				<span className={'fa-widget-card--resize-contents'}>
					{unit.cardSize === 0 ? (
						<a onClick={changeCardSize.bind(this, 1)}><RightOutlined /></a>
					) : null}
					{unit.cardSize === 1 ? (
						<a onClick={changeCardSize.bind(this, 0)}><LeftOutlined /></a>
					) : null}
				</span>
			</div>
		</div>

	);
}

export default React.memo(WidgetCard);
