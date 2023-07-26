from urllib.parse import urlparse, urljoin
import requests
from bs4 import BeautifulSoup
import openai

# set openai key
openai.api_key = ""

def extract_date_published(text):
    prompt = "Based on this content piece, extract 1) when it was published, 2) when it was updated, 3) the main topics this piece talks about (shared in short categories of text, for e.g. 'Sales Enablement', 'Marketing Sales collaboration', etc): \n\n" + text + "\n\n. "
    response = openai.Completion.create(
        engine='text-davinci-003',
        prompt=prompt,
        max_tokens=100,
        n=1,
        stop=None,
        temperature=0.4
    )
    date_published = response.choices[0].text.strip()
    
    print('OPENAI DETECTED Content metrics')
    print(date_published)
    print('')



    return date_published


def find_content_metrics(text):

    # Find the pricing page URL
    date_published = extract_date_published(text)

    if date_published:
        # Scrape the pricing buckets from the pricing page
        
        return date_published
        # print('Pricing Buckets:', pricing_buckets)
    else:
        print('No content metrics found.')

