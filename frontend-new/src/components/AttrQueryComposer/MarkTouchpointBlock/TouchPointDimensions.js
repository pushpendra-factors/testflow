import React, { useMemo } from 'react';
import styles from './index.module.scss';
import { Text, SVG } from '../../factorsComponents';
import TouchPointDimensionsList from './TouchPointDimensionsList';
import { useSelector } from 'react-redux';
import FaSelect from '../../FaSelect';

function TouchPointDimensions({
  touchPoint,
  tpDimensionsSelection,
  setTPDimensionsSelection,
}) {
  const { attr_dimensions } = useSelector((state) => state.coreQuery);

  const dimensionsHeading = useMemo(() => {
    const heading = attr_dimensions
      .filter((d) => d.touchPoint === touchPoint && d.enabled)
      .map((d) => d.title)
      .join(', ');
    if (heading.length > 30) {
      return heading.slice(0, 30) + '...';
    } else {
      return heading;
    }
  }, [attr_dimensions, touchPoint]);

  if (!dimensionsHeading) {
    return null;
  }

  return (
    <>
      <Text
        type='title'
        extraClass='text-sm mb-0 ml-4'
        color='grey'
        lineHeight='medium'
      >
        Include
      </Text>
      <div className='ml-4 relative'>
        <Text
          weight='bold'
          type='paragraph'
          extraClass={`mb-0 cursor-pointer flex items-center text-xs ${styles.dimensionsHeading}`}
          onClick={setTPDimensionsSelection.bind(this, true)}
        >
          {dimensionsHeading}
          <div className='ml-1'>
            <SVG name='caretDown' size={16}></SVG>
          </div>
        </Text>
        {tpDimensionsSelection ? (
          <FaSelect
            extraClass={styles.dimensionsSelect}
            onClickOutside={setTPDimensionsSelection.bind(this, false)}
            allowSearch={false}
          >
            <TouchPointDimensionsList touchPoint={touchPoint} />
          </FaSelect>
        ) : null}
      </div>
    </>
  );
}

export default TouchPointDimensions;
