import moment from 'moment-timezone'; 

const MomentTz = (props) => {
    
  const timeZone = localStorage.getItem('project_timeZone') || 'Asia/Kolkata';  
  moment.tz.setDefault(timeZone);
    if(props){
        // console.table("MomentTz timezone, props, formatted, unix -->",timeZone, props, moment.tz(`${props}`,timeZone).format(), moment.tz(`${props}`,timeZone).unix());
        // return moment.tz(`${props}`,timeZone)
        return moment(props)
    } 
  return moment([])
}
export default MomentTz 
 


