import React, { useState, useEffect } from 'react';
import styles from './index.module.scss';
import { Input, Button } from 'antd';
import { SVG, Text } from 'factorsComponents';
import { CaretDownOutlined, CaretUpOutlined } from '@ant-design/icons';

function GroupSelect2({
    groupedProperties, placeholder,
    optionClick, onClickOutside, extraClass,
    allowEmpty = false
}) {
    const [groupCollapseState, setGroupCollapseState] = useState({});
    const [searchTerm, setSearchTerm] = useState('');


    useEffect(() => {
        const groupColState = Object.assign({}, groupCollapseState);
        Object.keys(groupedProperties).forEach((index) => { groupColState[index] = true })
        setGroupCollapseState(groupColState);
    }, [groupedProperties]);

    const onInputSearch = (userInput) => {
        setSearchTerm(userInput.currentTarget.value);
    };

    const searchTermExists = (opts) => {
        let termExists = false;

        opts.forEach((grp) => {
            grp.values.forEach((val) => {
                if (val[0].toLowerCase().includes(searchTerm.toLowerCase())) {
                    termExists = true;
                }
            })
        })
        return termExists;
    }

    const renderEmptyOpt = () => {
        if (!searchTerm.length) return null;
        return (<div key={0} className={`fa-select-group-select--content`}>
            <div className={styles.dropdown__filter_select__option_group_container_sec}>
                <div className={`fa-select-group-select--options`}
                    onClick={() => optionClick('', [searchTerm])} >
                    <div>
                        <Text level={7} type={'title'} extraClass={'mr-2'}>Select:</Text>
                    </div>
                    <Text level={7} type={'title'} extraClass={'m-0'} weight={'thin'}>'{searchTerm}'</Text>
                </div>
            </div>
        </div>)
    }

    const renderOptions = (options) => {
        const renderGroupedOptions = [];
        console.log("rebuilding");
        options.forEach((group, grpIndex) => {
            const collState = groupCollapseState[grpIndex] || searchTerm.length > 0;
            const [showFull, setShowFull] = useState(false);
            let hasSearchTerm = false;
            const valuesOptions = [];
            console.log(group.icon)
            const groupItem = (
                <div key={group.label} className={`fa-select-group-select--content`}>
                    {<div className={'fa-select-group-select--option-group'}>
                        <div>
                            <SVG name={group.icon} color={'purple'} extraClass={'self-center'}></SVG>
                            <Text level={8} type={'title'} extraClass={'m-0 ml-2'} weight={'bold'}>{group.label}</Text>
                        </div>
                    </div>}

                    <div className={styles.dropdown__filter_select__option_group_container_sec}>
                        {collState
                            ? (() => {
                                group.values.forEach((val, i) => {
                                    if (val[0].toLowerCase().includes(searchTerm.toLowerCase())) {
                                        hasSearchTerm = true;
                                        valuesOptions.push(
                                            <div key={i} title={val[0]} className={`fa-select-group-select--options`}
                                                onClick={() => optionClick(group.label, val)} >
                                                {searchTerm.length > 0}
                                                <Text level={7} type={'title'} extraClass={'m-0'} weight={'thin'}>{val[0]}</Text>
                                            </div>
                                        );
                                    }
                                });
                                return showFull ? valuesOptions : valuesOptions.slice(0, 5);
                            })()
                            : null
                        }
                    </div>

                    {(valuesOptions.length > 5 && collState) ?
                        !showFull ?
                            <Button
                                className={styles.dropdown__filter_select__showhide}
                                type='text'
                                onClick={() => { setShowFull(true) }} icon={<CaretDownOutlined />}>
                                Show More ({valuesOptions.length - 5})
                            </Button> :
                            <Button
                                className={styles.dropdown__filter_select__showhide}
                                type='text'
                                onClick={() => { setShowFull(false) }} icon={<CaretUpOutlined />}>
                                Show Less
                            </Button> : null
                    }
                </div>
            );
            hasSearchTerm && renderGroupedOptions.push(groupItem);
        });
        if (allowEmpty) {
            renderGroupedOptions.push(renderEmptyOpt());
        }
        return renderGroupedOptions;
    };

    return (
        <>
        <div className='block-header'>abc</div>
            <div className={`${styles.dropdown__filter_select} fa-select fa-select--group-select ${extraClass}`}>
                <Input
                    className={styles.dropdown__filter_select__input}
                    placeholder={placeholder}
                    onKeyUp={onInputSearch}
                    prefix={(<SVG name="search" size={20} color={'grey'} />)}
                />
                <div
                    className={styles.dropdown__filter_select__content}
                >
                    {renderOptions(groupedProperties)}
                </div>
            </div>
            <div className={styles.dropdown__hd_overlay} onClick={onClickOutside}></div>
        </>
    );
}

export default GroupSelect2;
