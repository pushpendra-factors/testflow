import React from 'react';

const NoContent = (props) => {
    let paddingTop = props.paddingTop ? props.paddingTop : '12%'
    let containerStyle = { paddingTop: paddingTop };
    if (props.center) containerStyle.textAlign = 'center';

    return (
        <div style={ containerStyle }>
            <span style={{ fontWeight: '700', color: '#BBB', fontSize: '25px' }}>{ props.msg }</span>
        </div>
    );
}

export default NoContent;