import React, { useEffect, useState } from 'react';
import { Checkbox, Divider, Input, List, Modal } from 'antd';
import { SVG, Text } from 'Components/factorsComponents';
import VirtualList from 'rc-virtual-list';
import { useSelector } from 'react-redux';
import {
  assignUnitsToDashboard,
  DeleteUnitFromDashboard,
  fetchActiveDashboardUnits
} from 'Reducers/dashboard/services';
import { useDispatch } from 'react-redux';
import { WIDGET_DELETED } from 'Reducers/types';
import { getQueryType } from 'Utils/dataFormatter';
import useAutoFocus from 'hooks/useAutoFocus';

const ExistingReportsModal = ({
  isReportsModalOpen,
  setIsReportsModalOpen
}) => {
  const inputReference = useAutoFocus();
  const { dashboards, activeDashboard, activeDashboardUnits } = useSelector(
    (state) => state.dashboard
  );
  const { active_project } = useSelector((state) => state.global);
  const queries = useSelector((state) => state.queries);

  const ContainerHeight = 440;
  const dispatch = useDispatch();
  const [data, setData] = useState([]);
  const [selectedQuerySet, setSelectedQuerySet] = useState(new Set());
  const [prevSelectedQuerySet, setPrevSelectedQuerySet] = useState(new Set());
  const [prevSelectedQueryMainSet, setPrevSelectedQueryMainSet] = useState(
    new Map()
  );

  useEffect(() => {
    activeDashboardUnits.data.forEach((element, index) => {
      selectedQuerySet.add(element.query_id);
      prevSelectedQuerySet.add(element.query_id);
      prevSelectedQueryMainSet.set(element.query_id, element.id);
    });

    setData(queries.data);
  }, [activeDashboardUnits, activeDashboard]);

  const handleReportItemChange = (e, id) => {
    setSelectedQuerySet((prev) => {
      let x = new Set(prev);
      if (x.has(id)) {
        x.delete(id);
      } else {
        x.add(id);
      }
      return x;
    });
  };

  const HandleSearchReportModal = (e) => {
    setData(
      queries.data.filter((ele) =>
        ele.title?.toLowerCase().includes(e.target.value.toLowerCase())
      )
    );
  };
  const handleAddReport = async () => {
    // console.log({selectedQuerySet,prevSelectedQuerySet})

    try {
      // 1. Find all New Selected Queries
      // Those are available in newSelectedQuerySet
      let newQueries = [];
      selectedQuerySet.forEach((eachUnitId) => {
        if (prevSelectedQuerySet.has(eachUnitId) === false) {
          newQueries.push(eachUnitId);
        }
      });

      let deleteQueries = [];
      // Comparing Old Set of UnitIds with Current Selected QueryIDs
      // and finding which one is not there, means those will be not needed in our Dashboard
      prevSelectedQuerySet.forEach((eachUnitId) => {
        if (selectedQuerySet.has(eachUnitId) === false) {
          deleteQueries.push(eachUnitId);
        }
      });
      // 2. Check is new
      if (newQueries.length > 0) {
        let reqBody = newQueries.map((unitid) => {
          return {
            query_id: unitid
          };
        });
        await assignUnitsToDashboard(
          active_project?.id,
          activeDashboard?.id,
          reqBody
        );
      }

      if (deleteQueries.length > 0) {
        let deletePromises = deleteQueries.map((eachDeleteId) => {
          return DeleteUnitFromDashboard(
            active_project?.id,
            activeDashboard?.id,
            prevSelectedQueryMainSet.get(eachDeleteId)
          );
        });

        await Promise.all(deletePromises);
        deleteQueries.forEach((unitId) => {
          dispatch({
            type: WIDGET_DELETED,
            payload: unitId
          });
        });
      }

      if (newQueries.length > 0 || deleteQueries.length > 0) {
        // update Redux if size is > 0
        dispatch(
          fetchActiveDashboardUnits(active_project?.id, activeDashboard?.id)
        );
      }
      setIsReportsModalOpen(false);
    } catch (error) {
      console.error(error);
    }
  };

  const queryTypeName = {
    events: 'events_cq',
    funnel: 'funnels_cq',
    channel_v1: 'campaigns_cq',
    attribution: 'attributions_cq',
    profiles: 'profiles_cq',
    kpi: 'KPI_cq'
  };

  return (
    <>
      <Modal
        width={'712px'}
        okText='Add Report'
        visible={isReportsModalOpen}
        onOk={handleAddReport}
        onCancel={() => setIsReportsModalOpen(false)}
      >
        <Text type={'title'} level={3} weight={'bold'}>
          Reports
        </Text>
        <Input
          ref={inputReference}
          onChange={HandleSearchReportModal}
          className={`fa-global-search--input fa-global-search--input-fw fa-global-search--input-bgw py-1 mb-4`}
          placeholder='Search Reports'
          prefix={<SVG name='search' size={16} color={'grey'} />}
        />

        <Text type='title' level='6' className='mb-4'>
          All Saved Reports
        </Text>
        <Divider style={{ margin: '0' }} />

        <List>
          <VirtualList
            data={data}
            height={ContainerHeight}
            itemHeight={48}
            itemKey='email'
          >
            {(item, index) => {
              const queryType = getQueryType(item.query);

              let svgName = '';
              Object.entries(queryTypeName).forEach(([k, v]) => {
                if (queryType === k) {
                  svgName = v;
                }
              });
              if (queryType === 'profiles') {
                return <></>;
              } else
                return (
                  <List.Item key={index}>
                    <Checkbox
                      onChange={(e) => handleReportItemChange(e, item.id)}
                      defaultChecked={selectedQuerySet.has(item.id)}
                    >
                      {item.title}
                    </Checkbox>

                    <div>
                      <SVG name={svgName} size={24} color={'blue'} />
                    </div>
                  </List.Item>
                );
            }}
          </VirtualList>
        </List>
      </Modal>
    </>
  );
};

export default ExistingReportsModal;
