import React, { useState, useEffect } from 'react';
import styles from './index.module.scss';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';

import GroupSelect2 from '../../QueryComposer/GroupSelect2';
import AttrFilterBlock from '../AttrFilterBlock';

import { setTouchPointFilters } from '../../../reducers/coreQuery/middleware';

import { Button } from 'antd';
import { SVG } from '../../factorsComponents';
import TouchPointDimensions from './TouchPointDimensions';
import FaSelect from 'Components/FaSelect';

const MarkTouchpointBlock = ({
  touchPoint,
  touchPointOptions,
  setTouchpoint,
  campaign_config,
  activeProject,
  setTouchPointFilters,
  filters,
}) => {
  const [tpDimensionsSelection, setTPDimensionsSelection] = useState(false);
  const [selectVisible, setSelectVisible] = useState(false);
  const [filterDD, setFilterDD] = useState(false);
  const [moreOptions, setMoreOptions] = useState(false);
  const [filterProps, setFilterProperties] = useState({
    event: [],
    user: [],
  });

  useEffect(() => {
    if (campaign_config.properties && touchPoint) {
      const props = {};
      campaign_config.properties.forEach((prop, i) => {
        if (
          prop.label === touchPoint.toLowerCase() ||
          (prop.label === 'ad group' && touchPoint === 'AdGroup')
        ) {
          props[prop.label] = prop.values;
        }
      });
      setFilterProperties(props);
    }
  }, [campaign_config, touchPoint]);

  const editFilter = (index, val) => {
    const fltrs = [...filters];
    fltrs[index] = val;
    setTouchPointFilters(fltrs);
    setFilterDD(false);
  };

  const delFilter = (index) => {
    const fltrs = [...filters].filter((f, i) => i !== index);
    setTouchPointFilters(fltrs);
    setFilterDD(false);
  };

  const addFilterBlock = () => {
    setFilterDD(true);
  };

  const deleteItem = () => {
    setTouchpoint('');
    setFilters([]);
  };
  const toggleTouchPointSelect = () => {
    setSelectVisible(!selectVisible);
  };

  const onEventSelect = (val) => {
    if (val === 'Channel') {
      setTouchpoint('ChannelGroup');
    } else {
      setTouchpoint(val);
    }
    setTouchPointFilters([]);
    setSelectVisible(false);
  };

  const selectEvents = () => {
    return (
      <div className={styles.block__event_selector}>
        {selectVisible ? (
          <GroupSelect2
            groupedProperties={touchPointOptions}
            placeholder='Select Touchpoint'
            optionClick={(group, val) => onEventSelect(val[0])}
            onClickOutside={() => setSelectVisible(false)}
            extraClass={touchPoint ? styles.touchPointSelector : ''}
          ></GroupSelect2>
        ) : null}
      </div>
    );
  };

  const renderTouchPointSelect = () => {
    return (
      <div className={'flex justify-start items-center mt-2'}>
        {<Button type="text" onClick={toggleTouchPointSelect} icon={<SVG name={'plus'} color={'grey'} />}>Add a Touchpoint</Button>}
        {selectEvents()}
      </div>
    );
  };

  const insertFilter = (fil) => {
    const fltrs = [...filters];
    fltrs.push(fil);
    setTouchPointFilters(fltrs);
    setFilterDD(false);
  };

  const ifTouchPointFilter = () => {
    if (
      touchPoint === 'AdGroup' &&
      filterProps['ad group'] &&
      filterProps['ad group'].length
    ) {
      return true;
    }

    if (
      filterProps[touchPoint.toLowerCase()] &&
      filterProps[touchPoint.toLowerCase()].length
    ) {
      return true;
    }

    return false;
  };

  const additionalActions = () => {
    return (
      <div className={'fa--query_block--actions-cols flex'}>
        {ifTouchPointFilter() && (
          <div className={`relative`}>
            <Button
              type='text'
              size={'large'}
              onClick={() => setMoreOptions(true)}
              className={'ml-1 mr-1'}
            >
              <SVG name='more'></SVG>
            </Button>

            {moreOptions ? <FaSelect
              options={[[`Filter By`, 'filter']]}
              optionClick={(val) => {addFilterBlock(); setMoreOptions(false)}}
              onClickOutside={() => setMoreOptions(false)}
            ></FaSelect> : false}
          </div>

        )}
        <Button size={'large'} type='text' onClick={deleteItem}>
          <SVG name='trash'></SVG>
        </Button>
      </div>
    );
  };

  const renderFilterBlock = () => {
    if (filterProps) {
      const filtrs = [];

      filters.forEach((filt, id) => {
        filtrs.push(
          <div key={id} className={`mt-4`}>
            <AttrFilterBlock
              activeProject={activeProject}
              index={id}
              blockType={'event'}
              filterType={'channel'}
              filter={filt}
              insertFilter={(val) => editFilter(id, val)}
              closeFilter={() => setFilterDD(false)}
              deleteFilter={delFilter}
              closeFilter={() => setFilterDD(false)}
              typeProps={{ channel: 'all_ads' }}
              filterProps={filterProps}
              propsConstants={Object.keys(filterProps)}
            ></AttrFilterBlock>
          </div>
        );
      });

      if (filterDD) {
        filtrs.push(
          <div key={filtrs.length} className={`mt-4`}>
            <AttrFilterBlock
              activeProject={activeProject}
              blockType={'event'}
              filterType={'channel'}
              delBtnClass={styles.filterDelBtn}
              typeProps={{ channel: 'all_ads' }}
              filterProps={filterProps}
              propsConstants={Object.keys(filterProps)}
              insertFilter={insertFilter}
              deleteFilter={() => setFilterDD(false)}
              closeFilter={() => setFilterDD(false)}
            ></AttrFilterBlock>
          </div>
        );
      }

      return <div className={styles.block}>{filtrs}</div>;
    }
  };

  const renderMarkTouchpointBlockContent = () => {
    return (
      <div
        className={`${styles.block__content} fa--query_block_section--basic relative`}
      >
        {
          <Button type='link' onClick={toggleTouchPointSelect}>
            <SVG name='mouseevent' extraClass={'mr-1'}></SVG>
            {touchPoint === 'ChannelGroup' ? 'Channel' : touchPoint}
          </Button>
        }

        {touchPoint?.length ? (
          <TouchPointDimensions
            touchPoint={touchPoint}
            tpDimensionsSelection={tpDimensionsSelection}
            setTPDimensionsSelection={setTPDimensionsSelection}
          />
        ) : null}

        {selectEvents()}

        {additionalActions()}
      </div>
    );
  };

  return (
    <div className={styles.block}>
      {touchPoint?.length
        ? renderMarkTouchpointBlockContent()
        : renderTouchPointSelect()}
      {touchPoint?.length ? renderFilterBlock() : null}
    </div>
  );
};

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  touchPointOptions: state.coreQuery.touchpointOptions,
  filters: state.coreQuery.touchpoint_filters,
  campaign_config: state.coreQuery.campaign_config,
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      setTouchPointFilters,
    },
    dispatch
  );

export default connect(
  mapStateToProps,
  mapDispatchToProps
)(MarkTouchpointBlock);
