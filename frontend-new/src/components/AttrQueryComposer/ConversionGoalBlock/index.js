import React, {useState} from 'react';
import styles from './index.module.scss';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';

import { Button } from 'antd';
import { SVG, Text } from 'factorsComponents';

const ConversionGoalBlock = ({eventGoal, eventGoalChange}) => {

    const [selectVisible, setSelectVisible] = useState(false);

    const toggleEventSelect = () => {
        setSelectVisible(!selectVisible);
    }

    const addFilter = () => {};

    const deleteItem = () => {};

    const additionalActions = () => {
        return (
                <div className={'fa--query_block--actions'}>
                   <Button size={'large'} type="text" onClick={addFilter} className={'mr-1'}><SVG name="filter"></SVG></Button>
                   <Button size={'large'} type="text" onClick={deleteItem}><SVG name="trash"></SVG></Button>
                </div>
        );
      };

    const renderGoalBlockContent = () => {
        return (
            <div className={`${styles.block__content}`}>
                <Button 
                    size={'large'} 
                    type="link" 
                    onClick={toggleEventSelect}>
                        <SVG name="mouseevent" extraClass={'mr-1'}></SVG>
                         {eventGoal && eventGoal.label} 
                </Button> 

                {additionalActions()}
            </div>
        )
    };

    const renderGoalSelect = () => {
        <div className={`${styles.block__content}`}>
                
        </div>
    };

    return (
        <div className={styles.block}>
            {eventGoal? renderGoalBlockContent() : renderGoalSelect()}
        </div>
    )
}


export default ConversionGoalBlock;