import React, { useState, useEffect } from 'react';
import { useSelector } from 'react-redux';
import styles from './index.module.scss';
import { SVG, Text } from 'factorsComponents';
import { Button, InputNumber, Tooltip,Input } from 'antd';
import GroupSelect2 from 'Components/QueryComposer/GroupSelect2';
import FaDatepicker from 'Components/FaDatepicker';
import FaSelect from 'Components/FaSelect';
import moment from 'moment';
import { isArray } from 'lodash';
import { DISPLAY_PROP } from '../../../../../../../utils/constants';

const defaultOpProps = {
    "categorical": [
      '=',
      '!=',
      'contains',
      'does not contain'
    ],
    "numerical": [
      '=',
      '!=',
      '<',
      '<=',
      '>',
      '>='
    ],
    "datetime": [
      '='
    ]
  };

const GlobalFilterSelect = ({
    propOpts = [],
    operatorOpts=defaultOpProps,
    valueOpts=[],
    setValuesByProps,
    applyFilter,
    filter
},
) => {


    const [propState, setPropState] = useState({
        icon: '',
        name: '',
        type: ''
    });

    const [operatorState, setOperatorState] = useState("=");
    const [valuesState, setValuesState] = useState(null);

    const [propSelectOpen, setPropSelectOpen] = useState(false);
    const [operSelectOpen, setOperSelectOpen] = useState(false);
    const [valuesSelectionOpen, setValuesSelectionOpen] = useState(false);

    const [updateState, updateStateApply] = useState(false);

    const {userPropNames, eventPropNames} = useSelector((state) => state.coreQuery)
    const [containButton, setContainButton] = useState(true);

    useEffect(() => {
        if (filter) {
            const prop = filter.props;
            setPropState({ icon: prop[2], name: prop[0], type: prop[1] });
            setOperatorState(filter.operator);
            // Set values state
            setValues();
            setPropSelectOpen(false);
            setOperSelectOpen(false);
            setValuesSelectionOpen(false);

        }

    }, [filter])

    useEffect(() => {
        if(updateState && valuesState && propState.type !== 'numerical') {
            emitFilter();
            updateStateApply(false);
        }
    }, [updateState])


    const setValues = () => {
        let values;
        if (filter.props[1] === 'datetime') { 
            const parsedValues = (filter.values[0] ? (typeof filter.values[0] === 'string')? JSON.parse(filter.values) : filter.values : {});
            values = parseDateRangeFilter(parsedValues.fr, parsedValues.to)
        } else {
            values = filter.values;
        }
        setValuesState(values);
    }


    const emitFilter = () => {
        if(propState && operatorState && valuesState) {
            applyFilter({
                props: [propState.name, propState.type, propState.icon],
                operator: operatorState,
                values: valuesState
            })
        }
    }

    const operatorSelect = (op) => {
        setOperatorState(op);
        setValuesState(null);
        setOperSelectOpen(false);
    }

    const renderDisplayName = (propState) => {
        // propState?.name ? userPropNames[propState?.name] ? userPropNames[propState?.name] : propState?.name : 'Select Property'
        let propertyName = '';
        if(propState.name && (propState.icon === 'user' || propState.icon === 'user_g')) {
          propertyName = userPropNames[propState.name]?  userPropNames[propState.name] : propState.name;
        }
        if(propState.name && propState.icon === 'event') {
          propertyName = eventPropNames[propState.name]?  eventPropNames[propState.name] : propState.name;
        }
        if(!propState.name) {
          propertyName = 'Select Property';
        }
        return propertyName;
      }

    const propSelect = (prop) => {
        setPropState({ icon: prop[3], name: prop[1], type: prop[2] });
        setPropSelectOpen(false);
        setOperatorState("=");
        setValuesState(null);
        setValuesByProps(prop);
    }

    const valuesSelect = (val) => {
        setValuesState(val.map(vl => JSON.parse(vl)[0]));
        setValuesSelectionOpen(false);
        updateStateApply(true);
    }

    const onDateSelect = (rng) => {
        let startDate;
        let endDate;
        if(isArray(rng.startDate)) {
            startDate = rng.startDate[0].toDate().getTime();
            endDate = rng.startDate[1].toDate().getTime();
        } else {
            if(rng.startDate && rng.startDate._isAMomentObject){
                startDate = rng.startDate.toDate().getTime();
            } else {
                startDate = rng.startDate.getTime();
            }
    
            if(rng.endDate && rng.endDate._isAMomentObject){
                endDate = rng.endDate.toDate().getTime();
            } else {
                endDate = rng.endDate.getTime();
            }
        }
        
        const rangeValue = {
            "fr": startDate,
            "to": endDate,
            "ovp": false
        }

        setValuesState(JSON.stringify(rangeValue));
        updateStateApply(true);
    }
    const setNumericalValue = (ev) => {
        // onNumericalSelect(ev);

        setValuesState(String(ev.target.value).toString());
    }

    const parseDateRangeFilter = (fr, to) => {
        const fromVal = fr ? fr : new Date(moment().startOf('day')).getTime();
        const toVal = to ? to : new Date(moment()).getTime();
        return {
            from: fromVal,
            to: toVal,
            ovp: false
        }
        // return (moment(fromVal).format('MMM DD, YYYY') + ' - ' +
        //           moment(toVal).format('MMM DD, YYYY'));
    }

    const renderGroupDisplayName = (propState) => {
        let propertyName = '';
        if(!propState.name) {
          propertyName = 'Select Property';
        } else {
            propertyName = propState.name;
        }
        return propertyName;
      }

    const renderPropSelect = () => {
        return (<div className={styles.filter__propContainer}>

            <Tooltip title={renderGroupDisplayName(propState)}>
                <Button
                    icon={propState && propState.icon ? <SVG name={propState.icon} size={16} color={'purple'} /> : null}
                    className={`fa-button--truncate fa-button--truncate-xs`}
                    type="link"
                    onClick={() => setPropSelectOpen(!propSelectOpen)}> {renderDisplayName(propState)}
                </Button>
            </Tooltip>
            {propSelectOpen &&
                <GroupSelect2
                    groupedProperties={propOpts}
                    placeholder="Select Property"
                    optionClick={(group, val) => propSelect([...val, group])}
                    onClickOutside={() => setPropSelectOpen(false)}
                ></GroupSelect2>
            }

        </div>)
    }

    const renderOperatorSelector = () => {
        return (<div className={styles.filter__propContainer}>

            <Button
                className={`fa-button--truncate ml-2`}
                type="link"
                onClick={() => setOperSelectOpen(true)}> {operatorState ? operatorState : 'Select Operator'}
            </Button>

            {operSelectOpen &&
                <FaSelect
                    options={operatorOpts[propState.type].map(op => [op])}
                    optionClick={(val) => operatorSelect(val)}
                    onClickOutside={() => setOperSelectOpen(false)}
                >
                </FaSelect>
            }

        </div>)
    }

    const renderValuesSelector = () => {
        let selectionComponent = null;
        const values = [];
        
        selectionComponent = (<FaSelect
                multiSelect={true}
                options={valueOpts && valueOpts[propState.name]?.length ? valueOpts[propState.name].map(op => [op]) : []}
                applClick={(val) => valuesSelect(val)}
                onClickOutside={() => setValuesSelectionOpen(false)}
                selectedOpts={valuesState ? valuesState : []}
                allowSearch={true}
            >
            </FaSelect>);
        

        if (propState.type === 'datetime') {
            const parsedValues = (valuesState ? (typeof valuesState === 'string')? JSON.parse(valuesState) : valuesState : {});
            const fromRange = parsedValues.fr? parsedValues.fr : parsedValues.from;
            const dateRange = parseDateRangeFilter(fromRange, parsedValues.to);
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
                onSelect={(rng) => onDateSelect(rng)
                }
            />)
        }

        if (propState.type === 'numerical') {
            selectionComponent = (
                <div>
                {containButton && (
                  <Button
                  className={`fa-button--truncate filter-buttons-radius filter-buttons-margin`}
                  type='link'
                  onClick={() => setContainButton(false)}
                >
                  {valuesState ? valuesState : 'Enter Value'}
                </Button>)}
                {!containButton &&
                (<Input
                  type="number"
                  value={valuesState}
                  placeholder={'Enter Value'}
                  onPressEnter={()=>{
                    emitFilter()
                    setContainButton(true)}}
                  onChange={setNumericalValue}
                  className={`input-value filter-buttons-radius filter-buttons-margin`}
                ></Input>)
                }
                </div>
              );
        }
        if(!operatorState || !propState?.name) return null;

        return (
          <div className={`${styles.filter__propContainer} ml-2`}>
            {propState.type === 'categorical' ? (
              <Tooltip
                title={
                  valuesState && valuesState.length
                    ? valuesState
                        .map((vl) => (DISPLAY_PROP[vl] ? DISPLAY_PROP[vl] : vl))
                        .join(', ')
                    : null
                }
              >
                <Button
                  className={`fa-button--truncate`}
                  type='link'
                  onClick={() => setValuesSelectionOpen(!valuesSelectionOpen)}
                >
                  {' '}
                  {valuesState && valuesState.length
                    ? valuesState
                        .map((vl) => (DISPLAY_PROP[vl] ? DISPLAY_PROP[vl] : vl))
                        .join(', ')
                    : 'Select Values'}
                </Button>{' '}
              </Tooltip>
            ) : null}

            {valuesSelectionOpen || propState.type !== 'categorical'
              ? selectionComponent
              : null}
          </div>
        );

    }

    return (<div className={styles.filter}>
        {renderPropSelect()}

        {propState?.name ? renderOperatorSelector() : null}

        {operatorState? renderValuesSelector() : null}

    </div>);

}

export default GlobalFilterSelect;