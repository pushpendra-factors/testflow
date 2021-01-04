import React, {useState, useEffect} from 'react';
import FaSelect from '../../FaSelect';
import styles from './index.module.scss';

import { SVG, Text } from 'factorsComponents';
import { Button } from 'antd';

const MeasuresBlock = ({measures, onMeasureSelect, measures_metrics}) => {

    const [selectVisible, setSelectVisible] = useState([]);

    const deleteItem = (index) => {
        const measureStates = [...measures.filter((m,i) => i!==index)];
        onMeasureSelect(measureStates);
    }

    const additionalActions = (index) => {
        return (
                <div className={'fa--query_block--actions'}>
                   <Button size={'large'} type="text" onClick={() => deleteItem(index)}><SVG name="trash"></SVG></Button>
                </div>
        );
    };

    const toggleSelect = (index) => {
        const selectStates = [...selectVisible];
        selectStates[index] = !selectStates[index];
        setSelectVisible(selectStates);
    }

    const setMeasures = (val, index) => {
        const measureStates = [...measures];
        measureStates[index] ? 
            measureStates[index] = val :
            measureStates.push(val);
        onMeasureSelect(measureStates);
        toggleSelect(index);
    }

    const selectEvents = (index) => {
    
        return (
            <div className={styles.query_block__event_selector}>
                   {selectVisible[index]
                     ? <FaSelect
                            options={measures_metrics.map(m => [m.replaceAll('_', ' '), m])}
                            onClickOutside={() => toggleSelect(index)}
                            optionClick={(opts) => setMeasures(opts[1], index)}

                        ></FaSelect>

                     : null }
            </div>
        );
    };

    const renderMeasureBlockContent = (measure, index) => {
        return (
            <div className={`${styles.block__content}`}>
                {!selectVisible[index] && <Button 
                    size={'large'} 
                    type="link" 
                    className={styles.optText}
                    onClick={() => toggleSelect(index)}>
                         {measure && measure.replaceAll('_',' ')} 
                </Button> }

                {selectEvents(index)}
                {additionalActions(index)}
            </div>
        )
    };

    const renderMeasureSelect = (index) => {
        return (
            <div className={`${styles.block__content}`}>
                    <div className={'fa--query_block--add-event flex justify-center items-center mr-2'}>
                        <SVG name={'plus'} color={'purple'}></SVG>
                    </div>
                    
                    {!selectVisible[index] && <Button size={'large'} type="link" onClick={() => toggleSelect(index)}>Add new</Button>}
                    
                    {selectEvents(index)}
            </div> 
        )
    };

    const renderMeasures = () => {
        let msrs = [];
        if(measures && measures.length) {
            msrs = measures.map((measure, id) => {
                return renderMeasureBlockContent(measure, id);
            });
        }
        msrs.push(renderMeasureSelect(msrs.length))
        return msrs;
    }

    return (
        <div className={styles.block}>
            {renderMeasures()}
        </div>
    )
};

export default MeasuresBlock;