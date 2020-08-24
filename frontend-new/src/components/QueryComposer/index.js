import React, { useState } from 'react';
import { Drawer, Button } from 'antd';

import {SVG} from 'factorsComponents';
import styles from './index.module.scss';
import QueryBlock from './QueryBlock';

function QueryComposer({ drawerVisible, queries, onClose, runQuery, addEvent}) {

    if(!drawerVisible) {return null};

    const title = () => {
        return (<div className={styles.composer_title}>
            <div>
                <SVG name="teamfeed"></SVG>
                <span className={styles.composer_title__heading}>Find event funnel for</span>
            </div>
            <span className={styles.composer_title__help}>
                <SVG name="play"></SVG>
                    Help
            </span>
            
        </div>)
    }

    const queryList = () => {
        const blockList = [];

        queries.forEach((event, index) => {
            blockList.push(
                <div className={styles.composer_body__query_block}>
                    <QueryBlock index={index+1} event={event} eventChange={addEvent}></QueryBlock>
                </div>
            )
        });

        if(queries.length < 6) {
            blockList.push(
            <div className={styles.composer_body__query_block}>
                <QueryBlock index={queries.length+1} eventChange={addEvent}></QueryBlock>
            </div>
            )
        }

        return blockList;
    }

    const groupByBlock = () => {
        if(queries.length >= 2) {
            return (
                <div className={styles.composer_body__query_block}>
                    <span>Group By</span>
                </div>
            )
        }
    }

    const footer = () => {
        if(queries.length < 2) {return null}
        else {

            return (
                <div className={styles.composer_footer}>
                    <Button> Last Week </Button>
                    <Button type="primary" onClick={runQuery}>Run Query</Button> 
                </div>
            )
        }
    }

    return(
        <Drawer
        title={title()}
        placement="left"
        closable={true}
        visible={drawerVisible}
        onClose={onClose}
        mask={false}
        getContainer={false}
        width={"600px"}
        className={styles.query_composer}
      >
        <div className={styles.composer_body}>
            {queryList()}
            {groupByBlock()}
        </div>
        {footer()}
          
      </Drawer>
    )
}

export default QueryComposer;