import React from 'react';
import lazyWithRetry from 'Utils/lazyWithRetry';
import { Text } from 'factorsComponents';
import { Button } from 'antd';
import animationData from './assets/lottie/38064-error-cone.json';

const Lottie = lazyWithRetry(
  () => import(/* webpackChunkName: "animation" */ 'react-lottie')
);

const defaultOptions = {
  loop: true,
  autoplay: true,
  animationData,
  rendererSettings: {
    preserveAspectRatio: 'xMidYMid slice'
  }
};

class ErrorBoundary extends React.Component {
  constructor(props) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error) {
    // Update state so the next render will show the fallback UI.
    return { hasError: true };
  }

  componentDidCatch(error, errorInfo) {
    // You can also log the error to an error reporting service
    //   logErrorToMyService(error, errorInfo);
    console.log('logErrorToConsole:');
    console.error(error, errorInfo);
  }

  reloadPage() {
    window.location = '/';
    // window.location.reload();
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className='fa-container mt-24 flex flex-col items-center'>
          <Lottie options={defaultOptions} height={200} width={200} />
          <Text
            type='title'
            align='center'
            level={3}
            weight='bold'
            extraClass='ml-2 m-0'
          >
            Oops! Something went wrong.
          </Text>
          <Text
            type='title'
            align='center'
            level={5}
            weight='thin'
            extraClass='ml-2 m-0'
          >
            We're experiencing an internal server problem.
          </Text>
          <Button size='large' className='mt-4' onClick={this.reloadPage}>
            Try again!
          </Button>
        </div>
      );
    }

    return this.props.children;
  }
}

export default ErrorBoundary;
