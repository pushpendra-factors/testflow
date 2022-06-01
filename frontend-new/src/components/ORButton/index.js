import React from 'react';

import { Button} from 'antd';
import { SVG} from 'factorsComponents';


export default function orButton({index,setOrFilterIndex}){
    return(
        <Button
        type='text'
        index={index}
        onClick={ () => setOrFilterIndex(index)}
        size={'small'}
        className={`fa-btn--custom filter-buttons-margin btn-total-round filter-remove-button plus-button`}
        >
        <SVG name={'union'} />
      </Button>
    );
}