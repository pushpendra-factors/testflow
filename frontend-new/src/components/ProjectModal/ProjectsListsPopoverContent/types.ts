export interface ProjectListsPopoverContentType {
  variant: string;
  currentAgent: any;
  active_project: any;
  projects: Array<any>;
  showProjectsList: any;
  setShowPopOver: any;
  showUserSettingsModal: any;
  userLogout: () => void;
  setShowProjectsList: any;
  ShowPopOver?: boolean;
  searchProject: any;
  searchProjectName: string;
  setchangeProjectModal: any;
  setselectedProject: any;
  isFreePlan?: any;
}
