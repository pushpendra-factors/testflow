import React, {useState, useEffect} from 'react';
import styles from './index.module.scss';
import { SVG, Text } from '../../components/factorsComponents';
import { Button } from 'antd';

const FaToggleBtn = ({label, icon, state, onToggle}) => {

    const [toggleState, setToggleState] = useState(false);

    useEffect(() => {
        setToggleState(state);
    }, [state])


    // return (<div onClick={() => onToggle(label)} className={`${styles.btnContainer} ${toggleState && styles.active}`}>
    //     {icon && <SVG extraClass={styles.icon} name={icon}></SVG>}
        
    //     <span className={styles.label}>{label}</span>
    // </div>)
    return (<Button size={'large'} onClick={() => onToggle(label)} icon={icon ? <SVG name={icon} /> : null} className={`ml-2 ${toggleState ? 'active' : '' }`}>  
        <span>{label}</span>
    </Button>)

}

export default FaToggleBtn;