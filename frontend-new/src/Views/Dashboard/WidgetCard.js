import React, {
  useRef, useEffect, useCallback, useState
} from 'react';
import moment from 'moment';
import * as d3 from 'd3';
import { Button } from 'antd';
import { Text } from '../../components/factorsComponents';
// import { FullscreenOutlined, RightOutlined, LeftOutlined } from '@ant-design/icons';
import { RightOutlined, LeftOutlined, FullscreenOutlined } from '@ant-design/icons';
import CardContent from './CardContent';
import { useSelector, useDispatch } from 'react-redux';
import { DASHBOARD_UNIT_DATA_LOADED } from '../../reducers/types';
import { initialState, formatApiData } from '../CoreQuery/utils';
import { runQuery, getFunnelData } from '../../reducers/coreQuery/services';
import { cardClassNames } from '../../reducers/dashboard/utils';

function WidgetCard({
  unit,
  onDrop,
  setwidgetModal,
  durationObj
}) {
  const [resultState, setResultState] = useState(initialState);
  const { active_project } = useSelector(state => state.global);
  const { activeDashboardUnits } = useSelector(state => state.dashboard);
  const [resizerVisible, setResizerVisible] = useState(false);

  const dispatch = useDispatch();

  const getData = useCallback(async (refresh = false) => {
    try {
      setResizerVisible(false);
      setResultState({
        ...initialState,
        loading: true
      });

      if (unit.query.query.query_group) {
        let res;
        let queryGroup = unit.query.query.query_group;
        if (durationObj.from && durationObj.to) {
          queryGroup = queryGroup.map(elem => {
            return {
              ...elem,
              fr: moment(durationObj.from).startOf('day').utc().unix(),
              to: moment(durationObj.to).startOf('day').utc().unix()
            }
          })
        } else {
          queryGroup = queryGroup.map(elem => {
            return {
              ...elem,
              fr: moment().startOf('week').utc().unix(),
              to: moment().utc().unix()
            }
          })
        }
        if (refresh) {
          res = await runQuery(active_project.id, queryGroup);
        } else {
          res = await runQuery(active_project.id, queryGroup, { refresh: false, unit_id: unit.id, id: unit.dashboard_id });
        }
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
      } else {
        let res;
        let funnelQuery = unit.query.query;
        if (durationObj.from && durationObj.to) {
          funnelQuery = {
            ...funnelQuery,
            fr: moment(durationObj.from).startOf('day').utc().unix(),
            to: moment(durationObj.to).startOf('day').utc().unix()
          }
        } else {
          funnelQuery = {
            ...funnelQuery,
            fr: moment().startOf('week').utc().unix(),
            to: moment().utc().unix()
          }
        }
        if (refresh) {
          res = await getFunnelData(active_project.id, funnelQuery);
        } else {
          res = await getFunnelData(active_project.id, funnelQuery, { refresh: false, unit_id: unit.id, id: unit.dashboard_id });
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
  }, [active_project.id, unit.query, unit.dashboard_id, unit.id, dispatch, durationObj]);

  useEffect(() => {
    getData();
  }, [getData, durationObj]);

  const positionResizeContainer = useCallback(() => {
    // for charts to load properly and then show the expandable icons
    setTimeout(() => {
      try {
        setResizerVisible(true);
        const parentPositions = d3.select(`#card-${unit.id}`).node().getBoundingClientRect();
        d3.select(`#resize-${unit.id}`).style('left', parentPositions.right - 10 + 'px');
        const scrollTop = (window.pageYOffset !== undefined) ? window.pageYOffset : (document.documentElement || document.body.parentNode || document.body).scrollTop;
        d3.select(`#resize-${unit.id}`).style('top', parentPositions.top + (parentPositions.height / 2) - 10 + scrollTop + 'px');
      } catch (err) {
        console.log(err);
      }
    }, 1000);
  }, [unit.id]);

  const changeCardSize = useCallback((cardSize) => {
    setResizerVisible(false);
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

  useEffect(() => {
    positionResizeContainer();
  }, [dashboardsLoaded, positionResizeContainer]);

  return (
    <div className={`${unit.title} ${unit.className} py-4 px-2`} >
      <div id={`card-${unit.id}`} ref={cardRef} className={'fa-dashboard--widget-card w-full'}>
        {resizerVisible ? (
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
        ) : null}
        <div className={'fa-widget-card--top flex justify-between items-start'}>
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
    </div>

  );
}

export default React.memo(WidgetCard);
