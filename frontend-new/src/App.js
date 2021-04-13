import React, { useEffect, lazy, Suspense } from "react";
// import AppLayout from "./Views/AppLayout";
// import Login from "./Views/Pages/Login";
// import ForgotPassword from "./Views/Pages/ForgotPassword";
// import ResetPassword from "./Views/Pages/ResetPassword";
// import SignUp from "./Views/Pages/SignUp";
// import Activate from "./Views/Pages/Activate";
// import FactorsInsights from "./Views/Factors/FactorsInsights";
import { connect } from "react-redux";
import {
  BrowserRouter as Router,
  Route,
  Switch,
  Redirect,
} from "react-router-dom";
import PageSuspenseLoader from "./components/SuspenseLoaders/PageSuspenseLoader";
import * as Sentry from "@sentry/react";
import LogRocket from "logrocket";
import retryDynamicImport from 'Utils/dynamicImport';
import { FaErrorComp, FaErrorLog } from 'factorsComponents';
import {ErrorBoundary} from 'react-error-boundary';

const AppLayout = lazy(()=>retryDynamicImport(() => import("./Views/AppLayout")));
const Login = lazy(()=>retryDynamicImport(() => import("./Views/Pages/Login")));
const ForgotPassword = lazy(()=>retryDynamicImport(() => import("./Views/Pages/ForgotPassword")));
const ResetPassword = lazy(()=>retryDynamicImport(() => import("./Views/Pages/ResetPassword")));
const SignUp = lazy(()=>retryDynamicImport(() => import("./Views/Pages/SignUp")));
const Activate = lazy(()=>retryDynamicImport(() => import("./Views/Pages/Activate")));
const FactorsInsights = lazy(()=>retryDynamicImport(() => import("./Views/Factors/FactorsInsights")));

function App({ isAgentLoggedIn, agent_details }) {

  useEffect(() => {

    if (window.location.origin.startsWith("https://tufte-prod.factors.ai")) {
      window.location.replace("https://app.factors.ai/")
    }

    if (window.location.href.indexOf("?code=") > -1) {
      var searchParams = new URLSearchParams(window.location.search);
      if (searchParams) {
        let code = searchParams.get("code");
        let state = searchParams.get("state");
        console.log('code,state', code, state);
        localStorage.setItem('Linkedin_code', code);
        localStorage.setItem('Linkedin_state', state);
      }
      window.location.replace("/settings/#integrations");
    }


    if (Sentry) {
      Sentry.setUser({
        username: agent_details?.first_name,
        email: agent_details?.email,
        id: agent_details?.uuid,
      });
    }



    if (!process.env.NODE_ENV || process.env.NODE_ENV === 'development') {
      // DEV env
    } else {
      // PROD ENV

      //LogRocket
      if (LogRocket) {
        LogRocket.init('anylrg/tufte-prod');
        LogRocket.identify(agent_details?.uuid, {
          name: agent_details?.first_name,
          email: agent_details?.email,
        });
        LogRocket.getSessionURL(sessionURL => {
          Sentry.configureScope(scope => {
            scope.setExtra("sessionURL", sessionURL);
          });
        });
      }

      //intercom init and passing logged-in user-data 
      var APP_ID = "rvffkuu7";
      window.intercomSettings = {
        app_id: APP_ID,
        name: agent_details?.first_name,
        email: agent_details?.email,
        user_id: agent_details?.uuid,
      };

      (function () {
        var w = window;
        var ic = w.Intercom;
        if (typeof ic === "function") {
          ic("reattach_activator");
          ic("update", w.intercomSettings);
        } else {
          var d = document;
          var i = function () {
            i.c(arguments);
          };
          i.q = [];
          i.c = function (args) {
            i.q.push(args);
          };
          w.Intercom = i;
          var l = function () {
            var s = d.createElement("script");
            s.type = "text/javascript";
            s.async = true;
            s.src = "https://widget.intercom.io/widget/" + APP_ID;
            var x = d.getElementsByTagName("script")[0];
            x.parentNode.insertBefore(s, x);
          };
          if (document.readyState === "complete") {
            l();
          } else if (w.attachEvent) {
            w.attachEvent("onload", l);
          } else {
            w.addEventListener("load", l, false);
          }
        }
      })();

    }
  }, [agent_details]);

  return (
    <div className="App">
       <ErrorBoundary fallback={<FaErrorComp size={'medium'} title={'Bundle Error'} subtitle={ "We are facing trouble loading App Bundles. Drop us a message on the in-app chat."} />} onError={FaErrorLog}> 
      <Suspense fallback={<PageSuspenseLoader />}>
        <Router>
          <Switch>
            <Route exact path="/signup" name="login" component={SignUp} />
            <Route
              exact
              path="/activate"
              name="Activate"
              component={Activate}
            />
            <Route
              exact
              path="/setpassword"
              name="login"
              component={ResetPassword}
            />
            <Route
              exact
              path="/forgotpassword"
              name="login"
              component={ForgotPassword}
            />
            <Route exact path="/login" name="login" component={Login} />
            {isAgentLoggedIn ? (
              <Route
                exact
                path="/explain/insights"
                name="login"
                component={FactorsInsights}
              />
            ) : (
              <Redirect to="/login" />
            )}
            {isAgentLoggedIn ? (
              <Route path="/" name="Home" component={AppLayout} />
            ) : (
              <Redirect to="/login" />
            )}
          </Switch>
        </Router>
      </Suspense>
      </ErrorBoundary>
    </div>
  );
}

const mapStateToProps = (state) => ({
  isAgentLoggedIn: state.agent.isLoggedIn,
  agent_details: state.agent.agent_details,
});

export default connect(mapStateToProps)(App);
