import React, { Component } from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import {
  Input,
  Card,
  CardBody,
  CardHeader,
  Button,
} from 'reactstrap';
import Toggle from 'react-toggle';
import facebookSvg from '../../assets/img/settings/logo-social-fb-facebook-icon.svg';
import FacebookLogin from 'react-facebook-login';
import { addFacebookAccessToken, fetchProjectSettings } from '../../actions/projectsActions'
import Loading from '../../loading';
import Select from 'react-select';
import { createSelectOpts, makeSelectOpt } from '../../util';

var uniqid = require('uniqid')
const mapStateToProps = store => {
  return {
    currentProjectId: store.projects.currentProjectId,
    currentProjectSettings: store.projects.currentProjectSettings
    }
}
const mapDispatchToProps = dispatch => {
  return bindActionCreators({addFacebookAccessToken, fetchProjectSettings}, dispatch)
}


class Facebook extends Component {
    constructor(props) {
        super(props);
        this.state = {
          response: "",
          adAccounts : "",
          SelectedAdAccount : "",
        }
    }
    componentWillMount() {
      this.props.fetchProjectSettings(this.props.currentProjectId)
        .then((r) => {
          this.setState({ loaded: true });
        })
        .catch((r) => {
          this.setState({loaded: true, error: r.payload });
        });
    }
    responseFacebook = (response) => {
      this.setState({
        response: response
      })
      if(response.id != undefined) {
        fetch(`https://graph.facebook.com/v13.0/${response.id}/adaccounts?access_token=${response.accessToken}`)
        .then(res=> res.json().then((r)=> {
          let adAccounts = r.data.map(account => {
            return {value: account.id, label: account.id} 
          })
          this.setState({
            adAccounts
          })
        }))
        .catch(err=> console.log(err))
      }
    }

    renderFacebookLogin = () => {
      if(!(this.props.currentProjectSettings.int_facebook_access_token)) {
        return (
          <FacebookLogin
            appId={BUILD_CONFIG.facebook_app_id}
            fields="name,email,picture"
            scope="ads_read,email"
            callback={this.responseFacebook}
            cssClass='facebook-css'
          />
        )
      }
      else {
        return (
          <div>Logged In</div>
        )
      }
    }
    getAdAccountsOptSrc() {
      let opts = {}
      for(let i in this.state.adAccounts) {
        let adAccount = this.state.adAccounts[i];
        opts[adAccount.value] = adAccount.label;
      }
      return opts;
    }
    handleChange = e => {
      this.setState({
        SelectedAdAccount : e
      })
    }
    handleSubmit = e => {
      e.preventDefault();
      if (this.state.SelectAdAccount != "" ) {
        const data = {
          "int_facebook_user_id": this.state.response.id,
          "int_facebook_email": this.state.response.email,
          "int_facebook_ad_account": this.state.SelectedAdAccount.value,
          "project_id": this.props.currentProjectId.toString(),
          "int_facebook_access_token": this.state.response.accessToken,
        }
        this.props.addFacebookAccessToken(data).then(()=> this.props.fetchProjectSettings(this.props.currentProjectId)).catch((e)=> console.log(e))
      }
    }
    formComponent = () => {
      if (!(this.props.currentProjectSettings.int_facebook_access_token)) {
        if (this.state.adAccounts != "" && this.state.adAccounts.length != 0) {
          return (
            <div className="p-2">
              <form onSubmit={e => this.handleSubmit(e)}>
                <div className="w-50 pb-2">
                  <h5>Choose your ad account:</h5>
                  <Select
                  value={this.state.SelectedAdAccount}
                  onChange={this.handleChange}
                  options={createSelectOpts(this.getAdAccountsOptSrc())}
                  />
                </div>
                <input className="btn btn-primary shadow-none" type="submit" value="Submit"/>
              </form>
            </div>
            )
        }
        if (this.state.adAccounts != "" && this.state.adAccounts.length == 0) {
          return <div>You don't have any ad accounts associated to the id you logged in with.</div>
        }
    } 
    else {
      if(this.props.currentProjectSettings.int_facebook_ad_account !== "" || this.props.currentProjectSettings.int_facebook_ad_account !== undefined) {
        return <h5 className="p-2 m-2">Selected ad account: {this.props.currentProjectSettings.int_facebook_ad_account}</h5>
      }
    }
    return
  }
    render() {
        return (
          <div className='fapp-content fapp-content-margin'>
            <Card className='fapp-bordered-card'>
                <div>
                    <CardHeader className='fapp-button-header' style={{ marginBottom: '0' }}>
                        <strong>Facebook</strong>
                        <div style={{display: 'inline-block', float: 'right'}}>
                        {this.renderFacebookLogin()}    
                        </div>
                    </CardHeader>
                </div>
                {this.formComponent()}
            </Card>
          </div>
        )
      }
}
export default connect(mapStateToProps, mapDispatchToProps)(Facebook);