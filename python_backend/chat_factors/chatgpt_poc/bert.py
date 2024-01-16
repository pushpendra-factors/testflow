from transformers import AutoTokenizer, AutoModel, logging
import torch
logging.set_verbosity_error()

def get_tokenizer():
    return AutoTokenizer.from_pretrained('sentence-transformers/bert-base-nli-mean-tokens')

def get_model():
    return AutoModel.from_pretrained('sentence-transformers/bert-base-nli-mean-tokens',from_tf=True)

def mean_pooling(model_output, attention_mask):
    token_embeddings = model_output[0] #First element of model_output contains all token embeddings
    input_mask_expanded = attention_mask.unsqueeze(-1).expand(token_embeddings.size()).float()
    sum_embeddings = torch.sum(token_embeddings * input_mask_expanded, 1)
    sum_mask = torch.clamp(input_mask_expanded.sum(1), min=1e-9)
    return sum_embeddings / sum_mask

def embed_sentence(sent, tokenizer=None, model=None, normalise=False):
    tokenizer = get_tokenizer()
    model = get_model()
    ei = tokenizer([sent], padding=True, truncation=True, max_length=128, return_tensors='pt')
    #Compute token embeddings
    with torch.no_grad():
        mo = model(**ei)
    #Perform pooling. In this case, mean pooling
    pe = mean_pooling(mo, ei['attention_mask'])
    if normalise:
        pe = pe / pe.norm(dim=1)[:, None]
    return pe

def embed_sentences(sents, tokenizer=None, model=None, normalise=False):
    tokenizer = get_tokenizer()
    model = get_model()
    ei = tokenizer(sents, padding=True, truncation=True, max_length=128, return_tensors='pt')
    #Compute token embeddings
    with torch.no_grad():
        mo = model(**ei)
    #Perform pooling. In this case, mean pooling
    pe = mean_pooling(mo, ei['attention_mask'])
    if normalise:
        pe = pe / pe.norm(dim=1)[:, None]
    return pe