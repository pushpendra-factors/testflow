from transformers import AutoTokenizer, TFAutoModel, logging

import numpy as np

logging.set_verbosity_error()

def get_tokenizer():
    return AutoTokenizer.from_pretrained('sentence-transformers/bert-base-nli-mean-tokens')

def mean_pooling(model_output, attention_mask):
    token_embeddings = model_output # First element of model_output contains all token embeddings
    input_mask_expanded = np.expand_dims(attention_mask, axis=-1).astype(float)
    sum_embeddings = np.sum(token_embeddings * input_mask_expanded, axis=1)
    sum_mask = np.clip(input_mask_expanded.sum(axis=1), a_min=1e-9, a_max=None)
    return sum_embeddings / sum_mask

def embed_sentence(sentence, normalise=True):

    tokenizer = AutoTokenizer.from_pretrained("bert-base-uncased")
    model = TFAutoModel.from_pretrained("bert-base-uncased")

    # Tokenize input sentence
    inputs = tokenizer(sentence, padding=True, truncation=True,max_length=128, return_tensors="tf")

    # Get BERT model output
    outputs = model(inputs)

    # Extract embeddings from BERT model output
    cls_embedding = outputs.last_hidden_state.numpy()

    attention_mask = inputs['attention_mask'].numpy()

    # Perform mean pooling
    mean_embeddings = mean_pooling(cls_embedding, attention_mask)

    if normalise:
        mean_embeddings = mean_embeddings / np.linalg.norm(mean_embeddings, axis=1, keepdims=True)

    return mean_embeddings


def embed_sentences(sents, tokenizer=None, model=None, normalise=True):
    # Load pre-trained BERT tokenizer and model
    tokenizer = AutoTokenizer.from_pretrained("bert-base-uncased")
    model = TFAutoModel.from_pretrained("bert-base-uncased")

    # Tokenize input sentences
    inputs = tokenizer(sents, padding=True, truncation=True, max_length=128, return_tensors='tf')

    # Get BERT model output
    outputs = model(inputs)

    # Extract all token embeddings from BERT model output and convert to NumPy
    cls_embeddings = outputs.last_hidden_state.numpy()

    attention_mask = inputs['attention_mask'].numpy()

    # Perform mean pooling
    mean_embeddings = mean_pooling(cls_embeddings, attention_mask)

    # Perform normalization if specified
    if normalise:
        mean_embeddings = mean_embeddings / np.linalg.norm(mean_embeddings, axis=1, keepdims=True)

    return mean_embeddings
