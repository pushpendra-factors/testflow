import React from 'react';
import * as icons from '../svgIcons';
import Lottie from 'react-lottie';
import animationData from '../../assets/lottie/fa-loader.json';

const defaultOptions = {
    loop: true,
    autoplay: true,
    animationData: animationData,
    rendererSettings: {
      preserveAspectRatio: "xMidYMid slice"
    }
  };


class Spiner extends React.Component { 
  render() {
    const {
      name, size, color, className
    } = this.props; 
    const sizeCal = (size) => {
        switch(size){
            case 'large': return 200; break;
            case 'medium': return 100; break;
            case 'small': return 50; break;
            default: return 100; break; 
        } 
    }
    let finalSize = sizeCal(size);
    console.log('extraClass',className)
    return (
        <div className={className}>
            <Lottie 
            options={defaultOptions}
            height={finalSize}
            width={finalSize}
            />  
        </div>
    );
  }
}

export default Spiner;
