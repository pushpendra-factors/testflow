import { BaseService } from "../../classes/BaseServiceClass";

export class OTPService extends BaseService {

    getTouchPoints = () => {
        return this.get("/otp_rules");
    }

    createTouchPoint = (otp) => {
        return this.post("/otp_rules", otp);
    }

    modifyTouchPoint = (otp) => {
        return this.put("/otp_rules/" + otp.id, otp);
    }

    removeTouchPoint = (otpId) => {
        return this.del("/otp_rules/" + otpId);
    }
    
}




