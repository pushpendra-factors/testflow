import React from 'react';
import {Route, Redirect} from 'react-router-dom';
import { connect } from "react-redux";
import {isFromFactorsDomain} from './util';
 
const mapStateToProps = store => {
    return {
      currentAgent: store.agents.agent,
    };
  }

/**
 * Protected Routes in React using React Router
 * 
 * https://stackoverflow.com/a/43695765/3968921
 * https://www.youtube.com/watch?v=Y0-qdp-XBJg
 */

class InternalRoute extends React.Component{
    
    showInternalRoute(){
        return isFromFactorsDomain(this.props.currentAgent.email);
    }

    render(){
        const {component: Component, ...rest} = this.props;
        
        if(!this.showInternalRoute()){
            return(
                <Redirect to={{
                    pathname: "/",
                }} />
            )
        }
        
        return(
            <Route
                {...rest}
                render={(props)=>{
                    return <Component {...props} />
                }}
            />
        )
    }
}

export default connect(mapStateToProps, null)(InternalRoute);