import React, { memo, useContext, useEffect, useState } from 'react';
import cx from 'classnames';
import { Popover } from 'antd';
import { CoreQueryContext } from '../../../contexts/CoreQueryContext';
import { SVG } from '../../factorsComponents';
import styles from './FunnelsConversionDurationBlock.module.scss';
import FunnelsConversionDurationBlockContent from './FunnelsConversionDurationBlockContent';

const UnitsMapper = {
  D: 'Days',
  H: 'Hours',
  M: 'Minutes'
};

export const FunnelsConversionDurationBlockComponent = ({
  funnelConversionDurationNumber,
  funnelConversionDurationUnit,
  onChange
}) => {
  const [number, setNumber] = useState(30);
  const [unit, setUnit] = useState('D');
  const [open, setOpen] = useState(false);

  const handleApply = () => {
    onChange({
      funnelConversionDurationNumber: number,
      funnelConversionDurationUnit: unit
    });
    setOpen(false);
  };

  const handleOpen = () => {
    setOpen(true);
  };

  useEffect(() => {
    setNumber(funnelConversionDurationNumber);
    setUnit(funnelConversionDurationUnit);
  }, [funnelConversionDurationNumber, funnelConversionDurationUnit]);

  return (
    <Popover
      overlayStyle={{ width: '328px' }}
      placement='bottom'
      content={
        <FunnelsConversionDurationBlockContent
          funnelConversionDurationNumber={number}
          funnelConversionDurationUnit={unit}
          onNumberChange={setNumber}
          onUnitChange={setUnit}
          onApply={handleApply}
        />
      }
      trigger='click'
      visible={open}
    >
      <div
        className={cx(
          'flex items-center font-semibold cursor-pointer',
          styles['dropdown-placeholder']
        )}
        onClick={handleOpen}
      >
        <div>
          {funnelConversionDurationNumber}{' '}
          {UnitsMapper[funnelConversionDurationUnit]}
        </div>
        <SVG name='caretDown' color={'#1890ff'} />
      </div>
    </Popover>
  );
};

const FunnelsConversionDurationBlockMemoized = memo(
  FunnelsConversionDurationBlockComponent
);

const FunnelsConversionDurationBlock = (props) => {
  const coreQueryContext = useContext(CoreQueryContext);

  return (
    <FunnelsConversionDurationBlockMemoized
      funnelConversionDurationUnit={
        coreQueryContext.coreQueryState.funnelConversionDurationUnit
      }
      funnelConversionDurationNumber={
        coreQueryContext.coreQueryState.funnelConversionDurationNumber
      }
      onChange={coreQueryContext.updateCoreQueryReducer}
      {...props}
    />
  );
};

export default FunnelsConversionDurationBlock;
