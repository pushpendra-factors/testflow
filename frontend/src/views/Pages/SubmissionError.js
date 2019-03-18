import React from 'react';

const SubmissionError = (props) => {
    if(!props.message || props.message == "") return null;

    let marginTop = props.marginTop ? props.marginTop : '-20px';
    return (
        <div style={{marginTop: marginTop, marginBottom: '10px', color:'#d64541', fontWeight: '700', textAlign: 'center'}}>
            <span>{ props.message }</span>
        </div>
    );
}

export default SubmissionError;
