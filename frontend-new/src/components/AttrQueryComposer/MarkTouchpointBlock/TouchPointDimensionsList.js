import React, { memo, useMemo, useCallback } from 'react';
import { useSelector, useDispatch } from 'react-redux';
import OptionsPopover from '../../../Views/CoreQuery/AttributionsResult/OptionsPopover';
import { Text } from '../../factorsComponents';
import {
  INITIALIZE_CONTENT_GROUPS,
  INITIALIZE_TOUCHPOINT_DIMENSIONS,
} from '../../../reducers/types';

function TouchPointDimensionsList({ touchPoint }) {
  const { attr_dimensions, content_groups } = useSelector(
    (state) => state.coreQuery
  );
  const dispatch = useDispatch();

  const list_dimensions =
    touchPoint === 'LandingPage'
      ? content_groups.slice()
      : attr_dimensions.slice();

  const handleChange = useCallback(
    ({ header }) => {
      const changedDimensionIndex = list_dimensions.findIndex(
        (elem) => elem.header === header && elem.touchPoint === touchPoint
      );
      const currentEnabilityVal =
        list_dimensions[changedDimensionIndex].enabled;

      const newAttrDimensions = [
        ...list_dimensions.slice(0, changedDimensionIndex),
        {
          ...list_dimensions[changedDimensionIndex],
          enabled: !currentEnabilityVal,
        },
        ...list_dimensions.slice(changedDimensionIndex + 1),
      ];
      const isAtleastOneSelected =
        newAttrDimensions.filter(
          (d) => d.touchPoint === touchPoint && d.enabled
        ).length > 0;
      if (isAtleastOneSelected) {
        if (touchPoint === 'LandingPage') {
          dispatch({
            type: INITIALIZE_CONTENT_GROUPS,
            payload: newAttrDimensions,
          });
        } else
          dispatch({
            type: INITIALIZE_TOUCHPOINT_DIMENSIONS,
            payload: newAttrDimensions,
          });
      }
    },
    [list_dimensions, dispatch, touchPoint]
  );

  const keyOptions = useMemo(() => {
    return list_dimensions.filter(
      (elem) => elem.touchPoint === touchPoint && elem.type === 'key'
    );
  }, [list_dimensions, touchPoint]);

  const customOptions = useMemo(() => {
    return list_dimensions.filter(
      (elem) => elem.touchPoint === touchPoint && elem.type === 'custom'
    );
  }, [list_dimensions, touchPoint]);

  const contentGroups = useMemo(() => {
    return list_dimensions.filter(
      (elem) => elem.touchPoint === touchPoint && elem.type === 'content_group'
    );
  }, [list_dimensions, touchPoint]);

  return (
    <div className='p-1'>
      <div className='p-3'>
        <OptionsPopover options={keyOptions} onChange={handleChange} />
      </div>

      {customOptions.length ? (
        <>
          <hr />
          <Text extraClass='mb-0 mt-2 px-5' type='title' weight='bold'>
            Custom Dimensions
          </Text>
          <div className='px-3 py-1'>
            <OptionsPopover options={customOptions} onChange={handleChange} />
          </div>
        </>
      ) : null}

      {contentGroups.length ? (
        <>
          <hr />
          <Text extraClass='mb-0 mt-2 px-5' type='title' weight='bold'>
            Content Groups
          </Text>
          <div className='px-3 py-1'>
            <OptionsPopover options={contentGroups} onChange={handleChange} />
          </div>
        </>
      ) : null}
    </div>
  );
}

export default memo(TouchPointDimensionsList);
