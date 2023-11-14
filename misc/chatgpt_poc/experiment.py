from data_preparer import get_prepared_data
from sklearn.model_selection import train_test_split
from chat import chat_once_mode
import pdb
import os
os.environ['TF_CPP_MIN_LOG_LEVEL'] = '1'
os.environ['TOKENIZERS_PARALLELISM'] = 'false'

def qualitative_comparison_experiment():
    print_final_prompts = False
    df = get_prepared_data()
    # pdb.set_trace()
    tr_df, te_df = train_test_split(df, test_size=0.2)
    tr_data = list(tr_df[['prompt', 'completion', 'orig_completion']].T.to_dict().values())
    te_data = list(te_df[['prompt', 'completion', 'orig_completion']].T.to_dict().values())
    for te_datum in te_data[:2]:
        question = te_datum['prompt']
        expected_answer = te_datum['completion']
        expected_answer_orig = te_datum['orig_completion']
        ft = chat_once_mode(question, 'ft', silent=True, return_answer=True, return_prompt=True)
        ir = chat_once_mode(question, 'ir', silent=True, return_answer=True, return_prompt=True)
        print('Q:\t', question)
        print('A:\t', expected_answer_orig)
        print('A(r):\t', expected_answer)
        if print_final_prompts:
            print('FT(p):\t', ft['prompt'])
        print('FT:\t', ft['answer'])
        if print_final_prompts:
            print('IR(p):\t', ir['prompt'])
        print('IR:\t', ir['answer'])
        print('\n')


if __name__ == '__main__':
    qualitative_comparison_experiment()