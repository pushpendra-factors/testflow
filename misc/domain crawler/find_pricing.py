from urllib.parse import urlparse, urljoin
import requests
from bs4 import BeautifulSoup
import openai

openai.api_key = ''

def format_text(soup_text):
    formatted_text = ' '.join(soup_text.split())
    return formatted_text


def find_pricing_page(l1_urls):
    pricing_page_url = None

    print('traversing possible pricing pages')


    # Manually figuring out pricing page based on this
    # Use LLM here too
    pricing_keywords = ['pricing', 'plans', 'quote', 'buy', 'purchase']

    for url in l1_urls:
        path_segments = urlparse(url).path.split('/')
        first_segment = path_segments[1].lower() if len(path_segments) > 1 else ''
        
        if any(keyword in first_segment for keyword in pricing_keywords):
            return url            
    

def scrape_pricing_buckets(pricing_page_url):
    pricing_buckets = []

    if pricing_page_url:
        response = requests.get(pricing_page_url)
        soup = BeautifulSoup(response.content, 'html.parser')
        pricing_text = format_text(soup.get_text())
        # You may need to adjust the HTML elements and attributes based on the structure of the pricing page

        # print('PRICING SOUP')
        # print(pricing_text)
        # print('')

        # Use OpenAI language model to extract pricing information from the text
        extracted_pricing = extract_pricing_from_text(pricing_text)
        

    return pricing_buckets


def extract_pricing_from_text(text):
    prompt = "Extract pricing information:\n\n" + text + "\n\n Share the output as an array with strings, containing each pricing bucket with the plan name and price. If they have an Enterprise plan, share that also. Pricing:"
    response = openai.Completion.create(
        engine='text-davinci-003',
        prompt=prompt,
        max_tokens=100,
        n=1,
        stop=None,
        temperature=0.1
    )
    pricing_text = response.choices[0].text.strip()
    
    print('OPENAI PRICING TEXT')
    print(pricing_text)
    print('')

    return pricing_text


def find_pricing_elements(l1_urls):

    # Find the pricing page URL
    pricing_page_url = find_pricing_page(l1_urls)

    if pricing_page_url:
        # Scrape the pricing buckets from the pricing page
        pricing_buckets = scrape_pricing_buckets(pricing_page_url)
        print('Pricing Page URL:', pricing_page_url)
        # print('Pricing Buckets:', pricing_buckets)
    else:
        print('No pricing page found.')

