import React, {useState, useEffect} from 'react';
import styles from './index.module.scss';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';

import GroupSelect from '../../QueryComposer/GroupSelect';

import { Button } from 'antd';
import { SVG, Text } from 'factorsComponents';
import FaSelect from '../../FaSelect';

const AttributionOptions = ({models, window, setModelOpt, setWindowOpt}) => {

    const [selectVisibleModel, setSelectVisibleModel] = useState([false, false]);
    const [selectVisibleWindow, setSelectVisibleWindow] = useState(false);
    const [compareModelActive, setCompareModelActive] =useState(false);

    const modelOpts = [
        ['First Touch', 'First_Touch'], 
        ['Last Touch', 'Last_Touch'], 
        ['First Touch Non-Direct', 'First_Touch_ND'],
        ['Last Touch Non-Direct', 'Last_Touch_ND'],
        ['Linear', 'Linear'],
    ];

    useEffect(() => {
        if(models && models[1]) {
            setCompareModelActive(true);
        }
    }, [models])

    const toggleModelSelect = (id) => {
        const selectState = [...selectVisibleModel];
        selectState[id] = !selectState[id];
        setSelectVisibleModel(selectState);
    }

    const setModel = (val, index) => {
        const modelsState = [...models];
        modelsState[index] = val;
        setModelOpt(modelsState);
        toggleModelSelect(index);
    }

    const delModel = (index) => {
        const modelsState = models.filter((m, i) => i !== index);
        setModelOpt(modelsState);
        toggleModelSelect(index);
        index === 1 && setCompareModelActive(false);
    }

    const selectModel = (index) => {
        if(selectVisibleModel[index]) {
            
            return (<FaSelect 
                    options={modelOpts} 
                    delOption={'Remove Comparision'}
                    optionClick={(val) => setModel(val[1], index)}
                    onClickOutside={() => toggleModelSelect(index)}
                    delOptionClick={() => delModel(1)}
                    >

                    </FaSelect>)
        }
    }

    const renderModel = (index) => {
        if(models && models[index]) {
            return (<div className={styles.block__select_wrapper}>
                    <Button 
                        size={'large'} 
                        type="link" 
                        onClick={() => toggleModelSelect(index)}>
                            <SVG name="mouseevent" extraClass={'mr-1'}></SVG>
                            {
                                modelOpts.filter((md) => md[1] === models[index])[0][0]
                            } 
                    </Button>

                    {selectModel(index)}
                </div>)

        } else {
            return (
                <div className={styles.block__select_wrapper}>
                    <div className={styles.block__select_wrapper__block}>
                        {index < 1 &&
                            <div className={'fa--query_block--add-event flex justify-center items-center mr-2'}>
                                <SVG name={'plus'} color={'purple'}></SVG>
                            </div>
                        }
                
                        {!selectVisibleModel[index] && <Button size={'large'} type="link" onClick={() => toggleModelSelect(index)}>Add touchpoint</Button>}

                        {selectModel(index)} 
                    </div>
                </div>
            )
        }
    }

    const addModelAction = () => {
        return (
            <div className={'fa--query_block--actions'}>
                   <Button size={'large'} 
                        type="text" onClick={
                        () => setCompareModelActive(true)
                        } className={'mr-1'}><SVG name="compare"></SVG>
                    </Button>
            </div>
        )
    }

    const renderAttributionModel = () => {
        
            return (
                <div className={styles.block}>
                    <Text 
                        type={'paragraph'} color={`grey`} 
                        extraClass={`${styles.block__content__title_muted}`}> 
                        Attribution Model 
                    </Text>

                    <div className={`${styles.block__content}`}>
                        {renderModel(0)} 

                        {compareModelActive && 
                            <Text type={'paragraph'} 
                                color={`grey`} extraClass={`${styles.block__content__txt_muted}`}> 
                                compared to </Text>
                        }

                        {compareModelActive && renderModel(1)}

                        {!compareModelActive && addModelAction()}
                    </div>
                </div>
            
            )
    };

    const setWindow = (val) => {
        const win = parseInt(val.replace('days', '').trim());
        setWindowOpt(win);
        setSelectVisibleWindow(false);
    }

    const selectWindow = () => {
        if(selectVisibleWindow) {
            const opts = [...new Array(30).keys()].map((opt) => [`${opt+1} days`]);

            return (<FaSelect 
                    options={opts} 
                    optionClick={(val) => setWindow(val[0])}
                    onClickOutside={() => setSelectVisibleWindow(false)}
                    >
                    </FaSelect>)
        }
    }

    const renderWindow = () => {
        if((window !== null && window !== undefined) && window >= 0) {
            return (<div className={styles.block__select_wrapper}>
                    <Button 
                        size={'large'} 
                        type="link" 
                        onClick={() => setSelectVisibleWindow(!selectVisibleWindow)}>
                            <SVG name="clock" className={`mr-1`}></SVG>
                            {window} days 
                    </Button>

                    {selectWindow()}
                </div>)

        } else {
            return (
                <div className={styles.block__select_wrapper}>
                    <div className={styles.block__select_wrapper__block}>
                        <div className={'fa--query_block--add-event flex justify-center items-center mr-2'}>
                            <SVG name={'plus'} color={'purple'}></SVG>
                        </div>
                
                        {!selectVisibleWindow && <Button size={'large'} type="link" onClick={() => setSelectVisibleWindow(!selectVisibleWindow)}>Add Window</Button>}

                        {selectWindow()} 
                    </div>
                </div>
            )
        }
    }

    const renderAttributionWindow = () => {
            return (
                <div className={styles.block}>
                    <Text 
                        type={'paragraph'} color={`grey`} 
                        extraClass={`${styles.block__content__title_muted}`}> 
                        Attribution Window 
                    </Text>

                    <div className={`${styles.block__content}`}>
                        {renderWindow()}
                    </div>
                </div>)
    };

    return (
    <div className={`mt-2`}>
        {renderAttributionModel()}
        {renderAttributionWindow()}
    </div>)

}

const mapStateToProps = (state) => ({
    activeProject: state.global.active_project
});
  
const mapDispatchToProps = dispatch => bindActionCreators({}, dispatch);

export default connect(mapStateToProps, mapDispatchToProps)(AttributionOptions);