import React, {useState, useEffect} from 'react';
import styles from './index.module.scss';
import { SVG } from 'factorsComponents';

const FaSelect = ({ options, delOption, 
    optionClick, delOptionClick,
    onClickOutside }) => {


    const renderOptions = () => {
        const rendOpts = options.map((op, index) => {
            return (
                <div key={index} title={op[0]} className={`fa-select-group-select--options`}
                    onClick={() => optionClick(op)} >
                    {op[1] && <SVG name={op[1]} extraClass={'self-center'}></SVG>}
                    <span className={`ml-1 ${styles.optText}`}>{op[0]}</span>
                </div>
            )
        })

        if(delOption) {
            rendOpts.push(
                <div key={100} className={`fa-select-group-select--options ${styles.dropdown__del_opt}`}
                    onClick={delOptionClick} >
                    <SVG name={'remove'} extraClass={'self-center'}></SVG>
                    <span className={'ml-1'}>{delOption}</span>
                </div>
            )
        }

        return rendOpts;
    }
        

    return (<>
            <div className={`${styles.dropdown__select} fa-select fa-select--group-select`}>
            <div className={styles.dropdown__select__content}>
                {renderOptions()}
            </div>
            </div>
            <div className={styles.dropdown__hd_overlay} onClick={onClickOutside}></div>
    </>);

}

export default FaSelect;