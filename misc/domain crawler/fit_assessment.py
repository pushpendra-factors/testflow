from urllib.parse import urlparse, urljoin
import requests
from bs4 import BeautifulSoup
import openai

openai.api_key = ''

def extract_possible_services(text):

    # IDEA: based on a domain, find out the services they are likely to need. E.g. if you see mailmodo, an email smtp server, or email visualisation tools software would be able to sell to mailmodo because of what they do

    prompt = "Based on this website home page, first determine the type of business this is. Based on this, list down the different types of solutions and software tools that would play an integral part in helping their sales, marketing, product, and developer teams. the goal is to help figure out what products might help this company do better, so that we can match these answers to real market vendors who can sell to them. The response should include specific names of produts, such as 'IP-to-domain technology', 'Attribution software', 'visualization frameworks', etc : \n\n" + text + "\n\n. "
    response = openai.Completion.create(
        engine='text-davinci-003',
        prompt=prompt,
        max_tokens=100,
        n=1,
        stop=None,
        temperature=0.7
    )
    possible_services = response.choices[0].text.strip()
    
    print('OPENAI DETECTED POSSIBLE SERVICES')
    print(possible_services)
    print('')

    return possible_services


def find_possible_services(text):

    # Find the pricing page URL
    possible_services = extract_possible_services(text)

    if possible_services:
        # Scrape the pricing buckets from the pricing page
        
        return possible_services
        # print('Pricing Buckets:', pricing_buckets)
    else:
        print('No ideal company assumption possible.')

