import pdb
import json
from collections import defaultdict
import os
from .count_experiment import perform_count_experiment
from .mbd_emulator import perform_mbd_experiment

def main():
    perform_count_experiment()

if __name__ == "__main__":
    main()
