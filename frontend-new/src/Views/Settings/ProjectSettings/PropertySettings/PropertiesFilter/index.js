import React, {useEffect, useState} from 'react';
import { connect } from 'react-redux';
import styles from './index.module.scss';
import { SVG, Text } from 'factorsComponents';
import { Button, Tooltip, Input } from 'antd';
import GroupSelect from '../../../../../components/QueryComposer/GroupSelect';
import FaSelect from '../../../../../components/FaSelect';

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

function PropertyFilter({activeProject, propOpts = [], filter, insertFilter}) {

    const [propState, setPropState] = useState({});
    const [propSelectOpen, setPropSelectOpen] = useState(false);

    const [operatorState, setOperatorState] = useState("=");
    const [operSelectOpen, setOperSelectOpen] = useState(false);

    const [valueState, setValueState] = useState('');

    useState(() => {
        if(filter && filter.prop) {
            setPropState(filter.prop);
            setOperatorState(filter.operator);
            setValueState(filter.values);
        }
    }, [filter])
    

    const propSelect = (prop) => {
        setPropState({
            type: prop[2],
            name: prop[0],
            category: prop[1]
        })
        setPropSelectOpen(false);
        setOperatorState("=");
        setValueState("");
    }

    const operatorSelect = (op) => {
        setOperatorState(op[0]);
        setOperSelectOpen(false);
        setValueState("");
    }

    const captureInputValue = (val) => {
        setValueState(val.target.value);
    }

    const emitValue = (ev) => {
        if(valueState) {
            const filterToUpdate = {...filter};
            filterToUpdate.prop = propState;
            filterToUpdate.operator = operatorState;
            filterToUpdate.values = valueState;
            insertFilter(filterToUpdate);
        }
    }

    const renderPropSelect = () => {
        return (<div className={`${styles.filter__propContainer}`}>
            
            <Button 
                icon={propState && propState.icon? <SVG name={propState.type} size={16} color={'purple'} />: null} 
                className={`fa-button--truncate`} 
                type="link" 
                onClick={() => setPropSelectOpen(!propSelectOpen)}> {propState?.name? propState?.name : 'Select Property'} 
            </Button>

            {propSelectOpen && 
                <GroupSelect
                    extraClass={`fa-grp_noshadow fa-grp_pos-btn`}
                    groupedProperties={propOpts}
                    placeholder="Select Property"
                    optionClick={(group, val) => propSelect([...val, group])}
                    onClickOutside={() => setPropSelectOpen(false)}

                ></GroupSelect>
            }

        </div>)
    }

    const renderOperatorSelector = () => {
        if(!propState || !propState.name) return null;
        return (<div className={`${styles.filter__propContainer} ml-2`}>
            
            <Button 
                className={`fa-button--truncate`} 
                type="link" 
                onClick={() => setOperSelectOpen(!operSelectOpen)}> {operatorState? operatorState : 'Select Operator'} 
            </Button>

            {operSelectOpen && 
                <FaSelect 
                    extraClass={`fa-grp_noshadow fa-grp_pos-btn`}
                    options={defaultOpProps[propState.category].map(op => [op])}
                    optionClick={(val) => operatorSelect(val)}
                    onClickOutside={() => setOperSelectOpen(false)}
                >
                </FaSelect>
            }

        </div>)
    }

    const renderValuesSelector = () => {
        if(!operatorState || !propState || !propState.name) return null;

        let selectionComponent;
        selectionComponent = (<Input value={valueState} onBlur={emitValue} onChange={captureInputValue}></Input>);
        

        return (<div className={`${styles.filter__propContainer} ml-2`}>
                {selectionComponent}
        </div>)
        
    }

    return (<div className={`${styles.filter} mt-4`}>
        {renderPropSelect()}
        {renderOperatorSelector()}
        {renderValuesSelector()}
    </div>)
}

const mapStateToProps = (state) => ({
    activeProject: state.global.active_project,
});

export default connect(mapStateToProps, {})(PropertyFilter);