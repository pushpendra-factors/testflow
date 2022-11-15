import React, { useState, useEffect } from 'react';
import styles from './index.module.scss';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';

import AttrFilterBlock from '../AttrFilterBlock';

import {
  setTouchPointFilters,
  setTacticOfferType
} from 'Attribution/state/actions';

import { Button, Popover, Radio, Row } from 'antd';
import { SVG, Text } from 'Components/factorsComponents';
import TouchPointDimensions from './TouchPointDimensions';
import FaSelect from 'Components/FaSelect';

import ORButton from 'Components/ORButton';
import { compareFilters, groupFilters } from 'Utils/global';
import { InfoCircleOutlined } from '@ant-design/icons';

function MarkTouchpointBlock({
  touchPoint,
  touchPointOptions,
  setTouchpoint,
  campaign_config,
  activeProject,
  setTouchPointFilters,
  filters,
  setTacticOfferType,
  touchPointRef
}) {
  const [tpDimensionsSelection, setTPDimensionsSelection] = useState(false);
  const [selectVisible, setSelectVisible] = useState(false);
  const [filterDD, setFilterDD] = useState(false);
  const [moreOptions, setMoreOptions] = useState(false);
  const [filterProps, setFilterProperties] = useState({
    event: [],
    user: []
  });
  const [orFilterIndex, setOrFilterIndex] = useState(-1);

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
    const filtersSorted = [...filters];
    filtersSorted.sort(compareFilters);
    filtersSorted[index] = val;
    setTouchPointFilters(filtersSorted);
    setFilterDD(false);
  };

  const delFilter = (index) => {
    const filtersSorted = [...filters];
    filtersSorted.sort(compareFilters);
    const fltrs = filtersSorted.filter((f, i) => i !== index);
    setTouchPointFilters(fltrs);
    setFilterDD(false);
  };

  const closeFilter = () => {
    setFilterDD(false);
    setOrFilterIndex(-1);
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

  const selectTouchPointOpts = () => {
    let tchPointOpts = [];
    if (!touchPointRef || touchPointRef.includes('Tactic')) {
      tchPointOpts = touchPointOptions[0].values
        .filter((opt) => !['LandingPage'].includes(opt[0]))
        .map((option) => new Array(option[0], option[0]));
    } else {
      tchPointOpts = touchPointOptions[0].values
        .filter((opt) => !['AdGroup', 'Keyword'].includes(opt[0]))
        .map((option) => new Array(option[0], option[0]));
    }

    return tchPointOpts;
  };

  const selectEvents = () => (
    <div className={styles.block__event_selector}>
      {selectVisible ? (
        <FaSelect
          options={selectTouchPointOpts()}
          optionClick={(val) => onEventSelect(val[0])}
          onClickOutside={() => setSelectVisible(false)}
          extraClass={touchPoint ? styles.touchPointSelector : ''}
          showIcon={false}
        />
      ) : null}
    </div>
  );

  const renderTouchPointSelect = () => (
    <div
      className={`${styles.block_touchpoint_select} flex justify-start items-center mt-3`}
    >
      {
        <Button
          type='text'
          onClick={toggleTouchPointSelect}
          icon={<SVG name={'plus'} color={'grey'} />}
        >
          Add a Property
        </Button>
      }
      {selectEvents()}
    </div>
  );

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

  const additionalActions = () => (
    <div className={'fa--query_block--actions-cols flex'}>
      {ifTouchPointFilter() && (
        <div className={`relative`}>
          <Button
            type='text'
            size={'large'}
            onClick={() => setMoreOptions(true)}
            className={'fa-btn--custom mr-1 btn-total-round'}
          >
            <SVG name='more'></SVG>
          </Button>

          {moreOptions ? (
            <FaSelect
              options={[[`Filter By`, 'filter']]}
              optionClick={(val) => {
                addFilterBlock();
                setMoreOptions(false);
              }}
              onClickOutside={() => setMoreOptions(false)}
            ></FaSelect>
          ) : (
            false
          )}
        </div>
      )}
      <Button
        className={'fa-btn--custom btn-total-round'}
        size={'large'}
        type='text'
        onClick={deleteItem}
      >
        <SVG name='trash'></SVG>
      </Button>
    </div>
  );

  const renderFilterBlock = () => {
    if (filterProps) {
      const filtrs = [];
      let index = 0;
      let lastRef = 0;

      if (filters?.length) {
        const group = groupFilters(filters, 'ref');
        const filtersGroupedByRef = Object.values(group);
        const refValues = Object.keys(group);
        lastRef = parseInt(refValues[refValues.length - 1]);

        filtersGroupedByRef.forEach((filtersGr) => {
          const refValue = filtersGr[0].ref;
          if (filtersGr.length == 1) {
            const filt = filtersGr[0];
            filtrs.push(
              <div className={'fa--query_block--filters flex flex-row'}>
                <div key={index} className={`mt-2`}>
                  <AttrFilterBlock
                    activeProject={activeProject}
                    index={index}
                    blockType={'event'}
                    filterType={'channel'}
                    filter={filt}
                    deleteFilter={delFilter}
                    insertFilter={(val, index) => editFilter(index, val)}
                    closeFilter={closeFilter}
                    typeProps={{ channel: 'all_ads' }}
                    filterProps={filterProps}
                    propsConstants={Object.keys(filterProps)}
                    refValue={refValue}
                  ></AttrFilterBlock>
                </div>
                {index !== orFilterIndex && (
                  <div className={`mt-2`}>
                    <ORButton
                      index={index}
                      setOrFilterIndex={setOrFilterIndex}
                    />
                  </div>
                )}
                {index === orFilterIndex && (
                  <div key={'init'} className={`mt-2`}>
                    <AttrFilterBlock
                      activeProject={activeProject}
                      index={index}
                      blockType={'event'}
                      filterType={'channel'}
                      delBtnClass={styles.filterDelBtn}
                      deleteFilter={closeFilter}
                      insertFilter={insertFilter}
                      closeFilter={closeFilter}
                      typeProps={{ channel: 'all_ads' }}
                      filterProps={filterProps}
                      propsConstants={Object.keys(filterProps)}
                      refValue={refValue}
                      showOr={true}
                    ></AttrFilterBlock>
                  </div>
                )}
              </div>
            );
            index += 1;
          } else {
            filtrs.push(
              <div className={'fa--query_block--filters flex flex-row'}>
                <div key={index} className={`mt-2`}>
                  <AttrFilterBlock
                    activeProject={activeProject}
                    index={index}
                    blockType={'event'}
                    filterType={'channel'}
                    filter={filtersGr[0]}
                    deleteFilter={delFilter}
                    insertFilter={(val, index) => editFilter(index, val)}
                    closeFilter={closeFilter}
                    typeProps={{ channel: 'all_ads' }}
                    filterProps={filterProps}
                    propsConstants={Object.keys(filterProps)}
                    refValue={refValue}
                  ></AttrFilterBlock>
                </div>
                <div key={index + 1} className={`mt-2`}>
                  <AttrFilterBlock
                    activeProject={activeProject}
                    index={index + 1}
                    blockType={'event'}
                    filterType={'channel'}
                    filter={filtersGr[1]}
                    deleteFilter={delFilter}
                    insertFilter={(val, index) => editFilter(index, val)}
                    closeFilter={closeFilter}
                    typeProps={{ channel: 'all_ads' }}
                    filterProps={filterProps}
                    propsConstants={Object.keys(filterProps)}
                    refValue={refValue}
                    showOr={true}
                  ></AttrFilterBlock>
                </div>
              </div>
            );
            index += 2;
          }
        });
      }
      if (filterDD) {
        filtrs.push(
          <div key={filtrs.length} className={`mt-2`}>
            <AttrFilterBlock
              activeProject={activeProject}
              blockType={'event'}
              filterType={'channel'}
              delBtnClass={styles.filterDelBtn}
              typeProps={{ channel: 'all_ads' }}
              filterProps={filterProps}
              propsConstants={Object.keys(filterProps)}
              insertFilter={insertFilter}
              deleteFilter={() => closeFilter()}
              closeFilter={closeFilter}
              refValue={lastRef + 1}
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
        className={`${styles.block__content} fa--query_block_section--basic relative mt-3`}
      >
        {
          <Button
            type='link'
            onClick={toggleTouchPointSelect}
            className={'btn-total-round'}
          >
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

        <div className={styles.block__additional_actions}>
          {additionalActions()}
        </div>
      </div>
    );
  };

  const setTouchpointRef = (val) => {
    setTacticOfferType(val.target.value);
    setTouchpoint('');
  };

  return (
    <>
      <div className={styles.block}>
        <Row className={`mt-2`}>
          <Text
            type={'title'}
            level={7}
            weight={'bold'}
            color={'grey'}
            extraClass={'m-0 ml-2'}
          >
            Type
          </Text>
          <Popover
            className='p-1'
            placement='right'
            content={
              <>
                <b>Tactics</b> are methods in which you reach out to customers.{' '}
                <br />
                For e.g. Google Ads is a classic tactic. <br />
                <br />
                <b>Offers</b> are content that you serve to the visitor. <br />
                It's the offer itself that you present. Landing pages are
                offers.
              </>
            }
            trigger='hover'
            overlayStyle={{ width: '260px' }}
          >
            <InfoCircleOutlined />
          </Popover>
        </Row>
        <Row className={`mt-2 ml-2`}>
          <Radio.Group
            onChange={setTouchpointRef}
            value={touchPointRef ? touchPointRef : 'Tactic'}
          >
            <Radio value={`Tactic`}>Tactics</Radio>
            <Radio value={`Offer`}>Offers</Radio>
            <Radio value={`TacticOffer`}>Tactics and Offers</Radio>
          </Radio.Group>
        </Row>
      </div>
      <div className={`${styles.block} mt-4`}>
        <Row className={`mt-2`}>
          <Text
            type={'title'}
            level={7}
            weight={'bold'}
            color={'grey'}
            extraClass={'m-0 ml-2'}
          >
            Property
          </Text>
          <Popover
            className='p-1'
            placement='right'
            title={
              <>
                <b>There are a hierarchy of levels inside paid marketing.</b>
              </>
            }
            content={
              <>
                Select at what level you would like to run the analysis, between
                Source, Campaign, Adgroup, Creative, or Keyword levels.
              </>
            }
            trigger='hover'
            overlayStyle={{ width: '260px' }}
          >
            <InfoCircleOutlined />
          </Popover>
        </Row>

        <Row className={`ml-2`}>
          {touchPoint?.length
            ? renderMarkTouchpointBlockContent()
            : renderTouchPointSelect()}
        </Row>
        <Row className={`ml-2`}>
          {touchPoint?.length ? renderFilterBlock() : null}
        </Row>
      </div>
    </>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  touchPointOptions: state.coreQuery.touchpointOptions,
  touchPointRef: state.attributionDashboard.tacticOfferType,
  filters: state.attributionDashboard.touchpoint_filters,
  campaign_config: state.coreQuery.campaign_config
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      setTouchPointFilters,
      setTacticOfferType
    },
    dispatch
  );

export default connect(
  mapStateToProps,
  mapDispatchToProps
)(MarkTouchpointBlock);
