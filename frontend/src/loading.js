import React from 'react';
import loadingImage from './assets/img/loading.gif';


const Loading = (props) => {
  return (
    <div style={{paddingTop: '18%', textAlign: 'center'}} className='animated fadeIn fadeOut'>
      <img src={loadingImage} alt='Loading..' />
    </div>
  );
}

export default Loading;