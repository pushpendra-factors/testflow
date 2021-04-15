import React, {useState, useEffect} from 'react';
import styles from './index.module.scss';
import { SVG, Text } from '../factorsComponents';
import { Button, InputNumber, Tooltip } from 'antd';
import GroupSelect from '../QueryComposer/GroupSelect';
import FaDatepicker from '../FaDatepicker';
import FaSelect from '../FaSelect';
import moment from 'moment';

const FAFilterSelect = ({
        propOpts = [], 
        prop = [],
        propSelect,
        operatorOpts = [],
        operator = '=',
        operatorSelect,
        valueOpts = [],
        values = [],
        valuesSelect,
        onDateSelect,
        onNumericalSelect,
        filter
    }, 
        ) => {
    

    useEffect(() => {
        if(!filter) {
            setPropState({icon: '', name: ''});
            setPropSelectOpen(false);
            setOperSelectOpen(false);
            setValuesSelectionOpen(false);

        }
        
    }, [filter])

    useEffect(() => {
        setPropState({icon: prop[2], name: prop[0]});
        setPropSelectOpen(false);
    }, [prop])

    useEffect(() => {
        setOperSelectOpen(false);
    }, [operator])

    useEffect(() => {
        if(values?.length) {
            setValuesSelectionOpen(false);
        }
    }, [values])

    const [propState, setPropState] = useState({
        icon: '',
        name: ''
    });

    const [propSelectOpen, setPropSelectOpen] = useState(false);
    const [operSelectOpen, setOperSelectOpen] = useState(false);
    const [valuesSelectionOpen, setValuesSelectionOpen] = useState(false);

    const setNumericalValue = (ev) => {
        onNumericalSelect(ev);
    }

    const parseDateRangeFilter = (fr, to) => {
        const fromVal = fr? fr: new Date(moment().startOf('day')).getTime();
        const toVal = to? to: new Date(moment()).getTime();
        return {
            from: fromVal,
            to: toVal
        }
        // return (moment(fromVal).format('MMM DD, YYYY') + ' - ' +
        //           moment(toVal).format('MMM DD, YYYY'));
      }

    const renderPropSelect = () => {
        return (<div className={styles.filter__propContainer}>
            
            <Button 
                icon={propState && propState.icon? <SVG name={propState.icon} size={16} color={'purple'} />: null} 
                className={`fa-button--truncate`} 
                type="link" 
                onClick={() => setPropSelectOpen(!propSelectOpen)}> {propState?.name? propState?.name : 'Select Property'} 
            </Button>

            {propSelectOpen && 
                <GroupSelect
                    groupedProperties={propOpts}
                    placeholder="Select Property"
                    optionClick={(group, val) => propSelect([...val, group])}
                    onClickOutside={() => setPropSelectOpen(false)}

                ></GroupSelect>
            }

        </div>)
    }

    const renderOperatorSelector = () => {
        return (<div className={styles.filter__propContainer}>
            
            <Button 
                className={`fa-button--truncate ml-2`} 
                type="link" 
                onClick={() => setOperSelectOpen(!operSelectOpen)}> {operator? operator : 'Select Operator'} 
            </Button>

            {operSelectOpen && 
                <FaSelect 
                    options={operatorOpts.map(op => [op])}
                    optionClick={(val) => operatorSelect(val)}
                    onClickOutside={() => setOperSelectOpen(false)}
                >
                </FaSelect>
            }

        </div>)
    }

    const renderValuesSelector = () => {
        let selectionComponent;
        if(prop[1] === 'categorical') {
            selectionComponent = (<FaSelect 
                multiSelect={true}
                options={valueOpts.map(op => [op])}
                optionClick={(val) => valuesSelect(val[0])}
                onClickOutside={() => setValuesSelectionOpen(false)}
                selectedOpts={values}
                allowSearch={true}
                >
                </FaSelect>);
        }

        if(prop[1] === 'datetime') {
            const parsedValues = values && values.length ? JSON.parse(values) : {};
            const dateRange = parseDateRangeFilter(parsedValues.fr, parsedValues.to)
            const rang = {
                startDate: dateRange.from,
                endDate: dateRange.to,
              }
            
            selectionComponent = (<FaDatepicker
                customPicker
                presetRange
                monthPicker
                placement="topRight"
                range={rang}
                onSelect={(rng) => onDateSelect(rng)}
              />)
        }

        if(prop[1] === 'numerical') {
            selectionComponent = (<InputNumber onChange={setNumericalValue}></InputNumber>);
        }

        return (<div className={`${styles.filter__propContainer} ml-4`}>
                {prop[1] === 'categorical'?<> <Tooltip title={values && values.length? values.join(', ') : null}><Button 
                    className={`fa-button--truncate`} 
                    type="link" 
                    onClick={() => setValuesSelectionOpen(!valuesSelectionOpen)}> {values && values.length? values.join(', ') : 'Select Values'} 
                </Button> </Tooltip> {valuesSelectionOpen && selectionComponent} </> : null }

                {prop[1] !== 'categorical'? selectionComponent : null }
        </div>)
        
    }

    return (<div className={styles.filter}>
        {renderPropSelect()}

        {propState?.name ? renderOperatorSelector() : null}

        {operator? renderValuesSelector(): null}

    </div>);
    
}

export default FAFilterSelect;