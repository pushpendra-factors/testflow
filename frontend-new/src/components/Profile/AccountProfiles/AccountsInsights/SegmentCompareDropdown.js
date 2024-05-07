import React, { useMemo } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { Select } from 'antd';
import {
  selectAccountPayload,
  selectInsightsCompareSegmentBySegmentId
} from 'Reducers/accountProfilesView/selectors';
import { setInsightsCompareSegment } from 'Reducers/accountProfilesView/actions';
import { selectSegments } from 'Reducers/timelines/selectors';
import { GROUP_NAME_DOMAINS } from 'Components/GlobalFilter/FilterWrapper/utils';
import { reorderDefaultDomainSegmentsToTop } from '../accountProfiles.helpers';

function SegmentCompareDropdown() {
  const dispatch = useDispatch();
  const accountPayload = useSelector((state) => selectAccountPayload(state));
  const segments = useSelector(selectSegments);
  const compareSegmentId = useSelector((state) =>
    selectInsightsCompareSegmentBySegmentId(state, accountPayload.segment.id)
  );
  const segmentsList = useMemo(
    () => reorderDefaultDomainSegmentsToTop(segments[GROUP_NAME_DOMAINS]) || [],
    [segments]
  );

  const onChange = (value) => {
    dispatch(setInsightsCompareSegment(accountPayload.segment.id, value));
  };

  const options = useMemo(
    () =>
      segmentsList
        .filter((elem) => elem.id !== accountPayload.segment.id)
        .map((elem) => ({
          value: elem.id,
          label: elem.name
        })),
    [segmentsList, accountPayload.segment.id]
  );

  return (
    <Select
      showSearch
      placeholder='Compare with another segment'
      optionFilterProp='children'
      onChange={onChange}
      filterOption={(input, option) =>
        (option?.label ?? '').toLowerCase().includes(input.toLowerCase())
      }
      options={options}
      value={compareSegmentId}
      style={{
        width: '320px'
      }}
      allowClear
    />
  );
}

export default SegmentCompareDropdown;
