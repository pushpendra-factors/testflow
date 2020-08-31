import React, { useState } from 'react';
import { Modal, Switch, Input } from 'antd';

function FiltersInfo({ setDrawerVisible, eventsData, handleEventsVisibilityChange, handleGroupDataChange }) {
    
    const [visible, setVisible] = useState(false);

    const handleOk = () => {
        setVisible(false);
    }

    const handleCancel = () => {
        setVisible(false);
    }

    return (
        <>
            <Modal
                title="Edit Data"
                visible={visible}
                onOk={handleOk}
                onCancel={handleCancel}
            >
                {eventsData.map(elem=>{
                    return (
                        <div key={elem.index} className="mt-8">
                            <div className='my-4 font-bold text-2xl'>{elem.name}</div>
                            <Switch onChange={(value)=>handleEventsVisibilityChange(value, elem.index)} checkedChildren="visible" unCheckedChildren="not visible" checked={elem.display} />
                            {Object.keys(elem.data).map(d=>{
                                return (
                                    <Input type="number" onChange={(e)=>handleGroupDataChange(e.target.value, elem.index, d)} key={d} className="my-4" addonBefore={d} value={elem.data[d]} />
                                )
                            })}
                        </div>
                    )
                })}
            </Modal>
            <div className="mt-4 flex justify-end p-4">
                <a className="flex items-center" onClick={setVisible.bind(this, true)}>
                    <span className="mr-1">
                        <svg width="17" height="18" viewBox="0 0 17 18" fill="none" xmlns="http://www.w3.org/2000/svg">
                            <path d="M16 5.12653L12 1.12653C12.6069 0.400054 14.7336 -0.139838 16 1.12653C17.3554 2.48196 16.7265 4.51967 16 5.12653Z" fill="black" />
                            <path fillRule="evenodd" clipRule="evenodd" d="M0 17.1265L2.01335 11.105L10.9308 2.15259L14.9308 6.15259L6.05911 15.0904L0 17.1265ZM3.17055 13.9512L3.76138 12.1841L10.9336 4.98378L12.1076 6.15782L4.97321 13.3454L3.17055 13.9512Z" fill="black" />
                        </svg>
                    </span>
                    <span style={{ color: "#8692A3" }}>Edit Data</span>
                </a>
            </div>
        </>
    )
}

export default FiltersInfo;