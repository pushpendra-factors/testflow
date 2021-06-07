import React, { memo, useMemo, useCallback } from 'react';
import { useSelector, useDispatch } from 'react-redux';
import OptionsPopover from '../../../Views/CoreQuery/AttributionsResult/OptionsPopover';
import { Text } from '../../factorsComponents';
import { INITIALIZE_TOUCHPOINT_DIMENSIONS } from '../../../reducers/types';

function TouchPointDimensionsList({ touchPoint }) {
  const { attr_dimensions } = useSelector((state) => state.coreQuery);
  const dispatch = useDispatch();

  const handleChange = useCallback(
    ({ header }) => {
      const changedDimensionIndex = attr_dimensions.findIndex(
        (elem) => elem.header === header && elem.touchPoint === touchPoint
      );
      const isDisabled = attr_dimensions[changedDimensionIndex].disabled;
      if (isDisabled) {
        return false;
      }
      const currentEnabilityVal =
        attr_dimensions[changedDimensionIndex].enabled;
      const newAttrDimensions = [
        ...attr_dimensions.slice(0, changedDimensionIndex),
        {
          ...attr_dimensions[changedDimensionIndex],
          enabled: !currentEnabilityVal,
        },
        ...attr_dimensions.slice(changedDimensionIndex + 1),
      ];
      dispatch({
        type: INITIALIZE_TOUCHPOINT_DIMENSIONS,
        payload: newAttrDimensions,
      });
    },
    [attr_dimensions, dispatch, touchPoint]
  );

  const keyOptions = useMemo(() => {
    return attr_dimensions.filter(
      (elem) => elem.touchPoint === touchPoint && elem.type === 'key'
    );
  }, [attr_dimensions, touchPoint]);

  const customOptions = useMemo(() => {
    return attr_dimensions.filter(
      (elem) => elem.touchPoint === touchPoint && elem.type === 'custom'
    );
  }, [attr_dimensions, touchPoint]);

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
    </div>
  );
}

export default memo(TouchPointDimensionsList);
