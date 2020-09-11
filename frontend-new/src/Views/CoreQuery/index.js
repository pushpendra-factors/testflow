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
        console.log("clickedd!!")
        setDrawerVisible(false);
    }

    const title = () => {
        return (<div className={`flex justify-between items-center`}>
            <div className={`flex`}>
                <SVG name="teamfeed"></SVG>
                <Text type={'title'} level={4} weight={`bold`} extraClass={`ml-2 m-0`}>Find event funnel for</Text> 
            </div>
            <div className={`flex justify-end items-center`}>
                <Button type="text"><SVG name="play"></SVG>Help</Button>  
                <Button type="text" onClick={()=>closeDrawer()}><SVG name="times"></SVG></Button>  
            </div>
            
        </div>)
    }

    return (
        <>
             <Drawer
        title={title()}
        placement="left" 
        closable={false}
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