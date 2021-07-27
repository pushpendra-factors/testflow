import React, {useState, useEffect} from 'react';
import FaSelect from '../../FaSelect';
import styles from './index.module.scss';

import { SVG, Text } from 'factorsComponents';
import { Button } from 'antd';

const MeasuresBlock = ({measures, onMeasureSelect, measures_metrics}) => {

    const [selectVisible, setSelectVisible] = useState([false]);

    const deleteItem = (index) => {
        const measureStates = [...measures.filter((m,i) => i!==index)];
        onMeasureSelect(measureStates);
    }

    const additionalActions = (index) => {
        return (
                <div className={'fa--query_block--actions'}>
                   <Button type="text" onClick={() => deleteItem(index)} icon={<SVG name="trash"/>}/>
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
            <div key={index} className={`fa--query_block_section--basic flex items-center relative mt-3`}>
                {<Button  
                    type="link"  
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
            <div key={index} className={`flex items-center relative mt-4`}> 
                    
                    {<Button type="text" onClick={() => toggleSelect(index)} icon={<SVG name={'plus'} />}>Add new</Button>}
                    
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