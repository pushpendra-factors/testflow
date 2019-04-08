import React from 'react';
import loadingImage from './assets/img/loading.gif';


const Loading = (props) => {
  let paddingTop = !!props.paddingTop ? props.paddingTop: '18%';
  return (
    <div style={{paddingTop: paddingTop, textAlign: 'center'}} className='animated fadeIn fadeOut'>
      <img src={loadingImage} alt='Loading..' />
    </div>
  );
}

export default Loading;