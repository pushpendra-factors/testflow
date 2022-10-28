import pandas as pd
import numpy as np
import argparse



parser = argparse.ArgumentParser()
parser.add_argument("path", help="path to project files", type=str)
parser.add_argument("num_job", help="num of project per job", type=int)
args = parser.parse_args()


def chunks(L, n):
    """ Yield successive n-sized chunks from L.
    """
    for i in range(0, len(L), n):
        yield L[i:i+n]

class Projects:

    def __init__(self,path,num_per_run) -> None:
        self.path=path
        self.num_per_run = num_per_run
        self.prjs=[]
        self.prjList = []


    def __read_file__(self):
        df = pd.read_csv(self.path)
        column_name="project_id"
        if column_name in df:
            self.prjs = df[column_name].tolist()


    def __prj_list__(self):
        list_prj = []
        for p in chunks(self.prjs,self.num_per_run):
            list_prj.append(p)
        return list_prj

    def gen_list(self):
        self.__read_file__()
        if len(self.prjs)>0:
            r=self.__prj_list__()
        return r




if __name__=="__main__":
    path = args.path
    num_per_job = args.num_job
    prj = Projects(path,num_per_job)
    pr = prj.gen_list()
    for prList in pr:
        print(prList)
    


