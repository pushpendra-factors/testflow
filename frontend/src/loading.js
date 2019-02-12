import React from 'react';
import loadingImage from './assets/img/loading.gif';


const Loading = (props) => {
  return (
    <div style={{paddingTop: '18%', textAlign: 'center'}}>
      <img src={loadingImage} alt='Loading..' />
    </div>
  );
}

export default Loading;