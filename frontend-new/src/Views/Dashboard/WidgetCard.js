import React, {
  useRef, useEffect, useCallback, useState
} from 'react';
import * as d3 from 'd3';
import { Text } from '../../components/factorsComponents';
// import { FullscreenOutlined, RightOutlined, LeftOutlined } from '@ant-design/icons';
import { RightOutlined, LeftOutlined } from '@ant-design/icons';
import CardContent from './CardContent';
import { useSelector, useDispatch } from 'react-redux';
import { CARD_SIZE_CHANGED, DASHBOARD_UNIT_DATA_LOADED } from '../../reducers/types';
import { initialState } from '../CoreQuery/utils';
import { runQuery, getFunnelData } from '../../reducers/coreQuery/services';

function WidgetCard({
  unit
}) {
  const [resultState, setResultState] = useState(initialState);
  const { active_project } = useSelector(state => state.global);
  const dispatch = useDispatch();

  const getData = useCallback(async (refresh = false) => {
    try {
      setResultState({
        ...initialState,
        loading: true
      });

      if (unit.query.query.query_group) {
        let res;
        if (refresh) {
          res = await runQuery(active_project.id, unit.query.query.query_group);
        } else {
          res = await runQuery(active_project.id, unit.query.query.query_group, { refresh: false, unit_id: unit.id, id: unit.dashboard_id });
        }
        let resultantData = null;
        if (res.data.result) {
          // cached data
          resultantData = res.data.result.result_group[0];
        } else {
          // refreshed data
          resultantData = res.data.result_group[0];
        }
        setResultState({
          ...initialState,
          data: resultantData
        });
      } else {
        let res;
        if (refresh) {
          res = await getFunnelData(active_project.id, unit.query.query);
        } else {
          res = await getFunnelData(active_project.id, unit.query.query, { refresh: false, unit_id: unit.id, id: unit.dashboard_id });
        }
        let resultantData = null;
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
        setTimeout(() => {
          dispatch({ type: DASHBOARD_UNIT_DATA_LOADED, payload: unit.id });
        }, 1000);
      }
    } catch (err) {
      console.log(err);
      console.log(err.response);
      setResultState({
        ...initialState,
        error: true
      });
    }
  }, [active_project.id, unit.query, unit.dashboard_id, unit.id, dispatch]);

  useEffect(() => {
    getData();
  }, [getData]);

  const positionResizeContainer = useCallback(() => {
    const parentPositions = d3.select(`#card-${unit.id}`).node().getBoundingClientRect();
    d3.select(`#resize-${unit.id}`).style('left', parentPositions.right - 10 + 'px');
    const scrollTop = (window.pageYOffset !== undefined) ? window.pageYOffset : (document.documentElement || document.body.parentNode || document.body).scrollTop;
    d3.select(`#resize-${unit.id}`).style('top', parentPositions.top + (parentPositions.height / 2) - 10 + scrollTop + 'px');
  }, [unit.id]);

  const changeCardSize = useCallback((cardSize) => {
    dispatch({ type: CARD_SIZE_CHANGED, payload: { unit, cardSize } });
  }, [unit, dispatch]);

  const { dashboards_loaded } = useSelector(state => state.dashboard);

  const cardRef = useRef();

  useEffect(() => {
    setTimeout(() => {
      positionResizeContainer();
    }, 0);
  }, [dashboards_loaded, positionResizeContainer]);

  return (
		<div className={`${unit.title} ${unit.className} py-4 px-2`} >
			<div id={`card-${unit.id}`} ref={cardRef} style={{ transition: 'all 0.1s' }} className={'fa-dashboard--widget-card w-full'}>
				<div id={`resize-${unit.id}`} className={'fa-widget-card--resize-container'}>
					<span className={'fa-widget-card--resize-contents'}>
						{unit.cardSize === 'half-page' ? (
							<a onClick={changeCardSize.bind(this, 'full-page')}><RightOutlined /></a>
						) : null}
						{unit.cardSize === 'full-page' ? (
							<a onClick={changeCardSize.bind(this, 'half-page')}><LeftOutlined /></a>
						) : null}
					</span>
				</div>
				<div className={'fa-widget-card--top flex justify-between items-start'}>
					<div className={'w-full'} >
						<Text ellipsis type={'title'} level={5} weight={'bold'} extraClass={'m-0'}>{unit.title}</Text>
						<Text ellipsis type={'paragraph'} mini color={'grey'} extraClass={'m-0'}>{unit.description}</Text>
						<div className="mt-4">
							<CardContent
								unit={unit}
								resultState={resultState}
								dashboards_loaded={dashboards_loaded}
							/>
						</div>
					</div>
					{/* <div className={'flex flex-col justify-start items-start fa-widget-card--top-actions'}>
						<Button size={'large'} onClick={() => setwidgetModal(true)} icon={<FullscreenOutlined />} type="text" />
					</div> */}
				</div>
			</div>
		</div>

  );
}

export default React.memo(WidgetCard);
