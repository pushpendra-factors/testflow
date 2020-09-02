import React, { useState } from 'react';
import Header from './header';
import ResultsPage from './ResultsPage';
import QueryComposer from '../../components/QueryComposer';
import CoreQueryHome from '../CoreQueryHome';
import { Drawer, Button, Collapse, Select, Popover } from 'antd';
import {SVG, Text} from 'factorsComponents';
import styles from './index.module.scss';

function CoreQuery() {
    const [drawerVisible, setDrawerVisible] = useState(false);
    const [showResult, setShowResult] = useState(false);
    const [queries, setQueries] = useState([]);

    const queryChange = (newEvent, index, changeType = 'add') => {
        const queryupdated = [...queries];
        if(queryupdated[index]) {
            if(changeType === 'add') {
                queryupdated[index] = newEvent;
            } else {
                queryupdated.splice(index, 1);
            }
            
        } else {
            queryupdated.push(newEvent);
        }
        setQueries(queryupdated);
    }

    const runQuery = () => {
        setShowResult(true);
        closeDrawer();
    }

    const closeDrawer = () => {
        setDrawerVisible(false);
    }

    const title = () => {
        return (<div className={styles.composer_title}>
            <div>
                <SVG name="teamfeed"></SVG>
                <Text type={'title'} level={4} weight={`bold`} extraClass={`ml-2`}>Find event funnel for</Text> 
            </div>
            <span className={styles.composer_title__help}>
                <SVG name="play"></SVG>
                    Help
            </span>
            
        </div>)
    }

    return (
        <>
             <Drawer
        title={title()}
        placement="left"
        closable={true}
        visible={drawerVisible}
        onClose={closeDrawer} 
        getContainer={false}
        width={"600px"}
        className={`fa-drawer`}
      >

            <QueryComposer  
                queries={queries} 
                runQuery={runQuery}
                eventChange={queryChange}
            /> 
      </Drawer>

            {
                showResult ? (<ResultsPage setDrawerVisible={setDrawerVisible} queries={queries.map(elem => elem.label)} />) : (<CoreQueryHome setDrawerVisible={setDrawerVisible} />)
            }

        </>
    )
}

export default CoreQuery;