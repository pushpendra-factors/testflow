import React from 'react';
import { Dropdown, Button, Menu } from 'antd';
import { SVG } from '../factorsComponents';
import styles from './index.module.scss';

function ChartTypeDropdown({ menuItems, onClick, chartType }) {

    const menu = (
        <Menu className={styles.dropdownMenu}>
            {menuItems.map(item => {
                return (
                    <Menu.Item key={item.key} onClick={onClick} className={`${styles.dropdownMenuItem} ${chartType === item.key ? styles.active : ''}`}>
                        <div className={`flex items-center`}>
                            <SVG extraClass="mr-1" name={item.key} size={25} color={chartType === item.key ? '#8692A3' : '#3E516C'} />
                            <span className="mr-3">{item.name}</span>
                            {chartType === item.key ? (
                                <SVG name="checkmark" size={17} color="#8692A3" />
                            ) : null}
                        </div>
                    </Menu.Item>
                )
            })}
        </Menu>
    );

    return (
        <Dropdown overlay={menu}>
            <Button className={`ant-dropdown-link flex items-center ${styles.dropdownBtn}`}>
                <SVG name={chartType} size={25} color="#0E2647" />
                <SVG name={'dropdown'} size={25} color="#3E516C" />
            </Button>
        </Dropdown>
    )
}

export default ChartTypeDropdown;