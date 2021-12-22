import React from 'react';  
import NumberFormat from 'react-number-format';
import { isArray } from 'lodash';

const Number = ({
type, number, className, shortHand=false, suffix='', prefix=''
}) => {  
 
    const abbreviateNumber = n => {
        if (n < 1e3) return n;
        if (n >= 1e3 && n < 1e6) return +(n / 1e3).toFixed(1) + "K";
        if (n >= 1e6 && n < 1e9) return +(n / 1e6).toFixed(1) + "M";
        if (n >= 1e9 && n < 1e12) return +(n / 1e9).toFixed(1) + "B";
        if (n >= 1e12) return +(n / 1e12).toFixed(1) + "T";
      };

    const finalVal = shortHand ? abbreviateNumber(number) : number; 

    
    
    return (
        <span className={className}> 
            {shortHand ? `${prefix}${abbreviateNumber(number)}${suffix}` :
            <NumberFormat displayType={'text'} value={isArray(finalVal)? finalVal[0] : finalVal} thousandSeparator={true} decimalScale={1} suffix={suffix} prefix={prefix} /> 
            }
        </span>
    );
}


export default Number;
