import React, { useEffect, Suspense, useState } from 'react';
import { connect, useDispatch } from 'react-redux';
import PageSuspenseLoader from './components/SuspenseLoaders/PageSuspenseLoader';
import * as Sentry from '@sentry/react';
import LogRocket from 'logrocket';
import { FaErrorComp, FaErrorLog } from 'factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';
import factorsai from 'factorsai';
import {
  enableBingAdsIntegration,
  enableMarketoIntegration
} from 'Reducers/global';
import { SSO_LOGIN_FULFILLED } from './reducers/types';
import { sendSlackNotification } from './utils/slack';
import AdBlockerDetector from './components/AdBlockerDetector';
import { AppRoutes } from 'Routes';
import { ProductFruits } from 'react-product-fruits';
import { PRODUCTION_WORKSPACE_CODE } from 'Utils/productFruitsConfig';
import { ScrollToTop } from 'Routes/feature';

function App({
  agent_details,
  active_project,
  enableBingAdsIntegration,
  enableMarketoIntegration,
  plan
}) {
  const dispatch = useDispatch();
  const [userInfo, setUserInfo] = useState(null);

  const ssoLogin = () => {
    if (window.location.href.indexOf('?error=') > -1) {
      let searchParams = new URLSearchParams(window.location.search);
      if (searchParams) {
        let mode = searchParams.get('mode');
        let err = searchParams.get('error');
        if (mode == 'auth0' && err == '') {
          dispatch({ type: SSO_LOGIN_FULFILLED });
        }
      }
    }
  };

  ssoLogin();

  useEffect(() => {
    if (window.location.origin.startsWith('https://tufte-prod.factors.ai')) {
      window.location.replace('https://app.factors.ai/');
    }

    if (window.location.href.indexOf('?code=') > -1) {
      var searchParams = new URLSearchParams(window.location.search);
      if (searchParams) {
        let code = searchParams.get('code');
        let state = searchParams.get('state');
        localStorage.setItem('Linkedin_code', code);
        localStorage.setItem('Linkedin_state', state);
      }
      window.location.replace('/settings/integration');
    }

    if (window.location.href.indexOf('?bingadsint=') > -1) {
      var searchParams = new URLSearchParams(window.location.search);
      if (searchParams) {
        let projectID = searchParams.get('bingadsint');
        let email = searchParams.get('email');
        let projectname = searchParams.get('projectname');
        enableBingAdsIntegration(projectID)
          .then(() => {
            sendSlackNotification(email, projectname, 'Bing Ads');
            window.location.replace('/settings/integration');
          })
          .catch((err) => {
            console.log('bing ads enable error', err);
          });
      }
    }

    if (window.location.href.indexOf('?marketoInt=') > -1) {
      var searchParams = new URLSearchParams(window.location.search);
      if (searchParams) {
        let projectID = searchParams.get('marketoInt');
        let email = searchParams.get('email');
        let projectname = searchParams.get('projectname');
        enableMarketoIntegration(projectID)
          .then(() => {
            sendSlackNotification(email, projectname, 'Marketo');
            window.location.replace('/settings/integration');
          })
          .catch((err) => {
            console.log('Marketo enable error', err);
          });
      }
    }

    if (Sentry) {
      Sentry.setUser({
        username: agent_details?.first_name,
        email: agent_details?.email,
        id: agent_details?.uuid
      });
    }

    // Factorsai init
    if (window.location.href.indexOf('https://app.factors.ai') != -1) {
      factorsai.init('we0jyjxcs0ix4ggnkptymjh48ur8y7q7');
    } else {
      // Development hits will also be pushed to staging to avoid dependency with local api.
      factorsai.init('we0jyjxcs0ix4ggnkptymjh48ur8y7q7', {
        host: 'https://staging-api.factors.ai'
      });
    }

    if (!process.env.NODE_ENV || process.env.NODE_ENV === 'development') {
      // DEV env
    } else {
      // PROD ENV

      //Checking if it is PROD and not STAG
      if (window.location.href.indexOf('https://app.factors.ai/') != -1) {
        //LogRocket
        if (LogRocket) {
          LogRocket.init('anylrg/tufte-prod');
          LogRocket.identify(agent_details?.uuid, {
            name: agent_details?.first_name,
            email: agent_details?.email
          });
          LogRocket.getSessionURL((sessionURL) => {
            Sentry.configureScope((scope) => {
              scope.setExtra('sessionURL', sessionURL);
            });
          });
        }

        // Trackdesk - For affiliation tracking.
        if (window.trackdesk && typeof window.trackdesk == 'function') {
          window.trackdesk('factorsai', 'externalCid', {
            externalCid: agent_details?.email,
            revenueOriginId: 'fc51bc33-de16-4d8f-9270-320a311eb873'
          });
        }

        // Reditus - For affiliation tracking.
        if (window.gr && typeof window.gr == 'function') {
          window.gr('track', 'conversion', { email: agent_details?.email });
        }

        //intercom init and passing logged-in user-data
        var APP_ID = 'rvffkuu7';
        window.intercomSettings = {
          app_id: APP_ID,
          name: agent_details?.first_name,
          email: agent_details?.email,
          user_id: agent_details?.uuid
        };

        (function () {
          var w = window;
          var ic = w.Intercom;
          if (typeof ic === 'function') {
            ic('reattach_activator');
            ic('update', w.intercomSettings);
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
              var s = d.createElement('script');
              s.type = 'text/javascript';
              s.async = true;
              s.src = 'https://widget.intercom.io/widget/' + APP_ID;
              var x = d.getElementsByTagName('script')[0];
              x.parentNode.insertBefore(s, x);
            };
            if (document.readyState === 'complete') {
              l();
            } else if (w.attachEvent) {
              w.attachEvent('onload', l);
            } else {
              w.addEventListener('load', l, false);
            }
          }
        })();
      }
    }

    // if (
    //   window.location.href.indexOf('https://staging-app.factors.ai/') != -1 ||
    //   window.location.href.indexOf('http://factors-dev.com:3000/') != -1
    // ) {
    //   userflow.init('ct_ziy2e3t6sjdj7gh3pqfevszf3y');
    //   userflow.identify(agent_details?.uuid, {
    //     name: agent_details?.first_name,
    //     email: agent_details?.email,
    //     signed_up_at: agent_details?.signed_up_at
    //   });
    // }

    // if (window.location.href.indexOf('https://app.factors.ai/') != -1) {
    //   userflow.init('ct_4iqdnn267zdr5ednpbgbyvubky');
    //   userflow.identify(agent_details?.uuid, {
    //     name: agent_details?.first_name,
    //     email: agent_details?.email,
    //     signed_up_at: agent_details?.signed_up_at
    //   });
    // }
  }, [agent_details]);

  useEffect(() => {
    const tz = active_project?.time_zone;
    // const isTzEnabled = active_project?.is_multiple_project_timezone_enabled;
    if (tz) {
      localStorage.setItem('project_timeZone', tz);
    } else {
      localStorage.setItem('project_timeZone', 'Asia/Kolkata');
    }
  });

  useEffect(() => {
    ssoLogin();
  }, [agent_details]);

  useEffect(() => {
    if (agent_details && plan) {
      const userInfoObj = {
        username: agent_details?.email, // REQUIRED - any unique user identifier
        email: agent_details?.email,
        firstname: agent_details?.first_name,
        signUpAt: agent_details?.signed_up_at,
        props: {
          plan: plan.name
        }
      };
      setUserInfo(userInfoObj);
    }
  }, [agent_details, plan]);

  return (
    <AdBlockerDetector>
      <div className='App'>
        <ErrorBoundary
          fallback={
            <FaErrorComp
              size={'medium'}
              title={'Bundle Error'}
              subtitle={
                'We are facing trouble loading App Bundles. Drop us a message on the in-app chat.'
              }
            />
          }
          onError={FaErrorLog}
        >
          <Suspense fallback={<PageSuspenseLoader />}>
            {userInfo && (
              <ProductFruits
                workspaceCode={PRODUCTION_WORKSPACE_CODE}
                language='en'
                user={userInfo}
                lifeCycle={'unmount'}
              />
            )}
            <ScrollToTop />
            <AppRoutes />
          </Suspense>
        </ErrorBoundary>
      </div>
    </AdBlockerDetector>
  );
}

const mapStateToProps = (state) => ({
  agent_details: state.agent.agent_details,
  active_project: state.global.active_project,
  plan: state.featureConfig.plan
});

export default connect(mapStateToProps, {
  enableBingAdsIntegration,
  enableMarketoIntegration
})(App);
