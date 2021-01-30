import React, {useState} from 'react';
import styles from './index.module.scss';

import { SVG, Text } from "../../factorsComponents";
import {Collapse} from 'antd';

const {Panel} = Collapse;

const ComposerBlock = ({blockTitle, disabled = false, isOpen, showIcon=true, onClick, children}) => {

    const renderHeader = () => {
        return (
            <div className={`${styles.cmpBlock__title}`}>
                <div>
                  <Text
                    type={"title"}
                    level={6}
                    weight={"bold"}
                    disabled={disabled}
                    extraClass={"m-0 mb-2 inline"}
                    
                  >
                    {blockTitle}
                  </Text>
                </div>
                {showIcon && <div className={`${styles.cmpBlock__title__icon}`}>
                        <SVG name={isOpen ? 'minus' : 'plus'} color={disabled ? "black" : "gray"} onClick={e => {
                            onClick();
                        }}/>
                </div>}
            </div>
        )
    }

    return (
        <div className={`${styles.cmpBlock} fa--query_block bordered`}>
            <Collapse
                bordered={false}
                activeKey={isOpen ? [1] : [0]}
                expandIcon={() => {}}
                onChange={() => !disabled && onClick()}
            >
                <Panel header={renderHeader()} key={1}>
                    {children}
                </Panel>
            </Collapse>
        </div>
    )

}

export default ComposerBlock;