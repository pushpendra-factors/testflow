import React, { useState, useEffect } from 'react';
import styles from './index.module.scss';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import { Button, Radio } from 'antd';
import { SVG, Text } from '../../factorsComponents';
import FaSelect from '../../FaSelect';

import { setattrQueryType } from '../../../reducers/coreQuery/middleware';

const AttributionOptions = ({
  models,
  window,
  setModelOpt,
  setWindowOpt,
  timeline,
  setattrQueryType,
}) => {
  const [selectVisibleModel, setSelectVisibleModel] = useState([false, false]);
  const [selectVisibleWindow, setSelectVisibleWindow] = useState(false);
  const [compareModelActive, setCompareModelActive] = useState(false);

  const modelOpts = [
    ['First Touch', 'First_Touch'],
    ['Last Touch', 'Last_Touch'],
    ['First Touch Non-Direct', 'First_Touch_ND'],
    ['Last Touch Non-Direct', 'Last_Touch_ND'],
    ['Linear', 'Linear'],
  ];

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
            {index < 1 && (
              <div
                className={
                  'fa--query_block--add-event flex justify-center items-center mr-2'
                }
              >
                <SVG name={'plus'} color={'purple'}></SVG>
              </div>
            )}

            {!selectVisibleModel[index] && (
              <Button
                size={'large'}
                type='link'
                onClick={() => toggleModelSelect(index)}
              >
                Add Model
              </Button>
            )}

            {selectModel(index)}
          </div>
        </div>
      );
    }
  };

  const addModelAction = () => {
    return (
      <div className={'fa--query_block--actions mt-2'}>
        <Button
          type='text'
          onClick={() => setCompareModelActive(true)}
          className={'mr-1'}
        >
          <SVG name='compare'></SVG>
        </Button>
      </div>
    );
  };

  const renderAttributionModel = () => {
    return (
      <div className={`${styles.block}`}>
        <Text
          type={'paragraph'}
          color={`grey`}
          extraClass={`${styles.block__content__title_muted}`}
        >
          Attribution Model
        </Text>

        <div
          className={`${styles.block__content} flex items-center relative fa--query_block_section--basic`}
        >
          {renderModel(0)}

          {compareModelActive && (
            <div className={`${styles.block__select_wrapper} mr-1`}>
              <Text
                type={'paragraph'}
                color={`grey`}
                extraClass={`${styles.block__content__txt_muted}`}
              >
                compared to{' '}
              </Text>
            </div>
          )}

          {compareModelActive && renderModel(1)}

          {!compareModelActive && models[0] && addModelAction()}
        </div>
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
      const opts = [1, 3, 7, 14, 30, 60, 90].map((opt) => [
        `${opt} ${opt === 1 ? 'day' : 'days'}`,
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
        <div className={styles.block__select_wrapper}>
          <Button
            size={'large'}
            type='link'
            onClick={() => setSelectVisibleWindow(!selectVisibleWindow)}
          >
            <SVG name='clock' className={`mr-1`}></SVG>
            {window} {window === 1 ? 'day' : 'days'}
          </Button>

          {selectWindow()}
        </div>
      );
    } else {
      return (
        <div className={styles.block__select_wrapper}>
          <div className={styles.block__select_wrapper__block}>
            <div
              className={
                'fa--query_block--add-event flex justify-center items-center mr-2'
              }
            >
              <SVG name={'plus'} color={'purple'}></SVG>
            </div>

            {!selectVisibleWindow && (
              <Button
                size={'large'}
                type='link'
                onClick={() => setSelectVisibleWindow(!selectVisibleWindow)}
              >
                Add Window
              </Button>
            )}

            {selectWindow()}
          </div>
        </div>
      );
    }
  };

  const renderAttributionWindow = () => {
    return (
      <div className={styles.block}>
        <Text
          type={'paragraph'}
          color={`grey`}
          extraClass={`${styles.block__content__title_muted}`}
        >
          Attribution Window
        </Text>

        <div className={`${styles.block__content}`}>{renderWindow()}</div>
      </div>
    );
  };

  const renderTimeLine = () => {
    return (
      <Radio.Group
        onChange={(e) => setattrQueryType(e.target.value)}
        value={timeline}
      >
        <Radio value={'EngagementBased'}>Interaction Time</Radio>
        <Radio value={'ConversionBased'}>Conversion Time</Radio>
      </Radio.Group>
    );
  };

  const renderAttributionTimeline = () => {
    return (
      <div className={styles.block}>
        <Text
          type={'paragraph'}
          color={`grey`}
          extraClass={`${styles.block__content__title_muted}`}
        >
          Attribution Timeline
        </Text>

        <div className={`${styles.block__content} mt-4`}>
          {renderTimeLine()}
        </div>
      </div>
    );
  };

  return (
    <div className={`mt-2`}>
      {renderAttributionModel()}
      {renderAttributionWindow()}
      {renderAttributionTimeline()}
    </div>
  );
};

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  timeline: state.coreQuery.attr_query_type,
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      setattrQueryType,
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(AttributionOptions);
