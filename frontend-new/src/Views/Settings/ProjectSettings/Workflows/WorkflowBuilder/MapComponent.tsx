import React, {
    useEffect,
} from 'react';
import {  SVG } from 'factorsComponents';
import { 
    Button, 
    Select,
} from 'antd';

interface MapComponentProps {
    dropdownOptions1: any,
    dropdownOptions2: any,
    propertyMap: Array<[]>,
    setPropertyMap: any,
    limit: Number,
    isTemplate: Boolean
}
const MapComponent = ({
    dropdownOptions1,
    dropdownOptions2,
    propertyMap =[],
    setPropertyMap,
    limit,
    isTemplate
}:MapComponentProps ) => {
    
    useEffect(() => {
        if(!isTemplate){
            setPropertyMap(propertyMap)
        }else if (limit && limit > 0) {
            let arr = Array(limit).fill({});
            setPropertyMap(arr);
        }
    }, [limit])


    const handleChange = (value: String, index: any, type: String) => {
        let obj = {};
        let propertyMapings = propertyMap;
        if (propertyMapings[index]) {
            if (type == 'factors') {
                obj = { ...propertyMapings[index], factors: value }
            }
            else {
                obj = { ...propertyMapings[index], others: value }
            }
            propertyMapings[index] = obj;
        }
        else {
            if (type == 'factors') {
                obj = { factors: value }
            }
            else {
                obj = { others: value }
            }
            propertyMapings = propertyMap.push(obj)
        }
        setPropertyMap(propertyMapings);
    };

    const deletePropertyMap = (el: any) => {
        let newArr = propertyMap?.filter((item, index) => index !== el)
        setPropertyMap(newArr);
    }
    return (<>
        <div className='flex flex-col justify-start'>
            {propertyMap && propertyMap.map((item, index) => {
                return (<>
                    <div className='flex items-center mt-2'>
                        <div className=''>
                            <Select
                                options={dropdownOptions1}
                                onChange={(val) => handleChange(val, index, 'factors')}
                                style={{ width: 300 }}
                                showSearch
                                placeholder="Select property"
                                optionFilterProp="label"
                                value={item?.factors}
                                className='fa-select'
                            />
                        </div>
                        <div className='mr-4 ml-4'>
                            <SVG name='Arrowright' size={20} />
                        </div>
                        <div className=''>
                            <Select
                                options={dropdownOptions2}
                                onChange={(val) => handleChange(val, index, 'other')}
                                style={{ width: 300 }}
                                showSearch
                                placeholder="Select property"
                                optionFilterProp="label"
                                className='fa-select ml-4'
                                value={item?.others}
                            />
                        </div>
                        {!limit &&
                            <div className='ml-2'>
                                <Button size={'small'} icon={<SVG name='trash' size={14} />} onClick={() => deletePropertyMap(index)} />
                            </div>}
                    </div>
                </>)
            })}
        </div>
        {!limit && <Button className='mt-4' style={{ 'width': '200px' }} onClick={(() => setPropertyMap([...propertyMap, {}]))}>Add new</Button>}
    </>)
}
export default MapComponent