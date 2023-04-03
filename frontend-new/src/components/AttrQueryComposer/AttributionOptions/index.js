import React, { useState, useEffect } from 'react';
import styles from './index.module.scss';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import { Button, Tooltip, Radio, Popover } from 'antd';
import { SVG, Text } from '../../factorsComponents';
import FaSelect from '../../FaSelect';

import { setattrQueryType } from '../../../reducers/coreQuery/middleware';

const AttributionOptions = ({
  models,
  window,
  setModelOpt,
  setWindowOpt,
  timeline,
  setattrQueryType
}) => {
  const [selectVisibleModel, setSelectVisibleModel] = useState([false, false]);
  const [selectVisibleWindow, setSelectVisibleWindow] = useState(false);
  const [compareModelActive, setCompareModelActive] = useState(false);

  const [moreOptions, setMoreOptions] = useState(false);

  const [timelineSelect, setTimelineSelect] = useState(false);

  const modelOpts = [
    ['First Touch', 'First_Touch'],
    ['Last Touch', 'Last_Touch'],
    ['First Touch Non-Direct', 'First_Touch_ND'],
    ['Last Touch Non-Direct', 'Last_Touch_ND'],
    ['Linear', 'Linear'],
    ['U Shaped', 'U_Shaped'],
    ['W Shaped', 'W_Shaped'],
    ['Time Decay', 'Time_Decay'],
    ['Influence', 'Influence'],
    ['Last Campaign Touch', 'Last_Campaign_Touch']
  ];

  const timeLineMap = {
    EngagementBased: 'Interaction Time',
    ConversionBased: 'Conversion Time'
  };

  useEffect(() => {
    if (models && models[1]) {
      setCompareModelActive(true);
    }
  }, [models]);

  const toggleModelSelect = (id) => {
    const selectState = [...selectVisibleModel];
    selectState[id] = !selectState[id];
    setSelectVisibleModel(selectState);
  };

  const setModel = (val, index) => {
    const modelsState = [...models];
    modelsState[index] = val;
    setModelOpt(modelsState);
    toggleModelSelect(index);
  };

  const delModel = (index) => {
    const modelsState = models.filter((m, i) => i !== index);
    setModelOpt(modelsState);
    toggleModelSelect(index);
    index === 1 && setCompareModelActive(false);
  };

  const selectModel = (index) => {
    if (!selectVisibleModel[index]) return null;

    if (index === 0) {
      return (
        <FaSelect
          options={modelOpts}
          optionClick={(val) => setModel(val[1], index)}
          onClickOutside={() => toggleModelSelect(index)}
        ></FaSelect>
      );
    }

    if (index === 1) {
      return (
        <FaSelect
          options={modelOpts}
          delOption={'Remove Comparision'}
          optionClick={(val) => setModel(val[1], index)}
          onClickOutside={() => toggleModelSelect(index)}
          delOptionClick={() => delModel(index)}
        ></FaSelect>
      );
    }
  };

  const renderModel = (index) => {
    if (models && models[index]) {
      return (
        <div
          className={`${styles.block__select_wrapper} fa--query_block_section--basic`}
        >
          <Button type='link' onClick={() => toggleModelSelect(index)}>
            <SVG name='mouseevent' extraClass={'mr-1'}></SVG>
            {modelOpts.filter((md) => md[1] === models[index])[0][0]}
          </Button>
          {selectModel(index)}
        </div>
      );
    } else {
      return (
        <div
          className={`${styles.block__select_wrapper} fa--query_block_section--basic`}
        >
          <div className={styles.block__select_wrapper__block}>
            {
              <Button
                size={'normal'}
                type='text'
                onClick={() => toggleModelSelect(index)}
                icon={index < 1 && <SVG name={'plus'} color={'grey'} />}
              >
                Add Model
              </Button>
            }

            {selectModel(index)}
          </div>
        </div>
      );
    }
  };

  const modelActions = (selectFlag) => {
    if (selectFlag) {
    }
  };

  const addModelAction = () => {
    return (
      <div className={'fa--query_block--actions--cols relative ml-2'}>
        <Tooltip title={'Add Compare Model'}>
          <Button
            type='text'
            onClick={() => setCompareModelActive(true) && setMoreOptions(false)}
            className={'fa-btn--custom mr-1'}
          >
            <SVG name='compare'></SVG>
          </Button>
        </Tooltip>
      </div>
    );
  };

  const renderAttributionModel = () => {
    return (
      <div
        className={`${styles.block__content} mt-3 flex items-center relative fa--query_block_section--basic`}
      >
        {renderModel(0)}

        {compareModelActive && (
          <div className={`${styles.block__select_wrapper} mx-2`}>
            <Text
              type={'paragraph'}
              color={`grey`}
              extraClass={`${styles.block__content__txt_muted}`}
            >
              vs
            </Text>
          </div>
        )}

        {compareModelActive && renderModel(1)}

        {!compareModelActive && models[0] && (
          <div className={styles.block__additional_actions}>
            {addModelAction()}
          </div>
        )}
      </div>
    );
  };

  const setWindow = (val) => {
    const win = parseInt(val.replace('days', '').trim());
    setWindowOpt(win);
    setSelectVisibleWindow(false);
  };

  const selectWindow = () => {
    if (selectVisibleWindow) {
      const opts = [1, 3, 7, 14, 30, 60, 90, 180, 365].map((opt) => [
        `${opt} ${opt === 1 ? 'day' : 'days'}`
      ]);

      return (
        <FaSelect
          options={opts}
          optionClick={(val) => setWindow(val[0])}
          onClickOutside={() => setSelectVisibleWindow(false)}
        ></FaSelect>
      );
    }
  };

  const renderWindow = () => {
    if (window !== null && window !== undefined && window >= 0) {
      return (
        <div className={`relative mr-2`}>
          <Button
            size={'small'}
            type='link'
            onClick={() => setSelectVisibleWindow(!selectVisibleWindow)}
          >
            {window} {window === 1 ? 'day' : 'days'}
          </Button>
          {selectWindow()}
        </div>
      );
    } else {
      return (
        <div className={`relative mr-2`}>
          {
            <Popover
              title={
                <>
                  <b>Set your attribution window</b>
                </>
              }
              content={
                <>
                  This determines the earliest time prior to the occurrence of a
                  conversion event, within which the marketing touchpoint should
                  have happened for it to be considered in the attribution
                  analysis
                  <br />
                  <br />
                  <a href='#'>Learn More </a>
                </>
              }
              overlayStyle={{ width: '260px' }}
            >
              <Button
                size={'small'}
                type='link'
                onClick={() => setSelectVisibleWindow(!selectVisibleWindow)}
              >
                Add Window
              </Button>
            </Popover>
          }

          {selectWindow()}
        </div>
      );
    }
  };

  const renderAttributionWindow = () => {
    return (
      <>
        <Text
          type={'paragraph'}
          color={`grey`}
          extraClass={`${styles.block__content__txt_muted} mr-2`}
        >
          Within a window of
        </Text>

        {renderWindow()}
      </>
    );
  };

  const selectTimeline = () => {
    if (timelineSelect) {
      return (
        <FaSelect
          options={[
            ['Interaction Time', 'EngagementBased'],
            ['Conversion Time', 'ConversionBased']
          ]}
          optionClick={(val) =>
            setattrQueryType(val[1]) && setTimelineSelect(false)
          }
          onClickOutside={() => setTimelineSelect(false)}
        ></FaSelect>
      );
    }
  };

  const renderTimeLine = () => {
    if (timeline !== null && timeline !== undefined) {
      return (
        <div className={`relative`}>
          <Popover
            title={
              <>
                <b>Set your Interaction time</b>
              </>
            }
            content={
              <>
                <b>Interaction time</b> attributes conversions that happen
                during the month (01 to 30 June) to campaigns within that 30 day
                attribution window.
                <br />
                <br />
                <b>Conversion time</b> considers all the marketing touchpoints
                which happened between this month, till the end of your sales
                cycle, and attributes conversions.
                <br />
                <br />
                <a href='#'>Learn more</a> about the differences between the
                two.
              </>
            }
            overlayStyle={{ width: '260px' }}
          >
            <Button
              size={'small'}
              type='link'
              onClick={() => setTimelineSelect(!timelineSelect)}
            >
              {timeLineMap[timeline]}
            </Button>
          </Popover>

          {selectTimeline()}
        </div>
      );
    } else {
      return (
        <div className={styles.block__select_wrapper}>
          <div className={styles.block__select_wrapper__block}>
            {!timelineSelect && (
              <Button
                className={`relative`}
                size={'small'}
                type='link'
                onClick={() => setTimelineSelect(!timelineSelect)}
              >
                Add Timeline
              </Button>
            )}

            {selectTimeline()}
          </div>
        </div>
      );
    }
  };

  const renderAttributionTimeline = () => {
    return (
      <>
        <Text
          type={'paragraph'}
          color={`grey`}
          extraClass={`${styles.block__content__title_muted} mr-2`}
        >
          as timeline of
        </Text>

        {renderTimeLine()}
      </>
    );
  };

  return (
    <div className={`${styles.block}`}>
      {renderAttributionModel()}
      <div className={`flex items-center m-0 mt-2`}>
        <SVG name='clock' size={20} extraClass={`mx-2`}></SVG>
        {renderAttributionWindow()}
        {renderAttributionTimeline()}
      </div>
    </div>
  );
};

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  timeline: state.coreQuery.attr_query_type
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      setattrQueryType
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(AttributionOptions);
