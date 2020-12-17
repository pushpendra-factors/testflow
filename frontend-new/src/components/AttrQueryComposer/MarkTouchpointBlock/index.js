import React, {useState, useEffect} from 'react';
import styles from './index.module.scss';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';

import GroupSelect from '../../QueryComposer/GroupSelect';

import { Button } from 'antd';
import { SVG, Text } from 'factorsComponents';

const MarkTouchpointBlock = ({touchPoint, touchPointOptions, setTouchpoint}) => {

    const [selectVisible, setSelectVisible] = useState(false);

    const toggleTouchPointSelect = () => {
        setSelectVisible(!selectVisible);
    }

    const onEventSelect = (val) => {
        let currTouchpoint = (' ' + touchPoint).slice(1);
        currTouchpoint = val;
        setTouchpoint(currTouchpoint);
        setSelectVisible(false);
    };

    const selectEvents = () => {
    
        return (
            <div className={styles.query_block__event_selector}>
                   {selectVisible
                     ? <GroupSelect
                            groupedProperties={touchPointOptions}
                            placeholder="Select Touchpoint"
                            optionClick={(group, val) => onEventSelect(val[0])}
                            onClickOutside={() => setSelectVisible(false)}

                        ></GroupSelect>

                     : null }
            </div>
        );
    };

    const renderTouchPointSelect = () => {
        return (
            <div className={`${styles.block__content}`}>
                    <div className={'fa--query_block--add-event flex justify-center items-center mr-2'}>
                        <SVG name={'plus'} color={'purple'}></SVG>
                    </div>
                    
                    {!selectVisible && 
                        <Button 
                            size={'large'} 
                            type="link" 
                            onClick={toggleTouchPointSelect}>Add a Touchpoint</Button>}
                    
                    {selectEvents()}
            </div> 
        )
    }

    const addFilterBlock = () => {};

    const deleteItem = () => {};

    const additionalActions = () => {
        return (
                <div className={'fa--query_block--actions'}>
                   <Button size={'large'} type="text" onClick={addFilterBlock} className={'mr-1'}><SVG name="filter"></SVG></Button>
                   <Button size={'large'} type="text" onClick={deleteItem}><SVG name="trash"></SVG></Button>
                </div>
        );
    };

    

    const renderMarkTouchpointBlockContent = () => {
        return (
            <div className={`${styles.block__content}`}>
                {!selectVisible && <Button 
                    size={'large'} 
                    type="link" 
                    onClick={toggleTouchPointSelect}>
                        <SVG name="mouseevent" extraClass={'mr-1'}></SVG>
                         {touchPoint} 
                </Button> }

                {selectEvents()}

                {additionalActions()}
            </div>
        )
    };

    return (
        <div className={styles.block}>
            {touchPoint?.length? renderMarkTouchpointBlockContent() : renderTouchPointSelect()}
        </div>
    )
}

const mapStateToProps = (state) => ({
    activeProject: state.global.active_project,
    touchPointOptions: state.coreQuery.touchpointOptions
});
  
const mapDispatchToProps = dispatch => bindActionCreators({}, dispatch);

export default connect(mapStateToProps, mapDispatchToProps)(MarkTouchpointBlock);