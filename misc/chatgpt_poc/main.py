import argparse
from chat import chat_once_mode
import os
os.environ['TF_CPP_MIN_LOG_LEVEL'] = '1'
os.environ['TOKENIZERS_PARALLELISM'] = 'false'

def parse_args():
    parser = argparse.ArgumentParser(prog='GPT-driven Factors API request generator',
                                     description='Prompts a fine-tuned GPT model with a natural language question to generate runnable API requests.')
    parser.add_argument('prompt', type=str, help='The question you want to ask about your data (E.g., "How many users visited our website last month?")')
    parser.add_argument('--model', type=str, default='ft', help='The model to use -- fine-tuned ("ft") or information-retrieval ("ir").')
    parser.add_argument('--scratch', action='store_true', default=False, help='If "ft", fine-tune model again, and if "ir", generate embeddings from scratch.')
    parser.add_argument('-s', '--silent', default=False, action='store_true', help='Silent')
    args = vars(parser.parse_args())
    return parser, args

if __name__ == '__main__':
    parser, args = parse_args()
    prompt = args['prompt']
    model_type = args['model']
    silent = args['silent']
    scratch = args['scratch']
    if not silent:
        print('ARGUMENTS:', args)
    chat_once_mode(prompt, model_type, parser, scratch, silent)
