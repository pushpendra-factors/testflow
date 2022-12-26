import { BaseService } from '../../classes/BaseServiceClass';

export class OTPService extends BaseService {
  constructor(dispatch, projectId) {
    super(dispatch, projectId);
  }

  getTouchPoints = () => {
    return this.get('/otp_rules');
  };

  createTouchPoint = (otp) => {
    return this.post('/otp_rules', otp);
  };

  modifyTouchPoint = (otp, id) => {
    return this.put('/otp_rules/' + id, otp);
  };

  removeTouchPoint = (otpId) => {
    return this.del('/otp_rules/' + otpId);
  };
}
